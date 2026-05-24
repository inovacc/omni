package gopsclient

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

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
