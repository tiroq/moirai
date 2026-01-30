# moirai

![Moirai Logo](./assets/moirai.png)

Moirai is a CLI tool for managing oh-my-opencode configuration profiles. It discovers profile files, switches the active config by updating a symlink, and can help with backups and diffs so you can move between setups safely.

If you keep multiple oh-my-opencode profiles for different workflows, moirai gives you a small, focused interface to list and apply them without hand-editing files.

## Installation

Download the latest release from GitHub Releases, then install it manually:

```
chmod +x moirai-<os>-<arch>
mv moirai-<os>-<arch> /usr/local/bin/moirai
```

## Release (local snapshot)

Build a local snapshot release with GoReleaser:

```
goreleaser release --snapshot --clean
```

## CI (local)

Run the full CI suite locally:

```
task ci
```

## Release (publish)

Tag and push a release:

```
git tag v0.1.0 && git push origin v0.1.0
```

## Basic usage

List available profiles and show the active one:

```
moirai list
```

Switch to a profile:

```
moirai apply <profile>
```

## Config location

Moirai reads profiles from `~/.config/opencode/`.

## Safety note

Moirai treats the active config as a symlink to a profile file and uses backups when making changes. Review backups and symlinks before restoring or applying profiles.
