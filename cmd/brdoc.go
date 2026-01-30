package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/brdoc"
	"github.com/spf13/cobra"
)

// brdocCmd represents the brdoc command
var brdocCmd = &cobra.Command{
	Use:   "brdoc",
	Short: "Brazilian document utilities (CPF, CNPJ)",
	Long: `Brazilian document validation, generation, and formatting.

Subcommands:
  cpf     CPF (Cadastro de Pessoas Físicas) operations
  cnpj    CNPJ (Cadastro Nacional de Pessoa Jurídica) operations

Examples:
  omni brdoc cpf --generate           # generate a valid CPF
  omni brdoc cpf --validate 123.456.789-09
  omni brdoc cnpj --generate          # generate alphanumeric CNPJ
  omni brdoc cnpj --generate --legacy # generate numeric-only CNPJ`,
}

// cpfCmd represents the cpf subcommand
var cpfCmd = &cobra.Command{
	Use:   "cpf [CPF...]",
	Short: "CPF operations (generate, validate, format)",
	Long: `CPF (Cadastro de Pessoas Físicas) operations.

Flags:
  -g, --generate    Generate valid CPF(s)
  -v, --validate    Validate CPF(s)
  -f, --format      Format CPF(s) as XXX.XXX.XXX-XX
  -n, --count       Number of CPFs to generate (default 1)
  --json            Output as JSON

Examples:
  omni brdoc cpf --generate              # generate one CPF
  omni brdoc cpf --generate -n 5         # generate 5 CPFs
  omni brdoc cpf --validate 12345678909
  omni brdoc cpf --validate 123.456.789-09
  omni brdoc cpf --format 12345678909
  omni brdoc cpf --generate --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := brdoc.Options{}

		opts.Generate, _ = cmd.Flags().GetBool("generate")
		opts.Validate, _ = cmd.Flags().GetBool("validate")
		opts.Format, _ = cmd.Flags().GetBool("format")
		opts.Count, _ = cmd.Flags().GetInt("count")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return brdoc.RunCPF(cmd.OutOrStdout(), args, opts)
	},
}

// cnpjCmd represents the cnpj subcommand
var cnpjCmd = &cobra.Command{
	Use:   "cnpj [CNPJ...]",
	Short: "CNPJ operations (generate, validate, format)",
	Long: `CNPJ (Cadastro Nacional de Pessoa Jurídica) operations.

Supports both numeric and alphanumeric CNPJ formats per SERPRO specification.

Flags:
  -g, --generate    Generate valid CNPJ(s)
  -v, --validate    Validate CNPJ(s)
  -f, --format      Format CNPJ(s) as XX.XXX.XXX/XXXX-XX
  -n, --count       Number of CNPJs to generate (default 1)
  -l, --legacy      Generate numeric-only CNPJ (14 digits)
  --json            Output as JSON

Examples:
  omni brdoc cnpj --generate              # generate alphanumeric CNPJ
  omni brdoc cnpj --generate --legacy     # generate numeric-only CNPJ
  omni brdoc cnpj --generate -n 5         # generate 5 CNPJs
  omni brdoc cnpj --validate 12.ABC.345/01DE-35
  omni brdoc cnpj --validate 11222333000181
  omni brdoc cnpj --format 11222333000181
  omni brdoc cnpj --generate --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := brdoc.Options{}

		opts.Generate, _ = cmd.Flags().GetBool("generate")
		opts.Validate, _ = cmd.Flags().GetBool("validate")
		opts.Format, _ = cmd.Flags().GetBool("format")
		opts.Count, _ = cmd.Flags().GetInt("count")
		opts.Legacy, _ = cmd.Flags().GetBool("legacy")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return brdoc.RunCNPJ(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(brdocCmd)

	// Add subcommands
	brdocCmd.AddCommand(cpfCmd)
	brdocCmd.AddCommand(cnpjCmd)

	// CPF flags
	cpfCmd.Flags().BoolP("generate", "g", false, "generate valid CPF(s)")
	cpfCmd.Flags().BoolP("validate", "v", false, "validate CPF(s)")
	cpfCmd.Flags().BoolP("format", "f", false, "format CPF(s)")
	cpfCmd.Flags().IntP("count", "n", 1, "number of CPFs to generate")
	cpfCmd.Flags().Bool("json", false, "output as JSON")

	// CNPJ flags
	cnpjCmd.Flags().BoolP("generate", "g", false, "generate valid CNPJ(s)")
	cnpjCmd.Flags().BoolP("validate", "v", false, "validate CNPJ(s)")
	cnpjCmd.Flags().BoolP("format", "f", false, "format CNPJ(s)")
	cnpjCmd.Flags().IntP("count", "n", 1, "number of CNPJs to generate")
	cnpjCmd.Flags().BoolP("legacy", "l", false, "generate numeric-only CNPJ")
	cnpjCmd.Flags().Bool("json", false, "output as JSON")
}
