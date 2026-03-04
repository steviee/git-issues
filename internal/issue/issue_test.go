package issue

import (
	"strings"
	"testing"
)

func TestFrontmatterRoundTrip(t *testing.T) {
	original := &Issue{
		ID:       7,
		Title:    "Login schlägt bei leeren Passwörtern fehl",
		Status:   "open",
		Priority: "high",
		Labels:   []string{"bug", "auth"},
		Relations: Relations{
			Blocks:    []int{12, 15},
			DependsOn: []int{3},
		},
		Created: "2026-03-04",
		Updated: "2026-03-04",
		Body:    "Freitext Markdown body hier.\n",
	}

	data := Marshal(original)
	parsed, err := Parse(string(data))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if parsed.ID != original.ID {
		t.Errorf("ID: got %d, want %d", parsed.ID, original.ID)
	}
	if parsed.Title != original.Title {
		t.Errorf("Title: got %q, want %q", parsed.Title, original.Title)
	}
	if parsed.Status != original.Status {
		t.Errorf("Status: got %q, want %q", parsed.Status, original.Status)
	}
	if parsed.Priority != original.Priority {
		t.Errorf("Priority: got %q, want %q", parsed.Priority, original.Priority)
	}
	if len(parsed.Labels) != len(original.Labels) {
		t.Errorf("Labels length: got %d, want %d", len(parsed.Labels), len(original.Labels))
	} else {
		for i, l := range parsed.Labels {
			if l != original.Labels[i] {
				t.Errorf("Labels[%d]: got %q, want %q", i, l, original.Labels[i])
			}
		}
	}
	if parsed.Created != original.Created {
		t.Errorf("Created: got %q, want %q", parsed.Created, original.Created)
	}
	if parsed.Updated != original.Updated {
		t.Errorf("Updated: got %q, want %q", parsed.Updated, original.Updated)
	}

	// Relations
	if len(parsed.Relations.Blocks) != len(original.Relations.Blocks) {
		t.Errorf("Relations.Blocks length: got %d, want %d", len(parsed.Relations.Blocks), len(original.Relations.Blocks))
	} else {
		for i, id := range parsed.Relations.Blocks {
			if id != original.Relations.Blocks[i] {
				t.Errorf("Relations.Blocks[%d]: got %d, want %d", i, id, original.Relations.Blocks[i])
			}
		}
	}
	if len(parsed.Relations.DependsOn) != len(original.Relations.DependsOn) {
		t.Errorf("Relations.DependsOn length: got %d, want %d", len(parsed.Relations.DependsOn), len(original.Relations.DependsOn))
	} else {
		for i, id := range parsed.Relations.DependsOn {
			if id != original.Relations.DependsOn[i] {
				t.Errorf("Relations.DependsOn[%d]: got %d, want %d", i, id, original.Relations.DependsOn[i])
			}
		}
	}

	// Body
	if strings.TrimSpace(parsed.Body) != strings.TrimSpace(original.Body) {
		t.Errorf("Body: got %q, want %q", parsed.Body, original.Body)
	}
}

func TestParseWithBody(t *testing.T) {
	content := `---
id: 1
title: "Test issue"
status: open
priority: medium
labels: [bug]
created: 2026-03-04
updated: 2026-03-04
---

This is the body.

## Notes
Some notes here.
`
	issue, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if issue.ID != 1 {
		t.Errorf("ID: got %d, want 1", issue.ID)
	}
	if issue.Title != "Test issue" {
		t.Errorf("Title: got %q, want %q", issue.Title, "Test issue")
	}
	if !strings.Contains(issue.Body, "This is the body.") {
		t.Errorf("Body should contain 'This is the body.', got %q", issue.Body)
	}
	if !strings.Contains(issue.Body, "## Notes") {
		t.Errorf("Body should contain '## Notes', got %q", issue.Body)
	}
	if !strings.Contains(issue.Body, "Some notes here.") {
		t.Errorf("Body should contain 'Some notes here.', got %q", issue.Body)
	}
}

