package cmd

import (
	"fmt"
	"sort"

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

By default, output is concise: just names. Use --all to also show, for
workspaces, the stack each belongs to (if any) and their cloned repos; and
for stacks, their description and repos.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "workspaces"
		if len(args) == 1 {
			target = args[0]
		}

		switch target {
		case "workspaces", "workspace", "ws":
			return listWorkspaces(listAll)
		case "stacks", "stack":
			return listStacks(listAll)
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
// workspace names; with all set it also prints stack membership and repos.
func listWorkspaces(all bool) error {
	workspaces, err := workspace.ListWorkspaces()
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces found.")
		return nil
	}

	for _, ws := range workspaces {
		if !all {
			fmt.Println(ws.Name)
			continue
		}

		if ws.Stack != "" {
			fmt.Printf("%s (stack: %s)\n", ws.Name, ws.Stack)
		} else {
			fmt.Println(ws.Name)
		}
		for _, repo := range ws.Repos {
			fmt.Printf("  - %s\n", repo)
		}
	}

	return nil
}

// listStacks prints stacks defined in the config, sorted by name. In
// concise mode (the default) it prints just stack names; with all set it
// also prints each stack's description and repos.
func listStacks(all bool) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Stacks) == 0 {
		fmt.Println("No stacks configured.")
		return nil
	}

	names := make([]string, 0, len(cfg.Stacks))
	for name := range cfg.Stacks {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if !all {
			fmt.Println(name)
			continue
		}

		stack := cfg.Stacks[name]
		if stack.Description != "" {
			fmt.Printf("%s: %s\n", name, stack.Description)
		} else {
			fmt.Println(name)
		}
		for _, repo := range stack.Repos {
			fmt.Printf("  - %s\n", repo.URL)
		}
	}

	return nil
}
