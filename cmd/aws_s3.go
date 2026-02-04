package cmd

import (
	"context"
	"fmt"
	"time"

	awscommon "github.com/inovacc/omni/internal/cli/aws"
	"github.com/inovacc/omni/internal/cli/aws/s3"
	"github.com/spf13/cobra"
)

var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "AWS S3 operations",
	Long: `AWS S3 bucket and object operations.

Examples:
  # List buckets
  omni aws s3 ls

  # List objects in a bucket
  omni aws s3 ls s3://my-bucket/

  # Copy file to S3
  omni aws s3 cp file.txt s3://my-bucket/file.txt

  # Download file from S3
  omni aws s3 cp s3://my-bucket/file.txt ./file.txt

  # Remove object
  omni aws s3 rm s3://my-bucket/file.txt

  # Create bucket
  omni aws s3 mb s3://my-new-bucket

  # Generate presigned URL
  omni aws s3 presign s3://my-bucket/file.txt`,
}

var s3LsCmd = &cobra.Command{
	Use:   "ls [S3_URI]",
	Short: "List S3 objects or buckets",
	Long: `Lists S3 objects in a bucket or all buckets.

Examples:
  omni aws s3 ls
  omni aws s3 ls s3://my-bucket/
  omni aws s3 ls s3://my-bucket/prefix/ --recursive`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		output, _ := cmd.Flags().GetString("output")

		recursive, _ := cmd.Flags().GetBool("recursive")
		human, _ := cmd.Flags().GetBool("human-readable")
		summarize, _ := cmd.Flags().GetBool("summarize")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
			Output:  output,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := s3.NewClient(cfg, cmd.OutOrStdout(), awscommon.ParseOutputFormat(output))

		uri := ""
		if len(args) > 0 {
			uri = args[0]
		}

		return client.Ls(ctx, uri, s3.LsOptions{
			Recursive: recursive,
			Human:     human,
			Summarize: summarize,
		})
	},
}

var s3CpCmd = &cobra.Command{
	Use:   "cp <SOURCE> <DESTINATION>",
	Short: "Copy files to/from S3",
	Long: `Copies files between local filesystem and S3, or between S3 locations.

Examples:
  # Upload to S3
  omni aws s3 cp file.txt s3://my-bucket/file.txt

  # Download from S3
  omni aws s3 cp s3://my-bucket/file.txt ./file.txt

  # Copy between S3 locations
  omni aws s3 cp s3://bucket1/file.txt s3://bucket2/file.txt

  # Dry run
  omni aws s3 cp file.txt s3://my-bucket/file.txt --dryrun`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")

		recursive, _ := cmd.Flags().GetBool("recursive")
		dryRun, _ := cmd.Flags().GetBool("dryrun")
		quiet, _ := cmd.Flags().GetBool("quiet")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := s3.NewClient(cfg, cmd.OutOrStdout(), awscommon.OutputText)

		return client.Cp(ctx, cmd.OutOrStdout(), args[0], args[1], s3.CpOptions{
			Recursive: recursive,
			DryRun:    dryRun,
			Quiet:     quiet,
		})
	},
}

var s3RmCmd = &cobra.Command{
	Use:   "rm <S3_URI>",
	Short: "Remove S3 objects",
	Long: `Deletes objects from S3.

Examples:
  omni aws s3 rm s3://my-bucket/file.txt
  omni aws s3 rm s3://my-bucket/prefix/ --recursive
  omni aws s3 rm s3://my-bucket/prefix/ --recursive --dryrun`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")

		recursive, _ := cmd.Flags().GetBool("recursive")
		dryRun, _ := cmd.Flags().GetBool("dryrun")
		quiet, _ := cmd.Flags().GetBool("quiet")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := s3.NewClient(cfg, cmd.OutOrStdout(), awscommon.OutputText)

		return client.Rm(ctx, cmd.OutOrStdout(), args[0], recursive, dryRun, quiet)
	},
}

var s3MbCmd = &cobra.Command{
	Use:   "mb <S3_URI>",
	Short: "Create an S3 bucket",
	Long: `Creates an S3 bucket.

Examples:
  omni aws s3 mb s3://my-new-bucket
  omni aws s3 mb s3://my-new-bucket --region us-west-2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := s3.NewClient(cfg, cmd.OutOrStdout(), awscommon.OutputText)

		return client.Mb(ctx, cmd.OutOrStdout(), args[0], awscommon.GetRegion(opts))
	},
}

var s3RbCmd = &cobra.Command{
	Use:   "rb <S3_URI>",
	Short: "Remove an S3 bucket",
	Long: `Deletes an S3 bucket. The bucket must be empty unless --force is specified.

Examples:
  omni aws s3 rb s3://my-bucket
  omni aws s3 rb s3://my-bucket --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")

		force, _ := cmd.Flags().GetBool("force")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := s3.NewClient(cfg, cmd.OutOrStdout(), awscommon.OutputText)

		return client.Rb(ctx, cmd.OutOrStdout(), args[0], force)
	},
}

var s3PresignCmd = &cobra.Command{
	Use:   "presign <S3_URI>",
	Short: "Generate a presigned URL",
	Long: `Generates a presigned URL for an S3 object.

Examples:
  omni aws s3 presign s3://my-bucket/file.txt
  omni aws s3 presign s3://my-bucket/file.txt --expires-in 3600`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")

		expiresIn, _ := cmd.Flags().GetInt("expires-in")

		opts := awscommon.Options{
			Profile: profile,
			Region:  region,
		}

		cfg, err := awscommon.LoadConfig(ctx, opts)
		if err != nil {
			return err
		}

		client := s3.NewClient(cfg, cmd.OutOrStdout(), awscommon.OutputText)

		url, err := client.Presign(ctx, args[0], s3.PresignOptions{
			ExpiresIn: time.Duration(expiresIn) * time.Second,
		})
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout(), url)
		return nil
	},
}

func init() {
	awsCmd.AddCommand(s3Cmd)
	s3Cmd.AddCommand(s3LsCmd)
	s3Cmd.AddCommand(s3CpCmd)
	s3Cmd.AddCommand(s3RmCmd)
	s3Cmd.AddCommand(s3MbCmd)
	s3Cmd.AddCommand(s3RbCmd)
	s3Cmd.AddCommand(s3PresignCmd)

	// ls flags
	s3LsCmd.Flags().Bool("recursive", false, "list recursively")
	s3LsCmd.Flags().Bool("human-readable", false, "display file sizes in human-readable format")
	s3LsCmd.Flags().Bool("summarize", false, "display summary information")

	// cp flags
	s3CpCmd.Flags().Bool("recursive", false, "copy recursively")
	s3CpCmd.Flags().Bool("dryrun", false, "display operations without executing")
	s3CpCmd.Flags().Bool("quiet", false, "suppress output")

	// rm flags
	s3RmCmd.Flags().Bool("recursive", false, "delete recursively")
	s3RmCmd.Flags().Bool("dryrun", false, "display operations without executing")
	s3RmCmd.Flags().Bool("quiet", false, "suppress output")

	// rb flags
	s3RbCmd.Flags().Bool("force", false, "delete all objects before removing bucket")

	// presign flags
	s3PresignCmd.Flags().Int("expires-in", 900, "URL expiration time in seconds (default 15 minutes)")
}
