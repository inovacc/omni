package cache_test

import (
	"os"
	"testing"

	"github.com/inovacc/omni/pkg/video/cache"
)

func TestNew_API(t *testing.T) {
	tmp := t.TempDir()
	c := cache.New(tmp)
	if c == nil {
		t.Fatal("cache.New() returned nil")
	}
}

func TestStoreLoad_API(t *testing.T) {
	tmp := t.TempDir()
	c := cache.New(tmp)

	type entry struct {
		Value string `json:"value"`
	}

	if err := c.Store("test", "key1", entry{Value: "hello"}); err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	var got entry
	if !c.Load("test", "key1", &got) {
		t.Fatal("Load() returned false after Store()")
	}
	if got.Value != "hello" {
		t.Errorf("Load() got Value=%q, want %q", got.Value, "hello")
	}
}

func TestRemove_API(t *testing.T) {
	tmp := t.TempDir()
	c := cache.New(tmp)

	_ = c.Store("sec", "k", "data")
	if err := c.Remove("sec", "k"); err != nil {
		if !os.IsNotExist(err) {
			t.Fatalf("Remove() error = %v", err)
		}
	}

	var dst string
	if c.Load("sec", "k", &dst) {
		t.Fatal("Load() returned true after Remove()")
	}
}
