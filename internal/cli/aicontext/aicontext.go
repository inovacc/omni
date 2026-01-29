package aicontext

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Options configures the aicontext command behavior
type Options struct {
	JSON     bool   // --json: output as structured JSON
	Compact  bool   // --compact: omit examples and long descriptions
	Category string // --category: filter to specific category
}

// AIContext represents the complete AI context document
type AIContext struct {
	Overview     Overview      `json:"overview"`
	Categories   []Category    `json:"categories"`
	Commands     []CommandInfo `json:"commands"`
	LibraryAPI   LibraryAPI    `json:"library_api"`
	Architecture Architecture  `json:"architecture"`
}

// Overview describes the application
type Overview struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Principles  []string `json:"principles"`
	Features    []string `json:"features"`
}

// Category represents a command category
type Category struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
}

// CommandInfo represents detailed command documentation
type CommandInfo struct {
	Name        string     `json:"name"`
	Path        string     `json:"path"`
	Category    string     `json:"category"`
	Short       string     `json:"short"`
	Long        string     `json:"long,omitempty"`
	Usage       string     `json:"usage"`
	Flags       []FlagInfo `json:"flags,omitempty"`
	Examples    []string   `json:"examples,omitempty"`
	Subcommands []string   `json:"subcommands,omitempty"`
	ImportPath  string     `json:"import_path,omitempty"`
}

// FlagInfo represents a command flag
type FlagInfo struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
}

// LibraryAPI describes how to use commands as Go packages
type LibraryAPI struct {
	Description string   `json:"description"`
	Pattern     string   `json:"pattern"`
	Examples    []string `json:"examples"`
}

// Architecture describes the project structure
type Architecture struct {
	Description string            `json:"description"`
	Structure   map[string]string `json:"structure"`
}

// categoryMap maps command names to categories
var categoryMap = map[string]string{
	// Core
	"ls": "Core", "pwd": "Core", "cat": "Core", "date": "Core", "echo": "Core",
	"basename": "Core", "dirname": "Core", "realpath": "Core", "tree": "Core",
	"readlink": "Core", "yes": "Core",

	// File Operations
	"cp": "File Operations", "mv": "File Operations", "rm": "File Operations",
	"mkdir": "File Operations", "rmdir": "File Operations", "touch": "File Operations",
	"chmod": "File Operations", "chown": "File Operations", "ln": "File Operations",
	"stat": "File Operations", "file": "File Operations", "find": "File Operations",
	"dd": "File Operations", "sync": "File Operations",

	// Text Processing
	"grep": "Text Processing", "egrep": "Text Processing", "fgrep": "Text Processing",
	"sed": "Text Processing", "awk": "Text Processing", "head": "Text Processing",
	"tail": "Text Processing", "sort": "Text Processing", "uniq": "Text Processing",
	"cut": "Text Processing", "tr": "Text Processing", "wc": "Text Processing",
	"nl": "Text Processing", "paste": "Text Processing", "tac": "Text Processing",
	"column": "Text Processing", "fold": "Text Processing", "join": "Text Processing",
	"shuf": "Text Processing", "split": "Text Processing", "rev": "Text Processing",
	"comm": "Text Processing", "cmp": "Text Processing", "strings": "Text Processing",
	"diff": "Text Processing", "expand": "Text Processing", "unexpand": "Text Processing",

	// System Info
	"env": "System Info", "whoami": "System Info", "id": "System Info",
	"uname": "System Info", "uptime": "System Info", "df": "System Info",
	"du": "System Info", "ps": "System Info", "free": "System Info",
	"kill": "System Info", "arch": "System Info", "hostname": "System Info",
	"nproc": "System Info", "printenv": "System Info", "which": "System Info",

	// Archive & Compression
	"tar": "Archive", "zip": "Archive", "unzip": "Archive",
	"gzip": "Archive", "gunzip": "Archive", "zcat": "Archive",
	"bzip2": "Archive", "bunzip2": "Archive", "bzcat": "Archive",
	"xz": "Archive", "unxz": "Archive", "xzcat": "Archive",

	// Hash & Encoding
	"hash": "Hash & Encoding", "sha256sum": "Hash & Encoding", "sha512sum": "Hash & Encoding",
	"md5sum": "Hash & Encoding", "base64": "Hash & Encoding", "base32": "Hash & Encoding",
	"base58": "Hash & Encoding",

	// Data Processing
	"jq": "Data Processing", "yq": "Data Processing", "json": "Data Processing",
	"dotenv": "Data Processing",

	// Security
	"encrypt": "Security", "decrypt": "Security", "uuid": "Security",
	"random": "Security", "crypt": "Security",

	// Database
	"sqlite": "Database", "bbolt": "Database",

	// Code Generation
	"generate": "Code Generation",

	// Utilities
	"time": "Utilities", "sleep": "Utilities", "seq": "Utilities",
	"xargs": "Utilities", "watch": "Utilities", "tee": "Utilities",
	"true": "Utilities", "false": "Utilities", "test": "Utilities",

	// Tooling
	"lint": "Tooling", "logger": "Tooling", "cmdtree": "Tooling",
	"aicontext": "Tooling", "version": "Tooling", "help": "Tooling",
	"completion": "Tooling",
}

