package comparer

import (
	"testing"

	"github.com/inovacc/omni/pkg/twig/models"
)

func TestCompare_BothNil(t *testing.T) {
	result := Compare(nil, nil, CompareConfig{})

	if len(result.Changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(result.Changes))
	}
}

func TestCompare_AddedFiles(t *testing.T) {
	left := &models.JSONNode{
		Name:  "root",
		IsDir: true,
	}

	right := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "new.txt", IsDir: false, Hash: "abc123"},
		},
	}

	result := Compare(left, right, CompareConfig{})

	if result.Summary.Added != 1 {
		t.Errorf("expected 1 added, got %d", result.Summary.Added)
	}

	found := false

	for _, c := range result.Changes {
		if c.Type == Added && c.Path == "root/new.txt" {
			found = true

			if c.NewHash != "abc123" {
				t.Errorf("expected NewHash abc123, got %s", c.NewHash)
			}
		}
	}

	if !found {
		t.Error("expected to find added change for root/new.txt")
	}
}

func TestCompare_RemovedFiles(t *testing.T) {
	left := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "old.txt", IsDir: false, Hash: "def456"},
		},
	}

	right := &models.JSONNode{
		Name:  "root",
		IsDir: true,
	}

	result := Compare(left, right, CompareConfig{})

	if result.Summary.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", result.Summary.Removed)
	}

	found := false

	for _, c := range result.Changes {
		if c.Type == Removed && c.Path == "root/old.txt" {
			found = true

			if c.OldHash != "def456" {
				t.Errorf("expected OldHash def456, got %s", c.OldHash)
			}
		}
	}

	if !found {
		t.Error("expected to find removed change for root/old.txt")
	}
}

func TestCompare_ModifiedFiles(t *testing.T) {
	left := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "file.txt", IsDir: false, Hash: "hash1"},
		},
	}

	right := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "file.txt", IsDir: false, Hash: "hash2"},
		},
	}

	result := Compare(left, right, CompareConfig{})

	if result.Summary.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", result.Summary.Modified)
	}

	found := false

	for _, c := range result.Changes {
		if c.Type == Modified && c.Path == "root/file.txt" {
			found = true

			if c.OldHash != "hash1" || c.NewHash != "hash2" {
				t.Errorf("expected hashes hash1->hash2, got %s->%s", c.OldHash, c.NewHash)
			}
		}
	}

	if !found {
		t.Error("expected to find modified change for root/file.txt")
	}
}

func TestCompare_DetectMoves(t *testing.T) {
	left := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "src", IsDir: true, Children: []*models.JSONNode{
				{Name: "util.go", IsDir: false, Hash: "samehash"},
			}},
		},
	}

	right := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "pkg", IsDir: true, Children: []*models.JSONNode{
				{Name: "util.go", IsDir: false, Hash: "samehash"},
			}},
		},
	}

	result := Compare(left, right, CompareConfig{DetectMoves: true})

	if result.Summary.Moved != 1 {
		t.Errorf("expected 1 moved, got %d", result.Summary.Moved)
	}

	found := false

	for _, c := range result.Changes {
		if c.Type == Moved {
			found = true

			if c.Path != "root/pkg/util.go" {
				t.Errorf("expected move target root/pkg/util.go, got %s", c.Path)
			}

			if c.OldPath != "root/src/util.go" {
				t.Errorf("expected move source root/src/util.go, got %s", c.OldPath)
			}
		}
	}

	if !found {
		t.Error("expected to find moved change")
	}
}

func TestCompare_DetectMovesDisabled(t *testing.T) {
	left := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "old.txt", IsDir: false, Hash: "samehash"},
		},
	}

	right := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "new.txt", IsDir: false, Hash: "samehash"},
		},
	}

	result := Compare(left, right, CompareConfig{DetectMoves: false})

	if result.Summary.Moved != 0 {
		t.Errorf("expected 0 moved when DetectMoves=false, got %d", result.Summary.Moved)
	}

	if result.Summary.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", result.Summary.Removed)
	}

	if result.Summary.Added != 1 {
		t.Errorf("expected 1 added, got %d", result.Summary.Added)
	}
}

