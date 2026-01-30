package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/sqlite"
	"github.com/inovacc/omni/internal/logger"
	"github.com/spf13/cobra"
)

var sqliteCmd = &cobra.Command{
	Use:   "sqlite",
	Short: "SQLite database management",
	Long: `sqlite provides commands for working with SQLite databases.

This is a CLI wrapper around modernc.org/sqlite (pure Go, no CGO)
for database inspection, querying, and maintenance operations.

Subcommands:
  stats      Show database statistics
  tables     List all tables
  schema     Show table schema
  columns    Show table columns
  indexes    List all indexes
  query      Execute SQL query
  vacuum     Optimize database
  check      Verify database integrity
  dump       Export database as SQL
  import     Import SQL file

Examples:
  omni sqlite stats mydb.sqlite
  omni sqlite tables mydb.sqlite
  omni sqlite schema mydb.sqlite users
  omni sqlite query mydb.sqlite "SELECT * FROM users"
  omni sqlite dump mydb.sqlite > backup.sql
  omni sqlite import mydb.sqlite backup.sql`,
}

var sqliteStatsCmd = &cobra.Command{
	Use:   "stats <database>",
	Short: "Display database statistics",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return sqlite.RunStats(cmd.OutOrStdout(), args[0], sqlite.Options{JSON: jsonOutput})
	},
}

var sqliteTablesCmd = &cobra.Command{
	Use:   "tables <database>",
	Short: "List all tables in the database",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return sqlite.RunTables(cmd.OutOrStdout(), args[0], sqlite.Options{JSON: jsonOutput})
	},
}

var sqliteSchemaCmd = &cobra.Command{
	Use:   "schema <database> [table]",
	Short: "Show table schema",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		table := ""
		if len(args) > 1 {
			table = args[1]
		}

		return sqlite.RunSchema(cmd.OutOrStdout(), args[0], table, sqlite.Options{JSON: jsonOutput})
	},
}

var sqliteColumnsCmd = &cobra.Command{
	Use:   "columns <database> <table>",
	Short: "Show table columns",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return sqlite.RunColumns(cmd.OutOrStdout(), args[0], args[1], sqlite.Options{JSON: jsonOutput})
	},
}

var sqliteIndexesCmd = &cobra.Command{
	Use:   "indexes <database>",
	Short: "List all indexes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return sqlite.RunIndexes(cmd.OutOrStdout(), args[0], sqlite.Options{JSON: jsonOutput})
	},
}

var sqliteQueryCmd = &cobra.Command{
	Use:   "query <database> <sql>",
	Short: "Execute SQL query",
	Long: `Execute SQL query against a SQLite database.

Query logging can be enabled with the omni logger command:
  eval "$(omni logger --path /path/to/logs)"

With logging enabled, queries and results are recorded to log files.
Use --log-data to include result data in logs (use with caution for large results).

Examples:
  omni sqlite query mydb.sqlite "SELECT * FROM users"
  omni sqlite query mydb.sqlite "SELECT * FROM users" --header
  omni sqlite query mydb.sqlite "SELECT * FROM users" --json
  omni sqlite query mydb.sqlite "SELECT * FROM users" --log-data`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		header, _ := cmd.Flags().GetBool("header")
		separator, _ := cmd.Flags().GetString("separator")
		logData, _ := cmd.Flags().GetBool("log-data")

		// Get the global logger if available
		l := logger.Get()

		return sqlite.RunQuery(cmd.OutOrStdout(), args[0], args[1], sqlite.Options{
			JSON:      jsonOutput,
			Header:    header,
			Separator: separator,
			Logger:    l,
			LogData:   logData,
		})
	},
}

var sqliteVacuumCmd = &cobra.Command{
	Use:   "vacuum <database>",
	Short: "Optimize database",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return sqlite.RunVacuum(cmd.OutOrStdout(), args[0], sqlite.Options{JSON: jsonOutput})
	},
}

var sqliteCheckCmd = &cobra.Command{
	Use:   "check <database>",
	Short: "Verify database integrity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return sqlite.RunCheck(cmd.OutOrStdout(), args[0], sqlite.Options{JSON: jsonOutput})
	},
}

var sqliteDumpCmd = &cobra.Command{
	Use:   "dump <database> [table]",
	Short: "Export database as SQL",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		table := ""
		if len(args) > 1 {
			table = args[1]
		}

		return sqlite.RunDump(cmd.OutOrStdout(), args[0], table, sqlite.Options{})
	},
}

var sqliteImportCmd = &cobra.Command{
	Use:   "import <database> <sql-file>",
	Short: "Import SQL file into database",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		return sqlite.RunImport(cmd.OutOrStdout(), args[0], args[1], sqlite.Options{JSON: jsonOutput})
	},
}

func init() {
	rootCmd.AddCommand(sqliteCmd)

	// Add subcommands
	sqliteCmd.AddCommand(sqliteStatsCmd)
	sqliteCmd.AddCommand(sqliteTablesCmd)
	sqliteCmd.AddCommand(sqliteSchemaCmd)
	sqliteCmd.AddCommand(sqliteColumnsCmd)
	sqliteCmd.AddCommand(sqliteIndexesCmd)
	sqliteCmd.AddCommand(sqliteQueryCmd)
	sqliteCmd.AddCommand(sqliteVacuumCmd)
	sqliteCmd.AddCommand(sqliteCheckCmd)
	sqliteCmd.AddCommand(sqliteDumpCmd)
	sqliteCmd.AddCommand(sqliteImportCmd)

	// Add persistent flags to parent
	sqliteCmd.PersistentFlags().Bool("json", false, "output as JSON")

	// Add command-specific flags
	sqliteQueryCmd.Flags().BoolP("header", "H", false, "show column headers")
	sqliteQueryCmd.Flags().StringP("separator", "s", "|", "column separator")
	sqliteQueryCmd.Flags().Bool("log-data", false, "include result data in logs (use with caution for large results)")
}
