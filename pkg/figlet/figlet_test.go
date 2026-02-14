package figlet

import (
	"slices"
	"strings"
	"testing"
)

func TestListFonts(t *testing.T) {
	fonts := ListFonts()
	if len(fonts) == 0 {
		t.Fatal("ListFonts() returned empty list")
	}

	expected := []string{"banner", "block", "digital", "lean", "mini", "shadow", "slant", "small", "standard"}
	if len(fonts) != len(expected) {
		t.Errorf("ListFonts() = %d fonts, want %d", len(fonts), len(expected))
	}

	for _, name := range expected {
		found := slices.Contains(fonts, name)

		if !found {
			t.Errorf("ListFonts() missing font %q", name)
		}
	}
}

func TestLoadEmbedded(t *testing.T) {
	for _, name := range ListFonts() {
		t.Run(name, func(t *testing.T) {
			f, err := LoadEmbedded(name)
			if err != nil {
				t.Fatalf("LoadEmbedded(%q) error = %v", name, err)
			}

			if f.Height <= 0 {
				t.Errorf("font %q height = %d, want > 0", name, f.Height)
			}

			if len(f.Characters) < 95 {
				t.Errorf("font %q has %d chars, want >= 95 (printable ASCII)", name, len(f.Characters))
			}
			// Space character must exist
			if _, ok := f.Characters[' ']; !ok {
				t.Errorf("font %q missing space character", name)
			}
			// 'A' must exist
			if _, ok := f.Characters['A']; !ok {
				t.Errorf("font %q missing 'A' character", name)
			}
		})
	}
}

func TestLoadEmbeddedNotFound(t *testing.T) {
	_, err := LoadEmbedded("nonexistent")
	if err == nil {
		t.Error("LoadEmbedded(nonexistent) should return error")
	}
}

func TestRender(t *testing.T) {
	result, err := Render("Hi")
	if err != nil {
		t.Fatalf("Render(Hi) error = %v", err)
	}

	if result == "" {
		t.Error("Render(Hi) returned empty string")
	}

	lines := strings.Split(result, "\n")
	if len(lines) == 0 {
		t.Error("Render(Hi) returned no lines")
	}
}

func TestRenderLines(t *testing.T) {
	lines, err := RenderLines("A")
	if err != nil {
		t.Fatalf("RenderLines(A) error = %v", err)
	}

	f, _ := LoadEmbedded("standard")
	if len(lines) != f.Height {
		t.Errorf("RenderLines(A) = %d lines, want %d", len(lines), f.Height)
	}
}

func TestRenderWithFont(t *testing.T) {
	for _, name := range ListFonts() {
		t.Run(name, func(t *testing.T) {
			result, err := Render("Test", WithFont(name))
			if err != nil {
				t.Fatalf("Render with font %q error = %v", name, err)
			}

			if result == "" {
				t.Errorf("Render with font %q returned empty", name)
			}
		})
	}
}

func TestRenderWithWidth(t *testing.T) {
	lines, err := RenderLines("Hello World", WithWidth(20))
	if err != nil {
		t.Fatalf("RenderLines with width error = %v", err)
	}

	for i, line := range lines {
		if len(line) > 20 {
			t.Errorf("line %d length = %d, want <= 20", i, len(line))
		}
	}
}

func TestRenderEmptyString(t *testing.T) {
	lines, err := RenderLines("")
	if err != nil {
		t.Fatalf("RenderLines('') error = %v", err)
	}

	if lines != nil {
		t.Errorf("RenderLines('') = %v, want nil", lines)
	}
}

func TestRenderUnknownChar(t *testing.T) {
	// Should not error; unknown chars fall back to space
	_, err := Render("Hello \x01 World")
	if err != nil {
		t.Fatalf("Render with unknown char error = %v", err)
	}
}

func TestRenderInvalidFont(t *testing.T) {
	_, err := Render("Hi", WithFont("nosuchfont"))
	if err == nil {
		t.Error("Render with invalid font should return error")
	}
}

func TestRenderWithLoadedFont(t *testing.T) {
	f, err := LoadEmbedded("small")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Render("OK", WithLoadedFont(f))
	if err != nil {
		t.Fatalf("Render with loaded font error = %v", err)
	}

	if result == "" {
		t.Error("Render with loaded font returned empty")
	}
}

func TestParseHeaderErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"not flf", "garbage data"},
		{"too short", "flf2a"},
		{"missing params", "flf2a$ 6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseHeader(tt.input)
			if err == nil {
				t.Error("parseHeader should return error")
			}
		})
	}
}

func TestParseCodeTag(t *testing.T) {
	tests := []struct {
		input string
		want  int
		err   bool
	}{
		{"196", 196, false},
		{"0x00C4", 0xC4, false},
		{"0304", 0304, false},
		{"-1", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseCodeTag(tt.input)
			if tt.err {
				if err == nil {
					t.Error("parseCodeTag should return error")
				}

				return
			}

			if err != nil {
				t.Fatalf("parseCodeTag error = %v", err)
			}

			if got != tt.want {
				t.Errorf("parseCodeTag = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestHardblankReplacement(t *testing.T) {
	// The standard font uses '$' as hardblank
	// After rendering, no '$' from hardblanks should remain
	f, err := LoadEmbedded("standard")
	if err != nil {
		t.Fatal(err)
	}

	lines := renderText(f, " ", 0)
	for i, line := range lines {
		if strings.Contains(line, "$") {
			t.Errorf("line %d still contains hardblank: %q", i, line)
		}
	}
}
