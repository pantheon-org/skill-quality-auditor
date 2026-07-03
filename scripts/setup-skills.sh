#!/usr/bin/env bash
# Links pantheon-org helper skills from plugins/ to .claude/skills/ for
# Claude Code discovery. Tessl-managed registry skills are handled by tessl install.
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
PLUGINS_DIR="$REPO_ROOT/plugins/pantheon-org"
CLAUDE_SKILLS="$REPO_ROOT/.claude/skills"

mkdir -p "$CLAUDE_SKILLS"

linked=0
for skill_dir in "$PLUGINS_DIR"/*/; do
    [[ -d "$skill_dir" ]] || continue
    skill_name="$(basename "$skill_dir")"
    target="$CLAUDE_SKILLS/$skill_name"
    if [[ ! -e "$target" ]]; then
        ln -sf "../../plugins/pantheon-org/$skill_name" "$target"
        echo "  linked .claude/skills/$skill_name"
        linked=$((linked + 1))
    fi
done

if [[ $linked -gt 0 ]]; then
    echo "Skills linked: $linked"
else
    echo "Skills: all up to date"
fi
