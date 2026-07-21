package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/outoforbitdev/muster/internal/config"
)

func TestLaunchWorkspaceExisting(t *testing.T) {
	// Create a temporary workspace directory
	tmpDir, err := os.MkdirTemp("", "muster-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	workspacePath := filepath.Join(tmpDir, ".muster", "test-ws")
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	// Mock HOME to use our temp directory
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	cfg := &config.Config{
		Stacks: make(map[string]config.Stack),
		Defaults: config.Defaults{
			CheckoutBranchOnLaunch: false,
		},
	}

	path, err := LaunchWorkspace(cfg, "test-ws", []string{}, []string{}, "", false)
	if err != nil {
		t.Fatalf("LaunchWorkspace failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".muster", "test-ws")
	if path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, path)
	}
}

func TestLaunchWorkspaceNonExistent(t *testing.T) {
	// Create a temporary directory for config
	tmpDir, err := os.MkdirTemp("", "muster-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Mock HOME to use our temp directory
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	cfg := &config.Config{
		Stacks: make(map[string]config.Stack),
		Defaults: config.Defaults{
			CheckoutBranchOnLaunch: false,
		},
	}

	// Try to launch non-existent workspace with no repos
	// This should fail because no repos are specified
	_, err = LaunchWorkspace(cfg, "test-ws", []string{}, []string{}, "", false)
	if err == nil {
		t.Error("expected error for workspace with no repos")
	}
	if err.Error() != "failed to create workspace: no repos specified" {
		t.Errorf("unexpected error: %v", err)
	}
}
