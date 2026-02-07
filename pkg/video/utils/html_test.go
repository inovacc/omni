package utils

import "testing"

func TestCleanHTML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<b>Bold</b> &amp; <i>Italic</i>", "Bold & Italic"},
		{"<!-- comment -->text", "text"},
		{"Hello &lt;world&gt;", "Hello <world>"},
	}

	for _, tt := range tests {
		got := CleanHTML(tt.input)
		if got != tt.want {
			t.Errorf("CleanHTML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestUnescapeHTML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"&amp;", "&"},
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&quot;", "\""},
		{"&#39;", "'"},
		{"&#x27;", "'"},
		{"&#x41;", "A"},
		{"no entities", "no entities"},
	}

	for _, tt := range tests {
		got := UnescapeHTML(tt.input)
		if got != tt.want {
			t.Errorf("UnescapeHTML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestOGSearchProperty(t *testing.T) {
	html := `<html><head>
<meta property="og:title" content="Test Video">
<meta property="og:description" content="A test video">
<meta property="og:video" content="https://example.com/video.mp4">
<meta property="og:image" content="https://example.com/thumb.jpg">
</head></html>`

	if got := OGSearchTitle(html); got != "Test Video" {
		t.Errorf("OGSearchTitle = %q, want %q", got, "Test Video")
	}

	if got := OGSearchDescription(html); got != "A test video" {
		t.Errorf("OGSearchDescription = %q, want %q", got, "A test video")
	}

	if got := OGSearchVideoURL(html); got != "https://example.com/video.mp4" {
		t.Errorf("OGSearchVideoURL = %q", got)
	}

	if got := OGSearchThumbnail(html); got != "https://example.com/thumb.jpg" {
		t.Errorf("OGSearchThumbnail = %q", got)
	}
}

func TestHiddenInputs(t *testing.T) {
	html := `<form>
<input type="hidden" name="csrf" value="token123">
<input type="hidden" name="session" value="abc">
<input type="text" name="visible" value="shown">
</form>`

	result := HiddenInputs(html)
	if result["csrf"] != "token123" {
		t.Errorf("csrf = %q", result["csrf"])
	}

	if result["session"] != "abc" {
		t.Errorf("session = %q", result["session"])
	}

	if _, ok := result["visible"]; ok {
		t.Error("visible should not be in hidden inputs")
	}
}
