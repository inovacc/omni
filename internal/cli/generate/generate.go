package generate

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Options configures the generate command behavior
type Options struct {
	JSON bool // --json: output as JSON
}

// CobraInitOptions configures Cobra app initialization
type CobraInitOptions struct {
	Module      string // Go module path (e.g., github.com/user/myapp)
	AppName     string // Application name
	Description string // Application description
	Author      string // Author name
	License     string // License type (MIT, Apache-2.0, BSD-3)
	UseViper    bool   // Include viper for configuration
}

// CobraAddOptions configures adding a new command
type CobraAddOptions struct {
	Name        string // Command name
	Parent      string // Parent command (default: root)
	Description string // Command description
}

// InitResult represents the result of initialization
type InitResult struct {
	Status      string   `json:"status"`
	Path        string   `json:"path"`
	Module      string   `json:"module"`
	FilesCreated []string `json:"files_created"`
}

// AddResult represents the result of adding a command
type AddResult struct {
	Status  string `json:"status"`
	Command string `json:"command"`
	Parent  string `json:"parent"`
	File    string `json:"file"`
}

// RunCobraInit initializes a new Cobra CLI application
func RunCobraInit(w io.Writer, dir string, opts CobraInitOptions, genOpts Options) error {
	if opts.Module == "" {
		return fmt.Errorf("generate: module path is required")
	}

	if opts.AppName == "" {
		// Extract app name from module path
		parts := strings.Split(opts.Module, "/")
		opts.AppName = parts[len(parts)-1]
	}

	if opts.Description == "" {
		opts.Description = fmt.Sprintf("%s is a CLI application", opts.AppName)
	}

	// Create directory structure
	dirs := []string{
		filepath.Join(dir, "cmd"),
		filepath.Join(dir, "internal"),
	}

	if opts.UseViper {
		dirs = append(dirs, filepath.Join(dir, "internal", "config"))
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("generate: failed to create directory %s: %w", d, err)
		}
	}

	var filesCreated []string

	// Generate main.go
	mainPath := filepath.Join(dir, "main.go")
	if err := writeTemplate(mainPath, mainTemplate, opts); err != nil {
		return fmt.Errorf("generate: failed to create main.go: %w", err)
	}
	filesCreated = append(filesCreated, "main.go")

	// Generate cmd/root.go
	rootPath := filepath.Join(dir, "cmd", "root.go")
	if err := writeTemplate(rootPath, rootTemplate, opts); err != nil {
		return fmt.Errorf("generate: failed to create cmd/root.go: %w", err)
	}
	filesCreated = append(filesCreated, "cmd/root.go")

	// Generate cmd/version.go
	versionPath := filepath.Join(dir, "cmd", "version.go")
	if err := writeTemplate(versionPath, versionTemplate, opts); err != nil {
		return fmt.Errorf("generate: failed to create cmd/version.go: %w", err)
	}
	filesCreated = append(filesCreated, "cmd/version.go")

	// Generate go.mod
	goModPath := filepath.Join(dir, "go.mod")
	if err := writeTemplate(goModPath, goModTemplate, opts); err != nil {
		return fmt.Errorf("generate: failed to create go.mod: %w", err)
	}
	filesCreated = append(filesCreated, "go.mod")

	// Generate config if viper is enabled
	if opts.UseViper {
		configPath := filepath.Join(dir, "internal", "config", "config.go")
		if err := writeTemplate(configPath, configTemplate, opts); err != nil {
			return fmt.Errorf("generate: failed to create config.go: %w", err)
		}
		filesCreated = append(filesCreated, "internal/config/config.go")
	}

	// Generate LICENSE
	if opts.License != "" {
		licensePath := filepath.Join(dir, "LICENSE")
		if err := writeLicense(licensePath, opts.License, opts.Author); err != nil {
			return fmt.Errorf("generate: failed to create LICENSE: %w", err)
		}
		filesCreated = append(filesCreated, "LICENSE")
	}

	// Generate README.md
	readmePath := filepath.Join(dir, "README.md")
	if err := writeTemplate(readmePath, readmeTemplate, opts); err != nil {
		return fmt.Errorf("generate: failed to create README.md: %w", err)
	}
	filesCreated = append(filesCreated, "README.md")

	// Generate Taskfile.yml
	taskfilePath := filepath.Join(dir, "Taskfile.yml")
	if err := writeTemplate(taskfilePath, taskfileTemplate, opts); err != nil {
		return fmt.Errorf("generate: failed to create Taskfile.yml: %w", err)
	}
	filesCreated = append(filesCreated, "Taskfile.yml")

	// Generate .gitignore
	gitignorePath := filepath.Join(dir, ".gitignore")
	if err := writeTemplate(gitignorePath, gitignoreTemplate, opts); err != nil {
		return fmt.Errorf("generate: failed to create .gitignore: %w", err)
	}
	filesCreated = append(filesCreated, ".gitignore")

	if genOpts.JSON {
		result := InitResult{
			Status:       "created",
			Path:         dir,
			Module:       opts.Module,
			FilesCreated: filesCreated,
		}
		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created Cobra CLI application: %s\n", opts.AppName)
	_, _ = fmt.Fprintf(w, "Module: %s\n", opts.Module)
	_, _ = fmt.Fprintf(w, "Path: %s\n", dir)
	_, _ = fmt.Fprintln(w, "\nFiles created:")
	for _, f := range filesCreated {
		_, _ = fmt.Fprintf(w, "  - %s\n", f)
	}
	_, _ = fmt.Fprintln(w, "\nNext steps:")
	_, _ = fmt.Fprintf(w, "  cd %s\n", dir)
	_, _ = fmt.Fprintln(w, "  go mod tidy")
	_, _ = fmt.Fprintln(w, "  go build")
	_, _ = fmt.Fprintf(w, "  ./%s --help\n", opts.AppName)

	return nil
}

