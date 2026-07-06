---
title: "Finding: plumbing getplumber.io into skill-quality-auditor's CI"
type: FINDING
status: ACTIVE
date: 2026-07-04
value: MEDIUM
related:
  - ../../.plumber.yaml
  - ../../.github/workflows/ci.yml
  - ../../.github/workflows/skill-quality.yml
  - ../../docs/ADR/adr-034-agent-scan-skill-security-scanning.md
---
# Finding: plumbing getplumber.io into skill-quality-auditor's CI

> Plumber is an open-source CLI and GitHub Action that audits `.github/workflows/*.yml` for CI/CD supply-chain issues (unpinned actions, missing `permissions:` blocks, dangerous triggers, unverified script execution) and scores the result A to E. This repo already has an untracked `.plumber.yaml` in the working tree — the tool's own default config, generated locally, not yet wired into any workflow. Adding the official Action is a small, well-precedented change, but a first real run would immediately fail on several findings because none of this repo's third-party actions are pinned by commit SHA and two of the four workflows have no `permissions:` block.

## Summary

Confirmed via `getplumber.io`, `getplumber.io/docs`, and `github.com/getplumber/plumber` (fetched 2026-07-04). Plumber is a CI/CD compliance tool, not a code or dependency scanner — it parses committed workflow YAML (GitHub Actions and GitLab CI) and repo settings (branch protection, via API) and checks them against a rule set covering OWASP CICD-SEC categories: unpinned/untrusted actions, missing least-privilege permissions, `pull_request_target` with head checkout, secret leakage in CI config, Docker-in-Docker, and more. It maps to compliance frameworks (ISO 27001, SOC 2, NIS2, DORA) rather than to skill or application-code security.

The `.plumber.yaml` already sitting in this repo's working tree (`git status` shows it as untracked) is the CLI's own default template — it matches the upstream `plumber config generate` output almost exactly, both in the GitLab and GitHub control sections. The user's global mise config (`~/.config/mise/config.toml`) has `"github:getplumber/plumber" = "latest"` installed, which is consistent with someone running `plumber analyze` or `plumber config generate` locally in this checkout. Nothing in this repo currently invokes Plumber in CI; no `.github/workflows/plumber.yml` exists and there is no reference to "plumber" anywhere else in the tracked tree.

## Detail

**Adoption paths**, in order of fit for this repo:

| Path | Fit |
| --- | --- |
| Official GitHub Action (`getplumber/plumber@<version>`) | Best fit — this repo is GitHub-only, and the Action needs no separate install step |
| Local CLI (`brew install plumber` / `mise use -g github:getplumber/plumber`) | Already present on at least one contributor's machine; useful for pre-push checks but not a CI gate on its own |
| Plumber Platform (hosted, org-wide) | Out of scope — connects at the GitHub-org level, a decision for whoever owns the `pantheon-org` GitHub org, not this repo |

**Minimal Action wiring**, following the upstream README:

```yaml
name: Plumber

on:
  pull_request:
  push:
    branches: [main]

permissions:
  contents: read
  security-events: write

jobs:
  plumber:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      - uses: getplumber/plumber@<version>
        with:
          score-push: false
```

`score-push: false` matters here: enabling it publishes an official A-E badge plus the repo name to `score.getplumber.io`, publicly. That is an org-visibility decision, not a CI-config decision, and should not be flipped on without someone outside this task confirming PLG is comfortable with a public score for this repo.

**What a first run would actually flag**, checked against the four existing workflows (`ci.yml`, `docs.yml`, `release.yml`, `skill-quality.yml`) and the default `.plumber.yaml` already in the tree:

- `actionsMustBePinnedByCommitSha` — every third-party action in this repo is pinned by tag, not SHA: `golangci/golangci-lint-action@v7`, `jdx/mise-action@v2`, `googleapis/release-please-action@v4`, `goreleaser/goreleaser-action@v6`, `tesslio/setup-tessl@v2`. The default config only exempts `actions/*` and `github/*`.
- `githubActionMustComeFromAuthorizedSources` — the same five actions are not in `.plumber.yaml`'s `trustedGithubActions` allowlist (which currently only lists `docker/*`, `ossf/scorecard-action`, `anchore/scan-action`) and are not same-org, so they would need explicit entries.
- `workflowsMustDeclarePermissions` — `ci.yml` and `skill-quality.yml` have no `permissions:` block (both fall back to the default `GITHUB_TOKEN` scope). `docs.yml` and `release.yml` already declare `permissions:` and would pass.
- `branchMustBeProtected` — needs a token with `repo` (classic PAT) or "Administration: read" (fine-grained PAT) scope; the default `GITHUB_TOKEN` in a workflow does not carry that, so this control would most likely abstain rather than fail, unless a PAT is provisioned.

None of this is a reason not to adopt Plumber — it is exactly the "unwatched pipeline" gap the tool is built to surface, and this repo has never had a CI/CD-specific security scan (only `golangci-lint`, `go vet`, `gofmt`, `markdownlint`, `shellcheck`, and the skill-quality gates). But a first run wired in as a blocking gate would fail immediately on pre-existing debt unrelated to any single PR.

**Relationship to existing security work**: this is a different layer from `docs/ADR/adr-034-agent-scan-skill-security-scanning.md` (proposed 2026-07-04). Agent-scan inspects skill *content* (`cmd/assets/`) for prompt injection and secrets; Plumber inspects *pipeline configuration* (`.github/workflows/`) for supply-chain and least-privilege issues. Adopting both is complementary, not duplicative.

## Recommended Action

Feasible, low-effort to wire up, but should land in two steps rather than one, mirroring the proving-period pattern already used for the Tessl review step and proposed for agent-scan:

1. Add `.github/workflows/plumber.yml` with `score-push: false` and no `--fail-below`/blocking exit-code enforcement (advisory only, `continue-on-error: true`), so findings surface in PRs without breaking merges on day one.
2. Separately, fix the pre-existing gaps the first run will surface: pin the five third-party actions by commit SHA, add `permissions: { contents: read }` to `ci.yml` and `skill-quality.yml`, and extend `.plumber.yaml`'s `trustedGithubActions` list for the actions this repo actually uses.
3. Once green, promote the job from advisory to a required check (drop `continue-on-error`, decide a `--threshold`).
4. Commit `.plumber.yaml` itself once its comments are trimmed to what this repo needs — the generated default is close to 1,000 lines of mostly-commented-out GitLab controls that do not apply here (this repo has no `.gitlab-ci.yml`).

Two items are a human decision, not an agent one: whether to enable `score-push` (makes the repo name and score public — a call for whoever owns visibility/compliance posture for `pantheon-org`), and whether Plumber's CI/CD-compliance framing (ISO 27001 / SOC 2 readiness) is something this repo's stakeholders want to formally track. This finding does not create an ADR — no binding decision has been made yet, only an investigation. If the recommendation above is accepted, capture the decision as an ADR at that point.
