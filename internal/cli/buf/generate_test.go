package buf

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInit(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
		wantErr    bool
		check      func(string) bool
	}{
		{
			name:       "init without module name",
			moduleName: "",
			wantErr:    false,
			check: func(content string) bool {
				return strings.Contains(content, "version: v1") &&
					strings.Contains(content, "STANDARD") &&
					!strings.Contains(content, "name:")
			},
		},
		{
			name:       "init with module name",
			moduleName: "buf.build/test/repo",
			wantErr:    false,
			check: func(content string) bool {
				return strings.Contains(content, "version: v1") &&
					strings.Contains(content, "name: buf.build/test/repo")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			var buf bytes.Buffer

			err := RunInit(&buf, tmpDir, tt.moduleName)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunInit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check buf.yaml was created
			configPath := filepath.Join(tmpDir, "buf.yaml")

			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read buf.yaml: %v", err)
			}

			if tt.check != nil && !tt.check(string(content)) {
				t.Errorf("RunInit() created unexpected config:\n%s", string(content))
			}

			// Check output message
			if !strings.Contains(buf.String(), "Created buf.yaml") {
				t.Error("RunInit() should output success message")
			}
		})
	}
}

func TestRunInitExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing buf.yaml
	err := os.WriteFile(filepath.Join(tmpDir, "buf.yaml"), []byte("version: v1\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunInit(&buf, tmpDir, "")
	if err == nil {
		t.Error("RunInit() should error when buf.yaml already exists")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Error should mention buf.yaml already exists: %v", err)
	}
}

func TestRunLsFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	_ = os.MkdirAll(filepath.Join(tmpDir, "api", "v1"), 0755)

	// Create proto files
	files := []string{
		"test.proto",
		"api/service.proto",
		"api/v1/types.proto",
	}

	for _, f := range files {
		err := os.WriteFile(filepath.Join(tmpDir, f), []byte("syntax = \"proto3\";"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer

	err := RunLsFiles(&buf, tmpDir)
	if err != nil {
		t.Errorf("RunLsFiles() error = %v", err)
	}

	output := buf.String()

	for _, f := range files {
		// Convert to forward slashes for comparison
		expected := strings.ReplaceAll(f, "/", string(filepath.Separator))
		if !strings.Contains(output, "test.proto") {
			t.Errorf("RunLsFiles() output should contain %s:\n%s", expected, output)
		}
	}
}

func TestRunLsFilesEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer

	err := RunLsFiles(&buf, tmpDir)
	if err != nil {
		t.Errorf("RunLsFiles() error = %v", err)
	}

	// Empty directory should produce no output
	if strings.TrimSpace(buf.String()) != "" {
		t.Errorf("RunLsFiles() should produce no output for empty dir: %q", buf.String())
	}
}

func TestRunDepUpdate(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		config  string
		wantOut string
	}{
		{
			name:    "no deps",
			config:  "version: v1\n",
			wantOut: "No dependencies to update",
		},
		{
			name: "with deps",
			config: `version: v1
deps:
  - buf.build/googleapis/googleapis
  - buf.build/grpc/grpc
`,
			wantOut: "buf.build/googleapis/googleapis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			_ = os.MkdirAll(testDir, 0755)

			err := os.WriteFile(filepath.Join(testDir, "buf.yaml"), []byte(tt.config), 0644)
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer

			err = RunDepUpdate(&buf, testDir)
			if err != nil {
				t.Errorf("RunDepUpdate() error = %v", err)
			}

			if !strings.Contains(buf.String(), tt.wantOut) {
				t.Errorf("RunDepUpdate() output should contain %q:\n%s", tt.wantOut, buf.String())
			}
		})
	}
}

func TestRunGenerate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create proto file
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

	tests := []struct {
		name       string
		genConfig  string
		wantErr    bool
		wantOutput string
	}{
		{
			name: "generate with remote plugin",
			genConfig: `version: v1
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen/go
`,
			wantErr:    false,
			wantOutput: "Remote plugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			_ = os.MkdirAll(testDir, 0755)

			// Copy proto file
			_ = os.WriteFile(filepath.Join(testDir, "test.proto"), []byte(protoContent), 0644)

			// Write gen config
			err := os.WriteFile(filepath.Join(testDir, "buf.gen.yaml"), []byte(tt.genConfig), 0644)
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer

			opts := GenerateOptions{}
			err = RunGenerate(&buf, testDir, opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunGenerate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantOutput != "" && !strings.Contains(buf.String(), tt.wantOutput) {
				t.Errorf("RunGenerate() output should contain %q:\n%s", tt.wantOutput, buf.String())
			}
		})
	}
}

func TestRunGenerateNoConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create proto file but no config
	protoContent := `syntax = "proto3";
package test;
message User { string id = 1; }
`

	err := os.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(protoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := GenerateOptions{}

	err = RunGenerate(&buf, tmpDir, opts)
	if err == nil {
		t.Error("RunGenerate() should error when buf.gen.yaml is missing")
	}
}

func TestRunGenerateNoProtoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config but no proto files
	genConfig := `version: v1
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen/go
`

	err := os.WriteFile(filepath.Join(tmpDir, "buf.gen.yaml"), []byte(genConfig), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := GenerateOptions{}

	err = RunGenerate(&buf, tmpDir, opts)
	if err != nil {
		t.Errorf("RunGenerate() with no proto files should not error: %v", err)
	}

	if !strings.Contains(buf.String(), "No proto files found") {
		t.Error("RunGenerate() should indicate no proto files found")
	}
}

func TestExtractPluginName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"protoc-gen-go", "go"},
		{"protoc-gen-go-grpc", "go-grpc"},
		{"/usr/local/bin/protoc-gen-go", "go"},
		{"C:\\bin\\protoc-gen-go.exe", "go.exe"},
		{"custom-plugin", "custom-plugin"},
		{"go", "go"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := extractPluginName(tt.input); got != tt.want {
				t.Errorf("extractPluginName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
