---
title: "Plan: CI check for .tessl/plugins/pantheon-org mirror drift"
type: plan
status: draft
date: 2026-07-06
effort: S
related:
  - ../findings/tessl-mirror-drift-protection-2026-07-06.md
  - ../../docs/ADR/adr-047-tessl-mirror-ephemeral-diff.md
  - ../../.github/workflows/skill-quality.yml
  - ../../scripts/check-docs-drift.sh
---

# Plan: CI check for .tessl/plugins/pantheon-org mirror drift

## Goal

Implement ADR-047: catch drift between this repo's authored helper skills
(`.context/plugins/pantheon-org/**`) and what `tessl install` actually produces
(`.tessl/plugins/pantheon-org/**`), via a CI-only check — no tracking, no `.gitignore`
change. The decision (track vs. CI-only diff) is already made via `design-debate` and is
not open for re-litigation here; this plan is implementation only.

## Scope

**In scope:**

- A new script, `scripts/check-tessl-mirror-drift.sh`, matching this repo's existing
  `check-docs-drift.sh`/`check-plan-drift.sh` conventions (bash, `set -euo pipefail`,
  shellcheck-clean).
- Wiring it into `.github/workflows/skill-quality.yml`'s `quality-gate` job, which
  already triggers on `pull_request`/`push:main` for `paths: ['cmd/assets/**',
  '.context/plugins/**']` and already has a "helper skills" section (Discover helper
  skills → duplication → batch → structural eval) this fits alongside.
- Scoped to `pantheon-org/**` only. `pantheon-ai/**` and `tessl-labs/**` (third-party
  registry content, ~174MB combined) are explicitly out of scope — not diffed, not
  installed for the purpose of this check, not touched.
- Since nothing is tracked, no historical baseline is needed: in one CI run, the source
  is already checked out, `tessl install` runs fresh, then
  `.context/plugins/pantheon-org/` is diffed directly against
  `.tessl/plugins/pantheon-org/` in that same run. A `tessl install` bug (skipped file,
  truncated content, a newly-added skill not picked up) shows up immediately.
- Excludes each skill's `evals/` subdirectory from the diff — confirmed convention this
  session: `evals/` is deliberately not part of the installed/published skill for every
  existing helper skill, including `design-debate` (authored and installed this same
  session, `evals/` correctly absent from its mirror).
- Hard-fails the build on real divergence (enforcement, per ADR-047 — not an advisory
  cumulative-mode-style check like `docs-drift`/`plan-drift`). Zero divergence exists
  today across all 12 `pantheon-org` skills (verified during the `design-debate` that
  produced ADR-047), so there is nothing to grandfather.

**Out of scope:**

- Any change to `.gitignore` or tracking `.tessl/plugins/**` in git (rejected by
  ADR-047).
- Any check of `pantheon-ai/**` or `tessl-labs/**` content.
- Replacing or modifying the existing `tesslio/setup-tessl` / `tessl review run
  cmd/assets/` block in `skill-quality.yml` — that's separate, unrelated
  infrastructure (the Tier-2 Tessl review "proving period" block, slated for removal
  once the native eval runner has 2 weeks of green runs per a different plan).

## Phases

### Phase 1 — Script

1. Write `scripts/check-tessl-mirror-drift.sh`:
   - Run `tessl install` (assumes the `tessl` CLI is already on `PATH` — see Open
     Questions for how CI provides it).
   - For each skill directory under `.context/plugins/pantheon-org/<domain>/<skill>/`,
     diff its content (excluding `evals/`) against
     `.tessl/plugins/pantheon-org/<domain>/<skill>/`.
   - Print a clear diff summary per divergent skill; exit 1 if any divergence found,
     exit 0 otherwise.
2. `shellcheck scripts/check-tessl-mirror-drift.sh` clean.

Exit criterion: running the script locally against the current (non-divergent) repo
state exits 0 with no output; introducing a throwaway one-line edit to a source SKILL.md
without re-running `tessl install` causes it to exit 1 with a clear message naming the
divergent skill.

