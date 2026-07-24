#!/usr/bin/env bash
set -euo pipefail

# toolbox installs the entrypoint as a symlink under ~/.local/bin, so
# BASH_SOURCE must be resolved through the symlink chain to find the app's
# real install directory (~/.local/share/hello-app) before sourcing lib/.
source_path="${BASH_SOURCE[0]}"
while [ -h "$source_path" ]; do
  link_dir="$(cd -P "$(dirname "$source_path")" && pwd)"
  source_path="$(readlink "$source_path")"
  [[ $source_path != /* ]] && source_path="$link_dir/$source_path"
done
bin_dir="$(cd -P "$(dirname "$source_path")" && pwd)"
app_dir="$(cd "$bin_dir/.." && pwd)"

# shellcheck source=../lib/greeting.sh
source "$app_dir/lib/greeting.sh"

print_greeting "hello-app"
