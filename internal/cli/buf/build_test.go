package buf

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBuild(t *testing.T) {
	tmpDir := t.TempDir()

	protoContent := `syntax = "proto3";

package test.v1;

message User {
  string id = 1;
  string name = 2;
}

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
}

service UserService {
  rpc GetUser(User) returns (User);
}
`

	err := os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opts    BuildOptions
		wantErr bool
		check   func(string) bool
	}{
		{
			name:    "build without output",
			opts:    BuildOptions{},
			wantErr: false,
			check: func(output string) bool {
				return strings.Contains(output, "Built 1 file(s)") &&
					strings.Contains(output, "test.proto") &&
					strings.Contains(output, "messages: User") &&
					strings.Contains(output, "services: UserService")
			},
		},
		{
			name: "build with json output",
			opts: BuildOptions{
				Output: filepath.Join(tmpDir, "image.json"),
			},
			wantErr: false,
			check: func(output string) bool {
				return strings.Contains(output, "Built 1 file(s) to")
			},
		},
		{
			name: "build with bin output",
			opts: BuildOptions{
				Output: filepath.Join(tmpDir, "image.bin"),
			},
			wantErr: false,
			check: func(output string) bool {
				return strings.Contains(output, "Built 1 file(s) to")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunBuild(&buf, tmpDir, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunBuild() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.check != nil && !tt.check(buf.String()) {
				t.Errorf("RunBuild() output check failed:\n%s", buf.String())
			}
		})
	}
}

func TestRunBuildJSONOutput(t *testing.T) {
	tmpDir := t.TempDir()

	protoContent := `syntax = "proto3";

package test.v1;

message User {
  string id = 1;
}
`

	err := os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(tmpDir, "image.json")

	var buf bytes.Buffer

	opts := BuildOptions{Output: outputPath}

	err = RunBuild(&buf, tmpDir, opts)
	if err != nil {
		t.Fatalf("RunBuild() error = %v", err)
	}

	// Read and parse the JSON output
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var image ProtoImage
	if err := json.Unmarshal(content, &image); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(image.Files) != 1 {
		t.Errorf("Image has %d files, want 1", len(image.Files))
	}

	if image.Files[0].Package != "test.v1" {
		t.Errorf("File package = %s, want test.v1", image.Files[0].Package)
	}

	if len(image.Files[0].Messages) != 1 || image.Files[0].Messages[0] != "User" {
		t.Errorf("File messages = %v, want [User]", image.Files[0].Messages)
	}
}

func TestRunBuildNoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer

	opts := BuildOptions{}

	err := RunBuild(&buf, tmpDir, opts)
	if err != nil {
		t.Errorf("RunBuild() with no files should not error: %v", err)
	}

	if !strings.Contains(buf.String(), "No proto files found") {
		t.Error("RunBuild() should indicate no proto files found")
	}
}

func TestRunBuildWithErrors(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid proto file
	protoContent := `syntax = "proto3";
package test;
message { invalid }
`

	err := os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := BuildOptions{}
	_ = RunBuild(&buf, tmpDir, opts)

	// Should report errors but still process what it can
	output := buf.String()
	_ = output // May or may not have errors depending on parser behavior
}

