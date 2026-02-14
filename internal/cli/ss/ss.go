package ss

import (
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
	gnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Options configures the ss command behavior
type Options struct {
	All       bool   // -a: display all sockets
	Listening bool   // -l: display listening sockets only
	TCP       bool   // -t: display TCP sockets
	UDP       bool   // -u: display UDP sockets
	Unix      bool   // -x: display Unix sockets
	Processes bool   // -p: show process using socket
	Numeric   bool   // -n: don't resolve service names
	IPv4      bool   // -4: display only IPv4 sockets
	IPv6      bool   // -6: display only IPv6 sockets
	Summary   bool   // -s: print summary statistics
	Extended  bool   // -e: show extended socket info
	NoHeaders bool   // --no-header: don't print headers
	OutputFormat output.Format // output format (text/json/table)
	State     string // state: filter by state (established, listen, etc.)
}

// Socket represents a network socket
type Socket struct {
	Protocol    string `json:"protocol"`
	State       string `json:"state"`
	RecvQ       uint32 `json:"recv_q"`
	SendQ       uint32 `json:"send_q"`
	LocalAddr   string `json:"local_addr"`
	LocalPort   uint32 `json:"local_port"`
	RemoteAddr  string `json:"remote_addr"`
	RemotePort  uint32 `json:"remote_port"`
	PID         int32  `json:"pid,omitempty"`
	ProcessName string `json:"process_name,omitempty"`
	Family      string `json:"family,omitempty"` // ipv4, ipv6
	FD          uint32 `json:"fd,omitempty"`
}

// Summary represents socket statistics summary
type Summary struct {
	TCP   SummaryStats `json:"tcp"`
	UDP   SummaryStats `json:"udp"`
	Unix  SummaryStats `json:"unix,omitzero"`
	RAW   SummaryStats `json:"raw,omitzero"`
	Total int          `json:"total"`
}

// SummaryStats represents stats for a protocol
type SummaryStats struct {
	Total       int `json:"total"`
	Established int `json:"established,omitempty"`
	Listening   int `json:"listening,omitempty"`
	TimeWait    int `json:"time_wait,omitempty"`
	CloseWait   int `json:"close_wait,omitempty"`
	SynSent     int `json:"syn_sent,omitempty"`
	SynRecv     int `json:"syn_recv,omitempty"`
	FinWait1    int `json:"fin_wait1,omitempty"`
	FinWait2    int `json:"fin_wait2,omitempty"`
	Closing     int `json:"closing,omitempty"`
	LastAck     int `json:"last_ack,omitempty"`
}

// State constants
const (
	StateEstablished = "ESTABLISHED"
	StateListen      = "LISTEN"
	StateTimeWait    = "TIME_WAIT"
	StateCloseWait   = "CLOSE_WAIT"
	StateSynSent     = "SYN_SENT"
	StateSynRecv     = "SYN_RECV"
	StateFinWait1    = "FIN_WAIT1"
	StateFinWait2    = "FIN_WAIT2"
	StateClosing     = "CLOSING"
	StateLastAck     = "LAST_ACK"
	StateClosed      = "CLOSED"
	StateNone        = "NONE"
)

// Run executes the ss command
func Run(w io.Writer, opts Options) error {
	// Default to TCP if nothing specified
	if !opts.TCP && !opts.UDP && !opts.Unix {
		opts.TCP = true
		opts.UDP = true
	}

	f := output.New(w, opts.OutputFormat)

	if opts.Summary {
		return printSummary(w, opts, f)
	}

	sockets := getSockets(opts)

	// Enrich with process info if requested
	if opts.Processes {
		enrichWithProcessInfo(sockets)
	}

	if f.IsJSON() {
		return f.Print(sockets)
	}

	return printSockets(w, sockets, opts)
}

func getSockets(opts Options) []Socket {
	var (
		sockets []Socket
		kinds   []string
	)

	if opts.TCP {
		if opts.IPv4 && !opts.IPv6 {
			kinds = append(kinds, "tcp4")
		} else if opts.IPv6 && !opts.IPv4 {
			kinds = append(kinds, "tcp6")
		} else {
			kinds = append(kinds, "tcp")
		}
	}

	if opts.UDP {
		if opts.IPv4 && !opts.IPv6 {
			kinds = append(kinds, "udp4")
		} else if opts.IPv6 && !opts.IPv4 {
			kinds = append(kinds, "udp6")
		} else {
			kinds = append(kinds, "udp")
		}
	}

	if opts.Unix {
		kinds = append(kinds, "unix")
	}

	for _, kind := range kinds {
		conns, err := gnet.Connections(kind)
		if err != nil {
			continue // Skip errors for individual protocols
		}

		for _, conn := range conns {
			state := connStateString(conn.Status)

			// Filter by state
			if opts.State != "" {
				if !strings.EqualFold(state, opts.State) {
					continue
				}
			}

			// Filter listening only
			if opts.Listening && state != StateListen {
				continue
			}

			// Filter non-listening if not -a
			if !opts.All && !opts.Listening {
				if state == StateListen || state == StateNone {
					continue
				}
			}

			family := "ipv4"
			if conn.Family == 10 || conn.Family == 30 { // AF_INET6
				family = "ipv6"
			}

			proto := strings.TrimSuffix(kind, "4")
			proto = strings.TrimSuffix(proto, "6")

			sock := Socket{
				Protocol:   strings.ToUpper(proto),
				State:      state,
				LocalAddr:  conn.Laddr.IP,
				LocalPort:  conn.Laddr.Port,
				RemoteAddr: conn.Raddr.IP,
				RemotePort: conn.Raddr.Port,
				PID:        conn.Pid,
				FD:         conn.Fd,
				Family:     family,
			}

			sockets = append(sockets, sock)
		}
	}

	// Sort by protocol, then state, then local port
	sort.Slice(sockets, func(i, j int) bool {
		if sockets[i].Protocol != sockets[j].Protocol {
			return sockets[i].Protocol < sockets[j].Protocol
		}

		if sockets[i].State != sockets[j].State {
			// Put LISTEN first
			if sockets[i].State == StateListen {
				return true
			}

			if sockets[j].State == StateListen {
				return false
			}

			return sockets[i].State < sockets[j].State
		}

		return sockets[i].LocalPort < sockets[j].LocalPort
	})

	return sockets
}

func enrichWithProcessInfo(sockets []Socket) {
	// Get all processes
	procs, err := process.Processes()
	if err != nil {
		return
	}

	// Build PID -> name map
	pidMap := make(map[int32]string)

	for _, p := range procs {
		name, err := p.Name()
		if err == nil {
			pidMap[p.Pid] = name
		}
	}

	for i := range sockets {
		if sockets[i].PID > 0 {
			if name, ok := pidMap[sockets[i].PID]; ok {
				sockets[i].ProcessName = name
			}
		}
	}
}

func printSockets(w io.Writer, sockets []Socket, opts Options) error {
	if !opts.NoHeaders {
		if opts.Extended {
			_, _ = fmt.Fprintf(w, "%-5s %-12s %6s %6s %-25s %-25s %s\n",
				"Proto", "State", "Recv-Q", "Send-Q", "Local Address:Port", "Remote Address:Port", "Process")
		} else {
			_, _ = fmt.Fprintf(w, "%-5s %-12s %-25s %-25s %s\n",
				"Proto", "State", "Local Address:Port", "Remote Address:Port", "Process")
		}
	}

	for _, sock := range sockets {
		localAddr := formatAddr(sock.LocalAddr, sock.LocalPort, opts.Numeric)
		remoteAddr := formatAddr(sock.RemoteAddr, sock.RemotePort, opts.Numeric)

		proc := ""

		if opts.Processes && sock.PID > 0 {
			if sock.ProcessName != "" {
				proc = fmt.Sprintf("%s(%d)", sock.ProcessName, sock.PID)
			} else {
				proc = fmt.Sprintf("pid=%d", sock.PID)
			}
		}

		if opts.Extended {
			_, _ = fmt.Fprintf(w, "%-5s %-12s %6d %6d %-25s %-25s %s\n",
				sock.Protocol, sock.State, sock.RecvQ, sock.SendQ, localAddr, remoteAddr, proc)
		} else {
			_, _ = fmt.Fprintf(w, "%-5s %-12s %-25s %-25s %s\n",
				sock.Protocol, sock.State, localAddr, remoteAddr, proc)
		}
	}

	return nil
}

func formatAddr(ip string, port uint32, numeric bool) string {
	switch ip {
	case "":
		ip = "*"
	case "0.0.0.0", "::":
		ip = "*"
	}

	portStr := strconv.FormatUint(uint64(port), 10)
	if port == 0 {
		portStr = "*"
	} else if !numeric {
		// Try to resolve service name
		if svc := getServiceName(int(port)); svc != "" {
			portStr = svc
		}
	}

	// Handle IPv6
	if strings.Contains(ip, ":") && ip != "*" {
		return fmt.Sprintf("[%s]:%s", ip, portStr)
	}

	return fmt.Sprintf("%s:%s", ip, portStr)
}

func getServiceName(port int) string {
	// Common ports
	services := map[int]string{
		20:    "ftp-data",
		21:    "ftp",
		22:    "ssh",
		23:    "telnet",
		25:    "smtp",
		53:    "domain",
		80:    "http",
		110:   "pop3",
		143:   "imap",
		443:   "https",
		465:   "smtps",
		587:   "submission",
		993:   "imaps",
		995:   "pop3s",
		3306:  "mysql",
		5432:  "postgresql",
		6379:  "redis",
		8080:  "http-alt",
		27017: "mongodb",
	}

	return services[port]
}

func connStateString(status string) string {
	// gopsutil returns lowercase states
	status = strings.ToUpper(status)
	switch status {
	case "ESTABLISHED", "ESTAB":
		return StateEstablished
	case "LISTEN":
		return StateListen
	case "TIME_WAIT", "TIME-WAIT", "TIMEWAIT":
		return StateTimeWait
	case "CLOSE_WAIT", "CLOSE-WAIT", "CLOSEWAIT":
		return StateCloseWait
	case "SYN_SENT", "SYN-SENT", "SYNSENT":
		return StateSynSent
	case "SYN_RECV", "SYN-RECV", "SYNRECV":
		return StateSynRecv
	case "FIN_WAIT1", "FIN-WAIT-1", "FINWAIT1":
		return StateFinWait1
	case "FIN_WAIT2", "FIN-WAIT-2", "FINWAIT2":
		return StateFinWait2
	case "CLOSING":
		return StateClosing
	case "LAST_ACK", "LAST-ACK", "LASTACK":
		return StateLastAck
	case "CLOSED", "CLOSE":
		return StateClosed
	case "NONE", "":
		return StateNone
	default:
		return status
	}
}

func printSummary(w io.Writer, opts Options, f *output.Formatter) error {
	var summary Summary

	if opts.TCP || (!opts.UDP && !opts.Unix) {
		conns, err := gnet.Connections("tcp")
		if err == nil {
			for _, conn := range conns {
				summary.TCP.Total++

				switch connStateString(conn.Status) {
				case StateEstablished:
					summary.TCP.Established++
				case StateListen:
					summary.TCP.Listening++
				case StateTimeWait:
					summary.TCP.TimeWait++
				case StateCloseWait:
					summary.TCP.CloseWait++
				case StateSynSent:
					summary.TCP.SynSent++
				case StateSynRecv:
					summary.TCP.SynRecv++
				case StateFinWait1:
					summary.TCP.FinWait1++
				case StateFinWait2:
					summary.TCP.FinWait2++
				case StateClosing:
					summary.TCP.Closing++
				case StateLastAck:
					summary.TCP.LastAck++
				}
			}
		}
	}

	if opts.UDP {
		conns, err := gnet.Connections("udp")
		if err == nil {
			summary.UDP.Total = len(conns)
		}
	}

	summary.Total = summary.TCP.Total + summary.UDP.Total + summary.Unix.Total

	if f.IsJSON() {
		return f.Print(summary)
	}

	_, _ = fmt.Fprintf(w, "Total: %d\n\n", summary.Total)

	if opts.TCP || (!opts.UDP && !opts.Unix) {
		_, _ = fmt.Fprintf(w, "TCP:   %d (estab %d, listen %d, time_wait %d, close_wait %d)\n",
			summary.TCP.Total, summary.TCP.Established, summary.TCP.Listening,
			summary.TCP.TimeWait, summary.TCP.CloseWait)
	}

	if opts.UDP {
		_, _ = fmt.Fprintf(w, "UDP:   %d\n", summary.UDP.Total)
	}

	return nil
}

// ResolvePort tries to resolve a port number to a service name
func ResolvePort(port int) string {
	if svc := getServiceName(port); svc != "" {
		return svc
	}

	_, err := net.LookupPort("tcp", strconv.Itoa(port))
	if err == nil {
		return strconv.Itoa(port)
	}

	return strconv.Itoa(port)
}
