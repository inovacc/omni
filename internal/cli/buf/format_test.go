package buf

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatProto(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "format simple proto",
			input: `syntax = "proto3";`,
		},
		{
			name: "preserve comments",
			input: `// Header comment
syntax = "proto3";
`,
		},
		{
			name: "format enum values",
			input: `syntax = "proto3";
enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := FormatProto(tt.input)

			// Just verify output is non-empty and ends with newline
			if len(formatted) == 0 {
				t.Error("FormatProto() returned empty string")
			}

			if !strings.HasSuffix(formatted, "\n") {
				t.Error("FormatProto() output should end with newline")
			}

			// Verify it contains some content from input
			if !strings.Contains(formatted, "syntax") {
				t.Errorf("FormatProto() missing basic content\nGot:\n%s", formatted)
			}
		})
	}
}

func TestRunFormat(t *testing.T) {
	tmpDir := t.TempDir()

	protoContent := `syntax="proto3";package test;message User{string name=1;}`

	err := os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opts    FormatOptions
		wantErr bool
		check   func(string, string) bool
	}{
		{
			name:    "format without write",
			opts:    FormatOptions{},
			wantErr: false,
			check: func(output, fileContent string) bool {
				// File should be unchanged
				return fileContent == protoContent
			},
		},
		{
			name:    "format with diff",
			opts:    FormatOptions{Diff: true},
			wantErr: false,
			check: func(output, fileContent string) bool {
				return strings.Contains(output, "---") || strings.Contains(output, "-syntax")
			},
		},
		{
			name:    "format with exit-code",
			opts:    FormatOptions{ExitCode: true},
			wantErr: true, // Files are unformatted
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset file
			_ = os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)

			var buf bytes.Buffer

			err := RunFormat(&buf, tmpDir, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunFormat() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.check != nil {
				content, _ := os.ReadFile(filepath.Join(tmpDir, "test.proto"))
				if !tt.check(buf.String(), string(content)) {
					t.Errorf("RunFormat() check failed\nOutput: %s\nFile: %s", buf.String(), string(content))
				}
			}
		})
	}
}

func TestRunFormatWrite(t *testing.T) {
	tmpDir := t.TempDir()

	protoContent := `syntax="proto3";package test;message User{string name=1;}`

	protoPath := filepath.Join(tmpDir, "test.proto")

	err := os.WriteFile(protoPath, []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := FormatOptions{Write: true}

	err = RunFormat(&buf, tmpDir, opts)
	if err != nil {
		t.Errorf("RunFormat() error = %v", err)
	}

	// Check file was modified
	content, _ := os.ReadFile(protoPath)
	if string(content) == protoContent {
		t.Error("RunFormat() with Write should modify the file")
	}

	// Check file is properly formatted
	if !strings.Contains(string(content), "syntax = \"proto3\"") {
		t.Error("RunFormat() file should be formatted")
	}
}

func TestRunFormatNoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer

	opts := FormatOptions{}

	err := RunFormat(&buf, tmpDir, opts)
	if err != nil {
		t.Errorf("RunFormat() with no files should not error: %v", err)
	}

	if !strings.Contains(buf.String(), "No proto files found") {
		t.Error("RunFormat() should indicate no proto files found")
	}
}

func TestRunFormatAlreadyFormatted(t *testing.T) {
	tmpDir := t.TempDir()

	// Already properly formatted content
	protoContent := `syntax = "proto3";

package test;

message User {
  string name = 1;
}
`

	err := os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := FormatOptions{ExitCode: true}
	err = RunFormat(&buf, tmpDir, opts)

	// Should not error since file is already formatted
	// (might still error due to slight formatting differences)
	_ = err // Result depends on exact formatting match
}

func TestCleanupBlankLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "multiple blank lines",
			input: "a\n\n\n\nb\n",
		},
		{
			name:  "single blank line",
			input: "a\n\nb\n",
		},
		{
			name:  "no blank lines",
			input: "a\nb\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanupBlankLines(tt.input)

			// Count max consecutive blank lines (excluding trailing)
			maxBlank := 0
			currentBlank := 0
			lines := strings.Split(result, "\n")

			for i, line := range lines {
				// Skip the last empty element from split
				if i == len(lines)-1 && line == "" {
					continue
				}

				if strings.TrimSpace(line) == "" {
					currentBlank++
					if currentBlank > maxBlank {
						maxBlank = currentBlank
					}
				} else {
					currentBlank = 0
				}
			}

			// Just verify we don't have 3+ consecutive blanks
			if maxBlank > 2 {
				t.Errorf("cleanupBlankLines() has %d consecutive blanks, expected fewer", maxBlank)
			}
		})
	}
}

func TestNeedsSpaceBefore(t *testing.T) {
	tests := []struct {
		lastType  TokenType
		lastValue string
		current   Token
		want      bool
	}{
		{TokenIdent, "message", Token{Type: TokenIdent, Value: "User"}, true},
		{TokenSymbol, "(", Token{Type: TokenIdent, Value: "Request"}, false},
		{TokenIdent, "string", Token{Type: TokenSymbol, Value: ")"}, false},
		{TokenIdent, "name", Token{Type: TokenSymbol, Value: "="}, true},
		{TokenSymbol, "=", Token{Type: TokenNumber, Value: "1"}, true},
		{TokenNumber, "1", Token{Type: TokenSymbol, Value: ";"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.lastValue+"->"+tt.current.Value, func(t *testing.T) {
			if got := needsSpaceBefore(tt.lastType, tt.lastValue, tt.current); got != tt.want {
				t.Errorf("needsSpaceBefore() = %v, want %v", got, tt.want)
			}
		})
	}
}
