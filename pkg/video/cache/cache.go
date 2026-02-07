package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Cache provides filesystem-based caching using XDG directory conventions.
type Cache struct {
	dir string
}

// New creates a new Cache. If dir is empty, the default XDG cache directory is used.
func New(dir string) *Cache {
	if dir == "" {
		dir = defaultCacheDir()
	}

	return &Cache{dir: dir}
}

// Store saves data as JSON to the cache under the given section/key.
func (c *Cache) Store(section, key string, data any) error {
	path := c.path(section, key)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("cache: mkdir: %w", err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("cache: marshal: %w", err)
	}

	return os.WriteFile(path, b, 0o600)
}

// Load reads cached JSON data into dst. Returns false if not found.
func (c *Cache) Load(section, key string, dst any) bool {
	path := c.path(section, key)

	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	return json.Unmarshal(b, dst) == nil
}

// Remove deletes a cached entry.
func (c *Cache) Remove(section, key string) error {
	return os.Remove(c.path(section, key))
}

// Clear removes all cached data for a section, or all data if section is empty.
func (c *Cache) Clear(section string) error {
	if section == "" {
		return os.RemoveAll(c.dir)
	}

	return os.RemoveAll(filepath.Join(c.dir, section))
}

// Dir returns the cache directory path.
func (c *Cache) Dir() string {
	return c.dir
}

func (c *Cache) path(section, key string) string {
	return filepath.Join(c.dir, section, key+".json")
}

func defaultCacheDir() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "omni-video")
	}

	switch runtime.GOOS {
	case "windows":
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return filepath.Join(dir, "omni-video", "cache")
		}

		home, _ := os.UserHomeDir()

		return filepath.Join(home, "AppData", "Local", "omni-video", "cache")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".cache", "omni-video")
	}
}
