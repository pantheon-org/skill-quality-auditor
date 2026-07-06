---
title: "Findings: Eval Gating, Local Runs, and Bring-Your-Own-Key"
type: FINDING
status: ACTIVE
date: 2026-06-29
value: MEDIUM
themes:
  - EVAL
related:
  - ../plans/migrate-off-tessl-eval-2026-06-29.md
---
# Findings: eval gating, local runs, and bring-your-own-key

Date: 29-06-2026
Status: DECISION-SUPPORT, not actioned
Related plan: [migrate-off-tessl-eval-2026-06-29.md](../plans/migrate-off-tessl-eval-2026-06-29.md)
Author: investigation by AI agent, decisions pending human owner

> This is decision-support material produced while reviewing the Tessl
> migration plan. It records pushback on three questions: how to gate skill
> validation in CI, how to do it locally, and whether a consumer can bring
> their own key. The final approach must be decided and documented by a human
> maintainer before implementation.

## Framing problem underneath all three questions

The migration plan conflates two signals that should not share a gate:

1. **Structural D9** (`scorer/d9_eval_validation.go`): pure Go, deterministic,
   no network, no key. Already runs today in `go test` and `evaluate`.
2. **Runtime eval** (the proposed LLM-judge runner): non-deterministic, needs
   network, needs a key, costs money.

Option A as written replaces Tessl like-for-like, inheriting Tessl's
"one blocking eval" shape and putting the non-deterministic signal behind a
hard binary gate (`--fail-below 80`). Separating the two signals resolves all
three questions below.

## 1. CI gating

Three issues the plan does not confront:

- **Fork PRs cannot see the secret.** GitHub does not expose repository
  secrets to workflows triggered by `pull_request` from a fork. A required
  eval gate using `ANTHROPIC_API_KEY` therefore cannot run for external
  contributors. `pull_request_target` would expose secrets to untrusted
  checked-out code (token-exfiltration vector, do not use). This alone is
  decisive against "LLM-judge required on every PR".
- **Flaky knife-edge threshold.** A single LLM-judge sample at `--fail-below 80`
  is non-deterministic even at temperature 0. Re-runs flip red to green, which
  trains reviewers to ignore the gate.

Recommendation: **two-tier gate.**

- **Required, every PR, everywhere (incl. forks):** structural D9 via
  `go test ./scorer/...` and `skill-auditor evaluate cmd/assets --fail-below B`.
  Deterministic, free, no key. This is the real merge gate.
- **Advisory, non-blocking:** the LLM-judge runner posts a report or PR comment
  with per-scenario scores and a trend delta. Runs on same-repo pushes, on a
  `run-eval` label, or nightly against `main`. If it must block anything, it
  blocks the nightly/main build, not PRs, and only with N-sample median plus a
  margin band.

## 2. Local runs

- **Reproducibility trap.** If CI gates on LLM-judge and local can only do
  structure-only, contributors get "passes locally, fails in CI" with no way to
  reproduce. The two-tier model fixes this: the gating signal (structural) is
  identical locally and in CI, key or no key.
- **hk / pre-push.** The repo just migrated to hk. The runtime eval must NOT go
  in pre-push: too slow, needs network and a key, costs money per push.
  Pre-push runs the structural scorer only. This belongs in the plan now that
  hk is the hook runner.
- **One command, two tiers.** `skill-auditor eval` should detect key presence
  and degrade gracefully: no key -> structural only, said loudly; key present
  -> full actor + judge. Same command and flags locally and in CI, so the only
  variable is whether a key is available.

## 3. Bring-your-own-key (sub / API / local)

Most under-served question, and the most important for a tool other people run
on their own skills. The plan frames eval as something only this repo does. The
product is a CLI consumers point at their skills, so BYOK is a design
constraint, not a nice-to-have.

Problems with the plan as written:

- **Pinned-in-code judge model conflicts with BYOK.** Section 5 pins the model
  and prompt for reproducibility and hardcodes `ANTHROPIC_API_KEY`. A consumer
  cannot override either cleanly. Resolution: pin a default but make base URL,
  model id, and key source configurable, and record what was actually used in
  the JSON output. Reproducibility becomes "recorded and auditable", not
  "enforced and unchangeable". Those are different goals the plan treats as one.
- **Subscription vs API key is a real split.** Many likely consumers have Claude
  Code on a Max subscription with no separate API billing. `ANTHROPIC_API_KEY`
  is one auth path, not the only one. This is the genuine point in favour of
  Option C (Agent SDK / Claude Code), which can run off the existing OAuth
  session and sidesteps key provisioning entirely. The plan rejected C on
  "second runtime" grounds and missed this trade-off.
- **Provider-agnostic and regulated environments.** Skill content can be
  confidential; sending it to a hosted API is a data-governance decision the
  consumer must own. The eval boundary should be an interface, not a hardcoded
  Anthropic client. At minimum support an OpenAI-compatible base URL so a local
  model (Ollama, vLLM) or internal gateway can be slotted in. The runtime eval
  must be fully optional: a consumer who will not send content to any API still
  gets the structural D9 grade.

Honest answer: as currently written, the consumer cannot BYOK. To get there:
(a) eval optional, (b) key/model/endpoint configurable not pinned, (c) an
abstraction at the model-call boundary, (d) ideally an SDK/subscription path so
people use the auth they already have.

## Suggested plan edits (for later)

- Reframe Section 4 around the two-tier model: deterministic structural = the
  gate; LLM-judge = advisory signal.
- Add a CI subpoint on fork PRs and secret-visibility limits. This turns
  Open decision 3 (cadence) from a preference into a constraint.
- Add an hk/pre-push subsection: structural only in hooks, never the model call.
- Promote BYOK to an explicit design goal: configurable model/endpoint/key,
  optional runtime eval, recorded-not-pinned reproducibility, and re-open
  Option C on the subscription-auth merit specifically.
- Prefer adding a new section ("12. Gating, local runs, and bring-your-own-key")
  over rewriting Sections 4 to 6, so the original options analysis stays intact
  for the owner to compare against.
