#!/usr/bin/env bash
# Scans scripts/*/meta.json and apps/*/meta.json and regenerates manifest.json
# at the repository root. This is the single source of truth toolbox reads
# from; it must never be hand-edited (see .github/workflows/manifest.yml).
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

items="[]"

add_group() {
  local kind="$1" dir="$2"
  [ -d "$dir" ] || return 0

  for meta_path in "$dir"/*/meta.json; do
    [ -e "$meta_path" ] || continue

    local item_dir base_path dir_name meta_name
    item_dir="$(dirname "$meta_path")"
    dir_name="$(basename "$item_dir")"
    base_path="${dir}/${dir_name}"

    meta_name="$(jq -r '.name' "$meta_path")"
    if [ "$meta_name" != "$dir_name" ]; then
      echo "error: ${meta_path}: \"name\" (\"$meta_name\") must match its directory name (\"$dir_name\")" >&2
      exit 1
    fi

    # File list: everything in the item directory except meta.json itself,
    # relative to the item directory, sorted for deterministic output.
    local files_json
    files_json="$(cd "$item_dir" && find . -type f ! -name meta.json | sed 's#^\./##' | sort | jq -R . | jq -s .)"

    local entrypoint
    entrypoint="$(jq -r '.entrypoint // empty' "$meta_path")"
    if [ "$kind" = "script" ]; then
      # Scripts are exactly one file: the entrypoint itself.
      files_json="$(jq -n --arg e "$entrypoint" '[$e]')"
    fi

    local item
    item="$(jq -c \
      --arg kind "$kind" \
      --arg base_path "$base_path" \
      --argjson files "$files_json" \
      '. + {kind: $kind, base_path: $base_path, files: $files}' \
      "$meta_path")"

    items="$(jq -c --argjson item "$item" '. + [$item]' <<<"$items")"
  done
}

add_group "script" "scripts"
add_group "app" "apps"

generated_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
jq -n --arg generated_at "$generated_at" --argjson items "$items" \
  '{generated_at: $generated_at, items: ($items | sort_by(.name))}' \
  > manifest.json

echo "Wrote manifest.json with $(jq '.items | length' manifest.json) item(s)."
