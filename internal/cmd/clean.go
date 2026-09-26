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

	var rows [][]string
	var safeWorkspaces []workspace.Info
	unsafeCount := 0

	for _, ws := range workspaces {
		repos := ws.Repos
		if len(repos) == 0 {
			repos = []string{""}
		}

		var states []workspace.RepoState
		for _, repo := range repos {
			if repo == "" {
				continue
			}
			states = append(states, workspace.EvaluateRepo(filepath.Join(ws.Path, repo), repo, checkMerged))
		}

		for i, state := range states {
			status := state.Status
			if state.Detail != "" {
				status = status + ", " + state.Detail
			}
			if i == 0 {
				rows = append(rows, []string{ws.Name, ws.Stack, state.Repo, status})
			} else {
				rows = append(rows, []string{"", "", state.Repo, status})
			}
		}

		if workspace.IsWorkspaceSafe(states) {
			safeWorkspaces = append(safeWorkspaces, ws)
		} else {
			unsafeCount++
		}
	}

	renderTable(os.Stdout, []string{"WORKSPACE", "STACK", "REPO", "STATE"}, rows)
	fmt.Printf("\n%d workspace(s) safe to clean, %d not safe.\n", len(safeWorkspaces), unsafeCount)

	if !write {
		if len(safeWorkspaces) > 0 {
			fmt.Println("Re-run with --write to delete safe workspaces.")
		}
		return nil
	}

	return deleteWorkspaces(safeWorkspaces, yes)
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
