package cmd

import (
	"encoding/json"
	"os"

	"github.com/inovacc/omni/internal/cli/jsonfmt"
	"github.com/spf13/cobra"
)

// jsonCmd represents the json command
var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "JSON utilities (format, minify, validate)",
	Long: `JSON utilities for formatting, minifying, and validating JSON data.

Subcommands:
  fmt       Beautify/format JSON with indentation
  minify    Compact JSON by removing whitespace
  validate  Check if input is valid JSON
  stats     Show statistics about JSON data
  keys      List all keys in JSON object

Examples:
  omni json fmt file.json              # beautify JSON
  omni json minify file.json           # compact JSON
  omni json validate file.json         # check if valid
  echo '{"a":1}' | omni json fmt       # from stdin
  omni json stats file.json            # show statistics`,
}

// jsonFmtCmd formats JSON
var jsonFmtCmd = &cobra.Command{
	Use:     "fmt [FILE]...",
	Aliases: []string{"format", "beautify", "pretty"},
	Short:   "Beautify/format JSON with indentation",
	Long: `Format JSON with proper indentation and line breaks.

  -i, --indent=STR   indentation string (default "  ")
  -t, --tab          use tabs for indentation
  -s, --sort-keys    sort object keys alphabetically
  -e, --escape-html  escape HTML characters (<, >, &)

Examples:
  omni json fmt file.json              # beautify with 2-space indent
  omni json fmt -t file.json           # use tabs
  omni json fmt -s file.json           # sort keys
  echo '{"b":2,"a":1}' | omni json fmt -s`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := jsonfmt.Options{}

		opts.Indent, _ = cmd.Flags().GetString("indent")
		opts.Tab, _ = cmd.Flags().GetBool("tab")
		opts.SortKeys, _ = cmd.Flags().GetBool("sort-keys")
		opts.EscapeHTML, _ = cmd.Flags().GetBool("escape-html")

		return jsonfmt.RunJSONFmt(os.Stdout, args, opts)
	},
}

// jsonMinifyCmd minifies JSON
var jsonMinifyCmd = &cobra.Command{
	Use:     "minify [FILE]...",
	Aliases: []string{"min", "compact"},
	Short:   "Compact JSON by removing whitespace",
	Long: `Remove all unnecessary whitespace from JSON.

Examples:
  omni json minify file.json
  cat file.json | omni json minify
  omni json minify -s file.json        # also sort keys`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := jsonfmt.Options{Minify: true}

		opts.SortKeys, _ = cmd.Flags().GetBool("sort-keys")

		return jsonfmt.RunJSONFmt(os.Stdout, args, opts)
	},
}

// jsonValidateCmd validates JSON
var jsonValidateCmd = &cobra.Command{
	Use:     "validate [FILE]...",
	Aliases: []string{"check", "lint"},
	Short:   "Check if input is valid JSON",
	Long: `Validate JSON syntax without outputting the data.

Exit codes:
  0  Valid JSON
  1  Invalid JSON or error

  --json    output result as JSON

Examples:
  omni json validate file.json
  omni json validate --json file.json
  echo '{"valid": true}' | omni json validate`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := jsonfmt.Options{Validate: true}

		opts.JSON, _ = cmd.Flags().GetBool("json")

		return jsonfmt.RunJSONFmt(os.Stdout, args, opts)
	},
}

// jsonStatsCmd shows JSON statistics
var jsonStatsCmd = &cobra.Command{
	Use:   "stats [FILE]",
	Short: "Show statistics about JSON data",
	Long: `Display statistics about JSON data including type, depth, size, etc.

Examples:
  omni json stats file.json
  echo '[1,2,3]' | omni json stats`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var data []byte
		var err error
		var filename string

		if len(args) == 0 || args[0] == "-" {
			data, err = os.ReadFile("/dev/stdin")
			filename = "<stdin>"
		} else {
			data, err = os.ReadFile(args[0])
			filename = args[0]
		}

		if err != nil {
			// Try reading from stdin on Windows
			if len(args) == 0 {
				data, err = readStdin()
				if err != nil {
					return err
				}
				filename = "<stdin>"
			} else {
				return err
			}
		}

		stats, err := jsonfmt.GetStats(data)
		if err != nil {
			return err
		}

		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(stats)
		}

		_, _ = os.Stdout.WriteString("File: " + filename + "\n")
		_, _ = os.Stdout.WriteString("Type: " + stats.Type + "\n")
		if stats.Keys > 0 {
			_, _ = os.Stdout.WriteString("Keys: " + itoa(stats.Keys) + "\n")
		}
		if stats.Elements > 0 {
			_, _ = os.Stdout.WriteString("Elements: " + itoa(stats.Elements) + "\n")
		}
		_, _ = os.Stdout.WriteString("Depth: " + itoa(stats.Depth) + "\n")
		_, _ = os.Stdout.WriteString("Size: " + itoa(stats.Size) + " bytes\n")
		_, _ = os.Stdout.WriteString("Minified: " + itoa(stats.MinifiedLen) + " bytes\n")

		return nil
	},
}

// jsonKeysCmd lists JSON keys
var jsonKeysCmd = &cobra.Command{
	Use:   "keys [FILE]",
	Short: "List all keys in JSON object",
	Long: `List all keys (paths) in a JSON object recursively.

Examples:
  omni json keys file.json
  echo '{"a":{"b":1}}' | omni json keys`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var data []byte
		var err error

		if len(args) == 0 || args[0] == "-" {
			data, err = readStdin()
		} else {
			data, err = os.ReadFile(args[0])
		}

		if err != nil {
			return err
		}

		keys, err := jsonfmt.Keys(data)
		if err != nil {
			return err
		}

		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(keys)
		}

		for _, key := range keys {
			_, _ = os.Stdout.WriteString(key + "\n")
		}

		return nil
	},
}

func readStdin() ([]byte, error) {
	return os.ReadFile(os.Stdin.Name())
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	return string(digits)
}

func init() {
	rootCmd.AddCommand(jsonCmd)

	// Add subcommands
	jsonCmd.AddCommand(jsonFmtCmd)
	jsonCmd.AddCommand(jsonMinifyCmd)
	jsonCmd.AddCommand(jsonValidateCmd)
	jsonCmd.AddCommand(jsonStatsCmd)
	jsonCmd.AddCommand(jsonKeysCmd)

	// fmt flags
	jsonFmtCmd.Flags().StringP("indent", "i", "  ", "indentation string")
	jsonFmtCmd.Flags().BoolP("tab", "t", false, "use tabs for indentation")
	jsonFmtCmd.Flags().BoolP("sort-keys", "s", false, "sort object keys")
	jsonFmtCmd.Flags().BoolP("escape-html", "e", false, "escape HTML characters")

	// minify flags
	jsonMinifyCmd.Flags().BoolP("sort-keys", "s", false, "sort object keys")

	// validate flags
	jsonValidateCmd.Flags().Bool("json", false, "output result as JSON")

	// stats flags
	jsonStatsCmd.Flags().Bool("json", false, "output as JSON")

	// keys flags
	jsonKeysCmd.Flags().Bool("json", false, "output as JSON")
}
