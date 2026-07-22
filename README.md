# Muster

Workspace orchestrator for Claude Code. Create and manage multi-repo workspaces with coordinated branch checkout and automatic workspace configuration.

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

## Features

- **Multi-repo workspaces**: Clone multiple repositories into a single workspace
- **Coordinated branching**: Automatically check out branches across all repos with template substitution
- **Workspace configuration**: Auto-generate `CLAUDE.md` with workspace metadata
- **Stack management**: Define reusable repository collections in config
- **Claude Code integration**: Seamlessly launch Claude Code with workspace context

## Installation

### From Source

```bash
go build -o /usr/local/bin/muster ./cmd/muster
```

### Requirements

- Go 1.19 or later
- Git
- Claude Code (for `muster launch` integration)

## Setup

### Configuration File

Initialize your muster configuration:

```bash
muster init
```

This creates a new config file at `~/.config/muster/config.json` with example stacks and defaults. Then edit the file to customize your repositories and stacks.

See `config.example.json` for the complete example configuration structure.

## Usage

### Launch a Workspace

Launch an existing workspace or create a new one from a stack:

```bash
muster launch my-workspace --stack full-stack
```

For new workspaces, this will:
1. Create `~/.muster/my-workspace/`
2. Clone all repos from the stack
3. Checkout branches (with `{workspace}` substitution)
4. Generate `CLAUDE.md`
5. Launch Claude Code

For existing workspaces, it opens them directly in Claude Code.

### Launch with Specific Branch

Override branch checkout for all repos:

```bash
muster launch my-workspace --stack full-stack --branch main
```

Template substitution still applies:
- `--branch "feature-{workspace}"` → checks out `feature-my-workspace`

### Launch Without Branch Checkout

Skip automatic branch checkout:

```bash
muster launch my-workspace --stack full-stack --no-branch
```

### Launch Without Agent

Skip launching the agent (Claude Code by default):

```bash
muster launch my-workspace --stack full-stack --no-agent
```

### Launch With Editor

Launch the editor in addition to the agent:

```bash
muster launch my-workspace --stack full-stack --editor
```

### Launch With Custom Agent/Editor

Configure custom commands in your config file, then:

```bash
muster launch my-workspace --stack full-stack --agent --editor
```

Both `--agent` and `--editor` can be used together, and both respect your configured commands and defaults.

### Add Individual Repos

Mix stack repos with explicit URLs:

```bash
muster launch my-workspace --stack full-stack --repo https://github.com/yourorg/docs
```

Repeatable:
```bash
muster launch my-workspace \
  --repo https://github.com/yourorg/api \
  --repo https://github.com/yourorg/web
```

### Remove a Workspace

Delete a workspace with confirmation:

```bash
muster remove my-workspace
```

Skip confirmation:

```bash
muster remove my-workspace --yes
```

## Configuration Reference

### Config Structure

- **stacks**: Named collections of repositories
  - **repos**: Array of repository definitions
    - `url` (required): Git URL (HTTPS, HTTP, or SSH)
    - `templateBranchSyntax` (optional): Branch template with `{workspace}` placeholder
    - `description` (optional): Human-readable description
    - `directory` (optional): Custom directory name (default: inferred from URL)
  - `description` (optional): Stack description (included in generated CLAUDE.md)

- **defaults**: Global settings
  - `checkoutBranchOnLaunch` (boolean, default `true`): Enable automatic branch checkout
  - `templateBranchSyntax` (string): Default branch template (e.g., `"feature-{workspace}"`)
  - `agentCommand` (string, default `"claude --name {workspace}"`): Command to launch agent
  - `editorCommand` (string, default `"code {workspaceDirectory}"`): Command to launch editor
  - `launchAgent` (boolean, default `true`): Launch agent by default when no flags are set
  - `launchEditor` (boolean, default `false`): Launch editor by default when no flags are set

**Template Variables:**
- `{workspace}` — The workspace name (e.g., `"my-workspace"`)
- `{workspaceDirectory}` — The full path to the workspace (e.g., `/home/user/.muster/my-workspace`)

### Branch Checkout Precedence

1. CLI `--branch` flag → uses this branch for all repos
2. CLI `--no-branch` flag → skips checkout entirely
3. Per-repo `templateBranchSyntax` in config
4. Global `defaults.templateBranchSyntax` in config
5. Git default branch (after clone)

### Agent and Editor Launch Precedence

**Agent Launch:**
1. CLI `--no-agent` flag → do not launch
2. CLI `--agent` flag → launch
3. `defaults.launchAgent` config (default `true`) → launch if true
4. If not launched, no agent command is run

**Editor Launch:**
1. CLI `--no-editor` flag → do not launch
2. CLI `--editor` flag → launch
3. `defaults.launchEditor` config (default `false`) → launch if true
4. If not launched, no editor command is run

**Command Template Variables:**

Commands support two template variables:
- `{workspace}` — Replaced with the workspace name (e.g., `my-workspace`)
- `{workspaceDirectory}` — Replaced with the full workspace path (e.g., `/home/user/.muster/my-workspace`)

Examples:
- `claude --name {workspace}` → `claude --name my-workspace`
- `code {workspaceDirectory}` → `code /home/user/.muster/my-workspace`

## Workspace Layout

```
~/.muster/my-workspace/
├── CLAUDE.md              # Auto-generated workspace documentation
├── api/                   # Cloned repo (from URL)
├── web/                   # Cloned repo
└── types/                 # Custom directory name (from config)
```

## Development

### Build

```bash
go build -o bin/muster ./cmd/muster
```

### Test

```bash
go test ./...
```

### Lint

```bash
go fmt ./...
golangci-lint run ./...
```

## Design

See `docs/internal/DESIGN.md` for detailed architecture and implementation decisions.
