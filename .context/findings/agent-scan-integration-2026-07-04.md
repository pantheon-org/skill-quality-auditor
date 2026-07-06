---
title: "Finding: snyk/agent-scan as a PR security gate for cmd/assets/"
type: FINDING
status: ACTIVE
date: 2026-07-04
value: LOW
themes:
  - PR-TOOLING
  - SKILL-QUALITY
related:
  - ../plans/agent-scan-integration-2026-07-04.md
  - ../../docs/ADR/index.yaml
  - ../../.github/workflows/skill-quality.yml
---
# Finding: snyk/agent-scan as a PR security gate for `cmd/assets/`

> `snyk/agent-scan` is a real, actively maintained Snyk CLI (`snyk-agent-scan` on PyPI) that scans agent skills — including a single `SKILL.md` or a skills directory — for prompt injection, malware payloads, hardcoded secrets, and credential-handling issues. It fits this repo's use case directly, but its README explicitly labels CLI output as experimental, and it requires a new `SNYK_TOKEN` secret plus sharing skill content with Snyk's API.

## Summary

Confirmed via `gh repo view snyk/agent-scan` and the repo's README (fetched 2026-07-04). The tool is not a GitHub Action — it's a Python CLI installed and run via `uvx snyk-agent-scan@latest`. It auto-discovers agent components (MCP servers, skills) but also accepts an explicit path, so it can be pointed at exactly `cmd/assets/` (or `cmd/assets/SKILL.md`) without touching MCP configuration at all.

## Detail

**What it checks for skills specifically** (per `docs/issue-codes.md` in the upstream repo): Prompt Injection (E004), Malware Payloads (E006), Untrusted Content (W011), Credential Handling (W007), Hardcoded Secrets (W008). Snyk published a companion threat report on the agent-skill ecosystem alongside the 0.4 release that added skill scanning — this is a live area of vendor investment, not an afterthought feature.

**Usage relevant to this repo:**

```bash
export SNYK_TOKEN=your-api-token   # required
uvx snyk-agent-scan@latest cmd/assets/SKILL.md
# or scan the whole assets tree
uvx snyk-agent-scan@latest cmd/assets
```

Because we'd only ever point it at `cmd/assets/` (never at an MCP config), the tool's stdio-server consent prompt / `--dangerously-run-mcp-servers` flag is irrelevant to this integration — that flag only matters when scanning MCP server configs, which start subprocesses. Skill scanning does not execute anything from the skill file.

**Caveats surfaced by the README, verbatim:**

1. *"CLI output is experimental and subject to change... issue codes, field names, severity labels, and response structure... may change without notice between releases. We do not recommend building production workflows that depend on specific CLI output fields or issue codes."*
2. Skill content — including tool names, descriptions, and prompt text — is sent to the Snyk Agent Scan API for analysis. Snyk states it does not store or log tool-call contents/results, but the skill's own text is transmitted for scoring.
3. Requires a new `SNYK_TOKEN` secret (Snyk account + API token), separate from any credentials already in this repo (`ANTHROPIC_API_KEY`, `TESSL_TOKEN`).
4. Requires `uv` on the runner (not currently installed by any existing workflow step — would need `astral-sh/setup-uv` or equivalent added).

**Comparison to this repo's existing quality gates** (`.github/workflows/skill-quality.yml`): the repo already has precedent for exactly this shape of problem — an external, less-mature checker run alongside the primary gate without blocking merge. The Tessl review step (lines 107–118) runs with `continue-on-error: true` during what the workflow's own comments call a "proving period," explicitly deferred until "2 weeks of green CI runs." `agent-scan`'s own experimental-output disclaimer argues for the identical pattern: run it, surface results, don't gate on it yet.

## Recommended Action

Feasible and worth adding, advisory-only, following the same proving-period pattern already established for the Tessl review step. See the accompanying draft plan (`.context/plans/agent-scan-integration-2026-07-04.md`) for the phased rollout and `docs/ADR/adr-034-agent-scan-skill-security-scanning.md` for the binding decisions once reviewed. Two items need a human (not an agent) before any implementation: provisioning `SNYK_TOKEN` as a repo secret, and confirming the org is comfortable with skill content being sent to Snyk's API for analysis.
