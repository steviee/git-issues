package cmd

import (
	"fmt"

	"github.com/steviee/git-issues/internal/config"
	"github.com/steviee/git-issues/internal/git"
	"github.com/steviee/git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reopenCmd)
}

var reopenCmd = &cobra.Command{
	Use:   "reopen <id>",
	Short: "Reopen a closed issue",
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

		iss.Status = "open"
		iss.Closed = ""

		if err := issue.Save(issuesDir, iss); err != nil {
			return err
		}

		if cfg.AutoStage {
			git.Stage(iss.FilePath)
		}

		fmt.Printf("Reopened: #%d\n", id)
		return nil
	},
}
