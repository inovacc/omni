package cmd

import (
	"github.com/inovacc/omni/internal/cli/csvutil"
	"github.com/spf13/cobra"
)

var csvCmd = &cobra.Command{
	Use:   "csv",
	Short: "CSV utilities (convert to/from JSON)",
	Long: `CSV utilities for converting between CSV and JSON formats.

Subcommands:
  tojson    Convert CSV to JSON array
  fromjson  Convert JSON array to CSV

Examples:
  omni csv tojson file.csv             # convert CSV to JSON
  omni csv fromjson file.json          # convert JSON to CSV
  cat data.csv | omni csv tojson       # from stdin
  omni csv tojson -d ";" file.csv      # custom delimiter`,
}

var csvToJSONCmd = &cobra.Command{
	Use:     "tojson [FILE]",
	Aliases: []string{"json", "2json"},
	Short:   "Convert CSV to JSON array",
	Long: `Convert CSV data to JSON array of objects.

  --no-header          first row is data, not headers
  -d, --delimiter=STR  field delimiter (default ",")
  -a, --array          always output as array (even for single row)

Examples:
  omni csv tojson file.csv
  cat file.csv | omni csv tojson
  omni csv tojson -d ";" file.csv      # semicolon delimiter
  omni csv tojson --no-header file.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := csvutil.FromCSVOptions{Header: true}

		noHeader, _ := cmd.Flags().GetBool("no-header")
		opts.Header = !noHeader
		opts.Delimiter, _ = cmd.Flags().GetString("delimiter")
		opts.Array, _ = cmd.Flags().GetBool("array")

		return csvutil.RunFromCSV(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

var csvFromJSONCmd = &cobra.Command{
	Use:     "fromjson [FILE]",
	Aliases: []string{"from-json", "json2csv"},
	Short:   "Convert JSON array to CSV",
	Long: `Convert JSON array of objects to CSV format.

Nested objects are flattened with dot notation (e.g., address.city).

  --no-header          don't include header row
  -d, --delimiter=STR  field delimiter (default ",")
  --no-quotes          don't quote fields

Examples:
  omni csv fromjson file.json
  echo '[{"name":"John","age":30}]' | omni csv fromjson
  omni csv fromjson -d ";" file.json   # semicolon delimiter
  omni csv fromjson --no-header file.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := csvutil.ToCSVOptions{Header: true}

		noHeader, _ := cmd.Flags().GetBool("no-header")
		opts.Header = !noHeader
		opts.Delimiter, _ = cmd.Flags().GetString("delimiter")
		opts.NoQuotes, _ = cmd.Flags().GetBool("no-quotes")

		return csvutil.RunToCSV(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(csvCmd)
	csvCmd.AddCommand(csvToJSONCmd)
	csvCmd.AddCommand(csvFromJSONCmd)

	// tojson flags
	csvToJSONCmd.Flags().Bool("no-header", false, "first row is data, not headers")
	csvToJSONCmd.Flags().StringP("delimiter", "d", ",", "field delimiter")
	csvToJSONCmd.Flags().BoolP("array", "a", false, "always output as array")

	// fromjson flags
	csvFromJSONCmd.Flags().Bool("no-header", false, "don't include header row")
	csvFromJSONCmd.Flags().StringP("delimiter", "d", ",", "field delimiter")
	csvFromJSONCmd.Flags().Bool("no-quotes", false, "don't quote fields")
}
