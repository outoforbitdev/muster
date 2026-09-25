package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

// mustInitRepo creates a bare-minimum ".git" directory to simulate a cloned repo.
func mustInitRepo(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(path, ".git"), 0755); err != nil {
		t.Fatalf("failed to create fake repo at %q: %v", path, err)
	}
}

func TestListWorkspaces(t *testing.T) {
	t.Run("no workspaces root", func(t *testing.T) {
		withTempHome(t)

		got, err := ListWorkspaces()
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

		got, err := ListWorkspaces()
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

		got, err := ListWorkspaces()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []Info{
			{Name: "flat-ws", Stack: "", Path: filepath.Join(root, "flat-ws"), Repos: []string{"repo-a", "repo-b"}},
			{Name: "other-ws", Stack: "backend", Path: filepath.Join(root, "backend", "other-ws"), Repos: []string{"repo-d"}},
			{Name: "stacked-ws", Stack: "backend", Path: filepath.Join(root, "backend", "stacked-ws"), Repos: []string{"repo-c"}},
		}

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

		got, err := ListWorkspaces()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected no workspaces, got %v", got)
		}
	})
}
