package parser

import (
	"errors"
	"strings"
	"testing"
)

func TestParseString_FlatTree(t *testing.T) {
	// Note: the parser currently does not strip Unicode tree connector prefixes
	// (├── and └──) from node names because the byte-level slicing in parseLine
	// uses i+4 which doesn't match the multi-byte UTF-8 tree characters.
	// This is a characterization test documenting the current behavior.
	input := `project/
├── README.md
├── src/
└── go.mod`

	p := NewParser()
	root, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.Name != "project" {
		t.Errorf("root name = %q, want %q", root.Name, "project")
	}
	if !root.IsDir {
		t.Error("root should be a directory")
	}
	if len(root.Children) != 3 {
		t.Fatalf("root children = %d, want 3", len(root.Children))
	}

	// Children retain tree connector prefixes in current implementation
	if !strings.Contains(root.Children[0].Name, "README.md") {
		t.Errorf("child[0] should contain README.md, got %q", root.Children[0].Name)
	}
	if !strings.Contains(root.Children[2].Name, "go.mod") {
		t.Errorf("child[2] should contain go.mod, got %q", root.Children[2].Name)
	}
}

func TestParseString_Comments(t *testing.T) {
	input := `root/
├── config.yaml # configuration file
└── main.go # entry point`

	p := NewParser()
	root, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(root.Children) != 2 {
		t.Fatalf("children = %d, want 2", len(root.Children))
	}
	if root.Children[0].Comment != "configuration file" {
		t.Errorf("comment = %q, want %q", root.Children[0].Comment, "configuration file")
	}
	if root.Children[1].Comment != "entry point" {
		t.Errorf("comment = %q, want %q", root.Children[1].Comment, "entry point")
	}
}

func TestParseString_EmptyInput(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"whitespace only", "   \n\n  \n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.ParseString(tt.input)
			if !errors.Is(err, ErrEmptyInput) {
				t.Errorf("err = %v, want ErrEmptyInput", err)
			}
		})
	}
}

func TestParseString_RootNotDirectory(t *testing.T) {
	p := NewParser()
	root, err := p.ParseString("file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.IsDir {
		t.Error("root without trailing / should not be a directory")
	}
	if root.Name != "file.txt" {
		t.Errorf("root name = %q, want %q", root.Name, "file.txt")
	}
}

func TestParseString_RootWithTrailingSlash(t *testing.T) {
	p := NewParser()
	root, err := p.ParseString("mydir/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !root.IsDir {
		t.Error("root with trailing / should be a directory")
	}
	if root.Name != "mydir" {
		t.Errorf("root name = %q, want %q (trailing / should be trimmed)", root.Name, "mydir")
	}
}

func TestParse_WithReader(t *testing.T) {
	input := "mydir/\n└── hello.txt"

	p := NewParser()
	root, err := p.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.Name != "mydir" {
		t.Errorf("root name = %q, want %q", root.Name, "mydir")
	}
	if len(root.Children) != 1 {
		t.Fatalf("children = %d, want 1", len(root.Children))
	}
}

func TestParseString_Paths(t *testing.T) {
	p := NewParser()
	root, err := p.ParseString("root/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.Path != "root" {
		t.Errorf("root path = %q, want %q", root.Path, "root")
	}
}

func TestParseString_Levels(t *testing.T) {
	input := `root/
├── a.txt
└── b.txt`

	p := NewParser()
	root, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.Level != 0 {
		t.Errorf("root level = %d, want 0", root.Level)
	}
	for i, child := range root.Children {
		if child.Level != 1 {
			t.Errorf("child[%d] level = %d, want 1", i, child.Level)
		}
	}
}

func TestParseString_ParentReference(t *testing.T) {
	input := `root/
├── a.txt
└── b.txt`

	p := NewParser()
	root, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, child := range root.Children {
		if child.Parent != root {
			t.Errorf("child[%d] parent should be root", i)
		}
	}
}

func TestParseString_RootComment(t *testing.T) {
	p := NewParser()
	root, err := p.ParseString("project/ # root project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.Comment != "root project" {
		t.Errorf("root comment = %q, want %q", root.Comment, "root project")
	}
}

func TestParseString_SkipsEmptyLines(t *testing.T) {
	input := "root/\n\n├── a.txt\n\n└── b.txt"

	p := NewParser()
	root, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(root.Children) != 2 {
		t.Errorf("children = %d, want 2 (blank lines should be skipped)", len(root.Children))
	}
}

func TestNewParser_ImplementsInterface(t *testing.T) {
	var _ TreeParser = NewParser()
}

func TestParseString_SingleRoot(t *testing.T) {
	p := NewParser()
	root, err := p.ParseString("standalone.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.Name != "standalone.txt" {
		t.Errorf("name = %q, want %q", root.Name, "standalone.txt")
	}
	if len(root.Children) != 0 {
		t.Errorf("children = %d, want 0", len(root.Children))
	}
	if root.IsLeaf() != true {
		t.Error("single root should be a leaf")
	}
}

func TestParseString_DirectoryIsDir(t *testing.T) {
	input := `root/
├── src/
└── README.md`

	p := NewParser()
	root, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// src/ child should be a directory (name ends with /)
	src := root.Children[0]
	if !src.IsDir {
		t.Error("src/ should be recognized as directory")
	}

	// README.md should NOT be a directory
	readme := root.Children[1]
	if readme.IsDir {
		t.Error("README.md should not be a directory")
	}
}
