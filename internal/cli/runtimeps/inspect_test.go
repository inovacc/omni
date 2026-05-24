package runtimeps

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestRunInspect_SelfJSON(t *testing.T) {
	var buf bytes.Buffer
	pid := strconv.Itoa(os.Getpid())
	if err := RunInspect(context.Background(), &buf, pid, InspectOptions{Format: "json"}); err != nil {
		t.Fatalf("RunInspect: %v", err)
	}
	var r InspectReport
	if err := json.Unmarshal(buf.Bytes(), &r); err != nil {
		t.Fatalf("JSON unmarshal: %v\n%s", err, buf.String())
	}
	if r.Process.PID != int32(os.Getpid()) {
		t.Errorf("Report.Process.PID = %d, want %d", r.Process.PID, os.Getpid())
	}
	if r.Process.Runtime != "go" {
		t.Errorf("self runtime = %q, want %q", r.Process.Runtime, "go")
	}
	if r.Process.GoVersion == "" {
		t.Error("self GoVersion should be populated")
	}
}

func TestRunInspect_InvalidPID(t *testing.T) {
	var buf bytes.Buffer
	err := RunInspect(context.Background(), &buf, "not-a-pid", InspectOptions{})
	if err == nil || !strings.Contains(err.Error(), "invalid pid") {
		t.Errorf("want 'invalid pid' error; got %v", err)
	}
}

func TestRunInspect_TableOutputContainsPID(t *testing.T) {
	var buf bytes.Buffer
	pid := strconv.Itoa(os.Getpid())
	if err := RunInspect(context.Background(), &buf, pid, InspectOptions{Format: "table"}); err != nil {
		t.Fatalf("RunInspect: %v", err)
	}
	if !strings.Contains(buf.String(), "PID:") {
		t.Errorf("table output missing 'PID:' header:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), pid) {
		t.Errorf("table output missing self PID %s:\n%s", pid, buf.String())
	}
}