// categoryDescriptions provides descriptions for each category
var categoryDescriptions = map[string]string{
	"Core":            "Essential file system navigation and basic I/O operations",
	"File Operations": "File manipulation, permissions, and management commands",
	"Text Processing": "Text transformation, filtering, and analysis tools",
	"System Info":     "System information, process management, and environment",
	"Archive":         "Compression and archive management utilities",
	"Hash & Encoding": "Cryptographic hashes and encoding/decoding tools",
	"Data Processing": "JSON, YAML, and structured data manipulation",
	"Security":        "Encryption, random generation, and security utilities",
	"Database":        "Embedded database operations (SQLite, BoltDB)",
	"Code Generation": "Code scaffolding and generation tools",
	"Utilities":       "General-purpose helper utilities",
	"Tooling":         "Development and introspection tools",
}

// getCategory returns the category for a command name
func getCategory(name string) string {
	if cat, ok := categoryMap[name]; ok {
		return cat
	}

	return "Other"
}

// RunAIContext generates AI context documentation
func RunAIContext(w io.Writer, root *cobra.Command, opts Options) error {
	ctx := buildAIContext(root, opts)

	if opts.JSON {
		return writeJSON(w, ctx)
	}

	return writeMarkdown(w, ctx, opts)
}

// buildAIContext constructs the complete AI context
func buildAIContext(root *cobra.Command, opts Options) AIContext {
	commands := collectCommands(root, "", opts)
	categories := buildCategories(commands)

	// Filter by category if specified
	if opts.Category != "" {
		var filtered []CommandInfo

		for _, cmd := range commands {
			if strings.EqualFold(cmd.Category, opts.Category) {
				filtered = append(filtered, cmd)
			}
		}

		commands = filtered

		var filteredCats []Category

		for _, cat := range categories {
			if strings.EqualFold(cat.Name, opts.Category) {
				filteredCats = append(filteredCats, cat)
			}
		}

		categories = filteredCats
	}

	return AIContext{
		Overview:     buildOverview(),
		Categories:   categories,
		Commands:     commands,
		LibraryAPI:   buildLibraryAPI(),
		Architecture: buildArchitecture(),
	}
}

// buildOverview creates the application overview
func buildOverview() Overview {
	return Overview{
		Name:        "omni",
		Description: "Cross-platform, Go-native replacement for common shell utilities designed for Taskfile, CI/CD, and enterprise environments.",
		Principles: []string{
			"No exec: Never spawns external processes",
			"Pure Go: Standard library first, minimal dependencies",
			"Cross-platform: Linux, macOS, Windows support",
			"Library + CLI: All commands usable as Go packages",
			"Safe defaults: Destructive operations require explicit flags",
			"Testable: io.Writer pattern for all output",
		},
		Features: []string{
			"100+ commands implemented in pure Go",
			"JSON output mode for all commands",
			"Structured logging with OpenTelemetry support",
			"No external dependencies at runtime",
			"Consistent flag conventions across commands",
		},
	}
}

