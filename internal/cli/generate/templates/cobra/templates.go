// Package cobra provides unified Cobra CLI application templates.
// These templates are based on production patterns from the include folder
// and support both basic and full project structures.
package cobra

import "time"

// TemplateData contains all data needed for template rendering
type TemplateData struct {
	Module      string // Go module path (e.g., github.com/user/myapp)
	AppName     string // Application name
	Description string // Application description
	Author      string // Author name
	License     string // License type (MIT, Apache-2.0, BSD-3)
	UseViper    bool   // Include viper for configuration
	UseService  bool   // Include service pattern with inovacc/config
	Full        bool   // Full project with goreleaser, workflows, etc.
	Year        int    // Current year for license
}

// NewTemplateData creates template data with defaults
func NewTemplateData(module, appName, description string) TemplateData {
	return TemplateData{
		Module:      module,
		AppName:     appName,
		Description: description,
		Year:        time.Now().Year(),
	}
}

// MainTemplate generates the main.go entry point
const MainTemplate = `package main

import "{{.Module}}/cmd"

func main() {
	cmd.Execute()
}
`

// RootTemplate generates cmd/root.go with an optional service pattern
const RootTemplate = `package cmd

import (
{{if .UseService}}
	"fmt"
	"os"

	"{{.Module}}/internal/parameters"
	"{{.Module}}/internal/service"

	"github.com/inovacc/config"
{{else if .UseViper}}
	"{{.Module}}/internal/config"
{{end}}
	"github.com/spf13/cobra"
)
{{if or .UseViper .UseService}}
var cfgFile string
{{end}}
var rootCmd = &cobra.Command{
	Use:   "{{.AppName}}",
	Short: "{{.Description}}",
	Long: ` + "`" + `{{.Description}}

This is a CLI application built with Cobra.` + "`" + `,
{{if .UseService}}
	RunE: service.Handler,
{{end}}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
{{if or .UseViper .UseService}}
	cobra.OnInitialize(initConfig)
{{end}}
{{if .Full}}
	rootCmd.Version = GetVersionJSON()
	rootCmd.CompletionOptions.DisableDefaultCmd = true
{{end}}
{{if or .UseViper .UseService}}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file (default is config.yaml)")
{{else}}
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
{{end}}
}
{{if or .UseViper .UseService}}
// initConfig reads in config file and ENV variables if set.
func initConfig() {
{{if .UseService}}
	if cfgFile == "" {
		_, _ = fmt.Fprint(os.Stdout, "Using default config file: config.yaml")
	}

	// Load configuration from a file, applying defaults if needed
	if err := config.InitServiceConfig(&parameters.Service{}, cfgFile); err != nil {
		_, _ = fmt.Fprint(os.Stdout, "failed to load config: %w", err)
	}
{{else if .UseViper}}
	config.InitConfig("{{.AppName}}")
{{end}}
}
{{end}}
`

