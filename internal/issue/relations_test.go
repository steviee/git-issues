package issue

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/steviee/git-issues/internal/config"
)

func TestInverse(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"blocks", "depends-on"},
		{"depends-on", "blocks"},
		{"related-to", "related-to"},
		{"duplicates", "duplicates"},
	}

	for _, tt := range tests {
		got := Inverse(tt.input)
		if got != tt.want {
			t.Errorf("Inverse(%q): got %q, want %q", tt.input, got, tt.want)
		}
	}
}

// createTestIssueFile writes a minimal issue markdown file to the given directory.
func createTestIssueFile(t *testing.T, dir string, issue *Issue) {
	t.Helper()
	slug := GenerateSlug(issue.Title)
	filename := Filename(issue.ID, slug)
	issue.FilePath = filepath.Join(dir, filename)
	data := Marshal(issue)
	if err := os.WriteFile(issue.FilePath, data, 0644); err != nil {
		t.Fatalf("failed to write test issue file: %v", err)
	}
}

func TestAddRelationBlocksBothSides(t *testing.T) {
	tmpDir := t.TempDir()

	issue7 := &Issue{
		ID:       7,
		Title:    "Source issue",
		Status:   "open",
		Priority: "high",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}
	issue12 := &Issue{
		ID:       12,
		Title:    "Target issue",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}

	createTestIssueFile(t, tmpDir, issue7)
	createTestIssueFile(t, tmpDir, issue12)

	cfg := &config.Config{AutoStage: false}

	err := AddRelation(tmpDir, 7, "blocks", 12, cfg)
	if err != nil {
		t.Fatalf("AddRelation returned error: %v", err)
	}

	// Reload and verify source
	source, err := LoadByID(tmpDir, 7)
	if err != nil {
		t.Fatalf("LoadByID(7) returned error: %v", err)
	}
	if len(source.Relations.Blocks) != 1 || source.Relations.Blocks[0] != 12 {
		t.Errorf("source.Relations.Blocks: got %v, want [12]", source.Relations.Blocks)
	}

	// Reload and verify target (inverse: depends-on)
	target, err := LoadByID(tmpDir, 12)
	if err != nil {
		t.Fatalf("LoadByID(12) returned error: %v", err)
	}
	if len(target.Relations.DependsOn) != 1 || target.Relations.DependsOn[0] != 7 {
		t.Errorf("target.Relations.DependsOn: got %v, want [7]", target.Relations.DependsOn)
	}
}

func TestAddRelationDependsOnBothSides(t *testing.T) {
	tmpDir := t.TempDir()

	issue1 := &Issue{
		ID:       1,
		Title:    "Depends source",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}
	issue2 := &Issue{
		ID:       2,
		Title:    "Depends target",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}

	createTestIssueFile(t, tmpDir, issue1)
	createTestIssueFile(t, tmpDir, issue2)

	cfg := &config.Config{AutoStage: false}

	err := AddRelation(tmpDir, 1, "depends-on", 2, cfg)
	if err != nil {
		t.Fatalf("AddRelation returned error: %v", err)
	}

	source, err := LoadByID(tmpDir, 1)
	if err != nil {
		t.Fatalf("LoadByID(1) returned error: %v", err)
	}
	if len(source.Relations.DependsOn) != 1 || source.Relations.DependsOn[0] != 2 {
		t.Errorf("source.Relations.DependsOn: got %v, want [2]", source.Relations.DependsOn)
	}

	target, err := LoadByID(tmpDir, 2)
	if err != nil {
		t.Fatalf("LoadByID(2) returned error: %v", err)
	}
	if len(target.Relations.Blocks) != 1 || target.Relations.Blocks[0] != 1 {
		t.Errorf("target.Relations.Blocks: got %v, want [1]", target.Relations.Blocks)
	}
}

