package pager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PagerOptions configures the pager behavior.
type PagerOptions struct {
	LineNumbers bool // Show line numbers
	NoInit      bool // Don't clear screen on start
	Quit        bool // Quit if content fits on one screen
	IgnoreCase  bool // Case-insensitive search
	Chop        bool // Truncate long lines instead of wrapping
	Raw         bool // Show raw control characters
	Follow      bool // Follow mode (like tail -f)
}

// pagerModel represents the TUI state.
//
//nolint:recvcheck // bubbletea interface requires value receivers for Init/Update/View
type pagerModel struct {
	content     []string
	width       int
	height      int
	offset      int
	searchQuery string
	searching   bool
	searchIdx   int
	matches     []int
	opts        PagerOptions
	filename    string
	quit        bool
	message     string
}

// RunLess executes the less pager.
func RunLess(w io.Writer, args []string, opts PagerOptions) error {
	return runPager(w, args, opts, "less")
}

// RunMore executes the more pager.
func RunMore(w io.Writer, args []string, opts PagerOptions) error {
	// More is traditionally simpler than less
	opts.Quit = true // Quit when reaching end
	return runPager(w, args, opts, "more")
}

func runPager(_ io.Writer, args []string, opts PagerOptions, name string) error {
	// Note: io.Writer is unused because bubbletea manages its own terminal output
	var (
		content  []string
		filename string
	)

	if len(args) == 0 || args[0] == "-" {
		// Read from stdin
		filename = "(stdin)"

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			content = append(content, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	} else {
		// Read from file
		filename = args[0]

		file, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		defer func() {
			_ = file.Close()
		}()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			content = append(content, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	if len(content) == 0 {
		return nil
	}

	model := pagerModel{
		content:  content,
		opts:     opts,
		filename: filename,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()

	return err
}

func (m pagerModel) Init() tea.Cmd {
	return nil
}

func (m pagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 1 // Reserve line for status

		// Check if content fits and should quit
		if m.opts.Quit && len(m.content) <= m.height {
			m.quit = true
			return m, tea.Quit
		}

	case tea.KeyMsg:
		// Handle search input mode
		if m.searching {
			switch msg.String() {
			case "enter":
				m.searching = false
				m.findMatches()

				if len(m.matches) > 0 {
					m.offset = m.matches[0]
					m.searchIdx = 0
					m.message = fmt.Sprintf("Pattern found: %d matches", len(m.matches))
				} else {
					m.message = "Pattern not found"
				}
			case "esc":
				m.searching = false
				m.searchQuery = ""
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
				}
			}

			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit

		case "down", "j", "enter":
			if m.offset < len(m.content)-m.height {
				m.offset++
			}

		case "up", "k":
			if m.offset > 0 {
				m.offset--
			}

		case "pgdown", " ", "ctrl+f":
			m.offset += m.height
			if m.offset > len(m.content)-m.height {
				m.offset = len(m.content) - m.height
			}

			if m.offset < 0 {
				m.offset = 0
			}

		case "pgup", "ctrl+b":
			m.offset -= m.height
			if m.offset < 0 {
				m.offset = 0
			}

		case "home", "g":
			m.offset = 0

		case "end", "G":
			m.offset = max(len(m.content)-m.height, 0)

		case "/":
			m.searching = true
			m.searchQuery = ""
			m.message = ""

		case "n":
			// Next search match
			if len(m.matches) > 0 {
				m.searchIdx = (m.searchIdx + 1) % len(m.matches)
				m.offset = m.matches[m.searchIdx]
			}

		case "N":
			// Previous search match
			if len(m.matches) > 0 {
				m.searchIdx--
				if m.searchIdx < 0 {
					m.searchIdx = len(m.matches) - 1
				}

				m.offset = m.matches[m.searchIdx]
			}

		case "h":
			m.message = "j/k:scroll q:quit /:search n/N:next/prev g/G:top/bottom"
		}
	}

	return m, nil
}

func (m pagerModel) View() string {
	if m.quit {
		return ""
	}

	if m.height == 0 {
		return "Loading..."
	}

	var sb strings.Builder

	// Styles
	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("226")).Foreground(lipgloss.Color("0"))
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Calculate visible range
	start := m.offset

	end := min(m.offset+m.height, len(m.content))

	// Render visible lines
	for i := start; i < end; i++ {
		line := m.content[i]

		// Add line numbers if enabled
		if m.opts.LineNumbers {
			sb.WriteString(lineNumStyle.Render(fmt.Sprintf("%6d ", i+1)))
		}

		// Truncate or wrap long lines
		if m.opts.Chop && len(line) > m.width {
			line = line[:m.width-1] + ">"
		}

		// Highlight search matches
		if m.searchQuery != "" {
			line = highlightSearchMatches(line, m.searchQuery, m.opts.IgnoreCase, highlightStyle)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Pad remaining lines
	for i := end - start; i < m.height; i++ {
		sb.WriteString("~\n")
	}

	// Status line
	statusStyle := lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252"))

	var status string
	if m.searching {
		status = fmt.Sprintf("/%s", m.searchQuery)
	} else if m.message != "" {
		status = m.message
	} else {
		percent := 0
		if len(m.content) > m.height {
			percent = (m.offset * 100) / (len(m.content) - m.height)
		}

		if m.offset == 0 {
			status = fmt.Sprintf(" %s (TOP)", m.filename)
		} else if m.offset >= len(m.content)-m.height {
			status = fmt.Sprintf(" %s (END)", m.filename)
		} else {
			status = fmt.Sprintf(" %s (%d%%)", m.filename, percent)
		}
	}

	// Pad status to full width
	if len(status) < m.width {
		status += strings.Repeat(" ", m.width-len(status))
	}

	sb.WriteString(statusStyle.Render(status))

	return sb.String()
}

func (m *pagerModel) findMatches() {
	m.matches = nil
	if m.searchQuery == "" {
		return
	}

	pattern := m.searchQuery
	if m.opts.IgnoreCase {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return
	}

	for i, line := range m.content {
		if re.MatchString(line) {
			m.matches = append(m.matches, i)
		}
	}
}

func highlightSearchMatches(line, query string, ignoreCase bool, style lipgloss.Style) string {
	pattern := regexp.QuoteMeta(query)
	if ignoreCase {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return line
	}

	return re.ReplaceAllStringFunc(line, func(match string) string {
		return style.Render(match)
	})
}
