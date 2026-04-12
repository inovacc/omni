package youtube

import (
	"testing"

	"github.com/inovacc/omni/pkg/video/types"
)

// ---- YoutubeExtractor.Suitable ----

func TestYoutubeExtractor_Suitable(t *testing.T) {
	ext := &YoutubeExtractor{}

	tests := []struct {
		name    string
		url     string
		matches bool
	}{
		// Positive cases
		{"watch v=", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"youtu.be", "https://youtu.be/dQw4w9WgXcQ", true},
		{"shorts", "https://www.youtube.com/shorts/abc123defgh", true},
		{"embed", "https://www.youtube.com/embed/dQw4w9WgXcQ", true},
		{"v path", "https://www.youtube.com/v/dQw4w9WgXcQ", true},
		{"no scheme", "youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"mobile", "https://m.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"extra query", "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=42s", true},

		// Negative cases
		{"playlist no video", "https://www.youtube.com/playlist?list=PLtest", false},
		{"channel", "https://www.youtube.com/channel/UCxxxxxxx", false},
		{"vimeo", "https://vimeo.com/123456", false},
		{"empty", "", false},
		{"random", "https://example.com/watch?v=dQw4w9WgXcQ", false},
		{"malformed", "not-a-url", false},
		{"handle", "https://www.youtube.com/@username", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ext.Suitable(tt.url)
			if got != tt.matches {
				t.Errorf("Suitable(%q) = %v, want %v", tt.url, got, tt.matches)
			}
		})
	}
}

func TestYoutubeExtractor_Name(t *testing.T) {
	ext := &YoutubeExtractor{}
	if got := ext.Name(); got != "YouTube" {
		t.Errorf("Name() = %q, want %q", got, "YouTube")
	}
}

// ---- extractVideoID ----

