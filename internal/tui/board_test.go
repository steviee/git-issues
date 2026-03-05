package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/steviee/git-issues/internal/config"
	"github.com/steviee/git-issues/internal/issue"
)

func testIssues() []*issue.Issue {
	return []*issue.Issue{
		{ID: 1, Title: "Low open", Status: "open", Priority: "low", Labels: []string{}, Created: "2026-01-01", Updated: "2026-01-01"},
		{ID: 2, Title: "High open", Status: "open", Priority: "high", Labels: []string{"bug"}, Created: "2026-01-02", Updated: "2026-01-02"},
		{ID: 3, Title: "Critical open", Status: "open", Priority: "critical", Labels: []string{"auth"}, Created: "2026-01-03", Updated: "2026-01-03"},
		{ID: 4, Title: "In progress", Status: "in-progress", Priority: "medium", Labels: []string{"feature"}, Created: "2026-01-04", Updated: "2026-01-04"},
		{ID: 5, Title: "Closed issue", Status: "closed", Priority: "low", Labels: []string{}, Created: "2026-01-05", Updated: "2026-01-05", Closed: "2026-01-10"},
		{ID: 6, Title: "Wontfix issue", Status: "wontfix", Priority: "medium", Labels: []string{}, Created: "2026-01-06", Updated: "2026-01-06", Closed: "2026-01-11"},
	}
}

func testModel() Model {
	cfg := config.Default()
	return NewModel(testIssues(), "/tmp/test-issues", cfg)
}

// update is a test helper that sends a key message and returns the updated Model.
func update(m Model, msg tea.Msg) Model {
	updated, _ := m.Update(msg)
	return updated.(Model)
}

func keyMsg(r rune) tea.Msg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestGroupByStatus(t *testing.T) {
	m := testModel()

	// open column should have 3 issues
	if len(m.columns[0].Issues) != 3 {
		t.Errorf("open column: got %d issues, want 3", len(m.columns[0].Issues))
	}
	// in-progress should have 1
	if len(m.columns[1].Issues) != 1 {
		t.Errorf("in-progress column: got %d issues, want 1", len(m.columns[1].Issues))
	}
	// closed should have 1
	if len(m.columns[2].Issues) != 1 {
		t.Errorf("closed column: got %d issues, want 1", len(m.columns[2].Issues))
	}
	// wontfix should have 1
	if len(m.columns[3].Issues) != 1 {
		t.Errorf("wontfix column: got %d issues, want 1", len(m.columns[3].Issues))
	}
}

func TestColumnSortOrder(t *testing.T) {
	m := testModel()

	// open column: critical(3) > high(2) > low(1)
	col := m.columns[0]
	if col.Issues[0].ID != 3 {
		t.Errorf("first issue should be #3 (critical), got #%d", col.Issues[0].ID)
	}
	if col.Issues[1].ID != 2 {
		t.Errorf("second issue should be #2 (high), got #%d", col.Issues[1].ID)
	}
	if col.Issues[2].ID != 1 {
		t.Errorf("third issue should be #1 (low), got #%d", col.Issues[2].ID)
	}
}

func TestNavigationLeftRight(t *testing.T) {
	m := testModel()
	m.width = 120
	m.height = 40

	if m.activeCol != 0 {
		t.Fatalf("initial activeCol should be 0, got %d", m.activeCol)
	}

	// Move right
	m = update(m, keyMsg('l'))
	if m.activeCol != 1 {
		t.Errorf("after 'l': activeCol should be 1, got %d", m.activeCol)
	}

	// Move right again
	m = update(m, keyMsg('l'))
	if m.activeCol != 2 {
		t.Errorf("after 'l' again: activeCol should be 2, got %d", m.activeCol)
	}

	// Move left
	m = update(m, keyMsg('h'))
	if m.activeCol != 1 {
		t.Errorf("after 'h': activeCol should be 1, got %d", m.activeCol)
	}

	// Boundary: move left to 0, then try again
	m = update(m, keyMsg('h'))
	m = update(m, keyMsg('h'))
	if m.activeCol != 0 {
		t.Errorf("should clamp at 0, got %d", m.activeCol)
	}
}

func TestNavigationUpDown(t *testing.T) {
	m := testModel()
	m.width = 120
	m.height = 40

	col := m.columns[0]
	if col.Cursor != 0 {
		t.Fatalf("initial cursor should be 0, got %d", col.Cursor)
	}

	// Move down
	m = update(m, keyMsg('j'))
	if m.columns[0].Cursor != 1 {
		t.Errorf("after 'j': cursor should be 1, got %d", m.columns[0].Cursor)
	}

	// Move down again
	m = update(m, keyMsg('j'))
	if m.columns[0].Cursor != 2 {
		t.Errorf("after 'j' again: cursor should be 2, got %d", m.columns[0].Cursor)
	}

	// Boundary: can't go past last
	m = update(m, keyMsg('j'))
	if m.columns[0].Cursor != 2 {
		t.Errorf("should clamp at 2, got %d", m.columns[0].Cursor)
	}

	// Move up
	m = update(m, keyMsg('k'))
	if m.columns[0].Cursor != 1 {
		t.Errorf("after 'k': cursor should be 1, got %d", m.columns[0].Cursor)
	}
}

