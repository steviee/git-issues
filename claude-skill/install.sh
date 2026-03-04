#!/bin/bash
set -e

SKILL_DIR="$HOME/.claude/skills/git-issues"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

mkdir -p "$SKILL_DIR"
cp "$SCRIPT_DIR/SKILL.md" "$SKILL_DIR/SKILL.md"

echo "Installed git-issues skill to $SKILL_DIR/SKILL.md"
echo ""
echo "The skill is now available in Claude Code as /git-issues"
echo "It will also auto-activate when planning tasks in projects with .issues/"
