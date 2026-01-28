package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTr(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		set1   string
		set2   string
		opts   TrOptions
		expect string
	}{
		{
			name:   "lowercase to uppercase",
			input:  "hello",
			set1:   "a-z",
			set2:   "A-Z",
			opts:   TrOptions{},
			expect: "HELLO",
		},
		{
			name:   "delete characters",
			input:  "hello world",
			set1:   " ",
			set2:   "",
			opts:   TrOptions{Delete: true},
			expect: "helloworld",
		},
		{
			name:   "squeeze repeats",
			input:  "helllo   wooorld",
			set1:   "lo ",
			set2:   "",
			opts:   TrOptions{Squeeze: true},
			expect: "helo world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunTr(&buf, strings.NewReader(tt.input), tt.set1, tt.set2, tt.opts)
			if err != nil {
				t.Fatalf("RunTr() error = %v", err)
			}

			result := strings.TrimSpace(buf.String())
			if result != tt.expect {
				t.Errorf("RunTr() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestRunCut(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cut_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("cut fields", func(t *testing.T) {
		file := filepath.Join(tmpDir, "data.csv")

		content := "a,b,c\n1,2,3\nx,y,z"
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCut(&buf, []string{file}, CutOptions{Fields: "2", Delimiter: ","})
		if err != nil {
			t.Fatalf("RunCut() error = %v", err)
		}

		expected := "b\n2\ny"
		if strings.TrimSpace(buf.String()) != expected {
			t.Errorf("RunCut() = %v, want %v", buf.String(), expected)
		}
	})

	t.Run("cut characters", func(t *testing.T) {
		file := filepath.Join(tmpDir, "chars.txt")

		content := "hello\nworld"
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCut(&buf, []string{file}, CutOptions{Characters: "1-3"})
		if err != nil {
			t.Fatalf("RunCut() error = %v", err)
		}

		expected := "hel\nwor"
		if strings.TrimSpace(buf.String()) != expected {
			t.Errorf("RunCut() = %v, want %v", buf.String(), expected)
		}
	})
}

func TestRunNl(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nl_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "lines.txt")

	content := "first\nsecond\nthird"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunNl(&buf, []string{file}, NlOptions{})
	if err != nil {
		t.Fatalf("RunNl() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1") || !strings.Contains(output, "first") {
		t.Errorf("RunNl() missing line numbers: %v", output)
	}
}

func TestRunUniq(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "uniq_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("remove duplicates", func(t *testing.T) {
		file := filepath.Join(tmpDir, "dups.txt")

		content := "apple\nbanana\ncherry"
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, []string{file}, UniqOptions{})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunUniq() = %d lines, want 3", len(lines))
		}
	})

	t.Run("count occurrences", func(t *testing.T) {
		file := filepath.Join(tmpDir, "count.txt")

		content := "a\na\na\nb\nb\n"
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, []string{file}, UniqOptions{Count: true})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "3") {
			t.Errorf("RunUniq() missing count for 3 consecutive 'a': %v", output)
		}
	})

	t.Run("only duplicates", func(t *testing.T) {
		file := filepath.Join(tmpDir, "onlydups.txt")

		content := "unique\ndup\ndup\nanother"
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, []string{file}, UniqOptions{Repeated: true})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "dup" {
			t.Errorf("RunUniq() = %v, want 'dup'", output)
		}
	})
}

func TestRunPaste(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "paste_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("a\nb\nc"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(file2, []byte("1\n2\n3"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunPaste(&buf, []string{file1, file2}, PasteOptions{})
	if err != nil {
		t.Fatalf("RunPaste() error = %v", err)
	}

	expected := "a\t1\nb\t2\nc\t3"
	if strings.TrimSpace(buf.String()) != expected {
		t.Errorf("RunPaste() = %v, want %v", buf.String(), expected)
	}
}

func TestRunTac(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tac_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "reverse.txt")

	content := "first\nsecond\nthird"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunTac(&buf, []string{file}, TacOptions{})
	if err != nil {
		t.Fatalf("RunTac() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 || lines[0] != "third" || lines[2] != "first" {
		t.Errorf("RunTac() = %v, want reversed order", lines)
	}
}

func TestRunFold(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fold_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "longline.txt")

	content := "this is a very long line that should be wrapped"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunFold(&buf, []string{file}, FoldOptions{Width: 10})
	if err != nil {
		t.Fatalf("RunFold() error = %v", err)
	}

	lines := strings.SplitSeq(buf.String(), "\n")
	for line := range lines {
		if len(line) > 10 && !strings.HasSuffix(line, "\n") {
			t.Errorf("RunFold() line too long: %v", line)
		}
	}
}

func TestRunColumn(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "column_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "table.txt")

	content := "name,age,city\njohn,30,nyc\njane,25,la"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunColumn(&buf, []string{file}, ColumnOptions{Separator: ",", Table: true})
	if err != nil {
		t.Fatalf("RunColumn() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name") || !strings.Contains(output, "john") {
		t.Errorf("RunColumn() missing data: %v", output)
	}
}

func TestRunJoin(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "join_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("1 apple\n2 banana\n3 cherry"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(file2, []byte("1 red\n2 yellow\n3 red"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunJoin(&buf, []string{file1, file2}, JoinOptions{})
	if err != nil {
		t.Fatalf("RunJoin() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "apple") || !strings.Contains(output, "red") {
		t.Errorf("RunJoin() missing joined data: %v", output)
	}
}
