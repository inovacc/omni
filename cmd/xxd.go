package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/xxd"
	"github.com/spf13/cobra"
)

var xxdCmd = &cobra.Command{
	Use:   "xxd [OPTIONS] [FILE]",
	Short: "Make a hex dump or reverse it",
	Long: `Make a hex dump of a file or standard input, or reverse it.

xxd creates a hex dump of a given file or standard input. It can also convert
a hex dump back to its original binary form.

Output Modes:
  (default)    Traditional hex dump with addresses and ASCII
  -p, --plain  Output only hex bytes, no addresses or ASCII
  -i, --include  Output as C include file (array definition)
  -b, --bits   Binary digit dump instead of hex

Reverse Mode:
  -r, --reverse  Convert hex dump back to binary

Examples:
  # Basic hex dump
  omni xxd file.bin
  omni xxd < file.bin
  echo "hello" | omni xxd

  # Plain hex output (like 'hex encode' but for binary files)
  omni xxd -p file.bin

  # C include file output
  omni xxd -i data.bin > data.h

  # Binary dump (bits instead of hex)
  omni xxd -b file.bin

  # Reverse hex dump back to binary
  omni xxd -r hexdump.txt > original.bin
  omni xxd -p file.bin | omni xxd -r -p > copy.bin

  # Limit output to first N bytes
  omni xxd -l 16 file.bin

  # Start at offset
  omni xxd -s 100 file.bin

  # Custom columns and grouping
  omni xxd -c 8 -g 1 file.bin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xxd.Options{
			Columns:   16,
			Groups:    2,
			Length:    0,
			Seek:      0,
			Reverse:   false,
			Plain:     false,
			Include:   false,
			Uppercase: false,
			Bits:      false,
		}

		opts.Columns, _ = cmd.Flags().GetInt("cols")
		opts.Groups, _ = cmd.Flags().GetInt("groupsize")
		opts.Length, _ = cmd.Flags().GetInt("len")
		opts.Seek, _ = cmd.Flags().GetInt("seek")
		opts.Reverse, _ = cmd.Flags().GetBool("reverse")
		opts.Plain, _ = cmd.Flags().GetBool("plain")
		opts.Include, _ = cmd.Flags().GetBool("include")
		opts.Uppercase, _ = cmd.Flags().GetBool("uppercase")
		opts.Bits, _ = cmd.Flags().GetBool("bits")

		return xxd.Run(cmd.OutOrStdout(), os.Stdin, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(xxdCmd)

	xxdCmd.Flags().IntP("cols", "c", 16, "format <cols> octets per line (default 16)")
	xxdCmd.Flags().IntP("groupsize", "g", 2, "separate output with <bytes> spaces (default 2)")
	xxdCmd.Flags().IntP("len", "l", 0, "stop after <len> octets")
	xxdCmd.Flags().IntP("seek", "s", 0, "start at <seek> bytes offset")
	xxdCmd.Flags().BoolP("reverse", "r", false, "reverse operation: convert hex dump to binary")
	xxdCmd.Flags().BoolP("plain", "p", false, "output plain hex dump (no addresses or ASCII)")
	xxdCmd.Flags().BoolP("include", "i", false, "output in C include file style")
	xxdCmd.Flags().BoolP("uppercase", "u", false, "use uppercase hex letters")
	xxdCmd.Flags().BoolP("bits", "b", false, "binary digit dump (bits instead of hex)")
}
