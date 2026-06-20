package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/inovacc/omni/internal/cli/runtimeps"
	"github.com/inovacc/omni/pkg/procutil"
)

// gopsCmd lists Go processes and signals them. Replaces the older
// google/gops-based implementation with a pure-Go classifier under
// pkg/procutil (no embedded agent required).
var gopsCmd = &cobra.Command{
	Use:   "gops",
	Short: "List and signal running Go processes",
	Long: `Enumerate Go processes by inspecting binaries on disk via debug/buildinfo.
Pure Go — no external commands, no embedded agent required.

Examples:
  omni gops                                # list Go processes (table)
  omni gops --json                         # list as JSON
  omni gops kill 12345                     # signal by PID
  omni gops kill myapp                     # signal the single matching process
  omni gops kill myapp --recursive --yes   # signal every Go process named "myapp"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRuntimePsList(cmd, procutil.RuntimeGo)
	},
}

var gopsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Go processes (default action of `omni gops`)",
	Long: `List running Go processes. This is the default action of "omni gops".

Examples:
  omni gops list                  # list Go processes
  omni gops list --json           # list as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error { return runRuntimePsList(cmd, procutil.RuntimeGo) },
}

var gopsKillCmd = &cobra.Command{
	Use:   "kill <pid|name>",
	Short: "Signal one or more Go processes",
	Long: `Signal a Go process by PID or by name.

Examples:
  omni gops kill 12345            # signal by PID
  omni gops kill myapp --recursive --yes  # signal every matching process`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error { return runRuntimePsKill(cmd, procutil.RuntimeGo, args[0]) },
}

var gopsInspectCmd = &cobra.Command{
	Use:   "inspect <pid>",
	Short: "Detail report for a single Go process (build info + obfuscation)",
	Long: `Show a detailed report (build info and obfuscation status) for one Go process.

Examples:
  omni gops inspect 12345         # detail report for a PID`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format := "table"
		if j, _ := cmd.Flags().GetBool("json"); j {
			format = "json"
		}
		return runtimeps.RunInspect(cmd.Context(), cmd.OutOrStdout(), args[0], runtimeps.InspectOptions{Format: format})
	},
}

var gopsMonitorCmd = &cobra.Command{
	Use:   "monitor <pid>",
	Short: "Sample CPU/memory/IO/FD metrics for a process (single-shot or --watch)",
	Long: `Sample CPU, memory, IO, and FD metrics for a process, once or continuously.

Examples:
  omni gops monitor 12345         # single-shot sample
  omni gops monitor 12345 --watch # stream metrics as NDJSON`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		watch, _ := cmd.Flags().GetBool("watch")
		interval, _ := cmd.Flags().GetDuration("interval")
		format := "table"
		if j, _ := cmd.Flags().GetBool("json"); j {
			format = "json"
		}
		return runtimeps.RunMonitor(cmd.Context(), cmd.OutOrStdout(), args[0], runtimeps.MonitorOptions{
			Watch:    watch,
			Interval: interval,
			Format:   format,
		})
	},
}

var gopsObfuscationCmd = &cobra.Command{
	Use:   "obfuscation <pid|path>",
	Short: "Detect garble-style obfuscation in a Go binary (by PID or file path)",
	Long: `Detect garble-style obfuscation in a Go binary, by PID or by file path.

Examples:
  omni gops obfuscation 12345     # check a running process
  omni gops obfuscation ./app     # check a binary on disk`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format := "table"
		if j, _ := cmd.Flags().GetBool("json"); j {
			format = "json"
		}
		return runtimeps.RunObfuscation(cmd.Context(), cmd.OutOrStdout(), args[0], runtimeps.ObfuscationOptions{Format: format})
	},
}

var gopsTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Interactive TUI dashboard of Go processes (q to quit)",
	Long: `Launch an interactive TUI dashboard of Go processes. Press q to quit.

Examples:
  omni gops top                   # open the dashboard
  omni gops top --interval 2s     # refresh every 2 seconds`,
	RunE: func(cmd *cobra.Command, args []string) error {
		interval, _ := cmd.Flags().GetDuration("interval")
		all, _ := cmd.Flags().GetBool("all")
		return runtimeps.RunTop(cmd.Context(), interval, all)
	},
}

var gopsAgentCmd = &cobra.Command{
	Use:   "agent-cmd <pid> <stack|gc|memstats|version|stats|snapshot>",
	Short: "Send an opcode to a target's embedded gops agent",
	Long: `Send an opcode to a target process's embedded gops agent.

Examples:
  omni gops agent-cmd 12345 stack     # request a stack dump
  omni gops agent-cmd 12345 memstats  # request memory stats`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runtimeps.RunAgentCmd(cmd.Context(), cmd.OutOrStdout(), args[0], args[1])
	},
}

