---
title: "Finding: PR-Agent as an automated PR review bot"
type: finding
status: active
date: 2026-07-04
related:
  - ../../.github/workflows/skill-quality.yml
  - ../../.plumber.yaml
  - ../../docs/ADR/index.yaml
---
# Finding: PR-Agent as an automated PR review bot

> `The-PR-Agent/pr-agent` is the actively maintained, community-owned continuation of Codium/Qodo's original open-source PR-Agent (Qodo donated it back to the community; Qodo's own product moved on to a separate paid "Qodo 2.0" platform). It's a mature, Apache-2.0 tool that adds LLM-based `/review`, `/describe`, `/improve`, and `/ask` comments to pull requests. It would sit alongside — not replace — this repo's existing skill-quality LLM-judge and would need explicit scoping and a Plumber allowlist entry before it could run.

## Summary

Confirmed via GitHub (`The-PR-Agent/pr-agent` README and `docs/docs/installation/github.md`, fetched 2026-07-04). The project is the direct successor to `Codium-ai/pr-agent`: Qodo donated the original project to the community, it now lives under the `The-PR-Agent` GitHub org, is Apache-2.0 licensed, and carries the original project's history (1.6k forks). It is not the same thing as "Qodo Merge" / "Qodo 2.0", which is Qodo's own separate, hosted, freemium successor product — the README is explicit that this repo is "not the Qodo free tier."

## Detail

**What it does:** four LLM-backed tools that comment on a pull request:

- `/describe` — auto-generates a PR title/summary from the diff
- `/review` — a structured code-review comment (risk areas, effort estimate, checklist)
- `/improve` — inline code-suggestion comments
- `/ask` — answers free-form questions about the PR, including "ask on code lines"

Each tool call is a single LLM call (~30s), and PR-Agent has its own diff-compression strategy to fit large PRs into a model's context window without truncation.

**Deployment options:** CLI (`pip install pr-agent`), GitHub Action (Docker-based), GitHub App/webhook, or self-hosted. Also supports GitLab, Bitbucket, Azure DevOps, and Gitea — irrelevant here since this repo is GitHub-only.

**GitHub Action shape:**

```yaml
on:
  pull_request:
    types: [opened, reopened, ready_for_review]
  issue_comment:
jobs:
  pr_agent_job:
    if: ${{ github.event.sender.type != 'Bot' }}
    permissions:
      issues: write
      pull-requests: write
      contents: write
    steps:
      - uses: the-pr-agent/pr-agent@main   # or docker://pragent/pr-agent@sha256:<digest>
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          config.model: "anthropic/claude-3-opus-20240229"
          ANTHROPIC.KEY: ${{ secrets.ANTHROPIC_API_KEY }}
          github_action_config.auto_review: "true"
          github_action_config.auto_describe: "true"
          github_action_config.auto_improve: "true"
```

Configuration is via dotted/double-underscore env vars or a `.pr_agent.toml` file at repo root; either can pin the model and pass `extra_instructions` per tool.

**LLM providers:** OpenAI (default), Gemini, Anthropic/Claude, and Amazon Bedrock (with IAM-role credentials on AWS-hosted runners — no static keys needed there). This repo already provisions `ANTHROPIC_API_KEY` for the nightly LLM-judge in `skill-quality.yml`, so PR-Agent could reuse that secret via `config.model: "anthropic/claude-*"` + `ANTHROPIC.KEY: ${{ secrets.ANTHROPIC_API_KEY }}` rather than adding a new one.

**Overlap with existing tooling:** `skill-quality.yml` already runs an LLM-judge, but it's scoped narrowly — it scores `cmd/assets/` skill content against the 9-dimension framework, advisory-only, gated to schedule/dispatch/`run-eval`-label. PR-Agent is a general code-review bot that would comment on every PR touching any file (Go source, workflows, docs). These are different concerns and wouldn't literally duplicate each other, but running both against `cmd/assets/` changes would produce two separate LLM opinions on the same skill-content diff. If adopted, scoping PR-Agent's trigger to exclude `cmd/assets/**` (or configuring `ignore_pr_target_branches`/path filters) would avoid noisy, redundant commentary on skill files that already get the specialized eval.

**Supply-chain / Plumber implications** (this repo runs `.plumber.yaml` with `githubActionMustComeFromAuthorizedSources` and `actionsMustBePinnedByCommitSha` enabled — see ADR-036 through ADR-039 for the precedent):

1. `the-pr-agent/pr-agent` is a third-party, non-same-org action — it would need adding to `.plumber.yaml`'s `trustedGithubActions` list, the same way `getplumber/plumber` was added.
2. The README's own quick-start uses `uses: the-pr-agent/pr-agent@main` — a floating branch ref that Plumber's `actionsMustBePinnedByCommitSha` control would flag. The docs do offer a pinned alternative: `uses: docker://pragent/pr-agent@sha256:<digest>`, with published GitHub Artifact Attestations (`gh attestation verify`) to confirm the digest matches this repo's build. That pinned form is the only one that would pass this repo's existing Plumber gate.
3. The example workflow requests `contents: write` in addition to `issues: write` / `pull-requests: write`. `contents: write` is only needed for the `/update_changelog` tool; if that tool isn't enabled, the permissions block should drop to `issues: write` + `pull-requests: write` to match this repo's least-privilege posture (`workflowsMustDeclarePermissions` / no `write-all`).

**Data-handling note:** whichever LLM provider is configured, PR-Agent sends the (compressed) PR diff to that provider's API for every tool call. This repo's content is source code and documentation, not participant or personal data, so this doesn't raise the GDPR concerns that would apply to participant-facing content — but it's the same category of consideration already raised for `agent-scan` (`.context/findings/agent-scan-integration-2026-07-04.md`): a new external service sees repo diffs on every PR.

## Recommended Action

Worth a follow-up plan, advisory-only initially, scoped to non-`cmd/assets/` changes to avoid duplicating the existing skill-content LLM-judge. Two things need a human decision before implementation: (1) whether to reuse `ANTHROPIC_API_KEY` (cheaper, one fewer secret) or provision a separate key/budget for PR-Agent's own usage, and (2) whether `/improve` (inline code-suggestion comments) is wanted given this is a small, actively-maintained Go codebase where noisy suggestion comments could add review friction rather than reduce it. If the answer to (2) is "not yet," start with `/describe` + `/review` only and revisit `/improve` after a trial period — mirroring the proving-period pattern already used for the Tessl review step and `agent-scan`.