func TestExtractVideoID(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"https://www.youtube.com/embed/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"https://www.youtube.com/shorts/abc123defgh", "abc123defgh"},
		{"https://www.youtube.com/watch?v=abc&t=30", "abc"},
		{"https://youtube.com/watch?v=ABCDEFGHIJK", "ABCDEFGHIJK"},
		{"https://vimeo.com/123", ""},
		{"", ""},
		{"not-a-url", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := extractVideoID(tt.url)
			if got != tt.want {
				t.Errorf("extractVideoID(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

// ---- floatFromStr ----

func TestFloatFromStr(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"0", 0},
		{"123", 123},
		{"3600", 3600},
		{"", 0},
		{"abc", 0},
		{"12.5", 12}, // decimals ignored per implementation
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := floatFromStr(tt.input)
			if got != tt.want {
				t.Errorf("floatFromStr(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---- parseMimeType ----

func TestParseMimeType(t *testing.T) {
	tests := []struct {
		mime      string
		wantExt   string
		wantVCodec string
		wantACodec string
	}{
		{
			mime:      `video/mp4; codecs="avc1.640028, mp4a.40.2"`,
			wantExt:   "mp4",
			wantVCodec: "avc1.640028",
			wantACodec: "mp4a.40.2",
		},
		{
			mime:      `audio/mp4; codecs="mp4a.40.2"`,
			wantExt:   "mp4",
			wantVCodec: "none", // audio-only sets VCodec to "none"
			wantACodec: "mp4a.40.2",
		},
		{
			mime:      `video/webm; codecs="vp9"`,
			wantExt:   "webm",
			wantVCodec: "vp9",
			wantACodec: "",
		},
		{
			mime:      `video/webm; codecs="vp09.00.10.08, opus"`,
			wantExt:   "webm",
			wantVCodec: "vp09.00.10.08",
			wantACodec: "opus",
		},
		{
			mime:      `video/x-flv`,
			wantExt:   "flv",
			wantVCodec: "",
			wantACodec: "",
		},
		{
			mime:      `audio/webm; codecs="opus"`,
			wantExt:   "webm",
			wantVCodec: "none",
			wantACodec: "opus",
		},
		{
			mime:      `video/mp4; codecs="av01.0.05M.08"`,
			wantExt:   "mp4",
			wantVCodec: "av01.0.05M.08",
			wantACodec: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			f := &types.Format{}
			parseMimeType(tt.mime, f)
			if f.Ext != tt.wantExt {
				t.Errorf("Ext = %q, want %q", f.Ext, tt.wantExt)
			}
			if f.VCodec != tt.wantVCodec {
				t.Errorf("VCodec = %q, want %q", f.VCodec, tt.wantVCodec)
			}
			if f.ACodec != tt.wantACodec {
				t.Errorf("ACodec = %q, want %q", f.ACodec, tt.wantACodec)
			}
		})
	}
}

// ---- parseFormat ----

func TestParseFormat_BasicURL(t *testing.T) {
	ext := &YoutubeExtractor{}

	fm := map[string]any{
		"url":          "https://example.com/video.mp4",
		"itag":         "18",
		"mimeType":     `video/mp4; codecs="avc1.42001E, mp4a.40.2"`,
		"qualityLabel": "360p",
		"width":        float64(640),
		"height":       float64(360),
		"bitrate":      float64(500000),
	}

	f := ext.parseFormat(fm)
	if f == nil {
		t.Fatal("parseFormat returned nil")
	}
	if f.URL != "https://example.com/video.mp4" {
		t.Errorf("URL = %q, want original URL", f.URL)
	}
	if f.FormatID != "18" {
		t.Errorf("FormatID = %q, want %q", f.FormatID, "18")
	}
	if f.FormatNote != "360p" {
		t.Errorf("FormatNote = %q, want %q", f.FormatNote, "360p")
	}
	if f.Width == nil || *f.Width != 640 {
		t.Errorf("Width = %v, want 640", f.Width)
	}
	if f.Height == nil || *f.Height != 360 {
		t.Errorf("Height = %v, want 360", f.Height)
	}
}

func TestParseFormat_SignatureCipher(t *testing.T) {
	ext := &YoutubeExtractor{}

	fm := map[string]any{
		"signatureCipher": "url=https%3A%2F%2Fexample.com%2Fvideo.mp4&s=abc",
		"itag":            "137",
	}

	f := ext.parseFormat(fm)
	if f == nil {
		t.Fatal("parseFormat returned nil for signatureCipher")
	}
	if f.URL == "" {
		t.Error("expected URL to be extracted from signatureCipher")
	}
}

func TestParseFormat_NoURL(t *testing.T) {
	ext := &YoutubeExtractor{}
	fm := map[string]any{
		"itag": "18",
	}
	f := ext.parseFormat(fm)
	if f != nil {
		t.Error("parseFormat should return nil when no URL present")
	}
}

// ---- InnerTubeRequest ----

func TestInnerTubeRequest_Basic(t *testing.T) {
	cfg := clientWeb
	body := InnerTubeRequest("dQw4w9WgXcQ", cfg)

	if body["videoId"] != "dQw4w9WgXcQ" {
		t.Errorf("videoId = %v, want %q", body["videoId"], "dQw4w9WgXcQ")
	}
	if body["contentCheckOk"] != true {
		t.Error("contentCheckOk should be true")
	}
	ctx, ok := body["context"].(map[string]any)
	if !ok {
		t.Fatal("context missing or wrong type")
	}
	client, ok := ctx["client"].(map[string]any)
	if !ok {
		t.Fatal("context.client missing")
	}
	if client["clientName"] != cfg.Name {
		t.Errorf("clientName = %v, want %q", client["clientName"], cfg.Name)
	}
}

func TestInnerTubeRequest_AndroidParams(t *testing.T) {
	cfg := clientAndroid
	body := InnerTubeRequest("abc", cfg)

	if body["params"] != "CgIQBg==" {
		t.Errorf("Android client should have params set, got %v", body["params"])
	}
	ctx := body["context"].(map[string]any)
	client := ctx["client"].(map[string]any)
	if client["androidSdkVersion"] == nil {
		t.Error("Android client should set androidSdkVersion")
	}
}

func TestInnerTubeRequest_AllClients(t *testing.T) {
	for _, cfg := range clientOrder {
		t.Run(cfg.Name, func(t *testing.T) {
			body := InnerTubeRequest("testID", cfg)
			if body["videoId"] != "testID" {
				t.Errorf("%s: videoId missing", cfg.Name)
			}
		})
	}
}

// ---- InnerTubePlayerURL ----

func TestInnerTubePlayerURL(t *testing.T) {
	url := InnerTubePlayerURL("")
	if url == "" {
		t.Error("expected non-empty URL")
	}
	urlWithKey := InnerTubePlayerURL("mykey")
	if urlWithKey == url {
		t.Error("URL with key should differ from URL without key")
	}
	// Both should contain the player endpoint.
	for _, u := range []string{url, urlWithKey} {
		if len(u) < 10 {
			t.Errorf("URL too short: %q", u)
		}
	}
}

func TestInnerTubeBrowseURL(t *testing.T) {
	url := InnerTubeBrowseURL("")
	urlWithKey := InnerTubeBrowseURL("k")
	if url == "" || urlWithKey == "" {
		t.Error("expected non-empty URLs")
	}
	if url == urlWithKey {
		t.Error("browse URLs should differ when key is set")
	}
}

func TestInnerTubeSearchURL(t *testing.T) {
	url := InnerTubeSearchURL("")
	urlWithKey := InnerTubeSearchURL("k")
	if url == urlWithKey {
		t.Error("search URLs should differ when key is set")
	}
}

// ---- InnerTubeHeaders ----

func TestInnerTubeHeaders_WebClient(t *testing.T) {
	h := InnerTubeHeaders("dQw4w9WgXcQ", clientWeb, "")
	if h["Origin"] == "" {
		t.Error("Origin header missing")
	}
	if h["X-Youtube-Client-Name"] == "" {
		t.Error("web client should set X-Youtube-Client-Name")
	}
	if h["X-Youtube-Client-Version"] == "" {
		t.Error("web client should set X-Youtube-Client-Version")
	}
}

func TestInnerTubeHeaders_AndroidClient(t *testing.T) {
	h := InnerTubeHeaders("abc", clientAndroid, "")
	// Android clients should NOT set X-Youtube headers.
	if h["X-Youtube-Client-Name"] != "" {
		t.Error("Android client should NOT set X-Youtube-Client-Name")
	}
}

func TestInnerTubeHeaders_WithAuth(t *testing.T) {
	h := InnerTubeHeaders("abc", clientWeb, "SAPISIDHASH abc123")
	if h["Authorization"] != "SAPISIDHASH abc123" {
		t.Errorf("Authorization = %q", h["Authorization"])
	}
	if h["X-Goog-AuthUser"] != "0" {
		t.Errorf("X-Goog-AuthUser = %q", h["X-Goog-AuthUser"])
	}
}

// ---- normalizeChannelURL ----

func TestNormalizeChannelURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://www.youtube.com/@user", "https://www.youtube.com/@user"},
		{"www.youtube.com/@user", "https://www.youtube.com/@user"},
		{"  youtube.com/@user  ", "https://youtube.com/@user"},
		{"http://youtube.com/@user", "http://youtube.com/@user"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeChannelURL(tt.input)
			if got != tt.want {
				t.Errorf("normalizeChannelURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---- parseVideoRenderer ----

func TestParseVideoRenderer_Basic(t *testing.T) {
	ext := &YoutubeChannelExtractor{}

	renderer := map[string]any{
		"videoId": "testVideoID",
		"title": map[string]any{
			"runs": []any{
				map[string]any{"text": "Test Title"},
			},
		},
		"lengthText": map[string]any{
			"simpleText": "10:30",
		},
		"viewCountText": map[string]any{
			"simpleText": "12,345 views",
		},
		"publishedTimeText": map[string]any{
			"simpleText": "2 days ago",
		},
		"thumbnail": map[string]any{
			"thumbnails": []any{
				map[string]any{"url": "https://i.ytimg.com/vi/testVideoID/default.jpg"},
			},
		},
	}

	entry := ext.parseVideoRenderer(renderer)
	if entry == nil {
		t.Fatal("parseVideoRenderer returned nil")
	}
	if entry.ID != "testVideoID" {
		t.Errorf("ID = %q, want %q", entry.ID, "testVideoID")
	}
	if entry.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", entry.Title, "Test Title")
	}
	if entry.Duration != 630 { // 10*60 + 30
		t.Errorf("Duration = %v, want 630", entry.Duration)
	}
	if entry.UploadDate != "2 days ago" {
		t.Errorf("UploadDate = %q, want %q", entry.UploadDate, "2 days ago")
	}
	if len(entry.Thumbnails) == 0 {
		t.Error("expected thumbnails")
	}
}

func TestParseVideoRenderer_NoVideoID(t *testing.T) {
	ext := &YoutubeChannelExtractor{}
	renderer := map[string]any{
		"title": map[string]any{"simpleText": "no id"},
	}
	entry := ext.parseVideoRenderer(renderer)
	if entry != nil {
		t.Error("parseVideoRenderer should return nil when videoId missing")
	}
}

func TestParseVideoRenderer_SimpleTextTitle(t *testing.T) {
	ext := &YoutubeChannelExtractor{}
	renderer := map[string]any{
		"videoId": "abc123",
		"title":   map[string]any{"simpleText": "Simple Title"},
	}
	entry := ext.parseVideoRenderer(renderer)
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Title != "Simple Title" {
		t.Errorf("Title = %q, want %q", entry.Title, "Simple Title")
	}
}

// ---- parseGridContents ----

func TestParseGridContents_WithContinuationToken(t *testing.T) {
	ext := &YoutubeChannelExtractor{}

	contents := []any{
		map[string]any{
			"richItemRenderer": map[string]any{
				"content": map[string]any{
					"videoRenderer": map[string]any{
						"videoId": "vid1",
						"title":   map[string]any{"simpleText": "Video 1"},
					},
				},
			},
		},
		map[string]any{
			"continuationItemRenderer": map[string]any{
				"continuationEndpoint": map[string]any{
					"continuationCommand": map[string]any{
						"token": "nextPageToken",
					},
				},
			},
		},
	}

	entries, token := ext.parseGridContents(contents)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if token != "nextPageToken" {
		t.Errorf("token = %q, want %q", token, "nextPageToken")
	}
}

func TestParseGridContents_Empty(t *testing.T) {
	ext := &YoutubeChannelExtractor{}
	entries, token := ext.parseGridContents(nil)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
	if token != "" {
		t.Errorf("expected empty token, got %q", token)
	}
}

// ---- extractChannelMetadata ----

func TestExtractChannelMetadata(t *testing.T) {
	ext := &YoutubeChannelExtractor{}

	resp := map[string]any{
		"metadata": map[string]any{
			"channelMetadataRenderer": map[string]any{
				"title":       "My Channel",
				"description": "Channel desc",
				"channelUrl":  "https://www.youtube.com/channel/UC123",
				"externalId":  "UC123",
			},
		},
		"header": map[string]any{
			"c4TabbedHeaderRenderer": map[string]any{
				"subscriberCountText": map[string]any{
					"simpleText": "1.2M subscribers",
				},
			},
		},
	}

	info := &types.VideoInfo{
		Metadata: make(map[string]string),
	}
	ext.extractChannelMetadata(resp, info)

	if info.Title != "My Channel" {
		t.Errorf("Title = %q, want %q", info.Title, "My Channel")
	}
	if info.ChannelID != "UC123" {
		t.Errorf("ChannelID = %q, want %q", info.ChannelID, "UC123")
	}
	if info.Metadata["subscriber_count"] != "1.2M subscribers" {
		t.Errorf("subscriber_count = %q", info.Metadata["subscriber_count"])
	}
}
