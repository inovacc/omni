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
	JSON        bool
	Category    string
	NoStructure bool
}

// AIContext represents the complete AI context document
type AIContext struct {
	App        string            `json:"app"`
	Desc       string            `json:"desc"`
	Categories map[string][]CMD  `json:"categories"`
	Structure  map[string]string `json:"structure,omitempty"`
}

// CMD represents a command
type CMD struct {
	Cmd   string   `json:"cmd"`
	Desc  string   `json:"desc"`
	Flags []string `json:"flags,omitempty"`
	Sub   []string `json:"sub,omitempty"`
}

// categoryMap maps command names to categories
var categoryMap = map[string]string{
	// Core
	"ls": "core", "pwd": "core", "cat": "core", "date": "core", "echo": "core",
	"basename": "core", "dirname": "core", "realpath": "core", "tree": "core",
	"readlink": "core", "yes": "core",

	// File Operations
	"cp": "file", "mv": "file", "rm": "file", "mkdir": "file", "rmdir": "file",
	"touch": "file", "chmod": "file", "chown": "file", "ln": "file",
	"stat": "file", "file": "file", "find": "file", "dd": "file", "sync": "file",

	// Text Processing
	"grep": "text", "egrep": "text", "fgrep": "text", "sed": "text", "awk": "text",
	"head": "text", "tail": "text", "sort": "text", "uniq": "text", "cut": "text",
	"tr": "text", "wc": "text", "nl": "text", "paste": "text", "tac": "text",
	"column": "text", "fold": "text", "join": "text", "shuf": "text", "split": "text",
	"rev": "text", "comm": "text", "cmp": "text", "strings": "text", "diff": "text",
	"expand": "text", "unexpand": "text",

	// System Info
	"env": "sys", "whoami": "sys", "id": "sys", "uname": "sys", "uptime": "sys",
	"df": "sys", "du": "sys", "ps": "sys", "free": "sys", "kill": "sys",
	"arch": "sys", "hostname": "sys", "nproc": "sys", "printenv": "sys", "which": "sys",

	// Archive & Compression
	"tar": "archive", "zip": "archive", "unzip": "archive", "gzip": "archive",
	"gunzip": "archive", "zcat": "archive", "bzip2": "archive", "bunzip2": "archive",
	"bzcat": "archive", "xz": "archive", "unxz": "archive", "xzcat": "archive",

	// Hash & Encoding
	"hash": "hash", "sha256sum": "hash", "sha512sum": "hash", "md5sum": "hash",
	"base64": "hash", "base32": "hash", "base58": "hash",

	// Data Processing
	"jq": "data", "yq": "data", "json": "data", "dotenv": "data",

	// Security
	"encrypt": "security", "decrypt": "security", "uuid": "security",
	"random": "security", "crypt": "security",

	// Database
	"sqlite": "db", "bbolt": "db",

	// Code Generation
	"generate": "codegen",

	// Utilities
	"time": "util", "sleep": "util", "seq": "util", "xargs": "util",
	"watch": "util", "tee": "util", "true": "util", "false": "util", "test": "util",

	// Tooling
	"lint": "tools", "logger": "tools", "cmdtree": "tools", "aicontext": "tools",
	"version": "tools",
}

var categoryNames = map[string]string{
	"core":     "Core",
	"file":     "File",
	"text":     "Text",
	"sys":      "System",
	"archive":  "Archive",
	"hash":     "Hash/Encoding",
	"data":     "Data",
	"security": "Security",
	"db":       "Database",
	"codegen":  "CodeGen",
	"util":     "Utilities",
	"tools":    "Tools",
}

// RunAIContext generates AI context documentation
func RunAIContext(w io.Writer, root *cobra.Command, opts Options) error {
	ctx := buildContext(root, opts)
	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(ctx)
	}

	return writeMarkdown(w, ctx)
}

func buildContext(root *cobra.Command, opts Options) AIContext {
	categories := make(map[string][]CMD)

	for _, c := range root.Commands() {
		if c.Name() == "help" || c.Name() == "completion" || c.Hidden {
			continue
		}

		cat := categoryMap[c.Name()]
		if cat == "" {
			cat = "other"
		}

		if opts.Category != "" && cat != opts.Category {
			continue
		}

		cmd := CMD{
			Cmd:  c.Name(),
			Desc: c.Short,
		}

		// Collect flags (abbreviated)
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "help" {
				return
			}

			flag := "--" + f.Name
			if f.Shorthand != "" {
				flag = "-" + f.Shorthand + "/" + flag
			}

			cmd.Flags = append(cmd.Flags, flag)
		})

		// Collect subcommands
		for _, sub := range c.Commands() {
			if sub.Name() != "help" && !sub.Hidden {
				cmd.Sub = append(cmd.Sub, sub.Name())
			}
		}

		categories[cat] = append(categories[cat], cmd)
	}

	// Sort commands within categories
	for cat := range categories {
		sort.Slice(categories[cat], func(i, j int) bool {
			return categories[cat][i].Cmd < categories[cat][j].Cmd
		})
	}

	ctx := AIContext{
		App:        "omni",
		Desc:       "Cross-platform Go-native shell utilities (100+ commands, no exec, pure Go)",
		Categories: categories,
	}

	if !opts.NoStructure {
		ctx.Structure = map[string]string{
			"cmd/":          "CLI commands (Cobra)",
			"internal/cli/": "Command implementations",
			"tests/":        "Go integration tests",
			"testing/":      "Python black-box tests",
		}
	}

	return ctx
}

func writeMarkdown(w io.Writer, ctx AIContext) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n%s\n\n", ctx.App, ctx.Desc))

	// Sort categories
	cats := make([]string, 0, len(ctx.Categories))
	for cat := range ctx.Categories {
		cats = append(cats, cat)
	}

	sort.Strings(cats)

	for _, cat := range cats {
		cmds := ctx.Categories[cat]

		name := categoryNames[cat]
		if name == "" {
			name = cat
		}

		sb.WriteString(fmt.Sprintf("## %s\n\n", name))

		for _, cmd := range cmds {
			sb.WriteString(fmt.Sprintf("### %s\n%s\n", cmd.Cmd, cmd.Desc))

			if len(cmd.Flags) > 0 {
				sb.WriteString(fmt.Sprintf("Flags: `%s`\n", strings.Join(cmd.Flags, "` `")))
			}

			if len(cmd.Sub) > 0 {
				sb.WriteString(fmt.Sprintf("Sub: `%s`\n", strings.Join(cmd.Sub, "` `")))
			}

			sb.WriteString("\n")
		}
	}

	if len(ctx.Structure) > 0 {
		sb.WriteString("## Structure\n\n")

		keys := make([]string, 0, len(ctx.Structure))
		for k := range ctx.Structure {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("- `%s` %s\n", k, ctx.Structure[k]))
		}
	}

	_, err := io.WriteString(w, sb.String())

	return err
}
