# git-issues – Implementation Brief for Claude Code

## Project Goal

Build a CLI tool called `git-issues` (binary: `issues`) that provides a git-native issue tracker. Issues are stored as Markdown files inside the repository itself under `.issues/`, making them version-controlled alongside the code they describe.

---

## Tech Stack

- **Language:** Go 1.22+
- **Dependencies (minimal):**
  - `gopkg.in/yaml.v3` – Frontmatter parsing
  - `github.com/spf13/cobra` – CLI framework
  - No database, no network, no auth
- **Binary name:** `issues`
- **Install target:** `/usr/local/bin/issues`

---

## Repository Structure

```
git-issues/
  cmd/
    root.go
    new.go
    list.go
    show.go
    edit.go
    close.go
    reopen.go
    set.go
    relate.go
    unrelate.go
    labels.go
    graph.go
  internal/
    issue/
      issue.go       ← Issue struct + Frontmatter schema
      store.go       ← Read/write .issues/ directory
      relations.go   ← Bidirectional relation sync
      id.go          ← ID generation
      slug.go        ← Title → filename slug
    git/
      stage.go       ← git add wrapper
    config/
      config.go      ← .issues/.config.yml
  main.go
  go.mod
  go.sum
  README.md
  .issues/
    .agent.md        ← generated on `issues init`
    .config.yml      ← generated on `issues init`
```

---

## Data Model

### Issue Struct

```go
type Issue struct {
    // Frontmatter
    ID       int       `yaml:"id"`
    Title    string    `yaml:"title"`
    Status   string    `yaml:"status"`   // open | in-progress | closed | wontfix
    Priority string    `yaml:"priority"` // low | medium | high | critical
    Labels   []string  `yaml:"labels"`
    Relations Relations `yaml:"relations,omitempty"`
    Created  string    `yaml:"created"`  // ISO date YYYY-MM-DD
    Updated  string    `yaml:"updated"`
    Closed   string    `yaml:"closed,omitempty"`

    // Body (everything after frontmatter delimiter)
    Body string `yaml:"-"`

    // Derived (not stored)
    FilePath string `yaml:"-"`
}

type Relations struct {
    Blocks    []int `yaml:"blocks,omitempty"`
    DependsOn []int `yaml:"depends-on,omitempty"`
    RelatedTo []int `yaml:"related-to,omitempty"`
    Duplicates []int `yaml:"duplicates,omitempty"`
}
```

### File Format

```
---
id: 7
title: "Login schlägt bei leeren Passwörtern fehl"
status: open
priority: high
labels: [bug, auth]
relations:
  blocks: [12, 15]
  depends-on: [3]
created: 2026-03-04
updated: 2026-03-04
---

Freitext Markdown body hier.

## Notes
Weitere Notizen als freie Sections.
```

**Critical:** The frontmatter delimiter is exactly `---` on its own line. The file always starts with `---`, contains YAML, then `---`, then the body. If no body, a newline after the closing `---` is sufficient.

### Filename Convention

```
{id:04d}-{slug}.md
```

Slug generation:
1. Lowercase the title
2. Replace umlauts: ä→ae, ö→oe, ü→ue, ß→ss
3. Remove all characters except `[a-z0-9 -]`
4. Replace spaces with `-`
5. Truncate to 40 characters
6. Strip trailing `-`

Examples:
- `"Login schlägt bei leeren Passwörtern fehl"` → `0007-login-schlaegt-bei-leeren-passwoertern`
- `"Fix auth bug"` → `0001-fix-auth-bug`

**The filename never changes after creation**, even if the title is edited.

### ID Generation

Read all existing `.issues/*.md` files, parse their `id` frontmatter field, return `max(ids) + 1`. Start at 1 if no issues exist. IDs are never reused, never derived from filename.

---

## Store Layer (`internal/issue/store.go`)

```go
// Must implement:
func LoadAll(issuesDir string) ([]*Issue, error)
func LoadByID(issuesDir string, id int) (*Issue, error)
func Save(issuesDir string, issue *Issue) error   // writes file, updates `updated` field
func Delete(issuesDir string, id int) error
func NextID(issuesDir string) (int, error)
func IssuesDir() (string, error)  // finds .issues/ by walking up from cwd
```

