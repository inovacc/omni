package nethttp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNetscapeCookies(t *testing.T) {
	t.Run("valid cookies file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "cookies.txt")

		content := "# Netscape HTTP Cookie File\n" +
			"\n" +
			".example.com\tTRUE\t/\tFALSE\t1700000000\tsession_id\tabc123\n" +
			".secure.com\tTRUE\t/path\tTRUE\t1800000000\ttoken\txyz789\n" +
			".noexpire.com\tTRUE\t/\tFALSE\t0\tvisitor\tguest\n"

		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		cookies, err := LoadNetscapeCookies(path)
		if err != nil {
			t.Fatalf("LoadNetscapeCookies: %v", err)
		}

		if len(cookies) != 3 {
			t.Fatalf("got %d cookies, want 3", len(cookies))
		}

		// First cookie.
		c := cookies[0]
		if c.Domain != ".example.com" {
			t.Errorf("domain = %q, want %q", c.Domain, ".example.com")
		}
		if c.Path != "/" {
			t.Errorf("path = %q, want %q", c.Path, "/")
		}
		if c.Secure {
			t.Error("expected secure=false")
		}
		if c.Name != "session_id" {
			t.Errorf("name = %q, want %q", c.Name, "session_id")
		}
		if c.Value != "abc123" {
			t.Errorf("value = %q, want %q", c.Value, "abc123")
		}

		// Second cookie: secure flag.
		if !cookies[1].Secure {
			t.Error("expected secure=true for second cookie")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := LoadNetscapeCookies("/nonexistent/cookies.txt")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("malformed lines skipped", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "cookies.txt")

		content := "too\tfew\tfields\n" +
			".valid.com\tTRUE\t/\tFALSE\t0\tname\tvalue\n"

		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		cookies, err := LoadNetscapeCookies(path)
		if err != nil {
			t.Fatalf("LoadNetscapeCookies: %v", err)
		}

		if len(cookies) != 1 {
			t.Errorf("got %d cookies, want 1 (malformed line should be skipped)", len(cookies))
		}
	})
}
