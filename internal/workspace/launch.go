package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/outoforbitdev/muster/internal/claude"
	"github.com/outoforbitdev/muster/internal/config"
)

// LaunchWorkspace launches an existing workspace or creates a new one.
// Returns the workspace path and any error encountered.
func LaunchWorkspace(
	cfg *config.Config,
	workspace string,
	stackNames []string,
	repoURLs []string,
	branch string,
	noBranch bool,
) (string, error) {
	workspacePath := filepath.Join(os.Getenv("HOME"), ".muster", workspace)

	// Check if workspace exists
	if _, err := os.Stat(workspacePath); err == nil {
		// Workspace exists, just return the path
		return workspacePath, nil
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return "", fmt.Errorf("failed to check workspace: %w", err)
	}

	// Workspace doesn't exist, create it
	if err := CreateWorkspace(cfg, workspace, stackNames, repoURLs, branch, noBranch); err != nil {
		return "", fmt.Errorf("failed to create workspace: %w", err)
	}

	// Generate CLAUDE.md
	if len(stackNames) > 0 {
		stackName := stackNames[0]
		stack := cfg.GetStack(stackName)
		repos := make(map[string]string)
		for _, repo := range stack.Repos {
			rtc := RepoToClone{
				URL:       repo.URL,
				Directory: repo.Directory,
			}
			repoPath := getRepoPath(workspacePath, &rtc)
			repoName := filepath.Base(repoPath)
			repos[repoName] = repoPath
		}

		claudeContent := claude.GenerateCLAUDE(workspace, stack, repos)
		if err := claude.WriteCLAUDE(workspacePath, claudeContent); err != nil {
			// Don't fail the entire launch if CLAUDE.md generation fails
			fmt.Fprintf(os.Stderr, "warning: failed to generate CLAUDE.md: %v\n", err)
		}
	}

	return workspacePath, nil
}

// LaunchClaude launches Claude Code with the specified workspace name.
// Claude Code must be installed and available in PATH.
func LaunchClaude(workspacePath, workspace string) error {
	cmd := exec.Command("claude", "launch", "--name", workspace)
	cmd.Dir = workspacePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch Claude Code: %w", err)
	}
	return nil
}