func TestAddRelationRelatedTo(t *testing.T) {
	tmpDir := t.TempDir()

	issue3 := &Issue{
		ID:       3,
		Title:    "Related source",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}
	issue4 := &Issue{
		ID:       4,
		Title:    "Related target",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}

	createTestIssueFile(t, tmpDir, issue3)
	createTestIssueFile(t, tmpDir, issue4)

	cfg := &config.Config{AutoStage: false}

	err := AddRelation(tmpDir, 3, "related-to", 4, cfg)
	if err != nil {
		t.Fatalf("AddRelation returned error: %v", err)
	}

	source, err := LoadByID(tmpDir, 3)
	if err != nil {
		t.Fatalf("LoadByID(3) returned error: %v", err)
	}
	if len(source.Relations.RelatedTo) != 1 || source.Relations.RelatedTo[0] != 4 {
		t.Errorf("source.Relations.RelatedTo: got %v, want [4]", source.Relations.RelatedTo)
	}

	target, err := LoadByID(tmpDir, 4)
	if err != nil {
		t.Fatalf("LoadByID(4) returned error: %v", err)
	}
	if len(target.Relations.RelatedTo) != 1 || target.Relations.RelatedTo[0] != 3 {
		t.Errorf("target.Relations.RelatedTo: got %v, want [3]", target.Relations.RelatedTo)
	}
}

func TestAddRelationDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	issue5 := &Issue{
		ID:       5,
		Title:    "Dup source",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}
	issue6 := &Issue{
		ID:       6,
		Title:    "Dup target",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}

	createTestIssueFile(t, tmpDir, issue5)
	createTestIssueFile(t, tmpDir, issue6)

	cfg := &config.Config{AutoStage: false}

	err := AddRelation(tmpDir, 5, "duplicates", 6, cfg)
	if err != nil {
		t.Fatalf("AddRelation returned error: %v", err)
	}

	source, err := LoadByID(tmpDir, 5)
	if err != nil {
		t.Fatalf("LoadByID(5) returned error: %v", err)
	}
	if len(source.Relations.Duplicates) != 1 || source.Relations.Duplicates[0] != 6 {
		t.Errorf("source.Relations.Duplicates: got %v, want [6]", source.Relations.Duplicates)
	}

	target, err := LoadByID(tmpDir, 6)
	if err != nil {
		t.Fatalf("LoadByID(6) returned error: %v", err)
	}
	if len(target.Relations.Duplicates) != 1 || target.Relations.Duplicates[0] != 5 {
		t.Errorf("target.Relations.Duplicates: got %v, want [5]", target.Relations.Duplicates)
	}
}

func TestRemoveRelation(t *testing.T) {
	tmpDir := t.TempDir()

	issue7 := &Issue{
		ID:       7,
		Title:    "Remove source",
		Status:   "open",
		Priority: "high",
		Labels:   []string{},
		Relations: Relations{
			Blocks: []int{12},
		},
		Created: "2026-03-04",
		Updated: "2026-03-04",
	}
	issue12 := &Issue{
		ID:       12,
		Title:    "Remove target",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Relations: Relations{
			DependsOn: []int{7},
		},
		Created: "2026-03-04",
		Updated: "2026-03-04",
	}

	createTestIssueFile(t, tmpDir, issue7)
	createTestIssueFile(t, tmpDir, issue12)

	cfg := &config.Config{AutoStage: false}

	err := RemoveRelation(tmpDir, 7, "blocks", 12, cfg)
	if err != nil {
		t.Fatalf("RemoveRelation returned error: %v", err)
	}

	source, err := LoadByID(tmpDir, 7)
	if err != nil {
		t.Fatalf("LoadByID(7) returned error: %v", err)
	}
	if len(source.Relations.Blocks) != 0 {
		t.Errorf("source.Relations.Blocks should be empty, got %v", source.Relations.Blocks)
	}

	target, err := LoadByID(tmpDir, 12)
	if err != nil {
		t.Fatalf("LoadByID(12) returned error: %v", err)
	}
	if len(target.Relations.DependsOn) != 0 {
		t.Errorf("target.Relations.DependsOn should be empty, got %v", target.Relations.DependsOn)
	}
}

