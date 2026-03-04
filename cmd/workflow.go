package cmd

import (
	"fmt"

	"github.com/steviee/git-issues/internal/config"
	"github.com/steviee/git-issues/internal/git"
	"github.com/steviee/git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(claimCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(blockedCmd)
}

var claimCmd = &cobra.Command{
	Use:   "claim <id>",
	Short: "Claim an issue (set status to in-progress)",
	Long:  "Shortcut for 'issues set <id> status in-progress'. Designed for AI agent workflows.",
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

		iss.Status = "in-progress"
		iss.Closed = ""

		if err := issue.Save(issuesDir, iss); err != nil {
			return err
		}

		if cfg.AutoStage {
			git.Stage(iss.FilePath)
		}

		fmt.Printf("Claimed: #%d %s\n", id, iss.Title)
		return nil
	},
}

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark an issue as done (close it)",
	Long:  "Shortcut for 'issues close <id>'. Designed for AI agent workflows.",
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

		iss.Status = "closed"
		iss.Closed = issue.Today()

		if err := issue.Save(issuesDir, iss); err != nil {
			return err
		}

		if cfg.AutoStage {
			git.Stage(iss.FilePath)
		}

		fmt.Printf("Done: #%d %s\n", id, iss.Title)
		return nil
	},
}

var blockedCmd = &cobra.Command{
	Use:   "blocked [id]",
	Short: "List blocked issues or show blockers for a specific issue",
	Long:  "Without arguments: lists all open issues that are blocked. With an ID: shows what blocks that specific issue.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		issues, err := issue.LoadAll(issuesDir)
		if err != nil {
			return err
		}

		issueMap := make(map[int]*issue.Issue)
		for _, iss := range issues {
			issueMap[iss.ID] = iss
		}

		if len(args) == 1 {
			// Show blockers for specific issue
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			iss, ok := issueMap[id]
			if !ok {
				return fmt.Errorf("Error: issue #%d not found", id)
			}

			if len(iss.Relations.DependsOn) == 0 {
				fmt.Printf("#%d has no blockers.\n", id)
				return nil
			}

			hasOpenBlockers := false
			for _, depID := range iss.Relations.DependsOn {
				dep, ok := issueMap[depID]
				if !ok {
					fmt.Printf("  #%d (not found)\n", depID)
					continue
				}
				status := ""
				if dep.Status == "open" || dep.Status == "in-progress" {
					status = " ← OPEN"
					hasOpenBlockers = true
				}
				fmt.Printf("  #%d %s [%s]%s\n", dep.ID, dep.Title, dep.Status, status)
			}
			if !hasOpenBlockers {
				fmt.Printf("#%d is not blocked (all blockers are resolved).\n", id)
			}
			return nil
		}

		// List all blocked open issues
		found := false
		for _, iss := range issues {
			if iss.Status != "open" && iss.Status != "in-progress" {
				continue
			}
			var openBlockers []int
			for _, depID := range iss.Relations.DependsOn {
				dep, ok := issueMap[depID]
				if ok && (dep.Status == "open" || dep.Status == "in-progress") {
					openBlockers = append(openBlockers, depID)
				}
			}
			if len(openBlockers) > 0 {
				blockerStr := ""
				for i, bid := range openBlockers {
					if i > 0 {
						blockerStr += ", "
					}
					blockerStr += fmt.Sprintf("#%d", bid)
				}
				fmt.Printf("#%d %s [%s] ← blocked by %s\n", iss.ID, iss.Title, iss.Priority, blockerStr)
				found = true
			}
		}
		if !found {
			fmt.Println("No blocked issues.")
		}
		return nil
	},
}
