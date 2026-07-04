---
title: "ADR-034: Advisory snyk/agent-scan integration for cmd/assets/ skill security"
status: proposed
date: 2026-07-04
context:
  - path: ".context/findings/agent-scan-integration-2026-07-04.md"
  - path: ".context/plans/agent-scan-integration-2026-07-04.md"
---

**Status:** Proposed
**Date:** 2026-07-04

## Context

This repo's product is a `SKILL.md` (`cmd/assets/SKILL.md`) distributed via the Tessl registry — it is itself an agent skill, which puts it in scope for the exact class of supply-chain risk (prompt injection, malware payloads, hardcoded secrets in skill text) that `snyk/agent-scan` was built to catch. `.context/findings/agent-scan-integration-2026-07-04.md` confirmed the tool is real, actively maintained, and supports scanning a single `SKILL.md` or directory directly — but its README explicitly disclaims CLI output stability ("experimental and subject to change... we do not recommend building production workflows that depend on specific CLI output fields or issue codes") and it sends skill content to Snyk's API for analysis. `.context/plans/agent-scan-integration-2026-07-04.md` lays out a phased, advisory-first rollout. This ADR captures the binding decisions from that plan.

## Decision

1. **Integrate as an advisory-only step, not a merge-blocking gate, for at least an initial observation window.** `continue-on-error: true`, results posted as a PR comment. This mirrors the existing Tessl review step's proving-period pattern already in `skill-quality.yml`, and directly follows agent-scan's own disclaimer about output-format stability.
2. **Add the step to the existing `quality-gate` job in `.github/workflows/skill-quality.yml`**, not a new workflow file — it shares the same `cmd/assets/**` trigger path and job context as the other skill-asset checks (duplication, batch gate, structural eval).
3. **Scan target is `cmd/assets/` only — never an MCP configuration.** This repo ships no MCP config as a product artifact, so the integration never needs agent-scan's `--dangerously-run-mcp-servers` flag or its interactive stdio-server consent flow; that entire risk surface (agent-scan executes commands defined in MCP configs to introspect them) is out of scope by construction.
4. **Requires a new `SNYK_TOKEN` repository secret, provisioned by a human, not by an agent, before Phase 1 can merge.** Same treatment as `ANTHROPIC_API_KEY`/`TESSL_TOKEN` — credential provisioning is not something an implementation PR does on its own.
5. **Explicit human sign-off on data egress is a Phase 0 blocker.** Skill text (tool names, descriptions, prompt content) is sent to Snyk's Agent Scan API for analysis. This ADR does not itself constitute that sign-off — it records that sign-off is required before Phase 1 ships, per the plan's Phase 0.
6. **Gating is revisited after a fixed observation window** (proposed: 2 weeks of PRs touching `cmd/assets/`, matching the language already used for the Tessl proving period), based on observed false-positive rate, real findings caught, and CLI/schema stability — not left advisory indefinitely by default, and not promoted to a gate without that observation period.
7. **Uses `snyk-agent-scan@latest` for the advisory period**, accepting the risk that upstream behavior can shift without warning; pinning a specific version is deferred to the Phase 3 gating decision, since observing upstream's release cadence during the advisory window is itself useful input to that decision.

## Consequences

- **Easier:** skill-supply-chain risks specific to this repo's own product artifact (prompt injection, malware payloads, hardcoded secrets in `cmd/assets/SKILL.md`) get automated, PR-visible coverage that didn't exist before, using a tool purpose-built for exactly this artifact type.
- **Easier:** because the step never blocks merge during the advisory period, there's no risk of an unstable third-party CLI output format stalling the team's ability to ship — consistent with how the Tessl review step's proving period was already handled.
- **Harder:** a new external dependency (Snyk account, `SNYK_TOKEN`) and a new data-egress relationship (skill content sent to a third-party API) exist that didn't before — both require explicit, recorded human sign-off (Decision 4, 5) rather than an agent assuming consent.
- **Harder:** advisory-only means real findings can be present on `main` for the length of the observation window without blocking anything — acceptable short-term given the alternative (gating on an admittedly unstable output format) is worse, but this ADR's status should move to `accepted` (or be superseded) once Phase 3's gate/no-gate decision is made, not left `proposed` indefinitely.
