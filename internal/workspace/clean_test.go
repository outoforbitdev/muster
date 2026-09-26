package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// initTestRepo creates a git repo at path with an initial commit on branch
// "main", returning the repo path for convenience.
func initTestRepo(t *testing.T, path string) string {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	runGit(t, path, "init", "-b", "main")
	runGit(t, path, "commit", "--allow-empty", "-m", "initial commit")

	return path
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := runGitCommand(dir, args...)
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return out
}

func TestIsWorkspaceSafeAllClean(t *testing.T) {
	states := []RepoState{
		{Repo: "a", Status: StatusClean},
		{Repo: "b", Status: StatusClean},
	}
	if !IsWorkspaceSafe(states) {
		t.Errorf("expected workspace with all-clean repos to be safe")
	}
}

func TestIsWorkspaceSafeOneDirty(t *testing.T) {
	states := []RepoState{
		{Repo: "a", Status: StatusClean},
		{Repo: "b", Status: StatusDirty},
	}
	if IsWorkspaceSafe(states) {
		t.Errorf("expected workspace with a dirty repo to be unsafe")
	}
}

func TestIsWorkspaceSafeOneUnmerged(t *testing.T) {
	states := []RepoState{
		{Repo: "a", Status: StatusClean},
		{Repo: "b", Status: StatusUnmerged},
	}
	if IsWorkspaceSafe(states) {
		t.Errorf("expected workspace with an unmerged repo to be unsafe")
	}
}

func TestIsWorkspaceSafeOneUnknown(t *testing.T) {
	states := []RepoState{
		{Repo: "a", Status: StatusClean},
		{Repo: "b", Status: StatusUnknown},
	}
	if IsWorkspaceSafe(states) {
		t.Errorf("expected workspace with an unknown-status repo to be unsafe")
	}
}

func TestIsDirtyCleanRepo(t *testing.T) {
	repoPath := initTestRepo(t, filepath.Join(t.TempDir(), "repo"))

	dirty, err := isDirty(repoPath)
	if err != nil {
		t.Fatalf("isDirty returned error: %v", err)
	}
	if dirty {
		t.Errorf("expected clean repo to report dirty=false")
	}
}

func TestIsDirtyWithUntrackedFile(t *testing.T) {
	repoPath := initTestRepo(t, filepath.Join(t.TempDir(), "repo"))

	if err := os.WriteFile(filepath.Join(repoPath, "untracked.txt"), []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}

	dirty, err := isDirty(repoPath)
	if err != nil {
		t.Fatalf("isDirty returned error: %v", err)
	}
	if !dirty {
		t.Errorf("expected repo with untracked file to report dirty=true")
	}
}

// initRemoteWithClone creates a bare "remote" repo with an initial commit on
// main, then clones it into a working repo, returning the working repo path.
func initRemoteWithClone(t *testing.T) string {
	t.Helper()

	remotePath := filepath.Join(t.TempDir(), "remote.git")
	if err := os.MkdirAll(remotePath, 0o755); err != nil {
		t.Fatalf("failed to create remote dir: %v", err)
	}
	runGit(t, remotePath, "init", "--bare", "-b", "main")

	// Seed the bare remote via a throwaway working clone.
	seedPath := filepath.Join(t.TempDir(), "seed")
	if err := os.MkdirAll(seedPath, 0o755); err != nil {
		t.Fatalf("failed to create seed dir: %v", err)
	}
	runGit(t, seedPath, "init", "-b", "main")
	runGit(t, seedPath, "commit", "--allow-empty", "-m", "initial commit")
	runGit(t, seedPath, "remote", "add", "origin", remotePath)
	runGit(t, seedPath, "push", "origin", "main")

	clonePath := filepath.Join(t.TempDir(), "clone")
	out, err := runGitCommand(filepath.Dir(clonePath), "clone", remotePath, clonePath)
	if err != nil {
		t.Fatalf("failed to clone remote: %v\n%s", err, out)
	}

	return clonePath
}

func TestDefaultBranchPrefersOriginMain(t *testing.T) {
	repoPath := initRemoteWithClone(t)

	branch, err := defaultBranch(repoPath)
	if err != nil {
		t.Fatalf("defaultBranch returned error: %v", err)
	}
	if branch != "origin/main" {
		t.Errorf("expected origin/main, got %q", branch)
	}
}

func TestDefaultBranchFallsBackToMaster(t *testing.T) {
	remotePath := filepath.Join(t.TempDir(), "remote.git")
	if err := os.MkdirAll(remotePath, 0o755); err != nil {
		t.Fatalf("failed to create remote dir: %v", err)
	}
	runGit(t, remotePath, "init", "--bare", "-b", "master")

	seedPath := filepath.Join(t.TempDir(), "seed")
	if err := os.MkdirAll(seedPath, 0o755); err != nil {
		t.Fatalf("failed to create seed dir: %v", err)
	}
	runGit(t, seedPath, "init", "-b", "master")
	runGit(t, seedPath, "commit", "--allow-empty", "-m", "initial commit")
	runGit(t, seedPath, "remote", "add", "origin", remotePath)
	runGit(t, seedPath, "push", "origin", "master")

	clonePath := filepath.Join(t.TempDir(), "clone")
	if out, err := runGitCommand(filepath.Dir(clonePath), "clone", remotePath, clonePath); err != nil {
		t.Fatalf("failed to clone remote: %v\n%s", err, out)
	}

	branch, err := defaultBranch(clonePath)
	if err != nil {
		t.Fatalf("defaultBranch returned error: %v", err)
	}
	if branch != "origin/master" {
		t.Errorf("expected origin/master, got %q", branch)
	}
}

