package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// ASCII tree characters for consistent width across all terminals
const (
	treeMiddle  = "+-- "
	treeLast    = "\\-- "
	treeIndent  = "|   "
	treeSpace   = "    "
	includeHelp = true
	showHidden  = true
	maxDescLen  = 40
	commentCol  = 45
)

var cmdtreeCmd = &cobra.Command{
	Use:   "cmdtree",
	Short: "Display command tree visualization",
	Long:  "Display a tree visualization of all available commands with descriptions.",
	Run: func(cmd *cobra.Command, args []string) {
		var tree bytes.Buffer

		tree.WriteString("# Command Tree\n\n```\n")
		tree.Write(buildTree(rootCmd))
		tree.WriteString("```\n")

		cmd.Println(tree.String())
	},
}

func init() {
	rootCmd.AddCommand(cmdtreeCmd)
}

func buildTree(root *cobra.Command) []byte {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s\n", root.Use))
	printCommands(&buf, root.Commands(), "")

	return buf.Bytes()
}

func printCommands(w io.Writer, commands []*cobra.Command, prefix string) {
	var visible []*cobra.Command

	for _, c := range commands {
		if !includeHelp && (c.Name() == "help" || c.Name() == "completion") {
			continue
		}

		if !showHidden && c.Hidden {
			continue
		}

		visible = append(visible, c)
	}

	for i, c := range visible {
		isLast := i == len(visible)-1

		connector := treeMiddle
		if isLast {
			connector = treeLast
		}

		desc := c.Short
		if desc == "" {
			desc = c.Long
		}

		if len(desc) > maxDescLen {
			desc = fmt.Sprintf("%s...", desc[:maxDescLen-3])
		}

		cmdPart := prefix + connector + c.Name()

		padding := max(commentCol-len(cmdPart), 2)

		_, _ = fmt.Fprintf(w, "%s%s# %s\n", cmdPart, strings.Repeat(" ", padding), desc)

		if len(c.Commands()) > 0 {
			newPrefix := prefix + treeIndent
			if isLast {
				newPrefix = prefix + treeSpace
			}

			printCommands(w, c.Commands(), newPrefix)
		}
	}
}