// RunCobraAdd adds a new command to an existing Cobra application
func RunCobraAdd(w io.Writer, dir string, opts CobraAddOptions, genOpts Options) error {
	if opts.Name == "" {
		return fmt.Errorf("generate: command name is required")
	}

	if opts.Parent == "" {
		opts.Parent = "root"
	}

	if opts.Description == "" {
		opts.Description = fmt.Sprintf("%s command", opts.Name)
	}

	// Check if cmd directory exists
	cmdDir := filepath.Join(dir, "cmd")
	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		return fmt.Errorf("generate: cmd directory not found, is this a Cobra project?")
	}

	// Generate the command file
	cmdPath := filepath.Join(cmdDir, opts.Name+".go")
	if _, err := os.Stat(cmdPath); err == nil {
		return fmt.Errorf("generate: command %s already exists", opts.Name)
	}

	data := struct {
		Name        string
		Parent      string
		Description string
		NameTitle   string
	}{
		Name:        opts.Name,
		Parent:      opts.Parent,
		Description: opts.Description,
		NameTitle:   strings.Title(opts.Name),
	}

	if err := writeTemplate(cmdPath, commandTemplate, data); err != nil {
		return fmt.Errorf("generate: failed to create %s.go: %w", opts.Name, err)
	}

	if genOpts.JSON {
		result := AddResult{
			Status:  "created",
			Command: opts.Name,
			Parent:  opts.Parent,
			File:    filepath.Join("cmd", opts.Name+".go"),
		}
		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created command: %s\n", opts.Name)
	_, _ = fmt.Fprintf(w, "Parent: %s\n", opts.Parent)
	_, _ = fmt.Fprintf(w, "File: cmd/%s.go\n", opts.Name)

	return nil
}

func writeTemplate(path string, tmpl string, data any) error {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return t.Execute(f, data)
}

