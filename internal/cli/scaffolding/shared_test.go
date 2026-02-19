package scaffolding

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestWriteTemplate(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test.txt"

	err := WriteTemplate(fs, path, "Hello {{.Name}}", struct{ Name string }{"World"})
	if err != nil {
		t.Fatalf("WriteTemplate() error = %v", err)
	}

	data, err := afero.ReadFile(fs, path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(data) != "Hello World" {
		t.Errorf("got %q, want %q", string(data), "Hello World")
	}
}

func TestWriteTemplateInvalidTemplate(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := WriteTemplate(fs, "/test.txt", "{{.Invalid", nil)
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestWriteLicense(t *testing.T) {
	tests := []struct {
		name     string
		license  string
		wantErr  bool
		contains string
	}{
		{"MIT", "MIT", false, "MIT License"},
		{"Apache", "Apache-2.0", false, "Apache License"},
		{"BSD", "BSD-3", false, "BSD"},
		{"unknown", "unknown", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			err := WriteLicense(fs, "/LICENSE", tt.license, "Test Author")
			if (err != nil) != tt.wantErr {
				t.Fatalf("WriteLicense() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				data, err := afero.ReadFile(fs, "/LICENSE")
				if err != nil {
					t.Fatalf("ReadFile() error = %v", err)
				}
				if !strings.Contains(string(data), tt.contains) {
					t.Errorf("LICENSE doesn't contain %q", tt.contains)
				}
			}
		})
	}
}
