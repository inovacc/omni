package hashutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashString(t *testing.T) {
	tests := []struct {
		name string
		data string
		algo Algorithm
		want string
	}{
		{"md5 test", "test", MD5, "098f6bcd4621d373cade4e832627b4f6"},
		{"sha1 test", "test", SHA1, "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"},
		{"sha256 test", "test", SHA256, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"},
		{"sha256 empty", "", SHA256, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashString(tt.data, tt.algo)
			if got != tt.want {
				t.Errorf("HashString(%q, %s) = %v, want %v", tt.data, tt.algo, got, tt.want)
			}
		})
	}
}

func TestHashBytes(t *testing.T) {
	got := HashBytes([]byte("test"), MD5)
	want := "098f6bcd4621d373cade4e832627b4f6"
	if got != want {
		t.Errorf("HashBytes() = %v, want %v", got, want)
	}
}

func TestHashReader(t *testing.T) {
	r := strings.NewReader("test")
	got, err := HashReader(r, SHA256)
	if err != nil {
		t.Fatalf("HashReader() error = %v", err)
	}
	want := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if got != want {
		t.Errorf("HashReader() = %v, want %v", got, want)
	}
}

func TestHashFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hashutil_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("md5", func(t *testing.T) {
		got, err := HashFile(testFile, MD5)
		if err != nil {
			t.Fatalf("HashFile() error = %v", err)
		}
		if got != "098f6bcd4621d373cade4e832627b4f6" {
			t.Errorf("HashFile() = %v", got)
		}
	})

	t.Run("sha256", func(t *testing.T) {
		got, err := HashFile(testFile, SHA256)
		if err != nil {
			t.Fatalf("HashFile() error = %v", err)
		}
		if got != "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" {
			t.Errorf("HashFile() = %v", got)
		}
	})

	t.Run("sha512", func(t *testing.T) {
		got, err := HashFile(testFile, SHA512)
		if err != nil {
			t.Fatalf("HashFile() error = %v", err)
		}
		if len(got) != 128 {
			t.Errorf("SHA512 hash length = %d, want 128", len(got))
		}
	})

	t.Run("nonexistent", func(t *testing.T) {
		_, err := HashFile("/nonexistent/file.txt", SHA256)
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}

func TestHashConsistency(t *testing.T) {
	h1 := HashString("consistent", SHA256)
	h2 := HashString("consistent", SHA256)
	if h1 != h2 {
		t.Error("hashing the same data should produce consistent results")
	}
}

func TestDefaultAlgorithm(t *testing.T) {
	// Unknown algorithm should fall back to SHA256
	got := HashString("test", Algorithm("unknown"))
	want := HashString("test", SHA256)
	if got != want {
		t.Errorf("unknown algo = %v, want sha256 default = %v", got, want)
	}
}
