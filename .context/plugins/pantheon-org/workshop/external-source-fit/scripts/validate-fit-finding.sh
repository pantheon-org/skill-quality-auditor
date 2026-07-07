#!/usr/bin/env bash
# Validate the structured fit-assessment block embedded in an external-source-fit
# finding. Pure bash + awk/sed/grep — no python, node, or yq — per the repo rule
# "Avoid Python/Node.js scripts in skills".
#
# Usage: validate-fit-finding.sh <finding.md> [<finding.md> ...]
#
# The block is the fenced ```yaml ... ``` that immediately follows a line
# reading `<!-- fit-assessment -->`. The checks below MIRROR the contract in
# assets/schemas/fit-assessment.schema.json (that file is the human- and
# tool-facing spec; keep the two in sync when either changes).
set -euo pipefail

# ---- enum sets (mirror the schema) -----------------------------------------
VERDICTS="Good fit|Partial fit|No fit"
VALUES="LOW|MEDIUM|HIGH"
VEHICLES="go-cli|helper-skill|none"
ACTIONS="draft-plan|build-natively|record-and-hold|reject"
LEVELS="none|partial|full"
BOOLS="true|false"
TOP_KEYS="schema_version source characterisation overlap verdict vehicle_if_adopted salvageable recommendation value"
OVERLAP_KEYS="d1_d9_scorers validate_analyze duplication eval_runner helper_skills"

# Extract the YAML lines between the fence that follows the marker.
extract_block() {
  awk '
    /<!-- fit-assessment -->/ { seen=1; next }
    seen && /^```/ {
      if (infence) { exit }   # closing fence
      infence=1; next         # opening fence
    }
    infence { print }
  ' "$1"
}

# Value of a top-level scalar key, unquoted and trimmed ("" if absent).
top_scalar() {
  sed -nE "s/^$1:[[:space:]]*//p" "$2" | head -1 | sed -E 's/^"(.*)"$/\1/; s/^'\''(.*)'\''$/\1/; s/[[:space:]]+$//'
}

# Value of a key at a given indent (e.g. nested action:/present:/level:).
indented_scalar() {
  # $1 indent-spaces, $2 key, $3 file — returns all matches, one per line
  sed -nE "s/^$1$2:[[:space:]]*//p" "$3" | sed -E 's/^"(.*)"$/\1/; s/^'\''(.*)'\''$/\1/; s/[[:space:]]+$//'
}

in_set() { echo "$2" | grep -qxE "$1"; }

validate_block() {
  local blk="$1" errs=()

  # required top-level keys
  local k
  for k in $TOP_KEYS; do
    grep -qE "^$k:" "$blk" || errs+=("missing required key '$k'")
  done

  # schema_version == 1
  local sv; sv="$(top_scalar schema_version "$blk")"
  [[ "$sv" == "1" ]] || errs+=("schema_version must be 1 (got '${sv:-<empty>}')")

  # source.name and source.url present (2-space indent under source:)
  grep -qE "^  name:[[:space:]]*\S" "$blk" || errs+=("source.name is required and non-empty")
  grep -qE "^  url:[[:space:]]*https?://" "$blk" || errs+=("source.url is required and must start with http(s)://")

  # characterisation non-trivial (>= 20 chars on the value or a block scalar)
  local ch; ch="$(top_scalar characterisation "$blk")"
  if [[ "$ch" == ">-" || "$ch" == ">" || "$ch" == "|" || "$ch" == "|-" ]]; then
    : # block scalar; body follows on indented lines — presence is enough
    grep -qE "^  \S" "$blk" || errs+=("characterisation block scalar is empty")
  elif [[ ${#ch} -lt 20 ]]; then
    errs+=("characterisation must be >= 20 chars (got ${#ch})")
  fi

  # enum leaves
  local verdict value vehicle
  verdict="$(top_scalar verdict "$blk")"
  in_set "$VERDICTS" "$verdict" || errs+=("verdict '$verdict' not one of: ${VERDICTS//|/, }")
  value="$(top_scalar value "$blk")"
  in_set "$VALUES" "$value" || errs+=("value '$value' not one of: ${VALUES//|/, }")
  vehicle="$(top_scalar vehicle_if_adopted "$blk")"
  in_set "$VEHICLES" "$vehicle" || errs+=("vehicle_if_adopted '$vehicle' not one of: ${VEHICLES//|/, }")

  # overlap: all five surfaces present, each with a valid level
  for k in $OVERLAP_KEYS; do
    grep -qE "^  $k:" "$blk" || errs+=("overlap.$k is required")
  done
  local lvl
  while IFS= read -r lvl; do
    [[ -z "$lvl" ]] && continue
    in_set "$LEVELS" "$lvl" || errs+=("overlap level '$lvl' not one of: ${LEVELS//|/, }")
  done < <(indented_scalar "    " level "$blk")
  # count of levels must equal the five surfaces
  local nlev; nlev="$(indented_scalar "    " level "$blk" | grep -c . || true)"
  [[ "$nlev" -eq 5 ]] || errs+=("expected 5 overlap levels, found $nlev")

  # salvageable.present is boolean; description present
  local present; present="$(indented_scalar "  " present "$blk")"
  in_set "$BOOLS" "$present" || errs+=("salvageable.present '$present' must be true or false")
  grep -qE "^  description:[[:space:]]*\S" "$blk" || errs+=("salvageable.description is required and non-empty")

  # recommendation.action enum; detail present
  local action; action="$(indented_scalar "  " action "$blk")"
  in_set "$ACTIONS" "$action" || errs+=("recommendation.action '$action' not one of: ${ACTIONS//|/, }")
  grep -qE "^  detail:[[:space:]]*\S" "$blk" || errs+=("recommendation.detail is required and non-empty")

  if [[ ${#errs[@]} -gt 0 ]]; then
    printf '%s\n' "${errs[@]}"
    return 1
  fi
  return 0
}

if [[ $# -eq 0 ]]; then
  echo "usage: validate-fit-finding.sh <finding.md> [<finding.md> ...]" >&2
  exit 2
fi

fail=0
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

for f in "$@"; do
  if [[ ! -f "$f" ]]; then
    echo "SKIP: $f (not found)" >&2
    fail=1
    continue
  fi
  blk="$tmpdir/block.yaml"
  extract_block "$f" >"$blk"
  if [[ ! -s "$blk" ]]; then
    echo "FAIL: $f — no fit-assessment block found (expected '<!-- fit-assessment -->' then a yaml fence)" >&2
    fail=1
    continue
  fi
  if errs="$(validate_block "$blk")"; then
    echo "OK:   $f"
  else
    echo "FAIL: $f" >&2
    echo "$errs" | sed 's/^/    - /' >&2
    fail=1
  fi
done

exit "$fail"
