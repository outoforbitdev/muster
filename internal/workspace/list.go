package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
// A directory is treated as a workspace itself if it directly contains one
// or more cloned git repos. Otherwise it is treated as a stack container,
// and each of its subdirectories is listed as a workspace belonging to that
// stack. Results are sorted by stack, then by name.
func ListWorkspaces() ([]Info, error) {
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

		if isWorkspaceDir(path) {
			workspaces = append(workspaces, newInfo("", entry.Name(), path))
			continue
		}

		// Treat as a stack container; list its subdirectories as workspaces.
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

// newInfo builds an Info for a workspace directory, populating its cloned repos.
func newInfo(stack, name, path string) Info {
	return Info{
		Name:  name,
		Stack: stack,
		Path:  path,
		Repos: listRepos(path),
	}
}

// isWorkspaceDir reports whether path directly contains at least one cloned
// git repo, which distinguishes a workspace directory from a stack
// container directory.
func isWorkspaceDir(path string) bool {
	return len(listRepos(path)) > 0
}

// listRepos returns the names of immediate subdirectories of path that are
// cloned git repos, sorted alphabetically.
func listRepos(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	var repos []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(path, entry.Name(), ".git")); err == nil {
			repos = append(repos, entry.Name())
		}
	}

	sort.Strings(repos)
	return repos
}
