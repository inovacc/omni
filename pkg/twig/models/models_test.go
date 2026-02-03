package models

import (
	"encoding/json"
	"testing"
)

func TestNewNode(t *testing.T) {
	tests := []struct {
		name   string
		nName  string
		path   string
		isDir  bool
		wantNn string
		wantD  bool
	}{
		{
			name:   "create directory node",
			nName:  "src",
			path:   "/project/src",
			isDir:  true,
			wantNn: "src",
			wantD:  true,
		},
		{
			name:   "create file node",
			nName:  "main.go",
			path:   "/project/main.go",
			isDir:  false,
			wantNn: "main.go",
			wantD:  false,
		},
		{
			name:   "empty name",
			nName:  "",
			path:   "/project",
			isDir:  true,
			wantNn: "",
			wantD:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewNode(tt.nName, tt.path, tt.isDir)

			if node.Name != tt.wantNn {
				t.Errorf("NewNode().Name = %q, want %q", node.Name, tt.wantNn)
			}

			if node.Path != tt.path {
				t.Errorf("NewNode().Path = %q, want %q", node.Path, tt.path)
			}

			if node.IsDir != tt.wantD {
				t.Errorf("NewNode().IsDir = %v, want %v", node.IsDir, tt.wantD)
			}

			if node.Children == nil {
				t.Error("NewNode().Children should not be nil")
			}

			if len(node.Children) != 0 {
				t.Errorf("NewNode().Children should be empty, got %d", len(node.Children))
			}

			if node.Parent != nil {
				t.Error("NewNode().Parent should be nil")
			}
		})
	}
}

func TestNode_AddChild(t *testing.T) {
	parent := NewNode("parent", "/parent", true)
	child := NewNode("child", "/parent/child", false)

	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("AddChild: parent should have 1 child, got %d", len(parent.Children))
	}

	if parent.Children[0] != child {
		t.Error("AddChild: child should be in parent's children")
	}

	if child.Parent != parent {
		t.Error("AddChild: child's parent should be set")
	}

	if child.Level != 1 {
		t.Errorf("AddChild: child level should be 1, got %d", child.Level)
	}
}

func TestNode_AddChild_Multiple(t *testing.T) {
	parent := NewNode("parent", "/parent", true)
	child1 := NewNode("child1", "/parent/child1", false)
	child2 := NewNode("child2", "/parent/child2", true)

	parent.AddChild(child1)
	parent.AddChild(child2)

	if len(parent.Children) != 2 {
		t.Errorf("AddChild: parent should have 2 children, got %d", len(parent.Children))
	}
}

func TestNode_AddChild_Nested(t *testing.T) {
	root := NewNode("root", "/root", true)
	level1 := NewNode("level1", "/root/level1", true)
	level2 := NewNode("level2", "/root/level1/level2", true)
	level3 := NewNode("level3", "/root/level1/level2/level3", false)

	root.AddChild(level1)
	level1.AddChild(level2)
	level2.AddChild(level3)

	if level1.Level != 1 {
		t.Errorf("level1.Level = %d, want 1", level1.Level)
	}

	if level2.Level != 2 {
		t.Errorf("level2.Level = %d, want 2", level2.Level)
	}

	if level3.Level != 3 {
		t.Errorf("level3.Level = %d, want 3", level3.Level)
	}
}

func TestNode_IsLeaf(t *testing.T) {
	parent := NewNode("parent", "/parent", true)
	child := NewNode("child", "/parent/child", false)

	if !parent.IsLeaf() {
		t.Error("IsLeaf: node without children should be leaf")
	}

	parent.AddChild(child)

	if parent.IsLeaf() {
		t.Error("IsLeaf: node with children should not be leaf")
	}

	if !child.IsLeaf() {
		t.Error("IsLeaf: child without children should be leaf")
	}
}

func TestNode_FullPath(t *testing.T) {
	root := NewNode("project", "/project", true)
	child := NewNode("src", "/project/src", true)

	root.AddChild(child)

	// Root node (no parent) returns Name
	if got := root.FullPath(); got != "project" {
		t.Errorf("FullPath() for root = %q, want %q", got, "project")
	}

	// Child node returns Path
	if got := child.FullPath(); got != "/project/src" {
		t.Errorf("FullPath() for child = %q, want %q", got, "/project/src")
	}
}

