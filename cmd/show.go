package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	showCmd.Flags().String("format", "text", "Output format (text|json)")
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show issue details",
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

		iss, err := issue.LoadByID(issuesDir, id)
		if err != nil {
			return fmt.Errorf("Error: issue #%d not found", id)
		}

		format, _ := cmd.Flags().GetString("format")
		if format == "json" {
			data, err := json.MarshalIndent(iss, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		line := strings.Repeat("━", 54)

		fmt.Println(line)
		fmt.Printf("Issue #%d  ·  %s  ·  %s\n", iss.ID, iss.Priority, iss.Status)
		fmt.Println(iss.Title)
		fmt.Println(line)

		if len(iss.Labels) > 0 {
			fmt.Printf("Labels:   %s\n", strings.Join(iss.Labels, ", "))
		}
		fmt.Printf("Created:  %s\n", iss.Created)
		fmt.Printf("Updated:  %s\n", iss.Updated)
		if iss.Closed != "" {
			fmt.Printf("Closed:   %s\n", iss.Closed)
		}

		// Relations
		hasRelations := len(iss.Relations.Blocks) > 0 ||
			len(iss.Relations.DependsOn) > 0 ||
			len(iss.Relations.RelatedTo) > 0 ||
			len(iss.Relations.Duplicates) > 0

		if hasRelations {
			fmt.Println()
			fmt.Println("Relations")
			printRelationSection(issuesDir, "Blocks", iss.Relations.Blocks, false)
			printRelationSection(issuesDir, "Depends on", iss.Relations.DependsOn, true)
			printRelationSection(issuesDir, "Related to", iss.Relations.RelatedTo, false)
			printRelationSection(issuesDir, "Duplicates", iss.Relations.Duplicates, false)
		}

		fmt.Println()
		fmt.Println(line)

		if iss.Body != "" {
			fmt.Println()
			fmt.Println(iss.Body)
		}

		return nil
	},
}

func printRelationSection(issuesDir string, label string, ids []int, showBlocker bool) {
	if len(ids) == 0 {
		return
	}
	for i, id := range ids {
		prefix := fmt.Sprintf("  %-12s", label+":")
		if i > 0 {
			prefix = fmt.Sprintf("  %12s", "")
		}
		related, err := issue.LoadByID(issuesDir, id)
		if err != nil {
			fmt.Printf("%s#%-3d (not found)\n", prefix, id)
			continue
		}
		blocker := ""
		if showBlocker && (related.Status == "open" || related.Status == "in-progress") {
			blocker = "  ← BLOCKER"
		}
		fmt.Printf("%s#%-3d %-40s [%s]%s\n", prefix, id, related.Title, related.Status, blocker)
	}
}
