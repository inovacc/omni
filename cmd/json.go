package cmd

import (
	"encoding/json"
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/inovacc/omni/internal/cli/csvutil"
	"github.com/inovacc/omni/internal/cli/json2struct"
	"github.com/inovacc/omni/internal/cli/jsonfmt"
	"github.com/inovacc/omni/internal/cli/xmlutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
  toyaml    Convert JSON to YAML
  fromyaml  Convert YAML to JSON
  fromtoml  Convert TOML to JSON
  tostruct  Convert JSON to Go struct definition
  tocsv     Convert JSON array to CSV
  fromcsv   Convert CSV to JSON array
  toxml     Convert JSON to XML
  fromxml   Convert XML to JSON

Examples:
  omni json fmt file.json              # beautify JSON
  omni json minify file.json           # compact JSON
  omni json validate file.json         # check if valid
  echo '{"a":1}' | omni json fmt       # from stdin
  omni json stats file.json            # show statistics
  omni json toyaml file.json           # convert to YAML
  omni json fromyaml file.yaml         # convert from YAML
  omni json fromtoml file.toml         # convert from TOML
  omni json tostruct file.json         # convert to Go struct
  omni json tocsv file.json            # convert to CSV
  omni json fromcsv file.csv           # convert from CSV
  omni json toxml file.json            # convert to XML
  omni json fromxml file.xml           # convert from XML`,
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

// jsonToYAMLCmd converts JSON to YAML
var jsonToYAMLCmd = &cobra.Command{
	Use:     "toyaml [FILE]",
	Aliases: []string{"yaml", "to-yaml"},
	Short:   "Convert JSON to YAML",
	Long: `Convert JSON data to YAML format.

Examples:
  omni json toyaml file.json
  echo '{"name":"test"}' | omni json toyaml
  omni json toyaml file.json > output.yaml`,
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

		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}

		output, err := yaml.Marshal(v)
		if err != nil {
			return err
		}

		_, _ = os.Stdout.Write(output)
		return nil
	},
}

// jsonFromYAMLCmd converts YAML to JSON
var jsonFromYAMLCmd = &cobra.Command{
	Use:     "fromyaml [FILE]",
	Aliases: []string{"from-yaml", "y2j"},
	Short:   "Convert YAML to JSON",
	Long: `Convert YAML data to JSON format.

  -m, --minify    output minified JSON (no indentation)

Examples:
  omni json fromyaml file.yaml
  cat file.yaml | omni json fromyaml
  omni json fromyaml -m file.yaml     # minified output
  omni json fromyaml file.yaml > output.json`,
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

		var v any
		if err := yaml.Unmarshal(data, &v); err != nil {
			return err
		}

		minify, _ := cmd.Flags().GetBool("minify")

		var output []byte
		if minify {
			output, err = json.Marshal(v)
		} else {
			output, err = json.MarshalIndent(v, "", "  ")
		}

		if err != nil {
			return err
		}

		_, _ = os.Stdout.Write(output)
		_, _ = os.Stdout.WriteString("\n")
		return nil
	},
}

// jsonFromTOMLCmd converts TOML to JSON
var jsonFromTOMLCmd = &cobra.Command{
	Use:     "fromtoml [FILE]",
	Aliases: []string{"from-toml", "t2j"},
	Short:   "Convert TOML to JSON",
	Long: `Convert TOML data to JSON format.

  -m, --minify    output minified JSON (no indentation)

Examples:
  omni json fromtoml file.toml
  cat file.toml | omni json fromtoml
  omni json fromtoml -m file.toml     # minified output
  omni json fromtoml file.toml > output.json`,
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

		var v any
		if err := toml.Unmarshal(data, &v); err != nil {
			return err
		}

		minify, _ := cmd.Flags().GetBool("minify")

		var output []byte
		if minify {
			output, err = json.Marshal(v)
		} else {
			output, err = json.MarshalIndent(v, "", "  ")
		}

		if err != nil {
			return err
		}

		_, _ = os.Stdout.Write(output)
		_, _ = os.Stdout.WriteString("\n")
		return nil
	},
}

