package ps

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	t.Run("default output", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, Options{})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "PID") {
			t.Errorf("Run() should contain 'PID' header: %s", output)
		}
	})

	t.Run("all processes", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, Options{All: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Run() -a should produce output")
		}
	})

	t.Run("long format", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, Options{Long: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "UID") {
			t.Errorf("Run() -l should contain 'UID' header: %s", output)
		}
	})

	t.Run("full format", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, Options{Full: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Run() -f should produce output")
		}
	})

	t.Run("aux format", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, Options{Aux: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "USER") {
			t.Errorf("Run() aux should contain 'USER' header: %s", output)
		}
	})

	t.Run("no headers", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, Options{NoHeaders: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "PID") && strings.HasPrefix(output, " ") {
			// Check if PID appears as header (at start of line with spaces)
			lines := strings.Split(output, "\n")
			if len(lines) > 0 && strings.Contains(lines[0], "PID") && strings.Contains(lines[0], "TTY") {
				t.Errorf("Run() --no-headers should not have header line: %s", output)
			}
		}
	})

	t.Run("JSON output", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, Options{JSON: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
			t.Errorf("Run() -j should produce JSON array: %s", output)
		}
	})

	t.Run("sort by pid", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, Options{Sort: "pid"})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Run() --sort=pid should produce output")
		}
	})

	t.Run("sort by cpu", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, Options{Sort: "cpu"})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Run() --sort=cpu should produce output")
		}
	})

	t.Run("go only", func(t *testing.T) {
		var buf bytes.Buffer

		// This may return empty if no Go processes are running
		err := Run(&buf, Options{GoOnly: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		// The test itself is a Go process, so there should be at least one
		// But depending on the platform, this may not detect itself
		t.Logf("Run() --go output: %s", buf.String())
	})
}

func TestRunTop(t *testing.T) {
	t.Run("default top", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunTop(&buf, Options{}, 10)
		if err != nil {
			t.Fatalf("RunTop() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "PID") {
			t.Errorf("RunTop() should contain 'PID' header: %s", output)
		}
	})

	t.Run("top 5", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunTop(&buf, Options{}, 5)
		if err != nil {
			t.Fatalf("RunTop() error = %v", err)
		}

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		// Should have header + up to 5 processes
		if len(lines) > 6 {
			t.Errorf("RunTop() n=5 should have at most 6 lines (header + 5), got %d", len(lines))
		}
	})

	t.Run("top JSON", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunTop(&buf, Options{JSON: true}, 5)
		if err != nil {
			t.Fatalf("RunTop() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "[") {
			t.Errorf("RunTop() -j should produce JSON: %s", output)
		}
	})

	t.Run("top sort by mem", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunTop(&buf, Options{Sort: "mem"}, 10)
		if err != nil {
			t.Fatalf("RunTop() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunTop() --sort=mem should produce output")
		}
	})
}

func TestSortProcesses(t *testing.T) {
	procs := []Info{
		{PID: 3, CPU: 10.0, MEM: 5.0, Time: "00:01:00"},
		{PID: 1, CPU: 50.0, MEM: 20.0, Time: "00:05:00"},
		{PID: 2, CPU: 30.0, MEM: 10.0, Time: "00:03:00"},
	}

	t.Run("sort by pid", func(t *testing.T) {
		p := make([]Info, len(procs))
		copy(p, procs)
		sortProcesses(p, "pid")

		if p[0].PID != 1 || p[1].PID != 2 || p[2].PID != 3 {
			t.Errorf("sortProcesses('pid') incorrect order: %v", p)
		}
	})

	t.Run("sort by cpu", func(t *testing.T) {
		p := make([]Info, len(procs))
		copy(p, procs)
		sortProcesses(p, "cpu")

		if p[0].CPU != 50.0 {
			t.Errorf("sortProcesses('cpu') should have highest CPU first: %v", p)
		}
	})

	t.Run("sort by mem", func(t *testing.T) {
		p := make([]Info, len(procs))
		copy(p, procs)
		sortProcesses(p, "mem")

		if p[0].MEM != 20.0 {
			t.Errorf("sortProcesses('mem') should have highest MEM first: %v", p)
		}
	})
}

func TestFilterGoProcesses(t *testing.T) {
	procs := []Info{
		{PID: 1, IsGo: true},
		{PID: 2, IsGo: false},
		{PID: 3, IsGo: true},
	}

	result := filterGoProcesses(procs)

	if len(result) != 2 {
		t.Errorf("filterGoProcesses() should return 2 Go processes, got %d", len(result))
	}

	for _, p := range result {
		if !p.IsGo {
			t.Errorf("filterGoProcesses() returned non-Go process: %v", p)
		}
	}
}

func TestPrintJSON(t *testing.T) {
	procs := []Info{
		{PID: 1, Command: "test", CPU: 1.0, MEM: 2.0},
	}

	var buf bytes.Buffer

	err := printJSON(&buf, procs)
	if err != nil {
		t.Fatalf("printJSON() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"pid": 1`) {
		t.Errorf("printJSON() should contain pid: %s", output)
	}

	if !strings.Contains(output, `"command": "test"`) {
		t.Errorf("printJSON() should contain command: %s", output)
	}
}
