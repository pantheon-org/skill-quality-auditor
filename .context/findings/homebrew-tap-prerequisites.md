---
title: "Finding: Homebrew Tap Prerequisites"
type: FINDING
status: ACTIVE
date: 2026-04-28
value: LOW
related:
  - ../plans/onboarding-improvements-2026-04-27.md
---
# Finding: Homebrew tap prerequisites

**Date:** 2026-04-28
**Context:** Phase 2 of onboarding improvements — Homebrew distribution via GoReleaser

## What was done

`.goreleaser.yaml` now includes a `brews:` stanza that pushes a generated Formula to
`github.com/pantheon-org/homebrew-tap` on every tagged release.

## Blocking prerequisites

Two manual steps are required before the first release can publish the formula:

### 1. Create the `homebrew-tap` repository

Create a public repository at `github.com/pantheon-org/homebrew-tap`.
It needs no initial content — GoReleaser creates `Formula/skill-auditor.rb` on first run.

### 2. Add a `HOMEBREW_TAP_GITHUB_TOKEN` secret

Create a GitHub fine-grained PAT (or classic PAT with `repo` scope) that has **write access
to `pantheon-org/homebrew-tap`**. Add it as a repository secret named
`HOMEBREW_TAP_GITHUB_TOKEN` in `pantheon-org/skill-quality-auditor`.

GoReleaser reads this token from the environment via:

```yaml
repository:
  token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
```

## Validation

After the next tagged release, verify the formula landed:

```bash
brew tap pantheon-org/tap
brew install skill-auditor
skill-auditor version
```