func TestDeduplication(t *testing.T) {
	tmpDir := t.TempDir()

	issue1 := &Issue{
		ID:       1,
		Title:    "Dedup source",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}
	issue2 := &Issue{
		ID:       2,
		Title:    "Dedup target",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}

	createTestIssueFile(t, tmpDir, issue1)
	createTestIssueFile(t, tmpDir, issue2)

	cfg := &config.Config{AutoStage: false}

	// Add the same relation twice
	err := AddRelation(tmpDir, 1, "blocks", 2, cfg)
	if err != nil {
		t.Fatalf("first AddRelation returned error: %v", err)
	}
	err = AddRelation(tmpDir, 1, "blocks", 2, cfg)
	if err != nil {
		t.Fatalf("second AddRelation returned error: %v", err)
	}

	source, err := LoadByID(tmpDir, 1)
	if err != nil {
		t.Fatalf("LoadByID(1) returned error: %v", err)
	}
	if len(source.Relations.Blocks) != 1 {
		t.Errorf("source.Relations.Blocks should have 1 entry after adding twice, got %d: %v",
			len(source.Relations.Blocks), source.Relations.Blocks)
	}

	target, err := LoadByID(tmpDir, 2)
	if err != nil {
		t.Fatalf("LoadByID(2) returned error: %v", err)
	}
	if len(target.Relations.DependsOn) != 1 {
		t.Errorf("target.Relations.DependsOn should have 1 entry after adding twice, got %d: %v",
			len(target.Relations.DependsOn), target.Relations.DependsOn)
	}
}

func TestAddToSliceDeduplication(t *testing.T) {
	issue := &Issue{
		Relations: Relations{
			Blocks: []int{5},
		},
	}

	AddToSlice(issue, "blocks", 5)
	if len(issue.Relations.Blocks) != 1 {
		t.Errorf("Blocks should still have 1 entry, got %d", len(issue.Relations.Blocks))
	}

	AddToSlice(issue, "blocks", 10)
	if len(issue.Relations.Blocks) != 2 {
		t.Errorf("Blocks should have 2 entries, got %d", len(issue.Relations.Blocks))
	}
}

func TestRemoveFromSlice(t *testing.T) {
	issue := &Issue{
		Relations: Relations{
			Blocks: []int{5, 10, 15},
		},
	}

	RemoveFromSlice(issue, "blocks", 10)
	if len(issue.Relations.Blocks) != 2 {
		t.Fatalf("Blocks should have 2 entries after removal, got %d", len(issue.Relations.Blocks))
	}
	if issue.Relations.Blocks[0] != 5 || issue.Relations.Blocks[1] != 15 {
		t.Errorf("Blocks: got %v, want [5, 15]", issue.Relations.Blocks)
	}
}

func TestRemoveFromSliceNotPresent(t *testing.T) {
	issue := &Issue{
		Relations: Relations{
			Blocks: []int{5, 10},
		},
	}

	RemoveFromSlice(issue, "blocks", 99)
	if len(issue.Relations.Blocks) != 2 {
		t.Errorf("Blocks should still have 2 entries, got %d", len(issue.Relations.Blocks))
	}
}

func TestDiffRelationsAdditions(t *testing.T) {
	oldRel := Relations{}
	newRel := Relations{
		Blocks:    []int{10},
		DependsOn: []int{3, 5},
	}

	added, removed := DiffRelations(oldRel, newRel)

	if len(removed) != 0 {
		t.Errorf("expected no removals, got %v", removed)
	}
	if len(added) != 3 {
		t.Fatalf("expected 3 additions, got %d: %v", len(added), added)
	}

	// Verify the additions contain the expected entries
	addedMap := make(map[string][]int)
	for _, e := range added {
		addedMap[e.Relation] = append(addedMap[e.Relation], e.ID)
	}
	if ids, ok := addedMap["blocks"]; !ok || len(ids) != 1 || ids[0] != 10 {
		t.Errorf("expected blocks addition of [10], got %v", addedMap["blocks"])
	}
	if ids, ok := addedMap["depends-on"]; !ok || len(ids) != 2 {
		t.Errorf("expected depends-on additions of [3, 5], got %v", addedMap["depends-on"])
	}
}

