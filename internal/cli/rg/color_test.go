package rg

import (
	"regexp"
	"strings"
	"testing"
)

func TestColorize(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		color    string
		expected string
	}{
		{
			name:     "no color",
			text:     "hello",
			color:    "",
			expected: "hello",
		},
		{
			name:     "red color",
			text:     "error",
			color:    FgRed,
			expected: FgRed + "error" + Reset,
		},
		{
			name:     "bold magenta",
			text:     "path",
			color:    FgMagenta + Bold,
			expected: FgMagenta + Bold + "path" + Reset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Colorize(tt.text, tt.color)
			if result != tt.expected {
				t.Errorf("Colorize(%q, %q) = %q, want %q", tt.text, tt.color, result, tt.expected)
			}
		})
	}
}

func TestHighlightMatches(t *testing.T) {
	scheme := DefaultScheme()

	tests := []struct {
		name     string
		line     string
		pattern  string
		useColor bool
		wantHas  string
	}{
		{
			name:     "no color",
			line:     "hello world",
			pattern:  "world",
			useColor: false,
			wantHas:  "hello world",
		},
		{
			name:     "single match with color",
			line:     "hello world",
			pattern:  "world",
			useColor: true,
			wantHas:  scheme.Match + "world" + Reset,
		},
		{
			name:     "multiple matches",
			line:     "foo bar foo",
			pattern:  "foo",
			useColor: true,
			wantHas:  scheme.Match + "foo" + Reset,
		},
		{
			name:     "no match",
			line:     "hello world",
			pattern:  "xyz",
			useColor: true,
			wantHas:  "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(tt.pattern)

			result := HighlightMatches(tt.line, re, scheme, tt.useColor)
			if !strings.Contains(result, tt.wantHas) {
				t.Errorf("HighlightMatches() = %q, want to contain %q", result, tt.wantHas)
			}
		})
	}
}

func TestHighlightLiteralMatches(t *testing.T) {
	scheme := DefaultScheme()

	tests := []struct {
		name            string
		line            string
		pattern         string
		caseInsensitive bool
		useColor        bool
		wantHas         string
	}{
		{
			name:            "literal match",
			line:            "hello world",
			pattern:         "world",
			caseInsensitive: false,
			useColor:        true,
			wantHas:         scheme.Match + "world" + Reset,
		},
		{
			name:            "case insensitive match",
			line:            "Hello World",
			pattern:         "world",
			caseInsensitive: true,
			useColor:        true,
			wantHas:         scheme.Match + "World" + Reset,
		},
		{
			name:            "regex metachar treated literally",
			line:            "foo.bar",
			pattern:         ".",
			caseInsensitive: false,
			useColor:        true,
			wantHas:         scheme.Match + "." + Reset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightLiteralMatches(tt.line, tt.pattern, tt.caseInsensitive, scheme, tt.useColor)
			if !strings.Contains(result, tt.wantHas) {
				t.Errorf("HighlightLiteralMatches() = %q, want to contain %q", result, tt.wantHas)
			}
		})
	}
}

