package cmd

import (
	"context"
	"strings"

	awscommon "github.com/inovacc/omni/internal/cli/aws"
	"github.com/inovacc/omni/internal/cli/aws/ec2"
	"github.com/spf13/cobra"
)

var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "AWS EC2 operations",
	Long:  `AWS EC2 instance and resource operations.`,
}

var ec2DescribeInstancesCmd = &cobra.Command{
	Use:   "describe-instances",
	Short: "Describe EC2 instances",
	Long: `Describes one or more EC2 instances.

Examples:
  omni aws ec2 describe-instances
  omni aws ec2 describe-instances --instance-ids i-1234567890abcdef0
  omni aws ec2 describe-instances --filters "Name=tag:Name,Values=prod-*"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		instanceIds, _ := cmd.Flags().GetStringSlice("instance-ids")
		filters, _ := cmd.Flags().GetStringSlice("filters")
		maxResults, _ := cmd.Flags().GetInt32("max-results")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := ec2.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		instances, err := client.DescribeInstances(ctx, ec2.DescribeInstancesInput{
			InstanceIds: instanceIds,
			Filters:     parseEC2Filters(filters),
			MaxResults:  maxResults,
		})
		if err != nil {
			return err
		}

		return client.PrintInstances(instances)
	},
}

var ec2StartInstancesCmd = &cobra.Command{
	Use:   "start-instances",
	Short: "Start EC2 instances",
	Long: `Starts one or more stopped instances.

Examples:
  omni aws ec2 start-instances --instance-ids i-1234567890abcdef0
  omni aws ec2 start-instances --instance-ids i-1234,i-5678`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		instanceIds, _ := cmd.Flags().GetStringSlice("instance-ids")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := ec2.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		changes, err := client.StartInstances(ctx, instanceIds)
		if err != nil {
			return err
		}

		return client.PrintStateChanges(changes)
	},
}

var ec2StopInstancesCmd = &cobra.Command{
	Use:   "stop-instances",
	Short: "Stop EC2 instances",
	Long: `Stops one or more running instances.

Examples:
  omni aws ec2 stop-instances --instance-ids i-1234567890abcdef0
  omni aws ec2 stop-instances --instance-ids i-1234 --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		instanceIds, _ := cmd.Flags().GetStringSlice("instance-ids")
		force, _ := cmd.Flags().GetBool("force")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := ec2.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		changes, err := client.StopInstances(ctx, instanceIds, force)
		if err != nil {
			return err
		}

		return client.PrintStateChanges(changes)
	},
}

var ec2DescribeVpcsCmd = &cobra.Command{
	Use:   "describe-vpcs",
	Short: "Describe VPCs",
	Long: `Describes one or more VPCs.

Examples:
  omni aws ec2 describe-vpcs
  omni aws ec2 describe-vpcs --vpc-ids vpc-1234567890abcdef0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		vpcIds, _ := cmd.Flags().GetStringSlice("vpc-ids")
		filters, _ := cmd.Flags().GetStringSlice("filters")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := ec2.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		vpcs, err := client.DescribeVpcs(ctx, vpcIds, parseEC2Filters(filters))
		if err != nil {
			return err
		}

		return client.PrintVpcs(vpcs)
	},
}

var ec2DescribeSecurityGroupsCmd = &cobra.Command{
	Use:   "describe-security-groups",
	Short: "Describe security groups",
	Long: `Describes one or more security groups.

Examples:
  omni aws ec2 describe-security-groups
  omni aws ec2 describe-security-groups --group-ids sg-1234567890abcdef0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		groupIds, _ := cmd.Flags().GetStringSlice("group-ids")
		filters, _ := cmd.Flags().GetStringSlice("filters")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := ec2.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		groups, err := client.DescribeSecurityGroups(ctx, groupIds, parseEC2Filters(filters))
		if err != nil {
			return err
		}

		return client.PrintSecurityGroups(groups)
	},
}

func init() {
	awsCmd.AddCommand(ec2Cmd)
	ec2Cmd.AddCommand(ec2DescribeInstancesCmd)
	ec2Cmd.AddCommand(ec2StartInstancesCmd)
	ec2Cmd.AddCommand(ec2StopInstancesCmd)
	ec2Cmd.AddCommand(ec2DescribeVpcsCmd)
	ec2Cmd.AddCommand(ec2DescribeSecurityGroupsCmd)

	// describe-instances flags
	ec2DescribeInstancesCmd.Flags().StringSlice("instance-ids", nil, "instance IDs")
	ec2DescribeInstancesCmd.Flags().StringSlice("filters", nil, "filters in format 'Name=value,Values=v1,v2'")
	ec2DescribeInstancesCmd.Flags().Int32("max-results", 0, "maximum number of results")

	// start-instances flags
	ec2StartInstancesCmd.Flags().StringSlice("instance-ids", nil, "instance IDs (required)")
	_ = ec2StartInstancesCmd.MarkFlagRequired("instance-ids")

	// stop-instances flags
	ec2StopInstancesCmd.Flags().StringSlice("instance-ids", nil, "instance IDs (required)")
	ec2StopInstancesCmd.Flags().Bool("force", false, "force stop without graceful shutdown")
	_ = ec2StopInstancesCmd.MarkFlagRequired("instance-ids")

	// describe-vpcs flags
	ec2DescribeVpcsCmd.Flags().StringSlice("vpc-ids", nil, "VPC IDs")
	ec2DescribeVpcsCmd.Flags().StringSlice("filters", nil, "filters")

	// describe-security-groups flags
	ec2DescribeSecurityGroupsCmd.Flags().StringSlice("group-ids", nil, "security group IDs")
	ec2DescribeSecurityGroupsCmd.Flags().StringSlice("filters", nil, "filters")
}

// parseEC2Filters parses filter strings in AWS CLI format
// Format: "Name=tag:Name,Values=prod-*,staging-*"
func parseEC2Filters(filters []string) []ec2.Filter {
	var result []ec2.Filter

	for _, f := range filters {
		parts := strings.SplitN(f, ",", 2)
		if len(parts) < 2 {
			continue
		}

		var name string
		var values []string

		for _, part := range strings.Split(f, ",") {
			if strings.HasPrefix(part, "Name=") {
				name = strings.TrimPrefix(part, "Name=")
			} else if strings.HasPrefix(part, "Values=") {
				values = strings.Split(strings.TrimPrefix(part, "Values="), ",")
			}
		}

		if name != "" && len(values) > 0 {
			result = append(result, ec2.Filter{
				Name:   name,
				Values: values,
			})
		}
	}

	return result
}
