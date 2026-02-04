package cmd

import (
	"github.com/spf13/cobra"
)

var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Cloud profile management",
	Long: `Cloud profile management for AWS, Azure, and GCP.

Securely store and manage credentials for multiple cloud providers.
Credentials are encrypted using AES-256-GCM with per-profile key isolation.

Supported providers:
  aws    - Amazon Web Services
  azure  - Microsoft Azure
  gcp    - Google Cloud Platform

Examples:
  # Add an AWS profile
  omni cloud profile add myaws --provider aws --region us-east-1

  # List all profiles
  omni cloud profile list

  # Use a profile with AWS commands
  export OMNI_CLOUD_PROFILE=myaws
  omni aws s3 ls

  # Or use the omni: prefix
  omni aws s3 ls --profile omni:myaws`,
}

func init() {
	rootCmd.AddCommand(cloudCmd)
}