func TestDiffRelationsRemovals(t *testing.T) {
	oldRel := Relations{
		Blocks:    []int{10, 11},
		RelatedTo: []int{20},
	}
	newRel := Relations{
		Blocks: []int{10},
	}

	added, removed := DiffRelations(oldRel, newRel)

	if len(added) != 0 {
		t.Errorf("expected no additions, got %v", added)
	}
	if len(removed) != 2 {
		t.Fatalf("expected 2 removals, got %d: %v", len(removed), removed)
	}

	removedMap := make(map[string][]int)
	for _, e := range removed {
		removedMap[e.Relation] = append(removedMap[e.Relation], e.ID)
	}
	if ids, ok := removedMap["blocks"]; !ok || len(ids) != 1 || ids[0] != 11 {
		t.Errorf("expected blocks removal of [11], got %v", removedMap["blocks"])
	}
	if ids, ok := removedMap["related-to"]; !ok || len(ids) != 1 || ids[0] != 20 {
		t.Errorf("expected related-to removal of [20], got %v", removedMap["related-to"])
	}
}

func TestDiffRelationsMixed(t *testing.T) {
	oldRel := Relations{
		Blocks:     []int{10, 11},
		DependsOn:  []int{3},
		Duplicates: []int{99},
	}
	newRel := Relations{
		Blocks:    []int{11, 12},
		DependsOn: []int{3, 5},
	}

	added, removed := DiffRelations(oldRel, newRel)

	addedMap := make(map[string][]int)
	for _, e := range added {
		addedMap[e.Relation] = append(addedMap[e.Relation], e.ID)
	}
	removedMap := make(map[string][]int)
	for _, e := range removed {
		removedMap[e.Relation] = append(removedMap[e.Relation], e.ID)
	}

	// blocks: added 12, removed 10
	if ids := addedMap["blocks"]; len(ids) != 1 || ids[0] != 12 {
		t.Errorf("expected blocks addition of [12], got %v", ids)
	}
	if ids := removedMap["blocks"]; len(ids) != 1 || ids[0] != 10 {
		t.Errorf("expected blocks removal of [10], got %v", ids)
	}

	// depends-on: added 5, no removals
	if ids := addedMap["depends-on"]; len(ids) != 1 || ids[0] != 5 {
		t.Errorf("expected depends-on addition of [5], got %v", ids)
	}
	if ids := removedMap["depends-on"]; len(ids) != 0 {
		t.Errorf("expected no depends-on removals, got %v", ids)
	}

	// duplicates: removed 99
	if ids := removedMap["duplicates"]; len(ids) != 1 || ids[0] != 99 {
		t.Errorf("expected duplicates removal of [99], got %v", ids)
	}
}

func TestDiffRelationsNoChanges(t *testing.T) {
	rel := Relations{
		Blocks:    []int{10},
		DependsOn: []int{3},
	}

	added, removed := DiffRelations(rel, rel)

	if len(added) != 0 {
		t.Errorf("expected no additions for identical relations, got %v", added)
	}
	if len(removed) != 0 {
		t.Errorf("expected no removals for identical relations, got %v", removed)
	}
}

func TestAddRelationMissingSource(t *testing.T) {
	tmpDir := t.TempDir()

	issue2 := &Issue{
		ID:       2,
		Title:    "Existing target",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}
	createTestIssueFile(t, tmpDir, issue2)

	cfg := &config.Config{AutoStage: false}

	err := AddRelation(tmpDir, 999, "blocks", 2, cfg)
	if err == nil {
		t.Fatal("expected error when source ID not found, got nil")
	}
}

func TestAddRelationMissingTarget(t *testing.T) {
	tmpDir := t.TempDir()

	issue1 := &Issue{
		ID:       1,
		Title:    "Existing source",
		Status:   "open",
		Priority: "medium",
		Labels:   []string{},
		Created:  "2026-03-04",
		Updated:  "2026-03-04",
	}
	createTestIssueFile(t, tmpDir, issue1)

	cfg := &config.Config{AutoStage: false}

	err := AddRelation(tmpDir, 1, "blocks", 999, cfg)
	if err == nil {
		t.Fatal("expected error when target ID not found, got nil")
	}
}
