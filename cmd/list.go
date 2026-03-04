package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/steviee/git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	listCmd.Flags().String("status", "open", "Filter by status (open|in-progress|closed|wontfix|all)")
	listCmd.Flags().String("priority", "", "Filter by priority (low|medium|high|critical)")
	listCmd.Flags().StringSlice("label", nil, "Filter by label (OR logic, repeatable)")
	listCmd.Flags().Int("blocks", 0, "Filter issues that block this ID")
	listCmd.Flags().Int("depends-on", 0, "Filter issues that depend on this ID")
	listCmd.Flags().String("sort", "priority", "Sort by (priority|id|updated|created)")
	listCmd.Flags().String("format", "table", "Output format (table|json|ids)")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		labels, _ := cmd.Flags().GetStringSlice("label")
		blocksID, _ := cmd.Flags().GetInt("blocks")
		dependsOnID, _ := cmd.Flags().GetInt("depends-on")
		sortBy, _ := cmd.Flags().GetString("sort")
		format, _ := cmd.Flags().GetString("format")

		issues, err := issue.LoadAll(issuesDir)
		if err != nil {
			return err
		}

		// Build ID->Issue lookup for relation checks
		issueMap := make(map[int]*issue.Issue)
		for _, iss := range issues {
			issueMap[iss.ID] = iss
		}

		// Filter
		var filtered []*issue.Issue
		for _, iss := range issues {
			if status != "all" && iss.Status != status {
				continue
			}
			if priority != "" && iss.Priority != priority {
				continue
			}
			if len(labels) > 0 && !hasAnyLabel(iss.Labels, labels) {
				continue
			}
			if blocksID != 0 && !containsInt(iss.Relations.Blocks, blocksID) {
				continue
			}
			if dependsOnID != 0 && !containsInt(iss.Relations.DependsOn, dependsOnID) {
				continue
			}
			filtered = append(filtered, iss)
		}

		// Sort
		sortIssues(filtered, sortBy)

		// Output
		switch format {
		case "json":
			data, err := json.MarshalIndent(filtered, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		case "ids":
			ids := make([]string, len(filtered))
			for i, iss := range filtered {
				ids[i] = fmt.Sprintf("%d", iss.ID)
			}
			fmt.Println(strings.Join(ids, " "))
		default:
			printTable(filtered, issueMap)
		}

		return nil
	},
}

func hasAnyLabel(issueLabels, filterLabels []string) bool {
	for _, fl := range filterLabels {
		for _, il := range issueLabels {
			if il == fl {
				return true
			}
		}
	}
	return false
}

func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func sortIssues(issues []*issue.Issue, sortBy string) {
	sort.SliceStable(issues, func(i, j int) bool {
		switch sortBy {
		case "id":
			return issues[i].ID < issues[j].ID
		case "updated":
			return issues[i].Updated > issues[j].Updated
		case "created":
			return issues[i].Created > issues[j].Created
		default: // priority
			ri := issue.PriorityRank(issues[i].Priority)
			rj := issue.PriorityRank(issues[j].Priority)
			if ri != rj {
				return ri > rj
			}
			return issues[i].ID < issues[j].ID
		}
	})
}

func printTable(issues []*issue.Issue, issueMap map[int]*issue.Issue) {
	if len(issues) == 0 {
		fmt.Println("No issues found.")
		return
	}

	hasSymbols := false

	fmt.Fprintf(os.Stdout, "%-6s%-10s%-13s%-46s%-14s%s\n",
		"ID", "PRI", "STATUS", "TITLE", "LABELS", "")

	for _, iss := range issues {
		title := iss.Title
		if len(title) > 44 {
			title = title[:41] + "..."
		}

		labelsStr := ""
		if len(iss.Labels) > 0 {
			labelsStr = "[" + strings.Join(iss.Labels, ", ") + "]"
		}

		// Check relation indicators
		relStr := ""

		// Blocks open issues?
		var blocksOpen []int
		for _, bid := range iss.Relations.Blocks {
			if blocked, ok := issueMap[bid]; ok {
				if blocked.Status == "open" || blocked.Status == "in-progress" {
					blocksOpen = append(blocksOpen, bid)
				}
			}
		}
		if len(blocksOpen) > 0 {
			ids := make([]string, len(blocksOpen))
			for i, id := range blocksOpen {
				ids[i] = fmt.Sprintf("%d", id)
			}
			relStr = "⬛ " + strings.Join(ids, ",")
			hasSymbols = true
		}

		// Blocked by open issue?
		var blockedBy []int
		for _, did := range iss.Relations.DependsOn {
			if dep, ok := issueMap[did]; ok {
				if dep.Status == "open" || dep.Status == "in-progress" {
					blockedBy = append(blockedBy, did)
				}
			}
		}
		if len(blockedBy) > 0 {
			ids := make([]string, len(blockedBy))
			for i, id := range blockedBy {
				ids[i] = fmt.Sprintf("%d", id)
			}
			if relStr != "" {
				relStr += "  "
			}
			relStr += "⬜ " + strings.Join(ids, ",")
			hasSymbols = true
		}

		fmt.Fprintf(os.Stdout, "%04d  %-10s%-13s%-46s%-14s%s\n",
			iss.ID, iss.Priority, iss.Status, title, labelsStr, relStr)
	}

	if hasSymbols {
		fmt.Println()
		fmt.Println("⬛ = blocks open issues   ⬜ = blocked by open issue")
	}
}
