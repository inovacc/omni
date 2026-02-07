package format

import (
	"testing"

	"github.com/inovacc/omni/pkg/video/types"
)

func TestSortFormats(t *testing.T) {
	h480, h720, h1080 := 480, 720, 1080
	formats := []types.Format{
		{FormatID: "720", Height: &h720, Ext: "mp4", VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "480", Height: &h480, Ext: "mp4", VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "1080", Height: &h1080, Ext: "mp4", VCodec: "avc1", ACodec: "mp4a"},
	}

	SortFormats(formats)

	if formats[0].FormatID != "480" {
		t.Errorf("first format should be 480, got %s", formats[0].FormatID)
	}

	if formats[2].FormatID != "1080" {
		t.Errorf("last format should be 1080, got %s", formats[2].FormatID)
	}
}

func TestBestFormat(t *testing.T) {
	h720, h1080 := 720, 1080
	formats := []types.Format{
		{FormatID: "720", Height: &h720, VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "1080", Height: &h1080, VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "1080v", Height: &h1080, VCodec: "avc1", ACodec: ""},
	}
	SortFormats(formats)

	best := BestFormat(formats)
	if best == nil {
		t.Fatal("BestFormat returned nil")
	}
	// Should pick 1080 with both audio and video.
	if best.FormatID != "1080" {
		t.Errorf("BestFormat = %s, want 1080", best.FormatID)
	}
}

func TestWorstFormat(t *testing.T) {
	h480, h1080 := 480, 1080
	formats := []types.Format{
		{FormatID: "480", Height: &h480, VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "1080", Height: &h1080, VCodec: "avc1", ACodec: "mp4a"},
	}
	SortFormats(formats)

	worst := WorstFormat(formats)
	if worst == nil {
		t.Fatal("WorstFormat returned nil")
	}

	if worst.FormatID != "480" {
		t.Errorf("WorstFormat = %s, want 480", worst.FormatID)
	}
}

func TestFilterFormats(t *testing.T) {
	formats := []types.Format{
		{FormatID: "video", VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "audio", VCodec: "none", ACodec: "mp4a"},
		{FormatID: "video2", VCodec: "vp9", ACodec: ""},
	}

	audioOnly := FilterFormats(formats, func(f types.Format) bool {
		return f.HasAudio() && !f.HasVideo()
	})

	if len(audioOnly) != 1 || audioOnly[0].FormatID != "audio" {
		t.Errorf("FilterFormats audio-only = %v", audioOnly)
	}
}
