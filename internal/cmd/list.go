package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/outoforbitdev/muster/internal/config"
	"github.com/outoforbitdev/muster/internal/workspace"
)

var listCmd = &cobra.Command{
	Use:     "list [workspaces|stacks]",
	Aliases: []string{"ls"},
	Short:   "List workspaces or stacks",
	Long: `List available workspaces or stacks.

With no argument, or "workspaces", lists all workspaces found on disk along
with the stack they belong to (if any) and their cloned repos.

With "stacks", lists all stacks defined in the config along with their
description and repos.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "workspaces"
		if len(args) == 1 {
			target = args[0]
		}

		switch target {
		case "workspaces", "workspace", "ws":
			return listWorkspaces()
		case "stacks", "stack":
			return listStacks()
		default:
			return fmt.Errorf(`unknown list target %q: expected "workspaces" or "stacks"`, target)
		}
	},
}

// listWorkspaces prints all workspaces found on disk, grouped implicitly by
// the sorted order returned from workspace.ListWorkspaces.
func listWorkspaces() error {
	workspaces, err := workspace.ListWorkspaces()
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces found.")
		return nil
	}

	for _, ws := range workspaces {
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

// listStacks prints all stacks defined in the config, sorted by name.
func listStacks() error {
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
