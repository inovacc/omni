package templates

// CmdtreeTemplate generates cmd/cmdtree.go for scaffolded projects.
// This template has NO Go template variables — it's literal Go source.
// The cmdtree command is app-agnostic; it reads everything from rootCmd at runtime.
const CmdtreeTemplate = "package cmd\n" +
	"\n" +
	"import (\n" +
	"\t\"bytes\"\n" +
	"\t\"encoding/json\"\n" +
	"\t\"fmt\"\n" +
	"\t\"io\"\n" +
	"\t\"strings\"\n" +
	"\n" +
	"\t\"github.com/spf13/cobra\"\n" +
	"\t\"github.com/spf13/pflag\"\n" +
	")\n" +
	"\n" +
	"// ASCII tree characters for consistent width across all terminals\n" +
	"const (\n" +
	"\ttreeMiddle  = \"+-- \"\n" +
	"\ttreeLast    = \"\\\\-- \"\n" +
	"\ttreeIndent  = \"|   \"\n" +
	"\ttreeSpace   = \"    \"\n" +
	"\tincludeHelp = true\n" +
	"\tshowHidden  = true\n" +
	"\tmaxDescLen  = 40\n" +
	"\tcommentCol  = 45\n" +
	")\n" +
	"\n" +
	"// cmdtree flags\n" +
	"var (\n" +
	"\tcmdtreeVerbose bool\n" +
	"\tcmdtreeBrief   bool\n" +
	"\tcmdtreeCommand string\n" +
	"\tcmdtreeJSON    bool\n" +
	")\n" +
	"\n" +
	"// FlagDetail represents a single flag's information\n" +
	"type FlagDetail struct {\n" +
	"\tName        string " + "`" + "json:\"name\"" + "`" + "\n" +
	"\tShorthand   string " + "`" + "json:\"shorthand,omitempty\"" + "`" + "\n" +
	"\tType        string " + "`" + "json:\"type\"" + "`" + "\n" +
	"\tDefault     string " + "`" + "json:\"default\"" + "`" + "\n" +
	"\tDescription string " + "`" + "json:\"description\"" + "`" + "\n" +
	"}\n" +
	"\n" +
	"// CommandDetail represents a command's full information\n" +
	"type CommandDetail struct {\n" +
	"\tName        string          " + "`" + "json:\"name\"" + "`" + "\n" +
	"\tUse         string          " + "`" + "json:\"use\"" + "`" + "\n" +
	"\tShort       string          " + "`" + "json:\"short\"" + "`" + "\n" +
	"\tLong        string          " + "`" + "json:\"long,omitempty\"" + "`" + "\n" +
	"\tFlags       []FlagDetail    " + "`" + "json:\"flags,omitempty\"" + "`" + "\n" +
	"\tGlobalFlags []FlagDetail    " + "`" + "json:\"global_flags,omitempty\"" + "`" + "\n" +
	"\tSubcommands []CommandDetail " + "`" + "json:\"commands,omitempty\"" + "`" + "\n" +
	"}\n" +
	"\n" +
	"var cmdtreeCmd = &cobra.Command{\n" +
	"\tUse:   \"cmdtree\",\n" +
	"\tShort: \"Display command tree visualization\",\n" +
	"\tLong:  \"Display a tree visualization of all available commands with descriptions.\",\n" +
	"\tRunE: func(cmd *cobra.Command, args []string) error {\n" +
	"\t\tif cmdtreeJSON {\n" +
	"\t\t\treturn printJSONTree(cmd, rootCmd)\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tif cmdtreeCommand != \"\" {\n" +
	"\t\t\treturn printSingleCommand(cmd, rootCmd, cmdtreeCommand)\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tvar tree bytes.Buffer\n" +
	"\n" +
	"\t\ttree.WriteString(\"# Command Tree\\n\\n" + "```" + "\\n\")\n" +
	"\t\tif cmdtreeBrief || !cmdtreeVerbose {\n" +
	"\t\t\ttree.Write(buildTree(rootCmd))\n" +
	"\t\t} else {\n" +
	"\t\t\ttree.Write(buildVerboseTree(rootCmd))\n" +
	"\t\t}\n" +
	"\t\ttree.WriteString(\"" + "```" + "\\n\")\n" +
	"\n" +
	"\t\t_, _ = fmt.Fprintln(cmd.OutOrStdout(), tree.String())\n" +
	"\t\treturn nil\n" +
	"\t},\n" +
	"}\n" +
	"\n" +
	"func init() {\n" +
	"\trootCmd.AddCommand(cmdtreeCmd)\n" +
	"\n" +
	"\tcmdtreeCmd.Flags().BoolVarP(&cmdtreeVerbose, \"verbose\", \"v\", true, \"Show full details for all commands (default)\")\n" +
	"\tcmdtreeCmd.Flags().BoolVarP(&cmdtreeBrief, \"brief\", \"b\", false, \"Show compact tree with short descriptions only\")\n" +
	"\tcmdtreeCmd.Flags().StringVarP(&cmdtreeCommand, \"command\", \"c\", \"\", \"Show details for a specific command only\")\n" +
	"\tcmdtreeCmd.Flags().BoolVar(&cmdtreeJSON, \"json\", false, \"Output in JSON format\")\n" +
	"}\n" +
	"\n" +
	"func buildTree(root *cobra.Command) []byte {\n" +
	"\tvar buf bytes.Buffer\n" +
	"\n" +
	"\t_, _ = buf.WriteString(fmt.Sprintf(\"%s\\n\", root.Use))\n" +
	"\n" +
	"\t// Show global persistent flags\n" +
	"\tglobalFlags := collectPersistentFlags(root)\n" +
	"\tif len(globalFlags) > 0 {\n" +
	"\t\t_, _ = fmt.Fprintf(&buf, \"%s\\n\", treeIndent)\n" +
	"\t\t_, _ = fmt.Fprintf(&buf, \"%sGlobal flags:\\n\", treeIndent)\n" +
	"\n" +
	"\t\tfor _, f := range globalFlags {\n" +
	"\t\t\tprintFlagDetail(&buf, treeIndent+\"  \", f)\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t_, _ = fmt.Fprintf(&buf, \"%s\\n\", treeIndent)\n" +
	"\t}\n" +
	"\n" +
	"\tprintCommands(&buf, root.Commands(), \"\")\n" +
	"\n" +
	"\treturn buf.Bytes()\n" +
	"}\n" +
	"\n" +
	"func buildVerboseTree(root *cobra.Command) []byte {\n" +
	"\tvar buf bytes.Buffer\n" +
	"\n" +
	"\t_, _ = buf.WriteString(fmt.Sprintf(\"%s\\n\", root.Use))\n" +
	"\n" +
	"\t// Show global persistent flags\n" +
	"\tglobalFlags := collectPersistentFlags(root)\n" +
	"\tif len(globalFlags) > 0 {\n" +
	"\t\t_, _ = fmt.Fprintf(&buf, \"%s\\n\", treeIndent)\n" +
	"\t\t_, _ = fmt.Fprintf(&buf, \"%sGlobal Flags:\\n\", treeIndent)\n" +
	"\n" +
	"\t\tfor _, f := range globalFlags {\n" +
	"\t\t\tprintFlagDetail(&buf, treeIndent+\"  \", f)\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t_, _ = fmt.Fprintf(&buf, \"%s\\n\", treeIndent)\n" +
	"\t}\n" +
	"\n" +
	"\tprintVerboseCommands(&buf, root.Commands(), \"\")\n" +
	"\n" +
	"\treturn buf.Bytes()\n" +
	"}\n" +
	"\n" +
	"func printCommands(w io.Writer, commands []*cobra.Command, prefix string) {\n" +
	"\tvar visible []*cobra.Command\n" +
	"\n" +
	"\tfor _, c := range commands {\n" +
	"\t\tif !includeHelp && (c.Name() == \"help\" || c.Name() == \"completion\") {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tif !showHidden && c.Hidden {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tvisible = append(visible, c)\n" +
	"\t}\n" +
	"\n" +
	"\tfor i, c := range visible {\n" +
	"\t\tisLast := i == len(visible)-1\n" +
	"\n" +
	"\t\tconnector := treeMiddle\n" +
	"\t\tif isLast {\n" +
	"\t\t\tconnector = treeLast\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tdesc := c.Short\n" +
	"\t\tif desc == \"\" {\n" +
	"\t\t\tdesc = c.Long\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tif len(desc) > maxDescLen {\n" +
	"\t\t\tdesc = fmt.Sprintf(\"%s...\", desc[:maxDescLen-3])\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tcmdPart := prefix + connector + c.Name()\n" +
	"\n" +
	"\t\tpadding := max(commentCol-len(cmdPart), 2)\n" +
	"\n" +
	"\t\t_, _ = fmt.Fprintf(w, \"%s%s# %s\\n\", cmdPart, strings.Repeat(\" \", padding), desc)\n" +
	"\n" +
	"\t\tif len(c.Commands()) > 0 {\n" +
	"\t\t\tnewPrefix := prefix + treeIndent\n" +
	"\t\t\tif isLast {\n" +
	"\t\t\t\tnewPrefix = prefix + treeSpace\n" +
	"\t\t\t}\n" +
	"\n" +
	"\t\t\tprintCommands(w, c.Commands(), newPrefix)\n" +
	"\t\t}\n" +
	"\t}\n" +
	"}\n" +
	"\n" +
	"func printVerboseCommands(w io.Writer, commands []*cobra.Command, prefix string) {\n" +
	"\tvar visible []*cobra.Command\n" +
	"\n" +
	"\tfor _, c := range commands {\n" +
	"\t\tif !includeHelp && (c.Name() == \"help\" || c.Name() == \"completion\") {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tif !showHidden && c.Hidden {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tvisible = append(visible, c)\n" +
	"\t}\n" +
	"\n" +
	"\tfor i, c := range visible {\n" +
	"\t\tisLast := i == len(visible)-1\n" +
	"\n" +
	"\t\tconnector := treeMiddle\n" +
	"\t\tif isLast {\n" +
	"\t\t\tconnector = treeLast\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t// Print command name\n" +
	"\t\t_, _ = fmt.Fprintf(w, \"%s%s%s\\n\", prefix, connector, c.Name())\n" +
	"\n" +
	"\t\t// Determine the continuation prefix for details\n" +
	"\t\tdetailPrefix := prefix + treeIndent\n" +
	"\t\tif isLast {\n" +
	"\t\t\tdetailPrefix = prefix + treeSpace\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t// Print usage\n" +
	"\t\t_, _ = fmt.Fprintf(w, \"%sUsage: %s\\n\", detailPrefix, c.UseLine())\n" +
	"\n" +
	"\t\t// Print description\n" +
	"\t\tdesc := c.Short\n" +
	"\t\tif c.Long != \"\" {\n" +
	"\t\t\tdesc = c.Long\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tif desc != \"\" {\n" +
	"\t\t\t_, _ = fmt.Fprintf(w, \"%sDescription: %s\\n\", detailPrefix, desc)\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t// Print flags\n" +
	"\t\tflags := collectFlags(c)\n" +
	"\t\tif len(flags) > 0 {\n" +
	"\t\t\t_, _ = fmt.Fprintf(w, \"%s\\n\", detailPrefix)\n" +
	"\n" +
	"\t\t\t_, _ = fmt.Fprintf(w, \"%sFlags:\\n\", detailPrefix)\n" +
	"\t\t\tfor _, f := range flags {\n" +
	"\t\t\t\tprintFlagDetail(w, detailPrefix+\"  \", f)\n" +
	"\t\t\t}\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t// Add blank line between commands\n" +
	"\t\t_, _ = fmt.Fprintf(w, \"%s\\n\", detailPrefix)\n" +
	"\n" +
	"\t\t// Handle subcommands\n" +
	"\t\tif len(c.Commands()) > 0 {\n" +
	"\t\t\tprintVerboseCommands(w, c.Commands(), detailPrefix)\n" +
	"\t\t}\n" +
	"\t}\n" +
	"}\n" +
	"\n" +
	"func collectFlags(cmd *cobra.Command) []FlagDetail {\n" +
	"\tvar flags []FlagDetail\n" +
	"\n" +
	"\tcmd.LocalFlags().VisitAll(func(f *pflag.Flag) {\n" +
	"\t\tif f.Name == \"help\" {\n" +
	"\t\t\treturn\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t// Skip persistent flags (shown separately in global flags section)\n" +
	"\t\tif cmd.PersistentFlags().Lookup(f.Name) != nil {\n" +
	"\t\t\treturn\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tflags = append(flags, FlagDetail{\n" +
	"\t\t\tName:        f.Name,\n" +
	"\t\t\tShorthand:   f.Shorthand,\n" +
	"\t\t\tType:        f.Value.Type(),\n" +
	"\t\t\tDefault:     f.DefValue,\n" +
	"\t\t\tDescription: f.Usage,\n" +
	"\t\t})\n" +
	"\t})\n" +
	"\n" +
	"\treturn flags\n" +
	"}\n" +
	"\n" +
	"func collectPersistentFlags(cmd *cobra.Command) []FlagDetail {\n" +
	"\tvar flags []FlagDetail\n" +
	"\n" +
	"\tcmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {\n" +
	"\t\tif f.Name == \"help\" {\n" +
	"\t\t\treturn\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tflags = append(flags, FlagDetail{\n" +
	"\t\t\tName:        f.Name,\n" +
	"\t\t\tShorthand:   f.Shorthand,\n" +
	"\t\t\tType:        f.Value.Type(),\n" +
	"\t\t\tDefault:     f.DefValue,\n" +
	"\t\t\tDescription: f.Usage,\n" +
	"\t\t})\n" +
	"\t})\n" +
	"\n" +
	"\treturn flags\n" +
	"}\n" +
	"\n" +
	"func printFlagDetail(w io.Writer, prefix string, f FlagDetail) {\n" +
	"\tvar flagStr string\n" +
	"\n" +
	"\tif f.Shorthand != \"\" {\n" +
	"\t\tflagStr = fmt.Sprintf(\"-%s, --%s\", f.Shorthand, f.Name)\n" +
	"\t} else {\n" +
	"\t\tflagStr = fmt.Sprintf(\"    --%s\", f.Name)\n" +
	"\t}\n" +
	"\n" +
	"\t// Add type for non-bool flags\n" +
	"\tif f.Type != \"bool\" {\n" +
	"\t\tflagStr += \" \" + f.Type\n" +
	"\t}\n" +
	"\n" +
	"\t// Pad to align descriptions\n" +
	"\tpadding := max(26-len(flagStr), 2)\n" +
	"\t_, _ = fmt.Fprintf(w, \"%s%s%s%s\\n\", prefix, flagStr, strings.Repeat(\" \", padding), f.Description)\n" +
	"}\n" +
	"\n" +
	"func printSingleCommand(cobraCmd *cobra.Command, root *cobra.Command, cmdName string) error {\n" +
	"\ttarget := findCommand(root, cmdName)\n" +
	"\tif target == nil {\n" +
	"\t\treturn fmt.Errorf(\"command not found: %s\", cmdName)\n" +
	"\t}\n" +
	"\n" +
	"\tif cmdtreeJSON {\n" +
	"\t\tdetail := buildCommandDetail(target)\n" +
	"\t\tenc := json.NewEncoder(cobraCmd.OutOrStdout())\n" +
	"\t\tenc.SetIndent(\"\", \"  \")\n" +
	"\n" +
	"\t\tif err := enc.Encode(detail); err != nil {\n" +
	"\t\t\treturn fmt.Errorf(\"json encode: %w\", err)\n" +
	"\t\t}\n" +
	"\n" +
	"\t\treturn nil\n" +
	"\t}\n" +
	"\n" +
	"\tvar buf bytes.Buffer\n" +
	"\n" +
	"\t_, _ = buf.WriteString(fmt.Sprintf(\"# %s\\n\\n\", target.Name()))\n" +
	"\t_, _ = buf.WriteString(fmt.Sprintf(\"Usage: %s\\n\\n\", target.UseLine()))\n" +
	"\n" +
	"\tdesc := target.Short\n" +
	"\tif target.Long != \"\" {\n" +
	"\t\tdesc = target.Long\n" +
	"\t}\n" +
	"\n" +
	"\tif desc != \"\" {\n" +
	"\t\t_, _ = buf.WriteString(fmt.Sprintf(\"Description: %s\\n\\n\", desc))\n" +
	"\t}\n" +
	"\n" +
	"\tflags := collectFlags(target)\n" +
	"\tif len(flags) > 0 {\n" +
	"\t\t_, _ = buf.WriteString(\"Flags:\\n\")\n" +
	"\n" +
	"\t\tfor _, f := range flags {\n" +
	"\t\t\tprintFlagDetail(&buf, \"  \", f)\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t_, _ = buf.WriteString(\"\\n\")\n" +
	"\t}\n" +
	"\n" +
	"\tglobalFlags := collectPersistentFlags(target)\n" +
	"\tif len(globalFlags) > 0 {\n" +
	"\t\t_, _ = buf.WriteString(\"Global Flags:\\n\")\n" +
	"\n" +
	"\t\tfor _, f := range globalFlags {\n" +
	"\t\t\tprintFlagDetail(&buf, \"  \", f)\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t_, _ = buf.WriteString(\"\\n\")\n" +
	"\t}\n" +
	"\n" +
	"\tif len(target.Commands()) > 0 {\n" +
	"\t\t_, _ = buf.WriteString(\"Subcommands:\\n\")\n" +
	"\n" +
	"\t\tfor _, sub := range target.Commands() {\n" +
	"\t\t\tif !showHidden && sub.Hidden {\n" +
	"\t\t\t\tcontinue\n" +
	"\t\t\t}\n" +
	"\n" +
	"\t\t\t_, _ = fmt.Fprintf(&buf, \"  %s - %s\\n\", sub.Name(), sub.Short)\n" +
	"\t\t}\n" +
	"\t}\n" +
	"\n" +
	"\t_, _ = fmt.Fprint(cobraCmd.OutOrStdout(), buf.String())\n" +
	"\n" +
	"\treturn nil\n" +
	"}\n" +
	"\n" +
	"func findCommand(root *cobra.Command, name string) *cobra.Command {\n" +
	"\tif root.Name() == name {\n" +
	"\t\treturn root\n" +
	"\t}\n" +
	"\n" +
	"\tfor _, c := range root.Commands() {\n" +
	"\t\tif c.Name() == name {\n" +
	"\t\t\treturn c\n" +
	"\t\t}\n" +
	"\t\t// Search in subcommands\n" +
	"\t\tif found := findCommand(c, name); found != nil {\n" +
	"\t\t\treturn found\n" +
	"\t\t}\n" +
	"\t}\n" +
	"\n" +
	"\treturn nil\n" +
	"}\n" +
	"\n" +
	"func printJSONTree(cobraCmd *cobra.Command, root *cobra.Command) error {\n" +
	"\tdetail := buildCommandDetail(root)\n" +
	"\n" +
	"\tenc := json.NewEncoder(cobraCmd.OutOrStdout())\n" +
	"\tenc.SetIndent(\"\", \"  \")\n" +
	"\n" +
	"\tif err := enc.Encode(detail); err != nil {\n" +
	"\t\treturn fmt.Errorf(\"json encode: %w\", err)\n" +
	"\t}\n" +
	"\n" +
	"\treturn nil\n" +
	"}\n" +
	"\n" +
	"func buildCommandDetail(cmd *cobra.Command) CommandDetail {\n" +
	"\tdetail := CommandDetail{\n" +
	"\t\tName:        cmd.Name(),\n" +
	"\t\tUse:         cmd.UseLine(),\n" +
	"\t\tShort:       cmd.Short,\n" +
	"\t\tLong:        cmd.Long,\n" +
	"\t\tFlags:       collectFlags(cmd),\n" +
	"\t\tGlobalFlags: collectPersistentFlags(cmd),\n" +
	"\t}\n" +
	"\n" +
	"\tfor _, sub := range cmd.Commands() {\n" +
	"\t\tif !includeHelp && (sub.Name() == \"help\" || sub.Name() == \"completion\") {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tif !showHidden && sub.Hidden {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tdetail.Subcommands = append(detail.Subcommands, buildCommandDetail(sub))\n" +
	"\t}\n" +
	"\n" +
	"\treturn detail\n" +
	"}\n"
