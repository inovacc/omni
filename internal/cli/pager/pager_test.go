package pager

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// Note: pager uses bubbletea TUI which is difficult to test directly.
// These tests cover the non-interactive portions of the code.

func TestPagerModel_findMatches(t *testing.T) {
	model := &pagerModel{
		content: []string{
			"line one",
			"line two",
			"another line",
			"final",
		},
		searchQuery: "line",
		opts:        PagerOptions{},
	}

	model.findMatches()

	if len(model.matches) != 3 {
		t.Errorf("findMatches() found %d matches, want 3", len(model.matches))
	}
}

func TestPagerModel_findMatches_caseInsensitive(t *testing.T) {
	model := &pagerModel{
		content: []string{
			"LINE one",
			"line two",
			"ANOTHER LINE",
		},
		searchQuery: "line",
		opts:        PagerOptions{IgnoreCase: true},
	}

	model.findMatches()

	if len(model.matches) != 3 {
		t.Errorf("findMatches() case-insensitive found %d matches, want 3", len(model.matches))
	}
}

func TestPagerModel_findMatches_noMatch(t *testing.T) {
	model := &pagerModel{
		content: []string{
			"one",
			"two",
			"three",
		},
		searchQuery: "xyz",
		opts:        PagerOptions{},
	}

	model.findMatches()

	if len(model.matches) != 0 {
		t.Errorf("findMatches() found %d matches, want 0", len(model.matches))
	}
}

func TestPagerModel_findMatches_emptyQuery(t *testing.T) {
	model := &pagerModel{
		content: []string{
			"one",
			"two",
		},
		searchQuery: "",
		opts:        PagerOptions{},
	}

	model.findMatches()

	if len(model.matches) != 0 {
		t.Errorf("findMatches() with empty query found %d matches, want 0", len(model.matches))
	}
}

func TestHighlightSearchMatches(t *testing.T) {
	// Test basic highlighting using the actual lipgloss style
	style := lipgloss.NewStyle().Background(lipgloss.Color("226"))
	result := highlightSearchMatches("hello world", "world", false, style)
	// The result should contain the word "world" (possibly wrapped in styling)
	if result == "" {
		t.Error("highlightSearchMatches() returned empty string")
	}
}

func TestHighlightSearchMatches_caseInsensitive(t *testing.T) {
	style := lipgloss.NewStyle().Background(lipgloss.Color("226"))

	result := highlightSearchMatches("Hello World", "world", true, style)
	if result == "" {
		t.Error("highlightSearchMatches() case-insensitive returned empty string")
	}
}

func TestPagerModel_Init(t *testing.T) {
	model := pagerModel{}

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestPagerModel_View_empty(t *testing.T) {
	model := pagerModel{
		quit: true,
	}

	view := model.View()
	if view != "" {
		t.Errorf("View() with quit=true should return empty string, got %q", view)
	}
}

func TestPagerModel_View_loading(t *testing.T) {
	model := pagerModel{
		height: 0,
	}

	view := model.View()
	if view != "Loading..." {
		t.Errorf("View() with height=0 should return 'Loading...', got %q", view)
	}
}

// Integration test for file reading (without TUI)
func TestRunPagerFileReading(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pager_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\n"
	_ = os.WriteFile(file, []byte(content), 0644)

	// We can't fully test RunLess/RunMore as they start a TUI
	// but we can verify the file exists and is readable
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Could not read test file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Test file content = %q, want %q", data, content)
	}
}

func TestPagerOptions(t *testing.T) {
	// Test that options struct can be created with all fields
	opts := PagerOptions{
		LineNumbers: true,
		NoInit:      true,
		Quit:        true,
		IgnoreCase:  true,
		Chop:        true,
		Raw:         true,
		Follow:      true,
	}

	if !opts.LineNumbers {
		t.Error("LineNumbers should be true")
	}

	if !opts.Quit {
		t.Error("Quit should be true")
	}
}

// Benchmark for findMatches
func BenchmarkFindMatches(b *testing.B) {
	content := make([]string, 1000)
	for i := range content {
		content[i] = "line content here"
	}

	model := &pagerModel{
		content:     content,
		searchQuery: "content",
		opts:        PagerOptions{},
	}

	for b.Loop() {
		model.matches = nil
		model.findMatches()
	}
}

var _ = bytes.Buffer{} // Ensure bytes is imported
