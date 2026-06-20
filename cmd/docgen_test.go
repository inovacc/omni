package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateCommandReference_Deterministic(t *testing.T) {
	var a, b bytes.Buffer
	if err := GenerateCommandReference(&a); err != nil {
		t.Fatalf("first gen: %v", err)
	}
	if err := GenerateCommandReference(&b); err != nil {
		t.Fatalf("second gen: %v", err)
	}
	if a.String() != b.String() {
		t.Fatal("output is not deterministic across two runs")
	}
}

func TestGenerateCommandReference_Smoke(t *testing.T) {
	var buf bytes.Buffer
	if err := GenerateCommandReference(&buf); err != nil {
		t.Fatalf("gen: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"# omni Command Reference",
		"## Command Tree",
		"### ls -",                  // a known stable command
		"omni ls [file...] [flags]", // exact usage line — guards against a mangled or doubled prefix
	} {
		if !strings.Contains(out, want) {
			t.Errorf("reference missing %q", want)
		}
	}
	// Negative guard: cobra's UseLine() already includes the "omni " prefix for
	// subcommands of root. Prepending another "omni " would produce "omni omni ".
	if strings.Contains(out, "omni omni ") {
		t.Error("doubled command prefix: usage line has 'omni omni ' (do not prepend 'omni ' to UseLine())")
	}
	// No timestamp leak (case-insensitive guard against "generated on <date>").
	if strings.Contains(strings.ToLower(out), "generated on") {
		t.Error("output contains a timestamp marker; must be deterministic")
	}
}
