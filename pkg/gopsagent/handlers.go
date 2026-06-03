package gopsagent

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"
)

// idleReadTimeout bounds how long a handler waits for the next byte from a
// client before giving up. Without it, a client that connects and sends
// nothing leaks its handler goroutine forever (process-04). It is generous
// enough not to disturb well-behaved clients but closes hung connections.
const idleReadTimeout = 30 * time.Second

// profilerMu serializes the process-global profilers (CPU profile and
// execution trace). runtime/pprof.StartCPUProfile and runtime/trace.Start are
// process-wide singletons: concurrent OpCPUProfile/OpTrace requests would
// otherwise interfere with one another. A request that arrives while a profile
// is already running is rejected with a busy error rather than corrupting the
// in-flight profile (process-04).
var profilerMu sync.Mutex

// Snapshot is the JSON payload of OpRuntimeSnapshot and OpMetricsStream frames.
type Snapshot struct {
	Goroutines int    `json:"goroutines"`
	Threads    int    `json:"threads"`
	GCCount    uint32 `json:"gc_count"`
	HeapAlloc  uint64 `json:"heap_alloc"`
	HeapInUse  uint64 `json:"heap_in_use"`
	NextGC     uint64 `json:"next_gc"`
	NumCPU     int    `json:"num_cpu"`
	GoVersion  string `json:"go_version"`
}

// privileged reports whether an opcode is state-changing or
// information-disclosing and must therefore only be served when an AuthKey is
// configured (process-03). With the default Options{} (no AuthKey) these
// opcodes are silently ignored so the agent is closed-by-default; read-only
// introspection opcodes remain available unauthenticated as before.
func privileged(op byte) bool {
	switch op {
	case OpShutdown, OpSetGCPercent, OpCPUProfile, OpTrace, OpHeapProfile, OpStack:
		return true
	default:
		return false
	}
}

// handle reads one opcode (after optional HMAC challenge) and dispatches.
func (a *Agent) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	authed := len(a.opts.AuthKey) > 0
	if authed {
		if !runAuthChallenge(conn, a.opts.AuthKey) {
			return
		}
	}
	// Bound the opcode read so a client that connects but never sends an
	// opcode cannot leak this handler goroutine forever (process-04).
	_ = conn.SetReadDeadline(time.Now().Add(idleReadTimeout))
	op := make([]byte, 1)
	if _, err := io.ReadFull(conn, op); err != nil {
		return
	}
	// State-changing / disclosing opcodes require an AuthKey (process-03).
	if privileged(op[0]) && !authed {
		return
	}
	switch op[0] {
	case OpMetricsStream:
		runMetricsStream(conn)
	case OpVersion:
		_, _ = fmt.Fprintln(conn, runtime.Version())
	case OpStack:
		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)
		_, _ = conn.Write(buf[:n])
	case OpGC:
		runtime.GC()
	case OpMemStats:
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		_, _ = fmt.Fprintf(conn, "alloc=%d total=%d sys=%d gc=%d\n", m.Alloc, m.TotalAlloc, m.Sys, m.NumGC)
	case OpStats:
		_, _ = fmt.Fprintf(conn, "goroutines=%d threads=%d gc=%d\n",
			runtime.NumGoroutine(), pprof.Lookup("threadcreate").Count(), readGCCount())
	case OpHeapProfile:
		_ = pprof.Lookup("heap").WriteTo(conn, 0)
	case OpCPUProfile:
		// Protocol: 4-byte LE uint32 duration in seconds follows the opcode.
		runProfile(conn, pprof.StartCPUProfile, pprof.StopCPUProfile)
	case OpTrace:
		runProfile(conn, trace.Start, trace.Stop)
	case OpSetGCPercent:
		var pct int
		_, _ = fmt.Fscanf(conn, "%d", &pct)
		debug.SetGCPercent(pct)
	case OpRuntimeSnapshot:
		writeRuntimeSnapshot(conn)
	case OpShutdown:
		if a.opts.AllowShutdown {
			go func() { _ = a.Close() }()
		}
	}
}

