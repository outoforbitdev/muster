package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func withTempHome(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "muster-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	return tmpDir
}

func TestWorkspacePath(t *testing.T) {
	tmpDir := withTempHome(t)
	root := filepath.Join(tmpDir, ".muster", "workspaces")

	tests := []struct {
		name       string
		stackNames []string
		workspace  string
		want       string
	}{
		{"no stack", []string{}, "my-ws", filepath.Join(root, "my-ws")},
		{"single stack", []string{"backend"}, "my-ws", filepath.Join(root, "backend", "my-ws")},
		{"uses primary stack", []string{"backend", "frontend"}, "my-ws", filepath.Join(root, "backend", "my-ws")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WorkspacePath(tt.stackNames, tt.workspace)
			if got != tt.want {
				t.Errorf("WorkspacePath(%v, %q) = %q, want %q", tt.stackNames, tt.workspace, got, tt.want)
			}
		})
	}
}

func TestFindWorkspacePath(t *testing.T) {
	tmpDir := withTempHome(t)
	root := filepath.Join(tmpDir, ".muster", "workspaces")

	mustMkdir := func(path string) {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("failed to create dir %q: %v", path, err)
		}
	}

	mustMkdir(filepath.Join(root, "flat-ws"))
	mustMkdir(filepath.Join(root, "backend", "stacked-ws"))
	mustMkdir(filepath.Join(root, "backend", "ambiguous-ws"))
	mustMkdir(filepath.Join(root, "frontend", "ambiguous-ws"))

	t.Run("finds flat workspace", func(t *testing.T) {
		got, err := FindWorkspacePath("", "flat-ws")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join(root, "flat-ws"); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("finds stacked workspace without stack flag", func(t *testing.T) {
		got, err := FindWorkspacePath("", "stacked-ws")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join(root, "backend", "stacked-ws"); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("finds stacked workspace with stack flag", func(t *testing.T) {
		got, err := FindWorkspacePath("backend", "stacked-ws")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join(root, "backend", "stacked-ws"); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("errors on ambiguous workspace without stack flag", func(t *testing.T) {
		_, err := FindWorkspacePath("", "ambiguous-ws")
		if err == nil {
			t.Fatal("expected error for ambiguous workspace")
		}
	})

	t.Run("resolves ambiguous workspace with stack flag", func(t *testing.T) {
		got, err := FindWorkspacePath("frontend", "ambiguous-ws")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := filepath.Join(root, "frontend", "ambiguous-ws"); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("errors on missing workspace", func(t *testing.T) {
		_, err := FindWorkspacePath("", "does-not-exist")
		if err == nil {
			t.Fatal("expected error for missing workspace")
		}
	})

	t.Run("errors on missing workspace under given stack", func(t *testing.T) {
		_, err := FindWorkspacePath("backend", "does-not-exist")
		if err == nil {
			t.Fatal("expected error for missing workspace")
		}
	})
}
