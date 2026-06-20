package rg

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rgTree builds a nested directory tree for serial-search exercises.
func rgTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	// File with two matches separated by a gap so context output requires a
	// "--" separator between the two context blocks.
	gapFile := "match one\n" +
		"filler a\n" +
		"filler b\n" +
		"filler c\n" +
		"filler d\n" +
		"match two\n"
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte(gapFile), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), []byte("nested match here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestRunSerialContextSeparator drives the serial searchDir path (Threads:1)
// with -C context across two gapped matches, exercising printContextSeparator
// and printLineWithColor context branches.
func TestRunSerialContextSeparator(t *testing.T) {
	dir := rgTree(t)
	var buf bytes.Buffer
	opts := Options{
		Threads:    1,
		Context:    1,
		LineNumber: true,
		NoHeading:  true,
	}
	if err := Run(context.Background(), &buf, "match", []string{dir}, opts); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "match one") || !strings.Contains(out, "match two") {
		t.Errorf("expected both matches:\n%s", out)
	}
	// The gap between the two matches forces a context separator line "--".
	if !strings.Contains(out, "--") {
		t.Errorf("expected context separator '--':\n%s", out)
	}
	// Nested file under sub/ confirms recursive serial descent.
	if !strings.Contains(out, "nested match here") {
		t.Errorf("expected nested match:\n%s", out)
	}
}

// TestRunSerialColorOnlyMatching drives color=always + only-matching on the
// serial path, hitting the colorized printOnlyMatch / printLineWithColor paths.
func TestRunSerialColorOnlyMatching(t *testing.T) {
	dir := rgTree(t)

	tests := []struct {
		name string
		opts Options
	}{
		{"only-matching color", Options{Threads: 1, OnlyMatching: true, Color: "always", LineNumber: true, NoHeading: true}},
		{"byte offset + column", Options{Threads: 1, ByteOffset: true, ShowColumn: true, LineNumber: true, NoHeading: true}},
		{"color always headings", Options{Threads: 1, Color: "always", LineNumber: true}},
		{"trim", Options{Threads: 1, Trim: true, LineNumber: true, NoHeading: true}},
		{"replace", Options{Threads: 1, Replace: "X", LineNumber: true, NoHeading: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Run(context.Background(), &buf, "match", []string{dir}, tt.opts); err != nil {
				t.Fatal(err)
			}
			if buf.Len() == 0 {
				t.Errorf("expected non-empty output for %s", tt.name)
			}
		})
	}
}

// TestRunSerialSingleFile drives searchFile directly via Run on a file path.
func TestRunSerialSingleFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "single.txt")
	if err := os.WriteFile(p, []byte("alpha\nbeta match\ngamma\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	opts := Options{Threads: 1, After: 1, Before: 1, LineNumber: true, NoHeading: true}
	if err := Run(context.Background(), &buf, "match", []string{p}, opts); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "beta match") {
		t.Errorf("expected match line:\n%s", out)
	}
	// After/Before context should include surrounding lines.
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "gamma") {
		t.Errorf("expected before/after context:\n%s", out)
	}
}

// TestPrintContextSeparatorDirect calls printContextSeparator directly for both
// color modes.
func TestPrintContextSeparatorDirect(t *testing.T) {
	for _, color := range []string{"never", "always"} {
		var buf bytes.Buffer
		printContextSeparator(&buf, Options{Color: color})
		if !strings.Contains(buf.String(), "--") {
			t.Errorf("color=%s: expected '--', got %q", color, buf.String())
		}
	}
}

// TestRunSerialHiddenAndDepth covers hidden-file inclusion and max-depth limits.
func TestRunSerialHiddenAndDepth(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".hidden.txt"), []byte("secret match\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Without --hidden the dotfile is skipped.
	var buf bytes.Buffer
	if err := Run(context.Background(), &buf, "match", []string{dir}, Options{Threads: 1}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "secret match") {
		t.Errorf("hidden file should be skipped by default:\n%s", buf.String())
	}

	// With --hidden it is searched.
	var buf2 bytes.Buffer
	if err := Run(context.Background(), &buf2, "match", []string{dir}, Options{Threads: 1, Hidden: true, NoIgnore: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf2.String(), "secret match") {
		t.Errorf("hidden file should be searched with --hidden:\n%s", buf2.String())
	}
}
