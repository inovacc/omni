package cmd

import (
	"github.com/inovacc/omni/internal/cli/sqlfmt"
	"github.com/spf13/cobra"
)

var sqlCmd = &cobra.Command{
	Use:   "sql [FILE]",
	Short: "SQL utilities (format, minify, validate)",
	Long: `SQL utilities for formatting, minifying, and validating SQL.

When called directly, formats SQL (same as 'sql fmt').

Subcommands:
  fmt         Format/beautify SQL
  minify      Compact SQL
  validate    Validate SQL syntax

Examples:
  omni sql file.sql
  omni sql fmt file.sql
  omni sql minify file.sql
  omni sql validate file.sql
  echo 'select * from users' | omni sql
  omni sql "SELECT * FROM users WHERE id=1"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := sqlfmt.Options{}
		opts.Indent, _ = cmd.Flags().GetString("indent")
		opts.Uppercase, _ = cmd.Flags().GetBool("uppercase")

		return sqlfmt.Run(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

var sqlFmtCmd = &cobra.Command{
	Use:     "fmt [FILE]",
	Aliases: []string{"format", "beautify"},
	Short:   "Format/beautify SQL",
	Long: `Format SQL with proper indentation and keyword capitalization.

  -i, --indent=STR     indentation string (default "  ")
  -u, --uppercase      uppercase keywords (default: true)
  -d, --dialect=NAME   SQL dialect: mysql, postgres, sqlite (default: generic)

Examples:
  omni sql fmt file.sql
  omni sql fmt "select * from users where id = 1"
  cat file.sql | omni sql fmt
  omni sql fmt --indent "    " file.sql`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := sqlfmt.Options{}
		opts.Indent, _ = cmd.Flags().GetString("indent")
		opts.Uppercase, _ = cmd.Flags().GetBool("uppercase")
		opts.Dialect, _ = cmd.Flags().GetString("dialect")

		return sqlfmt.Run(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

var sqlMinifyCmd = &cobra.Command{
	Use:     "minify [FILE]",
	Aliases: []string{"min", "compact"},
	Short:   "Minify SQL",
	Long: `Minify SQL by removing unnecessary whitespace.

Examples:
  omni sql minify file.sql
  cat file.sql | omni sql minify`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := sqlfmt.Options{Minify: true}

		return sqlfmt.RunMinify(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

var sqlValidateCmd = &cobra.Command{
	Use:     "validate [FILE]",
	Aliases: []string{"check", "lint"},
	Short:   "Validate SQL syntax",
	Long: `Validate SQL syntax without outputting the query.

Exit codes:
  0  Valid SQL
  1  Invalid SQL or error

  --json    output result as JSON

Examples:
  omni sql validate file.sql
  omni sql validate "SELECT * FROM users"
  omni sql validate --json file.sql`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := sqlfmt.ValidateOptions{}
		opts.OutputFormat = getOutputOpts(cmd).GetFormat()
		opts.Dialect, _ = cmd.Flags().GetString("dialect")

		return sqlfmt.RunValidate(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(sqlCmd)
	sqlCmd.AddCommand(sqlFmtCmd)
	sqlCmd.AddCommand(sqlMinifyCmd)
	sqlCmd.AddCommand(sqlValidateCmd)

	// sql root flags
	sqlCmd.Flags().StringP("indent", "i", "  ", "indentation string")
	sqlCmd.Flags().BoolP("uppercase", "u", true, "uppercase keywords")

	// sql fmt flags
	sqlFmtCmd.Flags().StringP("indent", "i", "  ", "indentation string")
	sqlFmtCmd.Flags().BoolP("uppercase", "u", true, "uppercase keywords")
	sqlFmtCmd.Flags().StringP("dialect", "d", "generic", "SQL dialect (mysql, postgres, sqlite, generic)")

	// sql validate flags
	sqlValidateCmd.Flags().StringP("dialect", "d", "generic", "SQL dialect")
}
