package utils

import (
	"net/url"
	"testing"
)

func TestURLJoin(t *testing.T) {
	tests := []struct {
		base string
		path string
		want string
	}{
		// path is empty — return base
		{"https://example.com/a/b", "", "https://example.com/a/b"},
		// path already absolute
		{"https://example.com/a", "https://other.com/b", "https://other.com/b"},
		{"https://example.com/a", "http://other.com/b", "http://other.com/b"},
		// protocol-relative, https base
		{"https://example.com/", "//cdn.example.com/file.mp4", "https://cdn.example.com/file.mp4"},
		// protocol-relative, http base
		{"http://example.com/", "//cdn.example.com/file.mp4", "http://cdn.example.com/file.mp4"},
		// relative path
		{"https://example.com/a/b/", "c/d", "https://example.com/a/b/c/d"},
		// absolute path
		{"https://example.com/a/b/", "/c/d", "https://example.com/c/d"},
	}
	for _, tt := range tests {
		got := URLJoin(tt.base, tt.path)
		if got != tt.want {
			t.Errorf("URLJoin(%q, %q) = %q, want %q", tt.base, tt.path, got, tt.want)
		}
	}
}

func TestUpdateURLQuery(t *testing.T) {
	raw := "https://example.com/video?v=abc"
	got := UpdateURLQuery(raw, map[string]string{"quality": "hd"})
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("UpdateURLQuery returned invalid URL: %v", err)
	}
	if u.Query().Get("v") != "abc" {
		t.Errorf("v param lost, got %q", u.Query().Get("v"))
	}
	if u.Query().Get("quality") != "hd" {
		t.Errorf("quality param missing, got %q", u.Query().Get("quality"))
	}
}

func TestUpdateURLQuery_Empty(t *testing.T) {
	raw := "https://example.com/video"
	got := UpdateURLQuery(raw, map[string]string{})
	if got != raw {
		t.Errorf("UpdateURLQuery(empty params) = %q, want %q", got, raw)
	}
}

func TestSanitizeURL(t *testing.T) {
	got := SanitizeURL("https://example.com/video?v=abc#fragment")
	if got != "https://example.com/video?v=abc" {
		t.Errorf("SanitizeURL = %q, want no fragment", got)
	}
}

func TestSanitizeURL_NoFragment(t *testing.T) {
	raw := "https://example.com/video?v=abc"
	got := SanitizeURL(raw)
	if got != raw {
		t.Errorf("SanitizeURL (no fragment) = %q, want %q", got, raw)
	}
}

func TestParseQueryString(t *testing.T) {
	result := ParseQueryString("foo=bar&baz=qux")
	if result["foo"] != "bar" {
		t.Errorf("foo = %q, want bar", result["foo"])
	}
	if result["baz"] != "qux" {
		t.Errorf("baz = %q, want qux", result["baz"])
	}
}

func TestParseQueryString_Empty(t *testing.T) {
	result := ParseQueryString("")
	if len(result) != 0 {
		t.Errorf("ParseQueryString('') = %v, want empty", result)
	}
}

func TestExtractQueryParam(t *testing.T) {
	got := ExtractQueryParam("https://example.com/watch?v=dQw4w9WgXcQ&t=30", "v")
	if got != "dQw4w9WgXcQ" {
		t.Errorf("ExtractQueryParam v = %q", got)
	}
}

func TestExtractQueryParam_NotFound(t *testing.T) {
	got := ExtractQueryParam("https://example.com/watch?v=abc", "t")
	if got != "" {
		t.Errorf("ExtractQueryParam (not found) = %q, want empty", got)
	}
}

func TestExtractQueryParam_Invalid(t *testing.T) {
	got := ExtractQueryParam("::not-a-url", "v")
	if got != "" {
		t.Errorf("ExtractQueryParam(invalid) = %q, want empty", got)
	}
}