// runProfile reads a 4-byte LE duration (seconds, capped at 600), starts the
// profile writing into conn, sleeps, then stops. Only one process-global
// profiler may run at a time: a request arriving while another profile is
// in flight is rejected with a busy error rather than corrupting it
// (process-04).
func runProfile(conn net.Conn, start func(io.Writer) error, stop func()) {
	// Refresh the idle deadline for the framed duration read so a client that
	// sends the opcode but withholds the duration bytes cannot hang the
	// handler goroutine (process-04).
	_ = conn.SetReadDeadline(time.Now().Add(idleReadTimeout))
	var buf [4]byte
	if _, err := io.ReadFull(conn, buf[:]); err != nil {
		return
	}
	secs := binary.LittleEndian.Uint32(buf[:])
	if secs == 0 {
		secs = 30
	}
	if secs > 600 {
		secs = 600
	}
	if !profilerMu.TryLock() {
		_, _ = fmt.Fprint(conn, "error: profiler busy: another CPU profile or trace is already running")
		return
	}
	defer profilerMu.Unlock()
	if err := start(conn); err != nil {
		_, _ = fmt.Fprintf(conn, "error: %v", err)
		return
	}
	// Clear the read deadline for the duration of the profile so the long
	// sleep is not interrupted by a stale read deadline on the connection.
	_ = conn.SetReadDeadline(time.Time{})
	time.Sleep(time.Duration(secs) * time.Second)
	stop()
}

// runAuthChallenge performs the HMAC handshake: 32-byte nonce, client must
// reply with HMAC-SHA256(nonce, key) within 5 seconds. Returns true on pass.
func runAuthChallenge(conn net.Conn, key []byte) bool {
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return false
	}
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Write(nonce[:]); err != nil {
		return false
	}
	var resp [32]byte
	if _, err := io.ReadFull(conn, resp[:]); err != nil {
		return false
	}
	_ = conn.SetDeadline(time.Time{})
	return hmac.Equal(resp[:], expectedHMAC(key, nonce[:]))
}

// runMetricsStream pushes NDJSON snapshots until the client disconnects.
// Protocol: 4-byte LE uint32 interval-ms (clamped to [50ms, 60s]).
func runMetricsStream(conn net.Conn) {
	// Refresh the idle deadline for the framed interval read so a client that
	// sends the opcode but withholds the interval bytes cannot hang the
	// handler goroutine (process-04).
	_ = conn.SetReadDeadline(time.Now().Add(idleReadTimeout))
	var buf [4]byte
	if _, err := io.ReadFull(conn, buf[:]); err != nil {
		return
	}
	// The stream only writes from here on; clear the read deadline so it is
	// not torn down by a stale deadline. Client disconnects surface as an
	// encode error below.
	_ = conn.SetReadDeadline(time.Time{})
	ms := binary.LittleEndian.Uint32(buf[:])
	if ms < 50 {
		ms = 50
	}
	if ms > 60_000 {
		ms = 60_000
	}
	enc := json.NewEncoder(conn)
	t := time.NewTicker(time.Duration(ms) * time.Millisecond)
	defer t.Stop()
	for range t.C {
		if err := enc.Encode(currentSnapshot()); err != nil {
			return
		}
	}
}

func writeRuntimeSnapshot(conn net.Conn) {
	if err := json.NewEncoder(conn).Encode(currentSnapshot()); err != nil {
		_, _ = fmt.Fprintf(conn, "error: encode snapshot: %v", err)
	}
}

func currentSnapshot() Snapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return Snapshot{
		Goroutines: runtime.NumGoroutine(),
		Threads:    pprof.Lookup("threadcreate").Count(),
		GCCount:    m.NumGC,
		HeapAlloc:  m.HeapAlloc,
		HeapInUse:  m.HeapInuse,
		NextGC:     m.NextGC,
		NumCPU:     runtime.NumCPU(),
		GoVersion:  runtime.Version(),
	}
}

func readGCCount() uint32 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.NumGC
}
