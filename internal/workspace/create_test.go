package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/outoforbitdev/muster/internal/config"
)

func TestGetRepoPath(t *testing.T) {
	tests := []struct {
		name      string
		directory string
		url       string
		expected  string
	}{
		{
			name:      "with custom directory",
			directory: "types",
			url:       "https://github.com/test/shared-types",
			expected:  "/workspace/types",
		},
		{
			name:      "without custom directory",
			directory: "",
			url:       "https://github.com/test/api",
			expected:  "/workspace/api",
		},
		{
			name:      "without custom directory with .git suffix",
			directory: "",
			url:       "git@github.com:test/api.git",
			expected:  "/workspace/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rtc := RepoToClone{
				Directory: tt.directory,
				URL:       tt.url,
			}
			result := getRepoPath("/workspace", &rtc)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestShouldCheckoutBranch(t *testing.T) {
	tests := []struct {
		name                   string
		checkoutBranchOnLaunch bool
		cliBranch              string
		noBranch               bool
		expected               bool
	}{
		{"no-branch flag takes precedence", true, "main", true, false},
		{"cli branch takes precedence", true, "feature", false, true},
		{"config setting respected", false, "", false, false},
		{"default is true", true, "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Defaults: config.Defaults{
					CheckoutBranchOnLaunch: tt.checkoutBranchOnLaunch,
				},
			}
			result := shouldCheckoutBranch(cfg, tt.cliBranch, tt.noBranch)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCheckoutBranchInRepoSetsUpstreamToDefaultBranch(t *testing.T) {
	repoPath := initRemoteWithClone(t)

	if err := checkoutBranchInRepo(repoPath, "feature"); err != nil {
		t.Fatalf("checkoutBranchInRepo failed: %v", err)
	}

	upstream, err := runGitCommand(repoPath, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "feature@{u}")
	if err != nil {
		t.Fatalf("expected new branch to have an upstream configured, got error: %v\n%s", err, upstream)
	}
	if got := strings.TrimSpace(upstream); got != "origin/main" {
		t.Errorf("expected upstream origin/main, got %q", got)
	}
}

func TestRunBootstrapScriptExecutesCommandInRepoDirectory(t *testing.T) {
	repoPath := initRemoteWithClone(t)

	if err := runBootstrapScript(repoPath, "touch bootstrap-marker.txt"); err != nil {
		t.Fatalf("runBootstrapScript returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repoPath, "bootstrap-marker.txt")); err != nil {
		t.Errorf("expected bootstrap script to create marker file in repo directory: %v", err)
	}
}

func TestRunBootstrapScriptReturnsErrorOnFailure(t *testing.T) {
	repoPath := initRemoteWithClone(t)

	err := runBootstrapScript(repoPath, "exit 1")
	if err == nil {
		t.Fatal("expected error from failing bootstrap script, got nil")
	}
}

func TestCreateWorkspaceRunsBootstrapScript(t *testing.T) {
	remotePath := initBareRemote(t)
	t.Setenv("HOME", t.TempDir())

	cfg := &config.Config{
		Stacks: map[string]config.Stack{
			"test-stack": {
				Repos: []config.Repo{
					{
						URL:             remotePath,
						Directory:       "repo",
						BootstrapScript: "touch bootstrap-marker.txt",
					},
				},
			},
		},
	}

	if err := CreateWorkspace(cfg, "my-workspace", "test-stack", nil, "", true); err != nil {
		t.Fatalf("CreateWorkspace returned error: %v", err)
	}

	markerPath := filepath.Join(WorkspacePath("test-stack", "my-workspace"), "repo", "bootstrap-marker.txt")
	if _, err := os.Stat(markerPath); err != nil {
		t.Errorf("expected bootstrap script to run in cloned repo: %v", err)
	}
}

func TestCreateWorkspaceWarnsAndContinuesWhenBootstrapScriptFails(t *testing.T) {
	remotePath := initBareRemote(t)
	t.Setenv("HOME", t.TempDir())

	cfg := &config.Config{
		Stacks: map[string]config.Stack{
			"test-stack": {
				Repos: []config.Repo{
					{
						URL:             remotePath,
						Directory:       "repo",
						BootstrapScript: "exit 1",
					},
				},
			},
		},
	}

	if err := CreateWorkspace(cfg, "my-workspace", "test-stack", nil, "", true); err != nil {
		t.Fatalf("expected CreateWorkspace to succeed despite bootstrap script failure, got: %v", err)
	}

	repoPath := filepath.Join(WorkspacePath("test-stack", "my-workspace"), "repo")
	if _, err := os.Stat(repoPath); err != nil {
		t.Errorf("expected repo to still be cloned: %v", err)
	}
}

func TestDetermineBranch(t *testing.T) {
	tests := []struct {
		name           string
		cliBranch      string
		repoTemplate   string
		globalTemplate string
		expected       string
	}{
		{"cli branch takes precedence", "cli-branch", "repo-template", "global-template", "cli-branch"},
		{"repo template if no cli", "", "repo-template", "global-template", "repo-template"},
		{"global template if no repo", "", "", "global-template", "global-template"},
		{"empty if nothing set", "", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineBranch(tt.cliBranch, tt.repoTemplate, tt.globalTemplate)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
