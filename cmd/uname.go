package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// unameCmd represents the uname command
var unameCmd = &cobra.Command{
	Use:   "uname",
	Short: "Print system information",
	Long: `Print certain system information. With no OPTION, same as -s.

  -a, --all                print all information
  -s, --kernel-name        print the kernel name
  -n, --nodename           print the network node hostname
  -r, --kernel-release     print the kernel release
  -v, --kernel-version     print the kernel version
  -m, --machine            print the machine hardware name
  -p, --processor          print the processor type
  -i, --hardware-platform  print the hardware platform
  -o, --operating-system   print the operating system`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.UnameOptions{}

		opts.All, _ = cmd.Flags().GetBool("all")
		opts.KernelName, _ = cmd.Flags().GetBool("kernel-name")
		opts.NodeName, _ = cmd.Flags().GetBool("nodename")
		opts.KernelRelease, _ = cmd.Flags().GetBool("kernel-release")
		opts.KernelVersion, _ = cmd.Flags().GetBool("kernel-version")
		opts.Machine, _ = cmd.Flags().GetBool("machine")
		opts.Processor, _ = cmd.Flags().GetBool("processor")
		opts.HardwarePlatform, _ = cmd.Flags().GetBool("hardware-platform")
		opts.OperatingSystem, _ = cmd.Flags().GetBool("operating-system")

		return cli.RunUname(os.Stdout, opts)
	},
}

func init() {
	rootCmd.AddCommand(unameCmd)

	unameCmd.Flags().BoolP("all", "a", false, "print all information")
	unameCmd.Flags().BoolP("kernel-name", "s", false, "print the kernel name")
	unameCmd.Flags().BoolP("nodename", "n", false, "print the network node hostname")
	unameCmd.Flags().BoolP("kernel-release", "r", false, "print the kernel release")
	unameCmd.Flags().BoolP("kernel-version", "v", false, "print the kernel version")
	unameCmd.Flags().BoolP("machine", "m", false, "print the machine hardware name")
	unameCmd.Flags().BoolP("processor", "p", false, "print the processor type")
	unameCmd.Flags().BoolP("hardware-platform", "i", false, "print the hardware platform")
	unameCmd.Flags().BoolP("operating-system", "o", false, "print the operating system")
}
