// Package templates — daemon templates emitted by `omni scaffold cobra init --daemon`.
//
// The daemon pattern is a self-daemonizing PID-file based supervisor, as opposed
// to the OS-service-manager pattern emitted by --service. The two are
// complementary: a daemon-style app can ALSO register itself with the OS
// service manager via the platform-specific installService/uninstallService
// functions emitted alongside.
//
// File layout when --daemon is set:
//
//	internal/serverinfo/serverinfo.go   — PID/version JSON state + gopsutil PID validation
//	cmd/{app}/cmd_server.go             — Cobra subcommands: start/stop/restart/status/install/uninstall/run
//	cmd/{app}/server.go                 — Shared start/stop/daemonize logic
//	cmd/{app}/server_unix.go            — Unix helpers (//go:build !windows)
//	cmd/{app}/server_systemd.go         — systemd install/uninstall (//go:build !windows && !darwin)
//	cmd/{app}/server_darwin.go          — launchd install/uninstall (//go:build darwin)
//	cmd/{app}/server_windows.go         — Windows SCM install/uninstall + helpers (//go:build windows)
package templates

// ServerInfoTemplate generates internal/serverinfo/serverinfo.go.
const ServerInfoTemplate = `// Package serverinfo persists and inspects information about the running
// {{.AppName}} daemon instance via a JSON file under the user config dir.
package serverinfo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const appName = "{{.AppName}}"

var (
	ErrNoServerInfo    = errors.New("no server info file")
	ErrVersionMismatch = errors.New("running instance has a different version")
)

// Info describes a running {{.AppName}} instance.
type Info struct {
	PID       int       ` + "`json:\"pid\"`" + `
	StartedAt time.Time ` + "`json:\"started_at\"`" + `
	Version   string    ` + "`json:\"version,omitempty\"`" + `
}

// Version returns the module version baked in at build time, falling back to "dev".
func Version() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "dev"
}

// VersionMatch returns nil if i.Version is empty or matches the current binary.
func (i *Info) VersionMatch() error {
	cur := Version()
	if i.Version == "" || i.Version == cur {
		return nil
	}
	return fmt.Errorf("%w: running %s, binary is %s", ErrVersionMismatch, i.Version, cur)
}

// Path returns the on-disk location of the server info file.
func Path() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, appName, "server.json")
}

// Read loads the server info file if it exists.
func Read() (*Info, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoServerInfo
		}
		return nil, fmt.Errorf("read server info: %w", err)
	}
	var info Info
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("parse server info: %w", err)
	}
	return &info, nil
}

// IsRunning returns the recorded info iff the recorded PID is alive and looks
// like our binary. Stale info files are removed.
func IsRunning() *Info {
	info, err := Read()
	if err != nil {
		return nil
	}
	if isOurProcess(info.PID) {
		return info
	}
	Remove()
	return nil
}

func isOurProcess(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return false
	}
	name, err := p.Name()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(name), appName)
}

// Write records the current process as the running daemon.
func Write() error {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir serverinfo dir: %w", err)
	}
	data, err := json.MarshalIndent(Info{
		PID:       os.Getpid(),
		StartedAt: time.Now(),
		Version:   Version(),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal server info: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write server info: %w", err)
	}
	return nil
}

// Remove deletes the server info file. Safe to call when no file exists.
func Remove() {
	_ = os.Remove(Path())
}
`

// DaemonCmdTemplate generates cmd/{AppName}/cmd_server.go — the Cobra command group.
const DaemonCmdTemplate = `package main

import (
	"github.com/spf13/cobra"
)

// ServerForeground controls whether ` + "`{{.AppName}} server start`" + ` daemonizes (false)
// or stays in the foreground (true). Wired by --foreground on the start command.
var ServerForeground bool

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the {{.AppName}} daemon",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the {{.AppName}} daemon (forks to background unless --foreground)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serverStart()
	},
}

var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running {{.AppName}} daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serverStop()
	},
}

var serverRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the {{.AppName}} daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serverRestart()
	},
}

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show {{.AppName}} daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serverStatus()
	},
}

var serverInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install {{.AppName}} as an OS service (requires root/admin)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serverInstall()
	},
}

var serverUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the {{.AppName}} OS service (requires root/admin)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serverUninstall()
	},
}

func init() {
	serverStartCmd.Flags().BoolVar(&ServerForeground, "foreground", false, "stay in foreground (do not daemonize)")
	serverCmd.AddCommand(
		serverStartCmd,
		serverStopCmd,
		serverRestartCmd,
		serverStatusCmd,
		serverInstallCmd,
		serverUninstallCmd,
	)
	rootCmd.AddCommand(serverCmd)
}
`

