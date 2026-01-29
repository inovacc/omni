package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/jwt"
	"github.com/spf13/cobra"
)

var jwtCmd = &cobra.Command{
	Use:   "jwt",
	Short: "JWT (JSON Web Token) utilities",
	Long: `JWT (JSON Web Token) utilities.

Subcommands:
  decode    Decode and inspect a JWT token

Examples:
  omni jwt decode "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  echo $TOKEN | omni jwt decode
  omni jwt decode --header token.txt`,
}

var jwtDecodeCmd = &cobra.Command{
	Use:   "decode [TOKEN]",
	Short: "Decode and inspect a JWT token",
	Long: `Decode and inspect a JWT token.

Displays the header and payload of a JWT token. Does NOT verify the signature
(use a proper JWT library for that). Useful for debugging and inspection.

Examples:
  omni jwt decode "eyJhbGciOiJIUzI1NiIs..."
  omni jwt decode --header "eyJhbGciOiJIUzI1NiIs..."
  omni jwt decode --payload "eyJhbGciOiJIUzI1NiIs..."
  omni jwt decode --json "eyJhbGciOiJIUzI1NiIs..."
  echo $TOKEN | omni jwt decode`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := jwt.Options{}
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.Header, _ = cmd.Flags().GetBool("header")
		opts.Payload, _ = cmd.Flags().GetBool("payload")
		opts.Raw, _ = cmd.Flags().GetBool("raw")

		return jwt.RunDecode(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(jwtCmd)
	jwtCmd.AddCommand(jwtDecodeCmd)

	jwtDecodeCmd.Flags().Bool("json", false, "output as JSON")
	jwtDecodeCmd.Flags().BoolP("header", "H", false, "show only header")
	jwtDecodeCmd.Flags().BoolP("payload", "p", false, "show only payload")
	jwtDecodeCmd.Flags().Bool("raw", false, "output raw JSON without formatting")
}
