---
title: "ADR-008: Distribute skill-auditor via Homebrew tap"
status: accepted
date: 2026-06-30
context:
  - path: .context/findings/homebrew-tap-prerequisites.md
  - path: .context/plans/onboarding-improvements-2026-04-27.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

As part of multi-method distribution, Homebrew is a first-class install path for macOS users. GoReleaser v2 supports publishing Homebrew formulas to a tap repository on every tagged release. The `.goreleaser.yaml` file includes a `brews:` stanza configured to push to `github.com/pantheon-org/homebrew-tap`.

## Decision

Distribute `skill-auditor` via a Homebrew tap at `github.com/pantheon-org/homebrew-tap`. GoReleaser auto-generates and pushes `Formula/skill-auditor.rb` to the tap repository on every tagged release. Prerequisites:
- Create the `pantheon-org/homebrew-tap` public repository
- Add a `HOMEBREW_TAP_GITHUB_TOKEN` fine-grained PAT secret with write access to the tap repo

## Consequences

- macOS users install via `brew tap pantheon-org/tap && brew install skill-auditor`
- Formula is auto-maintained by GoReleaser — no manual version bumps
- Requires a GitHub PAT with cross-repo write access as a CI secret
- Only works on tagged releases (not on every commit)
