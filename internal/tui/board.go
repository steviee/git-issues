package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
	"github.com/steviee/git-issues/internal/config"
	"github.com/steviee/git-issues/internal/git"
	"github.com/steviee/git-issues/internal/issue"
)

// statusOrder defines the column order (left to right).
var statusOrder = []string{"open", "in-progress", "closed", "wontfix"}

// statusIndex returns the index in statusOrder, or -1.
func statusIndex(status string) int {
	for i, s := range statusOrder {
		if s == status {
			return i
		}
	}
	return -1
}

// savedMsg is sent after an issue is saved to disk.
type savedMsg struct {
	err error
}

// reloadMsg signals the board should reload issues from disk.
type reloadMsg struct{}

// FilterFunc is a predicate for filtering issues in the board.
type FilterFunc func(*issue.Issue) bool

// Model is the main bubbletea model for the board.
type Model struct {
	columns    [4]*Column
	activeCol  int
	width      int
	height     int
	keys       KeyMap
	issuesDir  string
	cfg        *config.Config
	filters    []FilterFunc
	showDetail bool
	err        error
	watcher    *fsnotify.Watcher
}

// NewModel creates a board model from the given issues.
func NewModel(issues []*issue.Issue, issuesDir string, cfg *config.Config, filters ...FilterFunc) Model {
	m := Model{
		keys:      DefaultKeyMap(),
		issuesDir: issuesDir,
		cfg:       cfg,
		filters:   filters,
	}

	m.populateColumns(issues)
	return m
}

