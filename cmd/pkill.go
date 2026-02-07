package cmd

import (
	"github.com/inovacc/omni/internal/cli/pkill"
	"github.com/spf13/cobra"
)

// pkillCmd represents the pkill command
var pkillCmd = &cobra.Command{
	Use:   "pkill [OPTIONS] PATTERN",
	Short: "Kill processes by name or pattern",
	Long: `Send a signal to processes matching a pattern.

By default, sends SIGTERM to matching processes. Use -l to list
matching processes without killing them (pgrep behavior).

Options:
  -signal     signal to send (default: TERM)
  -x          match process name exactly
  -f          match against full command line
  -n          select only the newest matching process
  -o          select only the oldest matching process
  -c          count matching processes
  -l          list matching processes (pgrep mode)
  -u USER     match only processes owned by USER
  -P PID      match only processes with parent PID
  -i          case insensitive matching

Examples:
  omni pkill firefox           # kill all firefox processes
  omni pkill -9 chrome         # send SIGKILL to chrome
  omni pkill -l python         # list python processes (pgrep)
  omni pkill -f "node server"  # match full command line
  omni pkill -n -l java        # show newest java process
  omni pkill -u root httpd     # kill httpd owned by root
  omni pkill -c nginx          # count nginx processes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		opts := pkill.Options{}
		opts.Signal, _ = cmd.Flags().GetString("signal")
		opts.Exact, _ = cmd.Flags().GetBool("exact")
		opts.Full, _ = cmd.Flags().GetBool("full")
		opts.Newest, _ = cmd.Flags().GetBool("newest")
		opts.Oldest, _ = cmd.Flags().GetBool("oldest")
		opts.Count, _ = cmd.Flags().GetBool("count")
		opts.ListOnly, _ = cmd.Flags().GetBool("list")
		opts.User, _ = cmd.Flags().GetString("user")
		opts.Parent, _ = cmd.Flags().GetInt("parent")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.IgnoreCase, _ = cmd.Flags().GetBool("ignore-case")

		return pkill.Run(cmd.OutOrStdout(), args[0], opts)
	},
}

// pgrepCmd is an alias for pkill -l
var pgrepCmd = &cobra.Command{
	Use:   "pgrep [OPTIONS] PATTERN",
	Short: "Find processes by name or pattern",
	Long: `List processes matching a pattern. Alias for pkill -l.

Options:
  -x          match process name exactly
  -f          match against full command line
  -n          select only the newest matching process
  -o          select only the oldest matching process
  -c          count matching processes
  -u USER     match only processes owned by USER
  -P PID      match only processes with parent PID
  -i          case insensitive matching

Examples:
  omni pgrep firefox           # find firefox processes
  omni pgrep -f "node server"  # match full command line
  omni pgrep -c python         # count python processes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		opts := pkill.Options{
			ListOnly: true, // pgrep always lists
		}
		opts.Exact, _ = cmd.Flags().GetBool("exact")
		opts.Full, _ = cmd.Flags().GetBool("full")
		opts.Newest, _ = cmd.Flags().GetBool("newest")
		opts.Oldest, _ = cmd.Flags().GetBool("oldest")
		opts.Count, _ = cmd.Flags().GetBool("count")
		opts.User, _ = cmd.Flags().GetString("user")
		opts.Parent, _ = cmd.Flags().GetInt("parent")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.IgnoreCase, _ = cmd.Flags().GetBool("ignore-case")

		return pkill.Run(cmd.OutOrStdout(), args[0], opts)
	},
}

func init() {
	rootCmd.AddCommand(pkillCmd)
	rootCmd.AddCommand(pgrepCmd)

	// pkill flags
	pkillCmd.Flags().StringP("signal", "s", "", "signal to send (default: TERM)")
	pkillCmd.Flags().BoolP("exact", "x", false, "match exactly")
	pkillCmd.Flags().BoolP("full", "f", false, "match against full command line")
	pkillCmd.Flags().BoolP("newest", "n", false, "select only the newest process")
	pkillCmd.Flags().BoolP("oldest", "o", false, "select only the oldest process")
	pkillCmd.Flags().BoolP("count", "c", false, "count matching processes")
	pkillCmd.Flags().BoolP("list", "l", false, "list matching processes (pgrep mode)")
	pkillCmd.Flags().StringP("user", "u", "", "match only processes owned by user")
	pkillCmd.Flags().IntP("parent", "P", 0, "match only processes with parent PID")
	pkillCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	pkillCmd.Flags().BoolP("json", "j", false, "output as JSON")
	pkillCmd.Flags().BoolP("ignore-case", "i", false, "case insensitive matching")

	// pgrep flags (subset of pkill)
	pgrepCmd.Flags().BoolP("exact", "x", false, "match exactly")
	pgrepCmd.Flags().BoolP("full", "f", false, "match against full command line")
	pgrepCmd.Flags().BoolP("newest", "n", false, "select only the newest process")
	pgrepCmd.Flags().BoolP("oldest", "o", false, "select only the oldest process")
	pgrepCmd.Flags().BoolP("count", "c", false, "count matching processes")
	pgrepCmd.Flags().StringP("user", "u", "", "match only processes owned by user")
	pgrepCmd.Flags().IntP("parent", "P", 0, "match only processes with parent PID")
	pgrepCmd.Flags().BoolP("json", "j", false, "output as JSON")
	pgrepCmd.Flags().BoolP("ignore-case", "i", false, "case insensitive matching")
}
