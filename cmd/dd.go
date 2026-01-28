package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/inovacc/omni/pkg/cli/dd"
	"github.com/spf13/cobra"
)

var ddCmd = &cobra.Command{
	Use:   "dd [OPERAND]...",
	Short: "Convert and copy a file",
	Long: `Copy a file, converting and formatting according to the operands.

  if=FILE     read from FILE instead of stdin
  of=FILE     write to FILE instead of stdout
  bs=BYTES    read and write up to BYTES bytes at a time
  ibs=BYTES   read up to BYTES bytes at a time (default: 512)
  obs=BYTES   write BYTES bytes at a time (default: 512)
  count=N     copy only N input blocks
  skip=N      skip N ibs-sized blocks at start of input
  seek=N      skip N obs-sized blocks at start of output
  conv=CONVS  convert the file as per the comma separated symbol list
  status=LEVEL  LEVEL of information to print to stderr:
              'none' suppresses everything but error messages,
              'noxfer' suppresses the final transfer statistics,
              'progress' shows periodic transfer statistics

CONV symbols:
  lcase       change upper case to lower case
  ucase       change lower case to upper case
  swab        swap every pair of input bytes
  notrunc     do not truncate the output file
  fsync       physically write output file data before finishing

BYTES may be followed by multiplicative suffixes:
  K=1024, M=1024*1024, G=1024*1024*1024

Examples:
  omni dd if=input.txt of=output.txt               # copy file
  omni dd if=/dev/zero of=file.bin bs=1M count=10  # create 10MB file
  omni dd if=file.txt conv=ucase                   # convert to uppercase
  omni dd if=disk.img of=backup.img bs=4K          # disk image backup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := dd.DdOptions{}

		// Parse operands
		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("dd: invalid operand %q", arg)
			}

			key, value := parts[0], parts[1]
			var err error

			switch key {
			case "if":
				opts.InputFile = value
			case "of":
				opts.OutputFile = value
			case "bs":
				opts.BlockSize, err = dd.ParseDdSize(value)
			case "ibs":
				opts.InputBS, err = dd.ParseDdSize(value)
			case "obs":
				opts.OutputBS, err = dd.ParseDdSize(value)
			case "count":
				opts.Count, err = dd.ParseDdSize(value)
			case "skip":
				opts.Skip, err = dd.ParseDdSize(value)
			case "seek":
				opts.Seek, err = dd.ParseDdSize(value)
			case "conv":
				opts.Conv = value
			case "status":
				opts.Status = value
			default:
				return fmt.Errorf("dd: unknown operand %q", key)
			}

			if err != nil {
				return fmt.Errorf("dd: invalid value for %s: %w", key, err)
			}
		}

		return dd.RunDd(os.Stdout, opts)
	},
}

func init() {
	rootCmd.AddCommand(ddCmd)
}
