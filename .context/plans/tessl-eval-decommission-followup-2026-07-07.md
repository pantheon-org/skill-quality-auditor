---
title: "Plan: Decommission the residual Tessl eval coupling (proving-period removal + Layer 3/4 cleanup)"
type: PLAN
status: DRAFT
date: 2026-07-07
value: MEDIUM
effort: S
themes:
  - EVAL
  - DISTRIBUTION
  - DOCS
related:
  - ../plans/migrate-off-tessl-eval-2026-06-29.md
  - ../plans/native-eval-runner-2026-07-01.md
  - ../findings/eval-gating-byok-2026-06-29.md
  - ../../docs/ADR/adr-001-native-eval-runner.md
  - ../../.github/workflows/skill-quality.yml
---
# Plan: decommission the residual Tessl eval coupling

Status: DRAFT for review
Date: 07-07-2026
Branch (proposed): `chore/tessl-eval-decommission`
Author: investigation by AI agent, decisions pending human owner

> Decision-support material. The final scope, CI change, and any secret
> removal must be decided and documented by a human maintainer before
> implementation.

## 1. Why this exists

This is the narrow follow-up to
[`migrate-off-tessl-eval-2026-06-29.md`](migrate-off-tessl-eval-2026-06-29.md),
which is now SUPERSEDED. The foundational work in that plan (the native Go eval
runner, Option A) already shipped via
[`native-eval-runner-2026-07-01.md`](native-eval-runner-2026-07-01.md) and
ADR-001. CI already runs `skill-auditor eval` on the relevant PR paths.

What remains is genuinely small and falls into two buckets: a date-gated CI
cleanup, and a documentation/packaging tidy. This plan tracks only that
residue so the index no longer carries a stale HIGH-value item that reads as
if the migration has not started.

## 2. Bucket A — proving-period removal (date-gated, ~15-07-2026)

`.github/workflows/skill-quality.yml` still carries an advisory Tessl step kept
deliberately alongside the native runner during a proving period
(`skill-quality.yml:252-263`):

```yaml
# ── Proving period: Tessl review kept alongside the native runner ──
# Remove this block + the TESSL_TOKEN secret after 2 weeks of green [runs]
- uses: tesslio/setup-tessl@25ec223fc0da33b41b8044ff5ab2b85235f4f91e # v2
  with:
    token: ${{ secrets.TESSL_TOKEN }}
- name: Tessl review run
  ...
  run: tessl review run cmd/assets/ --workspace pantheon-ai --json --threshold 80
```

The native runner shipped on 01-07-2026, so the two-week green window closes
around **15-07-2026**. Do not remove earlier: the whole point of the proving
period is to confirm the native runner and the Tessl review agree before the
advisory signal is dropped.

**Precondition to check on the day:** confirm CI has stayed green on the native
`skill-auditor eval` steps (`skill-quality.yml:54, 98, 154`) for the full
window with no native-vs-Tessl divergence that was resolved in Tessl's favour.

**Action when ripe:**

1. Delete the `setup-tessl` + `Tessl review run` block (`skill-quality.yml:252-263`)
   and its explanatory comment.
2. Remove the `TESSL_TOKEN` repository secret (owner action; note it here once done).
3. Confirm no other workflow references `TESSL_TOKEN`.

**Explicitly out of this bucket:** the `.tessl` mirror-drift job
(`skill-quality.yml:102-127`, `npm install -g tessl` + `tessl install` +
`check-tessl-mirror-drift.sh`). That job validates the *helper-skill
distribution* path (ADR-047/048), not eval, and stays until Bucket B decides
the wider distribution question. It does not use `TESSL_TOKEN`.

## 3. Bucket B — Layer 3/4 cleanup (not date-gated)

Carried forward from the superseded plan's Layers 3 and 4.

### B1. Owner decision — RESOLVED 07-07-2026: keep on registry for now

**Decision (owner, 07-07-2026):** leave the skill published on the Tessl
registry for now. No distribution change. Revisit as a separate call later if
the registry stops earning its keep. Consequence: Layer 4 wording must stay
honest about registry compliance while making clear that runtime evaluation is
now native (B2). `tile.json`, `tessl.json`, `.tessl/plugins/...`, the mirror-drift
job, and the MCP config files all stay untouched.

The original decision framing is retained below for context.

Does the skill stay published on the Tessl registry?

- **Keep on registry:** retain `tile.json`, `tessl.json`, `.tessl/plugins/...`,
  the `tessl install` consumption path, and the mirror-drift job. Then Layer 4
  wording must stay honest about registry compliance while making clear that
  *runtime evaluation is now native*.
- **Leave the registry:** schedule removal of the above as a separate
  distribution change. This is a product/distribution call, not an eval call,
  so it is deliberately parked here as a decision rather than pre-actioned.

Recommendation: **keep on registry for now** and only reconcile wording. The
registry is a distribution channel; there is no eval reason to drop it, and the
mirror-drift job already guards its integrity.

### B2. Wording reconciliation (Layer 4, mechanical, only where inaccurate)

Reword only claims that are now factually wrong (runtime eval is native, not
Tessl-hosted). Leave accurate registry/packaging references intact. Candidate
files (verify each still misstates the position before editing):

| File | Nature of edit |
| ---- | -------------- |
| `docs/reference/d9-eval-validation.md` | Replace any `tessl eval run` / `view-status` guidance with `skill-auditor eval`; keep the scenario-format description |
| `cmd/assets/references/tessl-compliance-framework.md` | Keep as registry-submission guidance; add a line that runtime eval is native |
| `cmd/assets/SKILL.md` | Only reword "tessl registry compliance" if B1 chooses to leave the registry |
| `README.md` | Confirm "Tessl tile" wording matches the B1 decision |
| `docs/d4-specification-compliance.md` | Reconcile any `tessl install` eval framing |
| `CONTRIBUTING.md` | Reconcile any `tessl eval` contributor instruction |

Note: `CHANGELOG.md` Tessl references are history and are deliberately left.
MCP config files (`.mcp.json`, `.mcpx.json`, `.vscode/mcp.json`, `opencode.json`)
are Layer 3 tooling and only change if B1 chooses to leave the registry.

## 4. Sequencing

- Bucket A is independent and date-gated: earliest sensible date ~15-07-2026.
- Bucket B1 is a decision that can be taken any time; B2 follows B1 and is
  mechanical. B2 should reference the same PR as A if timing lines up, to keep
  a single "final Tessl-eval cleanup" changeset.

## 5. Verification (once actioned)

- `grep -rn 'TESSL_TOKEN' .github/` returns nothing.
- `go test ./...` and `hk check` pass.
- `./dist/skill-auditor eval ./cmd/assets --fail-below 0` still runs the
  scenarios (native path unaffected).
- `grep -ril tessl` shows only intentional, documented references (registry
  packaging if B1 keeps it, mirror-drift job, CHANGELOG history).
- CI green with the advisory Tessl review step removed.

## 6. Open decisions for the owner

1. ~~**B1:** stay on the Tessl registry (recommended) or leave it?~~
   RESOLVED 07-07-2026: keep on registry for now (see B1).
2. Confirm the proving-period window (~15-07-2026) and that CI stayed green
   with no native-vs-Tessl divergence before removing the advisory step.
3. Who removes the `TESSL_TOKEN` secret (repo admin action, not code).
