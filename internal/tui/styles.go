package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("57")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("236"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	statusSyncedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42"))

	statusLaggingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))

	statusStaleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))

	statusErrorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196")).
				Background(lipgloss.Color("52"))

	detailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("252"))

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15"))

	historyHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("105"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