// VersionTemplate generates cmd/version.go with full version info
const VersionTemplate = `package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var jsonOutput bool

// Version information embedded at build time
var (
	// Version is the application version (from git tag or VERSION env)
	Version = "dev"

	// GitHash is the git commit hash
	GitHash = "none"

	// BuildTime is when the binary was built
	BuildTime = "unknown"
{{if .Full}}
	// BuildHash is a unique hash for this build
	BuildHash = "none"

	// GoVersion is the Go version used to build
	GoVersion = "unknown"

	// GOOS is the target operating system
	GOOS = "unknown"

	// GOARCH is the target architecture
	GOARCH = "unknown"
{{end}}
)

// VersionInfo contains all version metadata.
type VersionInfo struct {
	Version   string ` + "`" + `json:"version"` + "`" + `
	GitHash   string ` + "`" + `json:"git_hash"` + "`" + `
	BuildTime string ` + "`" + `json:"build_time"` + "`" + `
{{if .Full}}
	BuildHash string ` + "`" + `json:"build_hash"` + "`" + `
	GoVersion string ` + "`" + `json:"go_version"` + "`" + `
	GoOS      string ` + "`" + `json:"goos"` + "`" + `
	GoArch    string ` + "`" + `json:"goarch"` + "`" + `
{{end}}
}

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long: ` + "`" + `Display version information including:
	- Application version
	- Git commit hash
	- Build time{{if .Full}}
	- Build hash
	- Go version
	- OS/Architecture{{end}}` + "`" + `,
	Run: runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output version info as JSON")
}

func runVersion(cmd *cobra.Command, args []string) {
	if jsonOutput {
		fmt.Println(GetVersionJSON())
	} else {
		printVersion()
	}
}

// GetVersionInfo returns the version information.
func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version:   Version,
		GitHash:   GitHash,
		BuildTime: BuildTime,
{{if .Full}}
		BuildHash: BuildHash,
		GoVersion: GoVersion,
		GoOS:      GOOS,
		GoArch:    GOARCH,
{{end}}
	}
}

// GetVersionJSON returns the version information as a JSON string.
func GetVersionJSON() string {
	data, err := json.MarshalIndent(GetVersionInfo(), "", "  ")
	if err != nil {
		return "{}"
	}

	return string(data)
}

func printVersion() {
	fmt.Printf("Version:    %s\n", Version)
	fmt.Printf("Git Hash:   %s\n", GitHash)
	fmt.Printf("Build Time: %s\n", BuildTime)
{{if .Full}}
	fmt.Printf("Build Hash: %s\n", BuildHash)
	fmt.Printf("Go Version: %s\n", GoVersion)
	fmt.Printf("OS/Arch:    %s/%s\n", GOOS, GOARCH)
{{end}}
}
`

// ParametersTemplate generates internal/parameters/config.go
const ParametersTemplate = `package parameters

type Service struct {
	Port int    ` + "`" + `yaml:"port"` + "`" + `
	Host string ` + "`" + `yaml:"host"` + "`" + `
}
`

// ServiceTemplate generates internal/service/service.go
const ServiceTemplate = `package service

import (
	"fmt"
	"{{.Module}}/internal/parameters"

	"github.com/inovacc/config"
	"github.com/spf13/cobra"
)

func Handler(_ *cobra.Command, _ []string) error {
	// Get the loaded configuration with type safety
	cfg, err := config.GetServiceConfig[*parameters.Service]()
	if err != nil {
		return fmt.Errorf("failed to get service config: %w", err)
	}

	fmt.Printf("Service running on %s:%d\n", cfg.Host, cfg.Port)

	// Access base configuration
	baseCfg := config.GetBaseConfig()
	fmt.Printf("Application ID: %s\n", baseCfg.AppID)

	return nil
}
`

// ConfigTemplate generates internal/config/config.go (non-service viper)
const ConfigTemplate = `package config

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

// GoModTemplate generates go.mod
const GoModTemplate = `module {{.Module}}

go 1.21

require github.com/spf13/cobra v1.8.0
{{if .UseViper}}
require github.com/spf13/viper v1.18.0
{{end}}
{{if .UseService}}
require github.com/inovacc/config v1.2.2
{{end}}
`

// TaskfileTemplate generates Taskfile.yml with full CI/CD integration
const TaskfileTemplate = `version: '3'

vars:
  APP_NAME: {{.AppName}}
  MAIN_PACKAGE: .
  COVERAGE_FILE: coverage.out
  VERSION:
    sh: git describe --tags --always --dirty 2>/dev/null || echo "dev"
  COMMIT:
    sh: git rev-parse --short HEAD 2>/dev/null || echo "none"
  BUILD_DATE:
    sh: date -u +"%Y-%m-%dT%H:%M:%SZ"
{{if .Full}}
  GITHUB_OWNER:
    sh: echo "${GITHUB_OWNER:-{{.Author}}}"
{{end}}

