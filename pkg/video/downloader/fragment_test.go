package downloader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadFragmentState(t *testing.T) {
	t.Run("roundtrip", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")

		orig := &FragmentState{
			TotalFragments: 100,
			LastFragment:   42,
			Filename:       "video.mp4",
		}

		if err := SaveFragmentState(path, orig); err != nil {
			t.Fatalf("SaveFragmentState: %v", err)
		}

		loaded, err := LoadFragmentState(path)
		if err != nil {
			t.Fatalf("LoadFragmentState: %v", err)
		}

		if loaded.TotalFragments != orig.TotalFragments {
			t.Errorf("TotalFragments = %d, want %d", loaded.TotalFragments, orig.TotalFragments)
		}
		if loaded.LastFragment != orig.LastFragment {
			t.Errorf("LastFragment = %d, want %d", loaded.LastFragment, orig.LastFragment)
		}
		if loaded.Filename != orig.Filename {
			t.Errorf("Filename = %q, want %q", loaded.Filename, orig.Filename)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := LoadFragmentState("/nonexistent/path/state.json")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.json")
		if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
			t.Fatal(err)
		}

		_, err := LoadFragmentState(path)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestRemoveFragmentState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	if err := SaveFragmentState(path, &FragmentState{TotalFragments: 1}); err != nil {
		t.Fatal(err)
	}

	RemoveFragmentState(path)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}
}

func TestAppendToFile(t *testing.T) {
	t.Run("creates and appends", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "output.bin")

		if err := AppendToFile(path, []byte("hello")); err != nil {
			t.Fatalf("first append: %v", err)
		}

		if err := AppendToFile(path, []byte(" world")); err != nil {
			t.Fatalf("second append: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		if string(data) != "hello world" {
			t.Errorf("got %q, want %q", string(data), "hello world")
		}
	})

	t.Run("creates new file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "new.bin")

		if err := AppendToFile(path, []byte("data")); err != nil {
			t.Fatal(err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		if string(data) != "data" {
			t.Errorf("got %q, want %q", string(data), "data")
		}
	})
}