// DaemonServerTemplate generates cmd/{AppName}/server.go — shared logic.
const DaemonServerTemplate = `package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"{{.Module}}/internal/serverinfo"
)

const daemonEnvVar = "{{.AppNameUpper}}_DAEMON_CHILD"
const osServiceName = "{{.AppName}}"
const osServiceDisplay = "{{.AppName}} service"

// runServe is the daemon's main loop. TODO: fill in. It MUST call
// serverinfo.Write() once it is ready to serve, and serverinfo.Remove() on exit.
func runServe() error {
	if err := serverinfo.Write(); err != nil {
		return err
	}
	defer serverinfo.Remove()

	slog.Info("{{.AppName}} serving", slog.Int("pid", os.Getpid()))
	// Replace with your real server: HTTP, gRPC, worker loop, etc.
	select {}
}

func serverStart() error {
	if info := serverinfo.IsRunning(); info != nil {
		if err := info.VersionMatch(); err != nil {
			return fmt.Errorf("{{.AppName}} already running (PID %d) but %w — stop the running instance first", info.PID, err)
		}
		return fmt.Errorf("{{.AppName}} already running (PID %d, started %s)",
			info.PID, info.StartedAt.Format("2006-01-02 15:04:05"))
	}

	if ServerForeground {
		slog.Info("Starting {{.AppName}} in foreground")
		return runServe()
	}

	if os.Getenv(daemonEnvVar) == "1" {
		return runServe()
	}

	return daemonize()
}

func daemonize() error {
	binary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	cmd := exec.Command(binary, os.Args[1:]...)
	cmd.Env = append(os.Environ(), daemonEnvVar+"=1")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	fmt.Printf("{{.AppName}} started (PID %d)\n", cmd.Process.Pid)
	return nil
}

func serverStop() error {
	info, err := serverinfo.Read()
	if err != nil {
		return fmt.Errorf("{{.AppName}} is not running (no server info)")
	}
	if err := stopProcess(info.PID); err != nil {
		return err
	}
	fmt.Printf("Stop signal sent to PID %d\n", info.PID)
	return nil
}

func serverRestart() error {
	if info := serverinfo.IsRunning(); info != nil {
		slog.Info("Stopping running instance", slog.Int("pid", info.PID))
		if err := serverStop(); err != nil {
			slog.Warn("Failed to stop existing instance", slog.Any("error", err))
		}
		time.Sleep(2 * time.Second)
	}
	return serverStart()
}

func serverStatus() error {
	info := serverinfo.IsRunning()
	if info == nil {
		fmt.Println("{{.AppName}} is not running")
		return nil
	}
	uptime := time.Since(info.StartedAt).Truncate(time.Second)
	ver := info.Version
	if ver == "" {
		ver = "(unknown)"
	}
	fmt.Printf("{{.AppName}} is running\n  PID:     %d\n  Version: %s\n  Started: %s\n  Uptime:  %s\n",
		info.PID, ver, info.StartedAt.Format("2006-01-02 15:04:05"), uptime)
	if err := info.VersionMatch(); err != nil {
		fmt.Printf("\n  WARNING: %s\n", err)
	}
	return nil
}

func serverInstall() error {
	if !isPrivileged() {
		slog.Info("Elevating privileges for service install")
		return elevateAndRerun()
	}
	return installService(osServiceName, osServiceDisplay)
}

func serverUninstall() error {
	if !isPrivileged() {
		slog.Info("Elevating privileges for service uninstall")
		return elevateAndRerun()
	}
	return uninstallService(osServiceName)
}
`

// DaemonServerUnixTemplate generates cmd/{AppName}/server_unix.go (covers darwin too;
// install/uninstall on each Unix variant live in server_systemd.go / server_darwin.go).
const DaemonServerUnixTemplate = `//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"syscall"

	"{{.Module}}/internal/serverinfo"
)

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func stopProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process %d not found: %w", pid, err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		serverinfo.Remove()
		return fmt.Errorf("signal process %d: %w", pid, err)
	}
	return nil
}

func isPrivileged() bool { return os.Geteuid() == 0 }

func elevateAndRerun() error {
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command("sudo", append([]string{binary}, os.Args[1:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func resolveRealUser() string {
	if u := os.Getenv("SUDO_USER"); u != "" {
		return u
	}
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return "root"
}
`

