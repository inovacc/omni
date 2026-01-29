package dotenv

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDotenv(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dotenv_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("parse simple env file", func(t *testing.T) {
		file := filepath.Join(tmpDir, ".env1")
		_ = os.WriteFile(file, []byte("KEY1=value1\nKEY2=value2\n"), 0644)

		var buf bytes.Buffer

		err := RunDotenv(&buf, []string{file}, DotenvOptions{})
		if err != nil {
			t.Fatalf("RunDotenv() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "KEY1=value1") {
			t.Errorf("RunDotenv() missing KEY1: %s", output)
		}

		if !strings.Contains(output, "KEY2=value2") {
			t.Errorf("RunDotenv() missing KEY2: %s", output)
		}
	})

	t.Run("export format", func(t *testing.T) {
		file := filepath.Join(tmpDir, ".env2")
		_ = os.WriteFile(file, []byte("KEY=value\n"), 0644)

		var buf bytes.Buffer

		err := RunDotenv(&buf, []string{file}, DotenvOptions{Export: true})
		if err != nil {
			t.Fatalf("RunDotenv() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "export KEY=") {
			t.Errorf("RunDotenv() -e should use export: %s", output)
		}
	})

	t.Run("default to .env", func(t *testing.T) {
		// Change to temp dir
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)

		defer func() { _ = os.Chdir(origDir) }()

		_ = os.WriteFile(".env", []byte("DEFAULT=yes\n"), 0644)

		var buf bytes.Buffer

		err := RunDotenv(&buf, []string{}, DotenvOptions{})
		if err != nil {
			t.Fatalf("RunDotenv() error = %v", err)
		}

		if !strings.Contains(buf.String(), "DEFAULT=yes") {
			t.Errorf("RunDotenv() should use .env by default")
		}
	})

	t.Run("quiet mode on missing file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDotenv(&buf, []string{"/nonexistent/.env"}, DotenvOptions{Quiet: true})
		if err != nil {
			t.Fatalf("RunDotenv() quiet error = %v", err)
		}
	})
}

func TestParseDotenv(t *testing.T) {
	t.Run("basic key=value", func(t *testing.T) {
		input := strings.NewReader("KEY=value\n")

		vars, err := ParseDotenv(input, DotenvOptions{})
		if err != nil {
			t.Fatalf("ParseDotenv() error = %v", err)
		}

		if len(vars) != 1 || vars[0].Key != "KEY" || vars[0].Value != "value" {
			t.Errorf("ParseDotenv() = %v", vars)
		}
	})

	t.Run("skip comments", func(t *testing.T) {
		input := strings.NewReader("# comment\nKEY=value\n")
		vars, _ := ParseDotenv(input, DotenvOptions{})

		if len(vars) != 1 {
			t.Errorf("ParseDotenv() should skip comments: %d vars", len(vars))
		}
	})

	t.Run("skip empty lines", func(t *testing.T) {
		input := strings.NewReader("KEY1=v1\n\nKEY2=v2\n")
		vars, _ := ParseDotenv(input, DotenvOptions{})

		if len(vars) != 2 {
			t.Errorf("ParseDotenv() should skip empty lines: %d vars", len(vars))
		}
	})

	t.Run("double quoted value", func(t *testing.T) {
		input := strings.NewReader(`KEY="hello world"` + "\n")
		vars, _ := ParseDotenv(input, DotenvOptions{})

		if len(vars) != 1 || vars[0].Value != "hello world" {
			t.Errorf("ParseDotenv() double quote = %q", vars[0].Value)
		}
	})

	t.Run("single quoted value", func(t *testing.T) {
		input := strings.NewReader(`KEY='hello world'` + "\n")
		vars, _ := ParseDotenv(input, DotenvOptions{})

		if len(vars) != 1 || vars[0].Value != "hello world" {
			t.Errorf("ParseDotenv() single quote = %q", vars[0].Value)
		}
	})

	t.Run("export prefix", func(t *testing.T) {
		input := strings.NewReader("export KEY=value\n")
		vars, _ := ParseDotenv(input, DotenvOptions{})

		if len(vars) != 1 || vars[0].Key != "KEY" {
			t.Errorf("ParseDotenv() export prefix = %v", vars)
		}
	})

	t.Run("inline comment", func(t *testing.T) {
		input := strings.NewReader("KEY=value # inline comment\n")
		vars, _ := ParseDotenv(input, DotenvOptions{})

		if len(vars) != 1 || vars[0].Value != "value" {
			t.Errorf("ParseDotenv() inline comment = %q", vars[0].Value)
		}
	})

	t.Run("escape sequences in double quotes", func(t *testing.T) {
		input := strings.NewReader(`KEY="line1\nline2"` + "\n")
		vars, _ := ParseDotenv(input, DotenvOptions{})

		if len(vars) != 1 || !strings.Contains(vars[0].Value, "\n") {
			t.Errorf("ParseDotenv() escape = %q", vars[0].Value)
		}
	})

	t.Run("expand variables", func(t *testing.T) {
		input := strings.NewReader("BASE=/home\nFULL=$BASE/user\n")
		vars, _ := ParseDotenv(input, DotenvOptions{Expand: true})

		if len(vars) != 2 || vars[1].Value != "/home/user" {
			t.Errorf("ParseDotenv() expand = %q", vars[1].Value)
		}
	})
}

func TestParseDotenvLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantErr   bool
	}{
		{"simple", "KEY=value", "KEY", "value", false},
		{"with spaces", "KEY = value", "KEY", "value", false},
		{"empty value", "KEY=", "KEY", "", false},
		{"no equals", "NOEQUALS", "", "", true},
		{"empty key", "=value", "", "", true},
		{"export prefix", "export KEY=value", "KEY", "value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, err := parseDotenvLine(tt.line)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseDotenvLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if key != tt.wantKey {
					t.Errorf("parseDotenvLine() key = %q, want %q", key, tt.wantKey)
				}

				if value != tt.wantValue {
					t.Errorf("parseDotenvLine() value = %q, want %q", value, tt.wantValue)
				}
			}
		})
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"quoted"`, "quoted"},
		{`'single'`, "single"},
		{`unquoted`, "unquoted"},
		{`value # comment`, "value"},
		{`"has spaces"`, "has spaces"},
		{`   trimmed   `, "trimmed"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseValue(tt.input)
			if result != tt.expected {
				t.Errorf("parseValue(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadDotenv(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loaddotenv_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .env file
	file := filepath.Join(tmpDir, ".env")
	_ = os.WriteFile(file, []byte("TEST_VAR=test_value\n"), 0644)

	// Change to temp dir
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)

	defer func() { _ = os.Chdir(origDir) }()

	// Clear any existing var
	_ = os.Unsetenv("TEST_VAR")

	err = LoadDotenv()
	if err != nil {
		t.Fatalf("LoadDotenv() error = %v", err)
	}

	if os.Getenv("TEST_VAR") != "test_value" {
		t.Errorf("LoadDotenv() did not set TEST_VAR")
	}

	// Clean up
	_ = os.Unsetenv("TEST_VAR")
}

func TestLoadDotenvOverride(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "override_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, ".env")
	_ = os.WriteFile(file, []byte("OVERRIDE_VAR=new_value\n"), 0644)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)

	defer func() { _ = os.Chdir(origDir) }()

	// Set existing value
	_ = os.Setenv("OVERRIDE_VAR", "old_value")

	err = LoadDotenvOverride()
	if err != nil {
		t.Fatalf("LoadDotenvOverride() error = %v", err)
	}

	if os.Getenv("OVERRIDE_VAR") != "new_value" {
		t.Errorf("LoadDotenvOverride() should override: got %s", os.Getenv("OVERRIDE_VAR"))
	}

	// Clean up
	_ = os.Unsetenv("OVERRIDE_VAR")
}
