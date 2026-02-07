package comparer

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/inovacc/omni/pkg/twig/models"
)

// ChangeType represents the kind of filesystem change
type ChangeType string

const (
	Added    ChangeType = "added"
	Removed  ChangeType = "removed"
	Modified ChangeType = "modified"
	Moved    ChangeType = "moved"
)

// Change represents a single filesystem change between two snapshots
type Change struct {
	Type    ChangeType `json:"type"`
	Path    string     `json:"path"`
	OldPath string     `json:"old_path,omitempty"`
	IsDir   bool       `json:"is_dir"`
	OldHash string     `json:"old_hash,omitempty"`
	NewHash string     `json:"new_hash,omitempty"`
}

// Summary contains counts for each change type
type Summary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
	Moved    int `json:"moved"`
}

// CompareResult contains the full result of comparing two tree snapshots
type CompareResult struct {
	Changes   []Change `json:"changes"`
	Summary   Summary  `json:"summary"`
	LeftPath  string   `json:"left_path"`
	RightPath string   `json:"right_path"`
}

// CompareConfig controls comparison behavior
type CompareConfig struct {
	DetectMoves bool // Match removed+added pairs by hash
}

// flatNode is a flattened representation of a JSONNode with its relative path
type flatNode struct {
	relPath string
	isDir   bool
	hash    string
}

// Compare compares two JSON tree snapshots and returns the differences.
// The algorithm has 5 phases:
//  1. Flatten both trees to maps of relativePath -> flatNode
//  2. Find removed entries (in left, not in right)
//  3. Find added entries (in right, not in left)
//  4. Detect moves: match removed+added pairs by hash (if DetectMoves)
//  5. Find modified entries: same path, different hash
func Compare(left, right *models.JSONNode, cfg CompareConfig) *CompareResult {
	result := &CompareResult{}

	if left == nil && right == nil {
		return result
	}

	// Phase 1: Flatten both trees
	leftMap := make(map[string]flatNode)
	rightMap := make(map[string]flatNode)

	if left != nil {
		flattenTree(left, "", leftMap)
	}

	if right != nil {
		flattenTree(right, "", rightMap)
	}

	// Phase 2: Find removed (in left, not in right)
	var removed []Change
	for path, node := range leftMap {
		if _, exists := rightMap[path]; !exists {
			removed = append(removed, Change{
				Type:    Removed,
				Path:    path,
				IsDir:   node.isDir,
				OldHash: node.hash,
			})
		}
	}

	// Phase 3: Find added (in right, not in left)
	var added []Change
	for path, node := range rightMap {
		if _, exists := leftMap[path]; !exists {
			added = append(added, Change{
				Type:    Added,
				Path:    path,
				IsDir:   node.isDir,
				NewHash: node.hash,
			})
		}
	}

	// Phase 4: Detect moves (match removed+added by hash)
	var moved []Change
	if cfg.DetectMoves {
		// Build hash -> paths maps for removed and added
		removedByHash := make(map[string][]int) // hash -> indices in removed slice
		for i, c := range removed {
			if c.OldHash != "" {
				removedByHash[c.OldHash] = append(removedByHash[c.OldHash], i)
			}
		}

		matchedRemoved := make(map[int]bool)
		matchedAdded := make(map[int]bool)

		for i, c := range added {
			if c.NewHash == "" {
				continue
			}

			indices, ok := removedByHash[c.NewHash]
			if !ok || len(indices) == 0 {
				continue
			}

			// Use the first unmatched removed entry with this hash
			for _, ri := range indices {
				if matchedRemoved[ri] {
					continue
				}

				moved = append(moved, Change{
					Type:    Moved,
					Path:    c.Path,
					OldPath: removed[ri].Path,
					IsDir:   c.IsDir,
					OldHash: c.NewHash,
					NewHash: c.NewHash,
				})

				matchedRemoved[ri] = true
				matchedAdded[i] = true

				break
			}
		}

		// Filter out matched entries from removed and added
		if len(matchedRemoved) > 0 || len(matchedAdded) > 0 {
			var filteredRemoved []Change
			for i, c := range removed {
				if !matchedRemoved[i] {
					filteredRemoved = append(filteredRemoved, c)
				}
			}

			removed = filteredRemoved

			var filteredAdded []Change
			for i, c := range added {
				if !matchedAdded[i] {
					filteredAdded = append(filteredAdded, c)
				}
			}

			added = filteredAdded
		}
	}

	// Phase 5: Find modified (same path, different hash)
	var modified []Change
	for path, leftNode := range leftMap {
		rightNode, exists := rightMap[path]
		if !exists {
			continue
		}

		if leftNode.isDir || rightNode.isDir {
			continue // Skip directories for modification check
		}

		if leftNode.hash != "" && rightNode.hash != "" && leftNode.hash != rightNode.hash {
			modified = append(modified, Change{
				Type:    Modified,
				Path:    path,
				IsDir:   false,
				OldHash: leftNode.hash,
				NewHash: rightNode.hash,
			})
		}
	}

	// Sort all changes by path for deterministic output
	sort.Slice(removed, func(i, j int) bool { return removed[i].Path < removed[j].Path })
	sort.Slice(added, func(i, j int) bool { return added[i].Path < added[j].Path })
	sort.Slice(modified, func(i, j int) bool { return modified[i].Path < modified[j].Path })
	sort.Slice(moved, func(i, j int) bool { return moved[i].Path < moved[j].Path })

	// Combine all changes: removed, added, modified, moved
	result.Changes = append(result.Changes, removed...)
	result.Changes = append(result.Changes, added...)
	result.Changes = append(result.Changes, modified...)
	result.Changes = append(result.Changes, moved...)

	result.Summary = Summary{
		Added:    len(added),
		Removed:  len(removed),
		Modified: len(modified),
		Moved:    len(moved),
	}

	return result
}

// flattenTree recursively flattens a JSONNode tree into a map of relative path -> flatNode.
// The root node's name is used as the prefix.
func flattenTree(node *models.JSONNode, prefix string, out map[string]flatNode) {
	var relPath string
	if prefix == "" {
		relPath = node.Name
	} else {
		relPath = prefix + "/" + node.Name
	}

	// Normalize to forward slashes
	relPath = filepath.ToSlash(relPath)
	relPath = strings.TrimSuffix(relPath, "/")

	out[relPath] = flatNode{
		relPath: relPath,
		isDir:   node.IsDir,
		hash:    node.Hash,
	}

	for _, child := range node.Children {
		flattenTree(child, relPath, out)
	}
}
