package runtimeps

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/obfuscate"
	"github.com/inovacc/omni/pkg/procmetrics"
)

func TestRunMonitor_SingleShotJSON(t *testing.T) {
	var buf bytes.Buffer
	pid := strconv.Itoa(os.Getpid())
	if err := RunMonitor(context.Background(), &buf, pid, MonitorOptions{Format: "json"}); err != nil {
		t.Fatalf("RunMonitor: %v", err)
	}
	var m procmetrics.Metrics
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("JSON unmarshal: %v\n%s", err, buf.String())
	}
	if m.PID != int32(os.Getpid()) {
		t.Errorf("Metrics.PID = %d, want %d", m.PID, os.Getpid())
	}
}

func TestRunMonitor_InvalidPID(t *testing.T) {
	var buf bytes.Buffer
	err := RunMonitor(context.Background(), &buf, "abc", MonitorOptions{})
	if err == nil || !strings.Contains(err.Error(), "invalid pid") {
		t.Errorf("want 'invalid pid' error; got %v", err)
	}
}

func TestRunObfuscation_SelfBinaryByPID(t *testing.T) {
	var buf bytes.Buffer
	pid := strconv.Itoa(os.Getpid())
	if err := RunObfuscation(context.Background(), &buf, pid, ObfuscationOptions{Format: "json"}); err != nil {
		t.Fatalf("RunObfuscation: %v", err)
	}
	var v obfuscate.Verdict
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("JSON unmarshal: %v\n%s", err, buf.String())
	}
	if v.Verdict != obfuscate.VerdictClean {
		t.Errorf("self obfuscation verdict = %q, want %q", v.Verdict, obfuscate.VerdictClean)
	}
}

func TestRunObfuscation_BinaryPath(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable: %v", err)
	}
	var buf bytes.Buffer
	if err := RunObfuscation(context.Background(), &buf, exe, ObfuscationOptions{Format: "table"}); err != nil {
		t.Fatalf("RunObfuscation: %v", err)
	}
	if !strings.Contains(buf.String(), "Verdict:") {
		t.Errorf("table output missing 'Verdict:' header:\n%s", buf.String())
	}
}
