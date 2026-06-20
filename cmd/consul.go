package cmd

// helplint:ignore — Long strings need omni-usage examples added in a future pass.

import (
	"context"

	"github.com/inovacc/omni/internal/cli/consul"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
	"github.com/spf13/cobra"
)

var (
	consulAddr      string
	consulToken     string
	consulNamespace string
	consulTLSSkip   bool
)

var consulCmd = &cobra.Command{
	Use:   "consul",
	Short: "HashiCorp Consul CLI (read-only)",
	Long: `HashiCorp Consul CLI for read-only cluster/catalog/KV inspection.

This is a read-only MVP: members, KV get, and catalog services. Write/mutating
operations (kv put/delete, service register) are intentionally not implemented.

Environment variables:
  CONSUL_HTTP_ADDR   Consul HTTP address (default: http://127.0.0.1:8500)
  CONSUL_HTTP_TOKEN  ACL token (sent as X-Consul-Token)
  CONSUL_NAMESPACE   Consul namespace

Examples:
  omni consul members
  omni consul kv get myapp/config
  omni consul services --json`,
}

func getConsulClient() (*consul.Client, error) {
	return consul.New(consul.Options{
		Address:   consulAddr,
		Token:     consulToken,
		Namespace: consulNamespace,
		TLSSkip:   consulTLSSkip,
	})
}

func consulIsJSON(cmd *cobra.Command) bool {
	return getOutputOpts(cmd).GetFormat() == output.FormatJSON
}

var consulMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "List cluster members",
	Long: `List the Consul cluster members (GET /v1/agent/members).

Examples:
  omni consul members
  omni consul members --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getConsulClient()
		if err != nil {
			return err
		}

		members, err := client.Members(context.Background())
		if err != nil {
			return err
		}

		return consul.PrintMembers(cmd.OutOrStdout(), members, consulIsJSON(cmd))
	},
}

var consulKVCmd = &cobra.Command{
	Use:   "kv",
	Short: "Key-value store operations (read-only)",
	Long:  `Read keys from the Consul KV store. Only "get" is implemented (read-only).`,
}

var consulKVGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a value from the KV store",
	Long: `Get a value from the Consul KV store (GET /v1/kv/<key>).

The stored value is base64-decoded before printing.

Examples:
  omni consul kv get myapp/config
  omni consul kv get myapp/config --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getConsulClient()
		if err != nil {
			return err
		}

		value, err := client.KVGet(context.Background(), args[0])
		if err != nil {
			return err
		}

		return consul.PrintKV(cmd.OutOrStdout(), args[0], value, consulIsJSON(cmd))
	},
}

var consulServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List catalog services",
	Long: `List services from the Consul catalog (GET /v1/catalog/services).

Each service is shown with its registered tags.

Examples:
  omni consul services
  omni consul services --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getConsulClient()
		if err != nil {
			return err
		}

		services, err := client.Services(context.Background())
		if err != nil {
			return err
		}

		return consul.PrintServices(cmd.OutOrStdout(), services, consulIsJSON(cmd))
	},
}

func init() {
	rootCmd.AddCommand(consulCmd)

	consulCmd.PersistentFlags().StringVar(&consulAddr, "address", "", "Consul HTTP address (env: CONSUL_HTTP_ADDR)")
	consulCmd.PersistentFlags().StringVar(&consulToken, "token", "", "ACL token (env: CONSUL_HTTP_TOKEN)")
	consulCmd.PersistentFlags().StringVar(&consulNamespace, "namespace", "", "Consul namespace (env: CONSUL_NAMESPACE)")
	consulCmd.PersistentFlags().BoolVar(&consulTLSSkip, "tls-skip-verify", false, "Skip TLS verification")

	consulCmd.AddCommand(consulMembersCmd)
	consulCmd.AddCommand(consulServicesCmd)

	consulKVCmd.AddCommand(consulKVGetCmd)
	consulCmd.AddCommand(consulKVCmd)
}
