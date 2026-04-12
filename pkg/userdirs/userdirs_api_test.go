package userdirs_test

import (
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/userdirs"
)

func TestDownloadsDir_API(t *testing.T) {
	dir, err := userdirs.DownloadsDir()
	if err != nil {
		t.Fatalf("DownloadsDir() error = %v", err)
	}
	if dir == "" {
		t.Fatal("DownloadsDir() returned empty string")
	}
	if !strings.Contains(dir, "Downloads") {
		t.Errorf("DownloadsDir() = %q, expected to contain 'Downloads'", dir)
	}
}

func TestDocumentsDir_API(t *testing.T) {
	dir, err := userdirs.DocumentsDir()
	if err != nil {
		t.Fatalf("DocumentsDir() error = %v", err)
	}
	if dir == "" {
		t.Fatal("DocumentsDir() returned empty string")
	}
	if !strings.Contains(dir, "Documents") {
		t.Errorf("DocumentsDir() = %q, expected to contain 'Documents'", dir)
	}
}
