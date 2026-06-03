package cmd

import (
	"github.com/inovacc/omni/internal/cli/verify"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command.
var verifyCmd = &cobra.Command{
	Use:   "verify [OPTION]... FILE",
	Short: "Verify a detached minisign signature (fail-closed)",
	Long: `Verify FILE against a detached minisign signature and a public key.

Verification is fail-closed: it succeeds ONLY when the data signature AND the
trusted-comment (global) signature both verify and the signature key id matches
the public key. Any failure exits non-zero.

The public key is referenced by file path only; key material is never accepted
as a flag value.

  -k, --key FILE      path to the public key file (*.pub)
  -s, --sig FILE      path to the signature file (default: <FILE>.minisig)
      --bundle FILE   verify a Sigstore bundle (unsupported here; provided by
                      the separate github.com/inovacc/omni/contrib/sigstore-verify module)
      --trusted-root  Sigstore trusted-root path (sigstore-verify module only)
      --cert-identity Sigstore certificate identity (sigstore-verify module only)
      --cert-oidc-issuer Sigstore OIDC issuer (sigstore-verify module only)

Examples:
  omni verify --key release.pub artifact.tar.gz
  omni verify --key release.pub --sig artifact.sig artifact.tar.gz`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := verify.VerifyOptions{}
		opts.PubPath, _ = cmd.Flags().GetString("key")
		opts.SigPath, _ = cmd.Flags().GetString("sig")
		opts.BundlePath, _ = cmd.Flags().GetString("bundle")
		opts.TrustedRoot, _ = cmd.Flags().GetString("trusted-root")
		opts.CertIdentity, _ = cmd.Flags().GetString("cert-identity")
		opts.CertOIDCIssuer, _ = cmd.Flags().GetString("cert-oidc-issuer")
		return verify.RunVerify(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	verifyCmd.Flags().StringP("key", "k", "", "path to the public key file")
	verifyCmd.Flags().StringP("sig", "s", "", "path to the signature file (default: <FILE>.minisig)")
	verifyCmd.Flags().String("bundle", "", "verify a Sigstore bundle (unsupported here; see contrib/sigstore-verify module)")
	verifyCmd.Flags().String("trusted-root", "", "Sigstore trusted-root path (sigstore-verify module only)")
	verifyCmd.Flags().String("cert-identity", "", "Sigstore certificate identity (sigstore-verify module only)")
	verifyCmd.Flags().String("cert-oidc-issuer", "", "Sigstore OIDC issuer (sigstore-verify module only)")
}
