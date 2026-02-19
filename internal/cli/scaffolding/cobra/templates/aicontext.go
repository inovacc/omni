package templates

// AIContextTemplate generates cmd/aicontext.go for scaffolded projects.
// Template variables: {{.AppName}}, {{.Description}}
// Note: \x60 is the backtick character, used to avoid raw string issues.
const AIContextTemplate = "package cmd\n" +
	"\n" +
	"import (\n" +
	"\t\"encoding/json\"\n" +
	"\t\"fmt\"\n" +
	"\t\"strings\"\n" +
	"\n" +
	"\t\"github.com/spf13/cobra\"\n" +
	"\t\"github.com/spf13/pflag\"\n" +
	")\n" +
	"\n" +
	"var (\n" +
	"\taicontextJSON    bool\n" +
	"\taicontextCompact bool\n" +
	")\n" +
	"\n" +
	"var aicontextCmd = &cobra.Command{\n" +
	"\tUse:   \"aicontext\",\n" +
	"\tShort: \"Generate AI context documentation\",\n" +
	"\tLong: \"Generate structured documentation about {{.AppName}} for use by AI tools.\\n\" +\n" +
	"\t\t\"\\n\" +\n" +
	"\t\t\"Outputs a markdown (or JSON) reference of all commands, flags, and usage\\n\" +\n" +
	"\t\t\"patterns that AI assistants can consume as context.\\n\" +\n" +
	"\t\t\"\\n\" +\n" +
	"\t\t\"Examples:\\n\" +\n" +
	"\t\t\"  {{.AppName}} aicontext             # Markdown output (default)\\n\" +\n" +
	"\t\t\"  {{.AppName}} aicontext --json      # Structured JSON\\n\" +\n" +
	"\t\t\"  {{.AppName}} aicontext --compact   # Shorter output\",\n" +
	"\tRunE: runAIContext,\n" +
	"}\n" +
	"\n" +
	"func init() {\n" +
	"\trootCmd.AddCommand(aicontextCmd)\n" +
	"\n" +
	"\taicontextCmd.Flags().BoolVar(&aicontextJSON, \"json\", false, \"Output in JSON format\")\n" +
	"\taicontextCmd.Flags().BoolVar(&aicontextCompact, \"compact\", false, \"Shorter output\")\n" +
	"}\n" +
	"\n" +
	"// aiCommandInfo represents a command for JSON output\n" +
	"type aiCommandInfo struct {\n" +
	"\tName        string          \x60json:\"name\"\x60\n" +
	"\tUsage       string          \x60json:\"usage\"\x60\n" +
	"\tDescription string          \x60json:\"description\"\x60\n" +
	"\tFlags       []aiFlagInfo    \x60json:\"flags,omitempty\"\x60\n" +
	"\tSubcommands []aiCommandInfo \x60json:\"subcommands,omitempty\"\x60\n" +
	"}\n" +
	"\n" +
	"// aiFlagInfo represents a flag for JSON output\n" +
	"type aiFlagInfo struct {\n" +
	"\tName        string \x60json:\"name\"\x60\n" +
	"\tShorthand   string \x60json:\"shorthand,omitempty\"\x60\n" +
	"\tType        string \x60json:\"type\"\x60\n" +
	"\tDefault     string \x60json:\"default\"\x60\n" +
	"\tDescription string \x60json:\"description\"\x60\n" +
	"\tGlobal      bool   \x60json:\"global,omitempty\"\x60\n" +
	"}\n" +
	"\n" +
	"// aiContextDoc represents the full AI context for JSON output\n" +
	"type aiContextDoc struct {\n" +
	"\tTool        string          \x60json:\"tool\"\x60\n" +
	"\tVersion     string          \x60json:\"version\"\x60\n" +
	"\tDescription string          \x60json:\"description\"\x60\n" +
	"\tGlobalFlags []aiFlagInfo    \x60json:\"global_flags,omitempty\"\x60\n" +
	"\tCommands    []aiCommandInfo \x60json:\"commands\"\x60\n" +
	"}\n" +
	"\n" +
	"func runAIContext(cmd *cobra.Command, _ []string) error {\n" +
	"\tif aicontextJSON {\n" +
	"\t\treturn printAIContextJSON(cmd)\n" +
	"\t}\n" +
	"\n" +
	"\tif aicontextCompact {\n" +
	"\t\treturn printAIContextCompact(cmd)\n" +
	"\t}\n" +
	"\n" +
	"\treturn printAIContextMarkdown(cmd)\n" +
	"}\n" +
	"\n" +
	"func printAIContextMarkdown(cmd *cobra.Command) error {\n" +
	"\tvar b strings.Builder\n" +
	"\n" +
	"\tb.WriteString(\"# {{.AppName}} - AI Context\\n\\n\")\n" +
	"\tb.WriteString(\"## Overview\\n\\n\")\n" +
	"\tb.WriteString(\"{{.Description}}\\n\\n\")\n" +
	"\n" +
	"\t// Global flags\n" +
	"\tvar globalFlags []FlagDetail\n" +
	"\trootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {\n" +
	"\t\tif f.Name == \"help\" {\n" +
	"\t\t\treturn\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tglobalFlags = append(globalFlags, FlagDetail{\n" +
	"\t\t\tName:        f.Name,\n" +
	"\t\t\tShorthand:   f.Shorthand,\n" +
	"\t\t\tType:        f.Value.Type(),\n" +
	"\t\t\tDefault:     f.DefValue,\n" +
	"\t\t\tDescription: f.Usage,\n" +
	"\t\t})\n" +
	"\t})\n" +
	"\n" +
	"\tif len(globalFlags) > 0 {\n" +
	"\t\tb.WriteString(\"## Global Flags\\n\\n\")\n" +
	"\t\tb.WriteString(\"These flags apply to all commands:\\n\\n\")\n" +
	"\n" +
	"\t\tfor _, f := range globalFlags {\n" +
	"\t\t\tb.WriteString(fmt.Sprintf(\"- \\x60--%s\\x60\", f.Name))\n" +
	"\n" +
	"\t\t\tif f.Shorthand != \"\" {\n" +
	"\t\t\t\tb.WriteString(fmt.Sprintf(\", \\x60-%s\\x60\", f.Shorthand))\n" +
	"\t\t\t}\n" +
	"\n" +
	"\t\t\tb.WriteString(fmt.Sprintf(\" - %s\", f.Description))\n" +
	"\n" +
	"\t\t\tif f.Default != \"\" && f.Default != \"false\" {\n" +
	"\t\t\t\tb.WriteString(fmt.Sprintf(\" (default: %s)\", f.Default))\n" +
	"\t\t\t}\n" +
	"\n" +
	"\t\t\tb.WriteString(\"\\n\")\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tb.WriteString(\"\\n\")\n" +
	"\t}\n" +
	"\n" +
	"\t// Commands\n" +
	"\tb.WriteString(\"## Commands\\n\\n\")\n" +
	"\taiWriteCommandMarkdown(&b, rootCmd.Commands(), \"\")\n" +
	"\n" +
	"\t// Categories — customize for your app\n" +
	"\t// TODO: customize for your app\n" +
	"\tcategories := map[string][]string{\n" +
	"\t\t\"Core\":    {\"version\", \"help\"},\n" +
	"\t\t\"Tooling\": {\"cmdtree\", \"aicontext\"},\n" +
	"\t}\n" +
	"\n" +
	"\tb.WriteString(\"## Command Categories\\n\\n\")\n" +
	"\tfor cat, cmds := range categories {\n" +
	"\t\t_, _ = fmt.Fprintf(&b, \"### %s\\n\\n\", cat)\n" +
	"\t\tfor _, name := range cmds {\n" +
	"\t\t\t_, _ = fmt.Fprintf(&b, \"- \\x60%s\\x60\\n\", name)\n" +
	"\t\t}\n" +
	"\t\tb.WriteString(\"\\n\")\n" +
	"\t}\n" +
	"\n" +
	"\t// Structure — customize for your app\n" +
	"\t// TODO: customize for your app\n" +
	"\tb.WriteString(\"## Project Structure\\n\\n\")\n" +
	"\tb.WriteString(\"\\x60\\x60\\x60\\n\")\n" +
	"\tstructure := []string{\n" +
	"\t\t\"cmd/           # CLI commands (Cobra wrappers)\",\n" +
	"\t\t\"internal/      # Private application code\",\n" +
	"\t\t\"pkg/           # Public reusable libraries\",\n" +
	"\t\t\"main.go        # Entry point\",\n" +
	"\t}\n" +
	"\tfor _, s := range structure {\n" +
	"\t\t_, _ = fmt.Fprintf(&b, \"%s\\n\", s)\n" +
	"\t}\n" +
	"\tb.WriteString(\"\\x60\\x60\\x60\\n\")\n" +
	"\n" +
	"\t_, _ = fmt.Fprint(cmd.OutOrStdout(), b.String())\n" +
	"\n" +
	"\treturn nil\n" +
	"}\n" +
	"\n" +
	"func printAIContextCompact(cmd *cobra.Command) error {\n" +
	"\tvar b strings.Builder\n" +
	"\n" +
	"\tb.WriteString(\"# {{.AppName}} - {{.Description}}\\n\\n\")\n" +
	"\n" +
	"\t// Global flags\n" +
	"\tvar globalParts []string\n" +
	"\trootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {\n" +
	"\t\tif f.Name == \"help\" {\n" +
	"\t\t\treturn\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tglobalParts = append(globalParts, fmt.Sprintf(\"\\x60--%s\\x60 %s\", f.Name, f.Usage))\n" +
	"\t})\n" +
	"\n" +
	"\tif len(globalParts) > 0 {\n" +
	"\t\tb.WriteString(\"**Global:** \")\n" +
	"\t\tb.WriteString(strings.Join(globalParts, \", \"))\n" +
	"\t\tb.WriteString(\"\\n\\n\")\n" +
	"\t}\n" +
	"\n" +
	"\t// Commands\n" +
	"\tfor _, c := range rootCmd.Commands() {\n" +
	"\t\tif c.Hidden {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\taiWriteCompactCommand(&b, c, \"\")\n" +
	"\t}\n" +
	"\n" +
	"\t_, _ = fmt.Fprint(cmd.OutOrStdout(), b.String())\n" +
	"\n" +
	"\treturn nil\n" +
	"}\n" +
	"\n" +
	"func printAIContextJSON(cmd *cobra.Command) error {\n" +
	"\tdoc := aiContextDoc{\n" +
	"\t\tTool:        \"{{.AppName}}\",\n" +
	"\t\tVersion:     \"dev\",\n" +
	"\t\tDescription: \"{{.Description}}\",\n" +
	"\t}\n" +
	"\n" +
	"\t// Global flags\n" +
	"\trootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {\n" +
	"\t\tif f.Name == \"help\" {\n" +
	"\t\t\treturn\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tdoc.GlobalFlags = append(doc.GlobalFlags, aiFlagInfo{\n" +
	"\t\t\tName:        f.Name,\n" +
	"\t\t\tShorthand:   f.Shorthand,\n" +
	"\t\t\tType:        f.Value.Type(),\n" +
	"\t\t\tDefault:     f.DefValue,\n" +
	"\t\t\tDescription: f.Usage,\n" +
	"\t\t\tGlobal:      true,\n" +
	"\t\t})\n" +
	"\t})\n" +
	"\n" +
	"\t// Commands\n" +
	"\tfor _, c := range rootCmd.Commands() {\n" +
	"\t\tif c.Hidden {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tdoc.Commands = append(doc.Commands, aiBuildCommandInfo(c))\n" +
	"\t}\n" +
	"\n" +
	"\tenc := json.NewEncoder(cmd.OutOrStdout())\n" +
	"\tenc.SetIndent(\"\", \"  \")\n" +
	"\n" +
	"\tif err := enc.Encode(doc); err != nil {\n" +
	"\t\treturn fmt.Errorf(\"json encode: %w\", err)\n" +
	"\t}\n" +
	"\n" +
	"\treturn nil\n" +
	"}\n" +
	"\n" +
	"func aiBuildCommandInfo(cmd *cobra.Command) aiCommandInfo {\n" +
	"\tinfo := aiCommandInfo{\n" +
	"\t\tName:        cmd.Name(),\n" +
	"\t\tUsage:       cmd.UseLine(),\n" +
	"\t\tDescription: cmd.Short,\n" +
	"\t}\n" +
	"\n" +
	"\tcmd.LocalFlags().VisitAll(func(f *pflag.Flag) {\n" +
	"\t\tif f.Name == \"help\" {\n" +
	"\t\t\treturn\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tinfo.Flags = append(info.Flags, aiFlagInfo{\n" +
	"\t\t\tName:        f.Name,\n" +
	"\t\t\tShorthand:   f.Shorthand,\n" +
	"\t\t\tType:        f.Value.Type(),\n" +
	"\t\t\tDefault:     f.DefValue,\n" +
	"\t\t\tDescription: f.Usage,\n" +
	"\t\t})\n" +
	"\t})\n" +
	"\n" +
	"\tfor _, sub := range cmd.Commands() {\n" +
	"\t\tif sub.Hidden {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tinfo.Subcommands = append(info.Subcommands, aiBuildCommandInfo(sub))\n" +
	"\t}\n" +
	"\n" +
	"\treturn info\n" +
	"}\n" +
	"\n" +
	"func aiWriteCommandMarkdown(b *strings.Builder, commands []*cobra.Command, prefix string) {\n" +
	"\tfor _, c := range commands {\n" +
	"\t\tif c.Hidden {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\theading := \"###\"\n" +
	"\t\tif prefix != \"\" {\n" +
	"\t\t\theading = \"####\"\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tname := c.Name()\n" +
	"\t\tif prefix != \"\" {\n" +
	"\t\t\tname = prefix + \" \" + name\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t_, _ = fmt.Fprintf(b, \"%s %s\\n\\n\", heading, name)\n" +
	"\t\t_, _ = fmt.Fprintf(b, \"Usage: \\x60%s\\x60\\n\\n\", c.UseLine())\n" +
	"\t\t_, _ = fmt.Fprintf(b, \"%s\\n\\n\", c.Short)\n" +
	"\n" +
	"\t\t// Flags\n" +
	"\t\thasFlags := false\n" +
	"\n" +
	"\t\tc.LocalFlags().VisitAll(func(f *pflag.Flag) {\n" +
	"\t\t\tif f.Name == \"help\" {\n" +
	"\t\t\t\treturn\n" +
	"\t\t\t}\n" +
	"\n" +
	"\t\t\tif !hasFlags {\n" +
	"\t\t\t\tb.WriteString(\"Flags:\\n\")\n" +
	"\n" +
	"\t\t\t\thasFlags = true\n" +
	"\t\t\t}\n" +
	"\n" +
	"\t\t\t_, _ = fmt.Fprintf(b, \"- \\x60--%s\\x60\", f.Name)\n" +
	"\n" +
	"\t\t\tif f.Shorthand != \"\" {\n" +
	"\t\t\t\t_, _ = fmt.Fprintf(b, \", \\x60-%s\\x60\", f.Shorthand)\n" +
	"\t\t\t}\n" +
	"\n" +
	"\t\t\t_, _ = fmt.Fprintf(b, \" - %s\", f.Usage)\n" +
	"\n" +
	"\t\t\tif f.DefValue != \"\" && f.DefValue != \"false\" && f.DefValue != \"0\" {\n" +
	"\t\t\t\t_, _ = fmt.Fprintf(b, \" (default: %s)\", f.DefValue)\n" +
	"\t\t\t}\n" +
	"\n" +
	"\t\t\tb.WriteString(\"\\n\")\n" +
	"\t\t})\n" +
	"\n" +
	"\t\tif hasFlags {\n" +
	"\t\t\tb.WriteString(\"\\n\")\n" +
	"\t\t}\n" +
	"\n" +
	"\t\t// Subcommands\n" +
	"\t\tif len(c.Commands()) > 0 {\n" +
	"\t\t\taiWriteCommandMarkdown(b, c.Commands(), c.Name())\n" +
	"\t\t}\n" +
	"\t}\n" +
	"}\n" +
	"\n" +
	"func aiWriteCompactCommand(b *strings.Builder, cmd *cobra.Command, prefix string) {\n" +
	"\tname := cmd.Name()\n" +
	"\tif prefix != \"\" {\n" +
	"\t\tname = prefix + \" \" + name\n" +
	"\t}\n" +
	"\n" +
	"\t_, _ = fmt.Fprintf(b, \"- \\x60{{.AppName}} %s\\x60 - %s\", name, cmd.Short)\n" +
	"\n" +
	"\t// Inline flags\n" +
	"\tvar flagParts []string\n" +
	"\n" +
	"\tcmd.LocalFlags().VisitAll(func(f *pflag.Flag) {\n" +
	"\t\tif f.Name == \"help\" {\n" +
	"\t\t\treturn\n" +
	"\t\t}\n" +
	"\n" +
	"\t\tflagParts = append(flagParts, fmt.Sprintf(\"\\x60--%s\\x60\", f.Name))\n" +
	"\t})\n" +
	"\n" +
	"\tif len(flagParts) > 0 {\n" +
	"\t\t_, _ = fmt.Fprintf(b, \" [%s]\", strings.Join(flagParts, \", \"))\n" +
	"\t}\n" +
	"\n" +
	"\tb.WriteString(\"\\n\")\n" +
	"\n" +
	"\tfor _, sub := range cmd.Commands() {\n" +
	"\t\tif sub.Hidden {\n" +
	"\t\t\tcontinue\n" +
	"\t\t}\n" +
	"\n" +
	"\t\taiWriteCompactCommand(b, sub, name)\n" +
	"\t}\n" +
	"}\n"
