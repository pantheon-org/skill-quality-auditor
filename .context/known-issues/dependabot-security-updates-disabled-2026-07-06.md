---
title: "Known Issue: Dependabot security updates disabled — cooldown can delay critical fixes"
type: KNOWN_ISSUE
status: DONE
date: 2026-07-06
severity: HIGH
value: HIGH
themes:
  - GOVERNANCE
related:
  - ../../docs/ADR/adr-055-dependency-update-cooldown.md
  - ../../.github/dependabot.yml
---
# Known Issue: Dependabot security updates disabled — cooldown can delay critical fixes

> The 7-day dependency cooldown (ADR-055) relies on advisory-driven security updates
> bypassing it, but Dependabot alerts and security updates are both OFF on this repo, so
> a critical CVE currently gets no fast PR and is subject to the full 7-day delay.

## Resolution (2026-07-06)

Both settings were enabled via the GitHub API:

- `PUT /repos/pantheon-org/skill-quality-auditor/vulnerability-alerts` → `204`
- `PUT /repos/pantheon-org/skill-quality-auditor/automated-security-fixes` → `204`
  (verified: `automated-security-fixes` now reports `{"enabled":true,"paused":false}`)

Advisory-driven security updates now bypass the cooldown, so ADR-055's policy behaves as
intended: **7-day cooldown on version updates, unless security-critical.** Closed.

## Why this exists

ADR-055 adds `cooldown: { default-days: 7 }` to `.github/dependabot.yml` to defend against
supply-chain attacks (a compromised release sits unadopted through its detection window).
The "unless critical" carve-out depends on Dependabot **security updates** — which are
advisory-driven and bypass cooldown (dependabot-core #13979, fixed and deployed
2026-02-12).

But as of this date both prerequisites are OFF:

- `GET /repos/pantheon-org/skill-quality-auditor/vulnerability-alerts` → `404` (alerts disabled)
- `GET /repos/pantheon-org/skill-quality-auditor/automated-security-fixes` → `{"enabled":false}`

Enabling them is a repo Settings change (Advanced Security) requiring admin rights, so it
was routed to a repo admin rather than applied in the PR.

## Impact if unfixed

With security updates disabled, cooldown applies to **all** Dependabot PRs and there is no
separate fast path. A critical vulnerability in a GitHub Action would therefore be created
*slower* (weekly version PR, delayed a further 7 days) than before cooldown existed — the
opposite of the intended "unless critical" behaviour.

## Suggested fix (not yet applied — this is the tracked issue, not the fix)

A repo admin enables both, under **Settings → Advanced Security** (or via API with admin
scope):

```
PUT /repos/pantheon-org/skill-quality-auditor/vulnerability-alerts       # Dependabot alerts
PUT /repos/pantheon-org/skill-quality-auditor/automated-security-fixes   # Dependabot security updates
```

Once enabled, advisory-driven security PRs bypass the 7-day cooldown and the policy behaves
as "7 days unless critical". Until then, treat the cooldown as delaying *all* action
updates uniformly.
