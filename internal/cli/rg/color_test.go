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
