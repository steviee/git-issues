package issue

import (
	"os"
	"path/filepath"
	"strings"
)

func NextID(issuesDir string) (int, error) {
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return 0, err
	}

	maxID := 0
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
		if issue.ID > maxID {
			maxID = issue.ID
		}
	}

	return maxID + 1, nil
}