func TestCompare_NoChanges(t *testing.T) {
	tree := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "file.txt", IsDir: false, Hash: "same"},
		},
	}

	result := Compare(tree, tree, CompareConfig{})

	if len(result.Changes) != 0 {
		t.Errorf("expected 0 changes for identical trees, got %d", len(result.Changes))
	}
}

func TestCompare_ComplexScenario(t *testing.T) {
	left := &models.JSONNode{
		Name:  "project",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "src", IsDir: true, Children: []*models.JSONNode{
				{Name: "main.go", IsDir: false, Hash: "hash_main"},
				{Name: "utils.go", IsDir: false, Hash: "hash_utils"},
				{Name: "old.go", IsDir: false, Hash: "hash_old"},
			}},
			{Name: "docs", IsDir: true, Children: []*models.JSONNode{
				{Name: "readme.md", IsDir: false, Hash: "hash_readme"},
			}},
		},
	}

	right := &models.JSONNode{
		Name:  "project",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "src", IsDir: true, Children: []*models.JSONNode{
				{Name: "main.go", IsDir: false, Hash: "hash_main_v2"}, // modified
				{Name: "utils.go", IsDir: false, Hash: "hash_utils"},  // unchanged
				// old.go removed
			}},
			{Name: "docs", IsDir: true, Children: []*models.JSONNode{
				{Name: "readme.md", IsDir: false, Hash: "hash_readme"},
			}},
			{Name: "tests", IsDir: true, Children: []*models.JSONNode{ // new dir
				{Name: "main_test.go", IsDir: false, Hash: "hash_test"},
			}},
		},
	}

	result := Compare(left, right, CompareConfig{DetectMoves: true})

	if result.Summary.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", result.Summary.Modified)
	}

	if result.Summary.Removed != 1 {
		t.Errorf("expected 1 removed (old.go), got %d", result.Summary.Removed)
	}

	// tests dir + main_test.go = 2 added
	if result.Summary.Added != 2 {
		t.Errorf("expected 2 added (tests dir + main_test.go), got %d", result.Summary.Added)
	}
}

func TestCompare_LeftNilRightPopulated(t *testing.T) {
	right := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "file.txt", IsDir: false, Hash: "hash1"},
		},
	}

	result := Compare(nil, right, CompareConfig{})

	// root + file.txt = 2 added
	if result.Summary.Added != 2 {
		t.Errorf("expected 2 added, got %d", result.Summary.Added)
	}
}

func TestCompare_LeftPopulatedRightNil(t *testing.T) {
	left := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "file.txt", IsDir: false, Hash: "hash1"},
		},
	}

	result := Compare(left, nil, CompareConfig{})

	// root + file.txt = 2 removed
	if result.Summary.Removed != 2 {
		t.Errorf("expected 2 removed, got %d", result.Summary.Removed)
	}
}

func TestCompare_SortedOutput(t *testing.T) {
	left := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "z.txt", IsDir: false, Hash: "h1"},
			{Name: "a.txt", IsDir: false, Hash: "h2"},
			{Name: "m.txt", IsDir: false, Hash: "h3"},
		},
	}

	right := &models.JSONNode{
		Name:  "root",
		IsDir: true,
	}

	result := Compare(left, right, CompareConfig{})

	// All 3 files removed. Check they're sorted
	removedPaths := []string{}

	for _, c := range result.Changes {
		if c.Type == Removed && !c.IsDir {
			removedPaths = append(removedPaths, c.Path)
		}
	}

	for i := 1; i < len(removedPaths); i++ {
		if removedPaths[i-1] > removedPaths[i] {
			t.Errorf("changes not sorted: %s > %s", removedPaths[i-1], removedPaths[i])
		}
	}
}

func TestCompare_SkipsModifiedDirs(t *testing.T) {
	// Directories should not be reported as modified even if both exist
	left := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "dir", IsDir: true},
		},
	}

	right := &models.JSONNode{
		Name:  "root",
		IsDir: true,
		Children: []*models.JSONNode{
			{Name: "dir", IsDir: true},
		},
	}

	result := Compare(left, right, CompareConfig{})

	if result.Summary.Modified != 0 {
		t.Errorf("expected 0 modified for directory-only tree, got %d", result.Summary.Modified)
	}
}
