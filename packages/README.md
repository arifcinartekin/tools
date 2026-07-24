# packages/

Drop `.deb` files here and push to `main`. The `Build APT repository` workflow
(`.github/workflows/apt-repo.yml`) picks them up automatically: it imports
each one into `apt-repo/` with `reprepro`, signs the repository, and removes
the source `.deb` from this directory once it has been published.

This directory is an inbox, not storage — published packages live in
`apt-repo/pool/`, not here.
