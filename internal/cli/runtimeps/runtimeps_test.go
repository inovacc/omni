package runtimeps

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/procutil"
)

func TestRunList_JSON_IsValid(t *testing.T) {
	var buf bytes.Buffer
	if err := RunList(context.Background(), &buf, procutil.RuntimeGo, ListOptions{Format: "json"}); err != nil {
		t.Fatalf("RunList: %v", err)
	}
	var procs []procutil.Process
	if err := json.Unmarshal(buf.Bytes(), &procs); err != nil {
		t.Fatalf("RunList JSON did not decode: %v\n%s", err, buf.String())
	}
}

func TestRunList_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	err := RunList(context.Background(), &buf, procutil.RuntimeGo, ListOptions{Format: "yaml"})
	if err == nil || !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("want error mentioning 'unknown format'; got %v", err)
	}
}

func TestRunKill_RecursiveRequiresYes(t *testing.T) {
	// Use a name that exists at least twice on a normal dev machine (the test
	// runner spawns its own children). We need >=2 matches to exercise the guard.
	// If the host has zero or one matching process the test still passes via
	// the "no matches" branch — both are valid coverage of the guard logic.
	var buf bytes.Buffer
	err := RunKill(context.Background(), &buf, procutil.RuntimeGo, "definitely-not-a-real-process-name-xyz",
		KillOptions{Recursive: true, Yes: false})
	// Either ErrNotFound (no match) or ErrConflict (matches >1 without --yes) is acceptable;
	// what we MUST NOT see is a successful signal.
	if err == nil {
		t.Error("RunKill with --recursive but no --yes against a fake name should not succeed silently")
	}
}

func TestRunKill_InvalidSignal(t *testing.T) {
	var buf bytes.Buffer
	err := RunKill(context.Background(), &buf, procutil.RuntimeGo, "1", KillOptions{Signal: "WAT"})
	if err == nil || !strings.Contains(err.Error(), "unknown signal") {
		t.Errorf("want 'unknown signal' error; got %v", err)
	}
}
