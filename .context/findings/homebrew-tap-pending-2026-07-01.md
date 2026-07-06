---
title: "Finding: Homebrew Tap Repository Required"
type: finding
status: active
date: 2026-07-01
value: medium
related:
  - ../plans/onboarding-improvements-2026-04-27.md
---
# Finding: Homebrew Tap Repository Required

> Phase 2 of the onboarding improvements plan cannot be completed without org admin access.

## Summary

Homebrew installation is configured in `.goreleaser.yaml` but the target repository `pantheon-org/homebrew-tap` does not exist, and the `HOMEBREW_TAP_GITHUB_TOKEN` secret is not set. The next release will fail to publish the formula.

## What's needed

1. Create `github.com/pantheon-org/homebrew-tap` (empty repo, no template needed).
2. Set `HOMEBREW_TAP_GITHUB_TOKEN` as a repo secret in `pantheon-org/skill-quality-auditor` with a fine-grained PAT that has `contents:write` on `homebrew-tap`.
3. The next GoReleaser tag will auto-push the formula — no manual formula creation needed.

## What's already done

- `.goreleaser.yaml` has the `brews` block with correct repo name, homepage, description, license, and install/test stanzas.
- README documents `brew tap pantheon-org/tap && brew install skill-auditor` and `brew upgrade skill-auditor`.
- CI shellcheck covers `scripts/*.sh` (Phase 1 gate).
- `scripts/install.sh` and `cmd/update.go` shipped.
