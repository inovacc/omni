package ss

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestResolvePort(t *testing.T) {
	tests := []struct {
		port     int
		expected string
	}{
		{22, "ssh"},
		{80, "http"},
		{443, "https"},
		{3306, "mysql"},
		{5432, "postgresql"},
		{6379, "redis"},
		{8080, "http-alt"},
		{27017, "mongodb"},
		{9999, "9999"}, // unknown port
		{0, "0"},       // zero port
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := ResolvePort(tt.port)
			if result != tt.expected {
				t.Errorf("ResolvePort(%d) = %q, want %q", tt.port, result, tt.expected)
			}
		})
	}
}

func TestConnStateString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ESTABLISHED", StateEstablished},
		{"established", StateEstablished},
		{"ESTAB", StateEstablished},
		{"LISTEN", StateListen},
		{"listen", StateListen},
		{"TIME_WAIT", StateTimeWait},
		{"TIME-WAIT", StateTimeWait},
		{"TIMEWAIT", StateTimeWait},
		{"CLOSE_WAIT", StateCloseWait},
		{"CLOSE-WAIT", StateCloseWait},
		{"SYN_SENT", StateSynSent},
		{"SYN_RECV", StateSynRecv},
		{"FIN_WAIT1", StateFinWait1},
		{"FIN-WAIT-1", StateFinWait1},
		{"FIN_WAIT2", StateFinWait2},
		{"FIN-WAIT-2", StateFinWait2},
		{"CLOSING", StateClosing},
		{"LAST_ACK", StateLastAck},
		{"CLOSED", StateClosed},
		{"CLOSE", StateClosed},
		{"NONE", StateNone},
		{"", StateNone},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := connStateString(tt.input)
			if result != tt.expected {
				t.Errorf("connStateString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatAddr(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		port     uint32
		numeric  bool
		expected string
	}{
		{"empty_ip", "", 80, true, "*:80"},
		{"wildcard_v4", "0.0.0.0", 80, true, "*:80"},
		{"wildcard_v6", "::", 80, true, "*:80"},
		{"normal_v4_numeric", "192.168.1.1", 8080, true, "192.168.1.1:8080"},
		{"normal_v4_resolve", "192.168.1.1", 80, false, "192.168.1.1:http"},
		{"zero_port", "192.168.1.1", 0, true, "192.168.1.1:*"},
		{"ipv6_addr", "::1", 443, true, "[::1]:443"},
		{"ipv6_resolve", "::1", 443, false, "[::1]:https"},
		{"ipv6_full", "2001:db8::1", 22, false, "[2001:db8::1]:ssh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAddr(tt.ip, tt.port, tt.numeric)
			if result != tt.expected {
				t.Errorf("formatAddr(%q, %d, %v) = %q, want %q", tt.ip, tt.port, tt.numeric, result, tt.expected)
			}
		})
	}
}

func TestGetServiceName(t *testing.T) {
	tests := []struct {
		port     int
		expected string
	}{
		{22, "ssh"},
		{80, "http"},
		{443, "https"},
		{12345, ""},
		{0, ""},
	}

	for _, tt := range tests {
		result := getServiceName(tt.port)
		if result != tt.expected {
			t.Errorf("getServiceName(%d) = %q, want %q", tt.port, result, tt.expected)
		}
	}
}

func TestPrintSockets(t *testing.T) {
	sockets := []Socket{
		{Protocol: "TCP", State: "LISTEN", LocalAddr: "0.0.0.0", LocalPort: 80, RemoteAddr: "0.0.0.0", RemotePort: 0},
		{Protocol: "TCP", State: "ESTABLISHED", LocalAddr: "192.168.1.1", LocalPort: 54321, RemoteAddr: "10.0.0.1", RemotePort: 443},
	}

	t.Run("with_headers", func(t *testing.T) {
		var buf bytes.Buffer
		opts := Options{Numeric: true}
		err := printSockets(&buf, sockets, opts)
		if err != nil {
			t.Fatal(err)
		}
		output := buf.String()
		if !strings.Contains(output, "Proto") {
			t.Error("expected header with Proto")
		}
		if !strings.Contains(output, "LISTEN") {
			t.Error("expected LISTEN state")
		}
		if !strings.Contains(output, "ESTABLISHED") {
			t.Error("expected ESTABLISHED state")
		}
	})

	t.Run("no_headers", func(t *testing.T) {
		var buf bytes.Buffer
		opts := Options{NoHeaders: true, Numeric: true}
		err := printSockets(&buf, sockets, opts)
		if err != nil {
			t.Fatal(err)
		}
		output := buf.String()
		if strings.Contains(output, "Proto") {
			t.Error("expected no header")
		}
	})

	t.Run("extended", func(t *testing.T) {
		var buf bytes.Buffer
		opts := Options{Extended: true, Numeric: true}
		err := printSockets(&buf, sockets, opts)
		if err != nil {
			t.Fatal(err)
		}
		output := buf.String()
		if !strings.Contains(output, "Recv-Q") {
			t.Error("expected extended headers with Recv-Q")
		}
	})

	t.Run("with_process", func(t *testing.T) {
		procSockets := []Socket{
			{Protocol: "TCP", State: "LISTEN", LocalAddr: "0.0.0.0", LocalPort: 80,
				PID: 1234, ProcessName: "nginx"},
		}
		var buf bytes.Buffer
		opts := Options{Processes: true, Numeric: true}
		err := printSockets(&buf, procSockets, opts)
		if err != nil {
			t.Fatal(err)
		}
		output := buf.String()
		if !strings.Contains(output, "nginx(1234)") {
			t.Error("expected process info nginx(1234)")
		}
	})
}

func TestRun_Default(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{All: true, Numeric: true})
	if err != nil {
		t.Fatal(err)
	}
	// Should produce some output (headers at minimum)
	output := buf.String()
	if !strings.Contains(output, "Proto") {
		t.Error("expected Proto header in default output")
	}
}

func TestRun_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{All: true, JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	// Should be valid JSON
	var sockets []Socket
	if err := json.Unmarshal(buf.Bytes(), &sockets); err != nil {
		t.Errorf("expected valid JSON output, got error: %v", err)
	}
}

func TestRun_Summary(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{Summary: true})
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "Total:") {
		t.Error("expected Total: in summary output")
	}
}

func TestRun_SummaryJSON(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{Summary: true, JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	var summary Summary
	if err := json.Unmarshal(buf.Bytes(), &summary); err != nil {
		t.Errorf("expected valid JSON summary, got error: %v", err)
	}
}

func TestRun_TCPOnly(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{TCP: true, All: true, Numeric: true})
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// Check only TCP entries (skip header)
	for _, line := range lines[1:] {
		if !strings.HasPrefix(strings.TrimSpace(line), "TCP") {
			t.Errorf("expected TCP protocol, got line: %s", line)
		}
	}
}

func TestRun_UDPOnly(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{UDP: true, All: true, Numeric: true})
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(strings.TrimSpace(line), "UDP") {
			t.Errorf("expected UDP protocol, got line: %s", line)
		}
	}
}

func TestRun_Listening(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{Listening: true, Numeric: true})
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "LISTEN") {
			t.Errorf("expected only LISTEN state, got line: %s", line)
		}
	}
}

func TestRun_StateFilter(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{All: true, State: "ESTABLISHED", Numeric: true})
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "ESTABLISHED") {
			t.Errorf("expected only ESTABLISHED state, got line: %s", line)
		}
	}
}