### Phase 2 — CI wiring

1. Add a step to `skill-quality.yml`'s `quality-gate` job, after the existing helper-skill
   structural-eval step, that ensures the `tessl` CLI is available (resolution depends on
   Open Question 1) and runs `tessl install` then `scripts/check-tessl-mirror-drift.sh`.
2. Confirm the step only runs when helper skills exist in the diff (mirroring the
   existing `if: steps.helper_skills.outputs.count != '0'` guard on sibling steps) —
   or decide it should always run regardless of what changed, since the check verifies
   the whole `pantheon-org/**` mirror, not just the PR's diff (see Open Question 2).

Exit criterion: a live PR (or `workflow_dispatch` run) shows the new step passing on
current `main`, and a deliberately-broken throwaway PR (source edited, mirror not
reinstalled) shows it failing with a clear message.

### Phase 3 — Verify and land

1. Open a PR with the new script + workflow wiring.
2. Confirm `quality-gate` passes on the PR itself (the new step necessarily exercises
   itself, since the PR touches `.context/plugins/**`... actually it touches
   `scripts/` and `.github/workflows/`, not `.context/plugins/**` — confirm the
   workflow's path filter still triggers `quality-gate` for this PR, or use
   `workflow_dispatch` to force a run if not).
3. Mark this plan `status: done` once merged and confirmed passing.

## Open Questions

- **How does CI get the `tessl` CLI without depending on the soon-to-be-removed
  `tesslio/setup-tessl` action?** That action requires `TESSL_TOKEN` and is explicitly
  marked for deletion once the native eval runner proving period ends. `tessl.json`
  sources `pantheon-org/**` via `file:` paths (no registry involved), and `tessl install
  --help` runs with no auth prompt — suggesting `tessl install` for purely
  file-sourced plugins may not need `TESSL_TOKEN` at all. Not yet verified: whether a
  full `tessl install` run (not just `--help`) in a clean CI environment with no token
  succeeds. If it does, the simplest fix is `npm install -g tessl` (matching
  `mise.toml`'s `"npm:tessl" = "latest"`) with no token — decoupled entirely from the
  block slated for removal. If it doesn't, this needs its own auth path decided before
  Phase 2 can land.
- **Should the new step be gated on `helper_skills.outputs.count != '0'` like its
  siblings, or run unconditionally?** The check verifies the entire `pantheon-org/**`
  mirror's fidelity, not just whatever the current PR touched — a PR that touches
  `cmd/assets/**` only (triggering `quality-gate` via the workflow's path filter) but
  not `.context/plugins/**` would still benefit from catching a pre-existing mirror
  problem, but the existing sibling steps skip entirely when the PR touches no helper
  skills. Leaning toward "run whenever the job runs at all" for a stronger guarantee,
  but this changes the job's runtime characteristics slightly (always doing a full
  `tessl install`, not just when relevant) — a maintainer call, not decided here.
- **Hard-fail from day one, or advisory-warn for an initial rollout window?** The
  scope says hard-fail per ADR-047 and because zero divergence exists today, but this
  mechanism is entirely new and unproven in actual CI (only tested locally in this
  session). A short advisory-only window (e.g. `continue-on-error: true` for the first
  few merged PRs) would catch any CI-environment-specific surprise (e.g. the auth
  question above) without ever blocking a real merge on a false positive. Not decided
  — plan-review should weigh in.

## Verification

```bash
# Phase 1
scripts/check-tessl-mirror-drift.sh              # expect: exits 0, no divergence today
shellcheck scripts/check-tessl-mirror-drift.sh

# Introduce a throwaway divergence, confirm it's caught, then discard
echo "" >> .context/plugins/pantheon-org/planning/plan-create/SKILL.md
scripts/check-tessl-mirror-drift.sh              # expect: exits 1, names the divergent skill
git checkout -- .context/plugins/pantheon-org/planning/plan-create/SKILL.md

# Phase 2 + 3
gh workflow run skill-quality.yml                # or push a PR
gh run watch                                     # confirm the new step passes
```
