package cmd

import (
	"github.com/inovacc/omni/internal/cli/note"
	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Quick note taking to JSON in your Documents folder",
	Long: `Save quick notes to a JSON file in your user Documents directory.

By default notes are saved to:
  Windows: %USERPROFILE%\Documents\omni-notes.json
  Linux:   $HOME/Documents/omni-notes.json
  macOS:   $HOME/Documents/omni-notes.json

Examples:
  omni note add "buy milk"
  omni note add deploy production at 10pm
  omni note remove 1
  omni note remove 1770847088806767400
  omni note list
  omni note --list
  omni note --list --json
  omni note --file ./notes.json "local test note"`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := noteOptionsFromFlags(cmd)
		opts.List = noteBoolFlag(cmd, "list")
		opts.Limit = noteIntFlag(cmd, "limit")

		if !opts.List && len(args) == 0 {
			return cmd.Help()
		}

		return note.RunNote(cmd.OutOrStdout(), args, opts)
	},
}

var noteAddCmd = &cobra.Command{
	Use:   "add <TEXT...>",
	Short: "Add a new note entry",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := noteOptionsFromFlags(cmd)
		return note.RunNote(cmd.OutOrStdout(), args, opts)
	},
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List note entries",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := noteOptionsFromFlags(cmd)
		opts.List = true
		opts.Limit = noteIntFlag(cmd, "limit")
		return note.RunNote(cmd.OutOrStdout(), nil, opts)
	},
}

var noteRemoveCmd = &cobra.Command{
	Use:     "remove <INDEX_OR_ID>",
	Aliases: []string{"rm", "del"},
	Short:   "Remove a note entry by index or ID",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := noteOptionsFromFlags(cmd)
		return note.RunRemove(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(noteCmd)
	noteCmd.AddCommand(noteAddCmd)
	noteCmd.AddCommand(noteListCmd)
	noteCmd.AddCommand(noteRemoveCmd)

	noteCmd.PersistentFlags().String("file", "", "path to notes JSON file (default: Documents/omni-notes.json)")
	noteCmd.PersistentFlags().Bool("json", false, "output result in JSON")
	noteCmd.Flags().Bool("list", false, "list saved notes instead of adding a new one")
	noteCmd.Flags().IntP("limit", "n", 0, "show only last N notes (used with --list)")
	noteListCmd.Flags().IntP("limit", "n", 0, "show only last N notes")
}

func noteOptionsFromFlags(cmd *cobra.Command) note.Options {
	return note.Options{
		File: noteStringFlag(cmd, "file"),
		JSON: noteBoolFlag(cmd, "json"),
	}
}

func noteStringFlag(cmd *cobra.Command, name string) string {
	if cmd.Flags().Lookup(name) == nil {
		value, _ := cmd.InheritedFlags().GetString(name)
		return value
	}

	value, _ := cmd.Flags().GetString(name)
	return value
}

func noteBoolFlag(cmd *cobra.Command, name string) bool {
	if cmd.Flags().Lookup(name) == nil {
		value, _ := cmd.InheritedFlags().GetBool(name)
		return value
	}

	value, _ := cmd.Flags().GetBool(name)
	return value
}

func noteIntFlag(cmd *cobra.Command, name string) int {
	if cmd.Flags().Lookup(name) == nil {
		value, _ := cmd.InheritedFlags().GetInt(name)
		return value
	}

	value, _ := cmd.Flags().GetInt(name)
	return value
}
