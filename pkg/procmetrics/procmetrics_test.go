package procmetrics

import (
	"context"
	"os"
	"testing"
)

func TestCollect_SelfPID(t *testing.T) {
	c := NewCollector()
	m, err := c.Collect(context.Background(), int32(os.Getpid()))
	if err != nil {
		t.Fatalf("Collect(self): %v", err)
	}
	if m.PID != int32(os.Getpid()) {
		t.Errorf("Metrics.PID = %d, want %d", m.PID, os.Getpid())
	}
	// MemRSS must be > 0 — any live Go test process has resident memory.
	if m.MemRSS == 0 {
		t.Error("Metrics.MemRSS should be > 0 for the running test binary")
	}
}

func TestCollect_NonexistentPID(t *testing.T) {
	c := NewCollector()
	// PID 1 always exists; use a deliberately fake large PID instead.
	_, err := c.Collect(context.Background(), 99999999)
	if err == nil {
		t.Error("Collect on a fake PID should error")
	}
}
