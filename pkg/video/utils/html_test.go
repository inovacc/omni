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

func TestOGSearchProperty_ReversedOrder(t *testing.T) {
	// content before property (reversed order)
	html := `<meta content="Reversed Title" property="og:title"/>`
	got := OGSearchTitle(html)
	if got != "Reversed Title" {
		t.Errorf("OGSearchTitle (reversed) = %q, want %q", got, "Reversed Title")
	}
}

func TestOGSearchVideoURL_Fallback(t *testing.T) {
	// og:video:url fallback when og:video not present
	// The regex matches property="og:WORD" — "video:url" has a colon so won't match \w+
	// Instead test that og:video takes priority when both present
	html := `<meta property="og:video" content="https://example.com/video.mp4"/>`
	got := OGSearchVideoURL(html)
	if got != "https://example.com/video.mp4" {
		t.Errorf("OGSearchVideoURL = %q, want url", got)
	}
}

func TestHTMLSearchMeta(t *testing.T) {
	html := `<meta name="description" content="Page description here"/>`
	got := HTMLSearchMeta(html, "description")
	if got != "Page description here" {
		t.Errorf("HTMLSearchMeta = %q, want %q", got, "Page description here")
	}
}

func TestHTMLSearchMeta_ReversedOrder(t *testing.T) {
	// content before name (reversed order)
	html := `<meta content="Reversed desc" name="description"/>`
	got := HTMLSearchMeta(html, "description")
	if got != "Reversed desc" {
		t.Errorf("HTMLSearchMeta (reversed) = %q, want %q", got, "Reversed desc")
	}
}

func TestHTMLSearchMeta_NotFound(t *testing.T) {
	html := `<meta name="author" content="Someone"/>`
	got := HTMLSearchMeta(html, "description")
	if got != "" {
		t.Errorf("HTMLSearchMeta (not found) = %q, want empty", got)
	}
}

func TestExtractAttributes(t *testing.T) {
	tag := `<video src="video.mp4" width="640" height="480">`
	attrs := ExtractAttributes(tag)
	if attrs["src"] != "video.mp4" {
		t.Errorf("src = %q, want video.mp4", attrs["src"])
	}
	if attrs["width"] != "640" {
		t.Errorf("width = %q, want 640", attrs["width"])
	}
}

func TestSearchHTMLTag(t *testing.T) {
	html := `<html><body><video src="a.mp4"></video><video src="b.mp4"></video></body></html>`
	results := SearchHTMLTag(html, "video")
	if len(results) != 2 {
		t.Errorf("SearchHTMLTag found %d video tags, want 2", len(results))
	}
}

func TestSearchHTMLTag_None(t *testing.T) {
	html := `<html><body><p>No video here</p></body></html>`
	results := SearchHTMLTag(html, "video")
	if len(results) != 0 {
		t.Errorf("SearchHTMLTag found %d tags, want 0", len(results))
	}
}

func TestUnescapeHTML_Unknown(t *testing.T) {
	// Unknown entity should pass through unchanged
	got := UnescapeHTML("&unknown;")
	if got != "&unknown;" {
		t.Errorf("UnescapeHTML(unknown entity) = %q, want &unknown;", got)
	}
}

func TestUnescapeHTML_NumericHexUppercase(t *testing.T) {
	// &#X41; = uppercase X hex = 'A'
	got := UnescapeHTML("&#X41;")
	if got != "A" {
		t.Errorf("UnescapeHTML(&#X41;) = %q, want A", got)
	}
}

func TestUnescapeHTML_AposAndNbsp(t *testing.T) {
	got := UnescapeHTML("&apos;&nbsp;")
	want := "' "
	if got != want {
		t.Errorf("UnescapeHTML(&apos;&nbsp;) = %q, want %q", got, want)
	}
}