tasks:
  # === INFORMATION ===
  default:
    desc: List all available tasks
    cmds:
      - task --list
{{if .Full}}
  # === CODE GENERATION ===
  generate:
    desc: Generate version information
    cmds:
      - go generate ./...
{{end}}
  # === BUILD TASKS ===
  build:
    desc: Build the application
    cmds:
      - go build -ldflags "-X {{.Module}}/cmd.Version=$VERSION -X {{.Module}}/cmd.GitHash=$COMMIT -X {{.Module}}/cmd.BuildTime=$BUILD_DATE" -o $APP_NAME
    vars:
      VERSION: "{{"{{"}} .VERSION {{"}}"}}"
      COMMIT: "{{"{{"}} .COMMIT {{"}}"}}"
      BUILD_DATE: "{{"{{"}} .BUILD_DATE {{"}}"}}"
      APP_NAME: "{{"{{"}} .APP_NAME {{"}}"}}"
{{if .Full}}
  build:dev:
    desc: Build development snapshot with goreleaser
    deps: [generate]
    env:
      GITHUB_OWNER: "{{"{{"}} .GITHUB_OWNER {{"}}"}}"
    cmds:
      - goreleaser build --snapshot --clean

  build:prod:
    desc: Build production snapshot with goreleaser
    deps: [generate]
    env:
      GITHUB_OWNER: "{{"{{"}} .GITHUB_OWNER {{"}}"}}"
    cmds:
      - goreleaser --snapshot --skip=publish,announce --clean
{{end}}
  # === RUN TASKS ===
  run:
    desc: Run the application
    cmds:
      - go run {{"{{"}} .MAIN_PACKAGE {{"}}"}} $CLI_ARGS
    vars:
      CLI_ARGS: "{{"{{"}} .CLI_ARGS {{"}}"}}"

  # === TEST TASKS ===
  test:
    desc: Run all tests with coverage
    cmds:
      - go test -v -race -coverprofile={{"{{"}} .COVERAGE_FILE {{"}}"}} ./...

  test:coverage:
    desc: Generate HTML coverage report
    deps: [test]
    cmds:
      - go tool cover -html={{"{{"}} .COVERAGE_FILE {{"}}"}} -o coverage.html
      - echo "Coverage report generated at coverage.html"

  test:cover:
    desc: Run tests and show coverage percentage
    cmds:
      - go test -coverprofile={{"{{"}} .COVERAGE_FILE {{"}}"}} ./...
      - cmd: go tool cover -func={{"{{"}} .COVERAGE_FILE {{"}}"}} | findstr "total:"
        platforms: [windows]
      - cmd: go tool cover -func={{"{{"}} .COVERAGE_FILE {{"}}"}} | grep "total:"
        platforms: [linux, darwin]

  test:unit:
    desc: Run unit tests only (skip integration)
    cmds:
      - go test -v -short ./...

  # === CODE QUALITY ===
  fmt:
    desc: Format all Go code
    cmds:
      - go fmt ./...
      - goimports -w .

  vet:
    desc: Run go vet static analysis
    cmds:
      - go vet ./...

  lint:
    desc: Run golangci-lint
    cmds:
      - golangci-lint run ./...

  lint:fix:
    desc: Run linter and auto-fix issues
    cmds:
      - golangci-lint run --fix ./...

  check:
    desc: Run all quality checks (fmt, vet, lint, test)
    cmds:
      - task: fmt
      - task: vet
      - task: lint
      - task: test

  # === DEPENDENCY MANAGEMENT ===
  deps:
    desc: Download and tidy dependencies
    cmds:
      - go mod download
      - go mod tidy
      - go mod verify

  deps:upgrade:
    desc: Upgrade all dependencies to latest versions
    cmds:
      - go get -u ./...
      - go mod tidy
      - go mod verify

  # === CLEANUP ===
  clean:
    desc: Clean build artifacts and generated files
    cmds:
      - rm -f {{"{{"}} .APP_NAME {{"}}"}}
      - rm -f {{"{{"}} .COVERAGE_FILE {{"}}"}} coverage.html
{{if .Full}}
      - rm -rf dist/
{{end}}
  # === INSTALL ===
  install:
    desc: Install the application
    cmds:
      - go install -ldflags "-X {{.Module}}/cmd.Version=$VERSION -X {{.Module}}/cmd.GitHash=$COMMIT -X {{.Module}}/cmd.BuildTime=$BUILD_DATE"
    vars:
      VERSION: "{{"{{"}} .VERSION {{"}}"}}"
      COMMIT: "{{"{{"}} .COMMIT {{"}}"}}"
      BUILD_DATE: "{{"{{"}} .BUILD_DATE {{"}}"}}"
{{if .Full}}
  # === RELEASE TASKS ===
  release:
    desc: Create a production release with goreleaser (requires git tag)
    deps: [generate]
    env:
      GITHUB_OWNER: "{{"{{"}} .GITHUB_OWNER {{"}}"}}"
    cmds:
      - goreleaser release --clean

  release:snapshot:
    desc: Create a snapshot release (no git tag required)
    deps: [generate]
    env:
      GITHUB_OWNER: "{{"{{"}} .GITHUB_OWNER {{"}}"}}"
    cmds:
      - goreleaser release --snapshot --clean

  release:check:
    desc: Validate goreleaser configuration
    env:
      GITHUB_OWNER: "{{"{{"}} .GITHUB_OWNER {{"}}"}}"
    cmds:
      - goreleaser check
{{end}}
`

// GitignoreTemplate generates .gitignore
const GitignoreTemplate = `# Binaries
{{.AppName}}
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Go workspace
go.work
go.work.sum

