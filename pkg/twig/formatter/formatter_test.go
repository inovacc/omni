package formatter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/twig/models"
)

func TestDefaultFormatConfig(t *testing.T) {
	config := DefaultFormatConfig()

	if config == nil {
		t.Fatal("DefaultFormatConfig() returned nil")
	}

	if !config.ShowColors {
		t.Error("ShowColors should be true by default")
	}

	if !config.ShowDirSlash {
		t.Error("ShowDirSlash should be true by default")
	}

	if config.ShowSize {
		t.Error("ShowSize should be false by default")
	}

	if config.ShowDate {
		t.Error("ShowDate should be false by default")
	}
}

func TestNewFormatter(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		f := NewFormatter(nil)
		if f == nil {
			t.Fatal("NewFormatter(nil) returned nil")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &FormatConfig{ShowSize: true}

		f := NewFormatter(config)
		if f == nil {
			t.Fatal("NewFormatter() returned nil")
		}
	})
}

func TestFormatter_FormatSimple(t *testing.T) {
	root := models.NewNode("project", "/project", true)
	src := models.NewNode("src", "/project/src", true)
	main := models.NewNode("main.go", "/project/src/main.go", false)
	readme := models.NewNode("README.md", "/project/README.md", false)

	root.AddChild(src)
	root.AddChild(readme)
	src.AddChild(main)

	f := NewFormatter(&FormatConfig{ShowDirSlash: true})
	output := f.FormatSimple(root)

	if !strings.Contains(output, "project/") {
		t.Error("Output should contain 'project/'")
	}

	if !strings.Contains(output, "src/") {
		t.Error("Output should contain 'src/'")
	}

	if !strings.Contains(output, "main.go") {
		t.Error("Output should contain 'main.go'")
	}

	if !strings.Contains(output, "README.md") {
		t.Error("Output should contain 'README.md'")
	}

	if !strings.Contains(output, "├──") || !strings.Contains(output, "└──") {
		t.Error("Output should contain tree connectors")
	}
}

func TestFormatter_FormatSimple_Nil(t *testing.T) {
	f := NewFormatter(nil)
	output := f.FormatSimple(nil)

	if output != "" {
		t.Errorf("FormatSimple(nil) should return empty string, got %q", output)
	}
}

func TestFormatter_FormatSimple_WithHash(t *testing.T) {
	root := models.NewNode("project", "/project", true)
	file := models.NewNode("file.txt", "/project/file.txt", false)
	file.Hash = "abc123def456"

	root.AddChild(file)

	f := NewFormatter(&FormatConfig{ShowHash: true})
	output := f.FormatSimple(root)

	if !strings.Contains(output, "abc123def456") {
		t.Error("Output should contain file hash")
	}

	if !strings.Contains(output, "[abc123def456]") {
		t.Error("Output should contain hash in brackets")
	}
}

func TestFormatter_FormatSimple_WithComment(t *testing.T) {
	root := models.NewNode("project", "/project", true)
	file := models.NewNode("config.yaml", "/project/config.yaml", false)
	file.Comment = "configuration file"

	root.AddChild(file)

	f := NewFormatter(nil)
	output := f.FormatSimple(root)

	if !strings.Contains(output, "# configuration file") {
		t.Error("Output should contain comment")
	}
}

func TestFormatter_FormatSimple_NoDirSlash(t *testing.T) {
	root := models.NewNode("project", "/project", true)
	subdir := models.NewNode("src", "/project/src", true)

	root.AddChild(subdir)

	f := NewFormatter(&FormatConfig{ShowDirSlash: false})
	output := f.FormatSimple(root)

	if strings.Contains(output, "src/") {
		t.Error("Output should not contain trailing slash when ShowDirSlash is false")
	}

	if !strings.Contains(output, "src") {
		t.Error("Output should contain 'src' without slash")
	}
}

func TestFormatter_Format(t *testing.T) {
	root := models.NewNode("project", "/project", true)
	file := models.NewNode("main.go", "/project/main.go", false)

	root.AddChild(file)

	f := NewFormatter(&FormatConfig{ShowColors: false})
	output := f.Format(root)

	if output == "" {
		t.Error("Format() should produce output")
	}

	if !strings.Contains(output, "project") {
		t.Error("Format() output should contain 'project'")
	}

	if !strings.Contains(output, "main.go") {
		t.Error("Format() output should contain 'main.go'")
	}
}

func TestFormatter_Format_Nil(t *testing.T) {
	f := NewFormatter(nil)
	output := f.Format(nil)

	if output != "" {
		t.Errorf("Format(nil) should return empty string, got %q", output)
	}
}

func TestFormatter_FormatJSON(t *testing.T) {
	root := models.NewNode("project", "/project", true)
	file := models.NewNode("main.go", "/project/main.go", false)

	root.AddChild(file)

	stats := &models.TreeStats{
		TotalDirs:  1,
		TotalFiles: 1,
		MaxDepth:   1,
	}

	f := NewFormatter(nil)

	output, err := f.FormatJSON(root, stats)
	if err != nil {
		t.Fatalf("FormatJSON() error = %v", err)
	}

	if !strings.Contains(output, "\"tree\"") {
		t.Error("JSON should contain 'tree' key")
	}

	if !strings.Contains(output, "\"stats\"") {
		t.Error("JSON should contain 'stats' key")
	}

	if !strings.Contains(output, "\"name\": \"project\"") {
		t.Error("JSON should contain project name")
	}

	if !strings.Contains(output, "\"total_dirs\": 1") {
		t.Error("JSON should contain total_dirs")
	}
}