func TestParseWithNoBody(t *testing.T) {
	content := `---
id: 2
title: "No body issue"
status: open
priority: low
labels: []
created: 2026-03-04
updated: 2026-03-04
---
`
	issue, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if issue.ID != 2 {
		t.Errorf("ID: got %d, want 2", issue.ID)
	}
	if issue.Title != "No body issue" {
		t.Errorf("Title: got %q, want %q", issue.Title, "No body issue")
	}
	if issue.Body != "" {
		t.Errorf("Body should be empty, got %q", issue.Body)
	}
}

func TestParseWithRelations(t *testing.T) {
	content := `---
id: 5
title: "Issue with relations"
status: open
priority: high
labels: [feature]
relations:
  blocks: [10, 11]
  depends-on: [2, 3]
created: 2026-03-04
updated: 2026-03-04
---

Body text.
`
	issue, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(issue.Relations.Blocks) != 2 {
		t.Fatalf("Blocks: got %d entries, want 2", len(issue.Relations.Blocks))
	}
	if issue.Relations.Blocks[0] != 10 || issue.Relations.Blocks[1] != 11 {
		t.Errorf("Blocks: got %v, want [10, 11]", issue.Relations.Blocks)
	}

	if len(issue.Relations.DependsOn) != 2 {
		t.Fatalf("DependsOn: got %d entries, want 2", len(issue.Relations.DependsOn))
	}
	if issue.Relations.DependsOn[0] != 2 || issue.Relations.DependsOn[1] != 3 {
		t.Errorf("DependsOn: got %v, want [2, 3]", issue.Relations.DependsOn)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "missing opening delimiter",
			content: "id: 1\ntitle: test\n---\n",
		},
		{
			name:    "missing closing delimiter",
			content: "---\nid: 1\ntitle: test\n",
		},
		{
			name:    "empty content",
			content: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.content)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestValidateInvalidStatus(t *testing.T) {
	issue := &Issue{
		ID:       1,
		Title:    "Test",
		Status:   "invalid-status",
		Priority: "medium",
		Created:  "2026-03-04",
	}
	err := Validate(issue)
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("error should mention 'invalid status', got %q", err.Error())
	}
}

func TestValidateInvalidPriority(t *testing.T) {
	issue := &Issue{
		ID:       1,
		Title:    "Test",
		Status:   "open",
		Priority: "urgent",
		Created:  "2026-03-04",
	}
	err := Validate(issue)
	if err == nil {
		t.Fatal("expected error for invalid priority, got nil")
	}
	if !strings.Contains(err.Error(), "invalid priority") {
		t.Errorf("error should mention 'invalid priority', got %q", err.Error())
	}
}

func TestValidateEmptyTitle(t *testing.T) {
	issue := &Issue{
		ID:       1,
		Title:    "",
		Status:   "open",
		Priority: "medium",
		Created:  "2026-03-04",
	}
	err := Validate(issue)
	if err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error should mention 'title', got %q", err.Error())
	}
}

func TestValidateWhitespaceTitle(t *testing.T) {
	issue := &Issue{
		ID:       1,
		Title:    "   ",
		Status:   "open",
		Priority: "medium",
		Created:  "2026-03-04",
	}
	err := Validate(issue)
	if err == nil {
		t.Fatal("expected error for whitespace-only title, got nil")
	}
}

