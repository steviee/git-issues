package issue

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IssuesDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, ".issues")
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .issues/ directory found (run 'issues init')")
		}
		dir = parent
	}
}

func LoadAll(issuesDir string) ([]*Issue, error) {
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return nil, err
	}

	var issues []*Issue
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, ".") {
			continue
		}
		issue, err := ParseFile(filepath.Join(issuesDir, name))
		if err != nil {
			continue
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

func LoadByID(issuesDir string, id int) (*Issue, error) {
	issues, err := LoadAll(issuesDir)
	if err != nil {
		return nil, err
	}
	for _, issue := range issues {
		if issue.ID == id {
			return issue, nil
		}
	}
	return nil, fmt.Errorf("issue #%d not found", id)
}

func Save(issuesDir string, issue *Issue) error {
	issue.Updated = Today()

	if err := Validate(issue); err != nil {
		return err
	}

	data := Marshal(issue)

	if issue.FilePath == "" {
		slug := GenerateSlug(issue.Title)
		issue.FilePath = filepath.Join(issuesDir, Filename(issue.ID, slug))
	}

	return os.WriteFile(issue.FilePath, data, 0644)
}

func Delete(issuesDir string, id int) error {
	issue, err := LoadByID(issuesDir, id)
	if err != nil {
		return err
	}
	return os.Remove(issue.FilePath)
}
