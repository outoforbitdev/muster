# Changelog

All notable changes to this project will be documented in this file.

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Features
- `muster clean` command to find and delete workspaces that are safe to remove (no uncommitted changes, and either no divergence from `main`/`master` or divergent commits already merged)
- `muster list` (or `muster ls`) command to list workspaces or stacks, with `--all` for a table view of stack/repo details

### Fixed
- `muster launch` now sets upstream tracking (e.g. `origin/main`) on branches it creates, so they no longer start out untracked

---

## [0.4.0] - 2026-09-25

### Changed
- `muster launch --stack` now accepts a single stack name instead of multiple repeated flags

### Fixed
- Workspaces created from a stack are now nested under `~/.muster/workspaces/<stack>/<workspace>` to prevent name collisions between stacks; `muster remove` now accepts `--stack` to disambiguate if needed

---

## [0.3.1] - 2026-07-23

### Fixed
- Workspace directory structure: workspaces now created under `~/.muster/workspaces/` instead of directly in `~/.muster/`

---

## [0.3.0] - 2026-07-22

### Added
- `muster init` command to create example configuration file at `~/.config/muster/config.json`
- Embedded config template in binary for portability

---

## [0.2.1] - 2026-07-21

### Fixed
- Release pipeline: ignore dynamically generated `.release-notes.md` file to prevent "git dirty state" error

---

## [0.2.0] - 2026-07-21

### Added
- Project-level instructions (AGENTS.md and CLAUDE.md)

### Fixed
- Release workflow now creates git tag before GoReleaser runs (handles first release)
- Scorecard workflow now has correct permissions (security-events: write, id-token: write)

---

## [0.1.0] - 2026-07-21

### Added
- Multi-repo workspace orchestration
- Coordinated branching across repositories
- Workspace configuration generation
- Stack management for repository collections
- Claude Code integration with `muster launch` command