func TestValidateIDZero(t *testing.T) {
	issue := &Issue{
		ID:       0,
		Title:    "Test",
		Status:   "open",
		Priority: "medium",
		Created:  "2026-03-04",
	}
	err := Validate(issue)
	if err == nil {
		t.Fatal("expected error for ID=0, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("error should mention 'id', got %q", err.Error())
	}
}

func TestValidateIDNegative(t *testing.T) {
	issue := &Issue{
		ID:       -5,
		Title:    "Test",
		Status:   "open",
		Priority: "medium",
		Created:  "2026-03-04",
	}
	err := Validate(issue)
	if err == nil {
		t.Fatal("expected error for negative ID, got nil")
	}
}

func TestValidateValidIssue(t *testing.T) {
	issue := &Issue{
		ID:       1,
		Title:    "Valid issue",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}
	err := Validate(issue)
	if err != nil {
		t.Errorf("expected no error for valid issue, got %v", err)
	}
}

func TestValidateAllStatuses(t *testing.T) {
	for _, status := range ValidStatuses {
		issue := &Issue{
			ID:       1,
			Title:    "Test",
			Status:   status,
			Priority: "medium",
			Created:  "2026-03-04",
		}
		err := Validate(issue)
		if err != nil {
			t.Errorf("expected no error for status %q, got %v", status, err)
		}
	}
}

func TestValidateAllPriorities(t *testing.T) {
	for _, priority := range ValidPriorities {
		issue := &Issue{
			ID:       1,
			Title:    "Test",
			Status:   "open",
			Priority: priority,
			Created:  "2026-03-04",
		}
		err := Validate(issue)
		if err != nil {
			t.Errorf("expected no error for priority %q, got %v", priority, err)
		}
	}
}

func TestPriorityRank(t *testing.T) {
	tests := []struct {
		priority string
		want     int
	}{
		{"critical", 4},
		{"high", 3},
		{"medium", 2},
		{"low", 1},
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := PriorityRank(tt.priority)
		if got != tt.want {
			t.Errorf("PriorityRank(%q): got %d, want %d", tt.priority, got, tt.want)
		}
	}
}

func TestMarshalProducesFrontmatter(t *testing.T) {
	issue := &Issue{
		ID:       3,
		Title:    "Test marshal",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{"bug"},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}

	data := Marshal(issue)
	content := string(data)

	if !strings.HasPrefix(content, "---\n") {
		t.Error("marshaled output should start with ---")
	}

	// Should contain closing ---
	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) < 3 {
		t.Fatalf("expected at least 3 parts when splitting on ---, got %d", len(parts))
	}
}

func TestMarshalWithBody(t *testing.T) {
	issue := &Issue{
		ID:       1,
		Title:    "Test",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
		Body:     "Some body content.\n",
	}

	data := Marshal(issue)
	content := string(data)

	if !strings.Contains(content, "Some body content.") {
		t.Error("marshaled output should contain the body")
	}
}

func TestMarshalWithoutBody(t *testing.T) {
	issue := &Issue{
		ID:       1,
		Title:    "Test",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
		Body:     "",
	}

	data := Marshal(issue)
	content := string(data)

	// Should end with the closing --- and newline, no extra blank lines
	if !strings.HasSuffix(content, "---\n") {
		t.Errorf("marshaled output without body should end with '---\\n', got trailing: %q",
			content[len(content)-10:])
	}
}

func TestValidateInvalidCreatedDate(t *testing.T) {
	issue := &Issue{
		ID:       1,
		Title:    "Test",
		Status:   "open",
		Priority: "medium",
		Created:  "not-a-date",
	}
	err := Validate(issue)
	if err == nil {
		t.Fatal("expected error for invalid created date, got nil")
	}
	if !strings.Contains(err.Error(), "created date") {
		t.Errorf("error should mention 'created date', got %q", err.Error())
	}
}

func TestValidateEmptyCreatedDate(t *testing.T) {
	issue := &Issue{
		ID:       1,
		Title:    "Test",
		Status:   "open",
		Priority: "medium",
		Created:  "",
	}
	err := Validate(issue)
	if err == nil {
		t.Fatal("expected error for empty created date, got nil")
	}
}

func TestParseLabelsDefault(t *testing.T) {
	content := `---
id: 1
title: "Test"
status: open
priority: medium
created: 2026-03-04
updated: 2026-03-04
---
`
	issue, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if issue.Labels == nil {
		t.Error("Labels should be initialized to empty slice, not nil")
	}
	if len(issue.Labels) != 0 {
		t.Errorf("Labels should be empty, got %v", issue.Labels)
	}
}

func TestRoundTripWithClosedField(t *testing.T) {
	original := &Issue{
		ID:       10,
		Title:    "Closed issue",
		Status:   "closed",
		Priority: "low",
		Labels:   []string{},
		Created:  "2026-01-01",
		Updated:  "2026-03-04",
		Closed:   "2026-03-04",
	}

	data := Marshal(original)
	parsed, err := Parse(string(data))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if parsed.Closed != original.Closed {
		t.Errorf("Closed: got %q, want %q", parsed.Closed, original.Closed)
	}
}
