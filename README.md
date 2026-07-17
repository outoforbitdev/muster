# Muster

Workspace orchestrator for Claude Code. Create and manage multi-repo workspaces with coordinated branch checkout and automatic workspace configuration.

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

Create `~/.config/muster/config.json`:

```json
{
  "stacks": {
    "full-stack": {
      "repos": [
        {
          "url": "https://github.com/yourorg/api",
          "templateBranchSyntax": "main",
          "description": "Backend API service"
        },
        {
          "url": "https://github.com/yourorg/web",
          "templateBranchSyntax": "main",
          "description": "Next.js frontend"
        },
        {
          "url": "https://github.com/yourorg/shared-types",
          "templateBranchSyntax": "feature-{workspace}",
          "description": "Shared TypeScript types",
          "directory": "types"
        }
      ],
      "description": "api is the backend service. web is the Next.js frontend consuming api. shared-types holds TS types shared between both."
    }
  },
  "defaults": {
    "checkoutBranchOnLaunch": true,
    "templateBranchSyntax": "feature-{workspace}"
  }
}
```

See `config.example.json` for a complete example.

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

### Branch Checkout Precedence

1. CLI `--branch` flag → uses this branch for all repos
2. CLI `--no-branch` flag → skips checkout entirely
3. Per-repo `templateBranchSyntax` in config
4. Global `defaults.templateBranchSyntax` in config
5. Git default branch (after clone)

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
