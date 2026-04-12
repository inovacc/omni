package format

import (
	"testing"

	"github.com/inovacc/omni/pkg/video/types"
)

func TestSelectorBest(t *testing.T) {
	h480, h720 := 480, 720
	formats := []types.Format{
		{FormatID: "480", Height: &h480, VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "720", Height: &h720, VCodec: "avc1", ACodec: "mp4a"},
	}

	sel := NewSelector("best")

	result, err := sel.Select(formats)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 || result[0].FormatID != "720" {
		t.Errorf("best = %s, want 720", result[0].FormatID)
	}
}

func TestSelectorWorst(t *testing.T) {
	h480, h720 := 480, 720
	formats := []types.Format{
		{FormatID: "480", Height: &h480, VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "720", Height: &h720, VCodec: "avc1", ACodec: "mp4a"},
	}

	sel := NewSelector("worst")

	result, err := sel.Select(formats)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 || result[0].FormatID != "480" {
		t.Errorf("worst = %s, want 480", result[0].FormatID)
	}
}

func TestSelectorFormatID(t *testing.T) {
	formats := []types.Format{
		{FormatID: "18", VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "22", VCodec: "avc1", ACodec: "mp4a"},
	}

	sel := NewSelector("18")

	result, err := sel.Select(formats)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 || result[0].FormatID != "18" {
		t.Errorf("format_id = %s, want 18", result[0].FormatID)
	}
}

func TestSelectorFilter(t *testing.T) {
	h480, h720, h1080 := 480, 720, 1080
	formats := []types.Format{
		{FormatID: "480", Height: &h480, VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "720", Height: &h720, VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "1080", Height: &h1080, VCodec: "avc1", ACodec: "mp4a"},
	}

	sel := NewSelector("best[height<=720]")

	result, err := sel.Select(formats)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 || result[0].FormatID != "720" {
		t.Errorf("best[height<=720] = %s, want 720", result[0].FormatID)
	}
}

func TestSelectorEmpty(t *testing.T) {
	sel := NewSelector("best")

	_, err := sel.Select(nil)
	if err == nil {
		t.Error("expected error for empty formats")
	}
}

func TestSelectorBestVideo(t *testing.T) {
	h720 := 720
	formats := []types.Format{
		{FormatID: "v720", Height: &h720, VCodec: "avc1", ACodec: ""},
		{FormatID: "audio", VCodec: "none", ACodec: "mp4a"},
	}
	result, err := NewSelector("bestvideo").Select(formats)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].FormatID != "v720" {
		t.Errorf("bestvideo = %v, want v720", result)
	}
}

func TestSelectorWorstVideo(t *testing.T) {
	h480, h720 := 480, 720
	formats := []types.Format{
		{FormatID: "v720", Height: &h720, VCodec: "avc1", ACodec: ""},
		{FormatID: "v480", Height: &h480, VCodec: "avc1", ACodec: ""},
	}
	result, err := NewSelector("worstvideo").Select(formats)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].FormatID != "v480" {
		t.Errorf("worstvideo = %v, want v480", result)
	}
}

func TestSelectorBestAudio(t *testing.T) {
	formats := []types.Format{
		{FormatID: "audio", VCodec: "none", ACodec: "mp4a"},
		{FormatID: "video", VCodec: "avc1", ACodec: ""},
	}
	result, err := NewSelector("bestaudio").Select(formats)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].FormatID != "audio" {
		t.Errorf("bestaudio = %v, want audio", result)
	}
}

func TestSelectorWorstAudio(t *testing.T) {
	formats := []types.Format{
		{FormatID: "audio", VCodec: "none", ACodec: "mp4a"},
		{FormatID: "video", VCodec: "avc1", ACodec: ""},
	}
	result, err := NewSelector("worstaudio").Select(formats)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].FormatID != "audio" {
		t.Errorf("worstaudio = %v, want audio", result)
	}
}

func TestSelectorNoVideoFormats(t *testing.T) {
	formats := []types.Format{
		{FormatID: "audio", VCodec: "none", ACodec: "mp4a"},
	}
	_, err := NewSelector("bestvideo").Select(formats)
	if err == nil {
		t.Error("expected error for bestvideo with no video formats")
	}
}

func TestSelectorNoAudioFormats(t *testing.T) {
	formats := []types.Format{
		{FormatID: "video", VCodec: "avc1", ACodec: ""},
	}
	_, err := NewSelector("bestaudio").Select(formats)
	if err == nil {
		t.Error("expected error for bestaudio with no audio formats")
	}
}

func TestSelectorUnknownID(t *testing.T) {
	formats := []types.Format{
		{FormatID: "18", VCodec: "avc1", ACodec: "mp4a"},
	}
	_, err := NewSelector("999").Select(formats)
	if err == nil {
		t.Error("expected error for unknown format ID")
	}
}

func TestSelectorStringFilter(t *testing.T) {
	h720 := 720
	formats := []types.Format{
		{FormatID: "mp4", Height: &h720, Ext: "mp4", VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "webm", Height: &h720, Ext: "webm", VCodec: "vp9", ACodec: "opus"},
	}
	result, err := NewSelector("best[ext=mp4]").Select(formats)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].FormatID != "mp4" {
		t.Errorf("best[ext=mp4] = %v, want mp4", result)
	}
}

func TestSelectorStringFilterNegate(t *testing.T) {
	h720 := 720
	formats := []types.Format{
		{FormatID: "mp4", Height: &h720, Ext: "mp4", VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "webm", Height: &h720, Ext: "webm", VCodec: "vp9", ACodec: "opus"},
	}
	result, err := NewSelector("best[ext!=mp4]").Select(formats)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].FormatID != "webm" {
		t.Errorf("best[ext!=mp4] = %v, want webm", result)
	}
}

func TestSelectorDefaultEmpty(t *testing.T) {
	formats := []types.Format{
		{FormatID: "18", VCodec: "avc1", ACodec: "mp4a"},
	}
	// Empty spec defaults to "best"
	result, err := NewSelector("").Select(formats)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Errorf("default selector result = %v", result)
	}
}

func TestSelectorHeightGTE(t *testing.T) {
	h480, h720, h1080 := 480, 720, 1080
	formats := []types.Format{
		{FormatID: "480", Height: &h480, VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "720", Height: &h720, VCodec: "avc1", ACodec: "mp4a"},
		{FormatID: "1080", Height: &h1080, VCodec: "avc1", ACodec: "mp4a"},
	}
	result, err := NewSelector("best[height>=720]").Select(formats)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].FormatID != "1080" {
		t.Errorf("best[height>=720] = %v, want 1080", result)
	}
}

func TestSelectorNoMatchingFilter(t *testing.T) {
	h480 := 480
	formats := []types.Format{
		{FormatID: "480", Height: &h480, VCodec: "avc1", ACodec: "mp4a"},
	}
	_, err := NewSelector("best[height>=1080]").Select(formats)
	if err == nil {
		t.Error("expected error for filter matching no formats")
	}
}

func TestSelectorWorstVideoNoVideo(t *testing.T) {
	formats := []types.Format{
		{FormatID: "audio", VCodec: "none", ACodec: "mp4a"},
	}
	_, err := NewSelector("worstvideo").Select(formats)
	if err == nil {
		t.Error("expected error for worstvideo with no video formats")
	}
}

func TestSelectorWorstAudioNoAudio(t *testing.T) {
	formats := []types.Format{
		{FormatID: "video", VCodec: "avc1", ACodec: ""},
	}
	_, err := NewSelector("worstaudio").Select(formats)
	if err == nil {
		t.Error("expected error for worstaudio with no audio formats")
	}
}