// buildLibraryAPI creates the library API documentation
func buildLibraryAPI() LibraryAPI {
	return LibraryAPI{
		Description: "Command implementations live under internal/cli/ and cannot be imported by external projects (Go's internal package restriction). omni is designed as a CLI tool, not a library. To reuse the logic, fork the repository or use omni as a subprocess.",
		Pattern:     "internal/cli/<command>/<command>.go",
		Examples: []string{
			`// Internal structure (not importable externally):
// Each command follows the pattern:
//   - Options struct for configuration
//   - Run<Command>(w io.Writer, args []string, opts Options) error
//   - Helper functions for implementation

// Example: internal/cli/cat/cat.go
type CatOptions struct {
    NumberAll bool
    JSON      bool
}
func RunCat(w io.Writer, args []string, opts CatOptions) error`,
		},
	}
}

// buildArchitecture creates the architecture documentation
func buildArchitecture() Architecture {
	return Architecture{
		Description: "Hexagonal architecture with clear separation between CLI and library layers",
		Structure: map[string]string{
			"cmd/":             "Cobra CLI command definitions (thin wrappers)",
			"internal/cli/":    "Library implementations with Options structs and Run* functions",
			"internal/flags/":  "Feature flags and environment configuration",
			"internal/logger/": "Structured logging with slog",
			"main.go":          "Entry point calling cmd.Execute()",
		},
	}
}

// collectCommands recursively collects all commands
func collectCommands(cmd *cobra.Command, parentPath string, opts Options) []CommandInfo {
	var commands []CommandInfo

	for _, c := range cmd.Commands() {
		// Skip help and completion commands
		if c.Name() == "help" || c.Name() == "completion" {
			continue
		}

		// Skip hidden commands
		if c.Hidden {
			continue
		}

		path := c.Name()
		if parentPath != "" {
			path = parentPath + " " + c.Name()
		}

		info := CommandInfo{
			Name:     c.Name(),
			Path:     path,
			Category: getCategory(c.Name()),
			Short:    c.Short,
			Usage:    c.UseLine(),
		}

		// Include long description unless compact mode
		if !opts.Compact && c.Long != "" {
			info.Long = c.Long
			info.Examples = extractExamples(c.Long)
		}

		// Collect flags
		info.Flags = collectFlags(c)

		// Collect subcommand names
		for _, sub := range c.Commands() {
			if sub.Name() != "help" && !sub.Hidden {
				info.Subcommands = append(info.Subcommands, sub.Name())
			}
		}

		commands = append(commands, info)

		// Recurse into subcommands
		subCommands := collectCommands(c, path, opts)
		commands = append(commands, subCommands...)
	}

	return commands
}

// collectFlags collects flag information from a command
func collectFlags(cmd *cobra.Command) []FlagInfo {
	var flags []FlagInfo

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Skip help flag
		if f.Name == "help" {
			return
		}

		flags = append(flags, FlagInfo{
			Name:        f.Name,
			Shorthand:   f.Shorthand,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		})
	})

	return flags
}

// extractExamples parses examples from long description
func extractExamples(long string) []string {
	var examples []string

	lines := strings.Split(long, "\n")
	inExample := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect example blocks
		if strings.HasPrefix(trimmed, "omni ") || strings.HasPrefix(trimmed, "$ omni ") {
			examples = append(examples, strings.TrimPrefix(trimmed, "$ "))
			inExample = true
		} else if inExample && (trimmed == "" || !strings.HasPrefix(line, "  ")) {
			inExample = false
		}
	}

	return examples
}

// buildCategories groups commands by category
func buildCategories(commands []CommandInfo) []Category {
	catMap := make(map[string][]string)

	for _, cmd := range commands {
		// Only include top-level commands in category listing
		if !strings.Contains(cmd.Path, " ") {
			catMap[cmd.Category] = append(catMap[cmd.Category], cmd.Name)
		}
	}

	categories := make([]Category, 0, len(catMap))

	for name, cmds := range catMap {
		sort.Strings(cmds)
		categories = append(categories, Category{
			Name:        name,
			Description: categoryDescriptions[name],
			Commands:    cmds,
		})
	}

	// Sort categories by name
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	return categories
}