# Environment
.env

# IDE - JetBrains
.idea/**/workspace.xml
.idea/**/tasks.xml
.idea/**/usage.statistics.xml
.idea/**/dictionaries
.idea/**/shelf
.idea/**/aws.xml
.idea/**/contentModel.xml
.idea/**/dataSources/
.idea/**/dataSources.ids
.idea/**/dataSources.local.xml
.idea/**/sqlDataSources.xml
.idea/**/dynamic.xml
.idea/**/uiDesigner.xml
.idea/**/dbnavigator.xml
.idea/**/gradle.xml
.idea/**/libraries
.idea/**/mongoSettings.xml
.idea/**/replstate.xml
.idea/**/sonarlint/
.idea/**/httpRequests
.idea/**/caches/build_file_checksums.ser
.idea_modules/
*.iws
out/

# IDE - VSCode
.vscode/

# IDE - other
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Build artifacts
dist/
bin/
cmake-build-*/

# Generated files
cmd/version.go
.go-version
/pkg/pb/

# Coverage
coverage.out
coverage.html
`

// EditorConfigTemplate generates .editorconfig
const EditorConfigTemplate = `# EditorConfig helps maintain consistent coding styles
# More: https://editorconfig.org

root = true

[*]
charset = utf-8
end_of_line = lf
indent_style = space
indent_size = 4
tab_width = 4
insert_final_newline = false
max_line_length = 120

# Go files: use tabs
[*.{go,go2}]
indent_style = tab

# Shell scripts
[*.{sh,bash,zsh}]
indent_size = 2
tab_width = 2

# Config and infra files
[*.{conf,hcl,nomad,tf,tfvars}]
indent_size = 2
tab_width = 2

# Protocol Buffers
[*.proto]
indent_size = 2
tab_width = 2

# JSON-like files
[*.{har,jsb2,jsb3,json,jsonc,babelrc,eslintrc,prettierrc,stylelintrc}]
indent_size = 2

# Markdown
[*.md]
indent_size = 2
tab_width = 2
trim_trailing_whitespace = true
max_line_length = 200

# YAML
[*.{yml,yaml}]
indent_size = 2
tab_width = 2
insert_final_newline = true
trim_trailing_whitespace = true
max_line_length = 200
`

// GoreleaserTemplate generates .goreleaser.yaml
const GoreleaserTemplate = `# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
# vim: set ts=2 sw=2 tw=0 fo=cnqoj

version: 2

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64

