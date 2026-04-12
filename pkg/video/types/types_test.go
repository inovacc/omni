package types

import (
	"errors"
	"testing"
)

func ptr[T any](v T) *T { return &v }

func TestFormat_HasVideo(t *testing.T) {
	tests := []struct {
		name   string
		vcodec string
		want   bool
	}{
		{"empty vcodec", "", false},
		{"none vcodec", "none", false},
		{"h264", "avc1.42001E", true},
		{"vp9", "vp9", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Format{VCodec: tt.vcodec}
			if got := f.HasVideo(); got != tt.want {
				t.Errorf("HasVideo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormat_HasAudio(t *testing.T) {
	tests := []struct {
		name   string
		acodec string
		want   bool
	}{
		{"empty acodec", "", false},
		{"none acodec", "none", false},
		{"mp4a", "mp4a.40.2", true},
		{"opus", "opus", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Format{ACodec: tt.acodec}
			if got := f.HasAudio(); got != tt.want {
				t.Errorf("HasAudio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormat_GetFilesize(t *testing.T) {
	t.Run("explicit filesize", func(t *testing.T) {
		f := Format{Filesize: ptr(int64(1024))}
		if got := f.GetFilesize(); got != 1024 {
			t.Errorf("GetFilesize() = %d, want 1024", got)
		}
	})
	t.Run("approx filesize fallback", func(t *testing.T) {
		f := Format{FilesizeApprox: ptr(int64(2048))}
		if got := f.GetFilesize(); got != 2048 {
			t.Errorf("GetFilesize() = %d, want 2048", got)
		}
	})
	t.Run("zero when neither set", func(t *testing.T) {
		f := Format{}
		if got := f.GetFilesize(); got != 0 {
			t.Errorf("GetFilesize() = %d, want 0", got)
		}
	})
}

func TestFormat_FormatResolution(t *testing.T) {
	tests := []struct {
		name       string
		resolution string
		width      *int
		height     *int
		want       string
	}{
		{"explicit resolution", "1920x1080", nil, nil, "1920x1080"},
		{"height only", "", nil, ptr(720), "720p"},
		{"width and height", "", ptr(1280), ptr(720), "1280x720"},
		{"no info", "", nil, nil, "audio only"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Format{Resolution: tt.resolution, Width: tt.width, Height: tt.height}
			if got := f.FormatResolution(); got != tt.want {
				t.Errorf("FormatResolution() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractorError(t *testing.T) {
	cause := errors.New("network timeout")
	e := &ExtractorError{
		Extractor: "YouTube",
		VideoID:   "abc123",
		Message:   "fetch failed",
		Cause:     cause,
	}
	got := e.Error()
	want := "YouTube [abc123]: fetch failed: network timeout"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(e, cause) {
		t.Error("Unwrap() should return cause")
	}
}

func TestExtractorError_NoID(t *testing.T) {
	e := &ExtractorError{Extractor: "Generic", Message: "not found"}
	got := e.Error()
	want := "Generic: not found"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestUnsupportedError(t *testing.T) {
	e := &UnsupportedError{URL: "ftp://example.com/video"}
	got := e.Error()
	want := "unsupported URL: ftp://example.com/video"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestGeoRestrictedError(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		e := &GeoRestrictedError{Message: "not available in your region"}
		got := e.Error()
		want := "geo-restricted: not available in your region"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
	t.Run("without message", func(t *testing.T) {
		e := &GeoRestrictedError{}
		got := e.Error()
		want := "geo-restricted content"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}
