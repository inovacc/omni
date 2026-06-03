package cmd

import (
	"github.com/inovacc/omni/internal/cli/sign"
	"github.com/spf13/cobra"
)

// signCmd represents the sign command.
var signCmd = &cobra.Command{
	Use:   "sign [OPTION]... [FILE]",
	Short: "Create a detached minisign signature",
	Long: `Sign FILE (or standard input) with a minisign-compatible Ed25519 secret key,
producing a detached *.minisig signature using the prehashed ("ED") scheme.

The secret key is referenced by file path only; the passphrase is read from the
OMNI_SIGN_PASSPHRASE environment variable or an interactive prompt — NEVER from
a command-line flag. Key material is never accepted as a flag value.

  -k, --key FILE          path to the secret key file (*.key)
  -s, --sig FILE          output signature path (default: <FILE>.minisig)
  -t, --trusted-comment   trusted comment embedded in (and signed by) the signature
      --untrusted-comment first-line comment of the signature file (not signed)

Examples:
  OMNI_SIGN_PASSPHRASE=pw omni sign --key release.key artifact.tar.gz
  omni sign --key release.key --sig artifact.sig artifact.tar.gz`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := sign.SignOptions{}
		opts.KeyPath, _ = cmd.Flags().GetString("key")
		opts.SigPath, _ = cmd.Flags().GetString("sig")
		opts.TrustedComment, _ = cmd.Flags().GetString("trusted-comment")
		opts.UntrustedComment, _ = cmd.Flags().GetString("untrusted-comment")
		return sign.RunSign(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

// signKeygenCmd represents the `omni sign keygen` subcommand.
var signKeygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a passphrase-protected Ed25519 key pair",
	Long: `Generate a new minisign-compatible Ed25519 key pair. The secret key is
encrypted at rest with a passphrase (libsodium SENSITIVE scrypt cost) and
written 0600; the public key is written 0644.

The passphrase is read from OMNI_SIGN_PASSPHRASE or an interactive prompt; it is
never accepted as a flag value.

      --pub FILE      output path for the public key (*.pub)
      --key FILE      output path for the secret key (*.key)
      --comment TEXT  optional untrusted comment written to the key files
      --force         overwrite existing key files

Examples:
  OMNI_SIGN_PASSPHRASE=pw omni sign keygen --pub release.pub --key release.key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := sign.KeygenOptions{}
		opts.PubPath, _ = cmd.Flags().GetString("pub")
		opts.KeyPath, _ = cmd.Flags().GetString("key")
		opts.Comment, _ = cmd.Flags().GetString("comment")
		opts.Force, _ = cmd.Flags().GetBool("force")
		return sign.RunKeygen(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.AddCommand(signKeygenCmd)

	signCmd.Flags().StringP("key", "k", "", "path to the secret key file")
	signCmd.Flags().StringP("sig", "s", "", "output signature path (default: <FILE>.minisig)")
	signCmd.Flags().StringP("trusted-comment", "t", "", "trusted comment embedded in the signature")
	signCmd.Flags().String("untrusted-comment", "", "first-line comment of the signature file")

	signKeygenCmd.Flags().String("pub", "", "output path for the public key")
	signKeygenCmd.Flags().String("key", "", "output path for the secret key")
	signKeygenCmd.Flags().String("comment", "", "optional untrusted comment for the key files")
	signKeygenCmd.Flags().Bool("force", false, "overwrite existing key files")
}
