package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/outoforbitdev/muster/internal/config"
)

// writeConfig writes a minimal config.json under tempDir/.config/muster
// declaring the given stack names (each with one placeholder repo).
func writeConfig(t *testing.T, tempDir string, stackNames ...string) {
	t.Helper()

	configDir := filepath.Join(tempDir, ".config", "muster")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	stacks := make(map[string]config.Stack, len(stackNames))
	for _, name := range stackNames {
		stacks[name] = config.Stack{Repos: []config.Repo{{URL: "git@github.com:org/" + name + ".git"}}}
	}
	cfg := config.Config{Stacks: stacks}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), data, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}

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
	if err := os.MkdirAll(filepath.Join(root, "backend", "stacked-ws", "repo-b", ".git"), 0755); err != nil {
		t.Fatalf("failed to create fake repo: %v", err)
	}
	writeConfig(t, tempDir, "backend")

	t.Run("no target defaults to workspaces, concise by default", func(t *testing.T) {
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
		if bytes.Contains([]byte(out), []byte("repo-a")) {
			t.Errorf("expected concise output to omit repos, got %q", out)
		}
	})

	t.Run("explicit workspaces target, concise", func(t *testing.T) {
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

	t.Run("--all renders a table with repos and stack membership", func(t *testing.T) {
		out, err := captureStdout(t, func() error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			return listWorkspaces(cfg, true)
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) == 0 || !strings.Contains(lines[0], "WORKSPACE") || !strings.Contains(lines[0], "STACK") || !strings.Contains(lines[0], "REPO") {
			t.Fatalf("expected header row with WORKSPACE/STACK/REPO, got %q", out)
		}

		var flatRow, stackedRow string
		for _, line := range lines[1:] {
			if strings.Contains(line, "flat-ws") {
				flatRow = line
			}
			if strings.Contains(line, "stacked-ws") {
				stackedRow = line
			}
		}
		if !strings.Contains(flatRow, "repo-a") {
			t.Errorf("expected flat-ws row to contain repo-a, got %q", flatRow)
		}
		if !strings.Contains(stackedRow, "backend") || !strings.Contains(stackedRow, "repo-b") {
			t.Errorf("expected stacked-ws row to contain stack backend and repo-b, got %q", stackedRow)
		}
	})

	t.Run("no workspaces found", func(t *testing.T) {
		tempDir := withTempHome(t)
		writeConfig(t, tempDir)

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
				"repos": [
					{
						"url": "git@github.com:org/api.git",
						"description": "This is a fairly long description that should wrap across multiple lines in the table"
					},
					{"url": "git@github.com:org/worker.git"}
				]
			}
		},
		"defaults": {"checkoutBranchOnLaunch": true}
	}`
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	t.Run("concise by default", func(t *testing.T) {
		out, err := captureStdout(t, func() error {
			cmd := listCmd
			return cmd.RunE(cmd, []string{"stacks"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Contains([]byte(out), []byte("backend")) {
			t.Errorf("expected output to contain stack name, got %q", out)
		}
		if bytes.Contains([]byte(out), []byte("Backend services")) {
			t.Errorf("expected concise output to omit description, got %q", out)
		}
		if bytes.Contains([]byte(out), []byte("git@github.com:org/api.git")) {
			t.Errorf("expected concise output to omit repos, got %q", out)
		}
	})

	t.Run("--all renders a table with repos and wrapped descriptions", func(t *testing.T) {
		out, err := captureStdout(t, func() error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			return listStacks(cfg, true)
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) == 0 || !strings.Contains(lines[0], "STACK") || !strings.Contains(lines[0], "REPO") || !strings.Contains(lines[0], "DESCRIPTION") {
			t.Fatalf("expected header row with STACK/REPO/DESCRIPTION, got %q", out)
		}

		if !strings.Contains(out, "backend") {
			t.Errorf("expected output to contain stack name, got %q", out)
		}
		if !strings.Contains(out, "api") {
			t.Errorf("expected output to contain repo name derived from URL, got %q", out)
		}
		if !strings.Contains(out, "worker") {
			t.Errorf("expected output to contain second repo name, got %q", out)
		}
		// header + api's 2 wrapped description lines + worker's 1 line = 4.
		if len(lines) != 4 {
			t.Errorf("expected the long description to wrap across multiple rows, got %d lines: %q", len(lines), out)
		}
	})
}

func TestListCommand_UnknownTarget(t *testing.T) {
	withTempHome(t)

	cmd := listCmd
	err := cmd.RunE(cmd, []string{"bogus"})
	if err == nil {
		t.Fatal("expected error for unknown target")
	}
}
