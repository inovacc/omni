package note

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

func TestNoteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	// Use a nested path so save()'s MkdirAll branch runs.
	path := filepath.Join(dir, "sub", "notes.json")
	opts := Options{File: path}

	// Add three notes.
	for _, txt := range []string{"first", "second", "third"} {
		var buf bytes.Buffer
		if err := RunNote(&buf, []string{txt}, opts); err != nil {
			t.Fatalf("RunNote(%q): %v", txt, err)
		}
		if !strings.Contains(buf.String(), "Saved note") {
			t.Errorf("missing save confirmation: %q", buf.String())
		}
	}

	// List all (text mode).
	var list bytes.Buffer
	if err := runList(&list, path, Options{File: path}); err != nil {
		t.Fatalf("runList: %v", err)
	}
	for _, want := range []string{"first", "second", "third"} {
		if !strings.Contains(list.String(), want) {
			t.Errorf("list missing %q: %q", want, list.String())
		}
	}

	// List with limit keeps only the most recent.
	var limited bytes.Buffer
	if err := runList(&limited, path, Options{File: path, Limit: 1}); err != nil {
		t.Fatalf("runList limit: %v", err)
	}
	if strings.Contains(limited.String(), "first") {
		t.Errorf("limit=1 should drop oldest: %q", limited.String())
	}

	// List as JSON.
	var jsonOut bytes.Buffer
	if err := runList(&jsonOut, path, Options{File: path, OutputFormat: output.FormatJSON}); err != nil {
		t.Fatalf("runList json: %v", err)
	}
	if !strings.Contains(jsonOut.String(), "\"notes\"") {
		t.Errorf("json output missing notes key: %q", jsonOut.String())
	}

	// Remove by 1-based index.
	var rm bytes.Buffer
	if err := RunRemove(&rm, []string{"1"}, opts); err != nil {
		t.Fatalf("RunRemove index: %v", err)
	}
	if !strings.Contains(rm.String(), "Removed note") {
		t.Errorf("remove confirmation missing: %q", rm.String())
	}

	// Verify two remain.
	store, err := load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(store.Notes) != 2 {
		t.Fatalf("after remove want 2 notes, got %d", len(store.Notes))
	}

	// Remove by ID (JSON output path).
	id := store.Notes[0].ID
	var rmID bytes.Buffer
	if err := RunRemove(&rmID, []string{id}, Options{File: path, OutputFormat: output.FormatJSON}); err != nil {
		t.Fatalf("RunRemove by id: %v", err)
	}
}

func TestNoteRemoveErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notes.json")

	// Remove with no target.
	if err := RunRemove(&bytes.Buffer{}, nil, Options{File: path}); err == nil {
		t.Fatal("expected error for missing target")
	}
	if err := RunRemove(&bytes.Buffer{}, []string{"  "}, Options{File: path}); err == nil {
		t.Fatal("expected error for blank target")
	}

	// Remove from empty store.
	if err := RunRemove(&bytes.Buffer{}, []string{"1"}, Options{File: path}); err == nil {
		t.Fatal("expected error removing from empty store")
	}

	// Add one, then index out of range and unknown id.
	if err := RunNote(&bytes.Buffer{}, []string{"only"}, Options{File: path}); err != nil {
		t.Fatal(err)
	}
	if err := RunRemove(&bytes.Buffer{}, []string{"99"}, Options{File: path}); err == nil {
		t.Fatal("expected out-of-range error")
	}
	if err := RunRemove(&bytes.Buffer{}, []string{"no-such-id"}, Options{File: path}); err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestNoteRequiresText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "n.json")
	if err := RunNote(&bytes.Buffer{}, nil, Options{File: path}); err == nil {
		t.Fatal("expected error for empty note text")
	}
}

func TestNoteListViaRunNote(t *testing.T) {
	path := filepath.Join(t.TempDir(), "n.json")
	// Empty list prints a friendly message.
	var buf bytes.Buffer
	if err := RunNote(&buf, nil, Options{File: path, List: true}); err != nil {
		t.Fatalf("RunNote list: %v", err)
	}
	if !strings.Contains(buf.String(), "No notes yet") {
		t.Errorf("expected empty notice, got %q", buf.String())
	}
}

func TestNoteLoadParseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := load(path); err == nil {
		t.Fatal("expected parse error on malformed JSON")
	}
}

func TestNoteLoadEmptyAndMissing(t *testing.T) {
	dir := t.TempDir()
	// Missing file -> empty store, no error.
	if s, err := load(filepath.Join(dir, "absent.json")); err != nil || len(s.Notes) != 0 {
		t.Fatalf("missing file load = %+v, %v", s, err)
	}
	// Empty file -> empty store.
	empty := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if s, err := load(empty); err != nil || len(s.Notes) != 0 {
		t.Fatalf("empty file load = %+v, %v", s, err)
	}
}
