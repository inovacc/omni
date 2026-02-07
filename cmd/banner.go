package cmd

import (
	"github.com/inovacc/omni/internal/cli/banner"
	"github.com/spf13/cobra"
)

var bannerCmd = &cobra.Command{
	Use:   "banner [TEXT]",
	Short: "Generate ASCII art text banners",
	Long: `Generate FIGlet-style ASCII art text banners.

Supports multiple fonts and reads text from arguments or stdin.

  -f, --font=NAME   font name (default "standard")
  -w, --width=N     max output width (0 = unlimited)
  -l, --list        list available fonts

Examples:
  omni banner "Hello World"
  omni banner -f slant "omni"
  omni banner -f small "test"
  omni banner --list
  echo "piped" | omni banner`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := banner.Options{}

		opts.Font, _ = cmd.Flags().GetString("font")
		opts.Width, _ = cmd.Flags().GetInt("width")
		opts.List, _ = cmd.Flags().GetBool("list")

		return banner.RunBanner(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(bannerCmd)

	bannerCmd.Flags().StringP("font", "f", "standard", "font name")
	bannerCmd.Flags().IntP("width", "w", 0, "max output width (0 = unlimited)")
	bannerCmd.Flags().BoolP("list", "l", false, "list available fonts")
}
