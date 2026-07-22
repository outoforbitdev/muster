# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

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
