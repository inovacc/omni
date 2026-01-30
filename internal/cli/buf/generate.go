package buf

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunGenerate generates code from proto files
func RunGenerate(w io.Writer, dir string, opts GenerateOptions) error {
	// Load generation config
	configPath := opts.Template
	if configPath == "" {
		configPath = dir
	}

	config, err := LoadGenerateConfig(configPath)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	// Find proto files
	files, err := FindProtoFiles(dir, nil)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	if len(files) == 0 {
		_, _ = fmt.Fprintln(w, "No proto files found")

		return nil
	}

	// Clean output directories if configured
	if config.Clean {
		for _, plugin := range config.Plugins {
			outDir := plugin.Out
			if opts.Output != "" {
				outDir = filepath.Join(opts.Output, plugin.Out)
			}

			if err := os.RemoveAll(outDir); err != nil {
				_, _ = fmt.Fprintf(w, "Warning: failed to clean %s: %v\n", outDir, err)
			}
		}
	}

	// Create output directories
	for _, plugin := range config.Plugins {
		outDir := plugin.Out
		if opts.Output != "" {
			outDir = filepath.Join(opts.Output, plugin.Out)
		}

		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("buf: failed to create output directory %s: %w", outDir, err)
		}
	}

	// Run each plugin
	for _, plugin := range config.Plugins {
		outDir := plugin.Out
		if opts.Output != "" {
			outDir = filepath.Join(opts.Output, plugin.Out)
		}

		if plugin.Local != "" {
			// Run local plugin
			if err := runLocalPlugin(w, dir, files, plugin, outDir); err != nil {
				return fmt.Errorf("buf: plugin %s failed: %w", plugin.Local, err)
			}

			_, _ = fmt.Fprintf(w, "Generated with %s to %s\n", plugin.Local, outDir)
		} else if plugin.Remote != "" {
			// Remote plugins require protoc or buf to actually run
			_, _ = fmt.Fprintf(w, "Remote plugin %s would generate to %s\n", plugin.Remote, outDir)
			_, _ = fmt.Fprintln(w, "Note: Remote plugins require protoc or the full buf tool")
		}
	}

	_, _ = fmt.Fprintf(w, "\nProcessed %d proto file(s)\n", len(files))

	return nil
}

func runLocalPlugin(w io.Writer, dir string, files []string, plugin PluginConfig, outDir string) error {
	// Check if protoc is available
	protocPath, err := exec.LookPath("protoc")
	if err != nil {
		return fmt.Errorf("protoc not found in PATH")
	}

	// Build protoc command
	args := []string{
		"--proto_path=" + dir,
	}

	// Add plugin output
	pluginArg := fmt.Sprintf("--%s_out=%s", extractPluginName(plugin.Local), outDir)

	if len(plugin.Opt) > 0 {
		pluginArg = fmt.Sprintf("--%s_out=%s:%s", extractPluginName(plugin.Local), strings.Join(plugin.Opt, ","), outDir)
	}

	args = append(args, pluginArg)

	// Add proto files
	for _, file := range files {
		relPath, _ := filepath.Rel(dir, file)
		args = append(args, relPath)
	}

	cmd := exec.Command(protocPath, args...)
	cmd.Dir = dir
	cmd.Stdout = w
	cmd.Stderr = w

	return cmd.Run()
}

func extractPluginName(pluginPath string) string {
	// Extract plugin name from path like "protoc-gen-go" -> "go"
	base := filepath.Base(pluginPath)
	if after, ok := strings.CutPrefix(base, "protoc-gen-"); ok {
		return after
	}

	return base
}

// RunInit initializes a new buf module
func RunInit(w io.Writer, dir string, moduleName string) error {
	configPath := filepath.Join(dir, "buf.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("buf.yaml already exists")
	}

	// Create default config
	config := `version: v1
`

	if moduleName != "" {
		config += fmt.Sprintf("name: %s\n", moduleName)
	}

	config += `lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
`

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("buf: failed to write buf.yaml: %w", err)
	}

	_, _ = fmt.Fprintf(w, "Created buf.yaml in %s\n", dir)

	return nil
}

// RunLsFiles lists proto files in the module
func RunLsFiles(w io.Writer, dir string) error {
	files, err := FindProtoFiles(dir, nil)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	for _, file := range files {
		relPath, _ := filepath.Rel(dir, file)

		_, _ = fmt.Fprintln(w, relPath)
	}

	return nil
}

// RunDepUpdate updates dependencies (placeholder)
func RunDepUpdate(w io.Writer, dir string) error {
	config, err := LoadConfig(dir)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	if len(config.Deps) == 0 {
		_, _ = fmt.Fprintln(w, "No dependencies to update")

		return nil
	}

	_, _ = fmt.Fprintln(w, "Dependencies:")
	for _, dep := range config.Deps {
		_, _ = fmt.Fprintf(w, "  %s\n", dep)
	}

	_, _ = fmt.Fprintln(w, "\nNote: Dependency resolution requires network access to BSR")

	return nil
}
