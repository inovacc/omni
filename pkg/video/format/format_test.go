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

func TestBestFormat_Empty(t *testing.T) {
	if f := BestFormat(nil); f != nil {
		t.Errorf("BestFormat(nil) = %v, want nil", f)
	}
}

func TestWorstFormat_Empty(t *testing.T) {
	if f := WorstFormat(nil); f != nil {
		t.Errorf("WorstFormat(nil) = %v, want nil", f)
	}
}

func TestWorstFormat_NoAV(t *testing.T) {
	// All video-only — should fall back to first
	h480 := 480
	formats := []types.Format{
		{FormatID: "video-only", Height: &h480, VCodec: "avc1", ACodec: ""},
	}
	SortFormats(formats)
	f := WorstFormat(formats)
	if f == nil {
		t.Fatal("WorstFormat returned nil")
	}
	if f.FormatID != "video-only" {
		t.Errorf("WorstFormat = %s, want video-only", f.FormatID)
	}
}

func TestBestVideoOnly(t *testing.T) {
	h480, h720 := 480, 720
	formats := []types.Format{
		{FormatID: "v480", Height: &h480, VCodec: "avc1", ACodec: ""},
		{FormatID: "v720", Height: &h720, VCodec: "avc1", ACodec: ""},
		{FormatID: "audio", VCodec: "none", ACodec: "mp4a"},
	}
	SortFormats(formats)
	f := BestVideoOnly(formats)
	if f == nil {
		t.Fatal("BestVideoOnly returned nil")
	}
	if f.FormatID != "v720" {
		t.Errorf("BestVideoOnly = %s, want v720", f.FormatID)
	}
}

func TestBestVideoOnly_Empty(t *testing.T) {
	formats := []types.Format{
		{FormatID: "audio", VCodec: "none", ACodec: "mp4a"},
	}
	SortFormats(formats)
	f := BestVideoOnly(formats)
	if f != nil {
		t.Errorf("BestVideoOnly should be nil when no video formats, got %v", f)
	}
}

func TestBestAudioOnly(t *testing.T) {
	h720 := 720
	formats := []types.Format{
		{FormatID: "video-only", Height: &h720, VCodec: "avc1", ACodec: ""},
		{FormatID: "audio-only", VCodec: "none", ACodec: "mp4a"},
	}
	SortFormats(formats)
	f := BestAudioOnly(formats)
	if f == nil {
		t.Fatal("BestAudioOnly returned nil")
	}
	if f.FormatID != "audio-only" {
		t.Errorf("BestAudioOnly = %s, want audio-only", f.FormatID)
	}
}

func TestBestAudioOnly_Empty(t *testing.T) {
	formats := []types.Format{
		{FormatID: "video-only", VCodec: "avc1", ACodec: ""},
	}
	SortFormats(formats)
	f := BestAudioOnly(formats)
	if f != nil {
		t.Errorf("BestAudioOnly should be nil when no audio formats, got %v", f)
	}
}

func TestSortFormats_ExtPreference(t *testing.T) {
	h720 := 720
	formats := []types.Format{
		{FormatID: "webm", Height: &h720, Ext: "webm", VCodec: "vp9", ACodec: "opus"},
		{FormatID: "mp4", Height: &h720, Ext: "mp4", VCodec: "avc1", ACodec: "mp4a"},
	}
	SortFormats(formats)
	// webm (index 3) > mp4 (index 2) so webm should be last (better)
	if formats[len(formats)-1].FormatID != "webm" {
		t.Errorf("webm should rank higher than mp4, got last=%s", formats[len(formats)-1].FormatID)
	}
}

func TestSortFormats_Preference(t *testing.T) {
	pref := -1
	h720 := 720
	formats := []types.Format{
		{FormatID: "normal", Height: &h720, VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "preferred", Height: &h720, Preference: &pref, VCodec: "avc1", ACodec: "mp4a"},
	}
	SortFormats(formats)
	// Negative preference = lower rank = should come first
	if formats[0].FormatID != "preferred" {
		t.Errorf("preferred (pref=-1) should sort first, got %s", formats[0].FormatID)
	}
}