// writeJSON outputs the context as JSON
func writeJSON(w io.Writer, ctx AIContext) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(ctx)
}

// writeMarkdown outputs the context as Markdown
func writeMarkdown(w io.Writer, ctx AIContext, opts Options) error {
	var sb strings.Builder

	// Title
	sb.WriteString("# omni - AI Context Document\n\n")

	// Overview
	sb.WriteString("## Overview\n\n")
	sb.WriteString(ctx.Overview.Description + "\n\n")

	sb.WriteString("### Design Principles\n\n")

	for _, p := range ctx.Overview.Principles {
		sb.WriteString("- " + p + "\n")
	}

	sb.WriteString("\n")

	sb.WriteString("### Key Features\n\n")

	for _, f := range ctx.Overview.Features {
		sb.WriteString("- " + f + "\n")
	}

	sb.WriteString("\n")

	// Categories
	sb.WriteString("## Command Categories\n\n")

	for _, cat := range ctx.Categories {
		sb.WriteString(fmt.Sprintf("### %s\n\n", cat.Name))

		if cat.Description != "" {
			sb.WriteString(cat.Description + "\n\n")
		}

		sb.WriteString("Commands: `" + strings.Join(cat.Commands, "`, `") + "`\n\n")
	}

	// Command Reference
	sb.WriteString("## Complete Command Reference\n\n")

	for _, cmd := range ctx.Commands {
		sb.WriteString(fmt.Sprintf("### %s\n\n", cmd.Path))
		sb.WriteString(fmt.Sprintf("**Category:** %s\n\n", cmd.Category))
		sb.WriteString(fmt.Sprintf("**Usage:** `%s`\n\n", cmd.Usage))
		sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", cmd.Short))

		if !opts.Compact && cmd.Long != "" {
			sb.WriteString("**Details:**\n\n")
			sb.WriteString(cmd.Long + "\n\n")
		}

		if len(cmd.Flags) > 0 {
			sb.WriteString("**Flags:**\n\n")
			sb.WriteString("| Flag | Type | Default | Description |\n")
			sb.WriteString("|------|------|---------|-------------|\n")

			for _, f := range cmd.Flags {
				flag := "--" + f.Name
				if f.Shorthand != "" {
					flag = "-" + f.Shorthand + ", " + flag
				}

				def := f.Default
				if def == "" {
					def = "-"
				}

				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
					flag, f.Type, def, f.Description))
			}

			sb.WriteString("\n")
		}

		if !opts.Compact && len(cmd.Examples) > 0 {
			sb.WriteString("**Examples:**\n\n```bash\n")

			for _, ex := range cmd.Examples {
				sb.WriteString(ex + "\n")
			}

			sb.WriteString("```\n\n")
		}

		if len(cmd.Subcommands) > 0 {
			sb.WriteString(fmt.Sprintf("**Subcommands:** `%s`\n\n", strings.Join(cmd.Subcommands, "`, `")))
		}

		sb.WriteString("---\n\n")
	}

	// Library API
	sb.WriteString("## Library API\n\n")
	sb.WriteString(ctx.LibraryAPI.Description + "\n\n")
	sb.WriteString(fmt.Sprintf("**Import pattern:** `%s`\n\n", ctx.LibraryAPI.Pattern))

	if !opts.Compact {
		sb.WriteString("**Examples:**\n\n")

		for _, ex := range ctx.LibraryAPI.Examples {
			sb.WriteString("```go\n" + ex + "\n```\n\n")
		}
	}

	// Architecture
	sb.WriteString("## Architecture\n\n")
	sb.WriteString(ctx.Architecture.Description + "\n\n")
	sb.WriteString("```\n")
	sb.WriteString("omni/\n")

	// Sort keys for consistent output
	keys := make([]string, 0, len(ctx.Architecture.Structure))
	for k := range ctx.Architecture.Structure {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("  %s  # %s\n", k, ctx.Architecture.Structure[k]))
	}

	sb.WriteString("```\n")

	_, err := io.WriteString(w, sb.String())

	return err
}
