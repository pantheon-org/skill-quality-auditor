#!/usr/bin/env bash
# Pure-bash helpers for merge-status-sync.sh. No python/node — see
# .agents/RULES.md "Avoid Python/Node.js scripts in skills".
#
# frontmatter list parsing mirrors scripts/check-plan-drift.sh's existing
# `related:` walk (Decision 5 in
# .context/plans/post-merge-status-sync-2026-07-04.md).

# frontmatter_value <file> <key>
# Prints the scalar value of a top-level frontmatter key, e.g. "status".
frontmatter_value() {
  local file="$1" key="$2"
  sed -n '/^---$/,/^---$/p' "$file" | sed '1d;$d' |
    grep -E "^${key}: " | head -1 | sed -E "s/^${key}: *//" | tr -d '"'
}

# normalize_path <path>
# Collapses "." and ".." segments in an absolute path string without
# requiring the path to exist. Portable stand-in for GNU `realpath -m`,
# which macOS/BSD `realpath` doesn't support at all (no -m flag) — see
# .agents/RULES.md "Avoid Python/Node.js scripts in skills": this repo's
# skill scripts must not assume a GNU userland either.
normalize_path() {
  local input="$1" part result=""
  local IFS='/'
  for part in $input; do
    case "$part" in
      "" | ".") continue ;;
      "..") result="${result%/*}" ;;
      *) result="$result/$part" ;;
    esac
  done
  echo "$result"
}

# frontmatter_list <file> <key> <base_dir>
# Prints repo-relative paths resolved from a YAML list under <key>,
# resolving each entry against <base_dir> (the plan's own directory for
# `related:`, or $ROOT for an ADR's root-relative `context:` entries).
# Handles both plain-string list items (`- ../../foo.go`) and
# path-subkey items (`- path: "docs/ADR/adr-001.md"`).
#
# Uses `case` glob matching rather than `[[ =~ ]]`/BASH_REMATCH: under
# bash 3.2 (macOS's stock /bin/bash), BASH_REMATCH capture groups come
# back empty when this exact pattern runs inside a function loaded via
# `source` — a real, reproduced quirk, not a hypothetical one. `case` glob
# matching sidesteps it entirely and is portable across bash versions.
frontmatter_list() {
  local file="$1" key="$2" base_dir="$3"
  local frontmatter in_list=false
  frontmatter=$(sed -n '/^---$/,/^---$/p' "$file" | sed '1d;$d')

  while IFS= read -r line; do
    if [ "$line" = "${key}:" ]; then
      in_list=true
      continue
    fi
    if $in_list; then
      case "$line" in
        [a-zA-Z]* | ---)
          in_list=false
          continue
          ;;
      esac
      case "$line" in
        [[:space:]]*-[[:space:]]*)
          local raw="${line#*- }"
          raw="${raw#path: }"
          raw="${raw%\"}"
          raw="${raw#\"}"
          local resolved repo_rel
          resolved="$(normalize_path "$base_dir/$raw")"
          [ -n "$resolved" ] || continue
          repo_rel="${resolved#"$ROOT"/}"
          [ "$repo_rel" != "$resolved" ] || continue
          echo "$repo_rel"
          ;;
      esac
    fi
  done <<<"$frontmatter"
}

# phase_count <file>
# Counts "### Phase N" headings in the file body (frontmatter excluded by
# the heading pattern never appearing inside it).
phase_count() {
  grep -c '^### Phase [0-9]' "$1" 2>/dev/null || true
}

# is_governance_path <rel_path>
is_governance_path() {
  case "$1" in
    .context/* | docs/ADR/*) return 0 ;;
    *) return 1 ;;
  esac
}

# path_in_list <needle> <haystack_file>
# <haystack_file> holds one path per line (the PR's touched-files set).
path_in_list() {
  grep -qxF "$1" "$2" 2>/dev/null
}

# classify_signal <rel_path> <touched_file> <linked_paths_file>
# Prints one of: direct, frontmatter, file-touch, none.
classify_signal() {
  local rel_path="$1" touched_file="$2" linked_file="$3"
  if path_in_list "$rel_path" "$touched_file"; then
    echo "direct"
    return
  fi
  local link
  if [ -s "$linked_file" ]; then
    while IFS= read -r link; do
      [ -n "$link" ] || continue
      if path_in_list "$link" "$touched_file" && is_governance_path "$link"; then
        echo "frontmatter"
        return
      fi
    done <"$linked_file"
    while IFS= read -r link; do
      [ -n "$link" ] || continue
      if path_in_list "$link" "$touched_file"; then
        echo "file-touch"
        return
      fi
    done <"$linked_file"
  fi
  echo "none"
}

# detect_candidates <root> <touched_file>
# Prints tab-separated records:
#   type<TAB>path<TAB>status<TAB>signal<TAB>phase_count<TAB>auto_flip<TAB>target_status
# auto_flip is "1" or "0". phase_count is empty for ADR records.
detect_candidates() {
  local root="$1" touched_file="$2"
  local plans_dir="$root/.context/plans"
  local adr_dir="$root/docs/ADR"
  local links_file
  links_file=$(mktemp)
  trap 'rm -f "$links_file"' RETURN

  if [ -d "$plans_dir" ]; then
    local md
    while IFS= read -r -d '' md; do
      local status
      status=$(frontmatter_value "$md" "status")
      case "$status" in
        ACTIVE | DRAFT) ;;
        *) continue ;;
      esac
      local rel_path
      rel_path="${md#"$root"/}"
      frontmatter_list "$md" "related" "$(dirname "$md")" >"$links_file"
      local signal
      signal=$(classify_signal "$rel_path" "$touched_file" "$links_file")
      [ "$signal" != "none" ] || continue
      local phases single_phase auto_flip
      phases=$(phase_count "$md")
      [ "$phases" -le 1 ] 2>/dev/null && single_phase=1 || single_phase=0
      if [ "$single_phase" -eq 1 ] && { [ "$signal" = "direct" ] || [ "$signal" = "frontmatter" ]; }; then
        auto_flip=1
      else
        auto_flip=0
      fi
      printf 'plan\t%s\t%s\t%s\t%s\t%s\tDONE\n' "$rel_path" "$status" "$signal" "$phases" "$auto_flip"
    done < <(find "$plans_dir" -name '*.md' -print0 | sort -z)
  fi

  if [ -d "$adr_dir" ]; then
    local md
    while IFS= read -r -d '' md; do
      local status
      status=$(frontmatter_value "$md" "status")
      [ "$status" = "proposed" ] || continue
      local rel_path
      rel_path="${md#"$root"/}"
      frontmatter_list "$md" "context" "$root" >"$links_file"
      local signal
      signal=$(classify_signal "$rel_path" "$touched_file" "$links_file")
      [ "$signal" != "none" ] || continue
      # Decision 2: ADR flips always require confirmation, regardless of
      # signal strength.
      printf 'adr\t%s\t%s\t%s\t\t0\taccepted\n' "$rel_path" "$status" "$signal"
    done < <(find "$adr_dir" -maxdepth 1 -name 'adr-*.md' -print0 | sort -z)
  fi
}
