package cmd

import (
	"fmt"

	"github.com/inovacc/omni/internal/cli/bbolt"
	"github.com/spf13/cobra"
)

var bboltCmd = &cobra.Command{
	Use:   "bbolt",
	Short: "BoltDB database management",
	Long: `bbolt provides commands for working with BoltDB databases.

This is a CLI wrapper around etcd-io/bbolt for database inspection,
manipulation, and maintenance operations.

Subcommands:
  info       Display database page size
  stats      Show database statistics
  buckets    List all buckets
  keys       List keys in a bucket
  get        Get value for a key
  put        Store a key-value pair
  delete     Delete a key
  dump       Dump bucket contents
  compact    Compact database to new file
  check      Verify database integrity
  pages      List database pages
  page       Hex dump of a page
  create-bucket  Create a new bucket
  delete-bucket  Delete a bucket

Examples:
  omni bbolt stats mydb.bolt
  omni bbolt buckets mydb.bolt
  omni bbolt keys mydb.bolt users
  omni bbolt get mydb.bolt users user1
  omni bbolt put mydb.bolt config version 1.0.0
  omni bbolt compact mydb.bolt mydb-compact.bolt`,
}

var bboltInfoCmd = &cobra.Command{
	Use:   "info <database>",
	Short: "Display database information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return bbolt.RunInfo(cmd.OutOrStdout(), args[0], bbolt.Options{JSON: jsonOutput})
	},
}

var bboltStatsCmd = &cobra.Command{
	Use:   "stats <database>",
	Short: "Display database statistics",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return bbolt.RunStats(cmd.OutOrStdout(), args[0], bbolt.Options{JSON: jsonOutput})
	},
}

var bboltBucketsCmd = &cobra.Command{
	Use:   "buckets <database>",
	Short: "List all buckets in the database",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return bbolt.RunBuckets(cmd.OutOrStdout(), args[0], bbolt.Options{JSON: jsonOutput})
	},
}

var bboltKeysCmd = &cobra.Command{
	Use:   "keys <database> <bucket>",
	Short: "List keys in a bucket",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		prefix, _ := cmd.Flags().GetString("prefix")

		return bbolt.RunKeys(cmd.OutOrStdout(), args[0], args[1], bbolt.Options{
			JSON:   jsonOutput,
			Prefix: prefix,
		})
	},
}

var bboltGetCmd = &cobra.Command{
	Use:   "get <database> <bucket> <key>",
	Short: "Get value for a key",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		hexOutput, _ := cmd.Flags().GetBool("hex")

		return bbolt.RunGet(cmd.OutOrStdout(), args[0], args[1], args[2], bbolt.Options{
			JSON: jsonOutput,
			Hex:  hexOutput,
		})
	},
}

var bboltPutCmd = &cobra.Command{
	Use:   "put <database> <bucket> <key> <value>",
	Short: "Store a key-value pair",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return bbolt.RunPut(cmd.OutOrStdout(), args[0], args[1], args[2], args[3], bbolt.Options{JSON: jsonOutput})
	},
}

var bboltDeleteCmd = &cobra.Command{
	Use:   "delete <database> <bucket> <key>",
	Short: "Delete a key from a bucket",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return bbolt.RunDelete(cmd.OutOrStdout(), args[0], args[1], args[2], bbolt.Options{JSON: jsonOutput})
	},
}

var bboltDumpCmd = &cobra.Command{
	Use:   "dump <database> <bucket>",
	Short: "Dump all keys and values in a bucket",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		hexOutput, _ := cmd.Flags().GetBool("hex")
		prefix, _ := cmd.Flags().GetString("prefix")

		return bbolt.RunDump(cmd.OutOrStdout(), args[0], args[1], bbolt.Options{
			JSON:   jsonOutput,
			Hex:    hexOutput,
			Prefix: prefix,
		})
	},
}

var bboltCompactCmd = &cobra.Command{
	Use:   "compact <source> <destination>",
	Short: "Compact database to a new file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return bbolt.RunCompact(cmd.OutOrStdout(), args[0], args[1], bbolt.Options{JSON: jsonOutput})
	},
}

var bboltCheckCmd = &cobra.Command{
	Use:   "check <database>",
	Short: "Verify database integrity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return bbolt.RunCheck(cmd.OutOrStdout(), args[0], bbolt.Options{JSON: jsonOutput})
	},
}

var bboltPagesCmd = &cobra.Command{
	Use:   "pages <database>",
	Short: "List database pages",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return bbolt.RunPages(cmd.OutOrStdout(), args[0], bbolt.Options{JSON: jsonOutput})
	},
}

var bboltPageCmd = &cobra.Command{
	Use:   "page <database> <page-id>",
	Short: "Hex dump of a specific page",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		var pageID int
		if _, err := fmt.Sscanf(args[1], "%d", &pageID); err != nil {
			return fmt.Errorf("invalid page ID: %s", args[1])
		}

		return bbolt.RunPageDump(cmd.OutOrStdout(), args[0], pageID, bbolt.Options{JSON: jsonOutput})
	},
}

var bboltCreateBucketCmd = &cobra.Command{
	Use:   "create-bucket <database> <bucket>",
	Short: "Create a new bucket",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return bbolt.RunCreateBucket(cmd.OutOrStdout(), args[0], args[1], bbolt.Options{JSON: jsonOutput})
	},
}

var bboltDeleteBucketCmd = &cobra.Command{
	Use:   "delete-bucket <database> <bucket>",
	Short: "Delete a bucket",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return bbolt.RunDeleteBucket(cmd.OutOrStdout(), args[0], args[1], bbolt.Options{JSON: jsonOutput})
	},
}

func init() {
	rootCmd.AddCommand(bboltCmd)

	// Add subcommands
	bboltCmd.AddCommand(bboltInfoCmd)
	bboltCmd.AddCommand(bboltStatsCmd)
	bboltCmd.AddCommand(bboltBucketsCmd)
	bboltCmd.AddCommand(bboltKeysCmd)
	bboltCmd.AddCommand(bboltGetCmd)
	bboltCmd.AddCommand(bboltPutCmd)
	bboltCmd.AddCommand(bboltDeleteCmd)
	bboltCmd.AddCommand(bboltDumpCmd)
	bboltCmd.AddCommand(bboltCompactCmd)
	bboltCmd.AddCommand(bboltCheckCmd)
	bboltCmd.AddCommand(bboltPagesCmd)
	bboltCmd.AddCommand(bboltPageCmd)
	bboltCmd.AddCommand(bboltCreateBucketCmd)
	bboltCmd.AddCommand(bboltDeleteBucketCmd)

	// Add persistent flags to parent
	bboltCmd.PersistentFlags().Bool("json", false, "output as JSON")

	// Add command-specific flags
	bboltKeysCmd.Flags().String("prefix", "", "filter keys by prefix")
	bboltGetCmd.Flags().Bool("hex", false, "display value in hexadecimal")
	bboltDumpCmd.Flags().Bool("hex", false, "display values in hexadecimal")
	bboltDumpCmd.Flags().String("prefix", "", "filter keys by prefix")
}
