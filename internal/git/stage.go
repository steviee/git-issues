package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Stage(filePath string) error {
	cmd := exec.Command("git", "add", filePath)
	cmd.Dir = filepath.Dir(filePath)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not stage file (not a git repo?)\n")
		return nil
	}
	return nil
}
