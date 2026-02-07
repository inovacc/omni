package utils

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		restrict bool
		want     string
	}{
		{"empty", "", false, "_"},
		{"simple", "video.mp4", false, "video.mp4"},
		{"spaces", "my   video   file", false, "my video file"},
		{"slashes", "path/to/video", false, "path-to-video"},
		{"backslashes", "path\\to\\video", false, "path-to-video"},
		{"special chars", "video<>:\"|?*.mp4", false, "video.mp4"},
		{"control chars", "video\x00\x01\x1f.mp4", false, "video.mp4"},
		{"trailing dots", "video...", false, "video"},
		{"unicode", "vidéo café.mp4", false, "vidéo café.mp4"},
		{"unicode restrict", "vidéo café.mp4", true, "vido caf.mp4"},
		{"long name", string(make([]byte, 300)), false, "_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input, tt.restrict)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q, %v) = %q, want %q", tt.input, tt.restrict, got, tt.want)
			}
		})
	}
}

func TestSanitizeFilenameStrict(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world.mp4", "hello world.mp4"},
		{"hello<>world", "helloworld"},
		{"", "_"},
		{"---", "---"},
	}

	for _, tt := range tests {
		got := SanitizeFilenameStrict(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeFilenameStrict(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
