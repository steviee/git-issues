package cmd

import (
	"fmt"
	"sort"

	"github.com/steviee/git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	labelsCmd.Flags().String("sort", "count", "Sort by (count|alpha)")
	rootCmd.AddCommand(labelsCmd)
}

var labelsCmd = &cobra.Command{
	Use:   "labels",
	Short: "List all labels with frequency",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		sortBy, _ := cmd.Flags().GetString("sort")

		issues, err := issue.LoadAll(issuesDir)
		if err != nil {
			return err
		}

		counts := make(map[string]int)
		for _, iss := range issues {
			for _, label := range iss.Labels {
				counts[label]++
			}
		}

		type labelCount struct {
			label string
			count int
		}

		var sorted []labelCount
		for label, count := range counts {
			sorted = append(sorted, labelCount{label, count})
		}

		switch sortBy {
		case "alpha":
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].label < sorted[j].label
			})
		default: // count
			sort.Slice(sorted, func(i, j int) bool {
				if sorted[i].count != sorted[j].count {
					return sorted[i].count > sorted[j].count
				}
				return sorted[i].label < sorted[j].label
			})
		}

		maxLen := 0
		for _, lc := range sorted {
			if len(lc.label) > maxLen {
				maxLen = len(lc.label)
			}
		}

		for _, lc := range sorted {
			fmt.Printf("%-*s  %d\n", maxLen, lc.label, lc.count)
		}

		return nil
	},
}
