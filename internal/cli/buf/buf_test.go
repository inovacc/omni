package buf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		configYAML string
		wantErr    bool
		checkFunc  func(*Config) bool
	}{
		{
			name: "valid config with lint rules",
			configYAML: `version: v1
name: buf.build/test/repo
lint:
  use:
    - MINIMAL
    - BASIC
  except:
    - PACKAGE_VERSION_SUFFIX
`,
			wantErr: false,
			checkFunc: func(c *Config) bool {
				return c.Version == "v1" &&
					c.Name == "buf.build/test/repo" &&
					len(c.Lint.Use) == 2 &&
					c.Lint.Use[0] == "MINIMAL"
			},
		},
		{
			name: "valid config with breaking rules",
			configYAML: `version: v1
breaking:
  use:
    - FILE
  except:
    - FIELD_NO_DELETE
`,
			wantErr: false,
			checkFunc: func(c *Config) bool {
				return len(c.Breaking.Use) == 1 &&
					c.Breaking.Use[0] == "FILE"
			},
		},
		{
			name: "valid config with deps",
			configYAML: `version: v1
deps:
  - buf.build/googleapis/googleapis
  - buf.build/grpc/grpc
`,
			wantErr: false,
			checkFunc: func(c *Config) bool {
				return len(c.Deps) == 2
			},
		},
		{
			name:       "invalid yaml",
			configYAML: `version: v1\n  invalid: [`,
			wantErr:    true,
			checkFunc:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config file
			configPath := filepath.Join(tmpDir, tt.name)
			_ = os.MkdirAll(configPath, 0755)

			if tt.configYAML != "" {
				err := os.WriteFile(filepath.Join(configPath, "buf.yaml"), []byte(tt.configYAML), 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			config, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil && !tt.checkFunc(config) {
				t.Errorf("LoadConfig() returned unexpected config: %+v", config)
			}
		})
	}
}

func TestLoadConfigDefault(t *testing.T) {
	// Test loading from directory without buf.yaml
	tmpDir := t.TempDir()

	config, err := LoadConfig(tmpDir)
	if err != nil {
		t.Errorf("LoadConfig() should return default config, got error: %v", err)
	}

	if config.Version != "v1" {
		t.Errorf("Default config version = %s, want v1", config.Version)
	}

	if len(config.Lint.Use) == 0 || config.Lint.Use[0] != "STANDARD" {
		t.Errorf("Default config should use STANDARD lint rules")
	}
}

func TestLoadGenerateConfig(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		configYAML string
		wantErr    bool
		checkFunc  func(*GenerateConfig) bool
	}{
		{
			name: "valid generate config",
			configYAML: `version: v1
plugins:
  - local: protoc-gen-go
    out: gen/go
    opt:
      - paths=source_relative
  - remote: buf.build/protocolbuffers/go
    out: gen/go
`,
			wantErr: false,
			checkFunc: func(c *GenerateConfig) bool {
				return c.Version == "v1" &&
					len(c.Plugins) == 2 &&
					c.Plugins[0].Local == "protoc-gen-go" &&
					c.Plugins[1].Remote == "buf.build/protocolbuffers/go"
			},
		},
		{
			name: "config with managed mode",
			configYAML: `version: v1
managed:
  enabled: true
  go_package_prefix:
    default: github.com/test/repo/gen
plugins:
  - local: protoc-gen-go
    out: gen/go
`,
			wantErr: false,
			checkFunc: func(c *GenerateConfig) bool {
				return c.Managed.Enabled &&
					c.Managed.GoPackagePrefix.Default == "github.com/test/repo/gen"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, tt.name)
			_ = os.MkdirAll(configPath, 0755)

			err := os.WriteFile(filepath.Join(configPath, "buf.gen.yaml"), []byte(tt.configYAML), 0644)
			if err != nil {
				t.Fatal(err)
			}

			config, err := LoadGenerateConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadGenerateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil && !tt.checkFunc(config) {
				t.Errorf("LoadGenerateConfig() returned unexpected config: %+v", config)
			}
		})
	}
}

func TestLoadGenerateConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadGenerateConfig(tmpDir)
	if err == nil {
		t.Error("LoadGenerateConfig() should error when buf.gen.yaml not found")
	}
}

func TestFindProtoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure with proto files
	dirs := []string{
		"",
		"api",
		"api/v1",
		"vendor",
		"vendor/google",
	}

	for _, d := range dirs {
		_ = os.MkdirAll(filepath.Join(tmpDir, d), 0755)
	}

	// Create proto files
	protoFiles := []string{
		"test.proto",
		"api/service.proto",
		"api/v1/types.proto",
		"vendor/google/api.proto",
	}

	for _, f := range protoFiles {
		err := os.WriteFile(filepath.Join(tmpDir, f), []byte("syntax = \"proto3\";"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create non-proto file
	_ = os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Test"), 0644)

	tests := []struct {
		name        string
		excludePath []string
		wantCount   int
	}{
		{
			name:        "find all proto files",
			excludePath: nil,
			wantCount:   4,
		},
		{
			name:        "exclude vendor",
			excludePath: []string{"vendor"},
			wantCount:   3,
		},
		{
			name:        "exclude multiple paths",
			excludePath: []string{"vendor", "api"},
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := FindProtoFiles(tmpDir, tt.excludePath)
			if err != nil {
				t.Errorf("FindProtoFiles() error = %v", err)
				return
			}

			if len(files) != tt.wantCount {
				t.Errorf("FindProtoFiles() found %d files, want %d", len(files), tt.wantCount)
			}
		})
	}
}

func TestOutputResults(t *testing.T) {
	results := []LintResult{
		{File: "test.proto", Line: 1, Column: 1, Rule: "PACKAGE_DEFINED", Message: "Package not defined"},
		{File: "test.proto", Line: 5, Column: 3, Rule: "FIELD_LOWER_SNAKE_CASE", Message: "Field name should be lower_snake_case"},
	}

	tests := []struct {
		name     string
		format   string
		contains []string
	}{
		{
			name:   "text format",
			format: "text",
			contains: []string{
				"test.proto:1:1: PACKAGE_DEFINED: Package not defined",
				"test.proto:5:3: FIELD_LOWER_SNAKE_CASE: Field name should be lower_snake_case",
			},
		},
		{
			name:   "json format",
			format: "json",
			contains: []string{
				`"file": "test.proto"`,
				`"rule": "PACKAGE_DEFINED"`,
			},
		},
		{
			name:   "github-actions format",
			format: "github-actions",
			contains: []string{
				"::error file=test.proto,line=1,col=1::PACKAGE_DEFINED: Package not defined",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf testWriter

			err := OutputResults(&buf, results, tt.format)
			if err != nil {
				t.Errorf("OutputResults() error = %v", err)
				return
			}

			output := buf.String()
			for _, want := range tt.contains {
				if !containsString(output, want) {
					t.Errorf("OutputResults() output missing %q, got:\n%s", want, output)
				}
			}
		})
	}
}

type testWriter struct {
	data []byte
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *testWriter) String() string {
	return string(w.data)
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
