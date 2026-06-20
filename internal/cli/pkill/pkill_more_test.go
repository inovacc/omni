package pkill

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

func TestPkillInputValidation(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		opts    Options
	}{
		{"empty pattern", "", Options{}},
		{"invalid regex", "[unterminated", Options{}},
		{"invalid signal name", ".", Options{Signal: "NOTASIGNAL", ListOnly: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(&bytes.Buffer{}, tt.pattern, tt.opts)
			if !errors.Is(err, cmderr.ErrInvalidInput) {
				t.Fatalf("Run(%q) err=%v want ErrInvalidInput", tt.pattern, err)
			}
		})
	}
}

func TestPkillNoMatchSilentExit(t *testing.T) {
	// A pattern that should match no process name.
	pattern := "zzz-omni-no-such-process-xyzzy-9999"
	err := Run(&bytes.Buffer{}, pattern, Options{ListOnly: true})
	if err == nil {
		t.Fatal("expected SilentExit(1) on no match")
	}
	// SilentExit carries exit code 1.
	if cmderr.ExitCodeFor(err) != 1 {
		t.Fatalf("no-match exit code = %d want 1", cmderr.ExitCodeFor(err))
	}

	// JSON mode prints an empty array before the silent exit.
	var buf bytes.Buffer
	_ = Run(&buf, pattern, Options{ListOnly: true, OutputFormat: output.FormatJSON})
	if !strings.Contains(buf.String(), "[]") {
		t.Errorf("json no-match should print []: %q", buf.String())
	}
}

// TestPkillListAndCount exercises the non-destructive list and count paths
// against the always-present process table. A "." pattern matches every
// process name; ListOnly/Count never send a signal so nothing is killed.
func TestPkillListAndCount(t *testing.T) {
	var list bytes.Buffer
	if err := Run(&list, ".", Options{ListOnly: true}); err != nil {
		t.Fatalf("Run list: %v", err)
	}
	if list.Len() == 0 {
		t.Skip("no processes matched in this environment")
	}

	var count bytes.Buffer
	if err := Run(&count, ".", Options{Count: true}); err != nil {
		t.Fatalf("Run count: %v", err)
	}
	if strings.TrimSpace(count.String()) == "" {
		t.Errorf("count produced no output")
	}

	// Newest selection with list (still non-destructive).
	var newest bytes.Buffer
	if err := Run(&newest, ".", Options{ListOnly: true, Newest: true}); err != nil {
		t.Fatalf("Run newest: %v", err)
	}

	// Oldest selection with count, JSON mode.
	var oldest bytes.Buffer
	if err := Run(&oldest, ".", Options{Count: true, Oldest: true, OutputFormat: output.FormatJSON}); err != nil {
		t.Fatalf("Run oldest json: %v", err)
	}
	if !strings.Contains(oldest.String(), "count") {
		t.Errorf("json count missing key: %q", oldest.String())
	}
}
