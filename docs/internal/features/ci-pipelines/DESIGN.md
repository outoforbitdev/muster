# CI Pipeline Design for Muster

**Date:** 2026-07-21  
**Author:** Design Review Process  
**Status:** In Review

## Overview

This document specifies the CI/CD pipeline architecture for the Muster CLI tool. The pipeline consists of 7 workflows that provide automated testing, code quality checks, dependency management, security scanning, release management, and task tracking.

## Project Context

Muster is a Go CLI tool (v1.26.5+) for managing multi-repo workspaces. Current development practices:
- Testing: `go test ./... -v`
- Linting: `golangci-lint`
- Code formatting: `go fmt`
- Build: `go build -o bin/muster ./cmd/muster`
- Conventional commits in use
- Semantic versioning target
- No CI infrastructure currently in place

## Pipeline Architecture

### Workflow Overview

```
Pull Request Flow:
  ├─ test.yml (custom) ........... Go tests + linting (required gate)
  └─ check-tasks.yml (reusable) .. Task completion check (required gate)

Main Branch Flow (after merge):
  └─ release.yml (reusable) ....... Create release + build binaries (GoReleaser)

Scheduled:
  ├─ dependabot.yml (config) ..... Weekly dependency updates
  ├─ scorecard.yml (reusable) .... OSSF security scanning
  └─ labels.yml (reusable) ....... Label management
```

### 1. Test Workflow (`test.yml`)

**Trigger:** On every pull request  
**Type:** Custom (Go-specific)  
**Status:** Required gate for merge

**Purpose:** Validates code quality and correctness.

**Steps:**
1. Checkout code
2. Setup Go (v1.26.5+)
3. Run golangci-lint
4. Run go test ./... -v

**Behavior:**
- Fails PR if linting or tests fail
- Reports results to PR status check

### 2. Check Tasks Workflow (`check-tasks.yml`)

**Trigger:** On every pull request  
**Type:** Reusable (from `outoforbitdev/reusable-workflows-library`)  
**Status:** Required gate for merge

**Purpose:** Enforces task completion before merge.

**Behavior:**
- Scans PR for incomplete tasks (TODOs, unchecked checkboxes)
- Blocks merge if tasks remain incomplete
- Uses organization's standard task-checking action

### 3. Release Workflow (`release.yml`)

**Trigger:** On push to main branch  
**Type:** Reusable (from `outoforbitdev/reusable-workflows-library`)  
**Status:** Manual trigger or automatic based on CHANGELOG

**Purpose:** Creates GitHub releases with semantic versioning and pre-built binaries.

**Mechanism:**
- Watches for CHANGELOG.md updates on push to main
- `action-release-changelog` derives version from CHANGELOG
- Creates git tag with semantic version (v1.0.0, v1.1.0, etc.)
- Creates GitHub release
- Triggers GoReleaser to build binaries

**Build Targets (GoReleaser):**
- macOS: amd64, arm64 (Apple Silicon)
- Linux: amd64, arm64
- Windows: amd64

**Release Assets:**
- Compiled binaries for each platform
- Checksums (SHA256)
- Archives (.tar.gz, .zip)

**Configuration:** Requires `.goreleaser.yml` in repo root

### 4. GoReleaser Configuration (`.goreleaser.yml`)

**Purpose:** Defines binary build and release process.

**Key Settings:**
- Build targets: macOS (amd64/arm64), Linux (amd64/arm64), Windows (amd64)
- Output: Binaries, archives, checksums
- Code signing: None (can be added later if needed)

**Assets Generated:**
- muster-darwin-amd64, muster-darwin-arm64
- muster-linux-amd64, muster-linux-arm64
- muster-windows-amd64.exe
- Compressed archives (.tar.gz for Unix, .zip for Windows)
- SHA256SUMS checksum file

### 5. Dependabot Configuration (`.github/dependabot.yml`)

**Purpose:** Automated dependency update management.

**Settings:**
- Ecosystem: Go (go.mod)
- Schedule: Weekly
- Update strategy: Create PRs for all updates (patch, minor, major)
- No auto-merge

**Behavior:**
- Creates separate PR for each dependency update
- Runs existing test and check-tasks workflows on each PR
- Developers review and merge manually