func writeLicense(path, licenseType, author string) error {
	year := time.Now().Year()

	var content string
	switch strings.ToLower(licenseType) {
	case "mit":
		content = fmt.Sprintf(mitLicense, year, author)
	case "apache-2.0", "apache":
		content = fmt.Sprintf(apacheLicense, year, author)
	case "bsd-3", "bsd":
		content = fmt.Sprintf(bsdLicense, year, author)
	default:
		return fmt.Errorf("unknown license type: %s", licenseType)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// Templates

var mainTemplate = `package main

import "{{.Module}}/cmd"

func main() {
	cmd.Execute()
}
`

var rootTemplate = `package cmd

import (
	"fmt"
	"os"
{{if .UseViper}}
	"{{.Module}}/internal/config"

	"github.com/spf13/viper"
{{end}}
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "{{.AppName}}",
	Short: "{{.Description}}",
	Long: ` + "`" + `{{.Description}}

This is a CLI application built with Cobra.` + "`" + `,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
{{if .UseViper}}
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is $HOME/.{{.AppName}}.yaml)")
{{end}}
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
{{if .UseViper}}
func initConfig() {
	config.InitConfig("{{.AppName}}")
}
{{end}}
`

var versionTemplate = `package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("{{.AppName}} version %s\n", Version)
		fmt.Printf("  Commit: %s\n", Commit)
		fmt.Printf("  Built:  %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
`

var goModTemplate = `module {{.Module}}

go 1.21

require github.com/spf13/cobra v1.8.0
{{if .UseViper}}
require github.com/spf13/viper v1.18.0
{{end}}
`

var configTemplate = `package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var CfgFile string

func InitConfig(appName string) {
	if CfgFile != "" {
		viper.SetConfigFile(CfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("." + appName)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

func GetBool(key string) bool {
	return viper.GetBool(key)
}

func ConfigPath(appName string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "."+appName+".yaml")
}
`

var readmeTemplate = `# {{.AppName}}

{{.Description}}

## Installation

` + "```" + `bash
go install {{.Module}}@latest
` + "```" + `

## Usage

` + "```" + `bash
{{.AppName}} --help
` + "```" + `

## Commands

| Command | Description |
|---------|-------------|
| ` + "`" + `version` + "`" + ` | Print version information |

## Development

` + "```" + `bash
# Build
go build -o {{.AppName}}

# Run
./{{.AppName}} --help

# Test
go test ./...
` + "```" + `

## License

{{if .License}}{{.License}}{{else}}See LICENSE file{{end}}
`

var taskfileTemplate = `version: '3'

vars:
  APP_NAME: {{.AppName}}
  VERSION:
    sh: git describe --tags --always --dirty 2>/dev/null || echo "dev"
  COMMIT:
    sh: git rev-parse --short HEAD 2>/dev/null || echo "none"
  BUILD_DATE:
    sh: date -u +"%Y-%m-%dT%H:%M:%SZ"

tasks:
  default:
    cmds:
      - task --list

  build:
    desc: Build the application
    cmds:
      - go build -ldflags "-X {{.Module}}/cmd.Version=$VERSION -X {{.Module}}/cmd.Commit=$COMMIT -X {{.Module}}/cmd.BuildDate=$BUILD_DATE" -o $APP_NAME
    vars:
      VERSION: "{{"{{"}} .VERSION {{"}}"}}"
      COMMIT: "{{"{{"}} .COMMIT {{"}}"}}"
      BUILD_DATE: "{{"{{"}} .BUILD_DATE {{"}}"}}"
      APP_NAME: "{{"{{"}} .APP_NAME {{"}}"}}"

  run:
    desc: Run the application
    cmds:
      - go run . $CLI_ARGS
    vars:
      CLI_ARGS: "{{"{{"}} .CLI_ARGS {{"}}"}}"

  test:
    desc: Run tests
    cmds:
      - go test -v ./...

  lint:
    desc: Run linter
    cmds:
      - golangci-lint run ./...

  clean:
    desc: Clean build artifacts
    cmds:
      - rm -f $APP_NAME
    vars:
      APP_NAME: "{{"{{"}} .APP_NAME {{"}}"}}"

  tidy:
    desc: Tidy go modules
    cmds:
      - go mod tidy

  install:
    desc: Install the application
    cmds:
      - go install -ldflags "-X {{.Module}}/cmd.Version=$VERSION -X {{.Module}}/cmd.Commit=$COMMIT -X {{.Module}}/cmd.BuildDate=$BUILD_DATE"
    vars:
      VERSION: "{{"{{"}} .VERSION {{"}}"}}"
      COMMIT: "{{"{{"}} .COMMIT {{"}}"}}"
      BUILD_DATE: "{{"{{"}} .BUILD_DATE {{"}}"}}"
`

var gitignoreTemplate = `# Binaries
{{.AppName}}
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of go coverage
*.out

# Go workspace
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Build
dist/
bin/
`

var commandTemplate = `package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var {{.Name}}Cmd = &cobra.Command{
	Use:   "{{.Name}}",
	Short: "{{.Description}}",
	Long: ` + "`" + `{{.Description}}

Add more detailed description here.` + "`" + `,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("{{.Name}} called")
		return nil
	},
}

func init() {
	{{.Parent}}Cmd.AddCommand({{.Name}}Cmd)

	// Add flags here
	// {{.Name}}Cmd.Flags().StringP("example", "e", "", "example flag")
}
`

// License templates
var mitLicense = `MIT License

Copyright (c) %d %s

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

var apacheLicense = `Copyright %d %s

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
`

var bsdLicense = `BSD 3-Clause License

Copyright (c) %d, %s
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its
   contributors may be used to endorse or promote products derived from
   this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
`
