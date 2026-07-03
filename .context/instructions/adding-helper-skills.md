---
title: "Adding Helper Skills"
type: instruction
status: active
date: 2026-07-03
---

# Adding Helper Skills

This repo hosts **two kinds of skills**:

1. **Audited skills** — the `skill-quality-auditor` Tessl tile itself, at `cmd/assets/`. These are the Go CLI's self-skill assets (SKILL.md, evals, references). Evaluated by `./dist/skill-auditor eval ./cmd/assets`.

2. **Helper skills** — independent Tessl plugin skills under `plugins/` that provide agent workflows (e.g. socratic-method, context-file, rules-management). These are **not** evaluated by the Go CLI, but are available to AI agents working in the repo.

This document covers helper skills only.

## Directory structure

Each helper skill lives in its own directory under `plugins/`:

```
plugins/
└── <workspace>/<skill-name>/
    ├── tessl-package.json    # name + version
    ├── tile.json             # tile metadata, skill entries
    └── SKILL.md              # the skill content
```

## How to add a new helper skill

### 1. Choose a name

```
plugins/pantheon-org/<skill-name>/
```

Use `pantheon-org` as the workspace for skills maintained in this repo.
If the skill originates from the Tessl registry (e.g. `pantheon-ai/foo`), install it there and optionally fork to `pantheon-org/`.

### 2. Create the plugin files

Minimal required structure:

- **`tessl-package.json`** — package identity
- **`tile.json`** — tile metadata linking to the skill
- **`SKILL.md`** — the skill content (YAML frontmatter + markdown body)

Optionally include a `references/` directory for supporting docs linked from SKILL.md.

### 3. Register in `tessl.json`

Add the plugin to `tessl.json` dependencies using a `file:` source:

```json
"pantheon-org/<skill-name>": {
  "version": "0.1.0",
  "source": "file:plugins/pantheon-org/<skill-name>"
}
```

### 4. Install

```bash
tessl install
```

This copies the plugin into `.tessl/plugins/` and makes it available to AI agents.

### 5. Branch and commit

Branch from `main`, commit with a conventional message:

```bash
git checkout -b feat/tessl-skills
git add plugins/ tessl.json
git commit -m "feat(skills): add <skill-name> helper skill"
```

### 6. Publish (optional)

To publish a helper skill to the Tessl registry so it can be used outside this repo:

```bash
tessl login
# Create the workspace on tessl.io first, then:
tessl publish plugins/pantheon-org/<skill-name>
```

Update `tessl.json` to switch from `file:` source to the published version once published.

## Best practices

- **Keep SKILL.md focused on agent behaviour** — helper skills are consumed by AI agents, not humans. Use clear trigger descriptions, When-to-use/When-NOT-to-use sections, and concrete examples.
- **Include a `references/` directory** for deeper supporting material that would bloat the main SKILL.md.
- **Don't duplicate the audited skill** — helper skills are complementary. The `cmd/assets/` skill is the one that gets evaluated and version-bumped with releases.
- **Evaluate with the skill-quality-auditor** — you can still run `./dist/skill-auditor evaluate plugins/pantheon-org/<skill-name>` to score the helper skill against the quality framework.
