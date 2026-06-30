---
title: "ADR-020: Multi-method distribution strategy"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/onboarding-improvements.md
  - path: .context/findings/homebrew-tap-prerequisites.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The `skill-auditor` binary had no install path beyond `go install` and manual GitHub release downloads. Users needed to compile from source or navigate release assets. GoReleaser already produced cross-platform tarballs on every tagged release, but no install automation consumed them.

## Decision

Adopt a multi-method distribution strategy with four install paths:

1. **install.sh** — curl-pipe POSIX shell script from GitHub Releases (`scripts/install.sh`). Detects OS/arch, resolves latest release tag via GitHub API, downloads matching tarball, verifies checksums, installs to `/usr/local/bin`.
2. **Homebrew tap** — `pantheon-org/homebrew-tap` with GoReleaser auto-pushing `Formula/skill-auditor.rb` on every release (ADR-008).
3. **mise plugin** — Use `ubi` backend (Option A) initially, which auto-discovers GitHub releases. No custom plugin needed.
4. **`update` command** — `skill-auditor update --check` that checks for newer releases and performs a temp-file binary swap.

## Consequences

- Four install paths covering POSIX, Homebrew users, mise users, and existing binary users
- Install discovery happens via `README.md` and `skill-auditor update --check`
- Homebrew tap requires separate repo creation and a PAT secret
- mise ubi backend works out of the box for GitHub releases without custom plugin maintenance
- The `update` command is the upgrade path for users who installed via install.sh
