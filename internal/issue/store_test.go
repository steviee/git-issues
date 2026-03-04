package issue

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestIssuesDir(t *testing.T) {
	// Create a temp dir with .issues/ subdirectory
	root := t.TempDir()
	issuesDir := filepath.Join(root, ".issues")
	os.MkdirAll(issuesDir, 0755)

	// Create a nested subdirectory to test walking up
	nested := filepath.Join(root, "src", "pkg")
	os.MkdirAll(nested, 0755)

	// Save original dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Change to nested dir and try to find .issues/
	os.Chdir(nested)
	found, err := IssuesDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Resolve symlinks for macOS (/var → /private/var)
	expectedResolved, _ := filepath.EvalSymlinks(issuesDir)
	foundResolved, _ := filepath.EvalSymlinks(found)
	if foundResolved != expectedResolved {
		t.Errorf("expected %q, got %q", expectedResolved, foundResolved)
	}
}

func TestIssuesDirNotFound(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(dir)
	_, err := IssuesDir()
	if err == nil {
		t.Error("expected error when .issues/ not found")
	}
}

func TestLoadAll(t *testing.T) {
	dir := t.TempDir()
	writeIssueFile(t, dir, 1, "First issue")
	writeIssueFile(t, dir, 2, "Second issue")

	issues, err := LoadAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
}

func TestLoadAllSkipsDotFiles(t *testing.T) {
	dir := t.TempDir()
	writeIssueFile(t, dir, 1, "Normal issue")
	// Write a dotfile that looks like an issue
	os.WriteFile(filepath.Join(dir, ".agent.md"), []byte("---\nid: 99\ntitle: Agent\nstatus: open\npriority: medium\nlabels: []\ncreated: 2026-01-01\nupdated: 2026-01-01\n---\n"), 0644)

	issues, err := LoadAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue (dotfile skipped), got %d", len(issues))
	}
}

func TestLoadByID(t *testing.T) {
	dir := t.TempDir()
	writeIssueFile(t, dir, 1, "First")
	writeIssueFile(t, dir, 2, "Second")

	iss, err := LoadByID(dir, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if iss.Title != "Second" {
		t.Errorf("expected title 'Second', got %q", iss.Title)
	}
}

func TestLoadByIDNotFound(t *testing.T) {
	dir := t.TempDir()
	writeIssueFile(t, dir, 1, "First")

	_, err := LoadByID(dir, 99)
	if err == nil {
		t.Error("expected error for non-existent ID")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	iss := &Issue{
		ID:       1,
		Title:    "Test issue",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-01-01",
		Body:     "Test body",
	}

	err := Save(dir, iss)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Reload and verify
	loaded, err := LoadByID(dir, 1)
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}
	if loaded.Title != "Test issue" {
		t.Errorf("expected title 'Test issue', got %q", loaded.Title)
	}
	if loaded.Updated == "" {
		t.Error("expected updated field to be set")
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	writeIssueFile(t, dir, 1, "To delete")

	err := Delete(dir, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = LoadByID(dir, 1)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func writeIssueFile(t *testing.T, dir string, id int, title string) {
	t.Helper()
	content := "---\n" +
		"id: " + fmt.Sprintf("%d", id) + "\n" +
		"title: \"" + title + "\"\n" +
		"status: open\n" +
		"priority: medium\n" +
		"labels: []\n" +
		"created: 2026-01-01\n" +
		"updated: 2026-01-01\n" +
		"---\n"
	slug := GenerateSlug(title)
	filename := Filename(id, slug)
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

