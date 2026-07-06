---
title: "Plan: create a pr-merge skill"
type: plan
status: draft
date: 2026-07-06
value: medium
effort: M
related:
  - ../findings/pr-merge-validation-gap-2026-07-06.md
  - ../plugins/pantheon-org/workshop/pr-author/SKILL.md
  - ../plugins/pantheon-org/governance/adr-capture/SKILL.md
---

# Plan: create a pr-merge skill

## Goal

Formalize the validate-and-merge pattern used ad hoc for every PR this session
(#187–#199) into a new skill, `pr-merge`, so waiting for CI, handling a rebase
conflict on an auto-generated file, and merging safely is a documented procedure
instead of something re-derived from memory each time. Per
`.context/findings/pr-merge-validation-gap-2026-07-06.md`: no existing skill covers
this — `pr-author` is description-only, `adr-capture`'s `merge-status-sync.sh` is
post-merge only.

## Scope

**In scope:**

- A new skill, `pr-merge`, under `.context/plugins/pantheon-org/workshop/pr-merge/`
  (workshop domain, alongside `pr-author`, `session-reflection`, `docs-check`).
- Waiting for every PR check to reach a terminal state (not "started" — this
  session's finding flagged `gh pr merge` succeeding while checks were still
  "pending" as a real risk, since this repo has no required-status-checks branch
  protection forcing the wait), distinguishing a real failure from an expected
  `continue-on-error` advisory step (e.g. `check-tessl-mirror-drift.sh`'s rollout
  window, the Tier-2 LLM-judge eval).
- The regenerate-don't-hand-merge sequence (Rule 15) for a rebase conflict on an
  auto-generated file (`.context/index.yaml`, `docs/ADR/index.yaml`): `git checkout
  --ours <file>` → re-run the matching regenerate script → re-verify → force-push
  with `--force-with-lease`.
- A default merge strategy (squash, matching `pr-author`'s already-documented
  preference) and branch deletion after merge, matching `ways-of-working.md`'s
  "after merge" step.
- Explicit confirmation before merging, every time — never silent, per this
  project's own risk-taking guidance on shared-state, hard-to-reverse actions.
- Registering the skill in the `workshop` domain's `tile.json`, running
  `tessl install` to sync the mirror, and eval scenarios per Rule 2.

**Out of scope:**

- PR description authoring or updates — stays with `pr-author`.
- Post-merge plan/ADR status sync — stays with `adr-capture`'s
  `merge-status-sync.sh`; `pr-merge` may suggest running it as a final step but does
  not duplicate its logic.
- Any change to branch protection rules or required-status-checks configuration on
  GitHub itself — that's a repo-settings decision, not something this skill should
  do unprompted.

## Phases

### Phase 1 — Draft SKILL.md

1. Prerequisites, When to Use / When NOT to Use, Mindset.
2. Workflow: identify the PR → wait for terminal checks (poll or `Monitor`-based,
   matching this session's pattern) → classify any non-passing check as a real
   failure or an expected advisory (`continue-on-error`) step → if a conflict
   appears, apply the regenerate-don't-hand-merge sequence → confirm with the user
   → merge (squash default) → delete branch → suggest `merge-status-sync.sh
   --dry-run`.
3. Anti-Patterns: merging on "pending," hand-resolving generated-file conflicts,
   merging without confirmation, treating every non-green check as fatal
   (over-blocking on a known advisory step).
4. Troubleshooting table for common failure shapes (conflict on generated file,
   flaky check needing re-run, required review missing).

Exit criterion: `SKILL.md` passes `validate-context-frontmatter.sh`-equivalent
structural checks used for skills (frontmatter block with `name`/`description`).

### Phase 2 — Register and validate

1. Create the skill directory under
   `.context/plugins/pantheon-org/workshop/pr-merge/`.
2. Register in `.context/plugins/pantheon-org/workshop/tile.json`'s `skills` map.
3. Run `tessl install` to sync `.tessl/plugins/pantheon-org/workshop/pr-merge/`.
4. Run `./dist/skill-auditor evaluate <path> --store`; remediate if below B,
   matching this session's precedent for `design-debate` and `plan-review`.

Exit criterion: `skill-auditor evaluate` reports B or above; `check-tessl-mirror-drift.sh` (landed in #199) reports no divergence for the new skill.

### Phase 3 — Eval scenarios

1. Write `evals/scenario-01` (all checks green — straightforward merge, confirm
   before merging).
2. Write `evals/scenario-02` (a rebase conflict lands on an auto-generated file
   mid-wait — confirm the skill applies Rule 15's sequence, not a hand-merge).
3. Write `evals/scenario-03` (a check has genuinely failed, not an advisory
   `continue-on-error` step — confirm the skill does NOT merge and reports why).
4. Run the native eval runner (`./dist/skill-auditor eval
   .context/plugins/pantheon-org/workshop/pr-merge`).

Exit criterion: all 3 scenarios pass under the native structural eval gate.

### Phase 4 — Land

1. Update `docs/development/skills-and-rules.md`'s skill table with the new
   `pr-merge` entry (avoids the docs-drift gate failure PR #196 hit for the same
   reason).
2. Open a PR; confirm `skill-quality.yml`'s `quality-gate` passes (helper-skill
   duplication/batch/structural checks, plus the mirror-drift check from #199).
3. Merge — using `pr-merge` on itself once Phase 2's eval bar is met, or manually
   if that feels too recursive for the first-ever use.

Exit criterion: PR merged, `quality-gate` green, skill discoverable via
`.context/plugins/pantheon-org/workshop/pr-merge/SKILL.md`.

## Risks

- **Over-trusting a `continue-on-error` classification.** If the skill's logic for
  "this non-green check is an expected advisory, not a real failure" is wrong or
  goes stale (e.g. `check-tessl-mirror-drift.sh` flips to hard-fail per its own
  revisit trigger and the skill isn't updated), a real failure could be waved
  through. Mitigate by keeping the classification list small and explicit in the
  skill body, not inferred generically.
- **Confirmation fatigue.** If every merge requires a full stop-and-ask, the skill
  could feel like friction rather than help on genuinely green, uncontroversial
  PRs. Accepted tradeoff — per this project's own guidance, a merge is shared-state
  and hard to reverse, so the cost of asking is intentionally lower than the cost
  of an unwanted merge.
- **Recursion on first use.** Landing this skill's own PR via itself (Phase 4) is
  circular the first time — the skill doesn't exist in the mirror until after its
  own PR merges. Not a blocker, just worth calling out so the first landing isn't
  confused for a bug.

## Open Questions

- Should `pr-merge` end by automatically suggesting (not running)
  `merge-status-sync.sh --dry-run <pr-number>`, closing the loop between this
  skill and `adr-capture`'s post-merge sync? Leaning yes, but it's a small scope
  decision for `plan-review` or the implementer to confirm.
- Does this repo want `pr-merge` to also check for required PR approvals via the
  GitHub API, even though no branch protection currently enforces them? Doing so
  pre-emptively could future-proof the skill if protection is added later, but
  adds a check that's a no-op today.

## Verification

```bash
# Phase 1
.context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh .context/plans/pr-merge-skill-2026-07-06.md

# Phase 2
tessl install
./dist/skill-auditor evaluate .context/plugins/pantheon-org/workshop/pr-merge --store
./dist/skill-auditor validate artifacts .context/plugins/pantheon-org/workshop/pr-merge

# Phase 3
./dist/skill-auditor eval .context/plugins/pantheon-org/workshop/pr-merge

# Phase 4
gh pr checks <pr-number>
```