func TestHasDivergedNoNewCommits(t *testing.T) {
	repoPath := initRemoteWithClone(t)

	diverged, err := hasDiverged(repoPath, "origin/main")
	if err != nil {
		t.Fatalf("hasDiverged returned error: %v", err)
	}
	if diverged {
		t.Errorf("expected no divergence right after clone")
	}
}

func TestHasDivergedWithNewCommit(t *testing.T) {
	repoPath := initRemoteWithClone(t)

	runGit(t, repoPath, "checkout", "-b", "feature")
	runGit(t, repoPath, "commit", "--allow-empty", "-m", "feature work")

	diverged, err := hasDiverged(repoPath, "origin/main")
	if err != nil {
		t.Fatalf("hasDiverged returned error: %v", err)
	}
	if !diverged {
		t.Errorf("expected divergence after adding a commit ahead of origin/main")
	}
}

func TestEvaluateRepoDirty(t *testing.T) {
	repoPath := initRemoteWithClone(t)
	if err := os.WriteFile(filepath.Join(repoPath, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}

	state := EvaluateRepo(repoPath, "myrepo", nil)

	if state.Status != StatusDirty {
		t.Errorf("expected status %q, got %q (detail: %s)", StatusDirty, state.Status, state.Detail)
	}
}

func TestEvaluateRepoCleanNoDivergence(t *testing.T) {
	repoPath := initRemoteWithClone(t)

	state := EvaluateRepo(repoPath, "myrepo", nil)

	if state.Status != StatusClean {
		t.Errorf("expected status %q, got %q (detail: %s)", StatusClean, state.Status, state.Detail)
	}
}

func TestEvaluateRepoDivergedAndMerged(t *testing.T) {
	repoPath := initRemoteWithClone(t)
	runGit(t, repoPath, "checkout", "-b", "feature")
	runGit(t, repoPath, "commit", "--allow-empty", "-m", "feature work")

	fakeChecker := func(repoPath, branch string) (bool, error) {
		return true, nil
	}

	state := EvaluateRepo(repoPath, "myrepo", fakeChecker)

	if state.Status != StatusClean {
		t.Errorf("expected status %q, got %q (detail: %s)", StatusClean, state.Status, state.Detail)
	}
}

func TestEvaluateRepoDivergedAndUnmerged(t *testing.T) {
	repoPath := initRemoteWithClone(t)
	runGit(t, repoPath, "checkout", "-b", "feature")
	runGit(t, repoPath, "commit", "--allow-empty", "-m", "feature work")

	fakeChecker := func(repoPath, branch string) (bool, error) {
		return false, nil
	}

	state := EvaluateRepo(repoPath, "myrepo", fakeChecker)

	if state.Status != StatusUnmerged {
		t.Errorf("expected status %q, got %q (detail: %s)", StatusUnmerged, state.Status, state.Detail)
	}
}

func TestEvaluateRepoMergeCheckError(t *testing.T) {
	repoPath := initRemoteWithClone(t)
	runGit(t, repoPath, "checkout", "-b", "feature")
	runGit(t, repoPath, "commit", "--allow-empty", "-m", "feature work")

	fakeChecker := func(repoPath, branch string) (bool, error) {
		return false, fmt.Errorf("gh not authenticated")
	}

	state := EvaluateRepo(repoPath, "myrepo", fakeChecker)

	if state.Status != StatusUnknown {
		t.Errorf("expected status %q, got %q (detail: %s)", StatusUnknown, state.Status, state.Detail)
	}
}

func TestIsDirtyWithUncommittedChange(t *testing.T) {
	repoPath := initTestRepo(t, filepath.Join(t.TempDir(), "repo"))

	trackedFile := filepath.Join(repoPath, "tracked.txt")
	if err := os.WriteFile(trackedFile, []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to write tracked file: %v", err)
	}
	runGit(t, repoPath, "add", "tracked.txt")
	runGit(t, repoPath, "commit", "-m", "add tracked file")

	if err := os.WriteFile(trackedFile, []byte("modified"), 0o644); err != nil {
		t.Fatalf("failed to modify tracked file: %v", err)
	}

	dirty, err := isDirty(repoPath)
	if err != nil {
		t.Fatalf("isDirty returned error: %v", err)
	}
	if !dirty {
		t.Errorf("expected repo with unstaged modification to report dirty=true")
	}
}
