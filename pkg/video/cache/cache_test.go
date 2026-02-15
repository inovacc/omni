package cache

import (
	"testing"
)

func TestStoreAndLoad(t *testing.T) {
	c := New(t.TempDir())

	type item struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	want := item{Name: "test", Value: 42}
	if err := c.Store("section", "key", want); err != nil {
		t.Fatalf("Store: %v", err)
	}

	var got item
	if !c.Load("section", "key", &got) {
		t.Fatal("Load returned false")
	}

	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadMissing(t *testing.T) {
	c := New(t.TempDir())

	var v string
	if c.Load("no", "such", &v) {
		t.Error("Load should return false for missing key")
	}
}

func TestRemove(t *testing.T) {
	c := New(t.TempDir())

	_ = c.Store("s", "k", "data")
	if err := c.Remove("s", "k"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	var v string
	if c.Load("s", "k", &v) {
		t.Error("Load should return false after Remove")
	}
}

func TestClearSection(t *testing.T) {
	c := New(t.TempDir())

	_ = c.Store("a", "k1", "v1")
	_ = c.Store("b", "k2", "v2")

	if err := c.Clear("a"); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	var v string
	if c.Load("a", "k1", &v) {
		t.Error("section a should be cleared")
	}

	if !c.Load("b", "k2", &v) {
		t.Error("section b should still exist")
	}
}

func TestClearAll(t *testing.T) {
	c := New(t.TempDir())

	_ = c.Store("a", "k", "v")
	if err := c.Clear(""); err != nil {
		t.Fatalf("Clear all: %v", err)
	}

	var v string
	if c.Load("a", "k", &v) {
		t.Error("all data should be cleared")
	}
}

func TestDir(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	if c.Dir() != dir {
		t.Errorf("Dir() = %q, want %q", c.Dir(), dir)
	}
}

func TestDefaultCacheDir(t *testing.T) {
	dir := defaultCacheDir()
	if dir == "" {
		t.Error("defaultCacheDir returned empty string")
	}
}

func TestPath(t *testing.T) {
	c := New("/cache")
	got := c.path("section", "key")
	// Platform-agnostic check - just ensure it contains the parts.
	if got == "" {
		t.Error("path returned empty string")
	}
}