func TestParseColorMode(t *testing.T) {
	tests := []struct {
		input    string
		expected ColorMode
	}{
		{"auto", ColorAuto},
		{"always", ColorAlways},
		{"never", ColorNever},
		{"AUTO", ColorAuto},
		{"ALWAYS", ColorAlways},
		{"NEVER", ColorNever},
		{"invalid", ColorAuto}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseColorMode(tt.input)
			if result != tt.expected {
				t.Errorf("ParseColorMode(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseColorSpec(t *testing.T) {
	tests := []struct {
		spec      string
		component string
		attr      string
		value     string
		wantErr   bool
	}{
		{"path:fg:magenta", "path", "fg", "magenta", false},
		{"match:style:bold", "match", "style", "bold", false},
		{"line:fg:green", "line", "fg", "green", false},
		{"column:none", "column", "none", "", false},
		{"invalid", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			component, attr, value, err := ParseColorSpec(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseColorSpec(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
				return
			}

			if err == nil {
				if component != tt.component || attr != tt.attr || value != tt.value {
					t.Errorf("ParseColorSpec(%q) = (%q, %q, %q), want (%q, %q, %q)",
						tt.spec, component, attr, value, tt.component, tt.attr, tt.value)
				}
			}
		})
	}
}

func TestColorNameToCode(t *testing.T) {
	tests := []struct {
		name     string
		isBg     bool
		expected string
	}{
		{"red", false, "\033[31m"},
		{"red", true, "\033[41m"},
		{"green", false, "\033[32m"},
		{"blue", false, "\033[34m"},
		{"unknown", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorNameToCode(tt.name, tt.isBg)
			if result != tt.expected {
				t.Errorf("ColorNameToCode(%q, %v) = %q, want %q", tt.name, tt.isBg, result, tt.expected)
			}
		})
	}
}

func TestApplyColorSpec(t *testing.T) {
	scheme := DefaultScheme()

	err := ApplyColorSpec(&scheme, "path:fg:red")
	if err != nil {
		t.Errorf("ApplyColorSpec() error = %v", err)
	}

	if scheme.Path != "\033[31m" {
		t.Errorf("ApplyColorSpec() path = %q, want red", scheme.Path)
	}

	err = ApplyColorSpec(&scheme, "match:style:bold")
	if err != nil {
		t.Errorf("ApplyColorSpec() error = %v", err)
	}

	if scheme.Match != Bold {
		t.Errorf("ApplyColorSpec() match = %q, want bold", scheme.Match)
	}
}

func TestDefaultScheme(t *testing.T) {
	scheme := DefaultScheme()

	// Verify default colors match ripgrep
	if scheme.Path == "" {
		t.Error("DefaultScheme() path should not be empty")
	}

	if scheme.Line == "" {
		t.Error("DefaultScheme() line should not be empty")
	}

	if scheme.Match == "" {
		t.Error("DefaultScheme() match should not be empty")
	}
}

func TestNoColorScheme(t *testing.T) {
	scheme := NoColorScheme()

	if scheme.Path != "" {
		t.Errorf("NoColorScheme() path should be empty, got %q", scheme.Path)
	}

	if scheme.Line != "" {
		t.Errorf("NoColorScheme() line should be empty, got %q", scheme.Line)
	}

	if scheme.Match != "" {
		t.Errorf("NoColorScheme() match should be empty, got %q", scheme.Match)
	}

	if scheme.Separator != "" {
		t.Errorf("NoColorScheme() separator should be empty, got %q", scheme.Separator)
	}
}

func TestShouldUseColor(t *testing.T) {
	tests := []struct {
		name     string
		mode     ColorMode
		expected bool
	}{
		{"always", ColorAlways, true},
		{"never", ColorNever, false},
		// ColorAuto depends on terminal, skip in unit tests
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldUseColor(tt.mode)
			if result != tt.expected {
				t.Errorf("ShouldUseColor(%v) = %v, want %v", tt.mode, result, tt.expected)
			}
		})
	}
}

func TestFormatLineNumber(t *testing.T) {
	scheme := DefaultScheme()

	t.Run("no color", func(t *testing.T) {
		result := FormatLineNumber(42, scheme, false)
		if result != "42" {
			t.Errorf("FormatLineNumber() = %q, want '42'", result)
		}
	})

	t.Run("with color", func(t *testing.T) {
		result := FormatLineNumber(42, scheme, true)
		if !strings.Contains(result, "42") {
			t.Errorf("FormatLineNumber() should contain '42', got %q", result)
		}

		if !strings.Contains(result, scheme.Line) {
			t.Errorf("FormatLineNumber() should contain color code, got %q", result)
		}
	})

	t.Run("empty scheme", func(t *testing.T) {
		result := FormatLineNumber(42, NoColorScheme(), true)
		if result != "42" {
			t.Errorf("FormatLineNumber() with empty scheme = %q, want '42'", result)
		}
	})
}

func TestFormatColumn(t *testing.T) {
	scheme := DefaultScheme()

	t.Run("no color", func(t *testing.T) {
		result := FormatColumn(5, scheme, false)
		if result != "5" {
			t.Errorf("FormatColumn() = %q, want '5'", result)
		}
	})

	t.Run("with color", func(t *testing.T) {
		result := FormatColumn(5, scheme, true)
		if !strings.Contains(result, "5") {
			t.Errorf("FormatColumn() should contain '5', got %q", result)
		}

		if !strings.Contains(result, scheme.Column) {
			t.Errorf("FormatColumn() should contain color code, got %q", result)
		}
	})
}

