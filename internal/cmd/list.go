package cmd

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/outoforbitdev/muster/internal/config"
	"github.com/outoforbitdev/muster/internal/workspace"
)

var listAll bool

var listCmd = &cobra.Command{
	Use:     "list [workspaces|stacks]",
	Aliases: []string{"ls"},
	Short:   "List workspaces or stacks",
	Long: `List available workspaces or stacks.

With no argument, or "workspaces", lists all workspace names found on disk.

With "stacks", lists all stack names defined in the config.

By default, output is concise: just names. Use --all for a table view: for
workspaces, each workspace's stack (if any) and cloned repos; for stacks,
each stack's repos and their descriptions.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "workspaces"
		if len(args) == 1 {
			target = args[0]
		}

		switch target {
		case "workspaces", "workspace", "ws":
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			return listWorkspaces(cfg, listAll)
		case "stacks", "stack":
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			return listStacks(cfg, listAll)
		default:
			return fmt.Errorf(`unknown list target %q: expected "workspaces" or "stacks"`, target)
		}
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Show details: stack membership and repos for workspaces, description and repos for stacks")
}

// listWorkspaces prints workspaces found on disk, sorted per
// workspace.ListWorkspaces. In concise mode (the default) it prints just
// workspace names; with all set it prints a WORKSPACE/STACK/REPO table,
// repeating a workspace's name and stack only on its first repo row.
func listWorkspaces(cfg *config.Config, all bool) error {
	workspaces, err := workspace.ListWorkspaces(cfg)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces found.")
		return nil
	}

	if !all {
		for _, ws := range workspaces {
			fmt.Println(ws.Name)
		}
		return nil
	}

	var rows [][]string
	for _, ws := range workspaces {
		repos := ws.Repos
		if len(repos) == 0 {
			repos = []string{""}
		}
		for i, repo := range repos {
			if i == 0 {
				rows = append(rows, []string{ws.Name, ws.Stack, repo})
			} else {
				rows = append(rows, []string{"", "", repo})
			}
		}
	}

	renderTable(os.Stdout, []string{"WORKSPACE", "STACK", "REPO"}, rows)
	return nil
}

// listStacks prints stacks defined in the config, sorted by name. In
// concise mode (the default) it prints just stack names; with all set it
// prints a STACK/REPO/DESCRIPTION table, wrapping each repo's description
// and repeating a stack's name only on its first repo row.
func listStacks(cfg *config.Config, all bool) error {
	if len(cfg.Stacks) == 0 {
		fmt.Println("No stacks configured.")
		return nil
	}

	names := make([]string, 0, len(cfg.Stacks))
	for name := range cfg.Stacks {
		names = append(names, name)
	}
	sort.Strings(names)

	if !all {
		for _, name := range names {
			fmt.Println(name)
		}
		return nil
	}

	var rows [][]string
	for _, name := range names {
		stack := cfg.Stacks[name]
		if len(stack.Repos) == 0 {
			rows = append(rows, []string{name, "", ""})
			continue
		}

		for i, repo := range stack.Repos {
			descLines := wrapText(repo.Description, descriptionWrapWidth)
			for j, line := range descLines {
				switch {
				case i == 0 && j == 0:
					rows = append(rows, []string{name, repoName(repo), line})
				case j == 0:
					rows = append(rows, []string{"", repoName(repo), line})
				default:
					rows = append(rows, []string{"", "", line})
				}
			}
		}
	}

	renderTable(os.Stdout, []string{"STACK", "REPO", "DESCRIPTION"}, rows)
	return nil
}

// repoName derives a short, human-readable name for a repo: its configured
// directory override, or its URL's base name with a trailing ".git" stripped.
func repoName(repo config.Repo) string {
	if repo.Directory != "" {
		return repo.Directory
	}
	return strings.TrimSuffix(path.Base(repo.URL), ".git")
}
