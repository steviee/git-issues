package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/steviee/git-issues/internal/issue"
)

// RenderCard renders an issue as a card for the board.
func RenderCard(iss *issue.Issue, width int, selected bool) string {
	style := cardStyle.Width(width - 2) // account for border
	if selected {
		style = selectedCardStyle.Width(width - 2)
	}

	// Priority indicator
	priColor, ok := priorityColors[iss.Priority]
	if !ok {
		priColor = lipgloss.Color("248")
	}
	priStyle := lipgloss.NewStyle().Foreground(priColor).Bold(true)
	priStr := priStyle.Render(strings.ToUpper(iss.Priority))

	// ID line
	idLine := fmt.Sprintf("#%-4d %s", iss.ID, priStr)

	// Title (truncate to fit)
	maxTitleLen := width - 4
	title := iss.Title
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen-1] + "…"
	}

	// Labels
	var content string
	if len(iss.Labels) > 0 {
		labels := labelStyle.Render(strings.Join(iss.Labels, ", "))
		content = fmt.Sprintf("%s\n%s\n%s", idLine, title, labels)
	} else {
		content = fmt.Sprintf("%s\n%s", idLine, title)
	}

	return style.Render(content)
}
