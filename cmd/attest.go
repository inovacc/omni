package cmd

import (
	"github.com/inovacc/omni/internal/cli/attest"
	"github.com/spf13/cobra"
)

// attestCmd represents the `omni attest` command (generate).
var attestCmd = &cobra.Command{
	Use:   "attest [OPTION]... --artifact FILE",
	Short: "Generate a signed SLSA provenance attestation (DSSE/in-toto)",
	Long: `Generate an in-toto Statement v1 with a SLSA Provenance v1 predicate,
wrapped in a DSSE envelope and signed with a minisign-compatible Ed25519 secret
key (see 'omni sign keygen'). The passphrase is read from OMNI_SIGN_PASSPHRASE
— never a command-line flag. Key material is never accepted as a flag value.

The claimed SLSA level is fixed by ADR-0009 via builder.id; omni refuses to emit
any builder.id outside the ADR allowlist (no SLSA overclaim). There is no flag to
claim a higher level.

  -k, --key FILE          secret key file (*.key)
  -a, --artifact FILE     artifact to attest (its sha256 is the subject digest)
      --predicate-type T  predicate type (only: slsa-provenance)
      --predicate FILE    use a pre-built predicate JSON instead of building one
      --builder-id URI    builder.id (must be ADR-0009-allowed; default: local)
      --from-env          populate provenance from GITHUB_* env vars (release path)
  -o, --out FILE          output envelope path (default: stdout)

Examples:
  OMNI_SIGN_PASSPHRASE=pw omni attest --key release.key --artifact app.tar.gz --out app.intoto.jsonl
  OMNI_SIGN_PASSPHRASE=pw omni attest --key release.key --artifact app.tar.gz --from-env`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		opts := attest.GenOptions{}
		opts.KeyPath, _ = cmd.Flags().GetString("key")
		opts.ArtifactPath, _ = cmd.Flags().GetString("artifact")
		opts.PredicateType, _ = cmd.Flags().GetString("predicate-type")
		opts.PredicatePath, _ = cmd.Flags().GetString("predicate")
		opts.BuilderID, _ = cmd.Flags().GetString("builder-id")
		opts.FromEnv, _ = cmd.Flags().GetBool("from-env")
		opts.OutPath, _ = cmd.Flags().GetString("out")
		return attest.RunAttest(cmd.OutOrStdout(), opts)
	},
}

// attestVerifyCmd represents the `omni attest verify` subcommand.
var attestVerifyCmd = &cobra.Command{
	Use:   "verify [OPTION]... ENVELOPE",
	Short: "Verify a SLSA provenance attestation fail-closed",
	Long: `Verify a DSSE/in-toto SLSA provenance envelope against a public key.
Every failure mode (bad signature, wrong key, malformed envelope, digest
mismatch) exits non-zero with a classified error. With --artifact, also binds
the envelope to a specific artifact by sha256.

  -k, --key FILE       public key file (*.pub)
  -a, --artifact FILE  optional artifact to bind by sha256 to a subject

Examples:
  omni attest verify --key release.pub --artifact app.tar.gz app.intoto.jsonl`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := attest.VerifyOptions{}
		opts.KeyPath, _ = cmd.Flags().GetString("key")
		opts.ArtifactPath, _ = cmd.Flags().GetString("artifact")
		if len(args) == 1 {
			opts.EnvelopePath = args[0]
		} else {
			opts.EnvelopePath, _ = cmd.Flags().GetString("envelope")
		}
		return attest.RunVerify(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(attestCmd)
	attestCmd.AddCommand(attestVerifyCmd)

	attestCmd.Flags().StringP("key", "k", "", "secret key file (*.key)")
	attestCmd.Flags().StringP("artifact", "a", "", "artifact to attest")
	attestCmd.Flags().String("predicate-type", "slsa-provenance", "predicate type (only: slsa-provenance)")
	attestCmd.Flags().String("predicate", "", "pre-built predicate JSON file")
	attestCmd.Flags().String("builder-id", "", "builder.id (ADR-0009-allowed; default: local)")
	attestCmd.Flags().Bool("from-env", false, "populate provenance from GITHUB_* env vars")
	attestCmd.Flags().StringP("out", "o", "", "output envelope path (default: stdout)")

	attestVerifyCmd.Flags().StringP("key", "k", "", "public key file (*.pub)")
	attestVerifyCmd.Flags().StringP("artifact", "a", "", "artifact to bind by sha256")
	attestVerifyCmd.Flags().String("envelope", "", "envelope path (alternative to positional arg)")
}
