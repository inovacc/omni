package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/crypt"
	"github.com/spf13/cobra"
)

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt [OPTION]... [FILE]",
	Short: "Encrypt data using AES-256-GCM",
	Long: `Encrypt FILE or standard input using AES-256-GCM.

Uses PBKDF2 for key derivation with SHA-256.

  -p, --password STRING   password for encryption
  -P, --password-file FILE  read password from file
  -k, --key-file FILE     use key file for encryption
  -o, --output FILE       write output to file
  -a, --armor             ASCII armor (base64) output
  -i, --iterations N      PBKDF2 iterations (default 100000)

Password can also be set via omni_PASSWORD environment variable.

Examples:
  echo "secret" | omni encrypt -p mypassword
  omni encrypt -p mypassword -o secret.enc file.txt
  omni encrypt -P ~/.password -a file.txt
  omni_PASSWORD=pass omni encrypt file.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := crypt.CryptOptions{}

		opts.Password, _ = cmd.Flags().GetString("password")
		opts.PasswordFile, _ = cmd.Flags().GetString("password-file")
		opts.KeyFile, _ = cmd.Flags().GetString("key-file")
		opts.Output, _ = cmd.Flags().GetString("output")
		opts.Armor, _ = cmd.Flags().GetBool("armor")
		opts.Base64, _ = cmd.Flags().GetBool("base64")
		opts.Iterations, _ = cmd.Flags().GetInt("iterations")

		return crypt.RunEncrypt(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(encryptCmd)

	encryptCmd.Flags().StringP("password", "p", "", "password for encryption")
	encryptCmd.Flags().StringP("password-file", "P", "", "read password from file")
	encryptCmd.Flags().StringP("key-file", "k", "", "use key file for encryption")
	encryptCmd.Flags().StringP("output", "o", "", "write output to file")
	encryptCmd.Flags().BoolP("armor", "a", false, "ASCII armor (base64) output")
	encryptCmd.Flags().BoolP("base64", "b", false, "base64 output (same as -a)")
	encryptCmd.Flags().IntP("iterations", "i", 100000, "PBKDF2 iterations")
}
