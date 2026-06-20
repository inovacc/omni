package cmd

// helplint:ignore — Long strings need omni-usage examples added in a future pass.

import (
	"context"

	"github.com/inovacc/omni/internal/cli/nomad"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
	"github.com/spf13/cobra"
)

var (
	nomadAddr      string
	nomadToken     string
	nomadNamespace string
	nomadRegion    string
	nomadTLSSkip   bool
)

var nomadCmd = &cobra.Command{
	Use:   "nomad",
	Short: "HashiCorp Nomad CLI (read-only)",
	Long: `HashiCorp Nomad CLI for read-only job/node/allocation inspection.

This is a read-only MVP: job list, node list, and alloc list. Write/mutating
operations (job run/stop, node drain, alloc signal/restart) are intentionally
not implemented.

Environment variables:
  NOMAD_ADDR       Nomad HTTP address (default: http://127.0.0.1:4646)
  NOMAD_TOKEN      ACL token (sent as X-Nomad-Token)
  NOMAD_NAMESPACE  Nomad namespace
  NOMAD_REGION     Nomad region

Examples:
  omni nomad job list
  omni nomad node list
  omni nomad alloc list --json`,
}

func getNomadClient() (*nomad.Client, error) {
	return nomad.New(nomad.Options{
		Address:   nomadAddr,
		Token:     nomadToken,
		Namespace: nomadNamespace,
		Region:    nomadRegion,
		TLSSkip:   nomadTLSSkip,
	})
}

func nomadIsJSON(cmd *cobra.Command) bool {
	return getOutputOpts(cmd).GetFormat() == output.FormatJSON
}

var nomadJobCmd = &cobra.Command{
	Use:   "job",
	Short: "Job operations (read-only)",
	Long:  `Inspect Nomad jobs. Only "list" is implemented (read-only).`,
}

var nomadJobListCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs",
	Long: `List Nomad jobs (GET /v1/jobs).

Examples:
  omni nomad job list
  omni nomad job list --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getNomadClient()
		if err != nil {
			return err
		}

		jobs, err := client.JobList(context.Background())
		if err != nil {
			return err
		}

		return nomad.PrintJobs(cmd.OutOrStdout(), jobs, nomadIsJSON(cmd))
	},
}

var nomadNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Node operations (read-only)",
	Long:  `Inspect Nomad nodes. Only "list" is implemented (read-only).`,
}

var nomadNodeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List nodes",
	Long: `List Nomad client nodes (GET /v1/nodes).

Examples:
  omni nomad node list
  omni nomad node list --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getNomadClient()
		if err != nil {
			return err
		}

		nodes, err := client.NodeList(context.Background())
		if err != nil {
			return err
		}

		return nomad.PrintNodes(cmd.OutOrStdout(), nodes, nomadIsJSON(cmd))
	},
}

var nomadAllocCmd = &cobra.Command{
	Use:   "alloc",
	Short: "Allocation operations (read-only)",
	Long:  `Inspect Nomad allocations. Only "list" is implemented (read-only).`,
}

var nomadAllocListCmd = &cobra.Command{
	Use:   "list",
	Short: "List allocations",
	Long: `List Nomad allocations (GET /v1/allocations).

Examples:
  omni nomad alloc list
  omni nomad alloc list --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getNomadClient()
		if err != nil {
			return err
		}

		allocs, err := client.AllocList(context.Background())
		if err != nil {
			return err
		}

		return nomad.PrintAllocs(cmd.OutOrStdout(), allocs, nomadIsJSON(cmd))
	},
}

func init() {
	rootCmd.AddCommand(nomadCmd)

	nomadCmd.PersistentFlags().StringVar(&nomadAddr, "address", "", "Nomad HTTP address (env: NOMAD_ADDR)")
	nomadCmd.PersistentFlags().StringVar(&nomadToken, "token", "", "ACL token (env: NOMAD_TOKEN)")
	nomadCmd.PersistentFlags().StringVar(&nomadNamespace, "namespace", "", "Nomad namespace (env: NOMAD_NAMESPACE)")
	nomadCmd.PersistentFlags().StringVar(&nomadRegion, "region", "", "Nomad region (env: NOMAD_REGION)")
	nomadCmd.PersistentFlags().BoolVar(&nomadTLSSkip, "tls-skip-verify", false, "Skip TLS verification")

	nomadJobCmd.AddCommand(nomadJobListCmd)
	nomadCmd.AddCommand(nomadJobCmd)

	nomadNodeCmd.AddCommand(nomadNodeListCmd)
	nomadCmd.AddCommand(nomadNodeCmd)

	nomadAllocCmd.AddCommand(nomadAllocListCmd)
	nomadCmd.AddCommand(nomadAllocCmd)
}
