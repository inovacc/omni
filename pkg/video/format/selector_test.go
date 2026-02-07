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
