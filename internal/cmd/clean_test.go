package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/outoforbitdev/muster/internal/workspace"
)

// ansiEscapePattern matches ANSI color escape sequences (e.g. "\x1b[32m").
var ansiEscapePattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

// stripANSI removes ANSI color escape sequences, leaving only what's
// actually visible when the output is rendered in a terminal.
func stripANSI(s string) string {
	return ansiEscapePattern.ReplaceAllString(s, "")
}

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
	configureGitIdentityForTest(t, seedPath)
	runCleanGit(t, seedPath, "commit", "--allow-empty", "-m", "initial commit")
	runCleanGit(t, seedPath, "remote", "add", "origin", remotePath)
	runCleanGit(t, seedPath, "push", "origin", "main")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create parent dir: %v", err)
	}
	runCleanGit(t, "", "clone", remotePath, path)
	configureGitIdentityForTest(t, path)

	return path
}

// configureGitIdentityForTest sets a local git user.name/user.email in dir,
// so commits work in environments (like CI runners) with no global git
// identity configured.
func configureGitIdentityForTest(t *testing.T, dir string) {
	t.Helper()
	runCleanGit(t, dir, "config", "user.email", "test@example.com")
	runCleanGit(t, dir, "config", "user.name", "Test")
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

func TestRunCleanColorsSafeWorkspaceGreen(t *testing.T) {
	withoutNoColorEnv(t)
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

	if !strings.Contains(out, ansiGreen+"my-ws") {
		t.Errorf("expected workspace name to be colored green, got %q", out)
	}
	if !strings.Contains(out, ansiGreen+"repo-a") {
		t.Errorf("expected repo name to be colored green, got %q", out)
	}
	if strings.Contains(out, ansiRed) {
		t.Errorf("expected no red in output for an all-clean workspace, got %q", out)
	}
}

func TestRunCleanColorsDirtyRepoRed(t *testing.T) {
	withoutNoColorEnv(t)
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

	if !strings.Contains(out, ansiRed+"my-ws") {
		t.Errorf("expected workspace name to be colored red since a repo is dirty, got %q", out)
	}
	if !strings.Contains(out, ansiRed+"repo-a") {
		t.Errorf("expected dirty repo name to be colored red, got %q", out)
	}
}

func TestRunCleanRespectsNoColorEnv(t *testing.T) {
	withNoColorEnv(t, "1")
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

	if strings.Contains(out, ansiGreen) || strings.Contains(out, ansiRed) {
		t.Errorf("expected NO_COLOR to suppress all coloring, got %q", out)
	}
}

func TestRunCleanWrapsLongStateAcrossRowsEvenWithLongWorkspaceAndRepoNames(t *testing.T) {
	tempDir := withTempHome(t)
	writeConfig(t, tempDir)
	root := filepath.Join(tempDir, ".muster", "workspaces")

	// A long workspace/repo name eats into the terminal width available to
	// the STATE column; the wrap width must account for that, or the
	// terminal itself will re-wrap the row, breaking mid-word.
	repoPath := initCleanRepo(t, filepath.Join(root, "22-create-guideline-script", "reusable-workflows-library"))
	runCleanGit(t, repoPath, "checkout", "-b", "feature")
	runCleanGit(t, repoPath, "commit", "--allow-empty", "-m", "feature work")

	out, err := captureStdout(t, func() error {
		return runClean(nil, false, false, nil)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	var longestVisibleLen int
	var longestLine string
	for _, line := range lines {
		if visible := len(stripANSI(line)); visible > longestVisibleLen {
			longestVisibleLen = visible
			longestLine = line
		}
	}
	// A common terminal width; the full rendered row (all columns, tab
	// expansion included) must fit within it so the terminal never has to
	// re-wrap a row itself. ANSI color codes are invisible to the terminal,
	// so they're stripped before measuring.
	const commonTerminalWidth = 80
	if longestVisibleLen > commonTerminalWidth {
		t.Errorf("expected rendered row to fit within %d visible columns, got %d: %q", commonTerminalWidth, longestVisibleLen, longestLine)
	}
}

func TestRunCleanSkipsNonGitDirectoryWithoutBlockingSafety(t *testing.T) {
	withoutNoColorEnv(t)
	tempDir := withTempHome(t)
	writeConfig(t, tempDir)
	root := filepath.Join(tempDir, ".muster", "workspaces")

	initCleanRepo(t, filepath.Join(root, "my-ws", "repo-a"))
	if err := os.MkdirAll(filepath.Join(root, "my-ws", "broken-clone", "src"), 0o755); err != nil {
		t.Fatalf("failed to create non-git directory: %v", err)
	}

	out, err := captureStdout(t, func() error {
		return runClean(nil, false, false, nil)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var skippedLine string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "not a git repository") {
			skippedLine = line
		}
	}
	if skippedLine == "" {
		t.Fatalf("expected output to explain the non-git directory, got %q", out)
	}
	if strings.Contains(skippedLine, ansiRed) {
		t.Errorf("expected the non-git directory's row not to be colored red, got %q", skippedLine)
	}
	if !strings.Contains(out, "1 workspace(s) safe to clean") {
		t.Errorf("expected the workspace to still be safe despite the skipped non-git directory, got %q", out)
	}
}

func TestRunCleanWorkspaceWithNoReposIsNotSafe(t *testing.T) {
	tempDir := withTempHome(t)
	writeConfig(t, tempDir)
	root := filepath.Join(tempDir, ".muster", "workspaces")

	if err := os.MkdirAll(filepath.Join(root, "empty-ws"), 0o755); err != nil {
		t.Fatalf("failed to create empty workspace dir: %v", err)
	}

	out, err := captureStdout(t, func() error {
		return runClean(nil, false, false, nil)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "empty-ws") || !strings.Contains(out, "no repos found") {
		t.Errorf("expected output to report empty-ws has no repos, got %q", out)
	}
	if !strings.Contains(out, "0 workspace(s) safe to clean") {
		t.Errorf("expected a workspace with no repos to be reported as not safe, got %q", out)
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
