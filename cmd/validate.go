package cmd

import (
	"github.com/inovacc/omni/internal/cli/validate"
	"github.com/spf13/cobra"
)

// validateCmd is the parent command for data-format validators.
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate data formats",
	Long: `Validate that a value conforms to a well-known data format.

A malformed value exits 1 (validation failure); a missing argument exits 2.

Subcommands:
  email  Validate an email address
  ip     Validate an IPv4/IPv6 address

Examples:
  omni validate email user@example.com   # OK, exit 0
  omni validate ip 192.168.0.1           # OK, exit 0
  omni validate ip 999.1.1.1             # invalid, exit 1`,
}

// validateEmailCmd validates an email address.
var validateEmailCmd = &cobra.Command{
	Use:   "email ADDRESS",
	Short: "Validate an email address",
	Long: `Validate ADDRESS as an RFC 5322 email address.

A malformed address exits 1; a valid address prints "<addr>	OK".

Examples:
  omni validate email user@example.com
  omni validate email nope               # invalid, exit 1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return validate.RunEmail(cmd.OutOrStdout(), args[0])
	},
}

// validateIPCmd validates an IP address.
var validateIPCmd = &cobra.Command{
	Use:   "ip ADDRESS",
	Short: "Validate an IPv4/IPv6 address",
	Long: `Validate ADDRESS as an IPv4 or IPv6 address.

A malformed address exits 1; a valid address prints "<addr>	OK".

Examples:
  omni validate ip 192.168.0.1
  omni validate ip ::1
  omni validate ip 999.1.1.1             # invalid, exit 1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return validate.RunIP(cmd.OutOrStdout(), args[0])
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.AddCommand(validateEmailCmd, validateIPCmd)
}
