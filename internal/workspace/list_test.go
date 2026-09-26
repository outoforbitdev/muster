package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/outoforbitdev/muster/internal/config"
)

// mustInitRepo creates a bare-minimum ".git" directory to simulate a cloned repo.
func mustInitRepo(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(path, ".git"), 0755); err != nil {
		t.Fatalf("failed to create fake repo at %q: %v", path, err)
	}
}

func stacksConfig(stackNames ...string) *config.Config {
	stacks := make(map[string]config.Stack, len(stackNames))
	for _, name := range stackNames {
		stacks[name] = config.Stack{}
	}
	return &config.Config{Stacks: stacks}
}

func TestListWorkspaces(t *testing.T) {
	t.Run("no workspaces root", func(t *testing.T) {
		withTempHome(t)

		got, err := ListWorkspaces(stacksConfig())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected no workspaces, got %v", got)
		}
	})

	t.Run("empty workspaces root", func(t *testing.T) {
		tmpDir := withTempHome(t)
		root := filepath.Join(tmpDir, ".muster", "workspaces")
		if err := os.MkdirAll(root, 0755); err != nil {
			t.Fatalf("failed to create root: %v", err)
		}

		got, err := ListWorkspaces(stacksConfig())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected no workspaces, got %v", got)
		}
	})

	t.Run("flat and stacked workspaces with repos", func(t *testing.T) {
		tmpDir := withTempHome(t)
		root := filepath.Join(tmpDir, ".muster", "workspaces")

		// Flat workspace with two repos.
		mustInitRepo(t, filepath.Join(root, "flat-ws", "repo-a"))
		mustInitRepo(t, filepath.Join(root, "flat-ws", "repo-b"))

		// Stacked workspace with one repo.
		mustInitRepo(t, filepath.Join(root, "backend", "stacked-ws", "repo-c"))

		// A second stacked workspace under the same stack.
		mustInitRepo(t, filepath.Join(root, "backend", "other-ws", "repo-d"))

		got, err := ListWorkspaces(stacksConfig("backend"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []Info{
			{Name: "flat-ws", Stack: "", Path: filepath.Join(root, "flat-ws"), Repos: []string{"repo-a", "repo-b"}},
			{Name: "other-ws", Stack: "backend", Path: filepath.Join(root, "backend", "other-ws"), Repos: []string{"repo-d"}},
			{Name: "stacked-ws", Stack: "backend", Path: filepath.Join(root, "backend", "stacked-ws"), Repos: []string{"repo-c"}},
		}

		assertWorkspaces(t, got, want)
	})

	t.Run("flat workspace with a single, broken repo (no .git) is not exploded into a stack", func(t *testing.T) {
		tmpDir := withTempHome(t)
		root := filepath.Join(tmpDir, ".muster", "workspaces")

		// Simulates a workspace whose repo clone lost or never got its
		// .git directory. "backend" is a configured stack name, so it
		// must still be treated as a stack container, while this flat
		// workspace (whose name matches no stack) must be treated as a
		// single workspace, not exploded into per-subdirectory entries.
		if err := os.MkdirAll(filepath.Join(root, "broken-ws", "repo-a"), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		mustInitRepo(t, filepath.Join(root, "backend", "stacked-ws", "repo-c"))

		got, err := ListWorkspaces(stacksConfig("backend"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []Info{
			{Name: "broken-ws", Stack: "", Path: filepath.Join(root, "broken-ws"), Repos: []string{"repo-a"}},
			{Name: "stacked-ws", Stack: "backend", Path: filepath.Join(root, "backend", "stacked-ws"), Repos: []string{"repo-c"}},
		}

		assertWorkspaces(t, got, want)
	})

	t.Run("ignores non-workspace files", func(t *testing.T) {
		tmpDir := withTempHome(t)
		root := filepath.Join(tmpDir, ".muster", "workspaces")
		if err := os.MkdirAll(root, 0755); err != nil {
			t.Fatalf("failed to create root: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "README.txt"), []byte("hi"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		got, err := ListWorkspaces(stacksConfig())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected no workspaces, got %v", got)
		}
	})
}

func assertWorkspaces(t *testing.T, got, want []Info) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %d workspaces, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Name != want[i].Name || got[i].Stack != want[i].Stack || got[i].Path != want[i].Path {
			t.Errorf("workspace %d = %+v, want %+v", i, got[i], want[i])
		}
		if len(got[i].Repos) != len(want[i].Repos) {
			t.Errorf("workspace %d repos = %v, want %v", i, got[i].Repos, want[i].Repos)
			continue
		}
		for j := range want[i].Repos {
			if got[i].Repos[j] != want[i].Repos[j] {
				t.Errorf("workspace %d repos = %v, want %v", i, got[i].Repos, want[i].Repos)
			}
		}
	}
}
