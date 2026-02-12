package youtube

import "testing"

func TestChannelExtractor_Suitable(t *testing.T) {
	ext := &YoutubeChannelExtractor{}

	tests := []struct {
		name    string
		url     string
		matches bool
	}{
		// Positive cases.
		{"channel ID", "https://www.youtube.com/channel/UCxxxxxxx", true},
		{"channel ID with videos", "https://www.youtube.com/channel/UCxxxxxxx/videos", true},
		{"handle", "https://www.youtube.com/@username", true},
		{"handle with videos", "https://www.youtube.com/@username/videos", true},
		{"custom URL", "https://www.youtube.com/c/ChannelName", true},
		{"custom URL with videos", "https://www.youtube.com/c/ChannelName/videos", true},
		{"no scheme", "www.youtube.com/channel/UCxxxxxxx", true},
		{"mobile", "https://m.youtube.com/channel/UCxxxxxxx", true},
		{"handle dots", "https://www.youtube.com/@user.name", true},
		{"handle hyphens", "https://www.youtube.com/@user-name", true},
		{"trailing slash", "https://www.youtube.com/@username/", true},
		{"with query params", "https://www.youtube.com/@username?view=0&sort=dd", true},

		// Negative cases.
		{"watch URL", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", false},
		{"playlist URL", "https://www.youtube.com/playlist?list=PLtest", false},
		{"shorts URL", "https://www.youtube.com/shorts/abc123", false},
		{"youtu.be", "https://youtu.be/dQw4w9WgXcQ", false},
		{"empty", "", false},
		{"random", "https://example.com/channel/test", false},
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

func TestChannelExtractor_Name(t *testing.T) {
	ext := &YoutubeChannelExtractor{}
	if got := ext.Name(); got != "YoutubeChannel" {
		t.Errorf("Name() = %q, want %q", got, "YoutubeChannel")
	}
}

func TestParseDurationText(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"1:00", 60},
		{"12:34", 754},
		{"1:02:03", 3723},
		{"0:30", 30},
		{"10:00:00", 36000},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseDurationText(tt.input)
			if got != tt.want {
				t.Errorf("parseDurationText(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseViewCount(t *testing.T) {
	tests := []struct {
		input string
		want  *int64
	}{
		{"1,234 views", int64Ptr(1234)},
		{"1234 views", int64Ptr(1234)},
		{"0 views", int64Ptr(0)},
		{"No views", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseViewCount(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseViewCount(%q) = %v, want nil", tt.input, *got)
				}
				return
			}
			if got == nil {
				t.Errorf("parseViewCount(%q) = nil, want %v", tt.input, *tt.want)
				return
			}
			if *got != *tt.want {
				t.Errorf("parseViewCount(%q) = %v, want %v", tt.input, *got, *tt.want)
			}
		})
	}
}

func int64Ptr(n int64) *int64 { return &n }
