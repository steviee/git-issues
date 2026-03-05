package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Priority colors
	priorityColors = map[string]lipgloss.Color{
		"critical": lipgloss.Color("196"), // red
		"high":     lipgloss.Color("208"), // orange
		"medium":   lipgloss.Color("226"), // yellow
		"low":      lipgloss.Color("248"), // gray
	}

	// Column header styles
	columnHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1)

	activeColumnHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1).
				Foreground(lipgloss.Color("229")). // bright yellow
				Background(lipgloss.Color("63"))   // purple

	// Card styles
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			MarginBottom(0)

	selectedCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")). // purple
				Padding(0, 1).
				MarginBottom(0).
				Bold(true)

	// Column border
	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 0)

	activeColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(0, 0)

	// Labels
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			MarginTop(1)

	// Detail view
	detailBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				Padding(1, 2)

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true)

	detailFieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))
)
