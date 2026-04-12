package youtube

import "testing"

func TestPlaylistExtractor_Name(t *testing.T) {
	ext := &YoutubePlaylistExtractor{}
	if got := ext.Name(); got != "YoutubePlaylist" {
		t.Errorf("Name() = %q, want YoutubePlaylist", got)
	}
}

func TestPlaylistExtractor_Suitable(t *testing.T) {
	ext := &YoutubePlaylistExtractor{}

	tests := []struct {
		url     string
		want    bool
	}{
		{"https://www.youtube.com/playlist?list=PLtest123", true},
		{"https://m.youtube.com/playlist?list=PLtest123", true},
		{"https://www.youtube.com/playlist?list=PL_abc-XYZ", true},
		{"http://youtube.com/playlist?list=PLtest", true},
		{"youtube.com/playlist?list=PLtest", true},

		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", false},
		{"https://www.youtube.com/channel/UCtest", false},
		{"https://youtu.be/dQw4w9WgXcQ", false},
		{"https://www.youtube.com/shorts/abc", false},
		{"https://vimeo.com/playlist?list=test", false},
		{"", false},
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
