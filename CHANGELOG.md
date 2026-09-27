# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

---

## v0.6.0 - 2026-09-27

### Added
- Optional per-repo `bootstrapScript` config field to run a shell command (e.g. `pre-commit install`) in a repo's directory after clone and branch checkout; supports `{workspace}`/`{workspaceDirectory}` templating and warns without aborting workspace creation if it fails

---

## v0.5.0 - 2026-09-27

### Added
- `muster clean` command to find and delete workspaces that are safe to remove (no uncommitted changes, and either no divergence from `main`/`master` or divergent commits already merged), with output color-coded green/red by safety (respects `NO_COLOR`)
- `muster list` (or `muster ls`) command to list workspaces or stacks, with `--all` for a table view of stack/repo details

### Fixed
- `muster launch` now sets upstream tracking (e.g. `origin/main`) on branches it creates, so they no longer start out untracked
- `muster clean` no longer reports a "failed to check git status" error for repo directories that aren't actually git repositories (e.g. from a broken or incomplete clone); these are now reported as skipped and don't block a workspace from being safe to clean

---

## v0.4.0 - 2026-09-25

### Changed
- `muster launch --stack` now accepts a single stack name instead of multiple repeated flags

### Fixed
- Workspaces created from a stack are now nested under `~/.muster/workspaces/<stack>/<workspace>` to prevent name collisions between stacks; `muster remove` now accepts `--stack` to disambiguate if needed

---

## v0.3.1 - 2026-07-23

### Fixed
- Workspace directory structure: workspaces now created under `~/.muster/workspaces/` instead of directly in `~/.muster/`

---

## v0.3.0 - 2026-07-22

### Added
- `muster init` command to create example configuration file at `~/.config/muster/config.json`
- Embedded config template in binary for portability

---

## v0.2.1 - 2026-07-21

### Fixed
- Release pipeline: ignore dynamically generated `.release-notes.md` file to prevent "git dirty state" error

---

## v0.2.0 - 2026-07-21

### Added
- Project-level instructions (AGENTS.md and CLAUDE.md)

### Fixed
- Release workflow now creates git tag before GoReleaser runs (handles first release)
- Scorecard workflow now has correct permissions (security-events: write, id-token: write)

---

## v0.1.0 - 2026-07-21

### Added
- Multi-repo workspace orchestration
- Coordinated branching across repositories
- Workspace configuration generation
- Stack management for repository collections
- Claude Code integration with `muster launch` command
