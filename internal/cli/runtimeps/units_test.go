package runtimeps

import (
	"bytes"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/inovacc/omni/pkg/procmetrics"
	"github.com/inovacc/omni/pkg/procutil"
)

func TestOpcodeForName(t *testing.T) {
	known := []string{"stack", "gc", "memstats", "version", "stats", "snapshot"}
	for _, n := range known {
		t.Run(n, func(t *testing.T) {
			if _, err := OpcodeForName(n); err != nil {
				t.Errorf("OpcodeForName(%q) = err %v, want nil", n, err)
			}
		})
	}
	t.Run("unknown", func(t *testing.T) {
		_, err := OpcodeForName("bogus")
		if err == nil || !strings.Contains(err.Error(), "unknown agent cmd") {
			t.Errorf("OpcodeForName(bogus) err = %v, want 'unknown agent cmd'", err)
		}
	})
}

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		in   uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{2 * 1024, "2.00 KiB"},
		{5 * 1024 * 1024, "5.00 MiB"},
		{3 * 1024 * 1024 * 1024, "3.00 GiB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := humanBytes(tt.in); got != tt.want {
				t.Errorf("humanBytes(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestHumanBytesShort(t *testing.T) {
	tests := []struct {
		in   uint64
		want string
	}{
		{0, "0B"},
		{1023, "1023B"},
		{1024, "1.0KiB"},
		{1024 * 1024, "1.0MiB"},
		{1024 * 1024 * 1024, "1.0GiB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := humanBytesShort(tt.in); got != tt.want {
				t.Errorf("humanBytesShort(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{"short", "abc", 10, "abc"},
		{"exact", "abcde", 5, "abcde"},
		{"too long", "abcdefgh", 5, "abcd…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncate(tt.s, tt.n); got != tt.want {
				t.Errorf("truncate(%q,%d) = %q, want %q", tt.s, tt.n, got, tt.want)
			}
		})
	}
}

func TestRenderMetricsText(t *testing.T) {
	var buf bytes.Buffer
	m := procmetrics.Metrics{PID: 7, CPUPercent: 1.5, MemRSS: 2048, OpenFDs: 3}
	if err := renderMetrics(&buf, m, "table"); err != nil {
		t.Fatalf("renderMetrics: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"PID:", "CPU:", "Mem RSS:", "2.00 KiB"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderMetrics output missing %q in:\n%s", want, out)
		}
	}
}

// TestTopModelView covers the bubbletea View rendering for both the error path
// and the populated-list path without launching an interactive program.
func TestTopModelView(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		m := topModel{err: errString("boom")}
		got := m.View()
		if !strings.Contains(got, "error: boom") {
			t.Errorf("View() = %q, want it to contain the error", got)
		}
	})

	t.Run("populated", func(t *testing.T) {
		m := topModel{
			procs: []procutil.Process{
				{PID: 11, Name: "alpha-process-with-a-very-long-name-indeed", GoVersion: "go1.25.0"},
				{PID: 12, Name: "beta", GoVersion: "go1.25.0"},
			},
			perProc: map[int32]procmetrics.Metrics{
				11: {CPUPercent: 2.0, MemRSS: 4096, Goroutines: 8, GCCount: 1},
				12: {CPUPercent: 0.5, MemRSS: 1024},
			},
			selected: 0,
		}
		got := m.View()
		for _, want := range []string{"PID", "NAME", "CPU%", "beta", "selected:"} {
			if !strings.Contains(got, want) {
				t.Errorf("View() missing %q in:\n%s", want, got)
			}
		}
	})
}

// TestTopModelInitAndUpdate covers Init and the Update key-navigation and
// dataMsg branches deterministically (no TUI program, no real processes).
func TestTopModelInitAndUpdate(t *testing.T) {
	base := topModel{
		procs: []procutil.Process{{PID: 1}, {PID: 2}, {PID: 3}},
	}

	if cmd := base.Init(); cmd == nil {
		t.Error("Init() returned nil cmd, want a tick command")
	}

	// down/j moves selection forward, up/k moves it back, clamped at the ends.
	m, _ := base.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	tm := m.(topModel)
	if tm.selected != 1 {
		t.Errorf("after 'j' selected = %d, want 1", tm.selected)
	}
	m, _ = tm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	tm = m.(topModel)
	if tm.selected != 0 {
		t.Errorf("after 'k' selected = %d, want 0", tm.selected)
	}

	// quit key returns a non-nil command.
	if _, cmd := tm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}); cmd == nil {
		t.Error("'q' Update returned nil cmd, want tea.Quit")
	}

	// dataMsg with an error records the error; a success dataMsg replaces procs.
	m, _ = tm.Update(dataMsg{err: errString("collect failed")})
	tm = m.(topModel)
	if tm.err == nil {
		t.Error("dataMsg error not recorded on model")
	}
	m, _ = tm.Update(dataMsg{procs: []procutil.Process{{PID: 99}}})
	tm = m.(topModel)
	if tm.err != nil || len(tm.procs) != 1 || tm.procs[0].PID != 99 {
		t.Errorf("dataMsg success not applied: err=%v procs=%v", tm.err, tm.procs)
	}

	// tickMsg returns a batch command.
	if _, cmd := tm.Update(tickMsg{}); cmd == nil {
		t.Error("tickMsg Update returned nil cmd, want a batch")
	}
}

// errString is a tiny error type for constructing model error states in tests.
type errString string

func (e errString) Error() string { return string(e) }