`IssuesDir()` behavior: walk up from current working directory until finding `.issues/` or hitting filesystem root. Error if not found. This mirrors how `git` finds `.git/`.

`Save()` must:
1. Set `updated` to today's date
2. Marshal frontmatter as YAML
3. Write `---\n{yaml}---\n{body}\n`
4. If `auto_stage: true` in config, call `git add {filepath}`

---

## Config (`internal/config/config.go`)

File: `.issues/.config.yml`

```yaml
default_priority: medium
auto_stage: true
labels:
  - bug
  - feature
  - auth
  - security
  - docs
```

```go
type Config struct {
    DefaultPriority string   `yaml:"default_priority"`
    AutoStage       bool     `yaml:"auto_stage"`
    Labels          []string `yaml:"labels"`
}

func Load(issuesDir string) (*Config, error)
func Default() *Config  // returns sensible defaults if no config file exists
```

Default values: `default_priority: medium`, `auto_stage: true`.

---

## Git Integration (`internal/git/stage.go`)

```go
func Stage(filepath string) error {
    cmd := exec.Command("git", "add", filepath)
    cmd.Dir = filepath.Dir(filepath)
    return cmd.Run()
}
```

Only called when `config.AutoStage == true`. Fail silently (warn, don't error) if not inside a git repo.

---

## Commands

### `issues init`

```
issues init
```

Creates `.issues/` directory in current working directory with:
- `.issues/.config.yml` (default config)
- `.issues/.agent.md` (agent context, see Agent Docs section)
- `.issues/.gitignore` containing nothing (empty, so the dir is tracked)

Errors if `.issues/` already exists.

---

### `issues new`

```
issues new [--title "..."] [--priority low|medium|high|critical]
           [--label <label>] (repeatable)
           [--blocks <id>] (repeatable)
           [--depends-on <id>] (repeatable)
           [--related-to <id>] (repeatable)
           [--body "..."]
```

Behavior:
1. If `--title` not provided, open `$EDITOR` with template (see Editor Template below)
2. Assign next ID
3. Generate filename from title
4. Set `status: open`, `priority` from flag or config default, `created/updated` to today
5. Write file
6. For each relation flag, call the bidirectional sync (same as `issues relate`)
7. Stage if auto_stage

**Editor Template:**
```markdown
---
title: ""
priority: medium
labels: []
relations:
  blocks: []
  depends-on: []
---

<!-- Describe the issue here. Delete this line. -->
```

After editor closes, parse the file. If title is empty, abort with error.

**Output:**
```
Created: .issues/0007-login-schlaegt-bei-leeren-passwoertern.md (#7)
```

---

### `issues list`

```
issues list [--status open|in-progress|closed|wontfix|all]
            [--priority low|medium|high|critical]
            [--label <label>] (repeatable, OR logic)
            [--blocks <id>]
            [--depends-on <id>]
            [--sort priority|id|updated|created]  (default: priority desc, id asc)
            [--format table|json|ids]
```

Default: `--status open`, `--format table`

**Priority sort order:** critical > high > medium > low

**Table format:**
```
ID    PRI       STATUS       TITLE                                         LABELS
0003  critical  open         Auth-Refactor abschließen                     [auth]        ⬛ 7,15
0007  high      open         Login schlägt bei leeren Passwörtern fehl     [bug, auth]   ⬜ 3
0001  medium    in-progress  Session Timeout konfigurierbar machen         [feature]
```

Legend shown below table if any symbols present:
```
⬛ = blocks open issues   ⬜ = blocked by open issue
```

**JSON format:** Array of full Issue objects as JSON.

**IDs format:** Space-separated IDs, one line. For scripting:
```
3 7 1
```

---

### `issues show <id>`

```
issues show 7
```

Output:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Issue #7  ·  high  ·  open
Login schlägt bei leeren Passwörtern fehl
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Labels:   bug, auth
Created:  2026-03-04
Updated:  2026-03-04

Relations
  Blocks:     #12 Session Timeout konfigurierbar machen  [in-progress]
              #15 Password-Reset Flow                    [open]
  Depends on: #3  Auth-Refactor abschließen              [open]  ← BLOCKER

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Freitext Markdown body hier.
```

`← BLOCKER` is shown next to any `depends-on` entry where the referenced issue is still `open` or `in-progress`.

If issue not found: `Error: issue #99 not found` (exit code 1).

---

### `issues edit <id>`

```
issues edit 7
```

Opens `$EDITOR` with the raw `.md` file. After save:
1. Re-parse frontmatter
2. Validate required fields (id, title, status, priority)
3. Sync any relation changes bidirectionally (diff old vs new relations, add/remove accordingly)
4. Update `updated` field
5. Save + stage

---

### `issues close <id> [--wontfix] [--reason "..."]`

```
issues close 7
issues close 7 --wontfix --reason "Kein valider Use-Case"
```

1. If issue has open `depends-on` blockers, warn:
   ```
   Warning: #7 has open blocker: #3 (Auth-Refactor abschließen)
   Close anyway? [y/N]
   ```
2. Set `status: closed` (or `wontfix`), set `closed: <today>`
3. If `--reason` provided, append to body:
   ```markdown
   ## Closed
   Kein valider Use-Case
   ```
4. Save + stage

Output: `Closed: #7`

---

### `issues reopen <id>`

```
issues reopen 7
```

Sets `status: open`, removes `closed` field. Save + stage.

Output: `Reopened: #7`

---

### `issues set <id> <field> <value>`

```
issues set 7 priority critical
issues set 7 status in-progress
issues set 7 title "Neuer Titel"
issues set 7 label +security        ← add label
issues set 7 label -bug             ← remove label
```

Valid fields: `priority`, `status`, `title`, `label`

For `label`: prefix `+` to add, `-` to remove. Without prefix, replace all labels (value is comma-separated).

Validate enum values for `status` and `priority`. Error on invalid value.

Output: `Updated #7: priority = critical`

---

### `issues relate <id> <relation> <target-id>`

```
issues relate 7 blocks 12
issues relate 7 depends-on 3
issues relate 7 related-to 5
issues relate 7 duplicates 2
```

Valid relations: `blocks`, `depends-on`, `related-to`, `duplicates`

**Inverse mapping (both sides written atomically):**

| Relation | Written to source | Inverse written to target |
|---|---|---|
| `blocks` | `blocks: [target]` | `depends-on: [source]` |
| `depends-on` | `depends-on: [target]` | `blocks: [source]` |
| `related-to` | `related-to: [target]` | `related-to: [source]` |
| `duplicates` | `duplicates: [target]` | `duplicates: [source]` |

Deduplication: don't add if already present. Both files staged.

Error if either ID not found.

Output:
```
Related: #7 blocks #12
         #12 depends-on #7 (auto)
```

---

### `issues unrelate <id> <relation> <target-id>`

```
issues unrelate 7 blocks 12
```

Removes both sides of the relation. Both files staged.

Output:
```
Removed: #7 blocks #12
         #12 depends-on #7 (auto)
```

---

### `issues labels`

```
issues labels [--sort count|alpha]
```

Lists all labels used across all issues with frequency. Default sort: count desc.

Output:
```
auth       5
bug        4
feature    2
security   1
```

---

### `issues graph`

```
issues graph [--open-only] [--root <id>]
```

Renders ASCII dependency graph. `--open-only` hides closed issues.
`--root <id>` shows only the subgraph reachable from that issue.

Output:
```
#3  Auth-Refactor abschließen [open, critical]
└── blocks:
    #7  Login empty password [open, high]
        └── blocks:
            #12 Session Timeout [in-progress, medium]
            #15 Password-Reset [open, high]

#8  Unrelated feature [open, low]
    (no relations)
```

Cycle detection: if a cycle is found, print warning and break at the repeated node:
```
Warning: cycle detected at #7, truncating
```

---

## Relations Internal Implementation

```go
// internal/issue/relations.go

func AddRelation(store Store, sourceID int, relation string, targetID int) error {
    source, _ := store.LoadByID(sourceID)
    target, _ := store.LoadByID(targetID)

    addToRelation(source, relation, targetID)
    addToRelation(target, inverse(relation), sourceID)

    store.Save(source)
    store.Save(target)
    return nil
}

func inverse(relation string) string {
    switch relation {
    case "blocks":     return "depends-on"
    case "depends-on": return "blocks"
    default:           return relation  // related-to, duplicates are self-inverse
    }
}

func addToRelation(issue *Issue, relation string, id int) {
    // get the slice, append id if not already present, set back
}
```

---

## Validation

Validate on every write:

| Field | Rule |
|---|---|
| `id` | Integer > 0, unique |
| `title` | Non-empty string |
| `status` | One of: `open`, `in-progress`, `closed`, `wontfix` |
| `priority` | One of: `low`, `medium`, `high`, `critical` |
| `labels` | Array of strings, may be empty |
| `relations.*` | Arrays of valid existing IDs (warn, don't error on missing) |
| `created` | Valid ISO date |

On validation error: print specific message, exit code 1, do not write file.

---

## Error Handling

- Issue not found → `Error: issue #X not found` (exit 1)
- Not in a git-issues repo → `Error: no .issues/ directory found (run 'issues init')` (exit 1)
- Invalid field value → `Error: invalid status "foo", must be one of: open, in-progress, closed, wontfix` (exit 1)
- Editor exits without saving → `Aborted.` (exit 0)
- git add fails → `Warning: could not stage file (not a git repo?)` (continue, don't fail)

---

## Agent Documentation (`.issues/.agent.md`)

Generated by `issues init`. Content is fixed:

```markdown
# git-issues agent context

## Schema
Each issue is a Markdown file with YAML frontmatter.

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

## Commands
issues list --format ids --status open        # space-separated open IDs
issues list --format ids --priority high      # filter by priority
issues show <id>                              # full issue with resolved relations
issues set <id> status in-progress           # mark as started
issues set <id> status open                  # revert to open
issues close <id>                            # mark done
issues close <id> --wontfix                  # mark as won't fix
issues relate <id> blocks <id>               # declare blocking relation
issues graph --open-only                     # dependency tree

## Recommended workflow
1. issues list --format ids --status open → get IDs
2. issues show <id> → check for open blockers in "Depends on"
3. If blockers exist, work on those first
4. issues set <id> status in-progress → claim the issue
5. Do the work
6. issues close <id> → done, auto-staged for next commit
```

---

## README.md

Generate a concise README covering: install, init, basic workflow, command reference (one-liner per command). No marketing language.

---

## Testing

Write tests for:

1. **Slug generation** – umlauts, special chars, truncation
2. **ID generation** – empty dir returns 1, gaps in IDs handled correctly
3. **Frontmatter round-trip** – parse then serialize produces identical YAML
4. **Relation sync** – `relate 7 blocks 12` writes both files correctly
5. **Relation inverse** – all four relation types
6. **List filtering** – by status, priority, label (single and combined)
7. **Priority sort order** – critical > high > medium > low
8. **Close with blocker** – issues with open depends-on triggers warning
9. **Graph cycle detection** – doesn't infinite loop

Test using Go's standard `testing` package. No external test framework needed.

---

## Build & Install

```makefile
# Makefile
build:
    go build -o issues ./cmd/issues

install: build
    cp issues /usr/local/bin/issues

test:
    go test ./...

lint:
    go vet ./...
```

---

## Implementation Order

Suggested order to build incrementally:

1. `internal/issue/issue.go` – struct and frontmatter parse/write
2. `internal/issue/store.go` – LoadAll, LoadByID, Save, NextID, IssuesDir
3. `internal/issue/slug.go` – slug generation
4. `internal/config/config.go` – config load with defaults
5. `internal/git/stage.go` – git add wrapper
6. `cmd/root.go` + `main.go` – cobra setup
7. `issues init`
8. `issues new` (with --title flag first, editor later)
9. `issues list` (table format first, then json/ids)
10. `issues show`
11. `issues close` + `issues reopen`
12. `issues set`
13. `internal/issue/relations.go` + `issues relate` + `issues unrelate`
14. `issues edit` (editor integration)
15. `issues labels`
16. `issues graph`
17. Tests
18. README

---

## Out of Scope

Do not implement:
- Comments as separate objects (body sections are enough)
- Assignees (status is the assignee proxy)
- Milestones (use labels)
- Any network calls
- GitHub sync
- Color output (keep it simple, add later)
- Fuzzy search
- Interactive TUI
```

