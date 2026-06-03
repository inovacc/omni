package cmd

import (
	"github.com/inovacc/omni/internal/cli/sbom"
	"github.com/spf13/cobra"
)

// sbomVersion is the tool label embedded in generated SBOM documents. It is a
// package-level constant because cmd/ has no shared version accessor yet; when
// one is introduced, rootVersion should read from it.
const sbomVersion = "dev"

// rootVersion returns the omni version string used to label generated SBOMs.
func rootVersion() string { return sbomVersion }

var sbomCmd = &cobra.Command{
	Use:   "sbom [OPTION]... PATH",
	Short: "Generate an SBOM (SPDX 2.3 or CycloneDX 1.5) for a Go module dir or binary",
	Long: `Generate a byte-deterministic Software Bill of Materials for a Go module
directory (go.mod) or a built Go binary (debug/buildinfo). Every component
carries a normalized Go purl. Output is identical bytes for identical input,
enabling reproducible-build and golden-master pinning.

      --format spdx|cyclonedx   output format (default: spdx)
      --from   auto|module|binary  source kind (default: auto-detect from PATH)
      --source-date RFC3339     fixed creation timestamp (default: epoch)
      --out FILE                write to FILE instead of stdout
      --sign                    sign --out with a minisign key (requires --key, --out)
  -k, --key FILE                secret key path for --sign (passphrase via OMNI_SIGN_PASSPHRASE)
      --validate                validate emitted doc against the upstream schema
                                (requires building with -tags omni_sbomvalidate; else ErrUnsupported)

Examples:
  omni sbom . --format spdx
  omni sbom ./bin/omni --format cyclonedx
  omni sbom . --format spdx --out omni.spdx.json --sign --key release.key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := sbom.SBOMOptions{OmniVersion: rootVersion()}
		opts.Format, _ = cmd.Flags().GetString("format")
		opts.From, _ = cmd.Flags().GetString("from")
		opts.SourceDate, _ = cmd.Flags().GetString("source-date")
		opts.OutPath, _ = cmd.Flags().GetString("out")
		opts.Sign, _ = cmd.Flags().GetBool("sign")
		opts.KeyPath, _ = cmd.Flags().GetString("key")
		opts.Validate, _ = cmd.Flags().GetBool("validate")
		return sbom.RunSBOM(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(sbomCmd)
	sbomCmd.Flags().String("format", "spdx", "output format: spdx|cyclonedx")
	sbomCmd.Flags().String("from", "auto", "source kind: auto|module|binary")
	sbomCmd.Flags().String("source-date", "", "fixed RFC-3339 creation timestamp (default: epoch)")
	sbomCmd.Flags().String("out", "", "write to FILE instead of stdout")
	sbomCmd.Flags().Bool("sign", false, "sign --out with a minisign key")
	sbomCmd.Flags().StringP("key", "k", "", "secret key path for --sign")
	sbomCmd.Flags().Bool("validate", false, "validate emitted document against the upstream schema (requires -tags omni_sbomvalidate)")
}