func TestCalculateStats(t *testing.T) {
	// Build tree:
	// root/
	// ├── dir1/
	// │   ├── file1.txt
	// │   └── file2.txt
	// ├── dir2/
	// │   └── subdir/
	// │       └── file3.txt
	// └── file4.txt
	root := NewNode("root", "/root", true)
	dir1 := NewNode("dir1", "/root/dir1", true)
	dir2 := NewNode("dir2", "/root/dir2", true)
	subdir := NewNode("subdir", "/root/dir2/subdir", true)
	file1 := NewNode("file1.txt", "/root/dir1/file1.txt", false)
	file2 := NewNode("file2.txt", "/root/dir1/file2.txt", false)
	file3 := NewNode("file3.txt", "/root/dir2/subdir/file3.txt", false)
	file4 := NewNode("file4.txt", "/root/file4.txt", false)

	root.AddChild(dir1)
	root.AddChild(dir2)
	root.AddChild(file4)
	dir1.AddChild(file1)
	dir1.AddChild(file2)
	dir2.AddChild(subdir)
	subdir.AddChild(file3)

	stats := CalculateStats(root)

	if stats.TotalDirs != 4 {
		t.Errorf("TotalDirs = %d, want 4", stats.TotalDirs)
	}

	if stats.TotalFiles != 4 {
		t.Errorf("TotalFiles = %d, want 4", stats.TotalFiles)
	}

	if stats.MaxDepth != 3 {
		t.Errorf("MaxDepth = %d, want 3", stats.MaxDepth)
	}
}

func TestCalculateStats_EmptyTree(t *testing.T) {
	root := NewNode("root", "/root", true)
	stats := CalculateStats(root)

	if stats.TotalDirs != 1 {
		t.Errorf("TotalDirs = %d, want 1", stats.TotalDirs)
	}

	if stats.TotalFiles != 0 {
		t.Errorf("TotalFiles = %d, want 0", stats.TotalFiles)
	}

	if stats.MaxDepth != 0 {
		t.Errorf("MaxDepth = %d, want 0", stats.MaxDepth)
	}
}

func TestCalculateStats_Nil(t *testing.T) {
	stats := CalculateStats(nil)

	if stats.TotalDirs != 0 {
		t.Errorf("TotalDirs = %d, want 0", stats.TotalDirs)
	}

	if stats.TotalFiles != 0 {
		t.Errorf("TotalFiles = %d, want 0", stats.TotalFiles)
	}
}

func TestNode_ToJSON(t *testing.T) {
	root := NewNode("project", "/project", true)
	root.Hash = "abc123"
	root.Comment = "root dir"

	child := NewNode("main.go", "/project/main.go", false)
	child.Hash = "def456"
	root.AddChild(child)

	jsonNode := root.ToJSON()

	if jsonNode.Name != "project" {
		t.Errorf("ToJSON().Name = %q, want %q", jsonNode.Name, "project")
	}

	if jsonNode.Path != "/project" {
		t.Errorf("ToJSON().Path = %q, want %q", jsonNode.Path, "/project")
	}

	if !jsonNode.IsDir {
		t.Error("ToJSON().IsDir should be true")
	}

	if jsonNode.Hash != "abc123" {
		t.Errorf("ToJSON().Hash = %q, want %q", jsonNode.Hash, "abc123")
	}

	if jsonNode.Comment != "root dir" {
		t.Errorf("ToJSON().Comment = %q, want %q", jsonNode.Comment, "root dir")
	}

	if len(jsonNode.Children) != 1 {
		t.Errorf("ToJSON().Children length = %d, want 1", len(jsonNode.Children))
	}
}

func TestNode_ToJSON_Nil(t *testing.T) {
	var node *Node

	jsonNode := node.ToJSON()

	if jsonNode != nil {
		t.Error("ToJSON() on nil should return nil")
	}
}

func TestTreeStats_ToJSONStats(t *testing.T) {
	stats := &TreeStats{
		TotalDirs:  5,
		TotalFiles: 10,
		MaxDepth:   3,
	}

	jsonStats := stats.ToJSONStats()

	if jsonStats.TotalDirs != 5 {
		t.Errorf("ToJSONStats().TotalDirs = %d, want 5", jsonStats.TotalDirs)
	}

	if jsonStats.TotalFiles != 10 {
		t.Errorf("ToJSONStats().TotalFiles = %d, want 10", jsonStats.TotalFiles)
	}

	if jsonStats.MaxDepth != 3 {
		t.Errorf("ToJSONStats().MaxDepth = %d, want 3", jsonStats.MaxDepth)
	}

	if jsonStats.TotalItems != 15 {
		t.Errorf("ToJSONStats().TotalItems = %d, want 15", jsonStats.TotalItems)
	}
}

func TestTreeStats_ToJSONStats_Nil(t *testing.T) {
	var stats *TreeStats

	jsonStats := stats.ToJSONStats()

	if jsonStats != nil {
		t.Error("ToJSONStats() on nil should return nil")
	}
}

func TestNode_MarshalJSON(t *testing.T) {
	root := NewNode("project", "/project", true)
	child := NewNode("main.go", "/project/main.go", false)
	root.AddChild(child)

	data, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var result JSONNode

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if result.Name != "project" {
		t.Errorf("Unmarshaled Name = %q, want %q", result.Name, "project")
	}

	if len(result.Children) != 1 {
		t.Errorf("Unmarshaled Children length = %d, want 1", len(result.Children))
	}
}
