package cmd

import (
	"fmt"

	"git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	graphCmd.Flags().Bool("open-only", false, "Show only open/in-progress issues")
	graphCmd.Flags().Int("root", 0, "Show subgraph from this issue ID")
	rootCmd.AddCommand(graphCmd)
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Show dependency graph",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		openOnly, _ := cmd.Flags().GetBool("open-only")
		rootID, _ := cmd.Flags().GetInt("root")

		issues, err := issue.LoadAll(issuesDir)
		if err != nil {
			return err
		}

		// Build lookup
		issueMap := make(map[int]*issue.Issue)
		for _, iss := range issues {
			if openOnly && iss.Status != "open" && iss.Status != "in-progress" {
				continue
			}
			issueMap[iss.ID] = iss
		}

		if rootID != 0 {
			// Show subgraph from root
			visited := make(map[int]bool)
			printed := make(map[int]bool)
			printNode(issueMap, rootID, "", visited, printed)
			return nil
		}

		// Find root nodes (not blocked by anything in the visible set)
		blockedIDs := make(map[int]bool)
		for _, iss := range issueMap {
			for _, bid := range iss.Relations.Blocks {
				if _, ok := issueMap[bid]; ok {
					blockedIDs[bid] = true
				}
			}
		}

		// Print root nodes first, tracking all printed nodes globally
		globalPrinted := make(map[int]bool)
		for _, iss := range issues {
			if _, ok := issueMap[iss.ID]; !ok {
				continue
			}
			if blockedIDs[iss.ID] || globalPrinted[iss.ID] {
				continue
			}
			visited := make(map[int]bool)
			printNode(issueMap, iss.ID, "", visited, globalPrinted)
			fmt.Println()
		}

		// Print any remaining unprinted nodes (cycles or isolated)
		for _, iss := range issues {
			if _, ok := issueMap[iss.ID]; !ok {
				continue
			}
			if globalPrinted[iss.ID] {
				continue
			}
			visited := make(map[int]bool)
			printNode(issueMap, iss.ID, "", visited, globalPrinted)
			fmt.Println()
		}

		return nil
	},
}

func printNode(issueMap map[int]*issue.Issue, id int, indent string, visited map[int]bool, globalPrinted map[int]bool) {
	iss, ok := issueMap[id]
	if !ok {
		return
	}

	if visited[id] {
		fmt.Printf("%sWarning: cycle detected at #%d, truncating\n", indent, id)
		return
	}
	visited[id] = true
	globalPrinted[id] = true

	fmt.Printf("%s#%d  %s [%s, %s]\n", indent, iss.ID, iss.Title, iss.Status, iss.Priority)

	// Show blocks relations
	var blocksInSet []int
	for _, bid := range iss.Relations.Blocks {
		if _, ok := issueMap[bid]; ok {
			blocksInSet = append(blocksInSet, bid)
		}
	}

	if len(blocksInSet) > 0 {
		fmt.Printf("%s└── blocks:\n", indent)
		for _, bid := range blocksInSet {
			printNode(issueMap, bid, indent+"    ", visited, globalPrinted)
		}
	} else if indent == "" {
		fmt.Printf("%s    (no relations)\n", indent)
	}
}