func TestFormatPath(t *testing.T) {
	scheme := DefaultScheme()

	t.Run("no color", func(t *testing.T) {
		result := FormatPath("src/main.go", scheme, false)
		if result != "src/main.go" {
			t.Errorf("FormatPath() = %q, want 'src/main.go'", result)
		}
	})

	t.Run("with color", func(t *testing.T) {
		result := FormatPath("src/main.go", scheme, true)
		if !strings.Contains(result, "src/main.go") {
			t.Errorf("FormatPath() should contain path, got %q", result)
		}

		if !strings.Contains(result, scheme.Path) {
			t.Errorf("FormatPath() should contain color code, got %q", result)
		}
	})
}

func TestFormatSeparator(t *testing.T) {
	scheme := DefaultScheme()

	t.Run("no color", func(t *testing.T) {
		result := FormatSeparator(":", scheme, false)
		if result != ":" {
			t.Errorf("FormatSeparator() = %q, want ':'", result)
		}
	})

	t.Run("with color", func(t *testing.T) {
		result := FormatSeparator(":", scheme, true)
		if !strings.Contains(result, ":") {
			t.Errorf("FormatSeparator() should contain ':', got %q", result)
		}
	})

	t.Run("context separator", func(t *testing.T) {
		result := FormatSeparator("-", scheme, false)
		if result != "-" {
			t.Errorf("FormatSeparator() = %q, want '-'", result)
		}
	})
}

func TestApplyColorSpecErrors(t *testing.T) {
	scheme := DefaultScheme()

	t.Run("unknown attribute", func(t *testing.T) {
		err := ApplyColorSpec(&scheme, "path:unknown:value")
		if err == nil {
			t.Error("ApplyColorSpec() should error on unknown attribute")
		}
	})

	t.Run("unknown component", func(t *testing.T) {
		err := ApplyColorSpec(&scheme, "unknown:fg:red")
		if err == nil {
			t.Error("ApplyColorSpec() should error on unknown component")
		}
	})

	t.Run("invalid spec", func(t *testing.T) {
		err := ApplyColorSpec(&scheme, "invalid")
		if err == nil {
			t.Error("ApplyColorSpec() should error on invalid spec")
		}
	})
}

func TestApplyColorSpecComponents(t *testing.T) {
	t.Run("line component", func(t *testing.T) {
		scheme := DefaultScheme()

		err := ApplyColorSpec(&scheme, "line:fg:blue")
		if err != nil {
			t.Fatalf("ApplyColorSpec() error = %v", err)
		}

		if scheme.Line != "\033[34m" {
			t.Errorf("ApplyColorSpec() line = %q, want blue", scheme.Line)
		}
	})

	t.Run("column component", func(t *testing.T) {
		scheme := DefaultScheme()

		err := ApplyColorSpec(&scheme, "column:fg:yellow")
		if err != nil {
			t.Fatalf("ApplyColorSpec() error = %v", err)
		}

		if scheme.Column != "\033[33m" {
			t.Errorf("ApplyColorSpec() column = %q, want yellow", scheme.Column)
		}
	})

	t.Run("background color", func(t *testing.T) {
		scheme := DefaultScheme()

		err := ApplyColorSpec(&scheme, "match:bg:red")
		if err != nil {
			t.Fatalf("ApplyColorSpec() error = %v", err)
		}

		if scheme.Match != "\033[41m" {
			t.Errorf("ApplyColorSpec() match bg = %q, want red bg", scheme.Match)
		}
	})

	t.Run("style underline", func(t *testing.T) {
		scheme := DefaultScheme()

		err := ApplyColorSpec(&scheme, "path:style:underline")
		if err != nil {
			t.Fatalf("ApplyColorSpec() error = %v", err)
		}

		if scheme.Path != Underline {
			t.Errorf("ApplyColorSpec() path = %q, want underline", scheme.Path)
		}
	})

	t.Run("style nobold", func(t *testing.T) {
		scheme := DefaultScheme()

		err := ApplyColorSpec(&scheme, "match:style:nobold")
		if err != nil {
			t.Fatalf("ApplyColorSpec() error = %v", err)
		}

		if scheme.Match != "" {
			t.Errorf("ApplyColorSpec() match nobold = %q, want empty", scheme.Match)
		}
	})

	t.Run("none attribute", func(t *testing.T) {
		scheme := DefaultScheme()

		err := ApplyColorSpec(&scheme, "path:none")
		if err != nil {
			t.Fatalf("ApplyColorSpec() error = %v", err)
		}

		if scheme.Path != "" {
			t.Errorf("ApplyColorSpec() path none = %q, want empty", scheme.Path)
		}
	})
}

