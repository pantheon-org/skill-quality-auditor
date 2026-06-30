#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
AGENTS_SKILLS="$REPO_ROOT/.agents/skills"
CLAUDE_SKILLS="$REPO_ROOT/.claude/skills"

mkdir -p "$CLAUDE_SKILLS"

linked=0
for skill_dir in "$AGENTS_SKILLS"/*/; do
    [[ -d "$skill_dir" ]] || continue
    [[ -L "$skill_dir" ]] && continue  # skip tessl-managed symlinks
    skill_name="$(basename "$skill_dir")"
    target="$CLAUDE_SKILLS/$skill_name"
    if [[ ! -e "$target" ]]; then
        ln -sf "../../.agents/skills/$skill_name" "$target"
        echo "  linked .claude/skills/$skill_name"
        linked=$((linked + 1))
    fi
done

if [[ $linked -gt 0 ]]; then
    echo "Skills linked: $linked"
else
    echo "Skills: all up to date"
fi
