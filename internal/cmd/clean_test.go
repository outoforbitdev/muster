package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/outoforbitdev/muster/internal/workspace"
)

// initCleanRepo creates a bare "remote" repo with an initial commit on main
// and clones it into path, returning path.
func initCleanRepo(t *testing.T, path string) string {
	t.Helper()

	remotePath := filepath.Join(t.TempDir(), "remote.git")
	runCleanGit(t, "", "init", "--bare", "-b", "main", remotePath)

	seedPath := filepath.Join(t.TempDir(), "seed")
	if err := os.MkdirAll(seedPath, 0o755); err != nil {
		t.Fatalf("failed to create seed dir: %v", err)
	}
	runCleanGit(t, seedPath, "init", "-b", "main")
	runCleanGit(t, seedPath, "commit", "--allow-empty", "-m", "initial commit")
	runCleanGit(t, seedPath, "remote", "add", "origin", remotePath)
	runCleanGit(t, seedPath, "push", "origin", "main")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create parent dir: %v", err)
	}
	runCleanGit(t, "", "clone", remotePath, path)

	return path
}

func runCleanGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v (dir=%q) failed: %v\n%s", args, dir, err, out)
	}
}

func TestRunCleanDryRunReportsCleanWorkspaceAsSafe(t *testing.T) {
	tempDir := withTempHome(t)
	writeConfig(t, tempDir)
	root := filepath.Join(tempDir, ".muster", "workspaces")

	initCleanRepo(t, filepath.Join(root, "my-ws", "repo-a"))

	out, err := captureStdout(t, func() error {
		return runClean(nil, false, false, nil)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "my-ws") || !strings.Contains(out, "repo-a") || !strings.Contains(out, "clean") {
		t.Errorf("expected output to report my-ws/repo-a as clean, got %q", out)
	}
	if !strings.Contains(out, "safe to clean") {
		t.Errorf("expected summary mentioning safe-to-clean count, got %q", out)
	}

	if _, statErr := os.Stat(filepath.Join(root, "my-ws")); statErr != nil {
		t.Errorf("dry-run must not delete the workspace, but it's gone: %v", statErr)
	}
}

func TestRunCleanDryRunReportsDirtyWorkspaceAsUnsafe(t *testing.T) {
	tempDir := withTempHome(t)
	writeConfig(t, tempDir)
	root := filepath.Join(tempDir, ".muster", "workspaces")

	repoPath := initCleanRepo(t, filepath.Join(root, "my-ws", "repo-a"))
	if err := os.WriteFile(filepath.Join(repoPath, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}

	out, err := captureStdout(t, func() error {
		return runClean(nil, false, false, nil)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "dirty") {
		t.Errorf("expected output to report dirty status, got %q", out)
	}
	if !strings.Contains(out, "not safe") {
		t.Errorf("expected summary mentioning not-safe count, got %q", out)
	}
}

func TestRunCleanWriteDeletesOnlySafeWorkspaces(t *testing.T) {
	tempDir := withTempHome(t)
	writeConfig(t, tempDir)
	root := filepath.Join(tempDir, ".muster", "workspaces")

	initCleanRepo(t, filepath.Join(root, "safe-ws", "repo-a"))
	dirtyRepoPath := initCleanRepo(t, filepath.Join(root, "dirty-ws", "repo-a"))
	if err := os.WriteFile(filepath.Join(dirtyRepoPath, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}

	_, err := captureStdout(t, func() error {
		return runClean(nil, true, true, nil)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(root, "safe-ws")); !os.IsNotExist(statErr) {
		t.Errorf("expected safe-ws to be deleted, stat err: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, "dirty-ws")); statErr != nil {
		t.Errorf("expected dirty-ws to still exist, got stat err: %v", statErr)
	}
}

func TestRunCleanFiltersToNamedWorkspaces(t *testing.T) {
	tempDir := withTempHome(t)
	writeConfig(t, tempDir)
	root := filepath.Join(tempDir, ".muster", "workspaces")

	initCleanRepo(t, filepath.Join(root, "ws-one", "repo-a"))
	initCleanRepo(t, filepath.Join(root, "ws-two", "repo-a"))

	out, err := captureStdout(t, func() error {
		return runClean([]string{"ws-one"}, false, false, nil)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ws-one") {
		t.Errorf("expected output to include ws-one, got %q", out)
	}
	if strings.Contains(out, "ws-two") {
		t.Errorf("expected output to exclude ws-two, got %q", out)
	}
}

func TestRunCleanUnknownWorkspaceNameErrors(t *testing.T) {
	tempDir := withTempHome(t)
	writeConfig(t, tempDir)
	root := filepath.Join(tempDir, ".muster", "workspaces")
	initCleanRepo(t, filepath.Join(root, "ws-one", "repo-a"))

	err := runClean([]string{"does-not-exist"}, false, false, nil)
	if err == nil {
		t.Fatal("expected error for unknown workspace name")
	}
}

func TestRunCleanDivergedAndMergedIsSafe(t *testing.T) {
	tempDir := withTempHome(t)
	writeConfig(t, tempDir)
	root := filepath.Join(tempDir, ".muster", "workspaces")

	repoPath := initCleanRepo(t, filepath.Join(root, "my-ws", "repo-a"))
	runCleanGit(t, repoPath, "checkout", "-b", "feature")
	runCleanGit(t, repoPath, "commit", "--allow-empty", "-m", "feature work")

	fakeChecker := workspace.MergeChecker(func(repoPath, branch string) (bool, error) {
		return true, nil
	})

	out, err := captureStdout(t, func() error {
		return runClean(nil, false, false, fakeChecker)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "merged") {
		t.Errorf("expected output to report merged status, got %q", out)
	}
	if !strings.Contains(out, "1 workspace") || !strings.Contains(out, "safe to clean") {
		t.Errorf("expected summary reporting 1 safe workspace, got %q", out)
	}
}
