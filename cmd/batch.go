package cmd

import (
	"fmt"
	"strings"

	"github.com/steviee/git-issues/internal/config"
	"github.com/steviee/git-issues/internal/git"
	"github.com/steviee/git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	batchCloseCmd.Flags().Bool("wontfix", false, "Close as wontfix")
	batchCloseCmd.Flags().StringSlice("label", nil, "Filter by label (OR logic)")
	batchCloseCmd.Flags().String("status", "open", "Filter by current status")

	batchSetCmd.Flags().StringSlice("label", nil, "Filter by label (OR logic)")
	batchSetCmd.Flags().String("status", "open", "Filter by current status")

	batchCmd.AddCommand(batchCloseCmd)
	batchCmd.AddCommand(batchSetCmd)
	rootCmd.AddCommand(batchCmd)
}

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Bulk operations on multiple issues",
}

var batchCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "Close multiple issues matching filters",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		cfg, err := config.Load(issuesDir)
		if err != nil {
			return err
		}

		labels, _ := cmd.Flags().GetStringSlice("label")
		status, _ := cmd.Flags().GetString("status")
		wontfix, _ := cmd.Flags().GetBool("wontfix")

		issues, err := issue.LoadAll(issuesDir)
		if err != nil {
			return err
		}

		count := 0
		for _, iss := range issues {
			if status != "all" && iss.Status != status {
				continue
			}
			if len(labels) > 0 && !hasAnyLabel(iss.Labels, labels) {
				continue
			}

			if wontfix {
				iss.Status = "wontfix"
			} else {
				iss.Status = "closed"
			}
			iss.Closed = issue.Today()

			if err := issue.Save(issuesDir, iss); err != nil {
				fmt.Printf("Error closing #%d: %s\n", iss.ID, err)
				continue
			}
			if cfg.AutoStage {
				git.Stage(iss.FilePath)
			}
			count++
			fmt.Printf("Closed: #%d %s\n", iss.ID, iss.Title)
		}

		fmt.Printf("\n%d issues closed.\n", count)
		return nil
	},
}

var batchSetCmd = &cobra.Command{
	Use:   "set <field> <value>",
	Short: "Set a field on multiple issues matching filters",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		field := args[0]
		value := args[1]

		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		cfg, err := config.Load(issuesDir)
		if err != nil {
			return err
		}

		labels, _ := cmd.Flags().GetStringSlice("label")
		status, _ := cmd.Flags().GetString("status")

		issues, err := issue.LoadAll(issuesDir)
		if err != nil {
			return err
		}

		count := 0
		for _, iss := range issues {
			if status != "all" && iss.Status != status {
				continue
			}
			if len(labels) > 0 && !hasAnyLabel(iss.Labels, labels) {
				continue
			}

			switch field {
			case "priority":
				if !contains(issue.ValidPriorities, value) {
					return fmt.Errorf("Error: invalid priority %q, must be one of: %s", value, strings.Join(issue.ValidPriorities, ", "))
				}
				iss.Priority = value
			case "status":
				if !contains(issue.ValidStatuses, value) {
					return fmt.Errorf("Error: invalid status %q, must be one of: %s", value, strings.Join(issue.ValidStatuses, ", "))
				}
				iss.Status = value
				if value == "closed" || value == "wontfix" {
					iss.Closed = issue.Today()
				} else {
					iss.Closed = ""
				}
			case "label":
				if strings.HasPrefix(value, "+") {
					label := value[1:]
					if !containsStr(iss.Labels, label) {
						iss.Labels = append(iss.Labels, label)
					}
				} else if strings.HasPrefix(value, "-") {
					label := value[1:]
					iss.Labels = removeStr(iss.Labels, label)
				} else {
					iss.Labels = strings.Split(value, ",")
					for i := range iss.Labels {
						iss.Labels[i] = strings.TrimSpace(iss.Labels[i])
					}
				}
			default:
				return fmt.Errorf("Error: unknown field %q for batch set, must be one of: priority, status, label", field)
			}

			if err := issue.Save(issuesDir, iss); err != nil {
				fmt.Printf("Error updating #%d: %s\n", iss.ID, err)
				continue
			}
			if cfg.AutoStage {
				git.Stage(iss.FilePath)
			}
			count++
			fmt.Printf("Updated #%d: %s = %s\n", iss.ID, field, value)
		}

		fmt.Printf("\n%d issues updated.\n", count)
		return nil
	},
}
