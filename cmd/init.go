package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"git-issues/internal/config"

	"github.com/spf13/cobra"
)

const agentMD = `# git-issues agent context

## Schema
Each issue is a Markdown file with YAML frontmatter in .issues/.

Fields:
- id: integer, unique, never reused
- title: string
- status: open | in-progress | closed | wontfix
- priority: low | medium | high | critical
- labels: string array
- relations.blocks: int array (this issue blocks these IDs)
- relations.depends-on: int array (this issue needs these IDs first)
- relations.related-to: int array (loose relation)
- relations.duplicates: int array

## Agent-Optimized Commands (use these first)
issues next                                  # next unblocked highest-priority issue (one-liner)
issues next --format json                    # same as above, JSON output
issues claim <id>                            # set status to in-progress
issues done <id>                             # close the issue
issues blocked                               # list all blocked issues
issues blocked <id>                          # show what blocks a specific issue
issues check                                 # verify database consistency
issues check --fix                           # auto-fix broken relations

## Standard Commands
issues list --format ids --status open        # space-separated open IDs
issues list --format json --status open       # full JSON output
issues list --format ids --priority high      # filter by priority
issues show <id>                              # full issue with resolved relations
issues show <id> --format json                # JSON output
issues new --title "..." --priority high      # create issue inline
issues set <id> priority critical             # change priority
issues set <id> label +bug                    # add label
issues close <id>                             # close with blocker warning
issues close <id> --wontfix                   # mark as won't fix
issues relate <id> blocks <id>                # declare blocking relation
issues graph --open-only                      # dependency tree

## Bulk Operations
issues batch close --label <label>            # close all matching issues
issues batch set priority high --label <label># set field on all matching issues

## Recommended Workflow
1. issues next → get the next actionable issue
2. issues claim <id> → mark as in-progress
3. Do the work
4. issues done <id> → close, auto-staged for next commit

## Alternative Workflow (detailed)
1. issues list --format ids --status open → get IDs
2. issues show <id> → check for open blockers in "Depends on"
3. If blockers exist, work on those first
4. issues claim <id> → mark as in-progress
5. Do the work
6. issues done <id> → close, auto-staged for next commit

## Import GitHub Issues
To migrate existing GitHub issues into git-issues, use the gh CLI:

` + "`" + `` + "`" + `` + "`" + `bash
# Initialize git-issues if not already done
issues init

# Import all open GitHub issues
gh issue list --state open --json number,title,body,labels --limit 500 | \
  jq -c '.[]' | while read -r issue; do
    title=$(echo "$issue" | jq -r '.title')
    body=$(echo "$issue" | jq -r '.body // ""')
    labels=$(echo "$issue" | jq -r '[.labels[].name] | join(",")')
    if [ -n "$labels" ]; then
      issues new --title "$title" --body "$body" --label "$labels"
    else
      issues new --title "$title" --body "$body"
    fi
  done
` + "`" + `` + "`" + `` + "`" + `

After import, all issues are local Markdown files tracked by git.
From here, use git-issues commands to manage them.
`

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize git-issues in current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir := filepath.Join(".", ".issues")

		if _, err := os.Stat(issuesDir); err == nil {
			return fmt.Errorf(".issues/ already exists")
		}

		if err := os.MkdirAll(issuesDir, 0755); err != nil {
			return err
		}

		if err := config.WriteDefault(issuesDir); err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(issuesDir, ".agent.md"), []byte(agentMD), 0644); err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(issuesDir, ".gitignore"), []byte(""), 0644); err != nil {
			return err
		}

		fmt.Println("Initialized git-issues in .issues/")
		return nil
	},
}
