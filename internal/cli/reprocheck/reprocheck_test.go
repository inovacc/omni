package reprocheck_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/reprocheck"
)

func write(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestReproIdenticalPasses(t *testing.T) {
	d := t.TempDir()
	a := write(t, d, "a-omni", []byte("BINARYBYTES"))
	b := write(t, d, "b-omni", []byte("BINARYBYTES"))
	var w bytes.Buffer
	if err := reprocheck.Run(&w, reprocheck.Options{A: []string{a}, B: []string{b}}); err != nil {
		t.Fatalf("Run(identical) = %v, want nil", err)
	}
	if !strings.Contains(w.String(), "reproducible") {
		t.Errorf("output missing OK line: %q", w.String())
	}
}

func TestReproDriftFailsConflict(t *testing.T) {
	d := t.TempDir()
	a := write(t, d, "a-omni", []byte("BINARYBYTES"))
	b := write(t, d, "b-omni", []byte("DIFFERENT!!"))
	err := reprocheck.Run(&bytes.Buffer{}, reprocheck.Options{A: []string{a}, B: []string{b}})
	if !cmderr.IsConflict(err) {
		t.Fatalf("Run(drift) = %v, want cmderr.ErrConflict", err)
	}
}

func TestReproMismatchedListLen(t *testing.T) {
	err := reprocheck.Run(&bytes.Buffer{}, reprocheck.Options{A: []string{"x"}, B: nil})
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("Run(len mismatch) = %v, want ErrInvalidInput", err)
	}
}

func TestReproMissingFile(t *testing.T) {
	d := t.TempDir()
	a := write(t, d, "a", []byte("x"))
	err := reprocheck.Run(&bytes.Buffer{}, reprocheck.Options{A: []string{a}, B: []string{filepath.Join(d, "nope")}})
	if !cmderr.IsNotFound(err) {
		t.Fatalf("Run(missing) = %v, want ErrNotFound", err)
	}
}
