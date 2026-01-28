package df

import (
	"fmt"
	"io"

	"github.com/inovacc/omni/pkg/cli"
)

// DFOptions configures the df command behavior
type DFOptions struct {
	HumanReadable bool   // -h: print sizes in human readable format
	Inodes        bool   // -i: list inode information instead of block usage
	BlockSize     int64  // -B: scale sizes by SIZE
	Total         bool   // --total: produce a grand total
	Type          string // -t: limit listing to file systems of given TYPE
	ExcludeType   string // -x: exclude file systems of given TYPE
	Local         bool   // -l: limit listing to local file systems
	Portability   bool   // -P: use POSIX output format
}

// DFInfo represents disk free space information
type DFInfo struct {
	Filesystem string `json:"filesystem"`
	Type       string `json:"type"`
	Size       uint64 `json:"size"`
	Used       uint64 `json:"used"`
	Available  uint64 `json:"available"`
	UsePercent int    `json:"usePercent"`
	MountedOn  string `json:"mountedOn"`
	// Inode info
	Inodes      uint64 `json:"inodes,omitempty"`
	IUsed       uint64 `json:"iused,omitempty"`
	IFree       uint64 `json:"ifree,omitempty"`
	IUsePercent int    `json:"iusePercent,omitempty"`
}

// RunDF executes the df command
func RunDF(w io.Writer, args []string, opts DFOptions) error {
	if opts.BlockSize == 0 {
		if opts.HumanReadable {
			opts.BlockSize = 1
		} else {
			opts.BlockSize = 1024 // Default 1K blocks
		}
	}

	paths := args
	if len(paths) == 0 {
		paths = []string{"/"}
	}

	// Print header
	switch {
	case opts.Inodes:
		_, _ = fmt.Fprintf(w, "%-20s %10s %10s %10s %5s %s\n",
			"Filesystem", "Inodes", "IUsed", "IFree", "IUse%", "Mounted on")
	case opts.HumanReadable:
		_, _ = fmt.Fprintf(w, "%-20s %6s %6s %6s %5s %s\n",
			"Filesystem", "Size", "Used", "Avail", "Use%", "Mounted on")
	default:
		_, _ = fmt.Fprintf(w, "%-20s %10s %10s %10s %5s %s\n",
			"Filesystem", "1K-blocks", "Used", "Available", "Use%", "Mounted on")
	}

	var total DFInfo

	total.Filesystem = "total"

	for _, path := range paths {
		info, err := getDiskInfo(path)
		if err != nil {
			_, _ = fmt.Fprintf(w, "df: %s: %v\n", path, err)
			continue
		}

		printDFInfo(w, info, opts)

		// Accumulate totals
		total.Size += info.Size
		total.Used += info.Used
		total.Available += info.Available
		total.Inodes += info.Inodes
		total.IUsed += info.IUsed
		total.IFree += info.IFree
	}

	if opts.Total && len(paths) > 1 {
		if total.Size > 0 {
			total.UsePercent = int(float64(total.Used) / float64(total.Size) * 100)
		}

		if total.Inodes > 0 {
			total.IUsePercent = int(float64(total.IUsed) / float64(total.Inodes) * 100)
		}

		total.MountedOn = "-"
		printDFInfo(w, total, opts)
	}

	return nil
}

func printDFInfo(w io.Writer, info DFInfo, opts DFOptions) {
	switch {
	case opts.Inodes:
		_, _ = fmt.Fprintf(w, "%-20s %10d %10d %10d %4d%% %s\n",
			info.Filesystem,
			info.Inodes,
			info.IUsed,
			info.IFree,
			info.IUsePercent,
			info.MountedOn)
	case opts.HumanReadable:
		_, _ = fmt.Fprintf(w, "%-20s %6s %6s %6s %4d%% %s\n",
			info.Filesystem,
			cli.FormatHumanSize(int64(info.Size)),
			cli.FormatHumanSize(int64(info.Used)),
			cli.FormatHumanSize(int64(info.Available)),
			info.UsePercent,
			info.MountedOn)
	default:
		blocks := info.Size / uint64(opts.BlockSize)
		usedBlocks := info.Used / uint64(opts.BlockSize)
		availBlocks := info.Available / uint64(opts.BlockSize)

		_, _ = fmt.Fprintf(w, "%-20s %10d %10d %10d %4d%% %s\n",
			info.Filesystem,
			blocks,
			usedBlocks,
			availBlocks,
			info.UsePercent,
			info.MountedOn)
	}
}

// GetDiskFree returns disk space information for a path
func GetDiskFree(path string) (DFInfo, error) {
	return getDiskInfo(path)
}
