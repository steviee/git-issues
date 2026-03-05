#!/bin/bash
set -e

# Recommended: Install as a Claude Code plugin instead:
#   /plugin marketplace add steviee/git-issues
#   /plugin install git-issues

SKILL_DIR="$HOME/.claude/skills/git-issues"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

mkdir -p "$SKILL_DIR"
cp "$REPO_ROOT/skills/git-issues/SKILL.md" "$SKILL_DIR/SKILL.md"

echo "Installed git-issues skill to $SKILL_DIR/SKILL.md"
echo ""
echo "The skill is now available in Claude Code as /git-issues"
echo "It will also auto-activate when planning tasks in projects with .issues/"
echo ""
echo "Tip: You can also install as a plugin for automatic updates:"
echo "  /plugin marketplace add steviee/git-issues"
echo "  /plugin install git-issues"
