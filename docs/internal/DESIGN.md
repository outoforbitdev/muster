# Muster Design Document

## 1. CLI Interface

### Command: `muster launch`

```
muster launch <workspace> [--stack <stack>] [--repo <repo>]... [--branch <branch>] [--no-branch] [--agent] [--no-agent] [--editor] [--no-editor]
```

**Arguments:**
- `<workspace>` — Name of the workspace to create or launch. Directory: `~/.workspaces/<workspace>/`

**Flags:**
- `--stack <stack>` — Load repos from a named stack in config. Optional, repeatable if multiple stacks needed.
- `--repo <repo>` — Add individual repos (in addition to/instead of stacks). Must be a full git URL (SSH or HTTPS). Optional, repeatable.
- `--branch <branch>` — Default branch to check out for all repos. Can be a template like `feature/{workspace}` or literal branch name. Optional.
- `--no-branch` — Skip branch checkout; use whatever branch is already checked out (or default after clone). Optional.
- `--agent` — Launch the agent command after creating/opening the workspace. Optional.
- `--no-agent` — Do not launch the agent command. Optional.
- `--editor` — Launch the editor command after creating/opening the workspace. Optional.
- `--no-editor` — Do not launch the editor command. Optional.

**Behavior:**

**New workspace (doesn't exist):**
1. Create `~/.workspaces/<workspace>/` directory.
2. Clone all repos (from stacks + individual `--repo` flags) into subdirectories.
3. Checkout branches if applicable (based on config defaults or `--branch`/`--no-branch` flags).
4. Generate `CLAUDE.md` at workspace root from config stack descriptions.
5. Launch Claude Code for the user with `claude launch --name <workspace>` to name the session for the workspace.

**Existing workspace:**
1. Just launch Claude Code in `~/.workspaces/<workspace>/` (or print path and instructions for user to launch).
2. Do not re-clone, re-checkout, or overwrite existing repos.

### Command: `muster remove`

```
muster remove <workspace> [-y|--yes]
```

**Arguments:**
- `<workspace>` — Name of the workspace to remove. Deletes `~/.workspaces/<workspace>/` directory.

**Flags:**
- `-y`, `--yes` — Skip confirmation prompt and immediately delete the workspace. Optional.

**Behavior:**
1. If `-y`/`--yes` is not set: prompt user for confirmation before deletion.
2. Delete the workspace directory and all its contents.
3. Print confirmation message after deletion.

---

## 2. Configuration

**File location:** `~/.config/muster/config.json`

**Structure:**

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
          "templateBranchSyntax": "{workspace}",
          "description": "Shared TypeScript types",
          "directory": "types"
        }
      ],
      "description": "api is the backend service. web is the Next.js frontend consuming api. shared-types holds TS types shared between both."
    }
  },
  "defaults": {
    "checkoutBranchOnLaunch": true,
    "templateBranchSyntax": "{workspace}",
    "agentCommand": "claude --name {workspace}",
    "editorCommand": "code {workspaceDirectory}",
    "launchAgent": true,
    "launchEditor": false
  }
}
```

**Explanation:**

- **stacks**: Named collections of repos. Each stack defines:
  - `repos`: Array of repo objects, each with:
    - `url`: Git repo URL (full SSH or HTTPS URL only, e.g., `git@github.com:yourorg/api.git` or `https://github.com/yourorg/api`).
    - `templateBranchSyntax`: *(optional)* Branch template to checkout. Can use template syntax like `{workspace}` which gets replaced with the workspace name. If omitted, uses default after clone.
    - `description`: *(optional)* Human-readable text describing this repo.
    - `directory`: *(optional)* Custom directory name for the cloned repo. If omitted, uses git clone's default (last path component without `.git`).
  - `description`: Human-readable text describing the stack (templated into CLAUDE.md).

- **defaults**: Global defaults:
  - `checkoutBranchOnLaunch`: Boolean flag (default `true`) to enable automatic branch checkout on launch.
  - `templateBranchSyntax`: Template syntax for branch names, e.g., `{workspace}` or `feature-{workspace}`.
  - `agentCommand`: *(optional)* Command template to launch agent (default `"claude --name {workspace}"`). Supports `{workspace}` and `{workspaceDirectory}` substitution.
  - `editorCommand`: *(optional)* Command template to launch editor (default `"code {workspaceDirectory}"`). Supports `{workspace}` and `{workspaceDirectory}` substitution.
  - `launchAgent`: *(optional)* Boolean flag (default `true`) to launch agent by default when no CLI flags are set.
  - `launchEditor`: *(optional)* Boolean flag (default `false`) to launch editor by default when no CLI flags are set.

