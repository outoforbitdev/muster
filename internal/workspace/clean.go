package workspace

import (
	"fmt"
	"os/exec"
	"strings"
)

// runGitCommand runs git with the given args in dir, returning combined
// stdout+stderr output for error reporting.
func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// isDirty reports whether repoPath has any uncommitted or untracked changes.
func isDirty(repoPath string) (bool, error) {
	out, err := runGitCommand(repoPath, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}
	return strings.TrimSpace(out) != "", nil
}

// defaultBranch determines the remote-tracking ref for repoPath's default
// branch, trying origin/main before falling back to origin/master.
func defaultBranch(repoPath string) (string, error) {
	for _, candidate := range []string{"origin/main", "origin/master"} {
		if _, err := runGitCommand(repoPath, "rev-parse", "--verify", candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not determine default branch: no origin/main or origin/master found")
}

// hasDiverged reports whether HEAD has any commits not present on base.
func hasDiverged(repoPath, base string) (bool, error) {
	out, err := runGitCommand(repoPath, "rev-list", base+"..HEAD", "--count")
	if err != nil {
		return false, fmt.Errorf("failed to compare against %s: %w", base, err)
	}
	return strings.TrimSpace(out) != "0", nil
}

// currentBranch returns the name of the currently checked-out branch.
func currentBranch(repoPath string) (string, error) {
	out, err := runGitCommand(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to determine current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// Repo cleanliness statuses used in RepoState.Status.
const (
	StatusDirty    = "dirty"
	StatusClean    = "clean"
	StatusUnmerged = "unmerged"
	StatusUnknown  = "unknown"
)

// RepoState describes whether a repo is safe to clean up.
type RepoState struct {
	Repo   string
	Status string
	Detail string
}

// MergeChecker reports whether branch has been merged into its default
// branch, e.g. via a GitHub PR lookup. Implementations are injected so the
// evaluation logic can be tested without shelling out to gh/GitHub.
type MergeChecker func(repoPath, branch string) (bool, error)

// EvaluateRepo determines whether repoPath is safe to clean up: a repo is
// unsafe if it has uncommitted/untracked changes, or if it has diverged
// from its default branch without the divergent commits being merged. A
// nil checkMerged defaults to checking merge status via the gh CLI.
func EvaluateRepo(repoPath, repoName string, checkMerged MergeChecker) RepoState {
	if checkMerged == nil {
		checkMerged = checkMergedViaGH
	}

	dirty, err := isDirty(repoPath)
	if err != nil {
		return RepoState{Repo: repoName, Status: StatusUnknown, Detail: err.Error()}
	}
	if dirty {
		return RepoState{Repo: repoName, Status: StatusDirty, Detail: "uncommitted changes"}
	}

	if _, err := runGitCommand(repoPath, "fetch", "origin"); err != nil {
		return RepoState{Repo: repoName, Status: StatusUnknown, Detail: "failed to fetch: " + err.Error()}
	}

	base, err := defaultBranch(repoPath)
	if err != nil {
		return RepoState{Repo: repoName, Status: StatusUnknown, Detail: err.Error()}
	}

	diverged, err := hasDiverged(repoPath, base)
	if err != nil {
		return RepoState{Repo: repoName, Status: StatusUnknown, Detail: err.Error()}
	}
	if !diverged {
		return RepoState{Repo: repoName, Status: StatusClean, Detail: "no divergence from " + base}
	}

	branch, err := currentBranch(repoPath)
	if err != nil {
		return RepoState{Repo: repoName, Status: StatusUnknown, Detail: err.Error()}
	}

	merged, err := checkMerged(repoPath, branch)
	if err != nil {
		return RepoState{Repo: repoName, Status: StatusUnknown, Detail: err.Error()}
	}
	if merged {
		return RepoState{Repo: repoName, Status: StatusClean, Detail: "merged"}
	}
	return RepoState{Repo: repoName, Status: StatusUnmerged, Detail: "not merged"}
}

// IsWorkspaceSafe reports whether every repo in a workspace is safe to
// clean up: a workspace is only safe if all of its repos are StatusClean.
func IsWorkspaceSafe(states []RepoState) bool {
	for _, state := range states {
		if state.Status != StatusClean {
			return false
		}
	}
	return true
}

// checkMergedViaGH asks GitHub whether branch's pull request has been
// merged, since squash merges cannot be detected reliably from local git
// history alone.
func checkMergedViaGH(repoPath, branch string) (bool, error) {
	cmd := exec.Command("gh", "pr", "view", branch, "--json", "state", "-q", ".state")
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("gh pr view failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)) == "MERGED", nil
}
