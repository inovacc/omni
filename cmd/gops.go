package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/google/gops/goprocess"
	"github.com/spf13/cobra"
)

// GoProcessInfo represents Go process information for JSON output
type GoProcessInfo struct {
	PID       int    `json:"pid"`
	PPID      int    `json:"ppid"`
	Command   string `json:"command"`
	Version   string `json:"version"`
	BuildPath string `json:"build_path"`
}

// gopsCmd represents the gops command
var gopsCmd = &cobra.Command{
	Use:   "gops [PID]",
	Short: "Display Go process information",
	Long: `Display information about running Go processes.

Uses google/gops to detect Go processes and show their version
and build information.

Examples:
  omni gops           # list all Go processes
  omni gops -j        # output as JSON
  omni gops 1234      # show info for specific PID`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		procs := goprocess.FindAll()

		if len(procs) == 0 {
			if !jsonOutput {
				_, _ = fmt.Fprintln(os.Stdout, "No Go processes found")
			} else {
				_, _ = fmt.Fprintln(os.Stdout, "[]")
			}

			return nil
		}

		if jsonOutput {
			return printGopsJSON(procs)
		}

		return printGopsTable(procs)
	},
}

func printGopsTable(procs []goprocess.P) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "PID\tPPID\tCOMMAND\tVERSION\tBUILD PATH")

	for _, p := range procs {
		cmd := p.Exec
		if len(cmd) > 40 {
			cmd = cmd[:40] + "..."
		}

		path := p.Path
		if len(path) > 50 {
			path = "..." + path[len(path)-47:]
		}

		_, _ = fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%s\n",
			p.PID, p.PPID, cmd, p.BuildVersion, path)
	}

	return w.Flush()
}

func printGopsJSON(procs []goprocess.P) error {
	result := make([]GoProcessInfo, 0, len(procs))

	for _, p := range procs {
		result = append(result, GoProcessInfo{
			PID:       p.PID,
			PPID:      p.PPID,
			Command:   p.Exec,
			Version:   p.BuildVersion,
			BuildPath: p.Path,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode(result)
}

func init() {
	rootCmd.AddCommand(gopsCmd)

	gopsCmd.Flags().BoolP("json", "j", false, "output as JSON")
}
