package lsof

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
	gnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Options configures the lsof command behavior
type Options struct {
	PID          int           // -p: show files for specific PID
	User         string        // -u: show files for specific user
	Port         int           // -i: show files using specific port
	Protocol     string        // -i: filter by protocol (tcp, udp)
	Network      bool          // -i: show network files only
	Files        bool          // show file descriptors (default behavior)
	Command      string        // -c: filter by command name prefix
	NoHeaders    bool          // -n: don't print headers
	OutputFormat output.Format // output format (text/json/table)
	IPv4         bool          // -4: show only IPv4
	IPv6         bool          // -6: show only IPv6
	Listen       bool          // show only listening sockets
	Established  bool          // show only established connections
}

// OpenFile represents an open file or network connection
type OpenFile struct {
	Command    string `json:"command"`
	PID        int32  `json:"pid"`
	User       string `json:"user"`
	FD         string `json:"fd"`
	Type       string `json:"type"`
	Device     string `json:"device,omitempty"`
	Size       int64  `json:"size,omitempty"`
	Node       string `json:"node,omitempty"`
	Name       string `json:"name"`
	State      string `json:"state,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	LocalIP    string `json:"local_ip,omitempty"`
	LocalPort  uint32 `json:"local_port,omitempty"`
	RemoteIP   string `json:"remote_ip,omitempty"`
	RemotePort uint32 `json:"remote_port,omitempty"`
}

// Run executes the lsof command
func Run(w io.Writer, opts Options) error {
	// Get network connections (default behavior)
	// Note: Full file descriptor listing requires platform-specific code
	// For now, we focus on network connections which is cross-platform
	files, err := getNetworkFiles(opts)
	if err != nil {
		return fmt.Errorf("lsof: %w", err)
	}

	// Sort by PID then FD
	sort.Slice(files, func(i, j int) bool {
		if files[i].PID != files[j].PID {
			return files[i].PID < files[j].PID
		}

		return files[i].FD < files[j].FD
	})

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(files)
	}

	return printFiles(w, files, opts)
}

func getNetworkFiles(opts Options) ([]OpenFile, error) {
	var files []OpenFile

	// Get process info map
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	nameMap := make(map[int32]string)
	userMap := make(map[int32]string)

	for _, p := range procs {
		if name, err := p.Name(); err == nil {
			nameMap[p.Pid] = name
		}

		if user, err := p.Username(); err == nil {
			userMap[p.Pid] = user
		}
	}

	// Determine which protocols to query
	protocols := []string{"tcp", "udp"}
	if opts.Protocol != "" {
		protocols = []string{strings.ToLower(opts.Protocol)}
	}

	for _, proto := range protocols {
		conns, err := gnet.Connections(proto)
		if err != nil {
			continue
		}

		for _, conn := range conns {
			// Filter by PID
			if opts.PID > 0 && int(conn.Pid) != opts.PID {
				continue
			}

			// Filter by user
			if opts.User != "" {
				if user, ok := userMap[conn.Pid]; !ok || !strings.EqualFold(user, opts.User) {
					continue
				}
			}

			// Filter by command
			if opts.Command != "" {
				if name, ok := nameMap[conn.Pid]; !ok || !strings.HasPrefix(strings.ToLower(name), strings.ToLower(opts.Command)) {
					continue
				}
			}

			// Filter by port
			if opts.Port > 0 {
				if conn.Laddr.Port != uint32(opts.Port) && conn.Raddr.Port != uint32(opts.Port) {
					continue
				}
			}

			// Filter by IP version
			if opts.IPv4 && (conn.Family == 10 || conn.Family == 30) {
				continue
			}

			if opts.IPv6 && conn.Family == 2 {
				continue
			}

			// Filter by state
			state := strings.ToUpper(conn.Status)
			if opts.Listen && state != "LISTEN" {
				continue
			}

			if opts.Established && state != "ESTABLISHED" {
				continue
			}

			name := nameMap[conn.Pid]
			if name == "" {
				name = "?"
			}

			user := userMap[conn.Pid]
			if user == "" {
				user = "?"
			}

			// Format the connection name
			connName := formatConnection(conn)

			fdStr := strconv.FormatUint(uint64(conn.Fd), 10) + "u"

			fileType := "IPv4"
			if conn.Family == 10 || conn.Family == 30 {
				fileType = "IPv6"
			}

			file := OpenFile{
				Command:    name,
				PID:        conn.Pid,
				User:       user,
				FD:         fdStr,
				Type:       fileType,
				Node:       strings.ToUpper(proto),
				Name:       connName,
				State:      state,
				Protocol:   strings.ToUpper(proto),
				LocalIP:    conn.Laddr.IP,
				LocalPort:  conn.Laddr.Port,
				RemoteIP:   conn.Raddr.IP,
				RemotePort: conn.Raddr.Port,
			}

			files = append(files, file)
		}
	}

	return files, nil
}

func formatConnection(conn gnet.ConnectionStat) string {
	localAddr := conn.Laddr.IP
	if localAddr == "" || localAddr == "0.0.0.0" || localAddr == "::" {
		localAddr = "*"
	}

	remoteAddr := conn.Raddr.IP
	if remoteAddr == "" || remoteAddr == "0.0.0.0" || remoteAddr == "::" {
		remoteAddr = "*"
	}

	localPort := conn.Laddr.Port
	remotePort := conn.Raddr.Port

	local := fmt.Sprintf("%s:%d", localAddr, localPort)
	remote := fmt.Sprintf("%s:%d", remoteAddr, remotePort)

	state := conn.Status
	if state == "" {
		state = "NONE"
	}

	return fmt.Sprintf("%s->%s (%s)", local, remote, strings.ToUpper(state))
}

func printFiles(w io.Writer, files []OpenFile, opts Options) error {
	if !opts.NoHeaders {
		_, _ = fmt.Fprintf(w, "%-16s %7s %10s %4s %6s %8s %s\n",
			"COMMAND", "PID", "USER", "FD", "TYPE", "NODE", "NAME")
	}

	for _, f := range files {
		cmd := f.Command
		if len(cmd) > 16 {
			cmd = cmd[:16]
		}

		user := f.User
		if len(user) > 10 {
			user = user[:10]
		}

		_, _ = fmt.Fprintf(w, "%-16s %7d %10s %4s %6s %8s %s\n",
			cmd, f.PID, user, f.FD, f.Type, f.Node, f.Name)
	}

	return nil
}

// RunByPort shows open files for a specific port (convenience function)
func RunByPort(w io.Writer, port int, opts Options) error {
	opts.Port = port
	opts.Network = true

	return Run(w, opts)
}

// RunByPID shows open files for a specific process (convenience function)
func RunByPID(w io.Writer, pid int, opts Options) error {
	opts.PID = pid
	return Run(w, opts)
}
