# pkg.arifcinartekin.me

A single-repository "tool repository / package manager" for deploying
personal scripts and apps to multiple machines with one command, and an APT
repository published from the same repo via GitHub Pages.

Two independent ways to install things from here:

- **`toolbox`** — a small cross-platform CLI that installs/updates plain
  scripts and apps (`scripts/`, `apps/`) on any machine with `bash`: Linux
  (any distro — Debian/Ubuntu, Arch, Void, Fedora, Alpine, ...), macOS, and
  Termux.
- **APT** — for `.deb` packages (`packages/`), served as a real signed APT
  repository. Only relevant on Debian/Ubuntu-family Linux.

## Quick start

```sh
curl -fsSL https://pkg.arifcinartekin.me/sh | sh
```

This installs the `toolbox` CLI and, on Debian/Ubuntu-family systems, also
registers the APT repository. See [How `install.sh` works](#how-installsh-works)
for exactly what it does.

```sh
toolbox list                 # see everything available
toolbox install <name>       # install one item
toolbox update <name>        # update one item
toolbox update --all         # update everything installed
toolbox uninstall <name>     # remove an item
```

## Repository layout

```
scripts/<name>/<entrypoint>   Single-file tools (bash, python, ...)
scripts/<name>/meta.json
apps/<name>/...                Multi-file tools
apps/<name>/meta.json
packages/*.deb                 Drop .deb files here to publish them via APT
apt-repo/                      reprepro output (dists/, pool/, public.key) — generated, do not edit
cli/                            Source of the toolbox CLI (Go)
manifest.json                   Generated index of scripts/ and apps/ — do not edit
install.sh / sh                 Bootstrap installer, served at both paths
.github/workflows/              Automation (see below)
```

## Adding a new script

A script is a single executable file. Create `scripts/<name>/<name>.sh` (or
`.py`, etc.) plus a `scripts/<name>/meta.json`:

```json
{
  "name": "hello-toolbox",
  "description": "Prints a greeting.",
  "version": "1.0.0",
  "entrypoint": "hello-toolbox.sh",
  "target_path": "~/.local/bin/hello-toolbox",
  "os": ["linux", "darwin"],
  "executable": true
}
```

- `name` must match the directory name.
- `entrypoint` is the file downloaded and installed to `target_path`.
- `os` lists supported platforms (`linux`, `darwin`, or omit the field for
  "any"). `toolbox` refuses to install an item on an unsupported OS.
- Bump `version` (semver: `MAJOR.MINOR.PATCH`) whenever you change the
  script — `toolbox update` compares this against what's installed.

Commit and push to `main`. The manifest workflow regenerates `manifest.json`
automatically; nothing else to do.

## Adding a new app

An app is a directory of files installed as a unit, optionally with one
file symlinked into `PATH`. Create `apps/<name>/` with whatever file tree you
need, plus `apps/<name>/meta.json`:

```json
{
  "name": "hello-app",
  "description": "Multi-file greeter.",
  "version": "1.0.0",
  "entrypoint": "bin/hello-app.sh",
  "link_path": "~/.local/bin/hello-app",
  "target_path": "~/.local/share/hello-app",
  "os": ["linux", "darwin"],
  "executable": true
}
```

- The entire `apps/<name>/` tree (minus `meta.json`) is installed under
  `target_path`.
- If `entrypoint` and `link_path` are both set, `toolbox` symlinks
  `target_path/<entrypoint>` to `link_path` and makes it executable.
- If your entrypoint script needs to find other files in the app (e.g. a
  `lib/` directory), resolve its own location through the symlink first —
  see `apps/hello-app/bin/hello-app.sh` for the portable pattern, since
  `toolbox` installs the entrypoint as a symlink, not a copy.

## `meta.json` reference

| Field         | Required | Meaning                                                              |
|---------------|----------|------------------------------------------------------------------------|
| `name`        | yes      | Must match the directory name.                                       |
| `description` | yes      | Shown by `toolbox list`.                                             |
| `version`     | yes      | Semver `MAJOR.MINOR.PATCH`.                                          |
| `entrypoint`  | yes      | Relative path to the main file.                                      |
| `target_path` | yes      | Where it's installed (`~` is expanded to `$HOME`).                   |
| `os`          | no       | `["linux"]`, `["darwin"]`, or both. Omit for no restriction.         |
| `executable`  | no       | Whether the entrypoint should be `chmod +x`'d.                       |
| `link_path`   | apps only| Where the entrypoint is symlinked into `PATH`.                       |

`kind` (`script` vs `app`), `base_path`, and `files` are computed
automatically by the manifest builder — do not set them by hand.

## How `manifest.json` is generated

`.github/scripts/build-manifest.sh` scans every `scripts/*/meta.json` and
`apps/*/meta.json`, computes each item's file list, and writes
`manifest.json` at the repo root. The `Build manifest` workflow
(`.github/workflows/manifest.yml`) runs this on every push to `main` that
touches `scripts/`, `apps/`, or `install.sh`, and commits the result. It also
keeps the extensionless `sh` file in sync with `install.sh`. Never edit
`manifest.json` or `sh` by hand — they will be overwritten.

## How `install.sh` works

Run via `curl -fsSL https://pkg.arifcinartekin.me/sh | sh` (or
`/install.sh` — identical content, served at both paths). It:

1. Detects the platform: Termux, macOS, Debian/Ubuntu-family Linux (has
   `apt-get`), or any other Linux.
2. **Only on Debian/Ubuntu-family Linux**: installs `gnupg`/`ca-certificates`
   if missing, imports the repository's GPG key to
   `/usr/share/keyrings/pkg-arifcinartekin.gpg`, writes
   `/etc/apt/sources.list.d/pkg-arifcinartekin.list`, and runs
   `apt-get update`. `sudo` is used only for these steps, and only when not
   already root.
3. **On every platform**: downloads the `toolbox` CLI binary matching your
   OS/architecture from the latest GitHub Release and installs it to
   `~/.local/bin` (or `$PREFIX/bin` on Termux, which needs no root at all).
   It also ensures that directory is in the profile for the invoking user's
   shell (`.zshrc`, `.bashrc`, Fish's `config.fish`, or `.profile`). Open a
   new terminal after installation.

Every step is idempotent — re-running the script is safe and just
re-verifies/updates things in place.

## The `toolbox` CLI

Source in `cli/` (Go, compiled to a single static binary — no runtime
dependency on the target machine, which matters since it's fetched with a
one-line `curl` and run directly). Cross-compiled for
`linux/amd64`, `linux/arm64`, `linux/arm`, `darwin/amd64`, `darwin/arm64` by
`.github/workflows/cli-release.yml` on every push to `cli/`, published as a
GitHub Release. `install.sh` always grabs the latest release.

```sh
toolbox list                    # name, kind, versions (repo vs installed), description
toolbox install <name>...       # install one or more items
toolbox update <name>...        # update specific items
toolbox update --all            # update everything installed
toolbox uninstall <name>...     # remove one or more items
toolbox version
```

Local state (what's installed, at which version) lives in
`~/.toolbox/installed.json`. Override the repository URL for local testing
with `TOOLBOX_BASE_URL` (defaults to `https://pkg.arifcinartekin.me`).

## Using the APT repository

Debian/Ubuntu-family systems get this automatically from `install.sh`. To
set it up manually:

```sh
curl -fsSL https://pkg.arifcinartekin.me/apt-repo/public.key \
  | sudo gpg --batch --yes --dearmor -o /usr/share/keyrings/pkg-arifcinartekin.gpg

echo "deb [signed-by=/usr/share/keyrings/pkg-arifcinartekin.gpg] https://pkg.arifcinartekin.me/apt-repo stable main" \
  | sudo tee /etc/apt/sources.list.d/pkg-arifcinartekin.list

sudo apt update
```

### Publishing a `.deb`

Drop the file in `packages/` and push to `main`.
`.github/workflows/apt-repo.yml` imports it into `apt-repo/` with
`reprepro`, signs the repository with the GPG key held in repo secrets, and
removes the source `.deb` from `packages/` once published. `apt-repo/` is
generated — never edit `dists/`, `pool/`, or `public.key` by hand.

## Automation overview

| Workflow                            | Trigger                          | Does                                                        |
|--------------------------------------|-----------------------------------|---------------------------------------------------------------|
| `manifest.yml`                      | push touching `scripts/`, `apps/`, `install.sh` | Rebuilds `manifest.json`, syncs `sh` to `install.sh`, commits |
| `apt-repo.yml`                      | push touching `packages/`        | `reprepro includedeb` + sign + commit `apt-repo/`             |
| `cli-release.yml`                   | push touching `cli/`             | Cross-compiles `toolbox`, publishes a GitHub Release           |

GitHub Pages serves the repository root on `main` at
`https://pkg.arifcinartekin.me` (see `CNAME`).
