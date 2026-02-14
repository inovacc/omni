package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
	"github.com/inovacc/omni/internal/cli/vault"
	"github.com/spf13/cobra"
)

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "HashiCorp Vault CLI",
	Long: `HashiCorp Vault CLI for secrets management.

Environment variables:
  VAULT_ADDR       Vault server address (default: https://127.0.0.1:8200)
  VAULT_TOKEN      Authentication token
  VAULT_NAMESPACE  Vault namespace

Examples:
  omni vault status
  omni vault login -method=token
  omni vault kv get secret/myapp
  omni vault kv put secret/myapp key=value`,
}

// Global flags
var (
	vaultAddr      string
	vaultNamespace string
	vaultTLSSkip   bool
)

func getVaultClient() (*vault.Client, error) {
	return vault.New(vault.Options{
		Address:   vaultAddr,
		Namespace: vaultNamespace,
		TLSSkip:   vaultTLSSkip,
	})
}

// vault status
var vaultStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Vault server status",
	Long: `Show the status of the Vault server.

Examples:
  omni vault status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		status, err := client.Status()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		if getOutputOpts(cmd).GetFormat() == output.FormatJSON {
			return printJSON(status)
		}

		fmt.Printf("Key                Value\n")
		fmt.Printf("---                -----\n")
		fmt.Printf("Sealed             %t\n", status.Sealed)
		fmt.Printf("Initialized        %t\n", status.Initialized)
		fmt.Printf("Total Shares       %d\n", status.N)
		fmt.Printf("Threshold          %d\n", status.T)
		fmt.Printf("Version            %s\n", status.Version)
		fmt.Printf("Build Date         %s\n", status.BuildDate)
		fmt.Printf("Storage Type       %s\n", status.StorageType)
		fmt.Printf("Cluster Name       %s\n", status.ClusterName)
		fmt.Printf("Cluster ID         %s\n", status.ClusterID)

		return nil
	},
}

// vault login
var vaultLoginCmd = &cobra.Command{
	Use:   "login [token]",
	Short: "Authenticate to Vault",
	Long: `Authenticate to Vault using various methods.

Methods:
  token     - Direct token authentication (default)
  userpass  - Username/password authentication
  approle   - AppRole authentication

Examples:
  omni vault login                           # Prompts for token
  omni vault login s.xxxxx                   # Direct token
  omni vault login -method=userpass -username=admin
  omni vault login -method=approle -role-id=xxx -secret-id=xxx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		ctx := context.Background()
		method, _ := cmd.Flags().GetString("method")
		noStore, _ := cmd.Flags().GetBool("no-store")

		var token string

		switch method {
		case "token", "":
			if len(args) > 0 {
				token = args[0]
			} else {
				fmt.Print("Token (will be hidden): ")
				_, _ = fmt.Scanln(&token)
			}

			if err := client.LoginToken(token); err != nil {
				return err
			}

			fmt.Println("Success! You are now authenticated.")

		case "userpass":
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")
			mount, _ := cmd.Flags().GetString("path")

			if username == "" {
				return fmt.Errorf("username is required for userpass auth")
			}

			if password == "" {
				fmt.Print("Password (will be hidden): ")
				_, _ = fmt.Scanln(&password)
			}

			secret, err := client.LoginUserpass(ctx, username, password, mount)
			if err != nil {
				return err
			}

			if secret != nil && secret.Auth != nil {
				token = secret.Auth.ClientToken
				fmt.Printf("Success! You are now authenticated.\n")
				fmt.Printf("Token: %s\n", token)
				fmt.Printf("Token Duration: %ds\n", secret.Auth.LeaseDuration)
			}

		case "approle":
			roleID, _ := cmd.Flags().GetString("role-id")
			secretID, _ := cmd.Flags().GetString("secret-id")
			mount, _ := cmd.Flags().GetString("path")

			if roleID == "" {
				return fmt.Errorf("role-id is required for approle auth")
			}

			secret, err := client.LoginAppRole(ctx, roleID, secretID, mount)
			if err != nil {
				return err
			}

			if secret != nil && secret.Auth != nil {
				token = secret.Auth.ClientToken
				fmt.Printf("Success! You are now authenticated.\n")
				fmt.Printf("Token: %s\n", token)
				fmt.Printf("Token Duration: %ds\n", secret.Auth.LeaseDuration)
			}

		default:
			return fmt.Errorf("unsupported auth method: %s", method)
		}

		if !noStore && token != "" {
			client.SetToken(token)
			if err := client.SaveToken(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save token: %v\n", err)
			}
		}

		return nil
	},
}

