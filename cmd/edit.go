package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/steviee/git-issues/internal/config"
	"github.com/steviee/git-issues/internal/git"
	"github.com/steviee/git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an issue in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}

		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		cfg, err := config.Load(issuesDir)
		if err != nil {
			return err
		}

		iss, err := issue.LoadByID(issuesDir, id)
		if err != nil {
			return fmt.Errorf("Error: issue #%d not found", id)
		}

		// Snapshot old relations for diff (will be used in Phase 5)
		oldRelations := iss.Relations

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		editorCmd := exec.Command(editor, iss.FilePath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("editor failed: %w", err)
		}

		// Re-parse the file
		edited, err := issue.ParseFile(iss.FilePath)
		if err != nil {
			return fmt.Errorf("Error parsing edited file: %w", err)
		}

		if err := issue.Validate(edited); err != nil {
			return fmt.Errorf("Error: %w", err)
		}

		// Save (updates `updated` field)
		if err := issue.Save(issuesDir, edited); err != nil {
			return err
		}

		if cfg.AutoStage {
			git.Stage(edited.FilePath)
		}

		// Sync relation changes bidirectionally
		added, removed := issue.DiffRelations(oldRelations, edited.Relations)
		for _, entry := range added {
			// Source already has the relation from user edit, just update target's inverse
			target, err := issue.LoadByID(issuesDir, entry.ID)
			if err != nil {
				continue
			}
			issue.AddToSlice(target, issue.Inverse(entry.Relation), id)
			issue.Save(issuesDir, target)
			if cfg.AutoStage {
				git.Stage(target.FilePath)
			}
		}
		for _, entry := range removed {
			target, err := issue.LoadByID(issuesDir, entry.ID)
			if err != nil {
				continue
			}
			issue.RemoveFromSlice(target, issue.Inverse(entry.Relation), id)
			issue.Save(issuesDir, target)
			if cfg.AutoStage {
				git.Stage(target.FilePath)
			}
		}

		fmt.Printf("Updated: #%d\n", id)
		return nil
	},
}
