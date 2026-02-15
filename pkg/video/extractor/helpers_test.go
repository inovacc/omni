package extractor

import (
	"testing"
)

func TestBaseExtractor_SearchRegex(t *testing.T) {
	b := &BaseExtractor{ExtractorName: "test"}

	t.Run("match found", func(t *testing.T) {
		result, err := b.SearchRegex(
			[]string{`title="([^"]+)"`},
			`<h1 title="Hello World">`, "title", false,
		)
		if err != nil {
			t.Fatal(err)
		}
		if result != "Hello World" {
			t.Errorf("got %q, want %q", result, "Hello World")
		}
	})

	t.Run("no match non-fatal", func(t *testing.T) {
		result, err := b.SearchRegex(
			[]string{`notfound="([^"]+)"`},
			`<h1>nothing</h1>`, "field", false,
		)
		if err != nil {
			t.Fatal(err)
		}
		if result != "" {
			t.Errorf("got %q, want empty", result)
		}
	})

	t.Run("no match fatal", func(t *testing.T) {
		_, err := b.SearchRegex(
			[]string{`notfound="([^"]+)"`},
			`<h1>nothing</h1>`, "field", true,
		)
		if err == nil {
			t.Error("expected error for fatal missing match")
		}
	})

	t.Run("multiple patterns first wins", func(t *testing.T) {
		result, err := b.SearchRegex(
			[]string{`nomatch`, `id="(\d+)"`},
			`<div id="42">`, "id", false,
		)
		if err != nil {
			t.Fatal(err)
		}
		if result != "42" {
			t.Errorf("got %q, want %q", result, "42")
		}
	})
}

func TestBaseExtractor_ParseJSON(t *testing.T) {
	b := &BaseExtractor{ExtractorName: "test"}

	t.Run("valid json", func(t *testing.T) {
		result, err := b.ParseJSON(`{"key": "value", "num": 42}`)
		if err != nil {
			t.Fatal(err)
		}
		if result["key"] != "value" {
			t.Errorf("key = %v, want %q", result["key"], "value")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := b.ParseJSON(`not json`)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestParseM3U8Formats(t *testing.T) {
	manifest := "#EXTM3U\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=640x360,CODECS=\"avc1.4d401e,mp4a.40.2\"\n" +
		"low/stream.m3u8\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=2560000,RESOLUTION=1280x720,CODECS=\"avc1.4d401f,mp4a.40.2\"\n" +
		"mid/stream.m3u8\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=7680000,RESOLUTION=1920x1080\n" +
		"high/stream.m3u8\n"

	formats := ParseM3U8Formats(manifest, "https://example.com/master.m3u8", "video123")

	if len(formats) != 3 {
		t.Fatalf("got %d formats, want 3", len(formats))
	}

	tests := []struct {
		name       string
		idx        int
		resolution string
		protocol   string
		formatID   string
	}{
		{"low", 0, "640x360", "m3u8_native", "hls-0"},
		{"mid", 1, "1280x720", "m3u8_native", "hls-1"},
		{"high", 2, "1920x1080", "m3u8_native", "hls-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := formats[tt.idx]
			if f.Resolution != tt.resolution {
				t.Errorf("resolution = %q, want %q", f.Resolution, tt.resolution)
			}
			if f.Protocol != tt.protocol {
				t.Errorf("protocol = %q, want %q", f.Protocol, tt.protocol)
			}
			if f.FormatID != tt.formatID {
				t.Errorf("formatID = %q, want %q", f.FormatID, tt.formatID)
			}
			if f.URL == "" {
				t.Error("expected non-empty URL")
			}
		})
	}

	// Verify codecs parsed for first format.
	if formats[0].VCodec != "avc1.4d401e" {
		t.Errorf("vcodec = %q, want %q", formats[0].VCodec, "avc1.4d401e")
	}
	if formats[0].ACodec != "mp4a.40.2" {
		t.Errorf("acodec = %q, want %q", formats[0].ACodec, "mp4a.40.2")
	}
}
