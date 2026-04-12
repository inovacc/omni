package utils_test

import (
	"testing"

	"github.com/inovacc/omni/pkg/video/utils"
)

func TestSanitizeFilename_API(t *testing.T) {
	tests := []struct {
		input    string
		restrict bool
		wantNone bool
	}{
		{"My Video Title", false, false},
		{"Video: Part 1/2", false, false},
		{"My Video Title", true, false},
	}

	for _, tt := range tests {
		got := utils.SanitizeFilename(tt.input, tt.restrict)
		if tt.wantNone && got != "" {
			t.Errorf("SanitizeFilename(%q, %v) = %q, want empty", tt.input, tt.restrict, got)
		}
		if !tt.wantNone && got == "" {
			t.Errorf("SanitizeFilename(%q, %v) returned empty string", tt.input, tt.restrict)
		}
	}
}

func TestParseDuration_API(t *testing.T) {
	d, ok := utils.ParseDuration("1:30")
	if !ok {
		t.Fatal("ParseDuration('1:30') returned ok=false")
	}
	if d != 90 {
		t.Errorf("ParseDuration('1:30') = %v, want 90", d)
	}
}

func TestURLJoin_API(t *testing.T) {
	got := utils.URLJoin("https://example.com/path/", "../other")
	if got == "" {
		t.Fatal("URLJoin returned empty string")
	}
}
