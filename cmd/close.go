package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"git-issues/internal/config"
	"git-issues/internal/git"
	"git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	closeCmd.Flags().Bool("wontfix", false, "Close as wontfix")
	closeCmd.Flags().String("reason", "", "Reason for closing")
	rootCmd.AddCommand(closeCmd)
}

var closeCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close an issue",
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

		// Check for open blockers
		if len(iss.Relations.DependsOn) > 0 {
			var blockers []string
			for _, depID := range iss.Relations.DependsOn {
				dep, err := issue.LoadByID(issuesDir, depID)
				if err != nil {
					continue
				}
				if dep.Status == "open" || dep.Status == "in-progress" {
					blockers = append(blockers, fmt.Sprintf("#%d (%s)", dep.ID, dep.Title))
				}
			}
			if len(blockers) > 0 {
				for _, b := range blockers {
					fmt.Fprintf(os.Stderr, "Warning: #%d has open blocker: %s\n", id, b)
				}
				fmt.Fprint(os.Stderr, "Close anyway? [y/N] ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
			}
		}

		wontfix, _ := cmd.Flags().GetBool("wontfix")
		reason, _ := cmd.Flags().GetString("reason")

		if wontfix {
			iss.Status = "wontfix"
		} else {
			iss.Status = "closed"
		}
		iss.Closed = issue.Today()

		if reason != "" {
			if iss.Body != "" && !strings.HasSuffix(iss.Body, "\n") {
				iss.Body += "\n"
			}
			iss.Body += "\n## Closed\n" + reason + "\n"
		}

		if err := issue.Save(issuesDir, iss); err != nil {
			return err
		}

		if cfg.AutoStage {
			git.Stage(iss.FilePath)
		}

		fmt.Printf("Closed: #%d\n", id)
		return nil
	},
}