func TestMoveIssueRight(t *testing.T) {
	m := testModel()
	m.width = 120
	m.height = 40

	// Select first issue in open (#3 critical)
	selected := m.columns[0].Selected()
	if selected == nil || selected.ID != 3 {
		t.Fatalf("expected #3 selected, got %v", selected)
	}

	// Move right (open -> in-progress)
	cmd := m.moveIssue(1)

	// Check the issue moved
	if len(m.columns[0].Issues) != 2 {
		t.Errorf("open column should have 2 issues after move, got %d", len(m.columns[0].Issues))
	}
	if len(m.columns[1].Issues) != 2 {
		t.Errorf("in-progress column should have 2 issues after move, got %d", len(m.columns[1].Issues))
	}

	// Status should be updated
	found := false
	for _, iss := range m.columns[1].Issues {
		if iss.ID == 3 {
			found = true
			if iss.Status != "in-progress" {
				t.Errorf("moved issue status should be in-progress, got %s", iss.Status)
			}
			if iss.Closed != "" {
				t.Errorf("in-progress issue should not have closed date")
			}
		}
	}
	if !found {
		t.Error("issue #3 not found in in-progress column")
	}

	// cmd should be non-nil (save function)
	if cmd == nil {
		t.Error("moveIssue should return a save command")
	}
}

func TestMoveIssueToClosed(t *testing.T) {
	m := testModel()
	m.width = 120
	m.height = 40

	// Focus in-progress column (index 1)
	m.activeCol = 1
	selected := m.columns[1].Selected()
	if selected == nil || selected.ID != 4 {
		t.Fatalf("expected #4 selected in in-progress, got %v", selected)
	}

	// Move right (in-progress -> closed)
	m.moveIssue(1)

	// Check closed date was set
	for _, iss := range m.columns[2].Issues {
		if iss.ID == 4 {
			if iss.Closed == "" {
				t.Error("closed issue should have closed date")
			}
			if iss.Status != "closed" {
				t.Errorf("status should be closed, got %s", iss.Status)
			}
		}
	}
}

func TestMoveIssueBoundary(t *testing.T) {
	m := testModel()
	m.width = 120
	m.height = 40

	// Try to move left from first column
	cmd := m.moveIssue(-1)
	if cmd != nil {
		t.Error("should not be able to move left from first column")
	}

	// Try to move right from last column
	m.activeCol = 3
	cmd = m.moveIssue(1)
	if cmd != nil {
		t.Error("should not be able to move right from last column")
	}
}

func TestMoveFromEmptyColumn(t *testing.T) {
	cfg := config.Default()
	issues := []*issue.Issue{
		{ID: 1, Title: "Test", Status: "open", Priority: "medium", Labels: []string{}, Created: "2026-01-01", Updated: "2026-01-01"},
	}
	m := NewModel(issues, "/tmp/test", cfg)
	m.width = 120
	m.height = 40

	// Focus empty in-progress column
	m.activeCol = 1
	cmd := m.moveIssue(1)
	if cmd != nil {
		t.Error("should not be able to move from empty column")
	}
}

func TestDetailView(t *testing.T) {
	m := testModel()
	m.width = 120
	m.height = 40

	// Enter detail view
	m = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if !m.showDetail {
		t.Error("Enter should open detail view")
	}

	// Escape closes it
	m = update(m, tea.KeyMsg{Type: tea.KeyEscape})
	if m.showDetail {
		t.Error("Escape should close detail view")
	}
}

func TestColumnAddRemove(t *testing.T) {
	col := NewColumn("open")
	iss1 := &issue.Issue{ID: 1, Title: "A", Status: "open", Priority: "low", Labels: []string{}}
	iss2 := &issue.Issue{ID: 2, Title: "B", Status: "open", Priority: "high", Labels: []string{}}

	col.Add(iss1)
	col.Add(iss2)

	if len(col.Issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(col.Issues))
	}
	// high (#2) should come first
	if col.Issues[0].ID != 2 {
		t.Errorf("expected #2 first (high), got #%d", col.Issues[0].ID)
	}

	removed := col.Remove(2)
	if removed == nil || removed.ID != 2 {
		t.Errorf("expected to remove #2, got %v", removed)
	}
	if len(col.Issues) != 1 {
		t.Errorf("expected 1 issue after remove, got %d", len(col.Issues))
	}

	// Remove non-existent
	removed = col.Remove(99)
	if removed != nil {
		t.Error("removing non-existent should return nil")
	}
}

func TestStatusOrder(t *testing.T) {
	for i, status := range statusOrder {
		idx := statusIndex(status)
		if idx != i {
			t.Errorf("statusIndex(%q) = %d, want %d", status, idx, i)
		}
	}
	if statusIndex("unknown") != -1 {
		t.Error("unknown status should return -1")
	}
}