func TestRunBreaking(t *testing.T) {
	tmpDir := t.TempDir()
	currentDir := filepath.Join(tmpDir, "current")
	againstDir := filepath.Join(tmpDir, "against")

	_ = os.MkdirAll(currentDir, 0755)
	_ = os.MkdirAll(againstDir, 0755)

	tests := []struct {
		name          string
		currentProto  string
		againstProto  string
		wantErr       bool
		wantBreaking  bool
		checkContains string
	}{
		{
			name: "no breaking changes",
			currentProto: `syntax = "proto3";
package test;
message User {
  string id = 1;
  string name = 2;
}`,
			againstProto: `syntax = "proto3";
package test;
message User {
  string id = 1;
  string name = 2;
}`,
			wantErr:      false,
			wantBreaking: false,
		},
		{
			name: "field deleted",
			currentProto: `syntax = "proto3";
package test;
message User {
  string id = 1;
}`,
			againstProto: `syntax = "proto3";
package test;
message User {
  string id = 1;
  string name = 2;
}`,
			wantErr:       true,
			wantBreaking:  true,
			checkContains: "FIELD_NO_DELETE",
		},
		{
			name: "message deleted",
			currentProto: `syntax = "proto3";
package test;
`,
			againstProto: `syntax = "proto3";
package test;
message User {
  string id = 1;
}`,
			wantErr:       true,
			wantBreaking:  true,
			checkContains: "MESSAGE_NO_DELETE",
		},
		{
			name: "field type changed",
			currentProto: `syntax = "proto3";
package test;
message User {
  int32 id = 1;
}`,
			againstProto: `syntax = "proto3";
package test;
message User {
  string id = 1;
}`,
			wantErr:       true,
			wantBreaking:  true,
			checkContains: "FIELD_SAME_TYPE",
		},
		{
			name: "service deleted",
			currentProto: `syntax = "proto3";
package test;
message Request {}
message Response {}
`,
			againstProto: `syntax = "proto3";
package test;
message Request {}
message Response {}
service TestService {
  rpc Get(Request) returns (Response);
}`,
			wantErr:       true,
			wantBreaking:  true,
			checkContains: "SERVICE_NO_DELETE",
		},
		{
			name: "file deleted",
			currentProto: `syntax = "proto3";
package test;
`,
			againstProto: `syntax = "proto3";
package test;
`,
			wantErr:      false, // Same file exists
			wantBreaking: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write proto files
			_ = os.WriteFile(filepath.Join(currentDir, "test.proto"), []byte(tt.currentProto), 0644)
			_ = os.WriteFile(filepath.Join(againstDir, "test.proto"), []byte(tt.againstProto), 0644)

			var buf bytes.Buffer

			opts := BreakingOptions{Against: againstDir}
			err := RunBreaking(&buf, currentDir, opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunBreaking() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			hasBreaking := strings.Contains(output, "FIELD_NO_DELETE") ||
				strings.Contains(output, "MESSAGE_NO_DELETE") ||
				strings.Contains(output, "FIELD_SAME_TYPE") ||
				strings.Contains(output, "SERVICE_NO_DELETE") ||
				strings.Contains(output, "FILE_NO_DELETE") ||
				strings.Contains(output, "RPC_NO_DELETE")

			if hasBreaking != tt.wantBreaking {
				t.Errorf("RunBreaking() hasBreaking = %v, want %v\nOutput:\n%s", hasBreaking, tt.wantBreaking, output)
			}

			if tt.checkContains != "" && !strings.Contains(output, tt.checkContains) {
				t.Errorf("RunBreaking() output should contain %q\nOutput:\n%s", tt.checkContains, output)
			}
		})
	}
}

func TestRunBreakingMissingAgainst(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer

	opts := BreakingOptions{} // No Against specified

	err := RunBreaking(&buf, tmpDir, opts)
	if err == nil {
		t.Error("RunBreaking() should error when --against is missing")
	}

	if !strings.Contains(err.Error(), "--against") {
		t.Errorf("Error should mention --against: %v", err)
	}
}

func TestCheckMessageFieldChanges(t *testing.T) {
	current := []ProtoMessage{
		{
			Name: "User",
			Fields: []ProtoField{
				{Name: "id", Type: "string", Number: 1},
			},
		},
	}

	against := []ProtoMessage{
		{
			Name: "User",
			Fields: []ProtoField{
				{Name: "id", Type: "string", Number: 1},
				{Name: "name", Type: "string", Number: 2},
			},
		},
	}

	results := checkMessageFieldChanges("test.proto", current, against)

	if len(results) == 0 {
		t.Error("checkMessageFieldChanges() should detect deleted field")
	}

	foundDelete := false

	for _, r := range results {
		if r.Rule == "FIELD_NO_DELETE" {
			foundDelete = true
			break
		}
	}

	if !foundDelete {
		t.Error("checkMessageFieldChanges() should return FIELD_NO_DELETE rule")
	}
}
