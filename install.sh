#!/bin/sh
# Bootstrap installer for the pkg.arifcinartekin.me tool repository.
#
#   curl -fsSL https://pkg.arifcinartekin.me/sh | sh
#
# On Debian/Ubuntu-family Linux this also registers the APT repository
# (GPG key + /etc/apt/sources.list.d entry) so packages published under
# packages/*.deb can be installed with apt. On every supported platform
# (any Linux distro, macOS, Termux) it installs the "toolbox" CLI used to
# list/install/update/uninstall the scripts and apps in this repository.
#
# Safe to re-run: every step is idempotent.
set -e

BASE_URL="https://pkg.arifcinartekin.me"
GH_OWNER="arifcinartekin"
GH_REPO="tools"
KEYRING_PATH="/usr/share/keyrings/pkg-arifcinartekin.gpg"
SOURCES_LIST="/etc/apt/sources.list.d/pkg-arifcinartekin.list"
SOURCES_LINE="deb [signed-by=${KEYRING_PATH}] ${BASE_URL}/apt-repo stable main"

fatal() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

info() {
  printf '==> %s\n' "$1"
}

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

command_exists curl || fatal "curl is required but was not found. Install curl and re-run this script."

is_termux() {
  [ -n "${TERMUX_VERSION:-}" ] || { [ -n "${PREFIX:-}" ] && [ "${PREFIX#*com.termux}" != "$PREFIX" ]; }
}

# Runs its arguments as root: directly if already root, via sudo otherwise.
as_root() {
  if [ "$(id -u)" = "0" ]; then
    "$@"
  elif command_exists sudo; then
    sudo "$@"
  else
    fatal "root privileges are required (install sudo, or run this script as root)."
  fi
}

OS_KIND="unknown"    # termux | macos | linux-apt | linux-other
UNAME_S="$(uname -s)"

if is_termux; then
  OS_KIND="termux"
elif [ "$UNAME_S" = "Darwin" ]; then
  OS_KIND="macos"
elif [ "$UNAME_S" = "Linux" ]; then
  if command_exists apt-get; then
    OS_KIND="linux-apt"
  else
    OS_KIND="linux-other"
  fi
else
  fatal "unsupported platform: $UNAME_S"
fi

info "Detected platform: $OS_KIND"

# --- APT repository setup (Debian/Ubuntu and derivatives only) -------------

setup_apt_repo() {
  # The APT repo (and its signing key) only exist once the maintainer has
  # generated a GPG key and the apt-repo workflow has published it - don't
  # hard-fail the whole install over that; just skip this part.
  if ! curl -fsSL --head "${BASE_URL}/apt-repo/public.key" >/dev/null 2>&1; then
    info "APT repository not published yet (no key at ${BASE_URL}/apt-repo/public.key) - skipping APT setup."
    return 0
  fi

  MISSING_DEPS=""
  for dep in gnupg ca-certificates; do
    dpkg -s "$dep" >/dev/null 2>&1 || MISSING_DEPS="$MISSING_DEPS $dep"
  done
  if [ -n "$MISSING_DEPS" ]; then
    info "Installing missing dependencies:$MISSING_DEPS"
    as_root apt-get update -qq
    # shellcheck disable=SC2086
    as_root apt-get install -y -qq $MISSING_DEPS
  fi

  info "Installing GPG signing key to $KEYRING_PATH"
  TMP_KEY="$(mktemp)"
  if ! curl -fsSL "${BASE_URL}/apt-repo/public.key" -o "$TMP_KEY"; then
    rm -f "$TMP_KEY"
    fatal "failed to download ${BASE_URL}/apt-repo/public.key"
  fi
  as_root gpg --batch --yes --dearmor -o "$KEYRING_PATH" "$TMP_KEY"
  rm -f "$TMP_KEY"

  if [ -f "$SOURCES_LIST" ] && [ "$(cat "$SOURCES_LIST")" = "$SOURCES_LINE" ]; then
    info "APT source already up to date ($SOURCES_LIST)"
  else
    info "Writing APT source to $SOURCES_LIST"
    printf '%s\n' "$SOURCES_LINE" | as_root tee "$SOURCES_LIST" >/dev/null
  fi

  info "Running apt-get update"
  as_root apt-get update -qq
}

if [ "$OS_KIND" = "linux-apt" ]; then
  setup_apt_repo
else
  info "Skipping APT repository setup (not a Debian/Ubuntu-family system)."
fi

# --- toolbox CLI install (every platform) -----------------------------------

case "$(uname -m)" in
  x86_64|amd64) CLI_ARCH="amd64" ;;
  aarch64|arm64) CLI_ARCH="arm64" ;;
  armv7l|armv7) CLI_ARCH="arm" ;;
  *) fatal "unsupported CPU architecture: $(uname -m)" ;;
esac

case "$OS_KIND" in
  macos) CLI_OS="darwin" ;;
  termux|linux-apt|linux-other) CLI_OS="linux" ;;
esac

if [ "$OS_KIND" = "termux" ]; then
  BIN_DIR="${PREFIX}/bin"
else
  BIN_DIR="$HOME/.local/bin"
fi

CLI_ASSET="toolbox-${CLI_OS}-${CLI_ARCH}"
CLI_URL="https://github.com/${GH_OWNER}/${GH_REPO}/releases/latest/download/${CLI_ASSET}"
CLI_DEST="${BIN_DIR}/toolbox"

info "Installing toolbox CLI ($CLI_ASSET) to $CLI_DEST"
mkdir -p "$BIN_DIR"
TMP_CLI="$(mktemp)"
trap 'rm -f "$TMP_CLI"' EXIT
curl -fsSL "$CLI_URL" -o "$TMP_CLI" || fatal "failed to download $CLI_URL"
chmod +x "$TMP_CLI"
mv "$TMP_CLI" "$CLI_DEST"
trap - EXIT

# This process cannot change the parent shell's PATH (the usual install
# command pipes this script into sh), so persist the change in the profile
# for the shell that launched it instead. Do this even when the current PATH
# already contains BIN_DIR: it may have been set only for this shell session.
SHELL_NAME="${SHELL##*/}"
case "$SHELL_NAME" in
  zsh) PROFILE_PATH="$HOME/.zshrc" ;;
  bash) PROFILE_PATH="$HOME/.bashrc" ;;
  fish) PROFILE_PATH="$HOME/.config/fish/config.fish" ;;
  *) PROFILE_PATH="$HOME/.profile" ;;
esac

case "$SHELL_NAME" in
  fish) PATH_LINE="fish_add_path -g $BIN_DIR" ;;
  *) PATH_LINE="export PATH=\"$BIN_DIR:\$PATH\"" ;;
esac

if [ ! -f "$PROFILE_PATH" ] || ! grep -Fqx "$PATH_LINE" "$PROFILE_PATH" >/dev/null 2>&1; then
  info "Adding $BIN_DIR to PATH in $PROFILE_PATH"
  mkdir -p "$(dirname "$PROFILE_PATH")"
  {
    printf '\n# Added by the toolbox installer\n'
    printf '%s\n' "$PATH_LINE"
  } >> "$PROFILE_PATH"
else
  info "$BIN_DIR is already configured in $PROFILE_PATH"
fi
info "Open a new terminal, or run: $PATH_LINE"

info "Done. Run 'toolbox list' to see what's available in the repository."
