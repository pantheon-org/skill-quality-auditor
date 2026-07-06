---
title: "ADR-055: Dependency updates held for a 7-day cooldown unless security-critical"
status: accepted
date: 2026-07-06
context:
  - path: .context/known-issues/dependabot-security-updates-disabled-2026-07-06.md
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

Dependabot (added in the node20 sweep) opens version-update PRs as soon as a new release
exists. A supply-chain attack — a compromised or malicious release of a GitHub Action
(the tj-actions / CVE-2025-30066 class) — is most dangerous in the hours-to-days window
before it is detected and yanked. Adopting a release immediately maximises exposure to that
window.

The repo already applies this principle to mise-managed tools via
`mise.toml`'s `minimum_release_age = "7d"`. GitHub Actions had no equivalent delay.

Investigation (2026-07-06) established:

- Dependabot supports a `cooldown` option; `default-days` is supported for the
  `github-actions` ecosystem.
- `cooldown` is an option that also affects security-update PRs **by default** — but
  advisory-driven security updates **bypass** cooldown following the fix for
  dependabot-core #13979 (filed against the `github-actions` ecosystem, fixed and deployed
  2026-02-12). This makes "delay unless critical" achievable.
- **Prerequisite:** the bypass only exists if Dependabot security updates are enabled. On
  this repo, Dependabot alerts and security updates are currently OFF — tracked as
  known-issue `dependabot-security-updates-disabled-2026-07-06` and routed to a repo admin.

## Decision

1. Add `cooldown: { default-days: 7 }` to the `github-actions` ecosystem in
   `.github/dependabot.yml`. Version-update PRs are held for 7 days after a release.
2. Rely on Dependabot **security updates** to bypass cooldown for advisory-flagged
   (critical) fixes, giving "7 days unless critical".
3. Enabling Dependabot alerts + security updates is a required prerequisite for (2); it is
   an admin repo-settings change, tracked in the linked known-issue, not applied in this
   change.

## Consequences

- **Easier:** a compromised action release sits unadopted through its typical
  detection/yank window; the repo's supply-chain exposure to malicious releases drops
  sharply, consistent with the mise `minimum_release_age` posture.
- **Harder / risk:** until the prerequisite is met, cooldown applies to *all* Dependabot
  PRs with no fast path — a critical fix would be delayed the full 7 days, worse than no
  cooldown. This is the tracked known-issue; the cooldown's "unless critical" guarantee is
  not in force until security updates are enabled.
- Routine action updates arrive ~7 days later than upstream release; acceptable, and
  security-critical ones are exempt once the prerequisite is satisfied.
