package models

import (
	"encoding/json"
	"os"
)

// JSONNode represents a node in JSON-serializable format
type JSONNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"is_dir"`
	Hash     string      `json:"hash,omitempty"`
	Comment  string      `json:"comment,omitempty"`
	Children []*JSONNode `json:"children,omitempty"`
}

// JSONStats represents tree statistics in JSON format
type JSONStats struct {
	TotalDirs  int `json:"total_dirs"`
	TotalFiles int `json:"total_files"`
	MaxDepth   int `json:"max_depth"`
	TotalItems int `json:"total_items"`
}

// JSONOutput represents the complete JSON output structure
type JSONOutput struct {
	Tree  *JSONNode  `json:"tree"`
	Stats *JSONStats `json:"stats,omitempty"`
}

// ToJSON converts a Node tree to a JSONNode tree
func (n *Node) ToJSON() *JSONNode {
	if n == nil {
		return nil
	}

	jsonNode := &JSONNode{
		Name:    n.Name,
		Path:    n.Path,
		IsDir:   n.IsDir,
		Hash:    n.Hash,
		Comment: n.Comment,
	}

	if len(n.Children) > 0 {
		jsonNode.Children = make([]*JSONNode, len(n.Children))
		for i, child := range n.Children {
			jsonNode.Children[i] = child.ToJSON()
		}
	}

	return jsonNode
}

// ToJSONStats converts TreeStats to JSONStats
func (s *TreeStats) ToJSONStats() *JSONStats {
	if s == nil {
		return nil
	}

	return &JSONStats{
		TotalDirs:  s.TotalDirs,
		TotalFiles: s.TotalFiles,
		MaxDepth:   s.MaxDepth,
		TotalItems: s.TotalDirs + s.TotalFiles,
	}
}

// MarshalJSON returns the JSON encoding of the Node tree
func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.ToJSON())
}

// Node represents a file or directory in the tree structure
type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*Node
	Parent   *Node
	Level    int
	Comment  string // Optional comment from tree format
	FileInfo os.FileInfo
	Hash     string // File hash (SHA256)
}

// NewNode creates a new Node with the given name, path, and directory flag.
// The returned node has an empty children slice and no parent initially.
func NewNode(name string, path string, isDir bool) *Node {
	return &Node{
		Name:     name,
		Path:     path,
		IsDir:    isDir,
		Children: make([]*Node, 0),
	}
}

// AddChild adds a child node to this node, automatically setting the child's
// parent reference and level (parent level + 1).
func (n *Node) AddChild(child *Node) {
	child.Parent = n
	child.Level = n.Level + 1
	n.Children = append(n.Children, child)
}

// IsLeaf returns true if the node has no children, false otherwise.
func (n *Node) IsLeaf() bool {
	return len(n.Children) == 0
}

// FullPath returns the complete path of the node. For root nodes (no parent),
// returns the name. For all other nodes, returns the Path field.
func (n *Node) FullPath() string {
	if n.Parent == nil {
		return n.Name
	}

	return n.Path
}

// TreeStats contains statistics about the tree structure
type TreeStats struct {
	TotalDirs  int
	TotalFiles int
	MaxDepth   int
}

// CalculateStats calculates and returns statistics for the entire tree structure
// starting from the given root node. It traverses the tree recursively and counts
// total directories, total files, and maximum depth.
func CalculateStats(root *Node) *TreeStats {
	stats := &TreeStats{}
	calculateStatsRecursive(root, stats, 0)

	return stats
}

func calculateStatsRecursive(node *Node, stats *TreeStats, depth int) {
	if node == nil {
		return
	}

	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	if node.IsDir {
		stats.TotalDirs++
	} else {
		stats.TotalFiles++
	}

	for _, child := range node.Children {
		calculateStatsRecursive(child, stats, depth+1)
	}
}
