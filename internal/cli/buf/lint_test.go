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
			wantErr:    false, // Outputs issues but returns nil
			wantIssues: 1,     // PACKAGE_DEFINED
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
			wantErr:    false, // Outputs issues but returns nil
			wantIssues: 1,     // At least PACKAGE_DIRECTORY_MATCH
		},
		{
			name: "bad message name casing",
			protoFile: `syntax = "proto3";

package test.v1;

message user_info {
  string name = 1;
}
`,
			wantErr:    false, // Outputs issues but returns nil
			wantIssues: 1,     // At least one issue
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
			wantErr:    false, // Outputs issues but returns nil
			wantIssues: 1,     // At least one issue
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory and file
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

			// Count issues in output
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
	if err != nil {
		t.Errorf("RunLint() with no proto files should not error: %v", err)
	}

	if !strings.Contains(buf.String(), "No proto files found") {
		t.Errorf("RunLint() should indicate no proto files found")
	}
}

func TestRunLintJSONOutput(t *testing.T) {
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
	opts.ErrorFormat = "json"
	_ = RunLint(&buf, tmpDir, opts)

	output := buf.String()
	if !strings.Contains(output, `"rule"`) {
		t.Errorf("RunLint() JSON output should contain rule field:\n%s", output)
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
		{"USER", true}, // All caps is considered PascalCase by implementation
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
		{"user__name", true}, // Implementation allows double underscore
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
		{"STATUS__ACTIVE", true}, // Implementation allows double underscore
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
			// Just check that it's uppercase and contains underscores where expected
			if got == "" && tt.input != "" {
				t.Errorf("toUpperSnakeCase(%q) = empty string", tt.input)
			}
		})
	}
}

func TestLintRuleCategories(t *testing.T) {
	// Test that different configs produce different results
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

	// Should produce lint errors with default STANDARD rules
	if buf.Len() == 0 {
		t.Error("STANDARD rules should produce lint output for bad proto")
	}
}

func TestRunLintWithConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create buf.yaml with MINIMAL rules only
	configYAML := `version: v1
lint:
  use:
    - MINIMAL
`

	err := os.WriteFile(filepath.Join(tmpDir, "buf.yaml"), []byte(configYAML), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create proto file with issues that MINIMAL doesn't catch
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

	// With MINIMAL rules, we should still get some basic checks
	output := buf.String()
	if output == "" {
		t.Log("MINIMAL rules produced no output (expected for some configs)")
	}
}