// DaemonServerSystemdTemplate generates cmd/{AppName}/server_systemd.go.
const DaemonServerSystemdTemplate = `//go:build !windows && !darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const systemdUnitTemplate = ` + "`" + `[Unit]
Description=%s
After=network.target

[Service]
Type=simple
ExecStart=%s server start --foreground
Restart=on-failure
RestartSec=5
User=%s
WorkingDirectory=%s

[Install]
WantedBy=multi-user.target
` + "`" + `

func installService(serviceName, displayName string) error {
	binary, _ := os.Executable()
	realUser := resolveRealUser()
	workDir, _ := os.Getwd()

	unit := fmt.Sprintf(systemdUnitTemplate, displayName, binary, realUser, workDir)
	unitPath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)

	if err := os.WriteFile(unitPath, []byte(unit), 0o644); err != nil {
		return fmt.Errorf("write unit file: %w", err)
	}
	for _, args := range [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", serviceName},
	} {
		if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %s", strings.Join(args, " "), out)
		}
	}
	fmt.Printf("Service %q installed and enabled\n", serviceName)
	return nil
}

func uninstallService(serviceName string) error {
	unitPath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	for _, args := range [][]string{
		{"systemctl", "stop", serviceName},
		{"systemctl", "disable", serviceName},
	} {
		_, _ = exec.Command(args[0], args[1:]...).CombinedOutput()
	}
	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove unit file: %w", err)
	}
	_, _ = exec.Command("systemctl", "daemon-reload").CombinedOutput()
	fmt.Printf("Service %q uninstalled\n", serviceName)
	return nil
}
`

// DaemonServerDarwinTemplate generates cmd/{AppName}/server_darwin.go.
const DaemonServerDarwinTemplate = `//go:build darwin

package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
)

const launchdLabel = "com.{{.AppName}}.daemon"

func launchdPlistPath() string {
	return fmt.Sprintf("/Library/LaunchDaemons/%s.plist", launchdLabel)
}

func escapeXML(s string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}

func installService(serviceName, displayName string) error {
	_ = serviceName
	_ = displayName

	binary, _ := os.Executable()
	realUser := resolveRealUser()
	workDir, _ := os.Getwd()

	plist := fmt.Sprintf(` + "`" + `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key><string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>server</string>
		<string>start</string>
		<string>--foreground</string>
	</array>
	<key>RunAtLoad</key><true/>
	<key>KeepAlive</key>
	<dict><key>SuccessfulExit</key><false/></dict>
	<key>WorkingDirectory</key><string>%s</string>
	<key>UserName</key><string>%s</string>
	<key>StandardOutPath</key><string>/var/log/{{.AppName}}.out.log</string>
	<key>StandardErrorPath</key><string>/var/log/{{.AppName}}.err.log</string>
</dict>
</plist>
` + "`" + `,
		escapeXML(launchdLabel), escapeXML(binary), escapeXML(workDir), escapeXML(realUser))

	plistPath := launchdPlistPath()
	if err := os.WriteFile(plistPath, []byte(plist), 0o644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}
	if out, err := exec.Command("launchctl", "load", "-w", plistPath).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl load: %s", out)
	}
	fmt.Printf("Service %q installed and enabled\n", launchdLabel)
	return nil
}

func uninstallService(serviceName string) error {
	_ = serviceName
	plistPath := launchdPlistPath()
	_, _ = exec.Command("launchctl", "unload", "-w", plistPath).CombinedOutput()
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove plist: %w", err)
	}
	fmt.Printf("Service %q uninstalled\n", launchdLabel)
	return nil
}
`

// DaemonServerWindowsTemplate generates cmd/{AppName}/server_windows.go.
const DaemonServerWindowsTemplate = `//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"

	"{{.Module}}/internal/serverinfo"
)

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS,
	}
}

func stopProcess(pid int) error {
	pidStr := strconv.Itoa(pid)
	out, err := exec.Command("taskkill", "/PID", pidStr, "/T", "/F").CombinedOutput()
	if err != nil {
		serverinfo.Remove()
		return fmt.Errorf("taskkill PID %d: %s (%w)", pid, strings.TrimSpace(string(out)), err)
	}
	time.Sleep(500 * time.Millisecond)
	serverinfo.Remove()
	return nil
}

func isPrivileged() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

func elevateAndRerun() error {
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	args := strings.Join(os.Args[1:], " ")
	verb, _ := windows.UTF16PtrFromString("runas")
	exe, _ := windows.UTF16PtrFromString(binary)
	params, _ := windows.UTF16PtrFromString(args)
	cwd, _ := windows.UTF16PtrFromString("")
	return windows.ShellExecute(0, verb, exe, params, cwd, windows.SW_NORMAL)
}

func resolveRealUser() string {
	if u := os.Getenv("USERNAME"); u != "" {
		return u
	}
	return "SYSTEM"
}

func installService(serviceName, displayName string) error {
	binary, _ := os.Executable()
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.CreateService(serviceName, binary,
		mgr.Config{
			DisplayName: displayName,
			StartType:   mgr.StartAutomatic,
			Description: displayName,
		},
		"server", "start", "--foreground",
	)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	defer s.Close()

	fmt.Printf("Service %q installed (StartAutomatic)\n", serviceName)
	return nil
}

func uninstallService(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("open service: %w", err)
	}
	defer s.Close()

	if err := s.Delete(); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	fmt.Printf("Service %q uninstalled\n", serviceName)
	return nil
}
`
