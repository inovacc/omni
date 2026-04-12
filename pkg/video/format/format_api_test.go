package format_test

import (
	"testing"

	"github.com/inovacc/omni/pkg/video/format"
	"github.com/inovacc/omni/pkg/video/types"
)

func TestSortFormats_API(t *testing.T) {
	h360 := 360
	h720 := 720

	formats := []types.Format{
		{FormatID: "low", Ext: "mp4", Height: &h360},
		{FormatID: "high", Ext: "mp4", Height: &h720},
	}

	format.SortFormats(formats)

	// After sort, best quality (highest height) should be last.
	if formats[len(formats)-1].FormatID != "high" {
		t.Errorf("SortFormats: expected 'high' last, got %q", formats[len(formats)-1].FormatID)
	}
}

func TestSelectFormat_API(t *testing.T) {
	h720 := 720

	formats := []types.Format{
		{FormatID: "f1", Ext: "mp4", Height: &h720, URL: "http://example.com/f1.mp4"},
	}

	format.SortFormats(formats)

	sel := format.NewSelector("best")
	result, err := sel.Select(formats)
	if err != nil {
		t.Fatalf("NewSelector('best').Select() error = %v", err)
	}
	if len(result) == 0 {
		t.Fatal("NewSelector('best').Select() returned empty slice for non-empty formats")
	}
}
