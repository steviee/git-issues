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
	rootCmd.AddCommand(setCmd)
}

var setCmd = &cobra.Command{
	Use:   "set <id> <field> <value>",
	Short: "Set a field on an issue",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		field := args[1]
		value := args[2]

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
		case "title":
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("Error: title must not be empty")
			}
			iss.Title = value
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
			return fmt.Errorf("Error: unknown field %q, must be one of: priority, status, title, label", field)
		}

		if err := issue.Save(issuesDir, iss); err != nil {
			return err
		}

		if cfg.AutoStage {
			git.Stage(iss.FilePath)
		}

		fmt.Printf("Updated #%d: %s = %s\n", id, field, value)
		return nil
	},
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func containsStr(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func removeStr(slice []string, val string) []string {
	var result []string
	for _, s := range slice {
		if s != val {
			result = append(result, s)
		}
	}
	if result == nil {
		return []string{}
	}
	return result
}
