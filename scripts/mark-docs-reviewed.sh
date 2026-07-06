#!/usr/bin/env bash
# Records that a doc has been reviewed and confirmed accurate as of HEAD's
# commit, so check-docs-drift.sh's cumulative mode stops flagging source
# commits at or before that point. Does not require the doc to currently be
# flagged, and performs no authorization check — this is bookkeeping for an
# advisory tool, not an access control (see ADR-045).
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
SIDECAR="$ROOT/scripts/docs-drift-reviewed.jsonl"
CHECK_SCRIPT="$ROOT/scripts/check-docs-drift.sh"

if [ "$#" -eq 0 ]; then
    echo "Usage: $0 <doc_path> [doc_path...]" >&2
    echo "Records HEAD's commit date as the reviewed baseline for each doc." >&2
    exit 1
fi

mapped_docs=$(sed -n 's/^ *"\([^|]*\)|.*"$/\1/p' "$CHECK_SCRIPT")
reviewed_iso=$(git -C "$ROOT" log -1 --format='%cI')
reviewed_epoch=$(git -C "$ROOT" log -1 --format='%ct')

for doc in "$@"; do
    if ! grep -qxF "$doc" <<<"$mapped_docs"; then
        echo "Warning: $doc is not in check-docs-drift.sh's MAPPINGS — this entry will never be read." >&2
    fi

    if [ ! -f "$ROOT/$doc" ]; then
        echo "Warning: $doc does not exist in the working tree." >&2
    fi

    globs=$(sed -n "s#^ *\"$doc|\\(.*\\)\"\$#\\1#p" "$CHECK_SCRIPT")
    if [ -n "$globs" ]; then
        doc_date=$(git -C "$ROOT" log -1 --format=%cI -- "$doc" 2>/dev/null || true)
        if [ -n "$doc_date" ]; then
            IFS=';' read -ra glob_arr <<<"$globs"
            hits=""
            for g in "${glob_arr[@]}"; do
                matches=$(git -C "$ROOT" log --oneline --after="$doc_date" -- "$g" 2>/dev/null || true)
                [ -z "$matches" ] && continue
                hits="${hits}${matches}"$'\n'
            done
            if [ -n "$hits" ]; then
                hits="${hits%$'\n'}"
                echo "Confirming $doc reviewed despite these source changes since its last edit:"
                echo "    ${hits//$'\n'/$'\n    '}"
            fi
        fi
    fi

    echo "Marking $doc reviewed as of $reviewed_iso (HEAD)."

    tmp=$(mktemp)
    if [ -f "$SIDECAR" ]; then
        jq -c --arg d "$doc" 'select(.doc != $d)' "$SIDECAR" > "$tmp"
    fi
    jq -nc --arg doc "$doc" --arg iso "$reviewed_iso" --argjson epoch "$reviewed_epoch" \
        '{doc: $doc, reviewed_iso: $iso, reviewed_epoch: $epoch}' >> "$tmp"
    jq -sc 'sort_by(.doc)[]' "$tmp" > "$tmp.sorted"
    mv "$tmp.sorted" "$SIDECAR"
    rm -f "$tmp"
done
