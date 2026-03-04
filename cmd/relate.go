package cmd

import (
	"fmt"
	"strings"

	"github.com/steviee/git-issues/internal/config"
	"github.com/steviee/git-issues/internal/issue"

	"github.com/spf13/cobra"
)

var validRelations = []string{"blocks", "depends-on", "related-to", "duplicates"}

func init() {
	rootCmd.AddCommand(relateCmd)
}

var relateCmd = &cobra.Command{
	Use:   "relate <id> <relation> <target-id>",
	Short: "Add a relation between two issues",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceID, err := parseID(args[0])
		if err != nil {
			return err
		}
		relation := args[1]
		targetID, err := parseID(args[2])
		if err != nil {
			return err
		}

		if !contains(validRelations, relation) {
			return fmt.Errorf("Error: invalid relation %q, must be one of: %s", relation, strings.Join(validRelations, ", "))
		}

		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		cfg, err := config.Load(issuesDir)
		if err != nil {
			return err
		}

		if err := issue.AddRelation(issuesDir, sourceID, relation, targetID, cfg); err != nil {
			return err
		}

		fmt.Printf("Related: #%d %s #%d\n", sourceID, relation, targetID)
		fmt.Printf("         #%d %s #%d (auto)\n", targetID, issue.Inverse(relation), sourceID)
		return nil
	},
}
