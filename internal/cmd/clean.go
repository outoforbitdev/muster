package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/outoforbitdev/muster/internal/config"
	"github.com/outoforbitdev/muster/internal/workspace"
)

var (
	cleanWrite bool
	cleanYes   bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean [workspace...]",
	Short: "Clean up workspaces that are safe to delete",
	Long: `Report which workspaces are safe to delete, and optionally delete them.

A workspace is safe to clean only if every one of its repos is safe:
  - no uncommitted or untracked changes
  - either no divergence from its default branch (main/master), or its
    divergent commits have been merged (checked via "gh pr view")

By default this is a dry run: it only reports each repo's status. Use
--write to actually delete the workspaces classified as safe.

With no workspace names given, all workspaces are evaluated. With names
given, only those workspaces are evaluated (and, with --write, deleted).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClean(args, cleanWrite, cleanYes, nil)
	},
}

func init() {
	cleanCmd.Flags().BoolVarP(&cleanWrite, "write", "w", false, "Delete workspaces classified as safe to clean (default is dry-run)")
	cleanCmd.Flags().BoolVarP(&cleanYes, "yes", "y", false, "Skip confirmation prompt when deleting with --write")
}

// runClean evaluates the named workspaces (or all workspaces, if names is
// empty), prints a status table, and, if write is true, deletes the ones
// classified as safe. checkMerged is injected for testing; a nil value
// defaults to checking merge status via the gh CLI.
func runClean(names []string, write, yes bool, checkMerged workspace.MergeChecker) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	workspaces, err := workspace.ListWorkspaces(cfg)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	if len(names) > 0 {
		workspaces, err = filterWorkspaces(workspaces, names)
		if err != nil {
			return err
		}
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces found.")
		return nil
	}

	type evaluatedWorkspace struct {
		ws     workspace.Info
		states []workspace.RepoState
	}

	var evaluated []evaluatedWorkspace
	maxWorkspaceLen := len("WORKSPACE")
	maxStackLen := len("STACK")
	maxRepoLen := len("REPO")

	for _, ws := range workspaces {
		var states []workspace.RepoState
		if len(ws.Repos) == 0 {
			states = []workspace.RepoState{{Status: workspace.StatusUnknown, Detail: "no repos found"}}
		}
		for _, repo := range ws.Repos {
			states = append(states, workspace.EvaluateRepo(filepath.Join(ws.Path, repo), repo, checkMerged))
			if len(repo) > maxRepoLen {
				maxRepoLen = len(repo)
			}
		}
		if len(ws.Name) > maxWorkspaceLen {
			maxWorkspaceLen = len(ws.Name)
		}
		if len(ws.Stack) > maxStackLen {
			maxStackLen = len(ws.Stack)
		}
		evaluated = append(evaluated, evaluatedWorkspace{ws: ws, states: states})
	}

	stateWrapWidth := stateColumnWrapWidth(maxWorkspaceLen, maxStackLen, maxRepoLen)

	var rows []cleanRow
	var safeWorkspaces []workspace.Info
	unsafeCount := 0

	for _, ew := range evaluated {
		safe := workspace.IsWorkspaceSafe(ew.states)
		workspaceColor := ansiRed
		if safe {
			workspaceColor = ansiGreen
		}

		for i, state := range ew.states {
			repoColor := repoColorFor(state.Status)

			status := state.Status
			if state.Detail != "" {
				status = status + ", " + state.Detail
			}

			statusLines := wrapText(status, stateWrapWidth)
			for j, line := range statusLines {
				switch {
				case i == 0 && j == 0:
					rows = append(rows, cleanRow{workspace: ew.ws.Name, stack: ew.ws.Stack, repo: state.Repo, state: line, workspaceColor: workspaceColor, repoColor: repoColor})
				case j == 0:
					rows = append(rows, cleanRow{repo: state.Repo, state: line, workspaceColor: workspaceColor, repoColor: repoColor})
				default:
					rows = append(rows, cleanRow{state: line, workspaceColor: workspaceColor, repoColor: repoColor})
				}
			}
		}

		if safe {
			safeWorkspaces = append(safeWorkspaces, ew.ws)
		} else {
			unsafeCount++
		}
	}

	renderCleanTable(os.Stdout, maxWorkspaceLen, maxStackLen, maxRepoLen, rows)
	fmt.Printf("\n%d workspace(s) safe to clean, %d not safe.\n", len(safeWorkspaces), unsafeCount)

	if !write {
		if len(safeWorkspaces) > 0 {
			fmt.Println("Re-run with --write to delete safe workspaces.")
		}
		return nil
	}

	return deleteWorkspaces(safeWorkspaces, yes)
}

// tableTerminalWidth is the terminal width the STATE column is wrapped to
// fit, so a rendered row never needs re-wrapping by the terminal itself
// (which would break mid-word instead of on a word boundary).
const tableTerminalWidth = 80

// tabPadding matches the padding renderTable's tabwriter inserts between
// columns.
const tabPadding = 2

// minStateWrapWidth is an absolute floor on the STATE column's wrap width:
// below this, wrapping stops being useful (near one word per line), so it's
// preferable to let very long workspace/stack/repo names push the row past
// tableTerminalWidth rather than wrap STATE into an unreadable sliver.
const minStateWrapWidth = 10

// stateColumnWrapWidth computes how wide the STATE column's wrapped text
// can be so that, combined with the widest WORKSPACE, STACK, and REPO
// values (with tabwriter's padding between each), the full row fits within
// tableTerminalWidth whenever the other columns leave enough room.
func stateColumnWrapWidth(maxWorkspaceLen, maxStackLen, maxRepoLen int) int {
	reserved := maxWorkspaceLen + tabPadding + maxStackLen + tabPadding + maxRepoLen + tabPadding
	width := tableTerminalWidth - reserved
	if width < minStateWrapWidth {
		return minStateWrapWidth
	}
	return width
}

// filterWorkspaces returns the subset of workspaces whose Name matches one
// of names, erroring if any requested name has no match.
func filterWorkspaces(workspaces []workspace.Info, names []string) ([]workspace.Info, error) {
	wanted := make(map[string]bool, len(names))
	for _, name := range names {
		wanted[name] = true
	}

	var filtered []workspace.Info
	found := make(map[string]bool, len(names))
	for _, ws := range workspaces {
		if wanted[ws.Name] {
			filtered = append(filtered, ws)
			found[ws.Name] = true
		}
	}

	for _, name := range names {
		if !found[name] {
			return nil, fmt.Errorf("workspace %q not found", name)
		}
	}

	return filtered, nil
}

// deleteWorkspaces removes each workspace's directory from disk, prompting
// for confirmation first unless yes is true.
func deleteWorkspaces(safeWorkspaces []workspace.Info, yes bool) error {
	if len(safeWorkspaces) == 0 {
		return nil
	}

	if !yes {
		fmt.Println("\nAbout to delete:")
		for _, ws := range safeWorkspaces {
			fmt.Printf("  %s\n", ws.Name)
		}
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		if strings.TrimSpace(response) != "yes" {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	for _, ws := range safeWorkspaces {
		if err := os.RemoveAll(ws.Path); err != nil {
			return fmt.Errorf("failed to delete workspace %q: %w", ws.Name, err)
		}
		fmt.Printf("Deleted workspace %q\n", ws.Name)
	}

	return nil
}
