package cmd

import (
	"github.com/inovacc/omni/internal/cli/printf"
	"github.com/spf13/cobra"
)

var printfCmd = &cobra.Command{
	Use:   "printf FORMAT [ARG...]",
	Short: "Format and print data",
	Long: `Format and print data using printf-style format specifiers.

Format specifiers:
  %s    String
  %d    Decimal integer
  %i    Integer (same as %d)
  %o    Octal
  %x    Lowercase hexadecimal
  %X    Uppercase hexadecimal
  %b    Binary
  %f    Floating point
  %e    Scientific notation
  %g    Compact floating point
  %c    Character
  %q    Quoted string
  %%    Literal percent sign

Escape sequences:
  \n    Newline
  \t    Tab
  \r    Carriage return
  \\    Backslash
  \xHH  Hex character
  \NNN  Octal character

Width and precision:
  %10s   Right-aligned, width 10
  %-10s  Left-aligned, width 10
  %.5s   Max 5 characters
  %10.5s Width 10, max 5 characters
  %08d   Zero-padded, width 8

Examples:
  omni printf "Hello, %s!" World
  omni printf "Number: %d, Hex: %x" 255 255
  omni printf "Pi: %.2f" 3.14159
  omni printf "Name: %-10s Age: %3d" Alice 25
  omni printf "Tab:\tNewline:\n"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noNewline, _ := cmd.Flags().GetBool("no-newline")

		opts := printf.Options{
			NoNewline: noNewline,
		}

		return printf.Run(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(printfCmd)

	printfCmd.Flags().BoolP("no-newline", "n", false, "do not append a trailing newline")
}
