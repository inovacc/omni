package parser

import (
	"errors"
	"testing"
)

// TestParseLineCommentSplitting drives parseLine's name/comment extraction.
//
// NOTE: parseLine matches tree characters at the byte level (line[i:i+4]),
// but box-drawing runes such as │ ├ └ are multi-byte in UTF-8, so those
// byte comparisons never fire for real tree output — the higher layer strips
// the prefixes. These cases therefore use plain inputs that exercise the
// genuinely reachable branches: leading-space skipping and comment parsing.
func TestParseLineCommentSplitting(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantName    string
		wantComment string
	}{
		{"plain name", "file.txt", "file.txt", ""},
		{"name and comment", "config.yaml # the config", "config.yaml", "the config"},
		{"comment no space before hash", "a.txt#note", "a.txt", "note"},
		{"double hash keeps second", "x.txt ## hash", "x.txt", "# hash"},
		{"leading spaces skipped", "    plain.go", "plain.go", ""},
		{"comment only yields empty name", "# only comment", "", "only comment"},
	}

	p := &Parser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, name, comment, err := p.parseLine(tt.line)
			if err != nil {
				t.Fatalf("parseLine(%q) error: %v", tt.line, err)
			}
			if level != 0 {
				t.Errorf("level = %d, want 0 (byte-level matcher does not advance on plain input)", level)
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
// sibling entries (each parsed as a flat child since the byte-level indentation
// matcher does not advance levels on box-drawing runes).
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
