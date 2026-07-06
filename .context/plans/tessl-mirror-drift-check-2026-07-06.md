---
title: "Plan: CI check for .tessl/plugins/pantheon-org mirror drift"
type: plan
status: active
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
- Enforcement mode (hard-fail from day one vs. an initial advisory rollout window) is a
  maintainer call — see Decisions. ADR-047 frames this as enforcement in spirit (unlike
  `docs-drift`/`plan-drift`'s permanent advisory mode), and zero divergence exists today
  across all 12 `pantheon-org` skills (verified during the `design-debate` that produced
  ADR-047), so there is nothing to grandfather regardless of which mode is chosen.
- `quality-gate`'s trigger paths gain `scripts/check-tessl-mirror-drift.sh` and
  `.github/workflows/skill-quality.yml` itself, alongside the existing `cmd/assets/**`
  and `.context/plugins/**` — otherwise the PR that lands this check wouldn't trigger
  the job it adds a step to, and future edits to the script or its wiring would
  silently stop being re-verified.

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
   - Diff-only, no install side effects — the script assumes `.tessl/plugins/pantheon-org/**`
     is already populated by a prior `tessl install` (that's the CI step's job, per Phase 2,
     not the script's), so it can be re-run repeatedly against the same `.tessl/` state.
   - For each skill directory under `.context/plugins/pantheon-org/<domain>/<skill>/`, run
     `diff -rq --exclude=evals <source-dir> <installed-dir>` — content-only comparison;
     timestamps and permissions are not compared.
   - Treat a missing directory on either side as divergence: a source skill with no
     installed counterpart (newly added, not yet installed) and an installed skill with
     no source counterpart (deleted from source, stale in the mirror) both fail.
   - Print a clear diff summary per divergent skill (the skill name plus its `diff -rq`
     output); exit 1 if any divergence found, exit 0 otherwise.
2. `shellcheck scripts/check-tessl-mirror-drift.sh` clean.

Exit criterion: running the script locally against the current (non-divergent) repo
state exits 0 with no output; introducing a throwaway one-line edit to a source SKILL.md
without re-running `tessl install` causes it to exit 1 with a clear message naming the
divergent skill; a throwaway skill directory that exists on only one side is also caught
(see Verification).

### Phase 2 — Spike: verify `tessl` CLI works tokenless in clean CI

Gates Phase 3 — do not write the CI-wiring step until this goes green (Decisions).

1. Trigger a throwaway `workflow_dispatch` job (or a scratch run reusing an existing
   runner) that does nothing but `npm install -g tessl && tessl install` against a clean
   checkout, with no `TESSL_TOKEN` secret exposed.
2. Run it via `gh workflow run` / `gh run watch`; confirm it exits 0 with
   `pantheon-org/**` skills present under `.tessl/plugins/`.
3. If it fails needing a token, stop and resolve an auth path before Phase 3 starts. If
   it succeeds, remove the throwaway job (or fold its two commands straight into Phase
   3's step) and proceed.

Exit criterion: a real GitHub Actions run — not just local `tessl install --help` —
confirms `tessl install` succeeds with no `TESSL_TOKEN` in an environment matching
`quality-gate`'s runner.

### Phase 3 — CI wiring

1. Add a step to `skill-quality.yml`'s `quality-gate` job, after the existing helper-skill
   structural-eval step, that: (a) ensures the `tessl` CLI is available via
   `npm install -g tessl` — confirmed tokenless by Phase 2's spike, (b) removes any
   pre-existing `.tessl/plugins/pantheon-org` to avoid stale state masking a real
   divergence, (c) runs `tessl install`, (d) runs `scripts/check-tessl-mirror-drift.sh`
   (diff-only, per Phase 1) with `continue-on-error: true` for the advisory rollout
   (Decisions).
2. Widen `quality-gate`'s trigger paths to include `scripts/check-tessl-mirror-drift.sh`
   and `.github/workflows/skill-quality.yml` (see Scope) so this landing PR — and any
   future edit to the script or its wiring — actually exercises the check.
3. Gate the new step on `if: steps.helper_skills.outputs.count != '0'`, matching the
   existing sibling-step convention (Decisions).

Exit criterion: a live PR shows the new step passing (with `continue-on-error: true`) on
current `main`, and a deliberately-broken throwaway PR (source edited, mirror not
reinstalled) shows it failing with a clear message while not blocking the merge.

### Phase 4 — Verify and land

1. Open a PR with the new script + workflow wiring.
2. Confirm `quality-gate` passes on the PR itself — the widened path filter (Phase 3,
   step 2) means this PR's own diff (`scripts/**`, `.github/workflows/skill-quality.yml`)
   now triggers the job directly, no `workflow_dispatch` workaround needed.
3. Mark this plan `status: done` once merged and confirmed passing — but leave one
   open follow-up noted below (the hard-fail flip) rather than closing it silently.

## Decisions (resolved during plan-review)

- **Enforcement mode: advisory rollout, then flip.** The new step runs with
  `continue-on-error: true` initially rather than hard-failing from day one. **Revisit
  trigger:** after 5 merged PRs that exercise this step (or 2 weeks, whichever comes
  first) with zero false positives, open a follow-up PR removing `continue-on-error` and
  hard-fail per ADR-047's original intent. Recorded here, not left implicit, so
  "advisory" doesn't silently become "permanent" the way the known-issue on
  known-issues-lacking-enforcement warned against.
- **Gating: match the sibling convention.** The new step is gated on
  `if: steps.helper_skills.outputs.count != '0'`, same as the existing helper-skill
  steps in this job, rather than running unconditionally. Trade-off accepted: a PR
  touching only `cmd/assets/**` won't re-verify the mirror on that run, in exchange for
  consistency with how every other step in this section already behaves.
- **`tessl` CLI provisioning: gated spike first.** Phase 2 verifies `tessl install`
  succeeds tokenless in a clean CI runner before Phase 3's CI-wiring step is written,
  rather than discovering an auth requirement mid-implementation.

## Risks

- **`tessl` version drift**: `mise.toml` pins `"npm:tessl" = "latest"`, so CI tracks
  whatever's newest at each run — consistent with local dev, but an upstream change to
  `tessl install`'s file-filtering (e.g. what's excluded besides `evals/`) could produce
  a false positive with no advance signal. Accepted tradeoff, not pinned independently,
  to avoid a second place `tessl`'s version could drift from `mise.toml`. Caught safely
  during the advisory window (above) rather than blocking a merge.
- **Stale `.tessl` state**: mitigated by the CI step's `rm -rf .tessl/plugins/pantheon-org`
  before each `tessl install` (Phase 3, step 1b) — without it, a prior run's leftover
  files could mask a real divergence.
- **Advisory silently becoming permanent**: the biggest residual risk after the
  Decisions above isn't a false positive — it's nobody remembering to flip
  `continue-on-error` off. Mitigated by the concrete revisit trigger stated above; if
  that PR hasn't happened by the time this plan is reviewed again, treat it as overdue.

## Verification

```bash
# Phase 1
scripts/check-tessl-mirror-drift.sh              # expect: exits 0, no divergence today
shellcheck scripts/check-tessl-mirror-drift.sh

# Introduce a throwaway divergence, confirm it's caught, then discard
echo "" >> .context/plugins/pantheon-org/planning/plan-create/SKILL.md
scripts/check-tessl-mirror-drift.sh              # expect: exits 1, names the divergent skill
git checkout -- .context/plugins/pantheon-org/planning/plan-create/SKILL.md

# Asymmetric case: a skill directory present on only one side, then discard
mkdir -p .tessl/plugins/pantheon-org/workshop/throwaway-test-skill
scripts/check-tessl-mirror-drift.sh              # expect: exits 1, names throwaway-test-skill (installed-only)
rm -rf .tessl/plugins/pantheon-org/workshop/throwaway-test-skill

# Phase 2 (spike)
gh workflow run skill-quality.yml                # or a scratch workflow_dispatch job
gh run watch                                     # confirm tessl install succeeds tokenless

# Phase 3 + 4
gh workflow run skill-quality.yml                # or push a PR
gh run watch                                     # confirm the new step passes (continue-on-error)
```