// vault token
var vaultTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Token operations",
	Long:  `Token management operations.`,
}

var vaultTokenLookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup current token",
	Long: `Display information about the current token.

Examples:
  omni vault token lookup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		secret, err := client.TokenLookupSelf()
		if err != nil {
			return fmt.Errorf("failed to lookup token: %w", err)
		}

		if getOutputOpts(cmd).GetFormat() == output.FormatJSON {
			return printJSON(secret.Data)
		}

		fmt.Printf("Key                 Value\n")
		fmt.Printf("---                 -----\n")

		for k, v := range secret.Data {
			fmt.Printf("%-20s%v\n", k, v)
		}

		return nil
	},
}

var vaultTokenRenewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew current token",
	Long: `Renew the current token's lease.

Examples:
  omni vault token renew
  omni vault token renew -increment=3600`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		increment, _ := cmd.Flags().GetInt("increment")

		secret, err := client.TokenRenewSelf(increment)
		if err != nil {
			return fmt.Errorf("failed to renew token: %w", err)
		}

		if getOutputOpts(cmd).GetFormat() == output.FormatJSON {
			return printJSON(secret)
		}

		fmt.Println("Success! Token renewed.")

		if secret.Auth != nil {
			fmt.Printf("Token Duration: %ds\n", secret.Auth.LeaseDuration)
		}

		return nil
	},
}

var vaultTokenRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke current token",
	Long: `Revoke the current token.

Examples:
  omni vault token revoke`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		if err := client.TokenRevokeSelf(); err != nil {
			return fmt.Errorf("failed to revoke token: %w", err)
		}

		fmt.Println("Success! Token revoked.")

		return nil
	},
}

// vault read
var vaultReadCmd = &cobra.Command{
	Use:   "read <path>",
	Short: "Read secrets",
	Long: `Read a secret from Vault at the given path.

Examples:
  omni vault read secret/data/myapp
  omni vault read -field=password secret/data/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]
		field, _ := cmd.Flags().GetString("field")

		secret, err := client.Read(ctx, path)
		if err != nil {
			return err
		}

		if secret == nil {
			return fmt.Errorf("no secret found at %s", path)
		}

		if field != "" {
			if val, ok := secret.Data[field]; ok {
				fmt.Println(val)
				return nil
			}

			return fmt.Errorf("field %q not found", field)
		}

		if getOutputOpts(cmd).GetFormat() == output.FormatJSON {
			return printJSON(secret.Data)
		}

		fmt.Printf("Key                 Value\n")
		fmt.Printf("---                 -----\n")

		for k, v := range secret.Data {
			fmt.Printf("%-20s%v\n", k, v)
		}

		return nil
	},
}

// vault write
var vaultWriteCmd = &cobra.Command{
	Use:   "write <path> [key=value...]",
	Short: "Write secrets",
	Long: `Write a secret to Vault at the given path.

Examples:
  omni vault write secret/data/myapp key=value
  omni vault write secret/data/myapp username=admin password=secret`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]
		data := make(map[string]any)

		for _, arg := range args[1:] {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid key=value: %s", arg)
			}

			data[parts[0]] = parts[1]
		}

		secret, err := client.Write(ctx, path, data)
		if err != nil {
			return err
		}

		if getOutputOpts(cmd).GetFormat() == output.FormatJSON && secret != nil {
			return printJSON(secret.Data)
		}

		fmt.Println("Success! Data written to:", path)

		return nil
	},
}

// vault list
var vaultListCmd = &cobra.Command{
	Use:   "list <path>",
	Short: "List secrets",
	Long: `List secrets at the given path.

Examples:
  omni vault list secret/metadata/`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]

		secret, err := client.List(ctx, path)
		if err != nil {
			return err
		}

		if secret == nil || secret.Data == nil {
			return fmt.Errorf("no entries found at %s", path)
		}

		keys, ok := secret.Data["keys"].([]any)
		if !ok {
			return fmt.Errorf("unexpected response format")
		}

		if getOutputOpts(cmd).GetFormat() == output.FormatJSON {
			return printJSON(keys)
		}

		fmt.Println("Keys")
		fmt.Println("----")

		for _, k := range keys {
			fmt.Println(k)
		}

		return nil
	},
}

