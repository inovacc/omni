package parser

import (
	"errors"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/twig/models"
)

func TestParseString_FlatTree(t *testing.T) {
	// The parser strips the Unicode tree connector prefixes (├── and └──) from
	// node names. parseLine compares the multi-byte UTF-8 box-drawing glyphs with
	// strings.HasPrefix, so the connectors are recognized and removed.
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

	// Connector prefixes are stripped: names are clean.
	wantNames := []string{"README.md", "src", "go.mod"}
	for i, want := range wantNames {
		if got := root.Children[i].Name; got != want {
			t.Errorf("child[%d] name = %q, want %q", i, got, want)
		}
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

// flatNode is a depth-first flattening of a parsed tree used to pin the exact
// name, level, dir flag, and path of every node, including deep nesting.
type flatNode struct {
	name  string
	level int
	isDir bool
	path  string
}

func flatten(node *models.Node, out *[]flatNode) {
	*out = append(*out, flatNode{
		name:  node.Name,
		level: node.Level,
		isDir: node.IsDir,
		path:  node.Path,
	})
	for _, child := range node.Children {
		flatten(child, out)
	}
}

// TestParseString_NestedTree pins the corrected indentation handling: the
// box-drawing connectors are stripped from names and the "│   "/"    " indent
// units drive nesting so that deeply nested nodes attach to the right parent.
//
// Before the byte-vs-rune fix, parseLine's fixed-width byte windows never matched
// the multi-byte glyphs, so every line reported level 0, names kept their
// connector prefixes, and grandchildren were flattened into siblings of their
// parent's level. This test fails on that broken behavior.
func TestParseString_NestedTree(t *testing.T) {
	input := "project/\n" +
		"├── src/\n" +
		"│   ├── main.go\n" +
		"│   └── util/\n" +
		"│       └── helper.go\n" +
		"└── README.md"

	p := NewParser()
	root, err := p.ParseString(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []flatNode
	flatten(root, &got)

	want := []flatNode{
		{name: "project", level: 0, isDir: true, path: "project"},
		{name: "src", level: 1, isDir: true, path: "project/src"},
		{name: "main.go", level: 2, isDir: false, path: "project/src/main.go"},
		{name: "util", level: 2, isDir: true, path: "project/src/util"},
		{name: "helper.go", level: 3, isDir: false, path: "project/src/util/helper.go"},
		{name: "README.md", level: 1, isDir: false, path: "project/README.md"},
	}

	if len(got) != len(want) {
		t.Fatalf("node count = %d, want %d\ngot: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("node[%d] = %+v, want %+v", i, got[i], w)
		}
	}
}

// TestParseLine_LevelAndName checks parseLine directly across the indent unit and
// connector variants, with and without trailing spaces.
func TestParseLine_LevelAndName(t *testing.T) {
	p := &Parser{}

	tests := []struct {
		name      string
		line      string
		wantLevel int
		wantName  string
		wantComm  string
	}{
		{"mid connector, no indent", "├── README.md", 0, "README.md", ""},
		{"last connector, no indent", "└── go.mod", 0, "go.mod", ""},
		{"one pipe indent", "│   └── main.go", 1, "main.go", ""},
		{"blank indent unit", "    └── helper.go", 1, "helper.go", ""},
		{"two indent units", "│       └── deep.go", 2, "deep.go", ""},
		{"connector without trailing space", "├──config.yaml", 0, "config.yaml", ""},
		{"with comment", "├── main.go # entry", 0, "main.go", "entry"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, name, comment, err := p.parseLine(tt.line)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if level != tt.wantLevel {
				t.Errorf("level = %d, want %d", level, tt.wantLevel)
			}
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
			if comment != tt.wantComm {
				t.Errorf("comment = %q, want %q", comment, tt.wantComm)
			}
		})
	}
}
