#!/usr/bin/env bash
# Diffs .context/plugins/pantheon-org/** (source) against
# .tessl/plugins/pantheon-org/** (the mirror `tessl install` produces) for
# each helper skill, to catch a `tessl install` bug (skipped file, truncated
# content, a newly-added skill not picked up) or a stale mirror.
#
# Diff-only, no install side effects — this script assumes the mirror is
# already populated by a prior `tessl install` (the CI step's job, per
# .context/plans/tessl-mirror-drift-check-2026-07-06.md's Phase 3, not this
# script's), so it can be re-run repeatedly against the same `.tessl/` state.
#
# Content-only comparison (`diff -rq`); timestamps and permissions are not
# compared. `evals/` is excluded from every skill — deliberately not part of
# the installed/published skill for any helper skill in this repo. `.aislop/`
# is also excluded — a local, gitignored aislop-scan artifact directory that
# can appear anywhere under .context/** (see .gitignore) but is never source
# content; without this exclusion a contributor with a stale local scan
# session gets a false-positive divergence unrelated to the actual skill.
# A missing directory on either side counts as divergence: a source skill
# with no installed counterpart (newly added, not yet installed) and an
# installed skill with no source counterpart (deleted from source, stale in
# the mirror) both fail.
#
# Scoped to pantheon-org/** only — pantheon-ai/** and tessl-labs/** are
# third-party registry content, out of scope per ADR-047.
#
# Exits 1 if any skill diverges, 0 otherwise.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
SOURCE_ROOT="$ROOT/.context/plugins/pantheon-org"
MIRROR_ROOT="$ROOT/.tessl/plugins/pantheon-org"
COUNT=0

if [ ! -d "$SOURCE_ROOT" ]; then
    echo "✗ $SOURCE_ROOT not found — nothing to check" >&2
    exit 1
fi

if [ ! -d "$MIRROR_ROOT" ]; then
    echo "✗ $MIRROR_ROOT not found — run 'tessl install' first" >&2
    exit 1
fi

# One skill dir per <domain>/<skill>, two levels deep under SOURCE_ROOT.
while IFS= read -r source_dir; do
    skill_rel="${source_dir#"$SOURCE_ROOT"/}"
    mirror_dir="$MIRROR_ROOT/$skill_rel"

    if [ ! -d "$mirror_dir" ]; then
        echo "✗ $skill_rel: missing from mirror (source-only — newly added, not yet installed)"
        COUNT=$((COUNT + 1))
        continue
    fi

    diff_out=$(diff -rq --exclude=evals --exclude=.aislop "$source_dir" "$mirror_dir" 2>&1 || true)
    if [ -n "$diff_out" ]; then
        echo "✗ $skill_rel: source and mirror diverge"
        echo "    ${diff_out//$'\n'/$'\n    '}"
        COUNT=$((COUNT + 1))
    fi
done < <(find "$SOURCE_ROOT" -mindepth 2 -maxdepth 2 -type d | sort)

# Installed-only skills: present in the mirror but deleted from source.
while IFS= read -r mirror_dir; do
    skill_rel="${mirror_dir#"$MIRROR_ROOT"/}"
    source_dir="$SOURCE_ROOT/$skill_rel"

    if [ ! -d "$source_dir" ]; then
        echo "✗ $skill_rel: installed-only (deleted from source, stale in the mirror)"
        COUNT=$((COUNT + 1))
    fi
done < <(find "$MIRROR_ROOT" -mindepth 2 -maxdepth 2 -type d | sort)

if [ "$COUNT" -gt 0 ]; then
    echo
    echo "→ $COUNT pantheon-org skill(s) diverge between source and the installed mirror."
    exit 1
fi
exit 0