// populateColumns clears columns and fills them from the issue list.
func (m *Model) populateColumns(issues []*issue.Issue) {
	// Preserve cursor positions per status
	cursors := [4]int{}
	for i := range m.columns {
		if m.columns[i] != nil {
			cursors[i] = m.columns[i].Cursor
		}
	}

	for i, status := range statusOrder {
		m.columns[i] = NewColumn(status)
		for _, iss := range issues {
			if iss.Status == status {
				m.columns[i].Add(iss)
			}
		}
		// Restore cursor (clamped)
		if cursors[i] < len(m.columns[i].Issues) {
			m.columns[i].Cursor = cursors[i]
		}
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return m.watchFiles()
}

// watchFiles starts an fsnotify watcher on the issues directory.
func (m *Model) watchFiles() tea.Cmd {
	if m.issuesDir == "" {
		return nil
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil
	}
	if err := w.Add(m.issuesDir); err != nil {
		w.Close()
		return nil
	}
	m.watcher = w

	return func() tea.Msg {
		// Debounce: wait for events, then coalesce into one reload.
		for {
			select {
			case ev, ok := <-w.Events:
				if !ok {
					return nil
				}
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					// Short debounce to coalesce rapid writes
					time.Sleep(200 * time.Millisecond)
					// Drain any queued events
					for {
						select {
						case _, ok := <-w.Events:
							if !ok {
								return reloadMsg{}
							}
						default:
							return reloadMsg{}
						}
					}
				}
			case _, ok := <-w.Errors:
				if !ok {
					return nil
				}
			}
		}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case savedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil

	case reloadMsg:
		issues, err := issue.LoadAll(m.issuesDir)
		if err == nil {
			// Apply filters
			var filtered []*issue.Issue
			for _, iss := range issues {
				keep := true
				for _, f := range m.filters {
					if !f(iss) {
						keep = false
						break
					}
				}
				if keep {
					filtered = append(filtered, iss)
				}
			}
			m.populateColumns(filtered)
		}
		// Re-start the watcher for the next change
		return m, m.watchFiles()

	case tea.KeyMsg:
		// Detail view: only Esc and quit work
		if m.showDetail {
			switch {
			case key.Matches(msg, m.keys.Escape):
				m.showDetail = false
				return m, nil
			case key.Matches(msg, m.keys.Quit):
				if m.watcher != nil {
					m.watcher.Close()
				}
				return m, tea.Quit
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			if m.watcher != nil {
				m.watcher.Close()
			}
			return m, tea.Quit

		case key.Matches(msg, m.keys.Left):
			if m.activeCol > 0 {
				m.activeCol--
			}

		case key.Matches(msg, m.keys.Right):
			if m.activeCol < 3 {
				m.activeCol++
			}

		case key.Matches(msg, m.keys.Down):
			m.columns[m.activeCol].CursorDown()

		case key.Matches(msg, m.keys.Up):
			m.columns[m.activeCol].CursorUp()

		case key.Matches(msg, m.keys.MoveLeft):
			return m, m.moveIssue(-1)

		case key.Matches(msg, m.keys.MoveRight):
			return m, m.moveIssue(1)

		case key.Matches(msg, m.keys.Enter):
			if m.columns[m.activeCol].Selected() != nil {
				m.showDetail = true
			}
		}
	}

	return m, nil
}

// moveIssue moves the selected issue to an adjacent column.
func (m *Model) moveIssue(direction int) tea.Cmd {
	targetCol := m.activeCol + direction
	if targetCol < 0 || targetCol > 3 {
		return nil
	}

	col := m.columns[m.activeCol]
	iss := col.Selected()
	if iss == nil {
		return nil
	}

	// Remove from current column
	col.Remove(iss.ID)

	// Update status
	newStatus := statusOrder[targetCol]
	iss.Status = newStatus
	if newStatus == "closed" || newStatus == "wontfix" {
		iss.Closed = issue.Today()
	} else {
		iss.Closed = ""
	}

	// Add to target column
	m.columns[targetCol].Add(iss)

	// Save async
	issuesDir := m.issuesDir
	autoStage := m.cfg.AutoStage
	filePath := iss.FilePath
	return func() tea.Msg {
		err := issue.Save(issuesDir, iss)
		if err == nil && autoStage {
			git.Stage(filePath)
		}
		return savedMsg{err: err}
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.showDetail {
		return m.detailView()
	}

	return m.boardView()
}

func (m Model) boardView() string {
	colWidth := m.width / 4
	colHeight := m.height - 3 // room for status bar

	var cols []string
	for i := 0; i < 4; i++ {
		active := i == m.activeCol
		cols = append(cols, m.columns[i].Render(colWidth, colHeight, active))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	help := statusBarStyle.Render(" h/l: column  j/k: navigate  H/L: move issue  Enter: details  q: quit")

	if m.err != nil {
		help = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(
			fmt.Sprintf(" Error: %s", m.err)) + "\n" + help
		m.err = nil // show once
	}

	return board + "\n" + help
}

func (m Model) detailView() string {
	iss := m.columns[m.activeCol].Selected()
	if iss == nil {
		return "No issue selected"
	}

	separator := strings.Repeat("━", m.width-6)

	var b strings.Builder
	b.WriteString(separator + "\n")
	b.WriteString(detailTitleStyle.Render(
		fmt.Sprintf("Issue #%d  ·  %s  ·  %s", iss.ID, iss.Priority, iss.Status)) + "\n")
	b.WriteString(detailTitleStyle.Render(iss.Title) + "\n")
	b.WriteString(separator + "\n")

	if len(iss.Labels) > 0 {
		b.WriteString(detailFieldStyle.Render("Labels:   ") + strings.Join(iss.Labels, ", ") + "\n")
	}
	b.WriteString(detailFieldStyle.Render("Created:  ") + iss.Created + "\n")
	b.WriteString(detailFieldStyle.Render("Updated:  ") + iss.Updated + "\n")
	if iss.Closed != "" {
		b.WriteString(detailFieldStyle.Render("Closed:   ") + iss.Closed + "\n")
	}

	// Relations
	hasRelations := len(iss.Relations.Blocks) > 0 || len(iss.Relations.DependsOn) > 0 ||
		len(iss.Relations.RelatedTo) > 0 || len(iss.Relations.Duplicates) > 0
	if hasRelations {
		b.WriteString("\n" + detailTitleStyle.Render("Relations") + "\n")
		if len(iss.Relations.Blocks) > 0 {
			b.WriteString(detailFieldStyle.Render("  Blocks:     ") + formatIDs(iss.Relations.Blocks) + "\n")
		}
		if len(iss.Relations.DependsOn) > 0 {
			b.WriteString(detailFieldStyle.Render("  Depends on: ") + formatIDs(iss.Relations.DependsOn) + "\n")
		}
		if len(iss.Relations.RelatedTo) > 0 {
			b.WriteString(detailFieldStyle.Render("  Related to: ") + formatIDs(iss.Relations.RelatedTo) + "\n")
		}
		if len(iss.Relations.Duplicates) > 0 {
			b.WriteString(detailFieldStyle.Render("  Duplicates: ") + formatIDs(iss.Relations.Duplicates) + "\n")
		}
	}

	if iss.Body != "" {
		b.WriteString("\n" + separator + "\n\n")
		b.WriteString(iss.Body)
	}

	b.WriteString("\n\n" + statusBarStyle.Render(" Esc: back  q: quit"))

	return detailBorderStyle.Width(m.width - 4).Render(b.String())
}

func formatIDs(ids []int) string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = fmt.Sprintf("#%d", id)
	}
	return strings.Join(strs, ", ")
}
