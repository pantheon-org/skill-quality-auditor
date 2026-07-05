---
title: "Draft Plan: Integrate snyk/agent-scan as an Advisory PR Check"
type: plan
status: draft
date: 2026-07-04
effort: S
related:
  - ../findings/agent-scan-integration-2026-07-04.md
  - ../../docs/ADR/adr-034-agent-scan-skill-security-scanning.md
  - ../../.github/workflows/skill-quality.yml
  - ../instructions/ways-of-working.md
---

**Effort:** S — one new advisory step in an existing job, closely mirroring the already-shipped `pr-agent.yml` pattern. Blocked on a human prerequisite (Phase 0: provisioning `SNYK_TOKEN`), which isn't itself engineering effort but does gate the start date.

## Goal

Add `snyk/agent-scan` (`uvx snyk-agent-scan@latest`) as a step in `.github/workflows/skill-quality.yml` that scans `cmd/assets/` for prompt injection, malware payloads, hardcoded secrets, and credential-handling issues on every PR that touches skill assets — surfacing findings without blocking merge, until the tool's output format has proven stable enough to gate on. See `.context/findings/agent-scan-integration-2026-07-04.md` for the feasibility assessment this plan is based on.

## Scope

**In scope:**

- One new step in the existing `quality-gate` job in `skill-quality.yml`, triggered on the same `cmd/assets/**` path filter already used by the other steps in that workflow.
- `uv` installed on the runner (`astral-sh/setup-uv` or equivalent) so `uvx snyk-agent-scan@latest` can run without a prior `pip install`.
- Scanning `cmd/assets/` only — never an MCP configuration file — so the stdio-server consent/`--dangerously-run-mcp-servers` mechanics in the upstream tool never come into play for this integration.
- JSON output (`--json`) uploaded as a workflow artifact and posted as a PR comment, mirroring the existing "Post eval results comment" step's `actions/github-script` pattern.
- `continue-on-error: true` on the scan step itself, matching the Tessl review step's proving-period pattern.

**Out of scope (deferred):**

- Making the scan a blocking gate (`--fail-below`-style behavior) — deferred until the tool's output has been observed as stable across enough real PRs; agent-scan's own README calls its CLI output "experimental and subject to change."
- Scanning MCP configurations (`.mcp.json`, agent harness configs) — this repo doesn't ship any as a product artifact; if that changes, it's a separate decision with its own consent/sandboxing considerations (agent-scan executes stdio MCP server commands to introspect them).
- Nightly/scheduled runs — the existing nightly cron in this workflow is scoped to the LLM-judge eval advisory; adding agent-scan to it can be considered once the PR-triggered version has run long enough to be trusted, not before.
- Any workflow-file edits — this plan documents intent for review; implementation is a separate PR once the plan and ADR are accepted.

## Decisions (proposed — pending review)

