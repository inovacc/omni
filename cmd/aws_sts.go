package cmd

import (
	"context"

	awscommon "github.com/inovacc/omni/internal/cli/aws"
	"github.com/inovacc/omni/internal/cli/aws/sts"
	"github.com/spf13/cobra"
)

var stsCmd = &cobra.Command{
	Use:   "sts",
	Short: "AWS STS operations",
	Long:  `AWS Security Token Service (STS) operations.`,
}

var stsGetCallerIdentityCmd = &cobra.Command{
	Use:   "get-caller-identity",
	Short: "Get details about the IAM identity calling the API",
	Long: `Returns details about the IAM user or role whose credentials are used to call the operation.

Examples:
  omni aws sts get-caller-identity`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")
		endpointURL, _ := cmd.Flags().GetString("endpoint-url")

		opts := awscommon.Options{
			Profile:     profile,
			Region:      region,
			Output:      output,
			EndpointURL: endpointURL,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := sts.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output), awscommon.GetEndpointURL(opts))

		identity, err := client.GetCallerIdentity(ctx)
		if err != nil {
			return err
		}

		return client.PrintCallerIdentity(identity)
	},
}

var stsAssumeRoleCmd = &cobra.Command{
	Use:   "assume-role",
	Short: "Assume an IAM role",
	Long: `Returns a set of temporary security credentials that you can use to access AWS resources.

Examples:
  omni aws sts assume-role --role-arn arn:aws:iam::123456789012:role/MyRole --role-session-name MySession`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")
		endpointURL, _ := cmd.Flags().GetString("endpoint-url")

		roleArn, _ := cmd.Flags().GetString("role-arn")
		roleSessionName, _ := cmd.Flags().GetString("role-session-name")
		durationSeconds, _ := cmd.Flags().GetInt32("duration-seconds")
		externalId, _ := cmd.Flags().GetString("external-id")

		opts := awscommon.Options{
			Profile:     profile,
			Region:      region,
			Output:      output,
			EndpointURL: endpointURL,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := sts.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output), awscommon.GetEndpointURL(opts))

		result, err := client.AssumeRole(ctx, sts.AssumeRoleInput{
			RoleArn:         roleArn,
			RoleSessionName: roleSessionName,
			DurationSeconds: durationSeconds,
			ExternalId:      externalId,
		})
		if err != nil {
			return err
		}

		return client.PrintAssumeRole(result)
	},
}

func init() {
	awsCmd.AddCommand(stsCmd)
	stsCmd.AddCommand(stsGetCallerIdentityCmd)
	stsCmd.AddCommand(stsAssumeRoleCmd)

	// assume-role flags
	stsAssumeRoleCmd.Flags().String("role-arn", "", "ARN of the role to assume (required)")
	stsAssumeRoleCmd.Flags().String("role-session-name", "", "session name (required)")
	stsAssumeRoleCmd.Flags().Int32("duration-seconds", 3600, "duration of the session in seconds")
	stsAssumeRoleCmd.Flags().String("external-id", "", "external ID for cross-account access")
	_ = stsAssumeRoleCmd.MarkFlagRequired("role-arn")
	_ = stsAssumeRoleCmd.MarkFlagRequired("role-session-name")
}