// vault delete
var vaultDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete secrets",
	Long: `Delete a secret at the given path.

Examples:
  omni vault delete secret/data/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]

		_, err = client.Delete(ctx, path)
		if err != nil {
			return err
		}

		fmt.Println("Success! Data deleted at:", path)

		return nil
	},
}

// vault kv
var vaultKVCmd = &cobra.Command{
	Use:   "kv",
	Short: "KV secrets engine operations",
	Long:  `Interact with Vault's KV secrets engine (v2).`,
}

var kvMount string

// vault kv get
var vaultKVGetCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Get a secret from KV",
	Long: `Retrieve a secret from the KV secrets engine.

Examples:
  omni vault kv get secret/myapp
  omni vault kv get -mount=kv myapp
  omni vault kv get -version=2 secret/myapp
  omni vault kv get -field=password secret/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]
		version, _ := cmd.Flags().GetInt("version")
		field, _ := cmd.Flags().GetString("field")

		kv := client.NewKV(kvMount)

		secret, err := kv.Get(ctx, path, version)
		if err != nil {
			return err
		}

		if secret == nil {
			return fmt.Errorf("no secret found at %s", path)
		}

		if field != "" {
			if val, ok := secret.Data[field]; ok {
				fmt.Println(val)
				return nil
			}

			return fmt.Errorf("field %q not found", field)
		}

		if getOutputOpts(cmd).GetFormat() == output.FormatJSON {
			return printJSON(secret.Data)
		}

		fmt.Printf("====== Secret Path ======\n")
		fmt.Printf("%s/%s\n\n", kvMount, path)
		fmt.Printf("======= Metadata =======\n")
		fmt.Printf("Version: %d\n", secret.VersionMetadata.Version)
		fmt.Printf("Created: %s\n", secret.VersionMetadata.CreatedTime)

		if !secret.VersionMetadata.DeletionTime.IsZero() {
			fmt.Printf("Deleted: %s\n", secret.VersionMetadata.DeletionTime)
		}

		fmt.Printf("\n======== Data ========\n")
		fmt.Printf("Key                 Value\n")
		fmt.Printf("---                 -----\n")

		for k, v := range secret.Data {
			fmt.Printf("%-20s%v\n", k, v)
		}

		return nil
	},
}

// vault kv put
var vaultKVPutCmd = &cobra.Command{
	Use:   "put <path> [key=value...]",
	Short: "Put a secret into KV",
	Long: `Write a secret to the KV secrets engine.

Examples:
  omni vault kv put secret/myapp key=value
  omni vault kv put -mount=kv myapp username=admin password=secret`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]
		data := make(map[string]any)

		for _, arg := range args[1:] {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid key=value: %s", arg)
			}

			data[parts[0]] = parts[1]
		}

		kv := client.NewKV(kvMount)

		secret, err := kv.Put(ctx, path, data)
		if err != nil {
			return err
		}

		if getOutputOpts(cmd).GetFormat() == output.FormatJSON && secret != nil {
			return printJSON(map[string]any{
				"version":      secret.VersionMetadata.Version,
				"created_time": secret.VersionMetadata.CreatedTime,
			})
		}

		fmt.Printf("Success! Data written to: %s/%s\n", kvMount, path)

		if secret != nil {
			fmt.Printf("Version: %d\n", secret.VersionMetadata.Version)
		}

		return nil
	},
}

// vault kv delete
var vaultKVDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a secret from KV",
	Long: `Soft delete a secret from the KV secrets engine.

Examples:
  omni vault kv delete secret/myapp
  omni vault kv delete -versions=1,2,3 secret/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]
		versionsStr, _ := cmd.Flags().GetString("versions")

		kv := client.NewKV(kvMount)

		if versionsStr != "" {
			versions := parseVersions(versionsStr)
			if err := kv.DeleteVersions(ctx, path, versions); err != nil {
				return err
			}

			fmt.Printf("Success! Versions %v deleted at: %s/%s\n", versions, kvMount, path)
		} else {
			if err := kv.Delete(ctx, path); err != nil {
				return err
			}

			fmt.Printf("Success! Data deleted at: %s/%s\n", kvMount, path)
		}

		return nil
	},
}

