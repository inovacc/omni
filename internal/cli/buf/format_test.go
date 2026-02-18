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
		want  string
	}{
		{
			name:  "simple syntax statement",
			input: "syntax = \"proto3\";\n",
			want:  "syntax = \"proto3\";\n",
		},
		{
			name:  "collapse multiple blank lines",
			input: "syntax = \"proto3\";\n\n\n\npackage test;\n",
			want:  "syntax = \"proto3\";\n\npackage test;\n",
		},
		{
			name:  "preserve single blank line",
			input: "syntax = \"proto3\";\n\npackage test;\n",
			want:  "syntax = \"proto3\";\n\npackage test;\n",
		},
		{
			name: "format message with 2-space indent",
			input: `syntax = "proto3";
package test;
message User {
    string id = 1;
    string name = 2;
}
`,
			want: `syntax = "proto3";
package test;
message User {
  string id = 1;
  string name = 2;
}
`,
		},
		{
			name:  "expand compressed proto",
			input: `syntax = "proto3"; package test; message User { string id = 1; }`,
			want: `syntax = "proto3";
package test;
message User {
  string id = 1;
}
`,
		},
		{
			name: "format enum",
			input: `syntax = "proto3";
package test;
message User {
  string id = 1;
}
enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
}
`,
			want: `syntax = "proto3";
package test;
message User {
  string id = 1;
}
enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
}
`,
		},
		{
			name: "format service with rpc",
			input: `syntax = "proto3";
package test;
message User {}
service UserService {
  rpc GetUser(User) returns (User);
}
`,
			want: `syntax = "proto3";
package test;
message User {}
service UserService {
  rpc GetUser(User) returns (User);
}
`,
		},
		{
			name: "preserve trailing comment",
			input: `syntax = "proto3"; // version comment
package test;
`,
			want: `syntax = "proto3"; // version comment
package test;
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatProto(tt.input)
			if got != tt.want {
				t.Errorf("FormatProto() mismatch\ngot:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestRunFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Content with multiple blank lines
	protoContent := "syntax = \"proto3\";\n\n\n\npackage test;\n"

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
				return strings.Contains(output, "---")
			},
		},
		{
			name:    "format with exit-code",
			opts:    FormatOptions{ExitCode: true},
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

	protoContent := "syntax = \"proto3\";\n\n\n\npackage test;\n"

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

	content, _ := os.ReadFile(protoPath)
	if string(content) == protoContent {
		t.Error("RunFormat() with Write should modify the file")
	}

	if strings.Contains(string(content), "\n\n\n") {
		t.Error("RunFormat() should collapse multiple blank lines")
	}
}

func TestRunFormatNoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer

	err := RunFormat(&buf, tmpDir, FormatOptions{})
	if err != nil {
		t.Errorf("RunFormat() with no files should not error: %v", err)
	}

	if !strings.Contains(buf.String(), "No proto files found") {
		t.Error("RunFormat() should indicate no proto files found")
	}
}

func TestRunFormatAlreadyFormatted(t *testing.T) {
	tmpDir := t.TempDir()

	protoContent := "syntax = \"proto3\";\n\npackage test;\n"

	err := os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := FormatOptions{ExitCode: true}
	err = RunFormat(&buf, tmpDir, opts)

	if err != nil {
		t.Errorf("RunFormat() should not error for already formatted file: %v", err)
	}
}

func TestCleanupBlankLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "multiple blank lines collapsed",
			input: "a\n\n\n\nb\n",
			want:  "a\n\nb\n",
		},
		{
			name:  "single blank line preserved",
			input: "a\n\nb\n",
			want:  "a\n\nb\n",
		},
		{
			name:  "no blank lines unchanged",
			input: "a\nb\n",
			want:  "a\nb\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanupBlankLines(tt.input)
			if got != tt.want {
				t.Errorf("cleanupBlankLines() = %q, want %q", got, tt.want)
			}
		})
	}
}
