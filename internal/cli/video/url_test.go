package video

import (
	"net/url"
	"testing"
)

func TestNormalizeVideoURL_RemovesTQuery(t *testing.T) {
	in := "https://www.youtube.com/watch?v=YqHNOVlyIjU&t=208s"
	got := normalizeVideoURL(in)

	want := "https://www.youtube.com/watch?v=YqHNOVlyIjU"
	if got != want {
		t.Fatalf("normalizeVideoURL() = %q, want %q", got, want)
	}
}

func TestNormalizeVideoURL_RemovesTFragment(t *testing.T) {
	in := "https://www.youtube.com/watch?v=YqHNOVlyIjU#t=208s"
	got := normalizeVideoURL(in)

	want := "https://www.youtube.com/watch?v=YqHNOVlyIjU"
	if got != want {
		t.Fatalf("normalizeVideoURL() = %q, want %q", got, want)
	}
}

func TestNormalizeVideoURL_LeavesNonHTTPUntouched(t *testing.T) {
	in := "ytsearch:golang tutorial"

	got := normalizeVideoURL(in)
	if got != in {
		t.Fatalf("normalizeVideoURL() = %q, want %q", got, in)
	}
}

func TestNormalizeVideoURL_BareYouTubeID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"standard ID", "dQw4w9WgXcQ", "https://www.youtube.com/watch?v=dQw4w9WgXcQ"},
		{"ID with dash", "YqHNOVlyIjU", "https://www.youtube.com/watch?v=YqHNOVlyIjU"},
		{"ID with underscore", "abc_def-123", "https://www.youtube.com/watch?v=abc_def-123"},
		{"too short", "dQw4w9WgXc", "dQw4w9WgXc"},
		{"too long", "dQw4w9WgXcQQ", "dQw4w9WgXcQQ"},
		{"has space", "dQw4w9 gXcQ", "dQw4w9 gXcQ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeVideoURL(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeVideoURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeVideoURL_NoTUnchanged(t *testing.T) {
	in := "https://www.youtube.com/watch?v=YqHNOVlyIjU&list=abc123"
	got := normalizeVideoURL(in)

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", got, err)
	}

	q := parsed.Query()
	if q.Get("v") != "YqHNOVlyIjU" {
		t.Fatalf("query v = %q, want %q", q.Get("v"), "YqHNOVlyIjU")
	}

	if q.Get("list") != "abc123" {
		t.Fatalf("query list = %q, want %q", q.Get("list"), "abc123")
	}

	if q.Get("t") != "" {
		t.Fatalf("query t should be removed, got %q", q.Get("t"))
	}
}
