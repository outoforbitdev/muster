package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/outoforbitdev/muster/internal/config"
)

// CreateWorkspace creates a new workspace by cloning repos and checking out branches.
func CreateWorkspace(
	cfg *config.Config,
	workspace string,
	stackNames []string,
	repoURLs []string,
	branch string,
	noBranch bool,
) error {
	// Expand workspace path
	workspacePath := filepath.Join(os.Getenv("HOME"), ".muster", workspace)

	// Create workspace directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Collect repos from stacks and explicit URLs
	reposToClone := make([]RepoToClone, 0)

	// Add repos from stacks
	for _, stackName := range stackNames {
		stack := cfg.GetStack(stackName)
		if stack == nil {
			return fmt.Errorf("stack %q not found in config", stackName)
		}

		for _, repo := range stack.Repos {
			rtc := RepoToClone{
				URL:                  repo.URL,
				TemplateBranchSyntax: repo.TemplateBranchSyntax,
				Description:          repo.Description,
				Directory:            repo.Directory,
			}
			reposToClone = append(reposToClone, rtc)
		}
	}

	// Add explicit repos
	for _, url := range repoURLs {
		rtc := RepoToClone{
			URL: url,
		}
		reposToClone = append(reposToClone, rtc)
	}

	if len(reposToClone) == 0 {
		return fmt.Errorf("no repos specified")
	}

	// Determine branch checkout behavior
	checkoutBranch := shouldCheckoutBranch(cfg, branch, noBranch)

	// Clone and checkout repos
	for i, rtc := range reposToClone {
		repoPath := getRepoPath(workspacePath, &rtc)

		// Clone the repo
		if err := cloneRepo(rtc.URL, repoPath); err != nil {
			return fmt.Errorf("failed to clone repo %d (%s): %w", i, rtc.URL, err)
		}

		// Checkout branch if needed
		if checkoutBranch {
			targetBranch := determineBranch(branch, rtc.TemplateBranchSyntax, cfg.Defaults.TemplateBranchSyntax)
			if targetBranch != "" {
				substitutedBranch := SubstituteTemplate(targetBranch, workspace)
				if err := checkoutBranchInRepo(repoPath, substitutedBranch); err != nil {
					return fmt.Errorf("failed to checkout branch %q in repo %d (%s): %w", substitutedBranch, i, rtc.URL, err)
				}
			}
		}
	}

	return nil
}

// RepoToClone represents a repo to be cloned.
type RepoToClone struct {
	URL                  string
	TemplateBranchSyntax string
	Description          string
	Directory            string
}

// getRepoPath determines where to clone a repo based on its config.
func getRepoPath(workspacePath string, rtc *RepoToClone) string {
	if rtc.Directory != "" {
		return filepath.Join(workspacePath, rtc.Directory)
	}

	// Use git's default naming: last path component without .git
	repoName := path.Base(rtc.URL)
	if len(repoName) > 4 && repoName[len(repoName)-4:] == ".git" {
		repoName = repoName[:len(repoName)-4]
	}

	return filepath.Join(workspacePath, repoName)
}

// shouldCheckoutBranch determines if we should checkout branches based on precedence.
// Precedence:
// 1. If --no-branch is set, return false
// 2. If --branch is set, return true
// 3. Check checkoutBranchOnLaunch setting (default true)
func shouldCheckoutBranch(cfg *config.Config, branch string, noBranch bool) bool {
	if noBranch {
		return false
	}
	if branch != "" {
		return true
	}
	return cfg.Defaults.CheckoutBranchOnLaunch
}

// determineBranch determines which branch to checkout based on precedence.
// Precedence:
// 1. CLI --branch flag
// 2. Per-repo templateBranchSyntax
// 3. Global defaults.templateBranchSyntax
// 4. Empty string (use default after clone)
func determineBranch(cliBranch, repoTemplate, globalTemplate string) string {
	if cliBranch != "" {
		return cliBranch
	}
	if repoTemplate != "" {
		return repoTemplate
	}
	return globalTemplate
}

// cloneRepo clones a git repository to the specified path.
func cloneRepo(url, path string) error {
	cmd := exec.Command("git", "clone", url, path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}

// checkoutBranchInRepo checks out a branch in an existing repository.
func checkoutBranchInRepo(repoPath, branch string) error {
	cmd := exec.Command("git", "-C", repoPath, "checkout", branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}
	return nil
}
