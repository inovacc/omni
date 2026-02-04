package cmd

import (
	"context"

	awscommon "github.com/inovacc/omni/internal/cli/aws"
	"github.com/inovacc/omni/internal/cli/aws/iam"
	"github.com/spf13/cobra"
)

var iamCmd = &cobra.Command{
	Use:   "iam",
	Short: "AWS IAM operations",
	Long:  `AWS Identity and Access Management (IAM) operations.`,
}

var iamGetUserCmd = &cobra.Command{
	Use:   "get-user",
	Short: "Get IAM user information",
	Long: `Retrieves information about the specified IAM user, including the user's creation date,
path, unique ID, and ARN. If no user name is specified, returns information about
the IAM user whose credentials are used to call the operation.

Examples:
  omni aws iam get-user
  omni aws iam get-user --user-name myuser`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		userName, _ := cmd.Flags().GetString("user-name")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := iam.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		user, err := client.GetUser(ctx, userName)
		if err != nil {
			return err
		}

		return client.PrintUser(user)
	},
}

var iamGetRoleCmd = &cobra.Command{
	Use:   "get-role",
	Short: "Get IAM role information",
	Long: `Retrieves information about the specified role, including the role's path, GUID,
ARN, and the role's trust policy.

Examples:
  omni aws iam get-role --role-name MyRole`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		roleName, _ := cmd.Flags().GetString("role-name")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := iam.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		role, err := client.GetRole(ctx, roleName)
		if err != nil {
			return err
		}

		return client.PrintRole(role)
	},
}

var iamListRolesCmd = &cobra.Command{
	Use:   "list-roles",
	Short: "List IAM roles",
	Long: `Lists the IAM roles that have the specified path prefix.

Examples:
  omni aws iam list-roles
  omni aws iam list-roles --path-prefix /service-role/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		pathPrefix, _ := cmd.Flags().GetString("path-prefix")
		maxItems, _ := cmd.Flags().GetInt32("max-items")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := iam.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		roles, err := client.ListRoles(ctx, iam.ListRolesInput{
			PathPrefix: pathPrefix,
			MaxItems:   maxItems,
		})
		if err != nil {
			return err
		}

		return client.PrintRoles(roles)
	},
}

var iamListPoliciesCmd = &cobra.Command{
	Use:   "list-policies",
	Short: "List IAM policies",
	Long: `Lists all the managed policies that are available in your AWS account.

Examples:
  omni aws iam list-policies
  omni aws iam list-policies --scope Local
  omni aws iam list-policies --only-attached`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		scope, _ := cmd.Flags().GetString("scope")
		onlyAttached, _ := cmd.Flags().GetBool("only-attached")
		pathPrefix, _ := cmd.Flags().GetString("path-prefix")
		maxItems, _ := cmd.Flags().GetInt32("max-items")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := iam.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		policies, err := client.ListPolicies(ctx, iam.ListPoliciesInput{
			Scope:        scope,
			OnlyAttached: onlyAttached,
			PathPrefix:   pathPrefix,
			MaxItems:     maxItems,
		})
		if err != nil {
			return err
		}

		return client.PrintPolicies(policies)
	},
}

var iamGetPolicyCmd = &cobra.Command{
	Use:   "get-policy",
	Short: "Get IAM policy information",
	Long: `Retrieves information about the specified managed policy.

Examples:
  omni aws iam get-policy --policy-arn arn:aws:iam::123456789012:policy/MyPolicy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		policyArn, _ := cmd.Flags().GetString("policy-arn")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := iam.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		policy, err := client.GetPolicy(ctx, policyArn)
		if err != nil {
			return err
		}

		return client.PrintPolicy(policy)
	},
}

func init() {
	awsCmd.AddCommand(iamCmd)
	iamCmd.AddCommand(iamGetUserCmd)
	iamCmd.AddCommand(iamGetRoleCmd)
	iamCmd.AddCommand(iamListRolesCmd)
	iamCmd.AddCommand(iamListPoliciesCmd)
	iamCmd.AddCommand(iamGetPolicyCmd)

	// get-user flags
	iamGetUserCmd.Flags().String("user-name", "", "user name (optional, defaults to current user)")

	// get-role flags
	iamGetRoleCmd.Flags().String("role-name", "", "role name (required)")
	_ = iamGetRoleCmd.MarkFlagRequired("role-name")

	// list-roles flags
	iamListRolesCmd.Flags().String("path-prefix", "", "path prefix filter")
	iamListRolesCmd.Flags().Int32("max-items", 0, "maximum number of items")

	// list-policies flags
	iamListPoliciesCmd.Flags().String("scope", "All", "scope: All, AWS, Local")
	iamListPoliciesCmd.Flags().Bool("only-attached", false, "only show attached policies")
	iamListPoliciesCmd.Flags().String("path-prefix", "", "path prefix filter")
	iamListPoliciesCmd.Flags().Int32("max-items", 0, "maximum number of items")

	// get-policy flags
	iamGetPolicyCmd.Flags().String("policy-arn", "", "policy ARN (required)")
	_ = iamGetPolicyCmd.MarkFlagRequired("policy-arn")
}
