---
title: "ADR-029: Local skills migrated to .context/plugins/ as Tessl file-based plugins"
status: accepted
date: 2026-07-03
context:
  - path: .context/instructions/adding-helper-skills.md
  - path: .context/plugins/pantheon-org
---

**Status:** Accepted
**Date:** 2026-07-03

## Context

Six locally-developed skills (adr-capture, context-file, context-index, docs-check,
rules-management, session-reflection) lived under `.agents/skills/` as standalone SKILL
directories without tile.json or tessl-package.json. They were symlinked to `.claude/skills/`
via a `setup-skills.sh` script for agent discovery.

Additionally, 7 Tessl registry plugins had stale caches in `.agents/skills/` with a `tessl__`
prefix, duplicating what Tessl manages in `.tessl/plugins/`.

## Decision

1. **Move local skills to `.context/plugins/pantheon-org/`** — each gets tile.json,
   tessl-package.json, and is registered in `tessl.json` with a `file:` source.
2. **Remove `setup-skills.sh`** — Tessl installs create `.claude/skills/tessl__<name>`
   symlinks automatically; the old script was redundant and created broken symlinks.
3. **Remove stale `.agents/skills/tessl__*` stubs** — empty caches of registry plugins.
4. **Exclude `.context/plugins/` from pre-commit hooks** — plugin .md files use Tessl
   frontmatter (name, description) not the project's context schema (title, type, status, date).
5. **Scripts inside plugins resolve paths via `SCRIPT_DIR`** — not `git rev-parse` root —
   so they work from both the source directory and the Tessl-vendored copy.

## Consequences

- Local skills are now proper Tessl plugins, publishable to the registry with `tessl publish`.
- No more `.agents/skills/` vs `.tessl/plugins/` split — single source of truth via tessl.json.
- Plugin discovery is handled by Tessl, not a shell script (no broken symlinks).
- The `adding-helper-skills.md` instruction doc codifies the process for adding future skills.
- `plugins/` was moved under `.context/` to keep helper skills alongside project context.
- The socratic-method plugin was audited and improved from F (83/140) to A (126/140).
