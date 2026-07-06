---
title: "Finding: should .context/index.yaml be split into per-type files? — reviewed, decided against for now"
type: finding
status: active
date: 2026-07-06
related:
  - ../../.agents/RULES.md
  - ../../.context/plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh
  - ../../.tessl/plugins/pantheon-org/context-mgmt/context-index/scripts/regenerate-context-index.sh
---
# Finding: should .context/index.yaml be split into per-type files? — reviewed, decided against for now

> Three real git merge conflicts on `.context/index.yaml` in one hour (merging three independent PRs, each adding one unrelated entry) prompted the question: should the index be split by type (`.context/plans/index.yaml`, `.context/findings/index.yaml`, etc.) instead of staying one file? Reviewed via three independent adversarial agents (pro-split, anti-split, migration-risk) rather than a single pass. Decision: **don't split now** — the actual pain is already solved for near-zero cost by a rule written the same session, and splitting trades that solved problem for real new risk.

## Summary

At review time: 86 entries, 548 lines, one file, auto-regenerated wholesale by `scripts/regenerate-context-index.sh` on every `hk` pre-commit run, grouped into 6 sections (Known Issues, Plans, Findings, Analysis, Audits, Instructions).

**Root cause of the conflicts (found by the pro-split review, sharper than initially assumed):** it's not entries shifting position — it's a shared mutable header. `regenerate-context-index.sh` rewrites a `# Last updated: <date>` line and an aggregate `# N entries: X active, Y done...` line (built from `status_counts` over *every* entry) on every single regeneration, regardless of which section actually changed. Two branches that each add one file *anywhere* under `.context/` will independently regenerate different header values and collide at the top of the file — before any actual content conflict.

**Why split anyway falls short (anti-split + migration-risk reviews):**

1. `.agents/RULES.md` rule 16 ("Regenerate, don't hand-merge, conflicts on auto-generated files"), added the same session as this investigation, already resolves the actual git-conflict pain for near-zero cost — proven live three times in the hour that motivated this review, each resolved in under a minute with no risk of an inconsistent result.
2. Splitting narrows the collision domain (same-type changes still collide) but doesn't eliminate it — the pro-split review's own estimate is roughly a 6x reduction, not zero, since the root-cause header-churn mechanism would need a separate fix regardless of file count.
3. **`.tessl/plugins/pantheon-org/context-mgmt/context-index/` is a separate, already-diverged vendored copy of the generator/validator scripts** (`tessl.json` mode: `vendored`), confirmed via `diff -rq` to already differ from the `.context/plugins/` originals, and confirmed via grep to be referenced by no CI workflow. A split done in `.context/plugins/` would need a manual re-sync to this mirror with no automated safety net catching a missed one — the published tile would silently keep shipping single-file behavior.
4. **Highest silent-failure risk identified:** `regenerate-context-index.sh`'s `--check` mode currently does one string comparison against one file. If a split rewrote the generator to emit N files but the check-mode comparison wasn't carefully extended to verify all N, `hk check` would print "context index is fresh" while some files are genuinely stale — the exact failure the gate exists to catch, passing green. This is a real risk specifically because it's silent, not because it's certain.
5. 36+ files outside the index itself reference `.context/index.yaml` by exact path (skill scripts, ~15 `SKILL.md` files, `AGENTS.md`, eval scenario tasks that assert against the literal path, plans/findings citing it in prose). Migration is real engineering, not a rename.
6. Splitting would also require rebuilding the "one `cat`, whole-repo-state" property `AGENTS.md` documents as the intended read pattern — either as N separate reads, or a live/on-demand re-aggregation (the pro-split review's own proposed mitigation), which is more engineering than the problem being solved.

## Follow-up

No action taken — this finding documents a considered "not now" rather than a silent drop, per the "no man left behind" rule. Revisit only if: (a) entry count grows substantially further (watch for it crossing a few hundred, well beyond today's 86), or (b) rule 16 turns out insufficient in practice over the following weeks (conflicts recur despite it, or the regenerate-and-continue workflow proves error-prone in practice). Neither condition is met today.

A smaller, independent idea surfaced but not pursued: if header churn ever becomes annoying on its own (separately from the split question), dropping the `Last updated` line or making the aggregate count line conflict-tolerant would address the specific mechanism without touching the file's structure — much cheaper than a full split, and worth considering on its own merits if this resurfaces.
