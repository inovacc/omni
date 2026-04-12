package generic

import (
	"testing"
)

func TestGuessExt(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/video.mp4", "mp4"},
		{"https://example.com/video.webm", "webm"},
		{"https://example.com/playlist.m3u8", "mp4"},
		{"https://example.com/video.flv", "flv"},
		{"https://example.com/video.mkv", "mkv"},
		{"https://example.com/audio.m4a", "m4a"},
		{"https://example.com/audio.mp3", "mp3"},
		{"https://example.com/video", "mp4"}, // default
		{"https://example.com/VIDEO.MP4", "mp4"}, // case insensitive
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := guessExt(tt.url)
			if got != tt.want {
				t.Errorf("guessExt(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestGuessProtocol(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/playlist.m3u8", "m3u8_native"},
		{"https://example.com/video.mp4", "https"},
		{"http://example.com/video.mp4", "http"},
		{"http://cdn.example.com/live.M3U8", "m3u8_native"}, // case insensitive
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := guessProtocol(tt.url)
			if got != tt.want {
				t.Errorf("guessProtocol(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestExtractIDFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/path/video.mp4", "video"},
		{"https://example.com/path/video.mp4?t=123", "video"},
		{"https://example.com/path/myvideo", "myvideo"},
		{"https://example.com/", "example"}, // trailing slash stripped, extension (.com) removed
		{"https://example.com/path/", "path"},   // trailing slash stripped, last segment used
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := extractIDFromURL(tt.url)
			if got != tt.want {
				t.Errorf("extractIDFromURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			"title tag",
			"<html><head><title>My Video Title</title></head></html>",
			"My Video Title",
		},
		{
			"no title",
			"<html><head></head></html>",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTitle(tt.html)
			if got != tt.want {
				t.Errorf("extractTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenericExtractor_Name(t *testing.T) {
	e := &GenericExtractor{}
	if e.Name() != "Generic" {
		t.Errorf("Name() = %q, want %q", e.Name(), "Generic")
	}
}

func TestGenericExtractor_Suitable(t *testing.T) {
	e := &GenericExtractor{}
	tests := []struct {
		url  string
		want bool
	}{
		{"https://example.com/video.mp4", true},
		{"http://example.com/video.mp4", true},
		{"ftp://example.com/video.mp4", false},
		{"rtmp://stream.example.com/live", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := e.Suitable(tt.url)
			if got != tt.want {
				t.Errorf("Suitable(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
