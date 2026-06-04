package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/attest"
	"github.com/inovacc/omni/internal/cli/awk"
	"github.com/inovacc/omni/internal/cli/base"
	"github.com/inovacc/omni/internal/cli/caseconv"
	"github.com/inovacc/omni/internal/cli/cat"
	"github.com/inovacc/omni/internal/cli/column"
	"github.com/inovacc/omni/internal/cli/command"
	"github.com/inovacc/omni/internal/cli/cut"
	"github.com/inovacc/omni/internal/cli/fold"
	"github.com/inovacc/omni/internal/cli/grep"
	"github.com/inovacc/omni/internal/cli/hash"
	"github.com/inovacc/omni/internal/cli/head"
	"github.com/inovacc/omni/internal/cli/nl"
	"github.com/inovacc/omni/internal/cli/paste"
	"github.com/inovacc/omni/internal/cli/pipe"
	"github.com/inovacc/omni/internal/cli/rev"
	"github.com/inovacc/omni/internal/cli/sbom"
	"github.com/inovacc/omni/internal/cli/scan"
	"github.com/inovacc/omni/internal/cli/sed"
	"github.com/inovacc/omni/internal/cli/shuf"
	"github.com/inovacc/omni/internal/cli/sign"
	clstrings "github.com/inovacc/omni/internal/cli/strings"
	"github.com/inovacc/omni/internal/cli/tac"
	"github.com/inovacc/omni/internal/cli/tail"
	"github.com/inovacc/omni/internal/cli/text"
	"github.com/inovacc/omni/internal/cli/tr"
	"github.com/inovacc/omni/internal/cli/verify"
	"github.com/inovacc/omni/internal/cli/wc"
	"github.com/inovacc/omni/internal/cli/xxd"
	"github.com/spf13/cobra"
)

// buildPipeRegistry creates a unified command.Registry for commonly-piped commands.
// Commands registered here are dispatched directly without Cobra overhead.
// Commands not registered fall back to Cobra dispatch.
func buildPipeRegistry() *command.Registry {
	reg := command.NewRegistry()

	reg.Register("head", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return head.RunHead(w, r, args, head.HeadOptions{Lines: 10})
		},
	))

	reg.Register("tail", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return tail.RunTail(w, r, args, tail.TailOptions{Lines: 10})
		},
	))

	reg.Register("sort", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return text.RunSort(w, r, args, text.SortOptions{})
		},
	))

	reg.Register("uniq", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return text.RunUniq(w, r, args, text.UniqOptions{})
		},
	))

	reg.Register("cat", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return cat.RunCat(w, r, args, cat.CatOptions{})
		},
	))

	reg.Register("wc", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return wc.RunWC(w, r, args, wc.WCOptions{})
		},
	))

	reg.Register("cut", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return cut.RunCut(w, r, args, cut.CutOptions{})
		},
	))

	reg.Register("sed", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return sed.RunSed(w, r, args, sed.SedOptions{})
		},
	))

	reg.Register("nl", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return nl.RunNl(w, r, args, nl.NlOptions{})
		},
	))

	reg.Register("rev", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return rev.RunRev(w, r, args, rev.RevOptions{})
		},
	))

	reg.Register("tac", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return tac.RunTac(w, r, args, tac.TacOptions{})
		},
	))

	reg.Register("awk", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return awk.RunAwk(w, r, args, awk.AwkOptions{})
		},
	))

	reg.Register("fold", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return fold.RunFold(w, r, args, fold.FoldOptions{Width: 80})
		},
	))

	reg.Register("column", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return column.RunColumn(w, r, args, column.ColumnOptions{})
		},
	))

	reg.Register("paste", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return paste.RunPaste(w, r, args, paste.PasteOptions{})
		},
	))

	reg.Register("xxd", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return xxd.Run(w, r, args, xxd.Options{})
		},
	))

	// grep: first arg is the pattern, rest are file args
	reg.Register("grep", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("grep: missing pattern")
			}
			return grep.RunGrep(w, r, args[0], args[1:], grep.GrepOptions{})
		},
	))

	// tr: first arg is set1, optional second is set2
	reg.Register("tr", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("tr: missing operand")
			}
			set1 := args[0]
			set2 := ""
			if len(args) > 1 {
				set2 = args[1]
			}
			return tr.RunTr(w, r, set1, set2, tr.TrOptions{})
		},
	))

	// hash: first arg is algorithm, rest are file args
	reg.Register("hash", command.AdaptWriterArgs(
		func(w io.Writer, args []string) error {
			return hash.RunHash(w, args, hash.HashOptions{})
		},
	))

	reg.Register("base64", command.AdaptWriterArgs(
		func(w io.Writer, args []string) error {
			return base.RunBase64(w, args, base.BaseOptions{})
		},
	))

	reg.Register("base32", command.AdaptWriterArgs(
		func(w io.Writer, args []string) error {
			return base.RunBase32(w, args, base.BaseOptions{})
		},
	))

	reg.Register("caseconv", command.AdaptWriterArgs(
		func(w io.Writer, args []string) error {
			return caseconv.RunCase(w, args, caseconv.Options{Case: caseconv.CaseUpper})
		},
	))

	reg.Register("strings", command.AdaptWriterArgs(
		func(w io.Writer, args []string) error {
			return clstrings.RunStrings(w, args, clstrings.StringsOptions{})
		},
	))

	reg.Register("shuf", command.AdaptWriterArgs(
		func(w io.Writer, args []string) error {
			return shuf.RunShuf(w, args, shuf.ShufOptions{})
		},
	))

	// sign/verify: detached minisign over stdin or a file arg. keygen stays
	// Cobra-only — it has no stdin transform. Defaults read the passphrase from
	// OMNI_SIGN_PASSPHRASE; key material is referenced only by file path.
	reg.Register("sign", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return sign.RunSign(w, r, args, sign.SignOptions{})
		},
	))

	reg.Register("verify", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return verify.RunVerify(w, r, args, verify.VerifyOptions{})
		},
	))

	// attest-verify: verify a DSSE/SLSA provenance envelope read from stdin
	// against the public key whose path is args[0]. attest (generate) stays
	// Cobra-only — it consumes a file path and key/builder flags, not a stdin
	// stream. Artifact binding is unavailable over the pipe interface.
	reg.Register("attest-verify", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return attest.RunVerifyReader(w, r, args)
		},
	))

	// sbom: takes a PATH argument and ignores stdin (reader unused); registered
	// for unified-registry parity with sign/verify. Defaults to SPDX over an
	// auto-detected module-dir/binary path; the epoch-default timestamp keeps
	// the piped output deterministic.
	reg.Register("sbom", command.AdaptWriterReaderArgs(
		func(w io.Writer, _ io.Reader, args []string) error {
			return sbom.RunSBOM(w, args, sbom.SBOMOptions{Format: "spdx", From: "auto", OmniVersion: rootVersion()})
		},
	))

	// scan: reads an SBOM from stdin and scans it against the signed OSV DB.
	// Pipe stages take only (w, r, args), so the DB path/key and --fail-on gate
	// come from the OMNI_SCAN_* environment via OptionsFromEnv. scan source and
	// scan db update stay Cobra-only — they are not stdin transforms.
	reg.Register("scan", command.AdaptWriterReaderArgs(
		func(w io.Writer, r io.Reader, args []string) error {
			return scan.RunScanStdin(w, r, args, scan.OptionsFromEnv())
		},
	))

	return reg
}

