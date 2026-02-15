package video

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandOutputTemplate(t *testing.T) {
	info := &VideoInfo{
		ID:         "abc123",
		Title:      "Test Video",
		Uploader:   "TestUser",
		UploadDate: "20240101",
		Channel:    "TestChannel",
	}

	f := &Format{
		FormatID: "22",
		Ext:      "mp4",
		Height:   intPtr(720),
	}

	tests := []struct {
		tmpl string
		want string
	}{
		{"%(id)s.%(ext)s", "abc123.mp4"},
		{"%(title)s.%(ext)s", "Test Video.mp4"},
		{"%(uploader)s/%(title)s", "TestUser/Test Video.mp4"},
		{"%(channel)s-%(id)s.%(ext)s", "TestChannel-abc123.mp4"},
		{"output", "output.mp4"}, // auto-add extension
	}

	for _, tt := range tests {
		t.Run(tt.tmpl, func(t *testing.T) {
			got := expandOutputTemplate(tt.tmpl, info, f)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteInfoJSON(t *testing.T) {
	dir := t.TempDir()
	videoPath := filepath.Join(dir, "test.mp4")

	info := &VideoInfo{
		ID:    "test123",
		Title: "Test",
	}

	c := &Client{}
	if err := c.writeInfoJSON(info, videoPath); err != nil {
		t.Fatalf("writeInfoJSON: %v", err)
	}

	jsonPath := filepath.Join(dir, "test.info.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read info.json: %v", err)
	}

	var got VideoInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != "test123" {
		t.Errorf("ID = %q, want test123", got.ID)
	}
}

func TestWriteMarkdown(t *testing.T) {
	dir := t.TempDir()
	videoPath := filepath.Join(dir, "test.mp4")

	info := &VideoInfo{
		ID:          "test123",
		Title:       "My Video",
		WebpageURL:  "https://example.com/video",
		Description: "A test video description",
	}

	c := &Client{}
	if err := c.writeMarkdown(info, videoPath); err != nil {
		t.Fatalf("writeMarkdown: %v", err)
	}

	mdPath := filepath.Join(dir, "test.md")
	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# My Video") {
		t.Error("missing title")
	}
	if !strings.Contains(content, "https://example.com/video") {
		t.Error("missing link")
	}
	if !strings.Contains(content, "A test video description") {
		t.Error("missing description")
	}
}

func TestWriteMarkdownNoLink(t *testing.T) {
	dir := t.TempDir()
	videoPath := filepath.Join(dir, "test.mp4")

	info := &VideoInfo{ID: "x", Title: "Title Only"}

	c := &Client{}
	if err := c.writeMarkdown(info, videoPath); err != nil {
		t.Fatalf("writeMarkdown: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "test.md"))
	if strings.Contains(string(data), "**Link:**") {
		t.Error("should not have link section")
	}
}

func TestEnsureOutputDirDot(t *testing.T) {
	if err := ensureOutputDir("video.mp4"); err != nil {
		t.Fatalf("ensureOutputDir for current dir: %v", err)
	}
}

func TestOrDefault(t *testing.T) {
	if orDefault("", "mp4") != "mp4" {
		t.Error("expected default")
	}
	if orDefault("webm", "mp4") != "webm" {
		t.Error("expected webm")
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	existingFile := filepath.Join(dir, "exists.txt")
	_ = os.WriteFile(existingFile, []byte("test"), 0o644)

	if !fileExists(existingFile) {
		t.Error("expected file to exist")
	}

	if fileExists(filepath.Join(dir, "nonexistent.txt")) {
		t.Error("expected file to not exist")
	}
}

func TestMergeHeaders(t *testing.T) {
	c := &Client{
		opts: Options{
			Headers: map[string]string{"X-Global": "value1", "X-Common": "global"},
		},
	}

	formatHeaders := map[string]string{"X-Format": "value2", "X-Common": "format"}

	merged := c.mergeHeaders(formatHeaders)

	if merged["X-Global"] != "value1" {
		t.Error("missing global header")
	}
	if merged["X-Format"] != "value2" {
		t.Error("missing format header")
	}
	if merged["X-Common"] != "format" {
		t.Error("format headers should override global headers")
	}
}

func intPtr(n int) *int {
	return &n
}
