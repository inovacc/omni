package m3u8

import "testing"

func TestParseMasterPlaylist(t *testing.T) {
	content := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=640x360,CODECS="avc1.4d401e,mp4a.40.2"
360p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2560000,RESOLUTION=1280x720,CODECS="avc1.4d401f,mp4a.40.2"
720p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=7680000,RESOLUTION=1920x1080,CODECS="avc1.640028,mp4a.40.2"
1080p.m3u8`

	p, err := Parse(content)
	if err != nil {
		t.Fatal(err)
	}

	if p.Type != PlaylistMaster {
		t.Errorf("type = %v, want master", p.Type)
	}

	if len(p.Variants) != 3 {
		t.Fatalf("variants = %d, want 3", len(p.Variants))
	}

	v := p.Variants[1]
	if v.URL != "720p.m3u8" {
		t.Errorf("url = %s", v.URL)
	}

	if v.Width != 1280 || v.Height != 720 {
		t.Errorf("resolution = %dx%d", v.Width, v.Height)
	}

	if v.Bandwidth != 2560000 {
		t.Errorf("bandwidth = %d", v.Bandwidth)
	}
}

func TestParseMediaPlaylist(t *testing.T) {
	content := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
segment0.ts
#EXTINF:10.0,
segment1.ts
#EXTINF:8.5,
segment2.ts
#EXT-X-ENDLIST`

	p, err := Parse(content)
	if err != nil {
		t.Fatal(err)
	}

	if p.Type != PlaylistMedia {
		t.Errorf("type = %v, want media", p.Type)
	}

	if len(p.Segments) != 3 {
		t.Fatalf("segments = %d, want 3", len(p.Segments))
	}

	if p.Segments[0].URL != "segment0.ts" {
		t.Errorf("segment[0].url = %s", p.Segments[0].URL)
	}

	if p.Segments[0].Duration != 10.0 {
		t.Errorf("segment[0].duration = %f", p.Segments[0].Duration)
	}

	if p.Segments[2].Duration != 8.5 {
		t.Errorf("segment[2].duration = %f", p.Segments[2].Duration)
	}
}

func TestParseEncryptedPlaylist(t *testing.T) {
	content := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-KEY:METHOD=AES-128,URI="https://example.com/key",IV=0x00000000000000000000000000000001
#EXTINF:10.0,
segment0.ts
#EXTINF:10.0,
segment1.ts`

	p, err := Parse(content)
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Keys) != 1 {
		t.Fatalf("keys = %d, want 1", len(p.Keys))
	}

	key := p.Keys[0]
	if key.Method != "AES-128" {
		t.Errorf("method = %s", key.Method)
	}

	if key.URI != "https://example.com/key" {
		t.Errorf("uri = %s", key.URI)
	}

	if key.IV != "0x00000000000000000000000000000001" {
		t.Errorf("iv = %s", key.IV)
	}

	// Segments should reference the key.
	if p.Segments[0].Key == nil {
		t.Error("segment[0] key is nil")
	}

	if p.Segments[0].Key.Method != "AES-128" {
		t.Errorf("segment[0] key method = %s", p.Segments[0].Key.Method)
	}
}