var pipeCmd = &cobra.Command{
	Use:   "pipe {CMD}, {CMD}, ... | CMD | CMD",
	Short: "Chain omni commands without shell pipes",
	Long: `Chain multiple omni commands together, passing output from one to the next.

This allows creating pipelines of omni commands without using shell pipes,
making scripts more portable and avoiding shell-specific behavior.

Commands can be separated by:
  - Curly braces with commas: {cmd1}, {cmd2}, {cmd3} (recommended)
  - The | character: cmd1 | cmd2 | cmd3
  - A custom separator with --sep
  - As separate quoted arguments

Variable Substitution:
  Use $OUT (or custom var name with --var) to substitute previous output:
  - $OUT or ${OUT}     - single value substitution (uses last line)
  - [$OUT...]          - iterate over each line of output

Examples:
  # Using braces (recommended - clearest syntax)
  omni pipe '{ls -la}', '{grep .go}', '{wc -l}'
  omni pipe '{cat file.txt}', '{sort}', '{uniq}'
  omni pipe '{cat data.json}', '{jq .users[]}'

  # Using | separator (quote the whole thing)
  omni pipe "cat file.txt | grep pattern | sort | uniq"

  # Using separate arguments with | between
  omni pipe cat file.txt \| grep error \| sort \| uniq -c

  # Using custom separator
  omni pipe --sep "->" "cat file.txt -> grep error -> sort"

  # Multiple quoted commands
  omni pipe "cat file.txt" "grep pattern" "sort" "uniq"

  # With stdin
  echo "hello world" | omni pipe '{grep hello}', '{wc -l}'

  # Verbose mode to see intermediate results
  omni pipe -v '{cat file.txt}', '{head -10}', '{sort}'

  # JSON output with pipeline metadata
  omni pipe --json '{cat file.txt}', '{wc -l}'

  # Variable substitution - create folder with UUID
  omni pipe '{uuid -v 7}', '{mkdir $OUT}'

  # Custom variable name
  omni pipe --var UUID '{uuid -v 7}', '{mkdir $UUID}'

  # Iteration - create folder for each UUID
  omni pipe '{uuid -v 7 -n 10}', '{mkdir [$OUT...]}'

Supported commands include all omni commands:
  cat, grep, head, tail, sort, uniq, wc, cut, tr, sed, awk,
  base64, hex, json, jq, yq, curl, and many more.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := pipe.Options{}
		opts.OutputFormat = getOutputOpts(cmd).GetFormat()
		opts.Separator, _ = cmd.Flags().GetString("sep")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.VarName, _ = cmd.Flags().GetString("var")

		registry := pipe.NewRegistryWithUnified(rootCmd, buildPipeRegistry())

		// Check if we have stdin input
		stat, _ := os.Stdin.Stat()
		hasStdin := (stat.Mode() & os.ModeCharDevice) == 0

		if hasStdin && len(args) > 0 {
			// Read stdin and pass as initial input
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string

			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			input := strings.Join(lines, "\n")
			if len(lines) > 0 {
				input += "\n"
			}

			return pipe.RunWithInput(cmd.OutOrStdout(), input, args, opts, registry)
		}

		return pipe.Run(cmd.OutOrStdout(), args, opts, registry)
	},
}

func init() {
	rootCmd.AddCommand(pipeCmd)

	pipeCmd.Flags().StringP("sep", "s", "|", "command separator")
	pipeCmd.Flags().BoolP("verbose", "v", false, "show intermediate results")
	pipeCmd.Flags().String("var", "OUT", "variable name for output substitution (default: OUT)")
}
