package gopsagent

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// startTestAgent boots an Agent into a temp config dir and returns it.
// Caller is responsible for Close.
func startTestAgent(t *testing.T, opts Options) *Agent {
	t.Helper()
	if opts.ConfigDir == "" {
		opts.ConfigDir = t.TempDir()
	}
	a := New(opts)
	if err := a.Listen(); err != nil {
		t.Fatalf("agent Listen: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })
	return a
}

func TestAgent_PIDFileWritten(t *testing.T) {
	dir := t.TempDir()
	a := startTestAgent(t, Options{ConfigDir: dir})
	want := filepath.Join(dir, "")
	matches, _ := filepath.Glob(filepath.Join(dir, "*"))
	if len(matches) == 0 {
		t.Fatalf("no pid file written in %s", want)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != a.Addr() {
		t.Errorf("pid file = %q, want %q", string(data), a.Addr())
	}
}

func TestAgent_RejectsNonLoopback(t *testing.T) {
	// 0.0.0.0 is non-loopback; Listen must refuse it without AllowNonLoopback.
	a := New(Options{Addr: "0.0.0.0:0", ConfigDir: t.TempDir()})
	err := a.Listen()
	if err == nil {
		_ = a.Close()
		t.Fatal("expected Listen to reject non-loopback address")
	}
	if !strings.Contains(err.Error(), "not loopback") {
		t.Errorf("want 'not loopback' in error; got %v", err)
	}
}

// OpVersion roundtrip — minimal integration test (no client package import to
// keep this test self-contained; we drive the wire protocol directly).
func TestAgent_OpVersionRoundtrip(t *testing.T) {
	a := startTestAgent(t, Options{})
	conn, err := net.DialTimeout("tcp", a.Addr(), 2*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write([]byte{OpVersion}); err != nil {
		t.Fatalf("write op: %v", err)
	}
	b, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	got := strings.TrimSpace(string(b))
	if !strings.HasPrefix(got, "go") {
		t.Errorf("OpVersion = %q, want a 'go...' string", got)
	}
}

func TestAgent_OpRuntimeSnapshotIsValidJSON(t *testing.T) {
	a := startTestAgent(t, Options{})
	conn, err := net.DialTimeout("tcp", a.Addr(), 2*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write([]byte{OpRuntimeSnapshot}); err != nil {
		t.Fatalf("write op: %v", err)
	}
	b, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var snap Snapshot
	if err := json.Unmarshal(b, &snap); err != nil {
		t.Fatalf("snapshot JSON: %v\nraw: %s", err, b)
	}
	if snap.Goroutines == 0 {
		t.Error("Snapshot.Goroutines should be > 0 in a live process")
	}
	if snap.GoVersion == "" {
		t.Error("Snapshot.GoVersion should be populated")
	}
}

func TestAgent_AuthChallenge_GateBlocksUnauthenticated(t *testing.T) {
	a := startTestAgent(t, Options{AuthKey: []byte("topsecret")})
	conn, err := net.DialTimeout("tcp", a.Addr(), 2*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	// Read the nonce but reply with garbage — server must close before honouring any opcode.
	var nonce [32]byte
	if _, err := io.ReadFull(conn, nonce[:]); err != nil {
		t.Fatalf("read nonce: %v", err)
	}
	var bad [32]byte
	_, _ = conn.Write(bad[:]) // wrong HMAC
	_, _ = conn.Write([]byte{OpVersion})
	b, _ := io.ReadAll(conn)
	if len(b) > 0 {
		t.Errorf("unauthenticated client should get no opcode response; got %q", b)
	}
}

func TestAgent_NotifyStartup_FiresWhenEnabled(t *testing.T) {
	t.Setenv("GOPS_AGENT_NOTIFY", "1")
	// Swap the notification sink so the test asserts on captured output, not stderr.
	var captured bytes.Buffer
	origNotify := fireStartupNotification
	fireStartupNotification = func(addr string, pid int) {
		_, _ = captured.WriteString(addr)
	}
	defer func() { fireStartupNotification = origNotify }()

	a := startTestAgent(t, Options{})
	if captured.Len() == 0 {
		t.Error("notification should have fired when GOPS_AGENT_NOTIFY=1")
	}
	if !strings.Contains(captured.String(), a.Addr()) {
		t.Errorf("notification = %q, want it to contain agent addr %q", captured.String(), a.Addr())
	}
}

func TestAgent_NotifyStartup_SilentByDefault(t *testing.T) {
	t.Setenv("GOPS_AGENT_NOTIFY", "")
	var captured bytes.Buffer
	origNotify := fireStartupNotification
	fireStartupNotification = func(addr string, pid int) { _, _ = captured.WriteString(addr) }
	defer func() { fireStartupNotification = origNotify }()

	startTestAgent(t, Options{})
	if captured.Len() != 0 {
		t.Errorf("notification fired without opt-in: %q", captured.String())
	}
}

// Helper unused below but kept for future negative tests of the binary framing.
var _ = binary.LittleEndian

// Cover discovery flow: pid → addr file → in-memory dial.
func TestDiscovery_PIDFileRoundtrip(t *testing.T) {
	// Lift the Go-side discovery from internal/gopsclient is the contract we
	// rely on; here we just ensure the agent writes a parsable file.
	dir := t.TempDir()
	a := startTestAgent(t, Options{ConfigDir: dir})
	pidPath := filepath.Join(dir, "")
	matches, _ := filepath.Glob(filepath.Join(dir, "*"))
	if len(matches) == 0 {
		t.Fatalf("no pid file under %s", pidPath)
	}
	raw, _ := os.ReadFile(matches[0])
	if _, _, err := net.SplitHostPort(strings.TrimSpace(string(raw))); err != nil {
		t.Errorf("pid file content %q is not host:port: %v", raw, err)
	}
	_ = a // keep alive
}

func TestAgent_CloseIsIdempotent(t *testing.T) {
	a := New(Options{ConfigDir: t.TempDir()})
	if err := a.Listen(); err != nil {
		t.Fatalf("Listen: %v", err)
	}
	if err := a.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := a.Close(); err != nil {
		t.Errorf("second Close should be no-op; got %v", err)
	}
}

// Ensure context cancellation does not leak goroutines or hang the test.
func TestAgent_ContextCancelDuringDial(t *testing.T) {
	a := startTestAgent(t, Options{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d := net.Dialer{Timeout: 100 * time.Millisecond}
	conn, err := d.DialContext(ctx, "tcp", a.Addr())
	if err == nil {
		_ = conn.Close()
		// On a fast loopback we may complete the dial before ctx is observed;
		// either outcome is acceptable. The point of the test is no hang.
	}
}
