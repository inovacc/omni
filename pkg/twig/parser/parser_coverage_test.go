package parser

import (
	"errors"
	"testing"
)

// TestParseLineCommentSplitting drives parseLine's name/comment extraction.
//
// NOTE: the corrected parseLine (plan 025) consumes leading indentation units
// rune-correctly — one indentation unit (a pipe-continuation "│   " or a blank
// "    " of four columns) advances exactly one nesting level — before stripping
// the connector glyph and splitting name/comment. So a four-space leading
// indent now yields level 1, not the old buggy level 0.
func TestParseLineCommentSplitting(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantLevel   int
		wantName    string
		wantComment string
	}{
		{"plain name", "file.txt", 0, "file.txt", ""},
		{"name and comment", "config.yaml # the config", 0, "config.yaml", "the config"},
		{"comment no space before hash", "a.txt#note", 0, "a.txt", "note"},
		{"double hash keeps second", "x.txt ## hash", 0, "x.txt", "# hash"},
		{"leading spaces skipped", "    plain.go", 1, "plain.go", ""},
		{"comment only yields empty name", "# only comment", 0, "", "only comment"},
	}

	p := &Parser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, name, comment, err := p.parseLine(tt.line)
			if err != nil {
				t.Fatalf("parseLine(%q) error: %v", tt.line, err)
			}
			if level != tt.wantLevel {
				t.Errorf("level = %d, want %d", level, tt.wantLevel)
			}
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
			if comment != tt.wantComment {
				t.Errorf("comment = %q, want %q", comment, tt.wantComment)
			}
		})
	}
}

func TestParseLineEmptyName(t *testing.T) {
	p := &Parser{}
	_, _, _, err := p.parseLine("    ")
	if !errors.Is(err, ErrEmptyNodeName) {
		t.Errorf("parseLine(spaces only) err = %v, want ErrEmptyNodeName", err)
	}
}

// TestParseStringDeepFlat exercises the end-to-end Parse path through several
// sibling entries. With no leading indentation before each connector glyph,
// all three entries are direct children of root (level 1 siblings).
func TestParseStringDeepFlat(t *testing.T) {
	input := "root/\n" +
		"├── a.txt\n" +
		"├── b.txt\n" +
		"└── c.txt"

	p := NewParser()
	root, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("ParseString error: %v", err)
	}
	if len(root.Children) != 3 {
		t.Fatalf("root children = %d, want 3", len(root.Children))
	}
}
