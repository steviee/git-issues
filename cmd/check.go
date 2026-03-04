package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/steviee/git-issues/internal/issue"

	"github.com/spf13/cobra"
)

func init() {
	checkCmd.Flags().Bool("fix", false, "Auto-fix fixable issues (broken relations)")
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check issue database for inconsistencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.IssuesDir()
		if err != nil {
			return err
		}

		autoFix, _ := cmd.Flags().GetBool("fix")

		issues, err := issue.LoadAll(issuesDir)
		if err != nil {
			return err
		}

		issueMap := make(map[int]*issue.Issue)
		var errors []string
		var warnings []string
		var fixed []string

		// 1. Check for duplicate IDs
		idCount := make(map[int][]string)
		for _, iss := range issues {
			idCount[iss.ID] = append(idCount[iss.ID], iss.FilePath)
		}
		for id, files := range idCount {
			if len(files) > 1 {
				errors = append(errors, fmt.Sprintf("Duplicate ID #%d in files: %s", id, strings.Join(files, ", ")))
			}
		}

		// Build map (use first occurrence for each ID)
		for _, iss := range issues {
			if _, exists := issueMap[iss.ID]; !exists {
				issueMap[iss.ID] = iss
			}
		}

		for _, iss := range issues {
			// 2. Validation check
			if err := issue.Validate(iss); err != nil {
				errors = append(errors, fmt.Sprintf("#%d: validation error: %s", iss.ID, err))
			}

			// 3. Duplicate labels within an issue
			labelSeen := make(map[string]bool)
			for _, l := range iss.Labels {
				if labelSeen[l] {
					warnings = append(warnings, fmt.Sprintf("#%d: duplicate label %q", iss.ID, l))
				}
				labelSeen[l] = true
			}

			// 4. Orphan references (relations pointing to non-existent issues)
			checkOrphans := func(relation string, ids []int) {
				for _, refID := range ids {
					if _, ok := issueMap[refID]; !ok {
						errors = append(errors, fmt.Sprintf("#%d: %s references non-existent #%d", iss.ID, relation, refID))
						if autoFix {
							issue.RemoveFromSlice(iss, relation, refID)
							fixed = append(fixed, fmt.Sprintf("#%d: removed orphan %s reference to #%d", iss.ID, relation, refID))
						}
					}
				}
			}
			checkOrphans("blocks", iss.Relations.Blocks)
			checkOrphans("depends-on", iss.Relations.DependsOn)
			checkOrphans("related-to", iss.Relations.RelatedTo)
			checkOrphans("duplicates", iss.Relations.Duplicates)

			// 5. Bidirectional relation symmetry
			checkSymmetry := func(relation string, ids []int) {
				inv := issue.Inverse(relation)
				for _, refID := range ids {
					target, ok := issueMap[refID]
					if !ok {
						continue // already reported as orphan
					}
					invSlice := getRelSlice(target, inv)
					if !containsInt(invSlice, iss.ID) {
						errors = append(errors, fmt.Sprintf("#%d %s #%d but #%d does not have inverse %s #%d",
							iss.ID, relation, refID, refID, inv, iss.ID))
						if autoFix {
							issue.AddToSlice(target, inv, iss.ID)
							fixed = append(fixed, fmt.Sprintf("#%d: added missing %s #%d", refID, inv, iss.ID))
						}
					}
				}
			}
			checkSymmetry("blocks", iss.Relations.Blocks)
			checkSymmetry("depends-on", iss.Relations.DependsOn)
			checkSymmetry("related-to", iss.Relations.RelatedTo)
			checkSymmetry("duplicates", iss.Relations.Duplicates)

			// 6. Status inconsistencies
			if (iss.Status == "closed" || iss.Status == "wontfix") && iss.Closed == "" {
				warnings = append(warnings, fmt.Sprintf("#%d: status is %s but closed date is empty", iss.ID, iss.Status))
			}
			if iss.Status == "open" && iss.Closed != "" {
				warnings = append(warnings, fmt.Sprintf("#%d: status is open but has closed date %s", iss.ID, iss.Closed))
			}
		}

		// Save fixed issues
		if autoFix && len(fixed) > 0 {
			for _, iss := range issues {
				issue.Save(issuesDir, iss)
			}
		}

		// Output
		if len(errors) == 0 && len(warnings) == 0 {
			fmt.Println("No issues found. Database is consistent.")
			return nil
		}

		if len(errors) > 0 {
			fmt.Fprintf(os.Stderr, "Errors (%d):\n", len(errors))
			for _, e := range errors {
				fmt.Fprintf(os.Stderr, "  ✗ %s\n", e)
			}
		}

		if len(warnings) > 0 {
			fmt.Fprintf(os.Stderr, "Warnings (%d):\n", len(warnings))
			for _, w := range warnings {
				fmt.Fprintf(os.Stderr, "  ⚠ %s\n", w)
			}
		}

		if len(fixed) > 0 {
			fmt.Fprintf(os.Stdout, "Fixed (%d):\n", len(fixed))
			for _, f := range fixed {
				fmt.Fprintf(os.Stdout, "  ✓ %s\n", f)
			}
		}

		if len(errors) > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func getRelSlice(iss *issue.Issue, relation string) []int {
	switch relation {
	case "blocks":
		return iss.Relations.Blocks
	case "depends-on":
		return iss.Relations.DependsOn
	case "related-to":
		return iss.Relations.RelatedTo
	case "duplicates":
		return iss.Relations.Duplicates
	default:
		return nil
	}
}
