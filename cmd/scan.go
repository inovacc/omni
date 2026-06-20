package cmd

import (
	"os"
	"path/filepath"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/scan"
	"github.com/spf13/cobra"
)

// defaultDBBaseURL is the canonical release location of the signed OSV bundle
// (osv-db.zip + osv-db.zip.minisig). Overridable with --url.
const defaultDBBaseURL = "https://github.com/inovacc/omni/releases/latest/download"

// scanCmd represents the scan command: it matches an SBOM against a signed OSV
// vulnerability database and gates CI via --fail-on.
var scanCmd = &cobra.Command{
	Use:   "scan <sbom>",
	Short: "Scan an SBOM against a signed OSV vulnerability database",
	Long: `Scan an SPDX-2.3 or CycloneDX-1.5 JSON SBOM against a pkg/sign-signed OSV
vulnerability database (osv-db.zip) and report matching vulnerabilities.

The database bundle is signature-verified on load with the public key given by
--db-key; a tampered or unsigned bundle fails closed. --max-db-age gates
staleness (a DB older than the threshold fails loudly). --fail-on <severity>
exits non-zero (conflict) when any finding is at or above the threshold, for CI
gating.

  --db FILE          path to the signed osv-db.zip bundle (required)
  --db-key FILE      minisign public key (*.pub) used to verify the bundle (required)
  --db-sig FILE      detached signature path (default: <db>.minisig)
  --fail-on LEVEL    fail (exit 1) on a finding >= LEVEL (none|low|medium|high|critical)
  --json             emit JSON instead of the text table
  --max-db-age DUR   fail if the DB is older than DUR (e.g. 168h); 0 disables
  --online           enable OSV-API enrichment over net/http (opt-in)

Reachability-aware source scanning (omni scan source) is deferred in v1.0 per
ADR-0008 and returns "unsupported".

Examples:
  omni scan --db osv-db.zip --db-key db.pub sbom.json
  omni scan --db osv-db.zip --db-key db.pub --fail-on high sbom.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return scan.RunScan(cmd.OutOrStdout(), args, scanOptsFromFlags(cmd))
	},
}

// scanSourceCmd represents `omni scan source`: reachability-aware source
// scanning. Deferred in v1.0 (ADR-0008) — returns ErrUnsupported.
var scanSourceCmd = &cobra.Command{
	Use:   "source <pattern>",
	Short: "Reachability-aware Go source scan (deferred in v1.0 — returns unsupported)",
	Long: `Reachability-aware scanning of a Go source tree, reporting only vulnerabilities
whose vulnerable symbol is actually called.

DEFERRED (ADR-0008): reachability requires golang.org/x/vuln, which execs
"go list" (violating omni's no-exec rule) and pollutes the main go.mod via MVS.
This command returns "unsupported" in v1.0; its future home is a self-contained
contrib/govulncheck-scan module (see docs/BACKLOG.md).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return scan.RunScanSource(cmd.OutOrStdout(), args, scanOptsFromFlags(cmd))
	},
}

// scanDBCmd groups OSV database management subcommands.
var scanDBCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage the OSV vulnerability database",
}

// scanDBUpdateCmd downloads the signed OSV bundle, verifies its signature with
// the pinned public key, and writes it to the cache dir only if verification
// passes (fail-closed: a tampered download writes nothing).
var scanDBUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Download and verify the OSV vulnerability database",
	Long: `Download the signed OSV bundle (osv-db.zip + osv-db.zip.minisig) from --url,
verify its detached signature with the public key from --db-key, and write both
files into --cache-dir ONLY if verification passes. A tampered or unsigned
download is fail-closed: nothing is written and the command exits non-zero.

  --url FILE        base URL serving osv-db.zip and osv-db.zip.minisig
  --cache-dir DIR   destination directory (default: <user cache>/omni/osv-db)
  --db-key FILE     minisign public key (*.pub) used to verify the bundle (required)`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		baseURL, _ := cmd.Flags().GetString("url")
		cacheDir, _ := cmd.Flags().GetString("cache-dir")
		if cacheDir == "" {
			dir, err := defaultCacheDir()
			if err != nil {
				return cmderr.Wrap(cmderr.ErrIO,
					"scan db update: cannot resolve default cache dir: "+err.Error())
			}
			cacheDir = dir
		}
		dbKey, _ := cmd.Flags().GetString("db-key")
		return scan.RunDBUpdate(cmd.OutOrStdout(), scanOptsFromFlags(cmd), baseURL, cacheDir, dbKey)
	},
}

// defaultCacheDir returns the default OSV-DB cache directory:
// <os.UserCacheDir>/omni/osv-db.
func defaultCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "omni", "osv-db"), nil
}

// scanOptsFromFlags builds the internal scan Options from the command flags.
func scanOptsFromFlags(cmd *cobra.Command) scan.Options {
	opts := scan.Options{}
	opts.DBPath, _ = cmd.Flags().GetString("db")
	opts.DBKeyPath, _ = cmd.Flags().GetString("db-key")
	opts.DBSigPath, _ = cmd.Flags().GetString("db-sig")
	opts.FailOn, _ = cmd.Flags().GetString("fail-on")
	opts.OutputFormat = getOutputOpts(cmd).GetFormat()
	opts.MaxDBAge, _ = cmd.Flags().GetDuration("max-db-age")
	opts.Online, _ = cmd.Flags().GetBool("online")
	return opts
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.AddCommand(scanSourceCmd)
	scanCmd.AddCommand(scanDBCmd)
	scanDBCmd.AddCommand(scanDBUpdateCmd)

	// DB-resolution and gating flags live on the parent scanCmd as persistent
	// flags so the source subcommand inherits the same DB/--fail-on options.
	scanCmd.PersistentFlags().String("db", "", "path to the signed osv-db.zip bundle")
	scanCmd.PersistentFlags().String("db-key", "", "minisign public key (*.pub) used to verify the bundle")
	scanCmd.PersistentFlags().String("db-sig", "", "detached signature path (default: <db>.minisig)")
	scanCmd.PersistentFlags().String("fail-on", "", "fail on a finding >= LEVEL (none|low|medium|high|critical)")
	scanCmd.PersistentFlags().Duration("max-db-age", time.Duration(0), "fail if the DB is older than this (0 disables)")
	scanCmd.PersistentFlags().Bool("online", false, "enable OSV-API enrichment over net/http (opt-in)")

	// db update download flags (db-key is inherited from scanCmd's persistent flags).
	scanDBUpdateCmd.Flags().String("url", defaultDBBaseURL, "base URL serving osv-db.zip and osv-db.zip.minisig")
	scanDBUpdateCmd.Flags().String("cache-dir", "", "destination directory (default: <user cache>/omni/osv-db)")
}