1. **Advisory-only via `continue-on-error: true`, not a merge-blocking gate.** Justification: agent-scan's own README disclaims CLI output stability; gating on an unstable schema risks false-fail noise the maintainers have flagged as their own known limitation.
2. **Lives as a new step inside the existing `quality-gate` job in `skill-quality.yml`**, not a separate workflow file — it shares the same trigger paths (`cmd/assets/**`, `**/*.go`) and job context as the other skill-asset checks (duplication, batch gate, structural eval), so there's no reason to duplicate the `on:` block.
3. **Scan target is `cmd/assets/` only.** No MCP config in this repo needs scanning, so the integration never needs `--dangerously-run-mcp-servers` or the interactive consent flow — sidesteps that entire risk surface.
4. **New repo secret: `SNYK_TOKEN`.** This is a manual, human action (Snyk account signup + token generation + `gh secret set` or repo settings) — not something an agent should provision. Blocks Phase 1 until done.
5. **Results posted as a PR comment via `actions/github-script`**, reusing the shape already proven by the "Post eval results comment" step in the same workflow, rather than inventing a new reporting mechanism.
6. **Revisit gating after a fixed observation window** (proposed: 2 weeks of PRs touching `cmd/assets/`, mirroring the Tessl proving-period language already in this workflow's comments) rather than leaving it advisory indefinitely or picking an arbitrary PR count.
7. **Data-sharing disclosure stays out of code and lives in the ADR/finding.** Skill content is sent to Snyk's API for analysis; this needs explicit sign-off from whoever owns this repo's security/compliance posture before Phase 1 merges, not just an agent's assumption that it's fine.

## Phases

### Phase 0 — Prerequisites (human action, blocks Phase 1)

- Sign up for Snyk, generate an API token, add it as the `SNYK_TOKEN` repository secret.
- Confirm with whoever owns this repo's security posture that sending `cmd/assets/` skill content to Snyk's Agent Scan API for analysis is acceptable.
- Exit criterion: `SNYK_TOKEN` exists in repo secrets; sign-off recorded (e.g. as a comment on the ADR PR).

### Phase 1 — Wire in the advisory step

- Add `astral-sh/setup-uv` (or manual `uv` install) to the `quality-gate` job.
- Add a step: `uvx snyk-agent-scan@latest cmd/assets --json > agent-scan-results.json`, `continue-on-error: true`, `env: SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}`.
- Add an `actions/upload-artifact` step for `agent-scan-results.json`, `if-no-files-found: ignore`, matching the existing eval-results artifact step.
- Add a PR-comment step parsing `agent-scan-results.json` and posting a summary, guarded by `if: always() && github.event_name == 'pull_request'`, matching the existing eval-comment step's structure.
- Run the new step against this repo's own `cmd/assets/SKILL.md` as a smoke test (a skill-auditor SKILL.md scanning itself) before merging, to confirm the CLI runs cleanly in CI and produces parseable JSON.
- Exit criterion: a PR touching `cmd/assets/**` shows a new PR comment with agent-scan findings (or a clean-scan confirmation); the step never blocks merge regardless of findings.

### Phase 2 — Observe

- Let the advisory step run across real PRs for the window set in Decision 6.
- Track: any false positives, any real findings caught, any CLI crashes/schema changes that would have broken a hard gate.
- Exit criterion: enough observed runs to make an informed accept/reject/adjust call on gating.

### Phase 3 — Decide on gating

- Based on Phase 2 observations, either promote to a blocking gate (define what severity threshold blocks), keep it advisory indefinitely, or drop it if signal-to-noise is poor.
- Update `docs/ADR/adr-034-agent-scan-skill-security-scanning.md`'s status/decision if the outcome changes what was originally decided.
- Exit criterion: ADR-034 reflects the final state (accepted as-is, superseded, or amended).

## Risks

- **Output schema volatility** — agent-scan explicitly disclaims stable CLI output; a schema change could silently break the PR-comment parsing step. Mitigated by `continue-on-error: true` and by treating comment-parsing failures as "skip commenting," not "fail the job" (mirrors the existing eval-comment step's try/catch).
- **Data egress** — skill text goes to a third-party API. Mitigated by requiring explicit sign-off in Phase 0 before any code merges, not assuming consent.
- **Runner time / flakiness from `uvx` cold-starts** — first-run package resolution adds latency to every PR touching `cmd/assets/**`. Acceptable for an advisory step; would need addressing (caching, pinned version instead of `@latest`) before Phase 3 gating.
- **Alert fatigue** — even advisory, a noisy PR comment on every touch to `cmd/assets/**` risks being ignored if false-positive rate is high. Phase 2's explicit observation window exists to catch this before deciding on Phase 3.
- **Pinning `@latest`** — using `snyk-agent-scan@latest` means the tool can change behavior under us with no warning, compounding the output-volatility risk above. Worth reconsidering a pinned version once Phase 1 ships (not blocking Phase 1, since the point of the advisory period is partly to observe upstream's release cadence).

## Verification

```bash
# Local smoke test before wiring into CI (requires a personal SNYK_TOKEN)
export SNYK_TOKEN=...
uvx snyk-agent-scan@latest cmd/assets/SKILL.md --json

# After Phase 1 lands, confirm the workflow step runs and comments
gh pr view <pr-number> --json comments --jq '.comments[].body' | grep -i "agent scan"
```
