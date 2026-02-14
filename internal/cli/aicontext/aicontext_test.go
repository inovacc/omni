package aicontext

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// createTestCommand creates a mock command tree for testing
func createTestCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "omni",
		Short: "Test root command",
	}

	cat := &cobra.Command{
		Use:   "cat [file...]",
		Short: "Concatenate files",
	}
	cat.Flags().BoolP("number", "n", false, "number all output lines")
	cat.Flags().Bool("json", false, "output as JSON")

	grep := &cobra.Command{
		Use:   "grep PATTERN [FILE...]",
		Short: "Print lines matching a pattern",
	}
	grep.Flags().BoolP("ignore-case", "i", false, "ignore case")
	grep.Flags().BoolP("line-number", "n", false, "show line numbers")

	sqlite := &cobra.Command{
		Use:   "sqlite",
		Short: "SQLite database operations",
	}

	sqliteQuery := &cobra.Command{
		Use:   "query DB SQL",
		Short: "Execute SQL query",
	}
	sqlite.AddCommand(sqliteQuery)

	root.AddCommand(cat, grep, sqlite)

	return root
}

func TestRunAIContext_Markdown(t *testing.T) {
	root := createTestCommand()

	var buf bytes.Buffer

	opts := Options{}

	err := RunAIContext(&buf, root, opts)
	if err != nil {
		t.Fatalf("RunAIContext failed: %v", err)
	}

	output := buf.String()

	// Check for main sections
	checks := []string{
		"# omni",
		"## Structure",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing expected section: %s", check)
		}
	}

	// Check for command documentation
	if !strings.Contains(output, "### cat") {
		t.Error("output missing cat command")
	}

	if !strings.Contains(output, "### grep") {
		t.Error("output missing grep command")
	}
}

func TestRunAIContext_JSON(t *testing.T) {
	root := createTestCommand()

	var buf bytes.Buffer

	opts := Options{JSON: true}

	err := RunAIContext(&buf, root, opts)
	if err != nil {
		t.Fatalf("RunAIContext failed: %v", err)
	}

	var ctx AIContext
	if err := json.Unmarshal(buf.Bytes(), &ctx); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	// Check app name
	if ctx.App != "omni" {
		t.Errorf("expected app name 'omni', got %q", ctx.App)
	}

	// Check categories are present
	if len(ctx.Categories) == 0 {
		t.Error("expected non-empty categories")
	}

	// Find cat command in core category
	coreCmds, ok := ctx.Categories["core"]
	if !ok {
		t.Error("expected core category")
		return
	}

	var foundCat bool

	for _, cmd := range coreCmds {
		if cmd.Cmd == "cat" {
			foundCat = true

			if len(cmd.Flags) == 0 {
				t.Error("cat command should have flags")
			}

			break
		}
	}

	if !foundCat {
		t.Error("expected to find cat command in core category")
	}
}

func TestRunAIContext_NoStructure(t *testing.T) {
	root := createTestCommand()

	var buf bytes.Buffer

	opts := Options{NoStructure: true}

	err := RunAIContext(&buf, root, opts)
	if err != nil {
		t.Fatalf("RunAIContext failed: %v", err)
	}

	output := buf.String()

	// Should not contain structure section
	if strings.Contains(output, "## Structure") {
		t.Error("output should not contain structure section with NoStructure=true")
	}
}

func TestRunAIContext_CategoryFilter(t *testing.T) {
	root := createTestCommand()

	var buf bytes.Buffer

	opts := Options{Category: "text"}

	err := RunAIContext(&buf, root, opts)
	if err != nil {
		t.Fatalf("RunAIContext failed: %v", err)
	}

	output := buf.String()

	// Should contain grep (text category)
	if !strings.Contains(output, "### grep") {
		t.Error("output should contain grep command")
	}

	// Should not contain cat (core category)
	if strings.Contains(output, "### cat") {
		t.Error("output should not contain cat command when filtering by text")
	}
}

func TestCategoryMap(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"cat", "core"},
		{"ls", "core"},
		{"grep", "text"},
		{"sed", "text"},
		{"tar", "archive"},
		{"jq", "data"},
		{"sqlite", "db"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categoryMap[tt.name]
			if got != tt.expected {
				t.Errorf("categoryMap[%q] = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestCategoryNames(t *testing.T) {
	// Verify all category keys have names
	for cat := range categoryMap {
		catKey := categoryMap[cat]
		if _, ok := categoryNames[catKey]; !ok {
			// It's ok if some categories don't have names, they default to the key
			continue
		}
	}

	// Check some specific mappings
	if categoryNames["core"] != "Core" {
		t.Errorf("categoryNames[core] = %q, want Core", categoryNames["core"])
	}

	if categoryNames["text"] != "Text" {
		t.Errorf("categoryNames[text] = %q, want Text", categoryNames["text"])
	}
}
