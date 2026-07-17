package workspace

import (
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
		name                       string
		checkoutBranchOnLaunch     bool
		cliBranch                  string
		noBranch                   bool
		expected                   bool
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

func TestDetermineBranch(t *testing.T) {
	tests := []struct {
		name         string
		cliBranch    string
		repoTemplate string
		globalTemplate string
		expected     string
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
