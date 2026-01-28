package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHash(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hash_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		args      []string
		opts      HashOptions
		contains  string
		wantError bool
	}{
		{
			name:     "sha256 default",
			args:     []string{testFile},
			opts:     HashOptions{},
			contains: "test.txt", // just check filename appears
		},
		{
			name:     "md5",
			args:     []string{testFile},
			opts:     HashOptions{Algorithm: "md5"},
			contains: testFile,
		},
		{
			name:     "sha512",
			args:     []string{testFile},
			opts:     HashOptions{Algorithm: "sha512"},
			contains: testFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunHash(&buf, tt.args, tt.opts)
			if (err != nil) != tt.wantError {
				t.Errorf("RunHash() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.contains != "" && !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("RunHash() output = %v, want contains %v", buf.String(), tt.contains)
			}
		})
	}
}

func TestRunMD5Sum(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "md5_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunMD5Sum(&buf, []string{testFile}, HashOptions{})
	if err != nil {
		t.Fatalf("RunMD5Sum() error = %v", err)
	}

	// MD5 of "test" is 098f6bcd4621d373cade4e832627b4f6
	if !strings.Contains(buf.String(), "098f6bcd4621d373cade4e832627b4f6") {
		t.Errorf("RunMD5Sum() got = %v, want hash of 'test'", buf.String())
	}
}

func TestRunSHA256Sum(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sha256_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunSHA256Sum(&buf, []string{testFile}, HashOptions{})
	if err != nil {
		t.Fatalf("RunSHA256Sum() error = %v", err)
	}

	// SHA256 of "test" is 9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
	if !strings.Contains(buf.String(), "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08") {
		t.Errorf("RunSHA256Sum() got = %v, want hash of 'test'", buf.String())
	}
}
