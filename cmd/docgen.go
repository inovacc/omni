package cmd

// helplint:ignore — Long strings need omni-usage examples added in a future pass.

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/spf13/cobra"
)

// docCategoryOrder is the fixed, human-facing ordering of the command-reference
// sections. It is a slice (NOT a map) so the generated docs/COMMANDS.md is
// byte-deterministic — never range over a map to produce output. "Other
// Commands" is the catch-all for any command not assigned in docCategory and
// must appear just before the trailing "Command Tree" section.
var docCategoryOrder = []string{
	"Core Commands",
	"File Operations",
	"Text Processing",
	"System Information",
	"Process (runtime-aware)",
	"Flow Control",
	"Archive & Compression",
	"Hash & Encoding",
	"Data Processing",
	"Security & Random",
	"TUI Pagers",
	"Comparison",
	"Tooling",
	"Other Commands",
}

// otherCategory is the heading for commands not explicitly mapped in
// docCategory. A new command added via a cmd/*.go init() lands here until its
// name is added to docCategory.
const otherCategory = "Other Commands"

// docCategory maps a command name to its docCategoryOrder heading. Seeded from
// the groupings the hand-maintained docs/COMMANDS.md used. Any command not
// present here is emitted under "Other Commands".
var docCategory = map[string]string{
	// Core Commands
	"ls":        "Core Commands",
	"pwd":       "Core Commands",
	"cat":       "Core Commands",
	"date":      "Core Commands",
	"dirname":   "Core Commands",
	"basename":  "Core Commands",
	"realpath":  "Core Commands",
	"path":      "Core Commands",
	"echo":      "Core Commands",
	"testcheck": "Core Commands",

	// File Operations
	"cp":       "File Operations",
	"copy":     "File Operations",
	"mv":       "File Operations",
	"move":     "File Operations",
	"rm":       "File Operations",
	"remove":   "File Operations",
	"mkdir":    "File Operations",
	"rmdir":    "File Operations",
	"touch":    "File Operations",
	"stat":     "File Operations",
	"ln":       "File Operations",
	"readlink": "File Operations",
	"chmod":    "File Operations",
	"chown":    "File Operations",

	// Text Processing
	"grep":   "Text Processing",
	"egrep":  "Text Processing",
	"fgrep":  "Text Processing",
	"rg":     "Text Processing",
	"head":   "Text Processing",
	"tail":   "Text Processing",
	"sort":   "Text Processing",
	"uniq":   "Text Processing",
	"wc":     "Text Processing",
	"cut":    "Text Processing",
	"tr":     "Text Processing",
	"nl":     "Text Processing",
	"paste":  "Text Processing",
	"tac":    "Text Processing",
	"column": "Text Processing",
	"fold":   "Text Processing",
	"join":   "Text Processing",
	"sed":    "Text Processing",
	"awk":    "Text Processing",

	// System Information
	"env":    "System Information",
	"whoami": "System Information",
	"id":     "System Information",
	"uname":  "System Information",
	"uptime": "System Information",
	"free":   "System Information",
	"df":     "System Information",
	"du":     "System Information",
	"ps":     "System Information",
	"kill":   "System Information",
	"time":   "System Information",

	// Process (runtime-aware)
	"gops":   "Process (runtime-aware)",
	"nodeps": "Process (runtime-aware)",
	"pyps":   "Process (runtime-aware)",
	"javaps": "Process (runtime-aware)",

	// Flow Control
	"xargs": "Flow Control",
	"watch": "Flow Control",
	"yes":   "Flow Control",
	"nohup": "Flow Control",
	"pipe":  "Flow Control",

	// Archive & Compression
	"tar":   "Archive & Compression",
	"zip":   "Archive & Compression",
	"unzip": "Archive & Compression",

	// Hash & Encoding
	"hash":      "Hash & Encoding",
	"sha256sum": "Hash & Encoding",
	"sha512sum": "Hash & Encoding",
	"md5sum":    "Hash & Encoding",
	"base64":    "Hash & Encoding",
	"base32":    "Hash & Encoding",
	"base58":    "Hash & Encoding",
	"xxd":       "Hash & Encoding",

	// Data Processing
	"jq":     "Data Processing",
	"yq":     "Data Processing",
	"dotenv": "Data Processing",

	// Security & Random
	"sbom":       "Security & Random",
	"scan":       "Security & Random",
	"attest":     "Security & Random",
	"reprocheck": "Security & Random",
	"encrypt":    "Security & Random",
	"decrypt":    "Security & Random",
	"uuid":       "Security & Random",
	"random":     "Security & Random",

	// TUI Pagers
	"less": "TUI Pagers",
	"more": "TUI Pagers",

	// Comparison
	"diff": "Comparison",

	// Tooling
	"lint":    "Tooling",
	"cmdtree": "Tooling",
	"logger":  "Tooling",
	"version": "Tooling",
}

// GenerateCommandReference writes the canonical omni command reference
// (docs/COMMANDS.md content) to w. Output is deterministic — commands and
// categories are emitted in a fixed sorted order with no timestamps — so a
// CI drift check can diff a fresh regeneration against the committed file.
// It walks the in-process rootCmd tree (no exec, pure Go).
func GenerateCommandReference(w io.Writer) error {
	var sb strings.Builder

	sb.WriteString("# omni Command Reference\n\n")
	sb.WriteString("<!-- This file is auto-generated by tools/cmdref/cmdref.go -->\n")
	sb.WriteString("<!-- Run: go run tools/cmdref/cmdref.go   (or: task docs:commands) -->\n\n")

	// Bucket the visible top-level commands by category. cobra returns
	// rootCmd.Commands() already sorted alphabetically; we re-sort within each
	// category anyway to keep determinism explicit and independent of cobra.
	buckets := make(map[string][]*cobra.Command)
	for _, c := range rootCmd.Commands() {
		if c.Name() == "help" || c.Name() == "completion" || c.Hidden {
			continue
		}

		cat := docCategory[c.Name()]
		if cat == "" {
			cat = otherCategory
		}

		buckets[cat] = append(buckets[cat], c)
	}

	for _, cat := range docCategoryOrder {
		cmds := buckets[cat]
		if len(cmds) == 0 {
			continue
		}

		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name() < cmds[j].Name()
		})

		sb.WriteString(fmt.Sprintf("## %s\n\n", cat))

		for _, c := range cmds {
			writeCommandBlock(&sb, c)
		}
	}

	// Trailing "Command Tree" section — mirrors how docs/COMMANDS.md ends.
	sb.WriteString("## Command Tree\n\n")
	sb.WriteString("```\n")
	sb.Write(buildTree(rootCmd))
	sb.WriteString("```\n")

	if _, err := io.WriteString(w, sb.String()); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("docgen: write: %s", err))
	}

	return nil
}

// writeCommandBlock renders one "### <name> - <Short>" reference block followed
// by a fenced bash usage line (cobra's UseLine(), emitted AS-IS) and indented
// flag lines.
func writeCommandBlock(sb *strings.Builder, c *cobra.Command) {
	sb.WriteString(fmt.Sprintf("### %s - %s\n", c.Name(), c.Short))
	sb.WriteString("```bash\n")
	// c.UseLine() already prepends the parent command path ("omni ") for a
	// subcommand of root — emit it AS-IS. Prepending another "omni " would
	// produce a doubled "omni omni" prefix.
	sb.WriteString(c.UseLine() + "\n")

	for _, f := range collectFlags(c) {
		printFlagDetail(sb, "  ", f)
	}

	sb.WriteString("```\n\n")
}
