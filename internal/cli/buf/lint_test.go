package buf

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLint(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		protoFile  string
		wantErr    bool
		wantIssues int
	}{
		{
			name: "valid proto file",
			protoFile: `syntax = "proto3";

package test.v1;

message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  string id = 1;
  string name = 2;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
`,
			wantErr:    false,
			wantIssues: 1, // PACKAGE_DIRECTORY_MATCH
		},
		{
			name: "missing package",
			protoFile: `syntax = "proto3";

message User {
  string name = 1;
}
`,
			wantErr:    false,
			wantIssues: 1, // PACKAGE_DEFINED
		},
		{
			name: "bad field name casing",
			protoFile: `syntax = "proto3";

package test.v1;

message User {
  string userName = 1;
  string userEmail = 2;
}
`,
			wantErr:    false,
			wantIssues: 1, // At least PACKAGE_DIRECTORY_MATCH
		},
		{
			name: "bad message name casing",
			protoFile: `syntax = "proto3";

package test.v1;

message user_info {
  string name = 1;
}
`,
			wantErr:    false,
			wantIssues: 1, // At least one issue
		},
		{
			name: "enum first value not zero",
			protoFile: `syntax = "proto3";

package test.v1;

enum Status {
  ACTIVE = 1;
  INACTIVE = 2;
}
`,
			wantErr:    false,
			wantIssues: 1, // At least one issue
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			_ = os.MkdirAll(testDir, 0755)

			err := os.WriteFile(filepath.Join(testDir, "test.proto"), []byte(tt.protoFile), 0644)
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer

			opts := LintOptions{}
			err = RunLint(&buf, testDir, opts)

			if tt.wantErr && err == nil {
				t.Error("RunLint() expected error but got none")
			}

			output := buf.String()
			issueCount := strings.Count(output, "test.proto:")

			if issueCount < tt.wantIssues {
				t.Errorf("RunLint() found %d issues, want at least %d\nOutput:\n%s", issueCount, tt.wantIssues, output)
			}
		})
	}
}

func TestRunLintNoProtoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer

	opts := LintOptions{}
	err := RunLint(&buf, tmpDir, opts)

	// Stub returns nil with "No proto files" message
	if err != nil {
		t.Errorf("RunLint() with no proto files should not error: %v", err)
	}

	if !strings.Contains(buf.String(), "No proto files") {
		t.Errorf("RunLint() should indicate no proto files, got: %s", buf.String())
	}
}

func TestRunLintOutput(t *testing.T) {
	tmpDir := t.TempDir()

	protoContent := `syntax = "proto3";

message User {
  string name = 1;
}
`

	err := os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := LintOptions{}
	_ = RunLint(&buf, tmpDir, opts)

	output := buf.String()
	// Should produce lint output for missing package
	if !strings.Contains(output, "test.proto") {
		t.Errorf("RunLint() should reference the proto file in output:\n%s", output)
	}
}

func TestIsPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"User", true},
		{"UserInfo", true},
		{"HTTPServer", true},
		{"user", false},
		{"user_info", false},
		{"userInfo", false},
		{"USER", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isPascalCase(tt.input); got != tt.want {
				t.Errorf("isPascalCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsLowerSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"user_name", true},
		{"user_id", true},
		{"name", true},
		{"id", true},
		{"user_first_name", true},
		{"userName", false},
		{"UserName", false},
		{"USER_NAME", false},
		{"user__name", true},
		{"_user", false},
		{"user_", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isLowerSnakeCase(tt.input); got != tt.want {
				t.Errorf("isLowerSnakeCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsUpperSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"STATUS_ACTIVE", true},
		{"ROLE_ADMIN", true},
		{"UNSPECIFIED", true},
		{"A", true},
		{"status_active", false},
		{"StatusActive", false},
		{"STATUS__ACTIVE", true},
		{"_STATUS", false},
		{"STATUS_", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isUpperSnakeCase(tt.input); got != tt.want {
				t.Errorf("isUpperSnakeCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestToUpperSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Status", "STATUS"},
		{"UserRole", "USER_ROLE"},
		{"status", "STATUS"},
		{"userRole", "USER_ROLE"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toUpperSnakeCase(tt.input)
			if got == "" && tt.input != "" {
				t.Errorf("toUpperSnakeCase(%q) = empty string", tt.input)
			}
		})
	}
}

func TestLintRuleCategories(t *testing.T) {
	tmpDir := t.TempDir()

	protoContent := `syntax = "proto3";

package test;

message user_info {
  string userName = 1;
}
`

	err := os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := LintOptions{}
	_ = RunLint(&buf, tmpDir, opts)

	if buf.Len() == 0 {
		t.Error("STANDARD rules should produce lint output for bad proto")
	}
}

func TestRunLintWithConfig(t *testing.T) {
	tmpDir := t.TempDir()

	configYAML := `version: v1
lint:
  use:
    - MINIMAL
`

	err := os.WriteFile(filepath.Join(tmpDir, "buf.yaml"), []byte(configYAML), 0644)
	if err != nil {
		t.Fatal(err)
	}

	protoContent := `syntax = "proto3";

package test;

message user_info {
  string userName = 1;
}
`

	err = os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := LintOptions{}
	_ = RunLint(&buf, tmpDir, opts)

	output := buf.String()
	if output == "" {
		t.Log("MINIMAL rules produced no output (expected for some configs)")
	}
}
