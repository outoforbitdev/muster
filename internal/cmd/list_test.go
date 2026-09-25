package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// withTempHome points HOME at a fresh temp dir for the duration of the test.
func withTempHome(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	origHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Fatalf("failed to restore HOME: %v", err)
		}
	})

	return tempDir
}

// captureStdout runs fn and returns whatever it wrote to os.Stdout.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	fnErr := fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe writer: %v", err)
	}
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
	}

	return buf.String(), fnErr
}

func TestListCommand_Workspaces(t *testing.T) {
	tempDir := withTempHome(t)
	root := filepath.Join(tempDir, ".muster", "workspaces")

	if err := os.MkdirAll(filepath.Join(root, "flat-ws", "repo-a", ".git"), 0755); err != nil {
		t.Fatalf("failed to create fake repo: %v", err)
	}

	t.Run("no target defaults to workspaces", func(t *testing.T) {
		out, err := captureStdout(t, func() error {
			cmd := listCmd
			return cmd.RunE(cmd, []string{})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Contains([]byte(out), []byte("flat-ws")) {
			t.Errorf("expected output to contain %q, got %q", "flat-ws", out)
		}
		if !bytes.Contains([]byte(out), []byte("repo-a")) {
			t.Errorf("expected output to contain %q, got %q", "repo-a", out)
		}
	})

	t.Run("explicit workspaces target", func(t *testing.T) {
		out, err := captureStdout(t, func() error {
			cmd := listCmd
			return cmd.RunE(cmd, []string{"workspaces"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Contains([]byte(out), []byte("flat-ws")) {
			t.Errorf("expected output to contain %q, got %q", "flat-ws", out)
		}
	})

	t.Run("no workspaces found", func(t *testing.T) {
		withTempHome(t)

		out, err := captureStdout(t, func() error {
			cmd := listCmd
			return cmd.RunE(cmd, []string{"workspaces"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Contains([]byte(out), []byte("No workspaces found")) {
			t.Errorf("expected 'No workspaces found' message, got %q", out)
		}
	})
}

func TestListCommand_Stacks(t *testing.T) {
	tempDir := withTempHome(t)
	configDir := filepath.Join(tempDir, ".config", "muster")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configJSON := `{
		"stacks": {
			"backend": {
				"description": "Backend services",
				"repos": [{"url": "git@github.com:org/api.git"}]
			}
		},
		"defaults": {"checkoutBranchOnLaunch": true}
	}`
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	out, err := captureStdout(t, func() error {
		cmd := listCmd
		return cmd.RunE(cmd, []string{"stacks"})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains([]byte(out), []byte("backend: Backend services")) {
		t.Errorf("expected output to contain stack name and description, got %q", out)
	}
	if !bytes.Contains([]byte(out), []byte("git@github.com:org/api.git")) {
		t.Errorf("expected output to contain repo URL, got %q", out)
	}
}

func TestListCommand_UnknownTarget(t *testing.T) {
	withTempHome(t)

	cmd := listCmd
	err := cmd.RunE(cmd, []string{"bogus"})
	if err == nil {
		t.Fatal("expected error for unknown target")
	}
}