archives:
  - formats: [tar.gz]
    name_template: >-
      {{"{{"}} .ProjectName {{"}}"}}_
      {{"{{"}}- title .Os {{"}}"}}_
      {{"{{"}}- if eq .Arch "amd64" {{"}}"}}x86_64
      {{"{{"}}- else if eq .Arch "386" {{"}}"}}i386
      {{"{{"}}- else {{"}}"}}{{"{{"}}.Arch{{"}}"}}{{"{{"}} end {{"}}"}}
      {{"{{"}}- if .Arm {{"}}"}}v{{"{{"}}.Arm{{"}}"}}{{"{{"}} end {{"}}"}}
    format_overrides:
      - goos: windows
        formats: [zip]

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"

release:
  footer: >-

    ---

    Released by [GoReleaser](https://github.com/goreleaser/goreleaser).
`

// GolangciLintTemplate generates .golangci.yml (v2 format)
const GolangciLintTemplate = `# https://golangci-lint.run/docs/welcome/install/

version: "2"

run:
  timeout: 2m

linters:
  default: all
  enable:
    - govet
    - errcheck
    - bodyclose
    - copyloopvar
    - dogsled
    - errorlint
    - gocheckcompilerdirectives
    - gocritic
    - goprintffuncname
    - ineffassign
    - misspell
    - nakedret
    - nolintlint
    - staticcheck
    - unconvert
    - unparam
    - unused

  disable:
    - tagliatelle
    - gochecknoglobals
    - mnd
    - testpackage
    - varnamelen
    - paralleltest
    - gochecknoinits
    - funlen
    - cyclop
    - err113
    - wsl
    - exhaustruct
    - forcetypeassert
    - funcorder
    - gocognit
    - godot
    - godox
    - gosec
    - gosmopolitan
    - intrange
    - ireturn
    - lll
    - nestif
    - nlreturn
    - noctx
    - noinlineerr
    - nonamedreturns
    - perfsprint
    - revive
    - tagalign
    - testifylint
    - thelper
    - usetesting
    - wrapcheck
    - depguard
    - dupl
    - goconst
    - gocyclo
    - whitespace
`

// ReadmeTemplate generates README.md
const ReadmeTemplate = `# {{.AppName}}

{{.Description}}

## Installation

` + "```bash" + `
go install {{.Module}}@latest
` + "```" + `

## Usage

` + "```bash" + `
{{.AppName}} --help
` + "```" + `

## Commands

| Command | Description |
|---------|-------------|
| ` + "`version`" + ` | Print version information |

## Development

` + "```bash" + `
# Build
task build

# Run
task run

# Test
task test

# Lint
task lint
` + "```" + `
{{if .Full}}
## Release

` + "```bash" + `
# Create a snapshot release
task release:snapshot

# Create a production release (requires git tag)
git tag v1.0.0
task release
` + "```" + `
{{end}}
## License

{{if .License}}{{.License}}{{else}}See LICENSE file{{end}}
`

// ToolsTemplate generates tools.go for build dependencies
const ToolsTemplate = `//go:build tools

//go:generate go get github.com/inovacc/genversioninfo
//go:generate go run ./scripts/genversion/genversion.go

package tools

import (
	_ "github.com/inovacc/genversioninfo"
)
`

// CommandTemplate generates a new command file
const CommandTemplate = `package cmd

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

// WorkflowBuildTemplate generates .github/workflows/build.yml
const WorkflowBuildTemplate = `name: Build

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        os: [linux, windows, darwin]
        arch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Build
        env:
          GOOS: ${{"{{"}} matrix.os {{"}}"}}
          GOARCH: ${{"{{"}} matrix.arch {{"}}"}}
          CGO_ENABLED: 0
        run: go build -v ./...
`

// WorkflowTestTemplate generates .github/workflows/test.yml
const WorkflowTestTemplate = `name: Test

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Test
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
`

// WorkflowReleaseTemplate generates .github/workflows/release.yaml
const WorkflowReleaseTemplate = `name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{"{{"}} secrets.GITHUB_TOKEN {{"}}"}}
`

// MITLicense License templates
const MITLicense = `MIT License

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

const ApacheLicense = `Copyright %d %s

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

const BSDLicense = `BSD 3-Clause License

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
