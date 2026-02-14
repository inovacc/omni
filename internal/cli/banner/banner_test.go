package banner

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunBanner(t *testing.T) {
	t.Run("basic text", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunBanner(&buf, strings.NewReader(""), []string{"Hi"}, Options{})
		if err != nil {
			t.Fatalf("RunBanner() error = %v", err)
		}

		output := buf.String()
		if output == "" {
			t.Error("RunBanner() returned empty output")
		}
	})

	t.Run("multiple words", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunBanner(&buf, strings.NewReader(""), []string{"Hello", "World"}, Options{})
		if err != nil {
			t.Fatalf("RunBanner() error = %v", err)
		}

		if buf.String() == "" {
			t.Error("RunBanner() returned empty output")
		}
	})

	t.Run("stdin input", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunBanner(&buf, strings.NewReader("Test\n"), nil, Options{})
		if err != nil {
			t.Fatalf("RunBanner() error = %v", err)
		}

		if buf.String() == "" {
			t.Error("RunBanner() returned empty output")
		}
	})

	t.Run("no text error", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunBanner(&buf, strings.NewReader(""), nil, Options{})
		if err == nil {
			t.Error("RunBanner() should error with no text")
		}
	})
}

func TestRunBannerFonts(t *testing.T) {
	fonts := []string{"standard", "slant", "small", "mini", "banner", "block", "shadow", "lean", "digital"}
	for _, font := range fonts {
		t.Run(font, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunBanner(&buf, strings.NewReader(""), []string{"Test"}, Options{Font: font})
			if err != nil {
				t.Fatalf("RunBanner() with font %q error = %v", font, err)
			}

			if buf.String() == "" {
				t.Errorf("RunBanner() with font %q returned empty", font)
			}
		})
	}
}

func TestRunBannerInvalidFont(t *testing.T) {
	var buf bytes.Buffer

	err := RunBanner(&buf, strings.NewReader(""), []string{"Hi"}, Options{Font: "nosuchfont"})
	if err == nil {
		t.Error("RunBanner() should error with invalid font")
	}
}

func TestRunBannerList(t *testing.T) {
	var buf bytes.Buffer

	err := RunBanner(&buf, strings.NewReader(""), nil, Options{List: true})
	if err != nil {
		t.Fatalf("RunBanner() list error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "standard") {
		t.Error("RunBanner() list should contain 'standard'")
	}

	if !strings.Contains(output, "slant") {
		t.Error("RunBanner() list should contain 'slant'")
	}
}

func TestRunBannerWidth(t *testing.T) {
	var buf bytes.Buffer

	err := RunBanner(&buf, strings.NewReader(""), []string{"Hello"}, Options{Width: 20})
	if err != nil {
		t.Fatalf("RunBanner() width error = %v", err)
	}

	for line := range strings.SplitSeq(buf.String(), "\n") {
		if len(line) > 20 {
			t.Errorf("line exceeds width limit: %d > 20: %q", len(line), line)
		}
	}
}
