package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aygp-dr/realtime-sync-orchestrator/internal/sync"
)

type view int

const (
	viewDashboard view = iota
	viewDetail
	viewHelp
)

type tickMsg time.Time

// Model is the Bubble Tea model for the sync dashboard.
type Model struct {
	sources      []sync.Source
	cursor       int
	currentView  view
	rng          *rand.Rand
	tickInterval time.Duration
	width        int
	height       int
}

// NewModel creates a new TUI model with mock data.
func NewModel(seed int64, tickInterval time.Duration) Model {
	rng := rand.New(rand.NewSource(seed))
	return Model{
		sources:      sync.MockSources(rng),
		rng:          rng,
		tickInterval: tickInterval,
	}
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return tickCmd(m.tickInterval)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		for i := range m.sources {
			m.sources[i] = sync.RefreshSource(m.rng, m.sources[i])
		}
		return m, tickCmd(m.tickInterval)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		if m.currentView == viewDashboard {
			return m, tea.Quit
		}
		m.currentView = viewDashboard
		return m, nil
	case "j", "down":
		if m.currentView == viewDashboard && m.cursor < len(m.sources)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.currentView == viewDashboard && m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if m.currentView == viewDashboard {
			m.currentView = viewDetail
		}
	case "esc", "escape", "backspace":
		m.currentView = viewDashboard
	case "?":
		if m.currentView == viewHelp {
			m.currentView = viewDashboard
		} else {
			m.currentView = viewHelp
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.currentView {
	case viewDetail:
		return m.viewDetail()
	case viewHelp:
		return m.viewHelp()
	default:
		return m.viewDashboard()
	}
}

func (m Model) viewDashboard() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" RealTimeSyncOrchestrator "))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf(" %d sources | refreshing every %s", len(m.sources), m.tickInterval)))
	b.WriteString("\n\n")

	// Table header
	header := fmt.Sprintf("  %-20s %-20s %8s  %-20s %-8s %14s",
		"SOURCE", "TARGET", "LAG(ms)", "LAST SYNC", "STATUS", "RECORDS BEHIND")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", 96)))
	b.WriteString("\n")

	for i, s := range m.sources {
		row := m.formatRow(s, i == m.cursor)
		b.WriteString(row)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  j/k: navigate  enter: detail  ?: help  q: quit"))
	return b.String()
}

func (m Model) formatRow(s sync.Source, selected bool) string {
	cursor := "  "
	if selected {
		cursor = "> "
	}

	lagStr := fmt.Sprintf("%d", s.LagMs)
	if s.LagMs < 0 {
		lagStr = "ERR"
	}

	elapsed := time.Since(s.LastSync)
	lastSyncStr := formatElapsed(elapsed)

	statusStr := renderStatus(s.Status)

	recordsStr := fmt.Sprintf("%d", s.RecordsBehind)

	row := fmt.Sprintf("%s%-20s %-20s %8s  %-20s %s %14s",
		cursor, s.Name, s.Target, lagStr, lastSyncStr,
		statusStr, recordsStr)

	if selected {
		return selectedStyle.Render(row)
	}
	return row
}

func (m Model) viewDetail() string {
	if m.cursor >= len(m.sources) {
		return "No source selected"
	}
	s := m.sources[m.cursor]

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf(" %s -> %s ", s.Name, s.Target)))
	b.WriteString("\n\n")

	lines := []struct{ label, value string }{
		{"Source:", s.Name},
		{"Target:", s.Target},
		{"Lag:", fmt.Sprintf("%d ms", s.LagMs)},
		{"Status:", string(s.Status)},
		{"Last Sync:", s.LastSync.Format("15:04:05")},
		{"Records Behind:", fmt.Sprintf("%d", s.RecordsBehind)},
	}
	for _, l := range lines {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			detailLabelStyle.Render(fmt.Sprintf("%-16s", l.label)),
			detailValueStyle.Render(l.value)))
	}

	b.WriteString("\n")
	b.WriteString(historyHeaderStyle.Render("  Sync History (recent)"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("  %-20s %8s %-8s %10s", "TIME", "LAG(ms)", "STATUS", "DURATION")))
	b.WriteString("\n")

	start := 0
	if len(s.History) > 15 {
		start = len(s.History) - 15
	}
	for _, h := range s.History[start:] {
		lagStr := fmt.Sprintf("%d", h.LagMs)
		if h.LagMs < 0 {
			lagStr = "ERR"
		}
		b.WriteString(fmt.Sprintf("  %-20s %8s %s %10s\n",
			h.Timestamp.Format("15:04:05"),
			lagStr,
			renderStatusPadded(h.Status),
			h.Duration.Round(time.Millisecond)))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  esc/q: back  ?: help"))
	return b.String()
}

func (m Model) viewHelp() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" Help "))
	b.WriteString("\n\n")

	keys := []struct{ key, desc string }{
		{"j / k / ↑ / ↓", "Navigate sources"},
		{"enter", "View source detail"},
		{"esc / backspace / q", "Back to dashboard"},
		{"?", "Toggle help"},
		{"q (dashboard)", "Quit"},
		{"ctrl+c", "Force quit"},
	}
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			detailLabelStyle.Render(fmt.Sprintf("%-22s", k.key)),
			detailValueStyle.Render(k.desc)))
	}

	b.WriteString("\n")
	b.WriteString(historyHeaderStyle.Render("  Status Legend"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s  lag ≤ 100ms\n", statusSyncedStyle.Render("synced ")))
	b.WriteString(fmt.Sprintf("  %s  100ms < lag ≤ 5000ms\n", statusLaggingStyle.Render("lagging")))
	b.WriteString(fmt.Sprintf("  %s  lag > 5000ms or last sync > 60s ago\n", statusStaleStyle.Render("stale  ")))
	b.WriteString(fmt.Sprintf("  %s  negative lag (connection failure)\n", statusErrorStyle.Render("error  ")))

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  esc/q: back"))
	return b.String()
}

// Sources returns the current sources (used for JSON output).
func (m Model) Sources() []sync.Source {
	return m.sources
}

func renderStatus(s sync.Status) string {
	switch s {
	case sync.StatusSynced:
		return statusSyncedStyle.Render("synced ")
	case sync.StatusLagging:
		return statusLaggingStyle.Render("lagging")
	case sync.StatusStale:
		return statusStaleStyle.Render("stale  ")
	case sync.StatusError:
		return statusErrorStyle.Render("error  ")
	default:
		return string(s)
	}
}

func renderStatusPadded(s sync.Status) string {
	switch s {
	case sync.StatusSynced:
		return statusSyncedStyle.Render(fmt.Sprintf("%-8s", "synced"))
	case sync.StatusLagging:
		return statusLaggingStyle.Render(fmt.Sprintf("%-8s", "lagging"))
	case sync.StatusStale:
		return statusStaleStyle.Render(fmt.Sprintf("%-8s", "stale"))
	case sync.StatusError:
		return statusErrorStyle.Render(fmt.Sprintf("%-8s", "error"))
	default:
		return fmt.Sprintf("%-8s", s)
	}
}

func formatElapsed(d time.Duration) string {
	if d < time.Second {
		return "just now"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm%ds ago", int(d.Minutes()), int(d.Seconds())%60)
}
