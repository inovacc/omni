package runtimeps

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/procmetrics"
	"github.com/inovacc/omni/pkg/procutil"
)

// RunTop launches the bubbletea TUI dashboard for Go processes.
// includeSelf=true shows the omni process itself in the list.
//
// Adapted from github.com/inovacc/gops (MIT) — see THIRD_PARTY_LICENSES/gops-MIT.txt.
func RunTop(ctx context.Context, interval time.Duration, includeSelf bool) error {
	if !isTTY() {
		return cmderr.Wrap(cmderr.ErrUnsupported, "TUI requires a TTY; use `omni gops --json` for non-interactive output")
	}
	if interval <= 0 {
		interval = time.Second
	}
	m := topModel{
		ctx:         ctx,
		col:         procmetrics.NewCollector(),
		interval:    interval,
		includeSelf: includeSelf,
	}
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

type tickMsg time.Time

// dataMsg carries one async collection round-trip.
type dataMsg struct {
	procs   []procutil.Process
	perProc map[int32]procmetrics.Metrics
	err     error
}

type topModel struct {
	ctx         context.Context
	procs       []procutil.Process
	selected    int
	perProc     map[int32]procmetrics.Metrics
	col         *procmetrics.Collector
	interval    time.Duration
	includeSelf bool
	err         error
}

func (m topModel) Init() tea.Cmd { return tick(m.interval) }

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// collectCmd runs the blocking I/O in a goroutine and returns a dataMsg.
func collectCmd(ctx context.Context, col *procmetrics.Collector, interval time.Duration, includeSelf bool) tea.Cmd {
	return func() tea.Msg {
		tctx, cancel := context.WithTimeout(ctx, interval)
		defer cancel()
		procs, err := procutil.List(tctx, procutil.ListOptions{
			Runtime:     procutil.RuntimeGo,
			IncludeSelf: includeSelf,
		})
		if err != nil {
			return dataMsg{err: err}
		}
		perProc := make(map[int32]procmetrics.Metrics, len(procs))
		// Cap per-proc collection to keep tick within interval budget.
		limit := len(procs)
		if limit > 50 {
			limit = 50
		}
		for i := 0; i < limit; i++ {
			if mm, merr := col.Collect(tctx, procs[i].PID); merr == nil {
				perProc[procs[i].PID] = mm
			}
		}
		return dataMsg{procs: procs, perProc: perProc}
	}
}

func (m topModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, tea.Batch(tick(m.interval), collectCmd(m.ctx, m.col, m.interval, m.includeSelf))
	case dataMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			m.procs = msg.procs
			if m.selected >= len(msg.procs) {
				m.selected = 0
			}
			m.perProc = msg.perProc
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.procs)-1 {
				m.selected++
			}
		}
	}
	return m, nil
}

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	rowSelStyle = lipgloss.NewStyle().Reverse(true)
)

func (m topModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v\n(q to quit)", m.err)
	}
	var s strings.Builder
	s.WriteString(headerStyle.Render("omni gops top — q to quit, j/k to navigate") + "\n\n")
	header := fmt.Sprintf("%-7s %-22s %-10s %7s %12s", "PID", "NAME", "GO VER", "CPU%", "MEM RSS")
	s.WriteString(headerStyle.Render(header) + "\n")
	for i, p := range m.procs {
		mm := m.perProc[p.PID]
		line := fmt.Sprintf("%-7d %-22s %-10s %6.1f%% %12s",
			p.PID, truncate(p.Name, 22), p.GoVersion, mm.CPUPercent, humanBytesShort(mm.MemRSS))
		if i == m.selected {
			line = rowSelStyle.Render(line)
		}
		s.WriteString(line + "\n")
	}
	if len(m.procs) > 0 {
		selected := m.perProc[m.procs[m.selected].PID]
		fmt.Fprintf(&s, "\nselected: CPU %.1f%%  RSS %s  Goroutines %d  GC %d\n",
			selected.CPUPercent, humanBytesShort(selected.MemRSS), selected.Goroutines, selected.GCCount)
	}
	return s.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func humanBytesShort(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