// jsonToStructCmd converts JSON to Go struct
var jsonToStructCmd = &cobra.Command{
	Use:     "tostruct [FILE]",
	Aliases: []string{"2struct", "gostruct"},
	Short:   "Convert JSON to Go struct definition",
	Long: `Convert JSON data to a Go struct definition.

  -n, --name=NAME      struct name (default "Root")
  -p, --package=PKG    package name (default "main")
  --inline             inline nested structs
  --omitempty          add omitempty to all fields

Examples:
  omni json tostruct file.json
  echo '{"name":"test","count":1}' | omni json tostruct
  omni json tostruct -n User -p models file.json
  omni json tostruct --omitempty file.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := json2struct.Options{}

		opts.Name, _ = cmd.Flags().GetString("name")
		opts.Package, _ = cmd.Flags().GetString("package")
		opts.Inline, _ = cmd.Flags().GetBool("inline")
		opts.OmitEmpty, _ = cmd.Flags().GetBool("omitempty")

		return json2struct.RunJSON2Struct(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

// jsonToCSVCmd converts JSON to CSV
var jsonToCSVCmd = &cobra.Command{
	Use:     "tocsv [FILE]",
	Aliases: []string{"csv", "2csv"},
	Short:   "Convert JSON array to CSV",
	Long: `Convert JSON array of objects to CSV format.

Nested objects are flattened with dot notation (e.g., address.city).

  --no-header          don't include header row
  -d, --delimiter=STR  field delimiter (default ",")
  --no-quotes          don't quote fields

Examples:
  omni json tocsv file.json
  echo '[{"name":"John","age":30}]' | omni json tocsv
  omni json tocsv -d ";" file.json     # semicolon delimiter
  omni json tocsv --no-header file.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := csvutil.ToCSVOptions{Header: true}

		noHeader, _ := cmd.Flags().GetBool("no-header")
		opts.Header = !noHeader
		opts.Delimiter, _ = cmd.Flags().GetString("delimiter")
		opts.NoQuotes, _ = cmd.Flags().GetBool("no-quotes")

		return csvutil.RunToCSV(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

// jsonFromCSVCmd converts CSV to JSON
var jsonFromCSVCmd = &cobra.Command{
	Use:     "fromcsv [FILE]",
	Aliases: []string{"from-csv", "csv2json"},
	Short:   "Convert CSV to JSON array",
	Long: `Convert CSV data to JSON array of objects.

  --no-header          first row is data, not headers
  -d, --delimiter=STR  field delimiter (default ",")
  -a, --array          always output as array (even for single row)

Examples:
  omni json fromcsv file.csv
  cat file.csv | omni json fromcsv
  omni json fromcsv -d ";" file.csv    # semicolon delimiter
  omni json fromcsv --no-header file.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := csvutil.FromCSVOptions{Header: true}

		noHeader, _ := cmd.Flags().GetBool("no-header")
		opts.Header = !noHeader
		opts.Delimiter, _ = cmd.Flags().GetString("delimiter")
		opts.Array, _ = cmd.Flags().GetBool("array")

		return csvutil.RunFromCSV(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

// jsonToXMLCmd converts JSON to XML
var jsonToXMLCmd = &cobra.Command{
	Use:     "toxml [FILE]",
	Aliases: []string{"xml", "2xml"},
	Short:   "Convert JSON to XML",
	Long: `Convert JSON data to XML format.

  -r, --root=NAME      root element name (default "root")
  -i, --indent=STR     indentation string (default "  ")
  --item-tag=NAME      tag for array items (default "item")
  --attr-prefix=STR    prefix for attributes (default "-")

Examples:
  omni json toxml file.json
  echo '{"name":"John"}' | omni json toxml
  omni json toxml -r person file.json   # custom root
  omni json toxml --item-tag=entry file.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xmlutil.ToXMLOptions{}

		opts.Root, _ = cmd.Flags().GetString("root")
		opts.Indent, _ = cmd.Flags().GetString("indent")
		opts.ItemTag, _ = cmd.Flags().GetString("item-tag")
		opts.AttrPrefix, _ = cmd.Flags().GetString("attr-prefix")

		return xmlutil.RunToXML(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

// jsonFromXMLCmd converts XML to JSON
var jsonFromXMLCmd = &cobra.Command{
	Use:     "fromxml [FILE]",
	Aliases: []string{"from-xml", "xml2json"},
	Short:   "Convert XML to JSON",
	Long: `Convert XML data to JSON format.

  --attr-prefix=STR    prefix for attributes in JSON (default "-")
  --text-key=STR       key for text content (default "#text")

Examples:
  omni json fromxml file.xml
  cat file.xml | omni json fromxml
  omni json fromxml --attr-prefix=@ file.xml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xmlutil.FromXMLOptions{}

		opts.AttrPrefix, _ = cmd.Flags().GetString("attr-prefix")
		opts.TextKey, _ = cmd.Flags().GetString("text-key")

		return xmlutil.RunFromXML(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func readStdin() ([]byte, error) {
	return io.ReadAll(os.Stdin)
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
	jsonCmd.AddCommand(jsonToYAMLCmd)
	jsonCmd.AddCommand(jsonFromYAMLCmd)
	jsonCmd.AddCommand(jsonFromTOMLCmd)
	jsonCmd.AddCommand(jsonToStructCmd)
	jsonCmd.AddCommand(jsonToCSVCmd)
	jsonCmd.AddCommand(jsonFromCSVCmd)
	jsonCmd.AddCommand(jsonToXMLCmd)
	jsonCmd.AddCommand(jsonFromXMLCmd)

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

	// fromyaml flags
	jsonFromYAMLCmd.Flags().BoolP("minify", "m", false, "output minified JSON")

	// fromtoml flags
	jsonFromTOMLCmd.Flags().BoolP("minify", "m", false, "output minified JSON")

	// tostruct flags
	jsonToStructCmd.Flags().StringP("name", "n", "Root", "struct name")
	jsonToStructCmd.Flags().StringP("package", "p", "main", "package name")
	jsonToStructCmd.Flags().Bool("inline", false, "inline nested structs")
	jsonToStructCmd.Flags().Bool("omitempty", false, "add omitempty to all fields")

	// tocsv flags
	jsonToCSVCmd.Flags().Bool("no-header", false, "don't include header row")
	jsonToCSVCmd.Flags().StringP("delimiter", "d", ",", "field delimiter")
	jsonToCSVCmd.Flags().Bool("no-quotes", false, "don't quote fields")

	// fromcsv flags
	jsonFromCSVCmd.Flags().Bool("no-header", false, "first row is data, not headers")
	jsonFromCSVCmd.Flags().StringP("delimiter", "d", ",", "field delimiter")
	jsonFromCSVCmd.Flags().BoolP("array", "a", false, "always output as array")

	// toxml flags
	jsonToXMLCmd.Flags().StringP("root", "r", "root", "root element name")
	jsonToXMLCmd.Flags().StringP("indent", "i", "  ", "indentation string")
	jsonToXMLCmd.Flags().String("item-tag", "item", "tag for array items")
	jsonToXMLCmd.Flags().String("attr-prefix", "-", "prefix for attributes")

	// fromxml flags
	jsonFromXMLCmd.Flags().String("attr-prefix", "-", "prefix for attributes in JSON")
	jsonFromXMLCmd.Flags().String("text-key", "#text", "key for text content")
}
