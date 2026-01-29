package chmod

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRunChmod(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping chmod tests on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "chmod_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("octal mode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file1.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		var buf bytes.Buffer

		err := RunChmod(&buf, []string{"755", file}, ChmodOptions{})
		if err != nil {
			t.Fatalf("RunChmod() error = %v", err)
		}

		info, _ := os.Stat(file)
		if info.Mode().Perm() != 0755 {
			t.Errorf("RunChmod() mode = %o, want 0755", info.Mode().Perm())
		}
	})

	t.Run("symbolic mode u+x", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file2.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		var buf bytes.Buffer

		err := RunChmod(&buf, []string{"u+x", file}, ChmodOptions{})
		if err != nil {
			t.Fatalf("RunChmod() error = %v", err)
		}

		info, _ := os.Stat(file)
		if info.Mode().Perm()&0100 == 0 {
			t.Errorf("RunChmod() should set user execute bit")
		}
	})

	t.Run("symbolic mode go-w", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file3.txt")
		_ = os.WriteFile(file, []byte("content"), 0666)

		var buf bytes.Buffer

		err := RunChmod(&buf, []string{"go-w", file}, ChmodOptions{})
		if err != nil {
			t.Fatalf("RunChmod() error = %v", err)
		}

		info, _ := os.Stat(file)
		if info.Mode().Perm()&0022 != 0 {
			t.Errorf("RunChmod() should clear group/other write bits")
		}
	})

	t.Run("symbolic mode a=r", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file4.txt")
		_ = os.WriteFile(file, []byte("content"), 0777)

		var buf bytes.Buffer

		err := RunChmod(&buf, []string{"a=r", file}, ChmodOptions{})
		if err != nil {
			t.Fatalf("RunChmod() error = %v", err)
		}

		info, _ := os.Stat(file)
		if info.Mode().Perm() != 0444 {
			t.Errorf("RunChmod() mode = %o, want 0444", info.Mode().Perm())
		}
	})

	t.Run("verbose mode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file5.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		var buf bytes.Buffer

		err := RunChmod(&buf, []string{"755", file}, ChmodOptions{Verbose: true})
		if err != nil {
			t.Fatalf("RunChmod() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunChmod() -v should produce output")
		}
	})

	t.Run("recursive mode", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "subdir")
		_ = os.Mkdir(dir, 0755)
		file := filepath.Join(dir, "nested.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		var buf bytes.Buffer

		err := RunChmod(&buf, []string{"700", dir}, ChmodOptions{Recursive: true})
		if err != nil {
			t.Fatalf("RunChmod() -R error = %v", err)
		}

		info, _ := os.Stat(file)
		if info.Mode().Perm() != 0700 {
			t.Errorf("RunChmod() -R nested file mode = %o, want 0700", info.Mode().Perm())
		}
	})

	t.Run("reference file", func(t *testing.T) {
		ref := filepath.Join(tmpDir, "ref.txt")
		target := filepath.Join(tmpDir, "target.txt")
		_ = os.WriteFile(ref, []byte("ref"), 0755)
		_ = os.WriteFile(target, []byte("target"), 0644)

		var buf bytes.Buffer

		err := RunChmod(&buf, []string{"ignored", target}, ChmodOptions{Reference: ref})
		if err != nil {
			t.Fatalf("RunChmod() --reference error = %v", err)
		}

		info, _ := os.Stat(target)
		if info.Mode().Perm() != 0755 {
			t.Errorf("RunChmod() --reference mode = %o, want 0755", info.Mode().Perm())
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunChmod(&buf, []string{"755"}, ChmodOptions{})
		if err == nil {
			t.Error("RunChmod() expected error for missing operand")
		}
	})

	t.Run("invalid octal mode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file6.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		var buf bytes.Buffer

		err := RunChmod(&buf, []string{"999", file}, ChmodOptions{})
		if err == nil {
			t.Error("RunChmod() expected error for invalid octal mode")
		}
	})
}

func TestIsOctalMode(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"755", true},
		{"0644", true},
		{"777", true},
		{"000", true},
		{"u+x", false},
		{"go-w", false},
		{"a=rw", false},
		{"", false},
		{"888", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isOctalMode(tt.input)
			if got != tt.want {
				t.Errorf("isOctalMode(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestApplySymbolicPart(t *testing.T) {
	tests := []struct {
		name     string
		initial  fs.FileMode
		part     string
		expected fs.FileMode
	}{
		{"u+x", 0644, "u+x", 0744},
		{"u-w", 0644, "u-w", 0444},
		{"g+w", 0644, "g+w", 0664},
		{"o+r", 0640, "o+r", 0644},
		{"a+x", 0644, "a+x", 0755},
		{"go-rwx", 0777, "go-rwx", 0700},
		{"u=rwx", 0000, "u=rwx", 0700},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applySymbolicPart(tt.initial, tt.part)
			if got != tt.expected {
				t.Errorf("applySymbolicPart(%o, %q) = %o, want %o", tt.initial, tt.part, got, tt.expected)
			}
		})
	}
}

func TestParseSymbolicMode(t *testing.T) {
	tests := []struct {
		mode     string
		initial  fs.FileMode
		expected fs.FileMode
	}{
		{"u+x", 0644, 0744},
		{"u+x,g+x", 0644, 0754},
		{"a=r", 0777, 0444},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			fn, err := parseSymbolicMode(tt.mode)
			if err != nil {
				t.Fatalf("parseSymbolicMode() error = %v", err)
			}

			got := fn(tt.initial)
			if got != tt.expected {
				t.Errorf("parseSymbolicMode(%q)(%o) = %o, want %o", tt.mode, tt.initial, got, tt.expected)
			}
		})
	}
}
