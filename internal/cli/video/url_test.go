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
