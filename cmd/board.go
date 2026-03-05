package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/steviee/git-issues/internal/config"
	"github.com/steviee/git-issues/internal/issue"
	"github.com/steviee/git-issues/internal/tui"
)

func init() {
	boardCmd.Flags().String("priority", "", "Filter by priority (low|medium|high|critical)")
	boardCmd.Flags().StringSlice("label", nil, "Filter by label (OR logic, repeatable)")
	rootCmd.AddCommand(boardCmd)
}

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Interactive Kanban board",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		cfg, err := config.Load(issuesDir)
		if err != nil {
			return err
		}

		priority, _ := cmd.Flags().GetString("priority")
		labels, _ := cmd.Flags().GetStringSlice("label")

		// Build filter functions (used for initial load and live reload)
		var filters []tui.FilterFunc
		if priority != "" {
			p := priority // capture
			filters = append(filters, func(iss *issue.Issue) bool {
				return iss.Priority == p
			})
		}
		if len(labels) > 0 {
			ls := labels // capture
			filters = append(filters, func(iss *issue.Issue) bool {
				return hasAnyLabel(iss.Labels, ls)
			})
		}

		issues, err := issue.LoadAll(issuesDir)
		if err != nil {
			return err
		}

		// Apply filters for initial load
		var filtered []*issue.Issue
		for _, iss := range issues {
			keep := true
			for _, f := range filters {
				if !f(iss) {
					keep = false
					break
				}
			}
			if keep {
				filtered = append(filtered, iss)
			}
		}

		model := tui.NewModel(filtered, issuesDir, cfg, filters...)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running board: %w", err)
		}
		return nil
	},
}
