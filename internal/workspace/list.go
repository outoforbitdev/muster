package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/outoforbitdev/muster/internal/config"
)

// Info describes a workspace discovered on disk.
type Info struct {
	Name  string
	Stack string
	Path  string
	Repos []string
}

// ListWorkspaces discovers all workspaces under the workspaces root.
//
// Workspaces are created flat under the root, except when created from a
// stack, in which case they're nested one level under a directory named
// after that stack (see WorkspacePath). Since a stack directory's name
// always matches a configured stack, cfg.Stacks is used to tell stack
// container directories apart from flat workspace directories, rather than
// inspecting the workspace's contents (which may be incomplete or broken).
// Results are sorted by stack, then by name.
func ListWorkspaces(cfg *config.Config) ([]Info, error) {
	root := WorkspacesRoot()

	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read workspaces directory: %w", err)
	}

	var workspaces []Info

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(root, entry.Name())

		if _, isStack := cfg.Stacks[entry.Name()]; !isStack {
			workspaces = append(workspaces, newInfo("", entry.Name(), path))
			continue
		}

		// entry.Name() is a stack; its subdirectories are workspaces.
		subEntries, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		for _, sub := range subEntries {
			if !sub.IsDir() {
				continue
			}
			subPath := filepath.Join(path, sub.Name())
			workspaces = append(workspaces, newInfo(entry.Name(), sub.Name(), subPath))
		}
	}

	sort.Slice(workspaces, func(i, j int) bool {
		if workspaces[i].Stack != workspaces[j].Stack {
			return workspaces[i].Stack < workspaces[j].Stack
		}
		return workspaces[i].Name < workspaces[j].Name
	})

	return workspaces, nil
}

// newInfo builds an Info for a workspace directory, populating its repos.
func newInfo(stack, name, path string) Info {
	return Info{
		Name:  name,
		Stack: stack,
		Path:  path,
		Repos: listRepos(path),
	}
}

// listRepos returns the names of a workspace's immediate subdirectories,
// sorted alphabetically. Each one is expected to be a cloned repo, but
// directories are listed as-is (rather than requiring a ".git" entry) so a
// repo that failed to clone or lost its ".git" still shows up.
func listRepos(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	var repos []string
	for _, entry := range entries {
		if entry.IsDir() {
			repos = append(repos, entry.Name())
		}
	}

	sort.Strings(repos)
	return repos
}
