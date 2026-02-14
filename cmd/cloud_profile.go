package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/cloud/profile"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var cloudProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage cloud profiles",
	Long:  `Manage cloud profiles for AWS, Azure, and GCP.`,
}

var cloudProfileAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new cloud profile",
	Long: `Add a new cloud profile with encrypted credentials.

For AWS:
  Prompts for Access Key ID and Secret Access Key, or use flags.

For Azure:
  Prompts for Tenant ID, Client ID, Client Secret, and Subscription ID.

For GCP:
  Requires --key-file pointing to a service account JSON file.

Examples:
  # Add AWS profile interactively
  omni cloud profile add myaws --provider aws --region us-east-1

  # Add AWS profile with flags
  omni cloud profile add myaws --provider aws \
    --access-key-id AKIAXXXXXXXX \
    --secret-access-key XXXXXXXX

  # Add Azure profile
  omni cloud profile add myazure --provider azure

  # Add GCP profile
  omni cloud profile add mygcp --provider gcp --key-file /path/to/sa.json`,
	Args: cobra.ExactArgs(1),
	RunE: runCloudProfileAdd,
}

var cloudProfileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cloud profiles",
	Long: `List all cloud profiles, optionally filtered by provider.

Examples:
  omni cloud profile list
  omni cloud profile list --provider aws`,
	RunE: runCloudProfileList,
}

var cloudProfileShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show profile details",
	Long: `Show details of a cloud profile (without credentials).

Examples:
  omni cloud profile show myaws
  omni cloud profile show myaws --provider aws`,
	Args: cobra.ExactArgs(1),
	RunE: runCloudProfileShow,
}

var cloudProfileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set a profile as default",
	Long: `Set a profile as the default for its provider.

Examples:
  omni cloud profile use myaws
  omni cloud profile use myaws --provider aws`,
	Args: cobra.ExactArgs(1),
	RunE: runCloudProfileUse,
}

var cloudProfileDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a cloud profile",
	Long: `Delete a cloud profile and its encrypted credentials.

Examples:
  omni cloud profile delete myaws --provider aws
  omni cloud profile delete myaws --provider aws --force`,
	Args: cobra.ExactArgs(1),
	RunE: runCloudProfileDelete,
}