func TestFormatter_FormatJSON_NoStats(t *testing.T) {
	root := models.NewNode("project", "/project", true)

	f := NewFormatter(nil)

	output, err := f.FormatJSON(root, nil)
	if err != nil {
		t.Fatalf("FormatJSON() error = %v", err)
	}

	if !strings.Contains(output, "\"tree\"") {
		t.Error("JSON should contain 'tree' key")
	}

	if strings.Contains(output, "\"stats\"") {
		t.Error("JSON should not contain 'stats' when nil")
	}
}

func TestFormatter_FormatJSON_Nil(t *testing.T) {
	f := NewFormatter(nil)

	output, err := f.FormatJSON(nil, nil)
	if err != nil {
		t.Fatalf("FormatJSON(nil) error = %v", err)
	}

	if output != "{}" {
		t.Errorf("FormatJSON(nil) should return '{}', got %q", output)
	}
}

func TestFormatter_FlattenFilesHash(t *testing.T) {
	root := models.NewNode("project", "/project", true)
	dir := models.NewNode("src", "/project/src", true)
	file1 := models.NewNode("main.go", "/project/src/main.go", false)
	file1.Hash = "hash1"
	file2 := models.NewNode("README.md", "/project/README.md", false)
	file2.Hash = "hash2"

	root.AddChild(dir)
	root.AddChild(file2)
	dir.AddChild(file1)

	f := NewFormatter(&FormatConfig{FlattenFilesHash: true})
	output := f.FormatSimple(root)

	if !strings.Contains(output, "hash1") {
		t.Error("Output should contain hash1")
	}

	if !strings.Contains(output, "hash2") {
		t.Error("Output should contain hash2")
	}

	if !strings.Contains(output, "project/src/main.go") {
		t.Error("Output should contain flattened path")
	}
}

func TestFormatter_DeepNesting(t *testing.T) {
	root := models.NewNode("root", "/root", true)
	current := root

	for range 5 {
		dir := models.NewNode("level", "/level", true)
		current.AddChild(dir)
		current = dir
	}

	file := models.NewNode("deep.txt", "/deep.txt", false)
	current.AddChild(file)

	f := NewFormatter(&FormatConfig{ShowDirSlash: true})
	output := f.FormatSimple(root)

	// Check output contains deeply nested file
	if !strings.Contains(output, "deep.txt") {
		t.Error("Output should contain deeply nested file")
	}

	// Check output has multiple levels
	levelCount := strings.Count(output, "level")
	if levelCount != 5 {
		t.Errorf("Output should contain 5 'level' directories, got %d", levelCount)
	}
}

func TestFormatter_FormatJSONStream(t *testing.T) {
	root := models.NewNode("project", "/project", true)
	src := models.NewNode("src", "/project/src", true)
	main := models.NewNode("main.go", "/project/src/main.go", false)
	main.Hash = "abc123"

	root.AddChild(src)
	src.AddChild(main)

	stats := &models.TreeStats{
		TotalDirs:  2,
		TotalFiles: 1,
		MaxDepth:   2,
	}

	f := NewFormatter(nil)

	var buf bytes.Buffer

	err := f.FormatJSONStream(&buf, root, stats)
	if err != nil {
		t.Fatalf("FormatJSONStream() error = %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have: begin, node(project), node(src), node(main.go), stats, end = 6 lines
	if len(lines) != 6 {
		t.Errorf("expected 6 NDJSON lines, got %d", len(lines))

		for i, line := range lines {
			t.Logf("line %d: %s", i, line)
		}
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var msg StreamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}

	// Verify first line is begin
	var firstMsg StreamMessage

	_ = json.Unmarshal([]byte(lines[0]), &firstMsg)

	if firstMsg.Type != "begin" {
		t.Errorf("first message type should be 'begin', got %q", firstMsg.Type)
	}

	// Verify last line is end
	var lastMsg StreamMessage

	_ = json.Unmarshal([]byte(lines[len(lines)-1]), &lastMsg)

	if lastMsg.Type != "end" {
		t.Errorf("last message type should be 'end', got %q", lastMsg.Type)
	}

	// Verify stats line
	var statsMsg StreamMessage

	_ = json.Unmarshal([]byte(lines[len(lines)-2]), &statsMsg)

	if statsMsg.Type != "stats" {
		t.Errorf("second to last message type should be 'stats', got %q", statsMsg.Type)
	}
}

func TestFormatter_FormatJSONStream_Nil(t *testing.T) {
	f := NewFormatter(nil)

	var buf bytes.Buffer

	err := f.FormatJSONStream(&buf, nil, nil)
	if err != nil {
		t.Fatalf("FormatJSONStream(nil) error = %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("FormatJSONStream(nil) should produce empty output, got %q", buf.String())
	}
}

func TestFormatter_FormatJSONStream_NoStats(t *testing.T) {
	root := models.NewNode("root", "/root", true)
	file := models.NewNode("file.txt", "/root/file.txt", false)

	root.AddChild(file)

	f := NewFormatter(nil)

	var buf bytes.Buffer

	err := f.FormatJSONStream(&buf, root, nil)
	if err != nil {
		t.Fatalf("FormatJSONStream() error = %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have: begin, node(root), node(file.txt), end = 4 lines (no stats)
	if len(lines) != 4 {
		t.Errorf("expected 4 NDJSON lines without stats, got %d", len(lines))
	}

	// Verify no stats message
	for _, line := range lines {
		var msg StreamMessage

		_ = json.Unmarshal([]byte(line), &msg)

		if msg.Type == "stats" {
			t.Error("should not have stats message when stats is nil")
		}
	}
}
