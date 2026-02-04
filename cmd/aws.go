package cmd

import (
	"github.com/spf13/cobra"
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "AWS CLI operations",
	Long: `AWS CLI operations for core services.

Supported services:
  s3    - S3 bucket and object operations
  ec2   - EC2 instance operations
  iam   - IAM user, role, and policy operations
  sts   - STS identity and credential operations
  ssm   - SSM Parameter Store operations

Configuration:
  AWS credentials are loaded from standard AWS SDK sources:
  - Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
  - Shared credentials file (~/.aws/credentials)
  - IAM role (when running on EC2)

Examples:
  # Get caller identity
  omni aws sts get-caller-identity

  # List S3 buckets
  omni aws s3 ls

  # List S3 objects
  omni aws s3 ls s3://my-bucket/prefix/

  # Describe EC2 instances
  omni aws ec2 describe-instances

  # Get SSM parameter
  omni aws ssm get-parameter --name /app/config`,
}

func init() {
	rootCmd.AddCommand(awsCmd)

	// Global AWS flags
	awsCmd.PersistentFlags().String("profile", "", "AWS profile to use")
	awsCmd.PersistentFlags().String("region", "", "AWS region")
	awsCmd.PersistentFlags().String("output", "json", "output format: json, text, table")
}
