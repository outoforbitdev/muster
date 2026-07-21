# CI Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Set up 7 CI/CD workflows for automated testing, linting, dependency management, security scanning, release management, and task tracking in the Muster CLI repository.

**Architecture:** Create GitHub Actions workflows that enforce code quality on PRs (test + linting + task completion checks), manage dependencies via Dependabot (weekly), and automate releases via semantic versioning with GoReleaser to build cross-platform binaries. Reuse organization workflows where available to reduce maintenance.

**Tech Stack:** GitHub Actions, Go 1.26.5+, golangci-lint, GoReleaser, Dependabot, OSSF Scorecard

## Global Constraints

- Go version: 1.26.5+
- Build command: `go build -o bin/muster ./cmd/muster`
- Test command: `go test ./... -v`
- Lint command: `golangci-lint run ./...`
- Versioning: Semantic versioning (v1.0.0 format)
- Conventional commits required
- GoReleaser builds for: macOS (amd64/arm64), Linux (amd64/arm64), Windows (amd64)

---

### Task 1: Create .github Directory Structure

**Files:**
- Create: `.github/workflows/` (directory)
- Create: `.github/dependabot.yml` (placeholder, will be filled in later task)

**Interfaces:**
- Produces: Directory structure for all subsequent workflow files

- [ ] **Step 1: Create .github/workflows directory**

```bash
mkdir -p .github/workflows
```

- [ ] **Step 2: Verify directory exists**

```bash
ls -la .github/
```

Expected output should show `workflows` directory.

- [ ] **Step 3: Commit**

```bash
git add .github/
git commit -m "ci: create GitHub Actions directory structure

Create .github/workflows directory to hold all CI workflow definitions.

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
```

---

### Task 2: Create test.yml Workflow

**Files:**
- Create: `.github/workflows/test.yml`

**Interfaces:**
- Consumes: Nothing (independent)
- Produces: GitHub Actions workflow that runs on every PR

**Purpose:** Run Go tests and linting on every pull request to gate merges.

- [ ] **Step 1: Write test.yml**

Create `.github/workflows/test.yml` with the following content:

```yaml
name: Test

permissions:
  contents: read

on:
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    name: Test and Lint

    steps:
      - name: Checkout
        uses: actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5 # v4.3.1

      - name: Set up Go
        uses: actions/setup-go@41dfa10bad2ba9c5c13ce3406b6e8a7f49d11921 # v5.1.0
        with:
          go-version: '1.26.5'

      - name: Run linter
        uses: golangci/golangci-lint-action@adc3e2e007b5e92b5dc23f1ca4f6e538192ddf89 # v6.2.1
        with:
          version: latest
          args: --timeout=5m

      - name: Run tests
        run: go test ./... -v
```

- [ ] **Step 2: Verify workflow syntax**

```bash
cat .github/workflows/test.yml
```

Expected: YAML file with workflow definition, no errors.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/test.yml
git commit -m "ci: add test and lint workflow

Run Go tests and golangci-lint on every pull request to gate merges.
Triggers automated code quality checks before code can be merged.

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
```

---

### Task 3: Create .goreleaser.yml Configuration

**Files:**
- Create: `.goreleaser.yml`

**Interfaces:**
- Consumes: Nothing (independent)
- Produces: GoReleaser configuration for building cross-platform binaries

**Purpose:** Configure GoReleaser to build Muster CLI binaries for macOS, Linux, and Windows across multiple architectures.

- [ ] **Step 1: Write .goreleaser.yml**

Create `.goreleaser.yml` with the following content:

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: default
    main: ./cmd/muster
    binary: muster
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - id: default
    format_overrides:
      - goos: windows
        format: zip
    default_libname: muster
    default_binary_name: muster

checksum:
  name_template: 'SHA256SUMS'
  algorithm: sha256

release:
  draft: false
  prerelease: false
```

- [ ] **Step 2: Verify configuration syntax**

```bash
cat .goreleaser.yml
```

Expected: YAML file with build, archive, checksum, and release sections.

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.yml
git commit -m "ci: add GoReleaser configuration

Configure GoReleaser to build Muster CLI binaries for:
- macOS: amd64, arm64 (Apple Silicon)
- Linux: amd64, arm64
- Windows: amd64

Generates archives and SHA256 checksums for each platform.

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
```

---

### Task 4: Create CHANGELOG.md

**Files:**
- Create: `CHANGELOG.md`

**Interfaces:**
- Consumes: Nothing (independent)
- Produces: Changelog file used by release workflow to determine version numbers

**Purpose:** Create initial CHANGELOG.md with semantic versioning structure that the release workflow will parse to determine version bumps.

- [ ] **Step 1: Write CHANGELOG.md**

Create `CHANGELOG.md` with the following content:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial CI/CD pipeline setup with GitHub Actions

### Changed

### Deprecated

### Removed

### Fixed

### Security

---

## [1.0.0] - 2026-07-21

### Added
- Multi-repo workspace orchestration
- Coordinated branching across repositories
- Workspace configuration generation
- Stack management for repository collections
- Claude Code integration with `muster launch` command
```