var gopsTraceCmd = &cobra.Command{
	Use:   "trace <pid>",
	Short: "Capture a runtime trace from a target's embedded agent (analyze with `go tool trace`)",
	Long: `Capture a runtime trace from a target's embedded agent. Analyze with "go tool trace".

Examples:
  omni gops trace 12345                  # capture a trace to stdout
  omni gops trace 12345 -d 10s -f t.out  # 10s trace to a file`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dur, _ := cmd.Flags().GetDuration("duration")
		out, _ := cmd.Flags().GetString("file")
		return runtimeps.RunTrace(cmd.Context(), cmd.OutOrStdout(), args[0], runtimeps.TraceOptions{Duration: dur, OutFile: out})
	},
}

var gopsProfileCmd = &cobra.Command{
	Use:   "profile <pid>",
	Short: "Capture a CPU profile from a target's embedded agent (analyze with `go tool pprof`)",
	Long: `Capture a CPU profile from a target's embedded agent. Analyze with "go tool pprof".

Examples:
  omni gops profile 12345                 # capture a CPU profile to stdout
  omni gops profile 12345 -d 30s -f p.out # 30s profile to a file`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dur, _ := cmd.Flags().GetDuration("duration")
		out, _ := cmd.Flags().GetString("file")
		return runtimeps.RunProfile(cmd.Context(), cmd.OutOrStdout(), args[0], runtimeps.ProfileOptions{Duration: dur, OutFile: out})
	},
}

var gopsStreamCmd = &cobra.Command{
	Use:   "stream <pid>",
	Short: "Stream NDJSON runtime snapshots from a target's embedded agent",
	Long: `Stream NDJSON runtime snapshots from a target's embedded agent.

Examples:
  omni gops stream 12345                  # stream snapshots
  omni gops stream 12345 --interval 500ms # one snapshot every 500ms`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		interval, _ := cmd.Flags().GetDuration("interval")
		return runtimeps.RunStream(cmd.Context(), cmd.OutOrStdout(), args[0], runtimeps.StreamOptions{Interval: interval})
	},
}

func init() {
	registerRuntimePsFlags(gopsCmd, gopsListCmd, gopsKillCmd)
	gopsInspectCmd.Flags().BoolP("json", "j", false, "output as JSON")
	gopsObfuscationCmd.Flags().BoolP("json", "j", false, "output as JSON")
	gopsMonitorCmd.Flags().BoolP("json", "j", false, "output as JSON (ignored when --watch; that path streams NDJSON)")
	gopsMonitorCmd.Flags().BoolP("watch", "w", false, "stream metrics continuously as NDJSON")
	gopsMonitorCmd.Flags().DurationP("interval", "i", time.Second, "sampling interval when --watch")
	gopsTopCmd.Flags().DurationP("interval", "i", time.Second, "refresh interval")
	gopsTopCmd.Flags().BoolP("all", "a", false, "include the omni process itself")
	gopsTraceCmd.Flags().DurationP("duration", "d", 5*time.Second, "trace duration (max 600s)")
	gopsTraceCmd.Flags().StringP("file", "f", "", "output file (default stdout)")
	gopsProfileCmd.Flags().DurationP("duration", "d", 30*time.Second, "profile duration (max 600s)")
	gopsProfileCmd.Flags().StringP("file", "f", "", "output file (default stdout)")
	gopsStreamCmd.Flags().DurationP("interval", "i", time.Second, "snapshot interval (50ms-60s)")
	gopsCmd.AddCommand(
		gopsListCmd, gopsKillCmd, gopsInspectCmd, gopsMonitorCmd, gopsObfuscationCmd, gopsTopCmd,
		gopsAgentCmd, gopsTraceCmd, gopsProfileCmd, gopsStreamCmd,
	)
	rootCmd.AddCommand(gopsCmd)
}

// Shared helpers live in cmd/runtimeps_shared.go so nodeps/pyps/javaps can reuse them.

func runRuntimePsList(cmd *cobra.Command, rt procutil.Runtime) error {
	all, _ := cmd.Flags().GetBool("all")
	format := "table"
	if j, _ := cmd.Flags().GetBool("json"); j {
		format = "json"
	}
	return runtimeps.RunList(cmd.Context(), cmd.OutOrStdout(), rt, runtimeps.ListOptions{
		All:    all,
		Format: format,
	})
}

func runRuntimePsKill(cmd *cobra.Command, rt procutil.Runtime, target string) error {
	sig, _ := cmd.Flags().GetString("signal")
	rec, _ := cmd.Flags().GetBool("recursive")
	yes, _ := cmd.Flags().GetBool("yes")
	return runtimeps.RunKill(cmd.Context(), cmd.OutOrStdout(), rt, target, runtimeps.KillOptions{
		Signal:    sig,
		Recursive: rec,
		Yes:       yes,
	})
}

func registerRuntimePsFlags(root, listCmd, killCmd *cobra.Command) {
	for _, c := range []*cobra.Command{root, listCmd} {
		c.Flags().BoolP("all", "a", false, "include the omni process itself")
		c.Flags().BoolP("json", "j", false, "output as JSON")
	}
	killCmd.Flags().String("signal", "TERM", "signal to deliver: TERM|KILL|INT|HUP (Windows: TERM/KILL only)")
	killCmd.Flags().Bool("recursive", false, "kill every matching process (required when name matches >1 process)")
	killCmd.Flags().Bool("yes", false, "confirm --recursive without prompt (required by --recursive)")
}