**CLI flag precedence:**
1. If both `--branch` and `--no-branch` are specified, error out (mutually exclusive).
2. If `--no-branch` → skip all checkouts.
3. If `--branch <branch>` → override all repos and use this branch (template syntax still applies).
4. If no CLI flags: check `checkoutBranchOnLaunch` setting. If `true`, use per-repo `templateBranchSyntax` or global `templateBranchSyntax`. If `false`, skip checkout.
5. If no branch specified anywhere: use default branch after clone.

---

## 3. File Structure

```
muster/
├── cmd/
│   └── main.go                 # CLI entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Config loading, schema, validation
│   ├── workspace/
│   │   └── workspace.go        # Workspace creation, repo cloning, checkout
│   └── claude/
│       └── claude.go           # CLAUDE.md generation
├── go.mod
├── go.sum
├── .goreleaser.yaml            # GoReleaser config
├── Justfile                    # Local build/test/lint
├── README.md
└── .github/
    └── workflows/
        └── release.yml         # GitHub Actions: run GoReleaser on tag
```

---

## 4. Key Implementation Details

### Config Loading
- Load from `~/.config/muster/config.json`.
- If not found, use sensible defaults (empty stacks, warn user, or error—TBD).
- Validate that all referenced repos are parseable URLs.

### Workspace Creation
- Use `os/exec` to shell out to `git clone` for each repo.
- Clone into `~/.workspaces/<workspace>/<repo-name>/`.
- Handle checkout logic (in order):
  1. If `--no-branch` is set: skip checkout.
  2. If `--branch` is set: use that branch for all repos (with template substitution).
  3. If `checkoutBranchOnLaunch` is `false`: skip checkout.
  4. If `checkoutBranchOnLaunch` is `true`: use per-repo `templateBranchSyntax` (or fall back to global `templateBranchSyntax`, or if not set, use `feature/{workspace}`).
  5. If no branch specified anywhere: use default after clone.

### Branch Template Substitution
- Replace `{workspace}` with the workspace name.
- E.g., if workspace is `my-feature` and branch template is `feature-{workspace}`, checkout `feature-my-feature`.

### CLAUDE.md Generation
- Use a static template with placeholder substitution.
- Template format:
  ```
  # Workspace: {workspace}
  
  {description}
  
  ## Repos
  {repos}
  ```
- `{workspace}` is replaced with the workspace name.
- `{description}` is replaced with the stack description from config.
- `{repos}` is an auto-generated list of repos with their cloned paths (e.g., `- api: ~/.workspaces/<workspace>/api`).

### Launching Agent and Editor
- After workspace creation/opening, optionally launch the configured agent and editor commands.
- Editor launches before agent (to avoid agent blocking editor launch).
- Commands are executed via the shell (`sh -c`) with template substitution:
  - `{workspace}` → workspace name (e.g., `my-workspace`)
  - `{workspaceDirectory}` → full workspace path (e.g., `/home/user/.muster/my-workspace`)
- Both agent and editor are optional; either can be disabled via config or CLI flags.
- By default, agent launches (via `claude --name {workspace}`), editor does not launch.
- Launch behavior is controlled by CLI flags with precedence over config defaults.

---

## 5. Implementation Decisions

1. **Branch behavior precedence**: CLI flags (`--branch`, `--no-branch`) take priority. `--branch` and `--no-branch` are mutually exclusive. If no CLI flags provided, `checkoutBranchOnLaunch` setting is respected. If that is `true`, per-repo `templateBranchSyntax` (or global fallback) is used. If `false`, no checkout occurs.

2. **Error handling**: If a repo clone fails mid-launch, leave the partial workspace intact and inform the user. They can clean up with `muster remove <workspace>` if needed.

3. **Config validation**: Use strict validation. Error out if any repo URL is invalid or the clone fails. Do not skip invalid repos.

4. **Repo name inference**: Use git clone's default naming convention (last path component without `.git` suffix). Repos can optionally specify a custom `directory` field in config to override this.

5. **CLAUDE.md format**: Use a static template with placeholders for dynamic content:
   ```
   # Workspace: {workspace}
   
   {description}
   
   ## Repos
   {repos}
   ```
   Where `{description}` is the stack description and `{repos}` is an auto-generated list of repos with their paths.

---

## 6. MVP Scope

- ✅ Config loading from `~/.config/muster/config.json`
- ✅ `muster launch <workspace> [--stack] [--repo]` command
- ✅ `muster remove <workspace> [-y]` command
- ✅ Clone repos into `~/.workspaces/<workspace>/`
- ✅ Checkout branches (with template substitution)
- ✅ Generate `CLAUDE.md`
- ✅ Re-launch behavior (existing workspace → just print root path)
- ❌ GoReleaser, Homebrew, GitHub Actions (phase 2)
- ❌ Additional commands like `muster list`, etc. (phase 2)