### 6. OSSF Scorecard Workflow (`scorecard.yml`)

**Trigger:** On push to main + scheduled weekly  
**Type:** Reusable (from `outoforbitdev/reusable-workflows-library`)

**Purpose:** Security and best-practices scanning.

**Coverage:**
- Dependency management checks
- Branch protection rules
- Token permissions
- Security policy presence
- License presence
- More (full OSSF scorecard checks)

### 7. Label Manager Workflow (`labels.yml`)

**Trigger:** On repository events  
**Type:** Reusable (from `outoforbitdev/reusable-workflows-library`)

**Purpose:** Auto-manage issue and PR labels.

**Behavior:**
- Applies labels to issues/PRs based on configured rules
- Enforces consistent labeling across repository
- Configured via organization standards

## README Updates

The README.md will be updated to include status badges in a new section at the top, below the project title:

**Badges to add:**
1. Test workflow badge - links to test workflow results
2. Release workflow badge - links to release workflow results  
3. OpenSSF Scorecard badge - links to https://securityscorecards.dev/viewer/?uri=github.com/outoforbitdev/muster
4. Latest release badge - shows latest version
5. Open issues badge (optional)

These badges follow the pattern used in `outoforbitdev/library-galaxy-map` repository and provide at-a-glance status on project health and release versions.

## File Structure

```
muster/
├── .github/
│   ├── workflows/
│   │   ├── test.yml ..................... PR test gating
│   │   ├── check-tasks.yml .............. PR task completion check
│   │   ├── release.yml .................. Release + binary builds
│   │   ├── scorecard.yml ................ OSSF security scanning
│   │   └── labels.yml ................... Label management
│   └── dependabot.yml ................... Dependency update config
├── .goreleaser.yml ...................... Binary build configuration
├── CHANGELOG.md ......................... For version derivation (new)
├── README.md (updated) .................. Add status badges section
└── docs/
    └── superpowers/
        └── specs/
            └── 2026-07-21-ci-pipeline-design.md (this file)
```

## CHANGELOG.md Requirement

The release workflow depends on a CHANGELOG.md file to determine version numbers. Format:
- Top entry should reflect changes since last release
- Version derived from semantic versioning rules
- Examples:
  - `## [1.0.0] - 2026-07-21` for initial release
  - `## [1.1.0] - 2026-07-25` for minor version bump
  - `## [1.0.1] - 2026-07-22` for patch bump

**Note:** CHANGELOG.md will need to be created as part of implementation.

## Implementation Order

1. Create `.github/workflows/` directory structure
2. Write test.yml (custom Go workflow)
3. Write .goreleaser.yml (Go binary config)
4. Create wrapper workflows (release.yml, check-tasks.yml, scorecard.yml, labels.yml)
5. Configure .github/dependabot.yml
6. Create CHANGELOG.md template
7. Test all workflows (will require push to main branch)

## Dependencies & Prerequisites

**Required:**
- .github/workflows/ directory (new)
- CHANGELOG.md file (new)
- .goreleaser.yml configuration (new)
- GitHub token with contents:write permission (standard)

**Assumptions:**
- Organization's reusable workflows accessible at `outoforbitdev/reusable-workflows-library`
- GoReleaser action available in GitHub Actions marketplace
- Semantic versioning discipline maintained in CHANGELOG updates

## Success Criteria

1. ✅ Test workflow runs on every PR, gates on linting + tests
2. ✅ Check-tasks workflow prevents merge of incomplete tasks
3. ✅ Release workflow creates GitHub releases + binaries on main
4. ✅ GoReleaser builds for all 5 platform/architecture combinations
5. ✅ Dependabot creates weekly dependency update PRs
6. ✅ OSSF scorecard runs and reports security status
7. ✅ Label manager auto-applies labels
8. ✅ All workflows properly handle failures and report status
9. ✅ README updated with status badges (test, release, scorecard, latest release, open issues)

## Future Enhancements (Out of Scope)

- Code signing of binaries
- Container image builds (Docker)
- Automated changelog generation (currently manual)
- Workflow notifications (Slack, email)
- Test coverage threshold enforcement
- Automated backport workflows
