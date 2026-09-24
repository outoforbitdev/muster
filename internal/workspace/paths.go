package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// WorkspacesRoot returns the root directory where all workspaces live.
func WorkspacesRoot() string {
	return filepath.Join(os.Getenv("HOME"), ".muster", "workspaces")
}

// WorkspacePath returns the on-disk path for a workspace. To prevent name
// collisions between stacks, workspaces created from a stack are nested
// under a subdirectory named after that stack. Workspaces created without a
// stack (repos added via --repo only) live directly under the workspaces
// root.
func WorkspacePath(stackName, workspaceName string) string {
	root := WorkspacesRoot()
	if stackName != "" {
		return filepath.Join(root, stackName, workspaceName)
	}
	return filepath.Join(root, workspaceName)
}

// FindWorkspacePath locates an existing workspace directory by name.
//
// If stackName is non-empty, it looks only under that stack's subdirectory.
// Otherwise it searches the workspaces root: directly (for workspaces
// created without a stack) and one level of stack subdirectories. It
// returns an error if the workspace can't be found, or if the name is
// ambiguous across multiple stacks (in which case the caller should retry
// with an explicit stack name).
func FindWorkspacePath(stackName, workspaceName string) (string, error) {
	root := WorkspacesRoot()

	if stackName != "" {
		path := filepath.Join(root, stackName, workspaceName)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("workspace %q not found under stack %q", workspaceName, stackName)
			}
			return "", fmt.Errorf("failed to check workspace: %w", err)
		}
		return path, nil
	}

	var matches []string

	// Workspaces created without a stack live directly under the root.
	if info, err := os.Stat(filepath.Join(root, workspaceName)); err == nil && info.IsDir() {
		matches = append(matches, filepath.Join(root, workspaceName))
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check workspace: %w", err)
	}

	// Search one level of stack subdirectories.
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("workspace %q not found", workspaceName)
		}
		return "", fmt.Errorf("failed to read workspaces directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(root, entry.Name(), workspaceName)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			matches = append(matches, candidate)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("workspace %q not found", workspaceName)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("workspace %q exists under multiple stacks; specify which one with --stack", workspaceName)
	}
}
