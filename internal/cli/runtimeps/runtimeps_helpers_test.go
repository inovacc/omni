package runtimeps

import (
	"bytes"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/procutil"
)

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"0", true},
		{"12345", true},
		{"12a", false},
		{"a12", false},
		{" 12", false},
		{"-1", false},
		{"99999999", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := isNumeric(tt.in); got != tt.want {
				t.Errorf("isNumeric(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestStripExe(t *testing.T) {
	tests := []struct{ in, want string }{
		{"node", "node"},
		{"node.exe", "node"},
		{"Node.EXE", "node"},
		{"foo.bar.exe", "foo.bar"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := stripExe(tt.in); got != tt.want {
				t.Errorf("stripExe(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBaseExe(t *testing.T) {
	tests := []struct{ in, want string }{
		{"node", "node"},
		{"/usr/bin/node", "node"},
		{`C:\Program Files\node\node.exe`, "node.exe"},
		{"relative/path/python3", "python3"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := baseExe(tt.in); got != tt.want {
				t.Errorf("baseExe(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestRenderTableGo covers the Go-runtime header path with module/version cols.
func TestRenderTableGo(t *testing.T) {
	var buf bytes.Buffer
	procs := []procutil.Process{
		{PID: 10, PPID: 1, Name: "omni", GoVersion: "go1.25.0", Module: "github.com/inovacc/omni", ExePath: "/usr/local/bin/omni"},
	}
	if err := renderTable(&buf, procutil.RuntimeGo, procs); err != nil {
		t.Fatalf("renderTable(go): %v", err)
	}
	out := buf.String()
	for _, want := range []string{"GO VERSION", "MODULE", "go1.25.0", "github.com/inovacc/omni", "omni"} {
		if !strings.Contains(out, want) {
			t.Errorf("go table missing %q in:\n%s", want, out)
		}
	}
}

// TestRenderTableNonGo covers the non-Go header path (no module/version cols).
func TestRenderTableNonGo(t *testing.T) {
	var buf bytes.Buffer
	procs := []procutil.Process{
		{PID: 42, PPID: 1, Name: "node", ExePath: "/usr/bin/node"},
	}
	if err := renderTable(&buf, procutil.RuntimeNode, procs); err != nil {
		t.Fatalf("renderTable(node): %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "GO VERSION") {
		t.Errorf("non-go table should not have a GO VERSION column:\n%s", out)
	}
	for _, want := range []string{"PID", "NAME", "EXE", "node", "/usr/bin/node"} {
		if !strings.Contains(out, want) {
			t.Errorf("non-go table missing %q in:\n%s", want, out)
		}
	}
}

// TestRenderTableEmpty covers the "(no … processes found)" empty branch.
func TestRenderTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := renderTable(&buf, procutil.RuntimePython, nil); err != nil {
		t.Fatalf("renderTable(empty): %v", err)
	}
	if !strings.Contains(buf.String(), "no") || !strings.Contains(buf.String(), "found") {
		t.Errorf("empty table missing the no-processes notice:\n%s", buf.String())
	}
}
