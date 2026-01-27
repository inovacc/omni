package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt [OPTION]... [FILE]",
	Short: "Decrypt data using AES-256-GCM",
	Long: `Decrypt FILE or standard input using AES-256-GCM.

  -p, --password STRING   password for decryption
  -P, --password-file FILE  read password from file
  -k, --key-file FILE     use key file for decryption
  -o, --output FILE       write output to file
  -a, --armor             input is ASCII armored (base64)
  -i, --iterations N      PBKDF2 iterations (default 100000)

Password can also be set via omni_PASSWORD environment variable.

Examples:
  omni decrypt -p mypassword secret.enc
  omni decrypt -p mypassword -a < secret.b64
  omni decrypt -P ~/.password -o file.txt secret.enc
  cat secret.enc | omni_PASSWORD=pass omni decrypt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.CryptOptions{}

		opts.Password, _ = cmd.Flags().GetString("password")
		opts.PasswordFile, _ = cmd.Flags().GetString("password-file")
		opts.KeyFile, _ = cmd.Flags().GetString("key-file")
		opts.Output, _ = cmd.Flags().GetString("output")
		opts.Armor, _ = cmd.Flags().GetBool("armor")
		opts.Base64, _ = cmd.Flags().GetBool("base64")
		opts.Iterations, _ = cmd.Flags().GetInt("iterations")

		return cli.RunDecrypt(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)

	decryptCmd.Flags().StringP("password", "p", "", "password for decryption")
	decryptCmd.Flags().StringP("password-file", "P", "", "read password from file")
	decryptCmd.Flags().StringP("key-file", "k", "", "use key file for decryption")
	decryptCmd.Flags().StringP("output", "o", "", "write output to file")
	decryptCmd.Flags().BoolP("armor", "a", false, "input is ASCII armored (base64)")
	decryptCmd.Flags().BoolP("base64", "b", false, "input is base64 (same as -a)")
	decryptCmd.Flags().IntP("iterations", "i", 100000, "PBKDF2 iterations")
}
