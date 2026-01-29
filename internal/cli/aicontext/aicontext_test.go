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
		Long:  "Concatenate FILE(s) to standard output.\n\nExamples:\n  omni cat file.txt\n  omni cat -n file.txt",
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
		"# omni - AI Context Document",
		"## Overview",
		"## Command Categories",
		"## Complete Command Reference",
		"## Library API",
		"## Architecture",
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

	// Check overview
	if ctx.Overview.Name != "omni" {
		t.Errorf("expected overview name 'omni', got %q", ctx.Overview.Name)
	}

	if len(ctx.Overview.Principles) == 0 {
		t.Error("expected non-empty principles")
	}

	// Check commands are present
	if len(ctx.Commands) == 0 {
		t.Error("expected non-empty commands")
	}

	// Find cat command
	var foundCat bool

	for _, cmd := range ctx.Commands {
		if cmd.Name == "cat" {
			foundCat = true

			if len(cmd.Flags) == 0 {
				t.Error("cat command should have flags")
			}

			break
		}
	}

	if !foundCat {
		t.Error("expected to find cat command")
	}
}

func TestRunAIContext_Compact(t *testing.T) {
	root := createTestCommand()

	var fullBuf bytes.Buffer

	err := RunAIContext(&fullBuf, root, Options{})
	if err != nil {
		t.Fatalf("RunAIContext failed: %v", err)
	}

	var compactBuf bytes.Buffer

	err = RunAIContext(&compactBuf, root, Options{Compact: true})
	if err != nil {
		t.Fatalf("RunAIContext compact failed: %v", err)
	}

	// Compact output should be shorter
	if compactBuf.Len() >= fullBuf.Len() {
		t.Error("compact output should be shorter than full output")
	}

	// Compact output should not contain examples section
	if strings.Contains(compactBuf.String(), "**Examples:**") {
		t.Error("compact output should not contain examples")
	}
}

func TestRunAIContext_CategoryFilter(t *testing.T) {
	root := createTestCommand()

	var buf bytes.Buffer

	opts := Options{Category: "Text Processing"}

	err := RunAIContext(&buf, root, opts)
	if err != nil {
		t.Fatalf("RunAIContext failed: %v", err)
	}

	output := buf.String()

	// Should contain grep (Text Processing)
	if !strings.Contains(output, "### grep") {
		t.Error("output should contain grep command")
	}

	// Should not contain cat (Core) or sqlite (Database)
	// Note: cat might still appear in overview, but not as a command reference
	catCommandRef := "### cat\n\n**Category:** Core"
	if strings.Contains(output, catCommandRef) {
		t.Error("output should not contain cat command reference when filtering by Text Processing")
	}
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"cat", "Core"},
		{"ls", "Core"},
		{"grep", "Text Processing"},
		{"sed", "Text Processing"},
		{"tar", "Archive"},
		{"jq", "Data Processing"},
		{"sqlite", "Database"},
		{"unknown_command", "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCategory(tt.name)
			if got != tt.expected {
				t.Errorf("getCategory(%q) = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestCollectFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().BoolP("verbose", "v", false, "verbose output")
	cmd.Flags().StringP("output", "o", "stdout", "output file")
	cmd.Flags().Int("count", 10, "number of items")

	flags := collectFlags(cmd)

	if len(flags) != 3 {
		t.Errorf("expected 3 flags, got %d", len(flags))
	}

	// Check for verbose flag
	var foundVerbose bool

	for _, f := range flags {
		if f.Name == "verbose" {
			foundVerbose = true

			if f.Shorthand != "v" {
				t.Errorf("verbose shorthand = %q, want 'v'", f.Shorthand)
			}

			if f.Type != "bool" {
				t.Errorf("verbose type = %q, want 'bool'", f.Type)
			}
		}
	}

	if !foundVerbose {
		t.Error("expected to find verbose flag")
	}
}

func TestExtractExamples(t *testing.T) {
	long := `Concatenate FILE(s) to standard output.

Examples:
  omni cat file.txt
  omni cat -n file.txt
  $ omni cat --json data.json

With no FILE, read standard input.`

	examples := extractExamples(long)

	if len(examples) != 3 {
		t.Errorf("expected 3 examples, got %d: %v", len(examples), examples)
	}

	expected := []string{
		"omni cat file.txt",
		"omni cat -n file.txt",
		"omni cat --json data.json",
	}

	for i, ex := range expected {
		if i < len(examples) && examples[i] != ex {
			t.Errorf("example[%d] = %q, want %q", i, examples[i], ex)
		}
	}
}

func TestBuildCategories(t *testing.T) {
	commands := []CommandInfo{
		{Name: "cat", Path: "cat", Category: "Core"},
		{Name: "ls", Path: "ls", Category: "Core"},
		{Name: "grep", Path: "grep", Category: "Text Processing"},
		{Name: "query", Path: "sqlite query", Category: "Database"}, // subcommand
	}

	categories := buildCategories(commands)

	// Should not include subcommands in category listing
	var coreFound bool

	for _, cat := range categories {
		if cat.Name == "Core" {
			coreFound = true

			if len(cat.Commands) != 2 {
				t.Errorf("Core category should have 2 commands, got %d", len(cat.Commands))
			}
		}

		if cat.Name == "Database" {
			t.Error("Database category should not exist (only has subcommand)")
		}
	}

	if !coreFound {
		t.Error("Core category should exist")
	}
}
