package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/outoforbitdev/muster/internal/config"
)

func TestGenerateCLAUDE(t *testing.T) {
	t.Run("with stack description", func(t *testing.T) {
		stack := &config.Stack{
			Description: "My test stack with api and web",
		}
		repos := map[string]string{
			"api": "~/.workspaces/my-ws/api",
			"web": "~/.workspaces/my-ws/web",
		}

		content := GenerateCLAUDE("my-ws", stack, repos)

		if !strings.Contains(content, "# Workspace: my-ws") {
			t.Error("missing workspace title")
		}
		if !strings.Contains(content, "My test stack with api and web") {
			t.Error("missing stack description")
		}
		if !strings.Contains(content, "## Repos") {
			t.Error("missing repos section")
		}
		if !strings.Contains(content, "**api**") {
			t.Error("missing api repo")
		}
		if !strings.Contains(content, "**web**") {
			t.Error("missing web repo")
		}
	})

	t.Run("without stack description", func(t *testing.T) {
		stack := &config.Stack{
			Description: "",
		}
		repos := map[string]string{
			"api": "~/.workspaces/my-ws/api",
		}

		content := GenerateCLAUDE("my-ws", stack, repos)

		if !strings.Contains(content, "# Workspace: my-ws") {
			t.Error("missing workspace title")
		}
		if !strings.Contains(content, "## Repos") {
			t.Error("missing repos section")
		}
	})

	t.Run("nil stack", func(t *testing.T) {
		repos := map[string]string{
			"api": "~/.workspaces/my-ws/api",
		}

		content := GenerateCLAUDE("my-ws", nil, repos)

		if !strings.Contains(content, "# Workspace: my-ws") {
			t.Error("missing workspace title")
		}
		if !strings.Contains(content, "## Repos") {
			t.Error("missing repos section")
		}
	})
}

func TestWriteCLAUDE(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "muster-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	content := "# Test Content\n\nSome test content"
	if err := WriteCLAUDE(tmpDir, content); err != nil {
		t.Fatalf("WriteCLAUDE failed: %v", err)
	}

	// Check that the file was created
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}

	// Check content
	readContent, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}
	if string(readContent) != content {
		t.Errorf("content mismatch: expected %q, got %q", content, string(readContent))
	}
}