- [ ] **Step 2: Verify changelog structure**

```bash
head -30 CHANGELOG.md
```

Expected: Markdown file with semantic versioning format.

- [ ] **Step 3: Commit**

```bash
git add CHANGELOG.md
git commit -m "ci: add CHANGELOG.md

Create initial changelog following Keep a Changelog format.
Release workflow will parse version entries to determine semantic version bumps.

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
```

---

### Task 5: Create check-tasks.yml Wrapper Workflow

**Files:**
- Create: `.github/workflows/check-tasks.yml`

**Interfaces:**
- Consumes: Reusable workflow from `outoforbitdev/reusable-workflows-library`
- Produces: GitHub Actions workflow that checks task completion on PRs

**Purpose:** Create wrapper workflow that calls the organization's reusable check-tasks workflow to enforce task completion before merge.

- [ ] **Step 1: Write check-tasks.yml**

Create `.github/workflows/check-tasks.yml` with the following content:

```yaml
name: Check Tasks

permissions: read-all

on: pull_request

jobs:
  check-tasks:
    # yamllint disable-line rule:line-length
    uses: outoforbitdev/reusable-workflows-library/.github/workflows/check-tasks.yml@main
```

- [ ] **Step 2: Verify workflow file**

```bash
cat .github/workflows/check-tasks.yml
```

Expected: Minimal YAML file that calls the reusable workflow.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/check-tasks.yml
git commit -m "ci: add check-tasks workflow

Add wrapper workflow that calls the organization's reusable check-tasks workflow.
Ensures all tasks/TODOs in PRs are completed before merge is allowed.

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
```

---

### Task 6: Create release.yml Wrapper Workflow

**Files:**
- Create: `.github/workflows/release.yml`

**Interfaces:**
- Consumes: Reusable workflow from `outoforbitdev/reusable-workflows-library`, `.goreleaser.yml`, `CHANGELOG.md`
- Produces: GitHub Actions workflow that creates releases with semantic versioning

**Purpose:** Create wrapper workflow that calls the organization's reusable release workflow. This workflow watches for CHANGELOG.md updates and automatically creates releases with semantic versioning.

- [ ] **Step 1: Write release.yml**

Create `.github/workflows/release.yml` with the following content:

```yaml
name: Release

permissions: read-all

on:
  push:
    branches: [main]

jobs:
  release:
    # yamllint disable-line rule:line-length
    uses: outoforbitdev/reusable-workflows-library/.github/workflows/release.yml@main
    permissions:
      contents: write
```

- [ ] **Step 2: Verify workflow file**

```bash
cat .github/workflows/release.yml
```

Expected: YAML file that calls the reusable release workflow on main branch pushes.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: add release workflow

Add wrapper workflow that calls the organization's reusable release workflow.
Automatically creates GitHub releases with semantic versioning based on CHANGELOG.md.

The workflow:
- Watches for CHANGELOG.md updates on push to main
- Determines version bump via action-release-changelog
- Triggers GoReleaser to build cross-platform binaries
- Creates GitHub release with built artifacts

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
```

---

### Task 7: Create scorecard.yml Wrapper Workflow

**Files:**
- Create: `.github/workflows/scorecard.yml`

**Interfaces:**
- Consumes: Reusable workflow from `outoforbitdev/reusable-workflows-library`
- Produces: GitHub Actions workflow that runs OSSF security scanning

**Purpose:** Create wrapper workflow that calls the organization's reusable OSSF Scorecard workflow for security and best-practices scanning.

- [ ] **Step 1: Write scorecard.yml**

Create `.github/workflows/scorecard.yml` with the following content:

```yaml
name: OpenSSF Scorecard

permissions: read-all

on:
  push:
    branches: [main]
  schedule:
    - cron: '0 0 * * 0'

jobs:
  scorecard:
    # yamllint disable-line rule:line-length
    uses: outoforbitdev/reusable-workflows-library/.github/workflows/scorecard.yml@main
```

- [ ] **Step 2: Verify workflow file**

```bash
cat .github/workflows/scorecard.yml
```

Expected: YAML file with push trigger on main and weekly schedule.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/scorecard.yml
git commit -m "ci: add OpenSSF Scorecard workflow

Add wrapper workflow that calls the organization's reusable OSSF Scorecard workflow.
Runs security scanning and best-practices checks on push to main and weekly schedule.

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
```

---

### Task 8: Create labels.yml Wrapper Workflow

**Files:**
- Create: `.github/workflows/labels.yml`

**Interfaces:**
- Consumes: Reusable workflow from `outoforbitdev/reusable-workflows-library`
- Produces: GitHub Actions workflow that manages issue/PR labels

**Purpose:** Create wrapper workflow that calls the organization's reusable label manager workflow for automatic label application.

- [ ] **Step 1: Write labels.yml**

Create `.github/workflows/labels.yml` with the following content:

```yaml
name: Label Manager

permissions: read-all

on:
  issues:
  pull_request:
  pull_request_target:

jobs:
  label-manager:
    # yamllint disable-line rule:line-length
    uses: outoforbitdev/reusable-workflows-library/.github/workflows/label-manager.yml@main
    permissions:
      contents: read
      issues: write
      pull-requests: write
```

- [ ] **Step 2: Verify workflow file**

```bash
cat .github/workflows/labels.yml
```

Expected: YAML file with triggers for issues and PRs.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/labels.yml
git commit -m "ci: add label manager workflow

Add wrapper workflow that calls the organization's reusable label manager workflow.
Automatically applies labels to issues and pull requests based on configured rules.

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
```

---

### Task 9: Create dependabot.yml Configuration

**Files:**
- Create: `.github/dependabot.yml`

**Interfaces:**
- Consumes: Nothing (independent)
- Produces: Dependabot configuration for weekly Go dependency updates

**Purpose:** Configure Dependabot to check for Go dependency updates weekly and create PRs for all updates (no auto-merge).

- [ ] **Step 1: Write dependabot.yml**

Create `.github/dependabot.yml` with the following content:

```yaml
version: 2

updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "03:00"
    open-pull-requests-limit: 10
    reviewers:
      - "jaymirecki"
```

- [ ] **Step 2: Verify configuration**

```bash
cat .github/dependabot.yml
```

Expected: YAML file with gomod ecosystem, weekly schedule, and PR settings.

- [ ] **Step 3: Commit**

```bash
git add .github/dependabot.yml
git commit -m "ci: add Dependabot configuration

Configure Dependabot to check for Go dependency updates weekly.
Creates pull requests for all updates (patch, minor, major) for manual review.

Schedule: Every Monday at 03:00 UTC
Limit: Up to 10 open PRs at once

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
```

---

### Task 10: Update README with Status Badges

**Files:**
- Modify: `README.md` (add badges section after title)

**Interfaces:**
- Consumes: All workflows (test.yml, release.yml, scorecard.yml)
- Produces: Updated README with status badges

**Purpose:** Add status badges to README showing test results, release status, security scorecard, latest version, and open issues.

- [ ] **Step 1: Read current README**

```bash
head -20 README.md
```

Note the current structure (should start with `# Muster`).

- [ ] **Step 2: Update README with badges section**

Open `README.md` and add the following after the `# Muster` title line and before the ## Features section:

```markdown
<p>
  <a href="https://github.com/outoforbitdev/muster/actions?query=workflow%3ATest">
    <img alt="Test workflow" src="https://github.com/outoforbitdev/muster/workflows/Test/badge.svg">
  </a>
  <a href="https://github.com/outoforbitdev/muster/actions?query=workflow%3ARelease">
    <img alt="Release workflow" src="https://github.com/outoforbitdev/muster/workflows/Release/badge.svg">
  </a>
  <a href="https://securityscorecards.dev/viewer/?uri=github.com/outoforbitdev/muster">
    <img alt="OpenSSF Scorecard" src="https://api.securityscorecards.dev/projects/github.com/outoforbitdev/muster/badge">
  </a>
  <a href="https://github.com/outoforbitdev/muster/releases/latest">
    <img alt="Latest release" src="https://img.shields.io/github/v/release/outoforbitdev/muster?logo=github">
  </a>
  <a href="https://github.com/outoforbitdev/muster/issues">
    <img alt="Open issues" src="https://img.shields.io/github/issues/outoforbitdev/muster?logo=github">
  </a>
</p>
```

The section should go after line 3 (after `Workspace orchestrator for Claude Code...`) and before `## Features`.

- [ ] **Step 3: Verify README changes**

```bash
head -30 README.md
```

Expected: Badge section visible with links and shield.io badge URLs.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: add CI status badges to README

Add status badges showing:
- Test workflow status
- Release workflow status
- OpenSSF Scorecard security rating
- Latest GitHub release version
- Open issues count

Badges provide at-a-glance project health status.

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
```

---

## Implementation Notes

### Testing Workflows Locally

- `test.yml`, `check-tasks.yml`, `scorecard.yml`, and `labels.yml` require pushing to GitHub to test
- `.goreleaser.yml` and `.dependabot.yml` cannot be easily tested locally
- `CHANGELOG.md` format can be verified locally by reading the file

### First Release

After merging all workflows to main:
1. Update `CHANGELOG.md` with actual v1.0.0 entry
2. Push to main
3. The release workflow will automatically detect the version and create a GitHub release
4. GoReleaser will build binaries for all 5 platform/architecture combinations

### Dependabot First Run

Dependabot may take a few minutes to scan dependencies after `.github/dependabot.yml` is merged. Check the "Dependabot" tab in repository settings to verify configuration.

## Verification Checklist

After all tasks are complete and merged:

- [ ] All 5 workflow files exist in `.github/workflows/`
- [ ] `.goreleaser.yml` exists in repository root
- [ ] `CHANGELOG.md` exists in repository root
- [ ] `.github/dependabot.yml` exists
- [ ] README has status badges section
- [ ] Push a test commit to verify workflows trigger
- [ ] Create a test PR to verify test.yml and check-tasks.yml run
