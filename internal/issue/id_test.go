package issue

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTestIssue creates a minimal issue file in the given directory.
func writeTestIssue(t *testing.T, dir string, id int, filename string) {
	t.Helper()
	content := []byte("---\n" +
		"id: " + itoa(id) + "\n" +
		"title: \"Test\"\n" +
		"status: open\n" +
		"priority: medium\n" +
		"labels: []\n" +
		"created: 2026-01-01\n" +
		"updated: 2026-01-01\n" +
		"---\n")
	err := os.WriteFile(filepath.Join(dir, filename), content, 0644)
	if err != nil {
		t.Fatalf("failed to write test issue file: %v", err)
	}
}

// itoa converts an int to its string representation without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func TestNextID_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	got, err := NextID(dir)
	if err != nil {
		t.Fatalf("NextID() returned error: %v", err)
	}
	if got != 1 {
		t.Errorf("NextID() on empty dir = %d, want 1", got)
	}
}

func TestNextID_GapsInIDs(t *testing.T) {
	dir := t.TempDir()

	writeTestIssue(t, dir, 1, "0001-first.md")
	writeTestIssue(t, dir, 3, "0003-third.md")
	writeTestIssue(t, dir, 5, "0005-fifth.md")

	got, err := NextID(dir)
	if err != nil {
		t.Fatalf("NextID() returned error: %v", err)
	}
	if got != 6 {
		t.Errorf("NextID() with IDs 1,3,5 = %d, want 6", got)
	}
}

func TestNextID_SingleIssue(t *testing.T) {
	dir := t.TempDir()

	writeTestIssue(t, dir, 1, "0001-only-issue.md")

	got, err := NextID(dir)
	if err != nil {
		t.Fatalf("NextID() returned error: %v", err)
	}
	if got != 2 {
		t.Errorf("NextID() with single ID=1 = %d, want 2", got)
	}
}

func TestNextID_IgnoresNonMdFiles(t *testing.T) {
	dir := t.TempDir()

	writeTestIssue(t, dir, 10, "0010-real-issue.md")

	// Write a non-.md file that should be ignored
	err := os.WriteFile(filepath.Join(dir, ".config.yml"), []byte("auto_stage: true\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	got, err := NextID(dir)
	if err != nil {
		t.Fatalf("NextID() returned error: %v", err)
	}
	if got != 11 {
		t.Errorf("NextID() = %d, want 11", got)
	}
}

func TestNextID_IgnoresDotFiles(t *testing.T) {
	dir := t.TempDir()

	writeTestIssue(t, dir, 5, "0005-real.md")

	// Write a hidden .md file that should be ignored
	err := os.WriteFile(filepath.Join(dir, ".agent.md"), []byte("# agent context\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write hidden file: %v", err)
	}

	got, err := NextID(dir)
	if err != nil {
		t.Fatalf("NextID() returned error: %v", err)
	}
	if got != 6 {
		t.Errorf("NextID() = %d, want 6", got)
	}
}
