package note

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunNoteAddAndList(t *testing.T) {
	file := filepath.Join(t.TempDir(), "notes.json")

	var addOut bytes.Buffer
	if err := RunNote(&addOut, []string{"buy", "milk"}, Options{File: file}); err != nil {
		t.Fatalf("RunNote(add) error = %v", err)
	}

	if !strings.Contains(addOut.String(), "Saved note") {
		t.Fatalf("expected save message, got %q", addOut.String())
	}

	var listOut bytes.Buffer
	if err := RunNote(&listOut, nil, Options{File: file, List: true}); err != nil {
		t.Fatalf("RunNote(list) error = %v", err)
	}

	if !strings.Contains(listOut.String(), "buy milk") {
		t.Fatalf("expected listed note text, got %q", listOut.String())
	}
}

func TestRunNoteListJSON(t *testing.T) {
	file := filepath.Join(t.TempDir(), "notes.json")

	if err := RunNote(&bytes.Buffer{}, []string{"alpha"}, Options{File: file}); err != nil {
		t.Fatalf("RunNote(add) error = %v", err)
	}

	var out bytes.Buffer
	if err := RunNote(&out, nil, Options{File: file, List: true, JSON: true}); err != nil {
		t.Fatalf("RunNote(list json) error = %v", err)
	}

	var store Store
	if err := json.Unmarshal(out.Bytes(), &store); err != nil {
		t.Fatalf("json.Unmarshal output error = %v; output=%q", err, out.String())
	}

	if len(store.Notes) != 1 || store.Notes[0].Text != "alpha" {
		t.Fatalf("unexpected store output: %+v", store)
	}
}

func TestRunNoteRequiresText(t *testing.T) {
	err := RunNote(&bytes.Buffer{}, nil, Options{File: filepath.Join(t.TempDir(), "notes.json")})
	if err == nil {
		t.Fatal("expected error for empty note text")
	}
}

func TestRunRemoveByIndex(t *testing.T) {
	file := filepath.Join(t.TempDir(), "notes.json")

	if err := RunNote(&bytes.Buffer{}, []string{"alpha"}, Options{File: file}); err != nil {
		t.Fatalf("RunNote(add alpha) error = %v", err)
	}

	if err := RunNote(&bytes.Buffer{}, []string{"beta"}, Options{File: file}); err != nil {
		t.Fatalf("RunNote(add beta) error = %v", err)
	}

	var out bytes.Buffer
	if err := RunRemove(&out, []string{"1"}, Options{File: file}); err != nil {
		t.Fatalf("RunRemove(index) error = %v", err)
	}

	if !strings.Contains(out.String(), "Removed note #1") {
		t.Fatalf("expected remove message with index, got %q", out.String())
	}

	var listOut bytes.Buffer
	if err := RunNote(&listOut, nil, Options{File: file, List: true}); err != nil {
		t.Fatalf("RunNote(list) error = %v", err)
	}

	if strings.Contains(listOut.String(), "alpha") {
		t.Fatalf("expected alpha to be removed, got %q", listOut.String())
	}

	if !strings.Contains(listOut.String(), "beta") {
		t.Fatalf("expected beta to remain, got %q", listOut.String())
	}
}

func TestRunRemoveByID(t *testing.T) {
	file := filepath.Join(t.TempDir(), "notes.json")

	if err := RunNote(&bytes.Buffer{}, []string{"alpha"}, Options{File: file}); err != nil {
		t.Fatalf("RunNote(add alpha) error = %v", err)
	}

	if err := RunNote(&bytes.Buffer{}, []string{"beta"}, Options{File: file}); err != nil {
		t.Fatalf("RunNote(add beta) error = %v", err)
	}

	var out bytes.Buffer
	if err := RunNote(&out, nil, Options{File: file, List: true, JSON: true}); err != nil {
		t.Fatalf("RunNote(list json) error = %v", err)
	}

	var store Store
	if err := json.Unmarshal(out.Bytes(), &store); err != nil {
		t.Fatalf("json.Unmarshal output error = %v; output=%q", err, out.String())
	}

	if len(store.Notes) < 2 {
		t.Fatalf("expected at least 2 notes, got %d", len(store.Notes))
	}

	targetID := store.Notes[0].ID
	if err := RunRemove(&bytes.Buffer{}, []string{targetID}, Options{File: file}); err != nil {
		t.Fatalf("RunRemove(id) error = %v", err)
	}

	var listAfter bytes.Buffer
	if err := RunNote(&listAfter, nil, Options{File: file, List: true, JSON: true}); err != nil {
		t.Fatalf("RunNote(list json after remove) error = %v", err)
	}

	var after Store
	if err := json.Unmarshal(listAfter.Bytes(), &after); err != nil {
		t.Fatalf("json.Unmarshal after output error = %v; output=%q", err, listAfter.String())
	}

	for _, n := range after.Notes {
		if n.ID == targetID {
			t.Fatalf("expected id %s to be removed", targetID)
		}
	}
}

func TestRunRemoveRequiresTarget(t *testing.T) {
	err := RunRemove(&bytes.Buffer{}, nil, Options{File: filepath.Join(t.TempDir(), "notes.json")})
	if err == nil {
		t.Fatal("expected error for missing remove target")
	}
}

func TestResolveNotesPathDefault(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	oldHome, hadHome := os.LookupEnv("HOME")
	oldProfile, hadProfile := os.LookupEnv("USERPROFILE")
	_ = os.Setenv("HOME", home)
	_ = os.Setenv("USERPROFILE", home)

	t.Cleanup(func() {
		if hadHome {
			_ = os.Setenv("HOME", oldHome)
		} else {
			_ = os.Unsetenv("HOME")
		}

		if hadProfile {
			_ = os.Setenv("USERPROFILE", oldProfile)
		} else {
			_ = os.Unsetenv("USERPROFILE")
		}
	})

	path, err := resolveNotesPath("")
	if err != nil {
		t.Fatalf("resolveNotesPath() error = %v", err)
	}

	wantHome := home
	if runtime.GOOS == "windows" {
		wantHome = home
	}

	want := filepath.Join(wantHome, "Documents", defaultFileName)
	if path != want {
		t.Fatalf("resolveNotesPath() = %q, want %q", path, want)
	}
}