var cloudProfileImportCmd = &cobra.Command{
	Use:   "import [name]",
	Short: "Import credentials from existing cloud CLI",
	Long: `Import existing credentials from AWS, Azure, or GCP CLI configurations.

AWS:
  Reads from ~/.aws/credentials and ~/.aws/config
  Use --source to specify which AWS profile to import (default: "default")

Azure:
  Requires a service principal JSON file (Azure CLI tokens cannot be migrated)
  Use --source to specify the service principal file path
  Create one with: az ad sp create-for-rbac --name omni-sp --sdk-auth > ~/.azure/sp.json

GCP:
  Imports service account credentials from GOOGLE_APPLICATION_CREDENTIALS
  or ~/.config/gcloud/ directory
  Note: Application Default Credentials (authorized_user) cannot be migrated

Examples:
  # Import default AWS profile
  omni cloud profile import --provider aws

  # Import specific AWS profile with custom name
  omni cloud profile import myaws --provider aws --source prod

  # List available AWS profiles
  omni cloud profile import --provider aws --list

  # Import Azure service principal
  omni cloud profile import --provider azure --source ~/.azure/sp.json

  # Import GCP service account
  omni cloud profile import --provider gcp --source /path/to/sa.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCloudProfileImport,
}

func init() {
	cloudCmd.AddCommand(cloudProfileCmd)
	cloudProfileCmd.AddCommand(cloudProfileAddCmd)
	cloudProfileCmd.AddCommand(cloudProfileListCmd)
	cloudProfileCmd.AddCommand(cloudProfileShowCmd)
	cloudProfileCmd.AddCommand(cloudProfileUseCmd)
	cloudProfileCmd.AddCommand(cloudProfileDeleteCmd)
	cloudProfileCmd.AddCommand(cloudProfileImportCmd)

	// Add flags
	cloudProfileAddCmd.Flags().StringP("provider", "p", "", "Cloud provider (aws, azure, gcp) (required)")
	cloudProfileAddCmd.Flags().String("region", "", "Default region for the profile")
	cloudProfileAddCmd.Flags().String("account-id", "", "Account/Subscription ID")
	cloudProfileAddCmd.Flags().String("role-arn", "", "IAM Role ARN (AWS only)")
	cloudProfileAddCmd.Flags().Bool("default", false, "Set as default profile")

	// AWS-specific flags
	cloudProfileAddCmd.Flags().String("access-key-id", "", "AWS Access Key ID")
	cloudProfileAddCmd.Flags().String("secret-access-key", "", "AWS Secret Access Key")
	cloudProfileAddCmd.Flags().String("session-token", "", "AWS Session Token (optional)")

	// Azure-specific flags
	cloudProfileAddCmd.Flags().String("tenant-id", "", "Azure Tenant ID")
	cloudProfileAddCmd.Flags().String("client-id", "", "Azure Client ID")
	cloudProfileAddCmd.Flags().String("client-secret", "", "Azure Client Secret")
	cloudProfileAddCmd.Flags().String("subscription-id", "", "Azure Subscription ID")

	// GCP-specific flags
	cloudProfileAddCmd.Flags().String("key-file", "", "Path to GCP service account JSON file")

	_ = cloudProfileAddCmd.MarkFlagRequired("provider")

	// List/Show/Use/Delete flags
	cloudProfileListCmd.Flags().StringP("provider", "p", "", "Filter by provider (aws, azure, gcp)")
	cloudProfileShowCmd.Flags().StringP("provider", "p", "", "Provider (defaults to aws)")
	cloudProfileUseCmd.Flags().StringP("provider", "p", "", "Provider (defaults to aws)")
	cloudProfileDeleteCmd.Flags().StringP("provider", "p", "", "Provider (required)")
	cloudProfileDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")

	_ = cloudProfileDeleteCmd.MarkFlagRequired("provider")

	// Import flags
	cloudProfileImportCmd.Flags().StringP("provider", "p", "", "Cloud provider (aws, azure, gcp) (required)")
	cloudProfileImportCmd.Flags().StringP("source", "s", "", "Source profile/file to import from")
	cloudProfileImportCmd.Flags().Bool("list", false, "List available profiles/credentials to import")
	cloudProfileImportCmd.Flags().Bool("default", false, "Set as default profile after import")

	_ = cloudProfileImportCmd.MarkFlagRequired("provider")
}

func runCloudProfileAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	providerStr, _ := cmd.Flags().GetString("provider")

	if !profile.IsValidProvider(providerStr) {
		return fmt.Errorf("invalid provider: %s (use aws, azure, or gcp)", providerStr)
	}

	provider := profile.Provider(providerStr)

	// Create profile
	p := &profile.CloudProfile{
		Name:     name,
		Provider: provider,
	}

	// Set optional fields
	if region, _ := cmd.Flags().GetString("region"); region != "" {
		p.Region = region
	}

	if accountID, _ := cmd.Flags().GetString("account-id"); accountID != "" {
		p.AccountID = accountID
	}

	if roleArn, _ := cmd.Flags().GetString("role-arn"); roleArn != "" {
		p.RoleArn = roleArn
	}

	if isDefault, _ := cmd.Flags().GetBool("default"); isDefault {
		p.Default = true
	}

	// Get credentials based on provider
	var (
		creds profile.Credentials
		err   error
	)

	switch provider {
	case profile.ProviderAWS:
		creds, err = getAWSCredentials(cmd)
	case profile.ProviderAzure:
		creds, err = getAzureCredentials(cmd)
	case profile.ProviderGCP:
		creds, err = getGCPCredentials(cmd)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	if err != nil {
		return err
	}

	// Initialize service
	svc, err := profile.NewService()
	if err != nil {
		return fmt.Errorf("initializing profile service: %w", err)
	}

	// Add profile
	if err := svc.AddProfile(p, creds); err != nil {
		return fmt.Errorf("adding profile: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Profile '%s' added successfully.\n", name)
	if p.Default {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Set as default profile for %s.\n", provider)
	}

	return nil
}

func getAWSCredentials(cmd *cobra.Command) (*profile.AWSCredentials, error) {
	accessKeyID, _ := cmd.Flags().GetString("access-key-id")
	secretAccessKey, _ := cmd.Flags().GetString("secret-access-key")
	sessionToken, _ := cmd.Flags().GetString("session-token")

	// Prompt for missing values
	if accessKeyID == "" {
		var err error

		accessKeyID, err = promptInput("AWS Access Key ID: ")
		if err != nil {
			return nil, err
		}
	}

	if secretAccessKey == "" {
		var err error

		secretAccessKey, err = promptSecret("AWS Secret Access Key: ")
		if err != nil {
			return nil, err
		}
	}

	if accessKeyID == "" || secretAccessKey == "" {
		return nil, fmt.Errorf("access key ID and secret access key are required")
	}

	return &profile.AWSCredentials{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
	}, nil
}

func getAzureCredentials(cmd *cobra.Command) (*profile.AzureCredentials, error) {
	tenantID, _ := cmd.Flags().GetString("tenant-id")
	clientID, _ := cmd.Flags().GetString("client-id")
	clientSecret, _ := cmd.Flags().GetString("client-secret")
	subscriptionID, _ := cmd.Flags().GetString("subscription-id")

	// Prompt for missing values
	if tenantID == "" {
		var err error

		tenantID, err = promptInput("Azure Tenant ID: ")
		if err != nil {
			return nil, err
		}
	}

	if clientID == "" {
		var err error

		clientID, err = promptInput("Azure Client ID: ")
		if err != nil {
			return nil, err
		}
	}

	if clientSecret == "" {
		var err error

		clientSecret, err = promptSecret("Azure Client Secret: ")
		if err != nil {
			return nil, err
		}
	}

	if subscriptionID == "" {
		var err error

		subscriptionID, err = promptInput("Azure Subscription ID: ")
		if err != nil {
			return nil, err
		}
	}

	if tenantID == "" || clientID == "" || clientSecret == "" || subscriptionID == "" {
		return nil, fmt.Errorf("tenant ID, client ID, client secret, and subscription ID are required")
	}

	return &profile.AzureCredentials{
		TenantID:       tenantID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		SubscriptionID: subscriptionID,
	}, nil
}

func getGCPCredentials(cmd *cobra.Command) (*profile.GCPCredentials, error) {
	keyFile, _ := cmd.Flags().GetString("key-file")

	if keyFile == "" {
		var err error

		keyFile, err = promptInput("Path to service account JSON file: ")
		if err != nil {
			return nil, err
		}
	}

	if keyFile == "" {
		return nil, fmt.Errorf("key-file is required for GCP profiles")
	}

	// Read and parse the service account JSON
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("reading key file: %w", err)
	}

	var creds profile.GCPCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parsing key file: %w", err)
	}

	if creds.Type != "service_account" {
		return nil, fmt.Errorf("key file must be a service account key (got type: %s)", creds.Type)
	}

	return &creds, nil
}

func runCloudProfileList(cmd *cobra.Command, args []string) error {
	providerStr, _ := cmd.Flags().GetString("provider")

	svc, err := profile.NewService()
	if err != nil {
		return fmt.Errorf("initializing profile service: %w", err)
	}

	out := cmd.OutOrStdout()

	if providerStr != "" {
		if !profile.IsValidProvider(providerStr) {
			return fmt.Errorf("invalid provider: %s", providerStr)
		}

		provider := profile.Provider(providerStr)

		profiles, err := svc.ListProfiles(provider)
		if err != nil {
			return err
		}

		printProfileTable(out, provider, profiles)

		return nil
	}

	// List all providers
	allProfiles, err := svc.ListAllProviderProfiles()
	if err != nil {
		return err
	}

	if len(allProfiles) == 0 {
		_, _ = fmt.Fprintln(out, "No profiles configured.")
		return nil
	}

	for _, provider := range profile.ValidProviders {
		profiles := allProfiles[provider]
		if len(profiles) > 0 {
			_, _ = fmt.Fprintf(out, "\n%s:\n", strings.ToUpper(string(provider)))
			printProfileTable(out, provider, profiles)
		}
	}

	return nil
}

func printProfileTable(out io.Writer, provider profile.Provider, profiles []*profile.CloudProfile) {
	if len(profiles) == 0 {
		_, _ = fmt.Fprintln(out, "  No profiles configured.")
		return
	}

	// Header
	_, _ = fmt.Fprintf(out, "  %-20s %-12s %-8s\n", "NAME", "REGION", "DEFAULT")
	_, _ = fmt.Fprintf(out, "  %-20s %-12s %-8s\n", "----", "------", "-------")

	for _, p := range profiles {
		defaultMark := ""
		if p.Default {
			defaultMark = "*"
		}

		region := p.Region
		if region == "" {
			region = "-"
		}

		_, _ = fmt.Fprintf(out, "  %-20s %-12s %-8s\n", p.Name, region, defaultMark)
	}
}

func runCloudProfileShow(cmd *cobra.Command, args []string) error {
	name := args[0]
	providerStr, _ := cmd.Flags().GetString("provider")

	if providerStr == "" {
		providerStr = "aws" // Default to AWS
	}

	if !profile.IsValidProvider(providerStr) {
		return fmt.Errorf("invalid provider: %s", providerStr)
	}

	provider := profile.Provider(providerStr)

	svc, err := profile.NewService()
	if err != nil {
		return fmt.Errorf("initializing profile service: %w", err)
	}

	p, err := svc.GetProfile(provider, name)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "Name:          %s\n", p.Name)
	_, _ = fmt.Fprintf(out, "Provider:      %s\n", p.Provider)
	_, _ = fmt.Fprintf(out, "Region:        %s\n", valueOrDash(p.Region))
	_, _ = fmt.Fprintf(out, "Account ID:    %s\n", valueOrDash(p.AccountID))
	_, _ = fmt.Fprintf(out, "Role ARN:      %s\n", valueOrDash(p.RoleArn))
	_, _ = fmt.Fprintf(out, "Default:       %v\n", p.Default)
	_, _ = fmt.Fprintf(out, "Token Storage: %s\n", p.TokenStorage)

	_, _ = fmt.Fprintf(out, "Created:       %s\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
	if !p.LastUsedAt.IsZero() {
		_, _ = fmt.Fprintf(out, "Last Used:     %s\n", p.LastUsedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func runCloudProfileUse(cmd *cobra.Command, args []string) error {
	name := args[0]
	providerStr, _ := cmd.Flags().GetString("provider")

	if providerStr == "" {
		providerStr = "aws" // Default to AWS
	}

	if !profile.IsValidProvider(providerStr) {
		return fmt.Errorf("invalid provider: %s", providerStr)
	}

	provider := profile.Provider(providerStr)

	svc, err := profile.NewService()
	if err != nil {
		return fmt.Errorf("initializing profile service: %w", err)
	}

	if err := svc.SetDefault(provider, name); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Profile '%s' is now the default for %s.\n", name, provider)

	return nil
}

func runCloudProfileDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	providerStr, _ := cmd.Flags().GetString("provider")
	force, _ := cmd.Flags().GetBool("force")

	if !profile.IsValidProvider(providerStr) {
		return fmt.Errorf("invalid provider: %s", providerStr)
	}

	provider := profile.Provider(providerStr)

	if !force {
		confirm, err := promptInput(fmt.Sprintf("Delete profile '%s/%s'? (yes/no): ", provider, name))
		if err != nil {
			return err
		}

		if strings.ToLower(strings.TrimSpace(confirm)) != "yes" {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
			return nil
		}
	}

	svc, err := profile.NewService()
	if err != nil {
		return fmt.Errorf("initializing profile service: %w", err)
	}

	if err := svc.DeleteProfile(provider, name); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Profile '%s' deleted.\n", name)

	return nil
}

func promptInput(prompt string) (string, error) {
	_, _ = fmt.Fprint(os.Stderr, prompt)
	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

func promptSecret(prompt string) (string, error) {
	_, _ = fmt.Fprint(os.Stderr, prompt)

	// Check if stdin is a terminal
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		password, err := term.ReadPassword(fd)
		_, _ = fmt.Fprintln(os.Stderr) // New line after password

		if err != nil {
			return "", err
		}

		return string(password), nil
	}

	// Fall back to regular input if not a terminal
	return promptInput("")
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}

	return s
}

func runCloudProfileImport(cmd *cobra.Command, args []string) error {
	providerStr, _ := cmd.Flags().GetString("provider")
	source, _ := cmd.Flags().GetString("source")
	listOnly, _ := cmd.Flags().GetBool("list")
	setDefault, _ := cmd.Flags().GetBool("default")

	if !profile.IsValidProvider(providerStr) {
		return fmt.Errorf("invalid provider: %s (use aws, azure, or gcp)", providerStr)
	}

	provider := profile.Provider(providerStr)
	out := cmd.OutOrStdout()

	// Determine target name
	targetName := ""
	if len(args) > 0 {
		targetName = args[0]
	}

	switch provider {
	case profile.ProviderAWS:
		return importAWSProfile(out, source, targetName, listOnly, setDefault)
	case profile.ProviderAzure:
		return importAzureProfile(out, source, targetName, listOnly, setDefault)
	case profile.ProviderGCP:
		return importGCPProfile(out, source, targetName, listOnly, setDefault)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

func importAWSProfile(out io.Writer, source, targetName string, listOnly, setDefault bool) error {
	importer, err := profile.NewAWSImporter()
	if err != nil {
		return err
	}

	if listOnly {
		profiles, err := importer.ListProfiles()
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(out, "Available AWS profiles:")
		for _, p := range profiles {
			_, _ = fmt.Fprintf(out, "  %s\n", p)
		}

		return nil
	}

	opts := profile.ImportOptions{
		SourceProfile: source,
		TargetName:    targetName,
		SetDefault:    setDefault,
	}

	p, creds, err := importer.Import(opts)
	if err != nil {
		return err
	}

	// Save to omni profile
	svc, err := profile.NewService()
	if err != nil {
		return fmt.Errorf("initializing profile service: %w", err)
	}

	if err := svc.AddProfile(p, creds); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	sourceDesc := source
	if sourceDesc == "" {
		sourceDesc = "default"
	}

	_, _ = fmt.Fprintf(out, "Imported AWS profile '%s' from '%s'\n", p.Name, sourceDesc)
	if p.Region != "" {
		_, _ = fmt.Fprintf(out, "  Region: %s\n", p.Region)
	}

	if p.Default {
		_, _ = fmt.Fprintf(out, "  Set as default for aws\n")
	}

	return nil
}

func importAzureProfile(out io.Writer, source, targetName string, listOnly, setDefault bool) error {
	importer, err := profile.NewAzureImporter()
	if err != nil {
		return err
	}

	if listOnly {
		subs, err := importer.ListSubscriptions()
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(out, "Available Azure subscriptions:")
		_, _ = fmt.Fprintf(out, "  %-36s  %-30s  %s\n", "SUBSCRIPTION ID", "NAME", "DEFAULT")
		_, _ = fmt.Fprintf(out, "  %-36s  %-30s  %s\n", "---------------", "----", "-------")

		for _, s := range subs {
			def := ""
			if s.IsDefault {
				def = "*"
			}

			name := s.Name
			if len(name) > 30 {
				name = name[:27] + "..."
			}

			_, _ = fmt.Fprintf(out, "  %-36s  %-30s  %s\n", s.ID, name, def)
		}

		_, _ = fmt.Fprintln(out, "\nNote: To import, create a service principal:")
		_, _ = fmt.Fprintln(out, "  az ad sp create-for-rbac --name omni-sp --sdk-auth > ~/.azure/sp.json")
		_, _ = fmt.Fprintln(out, "  omni cloud profile import --provider azure --source ~/.azure/sp.json")

		return nil
	}

	opts := profile.ImportOptions{
		SourceProfile: source,
		TargetName:    targetName,
		SetDefault:    setDefault,
	}

	p, creds, err := importer.Import(opts)
	if err != nil {
		return err
	}

	svc, err := profile.NewService()
	if err != nil {
		return fmt.Errorf("initializing profile service: %w", err)
	}

	if err := svc.AddProfile(p, creds); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Imported Azure profile '%s'\n", p.Name)
	_, _ = fmt.Fprintf(out, "  Subscription: %s\n", creds.SubscriptionID)

	_, _ = fmt.Fprintf(out, "  Tenant: %s\n", creds.TenantID)
	if p.Default {
		_, _ = fmt.Fprintf(out, "  Set as default for azure\n")
	}

	return nil
}

func importGCPProfile(out io.Writer, source, targetName string, listOnly, setDefault bool) error {
	importer, err := profile.NewGCPImporter()
	if err != nil {
		return err
	}

	if listOnly {
		sources := importer.ListSources()
		if len(sources) == 0 {
			_, _ = fmt.Fprintln(out, "No GCP credentials found.")
			_, _ = fmt.Fprintln(out, "\nTo import, specify a service account key file:")
			_, _ = fmt.Fprintln(out, "  omni cloud profile import --provider gcp --source /path/to/sa.json")

			return nil
		}

		_, _ = fmt.Fprintln(out, "Available GCP credential sources:")
		for _, s := range sources {
			_, _ = fmt.Fprintf(out, "  %s\n", s)
		}

		return nil
	}

	opts := profile.ImportOptions{
		SourceProfile: source,
		TargetName:    targetName,
		SetDefault:    setDefault,
	}

	p, creds, err := importer.Import(opts)
	if err != nil {
		return err
	}

	svc, err := profile.NewService()
	if err != nil {
		return fmt.Errorf("initializing profile service: %w", err)
	}

	if err := svc.AddProfile(p, creds); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Imported GCP profile '%s'\n", p.Name)
	_, _ = fmt.Fprintf(out, "  Project: %s\n", creds.ProjectID)

	_, _ = fmt.Fprintf(out, "  Service Account: %s\n", creds.ClientEmail)
	if p.Default {
		_, _ = fmt.Fprintf(out, "  Set as default for gcp\n")
	}

	return nil
}
