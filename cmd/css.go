package cmd

import (
	"github.com/inovacc/omni/internal/cli/cssfmt"
	"github.com/spf13/cobra"
)

var cssCmd = &cobra.Command{
	Use:   "css [FILE]",
	Short: "CSS utilities (format, minify, validate)",
	Long: `CSS utilities for formatting, minifying, and validating CSS.

When called directly, formats CSS (same as 'css fmt').

Subcommands:
  fmt         Format/beautify CSS
  minify      Minify CSS
  validate    Validate CSS syntax

Examples:
  omni css file.css
  omni css fmt file.css
  omni css minify file.css
  omni css validate file.css
  echo 'body{margin:0}' | omni css
  omni css "body{margin:0;padding:0}"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cssfmt.Options{}
		opts.Indent, _ = cmd.Flags().GetString("indent")
		opts.SortProps, _ = cmd.Flags().GetBool("sort-props")
		opts.SortRules, _ = cmd.Flags().GetBool("sort-rules")

		return cssfmt.Run(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

var cssFmtCmd = &cobra.Command{
	Use:     "fmt [FILE]",
	Aliases: []string{"format", "beautify"},
	Short:   "Format/beautify CSS",
	Long: `Format CSS with proper indentation.

  -i, --indent=STR     indentation string (default "  ")
  --sort-props         sort properties alphabetically
  --sort-rules         sort selectors alphabetically

Examples:
  omni css fmt file.css
  omni css fmt "body{margin:0;padding:0}"
  cat file.css | omni css fmt
  omni css fmt --sort-props file.css`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cssfmt.Options{}
		opts.Indent, _ = cmd.Flags().GetString("indent")
		opts.SortProps, _ = cmd.Flags().GetBool("sort-props")
		opts.SortRules, _ = cmd.Flags().GetBool("sort-rules")

		return cssfmt.Run(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

var cssMinifyCmd = &cobra.Command{
	Use:     "minify [FILE]",
	Aliases: []string{"min", "compact"},
	Short:   "Minify CSS",
	Long: `Minify CSS by removing unnecessary whitespace and comments.

Examples:
  omni css minify file.css
  cat file.css | omni css minify`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cssfmt.RunMinify(cmd.OutOrStdout(), cmd.InOrStdin(), args, cssfmt.Options{})
	},
}

var cssValidateCmd = &cobra.Command{
	Use:     "validate [FILE]",
	Aliases: []string{"check", "lint"},
	Short:   "Validate CSS syntax",
	Long: `Validate CSS syntax.

Exit codes:
  0  Valid CSS
  1  Invalid CSS or error

  --json    output result as JSON

Examples:
  omni css validate file.css
  omni css validate "body { margin: 0; }"
  omni css validate --json file.css`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cssfmt.ValidateOptions{}
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return cssfmt.RunValidate(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(cssCmd)
	cssCmd.AddCommand(cssFmtCmd)
	cssCmd.AddCommand(cssMinifyCmd)
	cssCmd.AddCommand(cssValidateCmd)

	// css root flags
	cssCmd.Flags().StringP("indent", "i", "  ", "indentation string")
	cssCmd.Flags().Bool("sort-props", false, "sort properties alphabetically")
	cssCmd.Flags().Bool("sort-rules", false, "sort selectors alphabetically")

	// css fmt flags
	cssFmtCmd.Flags().StringP("indent", "i", "  ", "indentation string")
	cssFmtCmd.Flags().Bool("sort-props", false, "sort properties alphabetically")
	cssFmtCmd.Flags().Bool("sort-rules", false, "sort selectors alphabetically")

	// css validate flags
	cssValidateCmd.Flags().Bool("json", false, "output as JSON")
}
