package youtube

import "testing"

func TestSearchExtractor_Name(t *testing.T) {
	ext := &YoutubeSearchExtractor{}
	if got := ext.Name(); got != "YoutubeSearch" {
		t.Errorf("Name() = %q, want YoutubeSearch", got)
	}
}

func TestSearchExtractor_Suitable(t *testing.T) {
	ext := &YoutubeSearchExtractor{}

	tests := []struct {
		url  string
		want bool
	}{
		{"ytsearch:golang tutorial", true},
		{"ytsearch:music", true},
		{"ytsearch:", true}, // empty query is still "suitable" (rejected in Extract)
		{"ytsearch:hello world 123", true},

		{"https://www.youtube.com/watch?v=test", false},
		{"youtube.com/watch?v=test", false},
		{"ytsearchwrong", false},
		{"", false},
		{"search:golang", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := ext.Suitable(tt.url)
			if got != tt.want {
				t.Errorf("Suitable(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
