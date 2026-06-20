package procutil

import (
	"context"
	"testing"
)

func TestRuntimeString(t *testing.T) {
	cases := map[Runtime]string{
		RuntimeGo:      "go",
		RuntimeNode:    "node",
		RuntimePython:  "python",
		RuntimeJava:    "java",
		RuntimeUnknown: "unknown",
	}
	for r, want := range cases {
		if got := r.String(); got != want {
			t.Errorf("Runtime(%q).String() = %q, want %q", string(r), got, want)
		}
	}
}

func TestListAll_NoError(t *testing.T) {
	// Offline: enumerates this machine's process table. We only assert it does
	// not error and the result is usable (length may legitimately vary).
	procs, err := ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	_ = procs
}

func TestKillAllMatching_NumericNonexistentPID(t *testing.T) {
	// A very high PID is overwhelmingly unlikely to exist; the numeric branch
	// runs Kill directly and returns its error without touching the process table.
	results, err := KillAllMatching(context.Background(), "0", SigTerm, "")
	// pid 0 is invalid -> Kill returns an error.
	if err == nil {
		t.Error("KillAllMatching(\"0\") expected error from invalid pid")
	}
	if len(results) != 1 {
		t.Errorf("results len = %d, want 1", len(results))
	}
}

func TestKillAllMatching_NameNoMatch(t *testing.T) {
	// A process name that cannot exist exercises the enumeration + match loop
	// and the no-match error path, all without signaling any real process.
	_, err := KillAllMatching(context.Background(), "definitely-not-a-real-process-xyz123", SigKill, "")
	if err == nil {
		t.Error("KillAllMatching with impossible name should return no-match error")
	}
}

func TestKill_NegativePID(t *testing.T) {
	if err := Kill(-5, SigTerm); err == nil {
		t.Error("Kill(-5) should error")
	}
	if err := Kill(0, SigKill); err == nil {
		t.Error("Kill(0) should error")
	}
}
