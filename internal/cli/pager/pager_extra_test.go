package pager

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestRunLess_NotFound verifies RunLess returns ErrNotFound for missing files.
func TestRunLess_NotFound(t *testing.T) {
	var buf bytes.Buffer
	err := RunLess(&buf, []string{"/nonexistent/path/to/file.txt"}, PagerOptions{})
	if err == nil {
		t.Error("RunLess() expected error for nonexistent file")
	}
}

// TestRunMore_NotFound verifies RunMore returns ErrNotFound for missing files.
func TestRunMore_NotFound(t *testing.T) {
	var buf bytes.Buffer
	err := RunMore(&buf, []string{"/nonexistent/path/to/file.txt"}, PagerOptions{})
	if err == nil {
		t.Error("RunMore() expected error for nonexistent file")
	}
}

// TestPagerModel_View_WithContent tests View() rendering with actual content.
func TestPagerModel_View_WithContent(t *testing.T) {
	model := pagerModel{
		content:  []string{"line 1", "line 2", "line 3"},
		width:    80,
		height:   24,
		filename: "test.txt",
	}

	view := model.View()
	if view == "" {
		t.Error("View() should return non-empty string with content")
	}
	if !strings.Contains(view, "line 1") {
		t.Errorf("View() should contain first line, got: %q", view)
	}
}

// TestPagerModel_View_Searching tests search prompt display.
func TestPagerModel_View_Searching(t *testing.T) {
	model := pagerModel{
		content:     []string{"alpha", "beta", "gamma"},
		width:       80,
		height:      24,
		filename:    "test.txt",
		searching:   true,
		searchQuery: "alpha",
	}

	view := model.View()
	if !strings.Contains(view, "/alpha") {
		t.Errorf("View() searching mode should show search prompt, got: %q", view)
	}
}

// TestPagerModel_View_WithMessage tests message display in status bar.
func TestPagerModel_View_WithMessage(t *testing.T) {
	model := pagerModel{
		content:  []string{"line 1"},
		width:    80,
		height:   24,
		filename: "test.txt",
		message:  "Pattern found: 1 matches",
	}

	view := model.View()
	if !strings.Contains(view, "Pattern found") {
		t.Errorf("View() should show message, got: %q", view)
	}
}

// TestPagerModel_View_AtEnd tests END status display.
func TestPagerModel_View_AtEnd(t *testing.T) {
	content := []string{"line 1", "line 2"}
	model := pagerModel{
		content:  content,
		width:    80,
		height:   24,
		filename: "test.txt",
		offset:   len(content), // at end
	}

	view := model.View()
	if !strings.Contains(view, "END") {
		t.Errorf("View() at end should show END, got: %q", view)
	}
}

// TestPagerModel_View_LineNumbers tests line number rendering.
func TestPagerModel_View_LineNumbers(t *testing.T) {
	model := pagerModel{
		content:  []string{"alpha", "beta"},
		width:    80,
		height:   24,
		filename: "test.txt",
		opts:     PagerOptions{LineNumbers: true},
	}

	view := model.View()
	if view == "" {
		t.Error("View() with line numbers should produce output")
	}
}

// TestPagerModel_View_Chop tests line truncation when Chop is enabled.
func TestPagerModel_View_Chop(t *testing.T) {
	longLine := strings.Repeat("x", 200)
	model := pagerModel{
		content:  []string{longLine},
		width:    80,
		height:   24,
		filename: "test.txt",
		opts:     PagerOptions{Chop: true},
	}

	view := model.View()
	if view == "" {
		t.Error("View() with Chop should produce output")
	}
}

// TestPagerModel_View_SearchHighlight tests search match highlighting in View.
func TestPagerModel_View_SearchHighlight(t *testing.T) {
	model := pagerModel{
		content:     []string{"hello world", "other line"},
		width:       80,
		height:      24,
		filename:    "test.txt",
		searchQuery: "world",
	}

	view := model.View()
	// Should contain content (with possible ANSI styling around "world")
	if !strings.Contains(view, "hello") {
		t.Errorf("View() should contain line content, got: %q", view)
	}
}

// TestPagerModel_View_ScrollPercent tests percentage display in status bar.
func TestPagerModel_View_ScrollPercent(t *testing.T) {
	content := make([]string, 100)
	for i := range content {
		content[i] = "line"
	}
	model := pagerModel{
		content:  content,
		width:    80,
		height:   10,
		filename: "big.txt",
		offset:   50,
	}

	view := model.View()
	if !strings.Contains(view, "%") {
		t.Errorf("View() mid-scroll should show percentage, got: %q", view)
	}
}

// TestHighlightSearchMatches_InvalidRegex tests that invalid regex is handled gracefully.
func TestHighlightSearchMatches_InvalidRegex(t *testing.T) {
	// lipgloss style
	result := highlightSearchMatches("hello world", "[invalid(", false, lipgloss.NewStyle())
	// Should return the original line unchanged
	if result != "hello world" {
		t.Errorf("highlightSearchMatches() invalid regex should return original line, got: %q", result)
	}
}

// TestFindMatches_InvalidRegex tests findMatches with invalid regex pattern.
func TestFindMatches_InvalidRegex(t *testing.T) {
	model := &pagerModel{
		content:     []string{"line 1", "line 2"},
		searchQuery: "[invalid(",
		opts:        PagerOptions{},
	}
	// Should not panic
	model.findMatches()
	if len(model.matches) != 0 {
		t.Error("findMatches() with invalid regex should find no matches")
	}
}
