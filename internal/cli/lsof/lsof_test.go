package lsof

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	gnet "github.com/shirou/gopsutil/v3/net"
)

func TestFormatConnection(t *testing.T) {
	tests := []struct {
		name     string
		conn     gnet.ConnectionStat
		expected string
	}{
		{
			name: "normal_connection",
			conn: gnet.ConnectionStat{
				Laddr:  gnet.Addr{IP: "192.168.1.1", Port: 8080},
				Raddr:  gnet.Addr{IP: "10.0.0.1", Port: 443},
				Status: "ESTABLISHED",
			},
			expected: "192.168.1.1:8080->10.0.0.1:443 (ESTABLISHED)",
		},
		{
			name: "wildcard_local",
			conn: gnet.ConnectionStat{
				Laddr:  gnet.Addr{IP: "0.0.0.0", Port: 80},
				Raddr:  gnet.Addr{IP: "0.0.0.0", Port: 0},
				Status: "LISTEN",
			},
			expected: "*:80->*:0 (LISTEN)",
		},
		{
			name: "empty_addrs",
			conn: gnet.ConnectionStat{
				Laddr:  gnet.Addr{IP: "", Port: 53},
				Raddr:  gnet.Addr{IP: "", Port: 0},
				Status: "",
			},
			expected: "*:53->*:0 (NONE)",
		},
		{
			name: "ipv6_wildcard",
			conn: gnet.ConnectionStat{
				Laddr:  gnet.Addr{IP: "::", Port: 443},
				Raddr:  gnet.Addr{IP: "::", Port: 0},
				Status: "LISTEN",
			},
			expected: "*:443->*:0 (LISTEN)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatConnection(tt.conn)
			if result != tt.expected {
				t.Errorf("formatConnection() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPrintFiles(t *testing.T) {
	files := []OpenFile{
		{
			Command:  "nginx",
			PID:      1234,
			User:     "root",
			FD:       "10u",
			Type:     "IPv4",
			Node:     "TCP",
			Name:     "192.168.1.1:80->10.0.0.1:54321 (ESTABLISHED)",
			Protocol: "TCP",
		},
		{
			Command:  "postgres",
			PID:      5678,
			User:     "postgres",
			FD:       "5u",
			Type:     "IPv4",
			Node:     "TCP",
			Name:     "*:5432->*:0 (LISTEN)",
			Protocol: "TCP",
		},
	}

	t.Run("with_headers", func(t *testing.T) {
		var buf bytes.Buffer
		err := printFiles(&buf, files, Options{})
		if err != nil {
			t.Fatal(err)
		}
		output := buf.String()
		if !strings.Contains(output, "COMMAND") {
			t.Error("expected COMMAND header")
		}
		if !strings.Contains(output, "nginx") {
			t.Error("expected nginx in output")
		}
		if !strings.Contains(output, "postgres") {
			t.Error("expected postgres in output")
		}
	})

	t.Run("no_headers", func(t *testing.T) {
		var buf bytes.Buffer
		err := printFiles(&buf, files, Options{NoHeaders: true})
		if err != nil {
			t.Fatal(err)
		}
		output := buf.String()
		if strings.Contains(output, "COMMAND") {
			t.Error("expected no header")
		}
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 2 {
			t.Errorf("expected 2 data lines, got %d", len(lines))
		}
	})

	t.Run("long_command_truncated", func(t *testing.T) {
		longFiles := []OpenFile{
			{
				Command: "very-long-command-name-exceeds",
				PID:     1,
				User:    "user",
				FD:      "1u",
				Type:    "IPv4",
				Node:    "TCP",
				Name:    "*:80->*:0 (LISTEN)",
			},
		}
		var buf bytes.Buffer
		err := printFiles(&buf, longFiles, Options{NoHeaders: true})
		if err != nil {
			t.Fatal(err)
		}
		output := buf.String()
		// Command should be truncated to 16 chars
		if strings.Contains(output, "very-long-command-name-exceeds") {
			t.Error("expected command to be truncated to 16 chars")
		}
	})
}

func TestRun_Default(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{})
	if err != nil {
		t.Fatal(err)
	}
	// Should have header at minimum
	output := buf.String()
	if !strings.Contains(output, "COMMAND") {
		t.Error("expected COMMAND header in default output")
	}
}

func TestRun_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	var files []OpenFile
	if err := json.Unmarshal(buf.Bytes(), &files); err != nil {
		t.Errorf("expected valid JSON output, got error: %v", err)
	}
}

func TestRun_PortFilter(t *testing.T) {
	var buf bytes.Buffer
	// Use a port unlikely to have connections
	err := Run(&buf, Options{Port: 59999, JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	var files []OpenFile
	if err := json.Unmarshal(buf.Bytes(), &files); err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}
	// All results should match the port
	for _, f := range files {
		if f.LocalPort != 59999 && f.RemotePort != 59999 {
			t.Errorf("expected port 59999, got local=%d remote=%d", f.LocalPort, f.RemotePort)
		}
	}
}

func TestRunByPort(t *testing.T) {
	var buf bytes.Buffer
	err := RunByPort(&buf, 80, Options{JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	var files []OpenFile
	if err := json.Unmarshal(buf.Bytes(), &files); err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}
}

func TestRunByPID(t *testing.T) {
	var buf bytes.Buffer
	// Use own PID
	err := RunByPID(&buf, 1, Options{JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	var files []OpenFile
	if err := json.Unmarshal(buf.Bytes(), &files); err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}
}

func TestRun_ProtocolFilter(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{Protocol: "tcp", JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	var files []OpenFile
	if err := json.Unmarshal(buf.Bytes(), &files); err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}
	for _, f := range files {
		if f.Protocol != "TCP" {
			t.Errorf("expected TCP protocol, got %s", f.Protocol)
		}
	}
}

func TestRun_IPv4Only(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, Options{IPv4: true, JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	var files []OpenFile
	if err := json.Unmarshal(buf.Bytes(), &files); err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}
	for _, f := range files {
		if f.Type != "IPv4" {
			t.Errorf("expected IPv4 type, got %s", f.Type)
		}
	}
}
