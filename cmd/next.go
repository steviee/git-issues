package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	nextCmd.Flags().String("priority", "", "Minimum priority (low|medium|high|critical)")
	nextCmd.Flags().StringSlice("label", nil, "Filter by label (OR logic)")
	nextCmd.Flags().String("format", "text", "Output format (text|json)")
	rootCmd.AddCommand(nextCmd)
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show next actionable issue (unblocked, highest priority)",
	Long:  "Finds the highest-priority open issue that is not blocked by other open issues. Designed for AI agent workflows.",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		minPriority, _ := cmd.Flags().GetString("priority")
		labels, _ := cmd.Flags().GetStringSlice("label")

		issues, err := issue.LoadAll(issuesDir)
		if err != nil {
			return err
		}

		// Build map for status lookups
		issueMap := make(map[int]*issue.Issue)
		for _, iss := range issues {
			issueMap[iss.ID] = iss
		}

		// Filter: open only, unblocked, matching filters
		var candidates []*issue.Issue
		for _, iss := range issues {
			if iss.Status != "open" && iss.Status != "in-progress" {
				continue
			}

			// Check minimum priority
			if minPriority != "" && issue.PriorityRank(iss.Priority) < issue.PriorityRank(minPriority) {
				continue
			}

			// Check label filter
			if len(labels) > 0 && !hasAnyLabel(iss.Labels, labels) {
				continue
			}

			// Check if blocked
			blocked := false
			for _, depID := range iss.Relations.DependsOn {
				dep, ok := issueMap[depID]
				if ok && (dep.Status == "open" || dep.Status == "in-progress") {
					blocked = true
					break
				}
			}
			if blocked {
				continue
			}

			candidates = append(candidates, iss)
		}

		if len(candidates) == 0 {
			fmt.Println("No actionable issues found.")
			return nil
		}

		// Sort: priority desc, then ID asc
		sort.SliceStable(candidates, func(i, j int) bool {
			ri := issue.PriorityRank(candidates[i].Priority)
			rj := issue.PriorityRank(candidates[j].Priority)
			if ri != rj {
				return ri > rj
			}
			return candidates[i].ID < candidates[j].ID
		})

		best := candidates[0]

		format, _ := cmd.Flags().GetString("format")
		if format == "json" {
			data, err := json.MarshalIndent(best, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		// Single-line output optimized for agent parsing
		fmt.Printf("%d %s %s [%s]\n", best.ID, best.Priority, best.Title, joinLabels(best.Labels))
		return nil
	},
}

func joinLabels(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	result := labels[0]
	for _, l := range labels[1:] {
		result += "," + l
	}
	return result
}
