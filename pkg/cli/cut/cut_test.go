package cut

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCut(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cut_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("cut fields with tab delimiter", func(t *testing.T) {
		file := filepath.Join(tmpDir, "tab.txt")
		if err := os.WriteFile(file, []byte("one\ttwo\tthree\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCut(&buf, []string{file}, CutOptions{Fields: "2"})
		if err != nil {
			t.Fatalf("RunCut() error = %v", err)
		}

		if buf.String() != "two\n" {
			t.Errorf("RunCut() = %q, want 'two\\n'", buf.String())
		}
	})

	t.Run("cut fields with custom delimiter", func(t *testing.T) {
		file := filepath.Join(tmpDir, "comma.txt")
		if err := os.WriteFile(file, []byte("one,two,three\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCut(&buf, []string{file}, CutOptions{Fields: "1,3", Delimiter: ","})
		if err != nil {
			t.Fatalf("RunCut() error = %v", err)
		}

		if buf.String() != "one,three\n" {
			t.Errorf("RunCut() = %q, want 'one,three\\n'", buf.String())
		}
	})

	t.Run("cut bytes", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes.txt")
		if err := os.WriteFile(file, []byte("hello world\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCut(&buf, []string{file}, CutOptions{Bytes: "1-5"})
		if err != nil {
			t.Fatalf("RunCut() error = %v", err)
		}

		if buf.String() != "hello\n" {
			t.Errorf("RunCut() = %q, want 'hello\\n'", buf.String())
		}
	})

	t.Run("cut characters", func(t *testing.T) {
		file := filepath.Join(tmpDir, "chars.txt")
		if err := os.WriteFile(file, []byte("hello\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCut(&buf, []string{file}, CutOptions{Characters: "2,4"})
		if err != nil {
			t.Fatalf("RunCut() error = %v", err)
		}

		if buf.String() != "el\n" {
			t.Errorf("RunCut() = %q, want 'el\\n'", buf.String())
		}
	})

	t.Run("field range", func(t *testing.T) {
		file := filepath.Join(tmpDir, "range.txt")
		if err := os.WriteFile(file, []byte("a,b,c,d,e\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCut(&buf, []string{file}, CutOptions{Fields: "2-4", Delimiter: ","})
		if err != nil {
			t.Fatalf("RunCut() error = %v", err)
		}

		if buf.String() != "b,c,d\n" {
			t.Errorf("RunCut() = %q, want 'b,c,d\\n'", buf.String())
		}
	})

	t.Run("complement fields", func(t *testing.T) {
		file := filepath.Join(tmpDir, "comp.txt")
		if err := os.WriteFile(file, []byte("a,b,c,d\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCut(&buf, []string{file}, CutOptions{Fields: "2", Delimiter: ",", Complement: true})
		if err != nil {
			t.Fatalf("RunCut() error = %v", err)
		}

		if buf.String() != "a,c,d\n" {
			t.Errorf("RunCut() = %q, want 'a,c,d\\n'", buf.String())
		}
	})

	t.Run("only delimited lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "only.txt")
		if err := os.WriteFile(file, []byte("a,b,c\nno delimiter\nx,y,z\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCut(&buf, []string{file}, CutOptions{Fields: "2", Delimiter: ",", OnlyDelim: true})
		if err != nil {
			t.Fatalf("RunCut() error = %v", err)
		}

		if buf.String() != "b\ny\n" {
			t.Errorf("RunCut() = %q, want 'b\\ny\\n'", buf.String())
		}
	})

	t.Run("custom output delimiter", func(t *testing.T) {
		file := filepath.Join(tmpDir, "outdelim.txt")
		if err := os.WriteFile(file, []byte("a,b,c\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCut(&buf, []string{file}, CutOptions{Fields: "1-3", Delimiter: ",", OutputDelim: ":"})
		if err != nil {
			t.Fatalf("RunCut() error = %v", err)
		}

		if buf.String() != "a:b:c\n" {
			t.Errorf("RunCut() = %q, want 'a:b:c\\n'", buf.String())
		}
	})

	t.Run("no selection specified", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunCut(&buf, []string{}, CutOptions{})
		if err == nil {
			t.Error("RunCut() expected error for no selection")
		}
	})

	t.Run("multiple selection types", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunCut(&buf, []string{}, CutOptions{Fields: "1", Bytes: "1"})
		if err == nil {
			t.Error("RunCut() expected error for multiple selection types")
		}
	})

	t.Run("delimiter too long", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunCut(&buf, []string{}, CutOptions{Fields: "1", Delimiter: "ab"})
		if err == nil {
			t.Error("RunCut() expected error for delimiter > 1 char")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		// Implementation prints to stderr but continues
		_ = RunCut(&buf, []string{"/nonexistent/file.txt"}, CutOptions{Fields: "1"})
	})
}

func TestParseRanges(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		maxVal   int
		expected []int
		wantErr  bool
	}{
		{"single number", "3", 10, []int{3}, false},
		{"multiple numbers", "1,3,5", 10, []int{1, 3, 5}, false},
		{"range", "2-5", 10, []int{2, 3, 4, 5}, false},
		{"open start range", "-3", 10, []int{1, 2, 3}, false},
		{"open end range", "8-", 10, []int{8, 9, 10}, false},
		{"mixed", "1,3-5,8", 10, []int{1, 3, 4, 5, 8}, false},
		{"invalid number", "abc", 10, nil, true},
		{"invalid range", "a-b", 10, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRanges(tt.spec, tt.maxVal)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseRanges(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) != len(tt.expected) {
					t.Errorf("parseRanges(%q) = %v, want %v", tt.spec, result, tt.expected)
					return
				}

				for i := range result {
					if result[i] != tt.expected[i] {
						t.Errorf("parseRanges(%q)[%d] = %d, want %d", tt.spec, i, result[i], tt.expected[i])
					}
				}
			}
		})
	}
}

func TestComplementRanges(t *testing.T) {
	tests := []struct {
		name     string
		ranges   []int
		maxVal   int
		expected []int
	}{
		{"single excluded", []int{2}, 5, []int{1, 3, 4, 5}},
		{"multiple excluded", []int{1, 3, 5}, 5, []int{2, 4}},
		{"none excluded", []int{}, 3, []int{1, 2, 3}},
		{"all excluded", []int{1, 2, 3}, 3, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := complementRanges(tt.ranges, tt.maxVal)

			if len(result) != len(tt.expected) {
				t.Errorf("complementRanges() = %v, want %v", result, tt.expected)
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("complementRanges()[%d] = %d, want %d", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
