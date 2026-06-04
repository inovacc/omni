package gopsclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/gopsagent"
)

// startAgent boots an in-process agent in a temp config dir.
func startAgent(t *testing.T) *gopsagent.Agent {
	t.Helper()
	a := gopsagent.New(gopsagent.Options{ConfigDir: t.TempDir()})
	if err := a.Listen(); err != nil {
		t.Fatalf("agent Listen: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })
	return a
}

func TestClient_Call_OpVersion(t *testing.T) {
	a := startAgent(t)
	c := NewClient(a.Addr())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b, err := c.Call(ctx, OpVersion)
	if err != nil {
		t.Fatalf("Call OpVersion: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(b)), "go") {
		t.Errorf("OpVersion response = %q, want 'go...'", b)
	}
}

func TestClient_Call_OpRuntimeSnapshot(t *testing.T) {
	a := startAgent(t)
	c := NewClient(a.Addr())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b, err := c.Call(ctx, OpRuntimeSnapshot)
	if err != nil {
		t.Fatalf("Call OpRuntimeSnapshot: %v", err)
	}
	var snap gopsagent.Snapshot
	if err := json.Unmarshal(b, &snap); err != nil {
		t.Fatalf("Snapshot JSON: %v\nraw: %s", err, b)
	}
	if snap.GoVersion == "" || snap.Goroutines == 0 {
		t.Errorf("Snapshot fields incomplete: %+v", snap)
	}
}

func TestClient_Call_OpStats(t *testing.T) {
	a := startAgent(t)
	c := NewClient(a.Addr())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b, err := c.Call(ctx, OpStats)
	if err != nil {
		t.Fatalf("Call OpStats: %v", err)
	}
	out := string(b)
	for _, want := range []string{"goroutines=", "threads=", "gc="} {
		if !strings.Contains(out, want) {
			t.Errorf("OpStats response missing %q: %s", want, out)
		}
	}
}

func TestClient_Call_DialFailureIsClear(t *testing.T) {
	// Bind a port, then close it so the address is invalid.
	a := startAgent(t)
	addr := a.Addr()
	_ = a.Close()

	c := NewClient(addr)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	_, err := c.Call(ctx, OpVersion)
	if err == nil {
		t.Error("Call to closed agent should error")
	}
}

// TestClient_RejectsNonLoopbackAddr proves the SSRF guard: a discovered
// (same-user-writable pid file) address pointing off-host must be rejected
// with cmderr.ErrInvalidInput *before* any dial is attempted, unless the
// caller explicitly opts in to non-loopback dialing.
func TestClient_RejectsNonLoopbackAddr(t *testing.T) {
	// TEST-NET-1 (192.0.2.0/24, RFC 5737) is documentation-only and never
	// routable, so even if validation regressed the dial cannot reach anything.
	const evilAddr = "192.0.2.1:9999"

	cases := []struct {
		name string
		call func(c *Client, ctx context.Context) error
	}{
		{"Call", func(c *Client, ctx context.Context) error {
			_, err := c.Call(ctx, OpVersion)
			return err
		}},
		{"CallProfile", func(c *Client, ctx context.Context) error {
			return c.CallProfile(ctx, OpCPUProfile, time.Second, io.Discard)
		}},
		{"Stream", func(c *Client, ctx context.Context) error {
			return c.Stream(ctx, 100, io.Discard)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(evilAddr)
			// Long timeout: a regressed (non-validating) client would block on
			// the dial well past this point, so the test asserts validation
			// happens up front rather than relying on a fast dial failure.
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			start := time.Now()
			err := tc.call(c, ctx)
			if err == nil {
				t.Fatalf("%s to non-loopback %q should be rejected", tc.name, evilAddr)
			}
			if !errors.Is(err, cmderr.ErrInvalidInput) {
				t.Fatalf("%s error = %v, want cmderr.ErrInvalidInput", tc.name, err)
			}
			if elapsed := time.Since(start); elapsed > 2*time.Second {
				t.Fatalf("%s took %v — validation must reject before dialing", tc.name, elapsed)
			}
		})
	}
}

// TestClient_AllowsLoopbackVariants confirms the guard does not break the
// legitimate loopback addresses an agent actually writes.
func TestClient_AllowsLoopbackVariants(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:1", "localhost:1", "[::1]:1"} {
		if err := NewClient(addr).validateAddr(); err != nil {
			t.Errorf("validateAddr(%q) = %v, want nil", addr, err)
		}
	}
}

// TestClient_AllowNonLoopbackOptIn confirms the explicit opt-in mirrors the
// agent's AllowNonLoopback escape hatch.
func TestClient_AllowNonLoopbackOptIn(t *testing.T) {
	c := NewClient("192.0.2.1:9999")
	c.SetAllowNonLoopback(true)
	if err := c.validateAddr(); err != nil {
		t.Errorf("validateAddr with opt-in = %v, want nil", err)
	}
}

// TestClient_RejectsGarbageAddr confirms empty/malformed addresses fail closed.
func TestClient_RejectsGarbageAddr(t *testing.T) {
	for _, addr := range []string{"", "not-an-addr", "256.256.256.256:1", "127.0.0.1"} {
		if err := NewClient(addr).validateAddr(); !errors.Is(err, cmderr.ErrInvalidInput) {
			t.Errorf("validateAddr(%q) = %v, want cmderr.ErrInvalidInput", addr, err)
		}
	}
}

// silentListener accepts exactly one TCP connection, drains the client's
// request frame (auth no-op + opcode), then goes silent forever — modelling a
// planted/hung agent that completes the handshake but never streams data. The
// returned addr is loopback so it passes the SSRF guard. accepted is closed
// once the request frame has been read so the test can assert the stream was
// genuinely established before it stalled.
func silentListener(t *testing.T) (addr string, accepted <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	done := make(chan struct{})
	var once sync.Once
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		// Drain the 5-byte OpMetricsStream frame (opcode + 4-byte interval).
		// No GOPS_AGENT_KEY in the test env, so authenticate() is a no-op and
		// sends nothing first.
		var frame [5]byte
		_, _ = io.ReadFull(conn, frame[:])
		once.Do(func() { close(done) })
		// Hold the connection open but never write: the agent has gone dark.
		t.Cleanup(func() { _ = conn.Close() })
		select {} //nolint:revive // intentional silent hang for the test
	}()
	return ln.Addr().String(), done
}

// TestClient_Stream_IdleTimeout proves Stream does not block forever on a
// planted/silent agent (CWE-400): once the handshake completes, an idle read
// deadline must fire and return an error rather than hanging in io.Copy.
//
// RED note: a literal "hangs forever" assertion is timing-dependent and would
// have to wait out a watchdog, so this is structured as a fast GREEN assertion
// using a short test-only idle timeout. Against the pre-fix code (no read
// deadline at all) Stream never returns and the watchdog goroutine reports the
// hang; after the fix it returns a cmderr.ErrTimeout promptly.
func TestClient_Stream_IdleTimeout(t *testing.T) {
	addr, accepted := silentListener(t)

	c := NewClient(addr)
	c.setIdleTimeout(150 * time.Millisecond)

	ctx := context.Background()
	errCh := make(chan error, 1)
	go func() { errCh <- c.Stream(ctx, 100, io.Discard) }()

	// The stream must actually be established (frame read) before it stalls.
	select {
	case <-accepted:
	case <-time.After(2 * time.Second):
		t.Fatal("listener never received the stream request frame")
	}

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("Stream returned nil on a silent agent; expected an idle-timeout error")
		}
		if !errors.Is(err, cmderr.ErrTimeout) {
			t.Fatalf("Stream error = %v, want cmderr.ErrTimeout on idle agent", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Stream blocked indefinitely on a silent agent — no idle read deadline (CWE-400)")
	}
}

func TestDiscovery_AddrForPID(t *testing.T) {
	// Override config dir via env so AddrForPID consults the test dir.
	dir := t.TempDir()
	t.Setenv("GOPS_CONFIG_DIR", dir)
	a := gopsagent.New(gopsagent.Options{ConfigDir: dir})
	if err := a.Listen(); err != nil {
		t.Fatalf("agent Listen: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	addr, err := AddrForPID(os.Getpid())
	if err != nil {
		t.Fatalf("AddrForPID: %v", err)
	}
	if addr != a.Addr() {
		t.Errorf("AddrForPID = %q, want %q", addr, a.Addr())
	}
	if !HasAgent(os.Getpid()) {
		t.Error("HasAgent should be true for self pid after Listen")
	}
}
