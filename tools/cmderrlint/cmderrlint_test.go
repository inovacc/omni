package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// check parses src and runs the analyzer, returning the findings.
func check(t *testing.T, src string) []Finding {
	t.Helper()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	return Check(fset, f)
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		wantCount int
		wantExpr  string
	}{
		{
			name: "raw os.ErrNotExist return is flagged",
			src: `package p
import "os"
func run() error { return os.ErrNotExist }`,
			wantCount: 1,
			wantExpr:  "os.ErrNotExist",
		},
		{
			name: "fmt.Errorf %w-wrapping a raw sentinel is flagged",
			src: `package p
import (
	"fmt"
	"os"
)
func run() error { return fmt.Errorf("open: %w", os.ErrPermission) }`,
			wantCount: 1,
			wantExpr:  "os.ErrPermission",
		},
		{
			name: "io.ErrUnexpectedEOF raw return is flagged (genuine I/O error)",
			src: `package p
import "io"
func run() error { return io.ErrUnexpectedEOF }`,
			wantCount: 1,
			wantExpr:  "io.ErrUnexpectedEOF",
		},
		{
			name: "io.EOF is NOT flagged (dual-natured end-of-input control flow)",
			src: `package p
import "io"
func run() error { return io.EOF }`,
			wantCount: 0,
		},
		{
			name: "context.Canceled is NOT flagged (control flow)",
			src: `package p
import "context"
func run() error { return context.Canceled }`,
			wantCount: 0,
		},
		{
			name: "errors.Is comparison is NOT flagged (sentinel is an argument)",
			src: `package p
import (
	"errors"
	"os"
)
func run(err error) bool { return errors.Is(err, os.ErrNotExist) }`,
			wantCount: 0,
		},
		{
			name: "cmderr-wrapped sentinel is NOT flagged",
			src: `package p
import "github.com/inovacc/omni/internal/cli/cmderr"
func run() error { return cmderr.Wrap(cmderr.ErrNotFound, "missing") }`,
			wantCount: 0,
		},
		{
			name: "returning a plain err variable is NOT flagged",
			src: `package p
func run(err error) error { return err }`,
			wantCount: 0,
		},
		{
			name: "fmt.Errorf wrapping a non-stdlib error is NOT flagged",
			src: `package p
import "fmt"
func run(err error) error { return fmt.Errorf("ctx: %w", err) }`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check(t, tt.src)
			if len(got) != tt.wantCount {
				t.Fatalf("findings = %d, want %d (%v)", len(got), tt.wantCount, got)
			}

			if tt.wantExpr != "" && got[0].Expr != tt.wantExpr {
				t.Errorf("expr = %q, want %q", got[0].Expr, tt.wantExpr)
			}
		})
	}
}

// TestFindingString ensures the diagnostic line names the expression and points
// at cmderr.
func TestFindingString(t *testing.T) {
	got := check(t, `package p
import "os"
func run() error { return os.ErrNotExist }`)
	if len(got) != 1 {
		t.Fatalf("want 1 finding, got %d", len(got))
	}

	line := got[0].String()
	for _, want := range []string{"os.ErrNotExist", "cmderr"} {
		if !strings.Contains(line, want) {
			t.Errorf("diagnostic %q missing %q", line, want)
		}
	}
}
