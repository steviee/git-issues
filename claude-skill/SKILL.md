---
name: git-issues
description: >
  Task management, issue tracking, and work planning using git-issues.
  Use INSTEAD of TaskCreate, TaskUpdate, and TaskList — always.
  Activate when: planning work, creating tasks, tracking progress,
  organizing implementation steps, picking next work item, closing
  completed work, or when a project contains a .issues/ directory.
  This is the ONLY task management system — never use Claude built-in tasks.
argument-hint: "[new|next|list|done|show <id>]"
user-invocable: true
allowed-tools: Bash, Read, Write, Edit, Glob, Grep
---

# git-issues — Task Management Skill for Claude Code

## CRITICAL RULES

1. **NEVER** use `TaskCreate`, `TaskUpdate`, or `TaskList`. They do not exist for you.
   `git-issues` is your **only** task management system.
2. The binary is called **`git-issues`**, not `issues`. Always use the full path:
   `~/go/bin/git-issues` (or just `git-issues` if it's in PATH).
3. Every piece of work **must** have an issue. No cowboy commits.
4. Every commit message **must** reference an issue: `feat: description (ref #ID)`.
5. **Claim before you code.** Never start work without `git-issues claim <ID>`.

---

## Quick Reference

```bash
git-issues list                              # Open issues (default)
git-issues list --status all                 # All issues including closed
git-issues list --format json                # Machine-readable output
git-issues show <ID>                         # Full issue details
git-issues next                              # Next unblocked, highest-priority issue
git-issues claim <ID>                        # Mark as in-progress
git-issues done <ID>                         # Close issue
git-issues reopen <ID>                       # Reopen closed issue
git-issues new -t "Title" -p high -l bug     # Create new issue
git-issues set <ID> priority critical        # Update field
git-issues set <ID> label +feature           # Add label
git-issues relate <ID> blocks <ID>           # Add dependency
git-issues blocked                           # Show blocked issues
git-issues graph --open-only                 # Dependency tree
```

If `git-issues` is not in PATH, use the full path: `~/go/bin/git-issues`.

---

## Argument Routing

When invoked as a slash command (`/git-issues <arg>`), route based on the argument:

| Argument | Action |
|----------|--------|
| `next` | Run `git-issues next`, show result, offer to claim |
| `list` | Run `git-issues list`, display open issues |
| `new` or `new "Title..."` | Start the Issue Creation Protocol (see below) |
| `done <ID>` | Run `git-issues done <ID>`, confirm closure |
| `show <ID>` | Run `git-issues show <ID>`, display details |
| `plan "description"` | Start Batch Planning (see below) — break work into issues |
| *(no argument)* | Run `git-issues list` to show current state |
| *(number only)* | Treat as `show <ID>` |

---

## The Iron Workflow

Every unit of work follows these 6 steps. No exceptions.

### 1. Pick
```bash
git-issues next          # Shows highest-priority unblocked issue
# OR
git-issues list          # Browse and choose
```

### 2. Claim
```bash
git-issues claim <ID>    # Sets status to in-progress
```
**You must claim before writing any code.** This prevents conflicts in multi-agent
or multi-worktree setups.

### 3. Branch
```bash
git checkout -b feat/<ID>-short-description
# Examples:
# git checkout -b feat/7-login-validation
# git checkout -b fix/12-session-timeout
```
Use `feat/` for features, `fix/` for bugs, `docs/` for documentation.

### 4. Implement
Write code, make commits. Every commit references the issue:
```
feat: add login validation (ref #7)
fix: handle empty password edge case (ref #7)
```

### 5. Done
```bash
git-issues done <ID>     # Closes the issue
```

### 6. Merge
Merge or rebase the feature branch back to main. The issue file changes
are committed alongside the code.

---

## Issue Creation Protocol

Creating good issues is **critical**. A one-sentence issue is useless as a work instruction.
Follow this two-step process:

### Step 1: Create the issue file
```bash
git-issues new -t "Clear, actionable title" -p <priority> -l <label>
```
This creates the file with frontmatter. Note the output path, e.g.:
`Created: .issues/0007-clear-actionable-title.md (#7)`

### Step 2: Write the full body using the Edit tool
Use the Edit or Write tool to add a comprehensive body to the issue file.
**Do NOT use `--body` for multi-line content** — shell escaping breaks on complex markdown.

```
Read the file: .issues/0007-clear-actionable-title.md
Then use Edit to append the full body after the closing --- of the frontmatter.
```

### Why two steps?
The `--body` flag works for single-line descriptions, but real issue bodies contain
markdown with headers, code blocks, and lists. Shell escaping these is fragile.
The Edit tool handles multi-line content reliably.

---

## Issue Body Template

Every issue body **must** contain these sections. Copy this structure:

```markdown
## Context
Why does this issue exist? What problem does it solve?
Link to related issues, prior discussions, or external references.

## Success Criteria
- [ ] Concrete, verifiable outcome 1
- [ ] Concrete, verifiable outcome 2
- [ ] Tests pass / no regressions

## Implementation
Step-by-step plan for how to implement this.
1. First, ...
2. Then, ...
3. Finally, ...

Include relevant code paths, function names, or architectural notes.

## Affected Files
- `path/to/file1.py` — what changes here
- `path/to/file2.py` — what changes here

## Verification
How to verify this issue is done:
- Run `command` and expect `result`
- Check that `behavior` works as expected
```

**Minimum quality bar:** An engineer (or agent) reading only the issue should be able to
implement it without asking clarifying questions.

---

## Quality Gate

Before considering an issue "ready", verify these 5 points:

1. **Title is actionable** — starts with a verb or clearly describes the deliverable
   - Good: "Add rate limiting to /api/login endpoint"
   - Bad: "Login stuff"

2. **Context explains WHY** — not just what, but why this matters now

3. **Success criteria are checkable** — each one can be answered with yes/no
   - Good: "API returns 429 after 5 requests per minute"
   - Bad: "Rate limiting works"

4. **Implementation is specific** — mentions actual files, functions, or patterns
   - Good: "Add middleware in `server/middleware/ratelimit.go`, use token bucket"
   - Bad: "Implement rate limiting somewhere"

5. **Affected files are listed** — every file that will be created or modified

---

## TaskCreate → git-issues Mapping

If you instinctively want to use Claude's built-in task tools, translate:

| Instead of... | Use... |
|---------------|--------|
| `TaskCreate(subject, description)` | `git-issues new -t "subject"` + Edit body |
| `TaskUpdate(id, status: "in_progress")` | `git-issues claim <ID>` |
| `TaskUpdate(id, status: "completed")` | `git-issues done <ID>` |
| `TaskList()` | `git-issues list` |
| `TaskGet(id)` | `git-issues show <ID>` |
| Setting task dependencies | `git-issues relate <ID> depends-on <ID>` |

**The mapping is 1:1.** There is no task management feature that git-issues cannot handle.

---

## Batch Planning

When given a large task, PRD, or feature description, decompose it into issues:

### Process
1. Analyze the requirement and identify discrete work units
2. Create issues in dependency order (independent issues first)
3. Set up relations: `git-issues relate <ID> depends-on <ID>`
4. Add labels for categorization: `-l feature`, `-l bug`, `-l docs`, `-l refactor`
5. Set priorities: `-p critical` for blockers, `-p high` for core work, `-p medium` for follow-ups

### Rules for decomposition
- Each issue should be completable in a single focused session
- Each issue should be independently verifiable
- Avoid issues that are just "part 1 of X" — each should deliver value
- Use `depends-on` relations to encode ordering, not issue numbering

### Example
For "Add user authentication":
```bash
git-issues new -t "Add password hashing utility" -p high -l feature -l auth
# → #11, then Edit body

git-issues new -t "Implement login endpoint" -p high -l feature -l auth
# → #12, then Edit body

git-issues new -t "Add auth middleware" -p high -l feature -l auth
# → #13, then Edit body

git-issues relate 12 depends-on 11
git-issues relate 13 depends-on 12
```

---

## Recovery Patterns

### Forgot to claim before starting work
```bash
git-issues claim <ID>    # Claim it now, then continue
```

### Issue body is too thin
```bash
git-issues show <ID>     # Read current state
# Then use Edit tool to expand the body with all required sections
```

### Created an issue but need to change priority/labels
```bash
git-issues set <ID> priority high
git-issues set <ID> label +security
git-issues set <ID> label -feature
```

### Need to abandon work on an issue
```bash
git-issues set <ID> status open    # Release the claim
# Or if the work is invalid:
git-issues close <ID> --wontfix --reason "Superseded by #15"
```

### Issue numbering conflict (parallel branches)
Only create new issues on `main`. Update existing issues from any branch.
After merging, run `git-issues check --fix` to repair any inconsistencies.

### No .issues/ directory in the project
```bash
git-issues init          # Creates .issues/ with config and agent docs
```