// vault kv list
var vaultKVListCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List secrets in KV",
	Long: `List secrets at a path in the KV secrets engine.

Examples:
  omni vault kv list secret/
  omni vault kv list -mount=kv myapp/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := ""

		if len(args) > 0 {
			path = args[0]
		}

		kv := client.NewKV(kvMount)

		keys, err := kv.List(ctx, path)
		if err != nil {
			return err
		}

		if keys == nil || len(keys) == 0 {
			return fmt.Errorf("no entries found")
		}

		if getOutputOpts(cmd).GetFormat() == output.FormatJSON {
			return printJSON(keys)
		}

		fmt.Println("Keys")
		fmt.Println("----")

		for _, k := range keys {
			fmt.Println(k)
		}

		return nil
	},
}

// vault kv destroy
var vaultKVDestroyCmd = &cobra.Command{
	Use:   "destroy <path>",
	Short: "Permanently destroy secret versions",
	Long: `Permanently destroy versions of a secret in the KV secrets engine.

Examples:
  omni vault kv destroy -versions=1,2,3 secret/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]
		versionsStr, _ := cmd.Flags().GetString("versions")

		if versionsStr == "" {
			return fmt.Errorf("-versions flag is required")
		}

		versions := parseVersions(versionsStr)
		kv := client.NewKV(kvMount)

		if err := kv.Destroy(ctx, path, versions); err != nil {
			return err
		}

		fmt.Printf("Success! Versions %v destroyed at: %s/%s\n", versions, kvMount, path)

		return nil
	},
}

// vault kv undelete
var vaultKVUndeleteCmd = &cobra.Command{
	Use:   "undelete <path>",
	Short: "Restore deleted secret versions",
	Long: `Restore deleted versions of a secret in the KV secrets engine.

Examples:
  omni vault kv undelete -versions=1,2,3 secret/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]
		versionsStr, _ := cmd.Flags().GetString("versions")

		if versionsStr == "" {
			return fmt.Errorf("-versions flag is required")
		}

		versions := parseVersions(versionsStr)
		kv := client.NewKV(kvMount)

		if err := kv.Undelete(ctx, path, versions); err != nil {
			return err
		}

		fmt.Printf("Success! Versions %v restored at: %s/%s\n", versions, kvMount, path)

		return nil
	},
}

// vault kv metadata
var vaultKVMetadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "KV metadata operations",
	Long:  `Manage metadata for secrets in the KV secrets engine.`,
}

var vaultKVMetadataGetCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Get secret metadata",
	Long: `Retrieve metadata for a secret.

Examples:
  omni vault kv metadata get secret/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]

		kv := client.NewKV(kvMount)

		metadata, err := kv.GetMetadata(ctx, path)
		if err != nil {
			return err
		}

		if getOutputOpts(cmd).GetFormat() == output.FormatJSON {
			return printJSON(metadata)
		}

		fmt.Printf("====== Metadata Path ======\n")
		fmt.Printf("%s/metadata/%s\n\n", kvMount, path)
		fmt.Printf("Current Version: %d\n", metadata.CurrentVersion)
		fmt.Printf("Oldest Version: %d\n", metadata.OldestVersion)
		fmt.Printf("Created: %s\n", metadata.CreatedTime)
		fmt.Printf("Updated: %s\n", metadata.UpdatedTime)
		fmt.Printf("Max Versions: %d\n", metadata.MaxVersions)

		if len(metadata.Versions) > 0 {
			fmt.Printf("\n======= Versions =======\n")

			for v, info := range metadata.Versions {
				fmt.Printf("Version %s: created=%s destroyed=%t\n", v, info.CreatedTime, info.Destroyed)
			}
		}

		return nil
	},
}

var vaultKVMetadataDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete all versions and metadata",
	Long: `Permanently delete all versions and metadata for a secret.

Examples:
  omni vault kv metadata delete secret/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getVaultClient()
		if err != nil {
			return err
		}

		if err := client.LoadToken(); err != nil {
			return fmt.Errorf("no token available: %w", err)
		}

		ctx := context.Background()
		path := args[0]

		kv := client.NewKV(kvMount)

		if err := kv.DeleteMetadata(ctx, path); err != nil {
			return err
		}

		fmt.Printf("Success! All versions and metadata deleted at: %s/%s\n", kvMount, path)

		return nil
	},
}

func parseVersions(s string) []int {
	var versions []int

	for v := range strings.SplitSeq(s, ",") {
		v = strings.TrimSpace(v)

		var ver int
		if _, err := fmt.Sscanf(v, "%d", &ver); err == nil {
			versions = append(versions, ver)
		}
	}

	return versions
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	return enc.Encode(v)
}

func init() {
	rootCmd.AddCommand(vaultCmd)

	// Global flags
	vaultCmd.PersistentFlags().StringVar(&vaultAddr, "address", "", "Vault server address")
	vaultCmd.PersistentFlags().StringVar(&vaultNamespace, "namespace", "", "Vault namespace")
	vaultCmd.PersistentFlags().BoolVar(&vaultTLSSkip, "tls-skip-verify", false, "Skip TLS verification")

	// status
	vaultCmd.AddCommand(vaultStatusCmd)

	// login
	vaultLoginCmd.Flags().String("method", "token", "Auth method (token, userpass, approle)")
	vaultLoginCmd.Flags().String("username", "", "Username for userpass auth")
	vaultLoginCmd.Flags().String("password", "", "Password for userpass auth")
	vaultLoginCmd.Flags().String("role-id", "", "Role ID for approle auth")
	vaultLoginCmd.Flags().String("secret-id", "", "Secret ID for approle auth")
	vaultLoginCmd.Flags().String("path", "", "Mount path for auth method")
	vaultLoginCmd.Flags().Bool("no-store", false, "Don't save token to file")
	vaultCmd.AddCommand(vaultLoginCmd)

	// token
	vaultTokenRenewCmd.Flags().Int("increment", 0, "Lease increment in seconds")
	vaultTokenCmd.AddCommand(vaultTokenLookupCmd)
	vaultTokenCmd.AddCommand(vaultTokenRenewCmd)
	vaultTokenCmd.AddCommand(vaultTokenRevokeCmd)
	vaultCmd.AddCommand(vaultTokenCmd)

	// read
	vaultReadCmd.Flags().String("field", "", "Print only this field")
	vaultCmd.AddCommand(vaultReadCmd)

	// write
	vaultCmd.AddCommand(vaultWriteCmd)

	// list
	vaultCmd.AddCommand(vaultListCmd)

	// delete
	vaultCmd.AddCommand(vaultDeleteCmd)

	// kv
	vaultKVCmd.PersistentFlags().StringVar(&kvMount, "mount", "secret", "KV mount path")

	// kv get
	vaultKVGetCmd.Flags().Int("version", 0, "Version to retrieve (0 = latest)")
	vaultKVGetCmd.Flags().String("field", "", "Print only this field")
	vaultKVCmd.AddCommand(vaultKVGetCmd)

	// kv put
	vaultKVCmd.AddCommand(vaultKVPutCmd)

	// kv delete
	vaultKVDeleteCmd.Flags().String("versions", "", "Comma-separated versions to delete")
	vaultKVCmd.AddCommand(vaultKVDeleteCmd)

	// kv list
	vaultKVCmd.AddCommand(vaultKVListCmd)

	// kv destroy
	vaultKVDestroyCmd.Flags().String("versions", "", "Comma-separated versions to destroy (required)")
	vaultKVCmd.AddCommand(vaultKVDestroyCmd)

	// kv undelete
	vaultKVUndeleteCmd.Flags().String("versions", "", "Comma-separated versions to restore (required)")
	vaultKVCmd.AddCommand(vaultKVUndeleteCmd)

	// kv metadata
	vaultKVMetadataCmd.AddCommand(vaultKVMetadataGetCmd)
	vaultKVMetadataCmd.AddCommand(vaultKVMetadataDeleteCmd)
	vaultKVCmd.AddCommand(vaultKVMetadataCmd)

	vaultCmd.AddCommand(vaultKVCmd)
}