func TestColorNameToCodeAllColors(t *testing.T) {
	colors := []struct {
		name   string
		fgCode string
		bgCode string
	}{
		{"black", "\033[30m", "\033[40m"},
		{"red", "\033[31m", "\033[41m"},
		{"green", "\033[32m", "\033[42m"},
		{"yellow", "\033[33m", "\033[43m"},
		{"blue", "\033[34m", "\033[44m"},
		{"magenta", "\033[35m", "\033[45m"},
		{"cyan", "\033[36m", "\033[46m"},
		{"white", "\033[37m", "\033[47m"},
	}

	for _, c := range colors {
		t.Run(c.name+"_fg", func(t *testing.T) {
			result := ColorNameToCode(c.name, false)
			if result != c.fgCode {
				t.Errorf("ColorNameToCode(%q, false) = %q, want %q", c.name, result, c.fgCode)
			}
		})

		t.Run(c.name+"_bg", func(t *testing.T) {
			result := ColorNameToCode(c.name, true)
			if result != c.bgCode {
				t.Errorf("ColorNameToCode(%q, true) = %q, want %q", c.name, result, c.bgCode)
			}
		})
	}
}

func TestHighlightMatchesEdgeCases(t *testing.T) {
	scheme := DefaultScheme()

	t.Run("nil regex", func(t *testing.T) {
		result := HighlightMatches("hello world", nil, scheme, true)
		if result != "hello world" {
			t.Errorf("HighlightMatches() with nil regex = %q, want original", result)
		}
	})

	t.Run("empty line", func(t *testing.T) {
		re := regexp.MustCompile("foo")

		result := HighlightMatches("", re, scheme, true)
		if result != "" {
			t.Errorf("HighlightMatches() empty line = %q, want empty", result)
		}
	})

	t.Run("match at start", func(t *testing.T) {
		re := regexp.MustCompile("hello")

		result := HighlightMatches("hello world", re, scheme, true)
		if !strings.HasPrefix(result, scheme.Match) {
			t.Errorf("HighlightMatches() match at start should start with color")
		}
	})

	t.Run("match at end", func(t *testing.T) {
		re := regexp.MustCompile("world")

		result := HighlightMatches("hello world", re, scheme, true)
		if !strings.HasSuffix(result, Reset) {
			t.Errorf("HighlightMatches() match at end should end with reset")
		}
	})
}

func TestHighlightLiteralMatchesEdgeCases(t *testing.T) {
	scheme := DefaultScheme()

	t.Run("empty pattern", func(t *testing.T) {
		result := HighlightLiteralMatches("hello", "", false, scheme, true)
		if result != "hello" {
			t.Errorf("HighlightLiteralMatches() empty pattern = %q, want original", result)
		}
	})

	t.Run("no color", func(t *testing.T) {
		result := HighlightLiteralMatches("hello world", "world", false, scheme, false)
		if result != "hello world" {
			t.Errorf("HighlightLiteralMatches() no color = %q, want original", result)
		}
	})

	t.Run("overlapping patterns don't exist", func(t *testing.T) {
		// Pattern "aa" in "aaa" should find 1 match (non-overlapping)
		result := HighlightLiteralMatches("aaa", "aa", false, scheme, true)
		// Should highlight first "aa" leaving one "a"
		matchCount := strings.Count(result, scheme.Match)
		if matchCount != 1 {
			t.Errorf("HighlightLiteralMatches() should find 1 non-overlapping match, got %d", matchCount)
		}
	})
}
