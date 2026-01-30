package cmd

import (
	"github.com/inovacc/omni/internal/cli/find"
	"github.com/spf13/cobra"
)

var (
	findName       string
	findIName      string
	findPath       string
	findIPath      string
	findRegex      string
	findIRegex     string
	findType       string
	findSize       string
	findMinDepth   int
	findMaxDepth   int
	findMTime      string
	findMMin       string
	findATime      string
	findAMin       string
	findEmpty      bool
	findExecutable bool
	findReadable   bool
	findWritable   bool
	findPrint0     bool
	findNot        bool
	findJSON       bool
)

var findCmd = &cobra.Command{
	Use:   "find [path...] [expression]",
	Short: "Search for files in a directory hierarchy",
	Long: `Search for files in a directory hierarchy.

Tests:
  -name PATTERN      file name matches shell PATTERN
  -iname PATTERN     like -name, but case insensitive
  -path PATTERN      path matches shell PATTERN
  -ipath PATTERN     like -path, but case insensitive
  -regex PATTERN     path matches regular expression PATTERN
  -iregex PATTERN    like -regex, but case insensitive
  -type TYPE         file type: f(ile), d(ir), l(ink), p(ipe), s(ocket)
  -size N[cwbkMGTP]  file size is N units (c=bytes, k=KB, M=MB, G=GB)
                     prefix + for greater than, - for less than
  -mindepth N        do not apply tests at levels less than N
  -maxdepth N        descend at most N levels
  -mtime N           modified N*24 hours ago (+N=more than, -N=less than)
  -mmin N            modified N minutes ago
  -atime N           accessed N*24 hours ago
  -amin N            accessed N minutes ago
  -empty             file is empty or directory has no entries
  -executable        matches files which are executable
  -readable          matches files which are readable
  -writable          matches files which are writable

Actions:
  -print0            print full path with null terminator

Operators:
  -not               negate the next test

Examples:
  omni find . -name "*.go"                    # find Go files
  omni find . -type f -size +1M               # find files larger than 1MB
  omni find /tmp -type f -mtime +7            # files modified more than 7 days ago
  omni find . -name "*.log" -empty            # find empty log files
  omni find . -type d -name "node_modules"    # find node_modules directories
  omni find . -maxdepth 2 -type f             # files at most 2 levels deep
  omni find . -name "*.txt" -print0           # null-separated output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := find.FindOptions{
			Name:       findName,
			IName:      findIName,
			Path:       findPath,
			IPath:      findIPath,
			Regex:      findRegex,
			IRegex:     findIRegex,
			Type:       findType,
			Size:       findSize,
			MinDepth:   findMinDepth,
			MaxDepth:   findMaxDepth,
			MTime:      findMTime,
			MMin:       findMMin,
			ATime:      findATime,
			AMin:       findAMin,
			Empty:      findEmpty,
			Executable: findExecutable,
			Readable:   findReadable,
			Writable:   findWritable,
			Print0:     findPrint0,
			Not:        findNot,
			JSON:       findJSON,
		}

		return find.RunFind(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().StringVarP(&findName, "name", "", "", "file name matches pattern")
	findCmd.Flags().StringVarP(&findIName, "iname", "", "", "case insensitive name match")
	findCmd.Flags().StringVarP(&findPath, "path", "", "", "path matches pattern")
	findCmd.Flags().StringVarP(&findIPath, "ipath", "", "", "case insensitive path match")
	findCmd.Flags().StringVarP(&findRegex, "regex", "", "", "path matches regex")
	findCmd.Flags().StringVarP(&findIRegex, "iregex", "", "", "case insensitive regex")
	findCmd.Flags().StringVarP(&findType, "type", "", "", "file type (f=file, d=dir, l=link)")
	findCmd.Flags().StringVarP(&findSize, "size", "", "", "file size [+-]N[ckMG]")
	findCmd.Flags().IntVarP(&findMinDepth, "mindepth", "", 0, "minimum depth")
	findCmd.Flags().IntVarP(&findMaxDepth, "maxdepth", "", 0, "maximum depth (0=unlimited)")
	findCmd.Flags().StringVarP(&findMTime, "mtime", "", "", "modification time [+-]N days")
	findCmd.Flags().StringVarP(&findMMin, "mmin", "", "", "modification time [+-]N minutes")
	findCmd.Flags().StringVarP(&findATime, "atime", "", "", "access time [+-]N days")
	findCmd.Flags().StringVarP(&findAMin, "amin", "", "", "access time [+-]N minutes")
	findCmd.Flags().BoolVarP(&findEmpty, "empty", "", false, "file is empty")
	findCmd.Flags().BoolVarP(&findExecutable, "executable", "", false, "file is executable")
	findCmd.Flags().BoolVarP(&findReadable, "readable", "", false, "file is readable")
	findCmd.Flags().BoolVarP(&findWritable, "writable", "", false, "file is writable")
	findCmd.Flags().BoolVarP(&findPrint0, "print0", "0", false, "print with null terminator")
	findCmd.Flags().BoolVarP(&findNot, "not", "", false, "negate next test")
	findCmd.Flags().BoolVarP(&findJSON, "json", "", false, "output in JSON format")
}
