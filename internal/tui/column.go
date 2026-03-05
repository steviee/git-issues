package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/steviee/git-issues/internal/issue"
)

// Column represents a status column on the board.
type Column struct {
	Status string
	Issues []*issue.Issue
	Cursor int
	Offset int // scroll offset for windowing
}

// NewColumn creates a column for the given status.
func NewColumn(status string) *Column {
	return &Column{
		Status: status,
		Issues: nil,
		Cursor: 0,
		Offset: 0,
	}
}

// Add adds an issue to the column in sorted position.
func (c *Column) Add(iss *issue.Issue) {
	c.Issues = append(c.Issues, iss)
	c.sortIssues()
}

// Remove removes an issue by ID and returns it. Returns nil if not found.
func (c *Column) Remove(id int) *issue.Issue {
	for i, iss := range c.Issues {
		if iss.ID == id {
			removed := c.Issues[i]
			c.Issues = append(c.Issues[:i], c.Issues[i+1:]...)
			// Adjust cursor
			if c.Cursor >= len(c.Issues) && c.Cursor > 0 {
				c.Cursor = len(c.Issues) - 1
			}
			c.clampOffset()
			return removed
		}
	}
	return nil
}

// Selected returns the currently selected issue, or nil if empty.
func (c *Column) Selected() *issue.Issue {
	if len(c.Issues) == 0 {
		return nil
	}
	if c.Cursor < 0 || c.Cursor >= len(c.Issues) {
		return nil
	}
	return c.Issues[c.Cursor]
}

// CursorDown moves the cursor down.
func (c *Column) CursorDown() {
	if c.Cursor < len(c.Issues)-1 {
		c.Cursor++
		c.clampOffset()
	}
}

// CursorUp moves the cursor up.
func (c *Column) CursorUp() {
	if c.Cursor > 0 {
		c.Cursor--
		c.clampOffset()
	}
}

func (c *Column) clampOffset() {
	// Will be called with visible count from Render
}

// Render renders the column with the given dimensions.
func (c *Column) Render(width, height int, active bool) string {
	// Header
	title := fmt.Sprintf(" %s (%d) ", strings.ToUpper(c.Status), len(c.Issues))
	headerStyle := columnHeaderStyle
	if active {
		headerStyle = activeColumnHeaderStyle
	}
	header := headerStyle.Width(width - 2).Render(title)

	// Calculate available card space
	innerWidth := width - 2 // column border
	cardHeight := 4         // approx lines per card (border + content)
	visibleCards := (height - 3) / cardHeight
	if visibleCards < 1 {
		visibleCards = 1
	}

	// Adjust offset for scrolling
	if c.Cursor < c.Offset {
		c.Offset = c.Cursor
	}
	if c.Cursor >= c.Offset+visibleCards {
		c.Offset = c.Cursor - visibleCards + 1
	}
	if c.Offset < 0 {
		c.Offset = 0
	}

	// Render visible cards
	var cards []string
	end := c.Offset + visibleCards
	if end > len(c.Issues) {
		end = len(c.Issues)
	}

	for i := c.Offset; i < end; i++ {
		selected := active && i == c.Cursor
		card := RenderCard(c.Issues[i], innerWidth, selected)
		cards = append(cards, card)
	}

	// Scroll indicators
	if c.Offset > 0 {
		cards = append([]string{fmt.Sprintf(" ↑ %d more", c.Offset)}, cards...)
	}
	remaining := len(c.Issues) - end
	if remaining > 0 {
		cards = append(cards, fmt.Sprintf(" ↓ %d more", remaining))
	}

	body := strings.Join(cards, "\n")

	// Pad body to fill height
	bodyLines := strings.Count(body, "\n") + 1
	if body == "" {
		bodyLines = 0
	}
	targetLines := height - 3 // header + border
	for bodyLines < targetLines {
		body += "\n"
		bodyLines++
	}

	content := header + "\n" + body

	colStyle := columnStyle.Width(width).Height(height)
	if active {
		colStyle = activeColumnStyle.Width(width).Height(height)
	}

	return colStyle.Render(content)
}

func (c *Column) sortIssues() {
	sort.SliceStable(c.Issues, func(i, j int) bool {
		ri := issue.PriorityRank(c.Issues[i].Priority)
		rj := issue.PriorityRank(c.Issues[j].Priority)
		if ri != rj {
			return ri > rj
		}
		return c.Issues[i].ID < c.Issues[j].ID
	})
}
