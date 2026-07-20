package config

import (
	"testing"
)

func TestValidateGitURL(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"https://github.com/yourorg/api", true},
		{"https://github.com/yourorg/api.git", true},
		{"http://github.com/yourorg/api", true},
		{"git@github.com:yourorg/api.git", true},
		{"git@github.com:yourorg/api", true},
		{"git://github.com/yourorg/api.git", true},
		{"", false},
		{"not-a-url", false},
		{"ftp://github.com/yourorg/api", false},
		{"https://", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := validateGitURL(tt.url)
			if tt.valid && err != nil {
				t.Errorf("expected valid URL, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected invalid URL, got no error")
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			Stacks: map[string]Stack{
				"test": {
					Repos: []Repo{
						{
							URL:         "https://github.com/test/repo",
							Description: "Test repo",
						},
					},
				},
			},
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		cfg := &Config{
			Stacks: map[string]Stack{
				"test": {
					Repos: []Repo{
						{
							URL: "not-a-url",
						},
					},
				},
			},
		}
		if err := cfg.Validate(); err == nil {
			t.Error("expected validation error for invalid URL")
		}
	})

	t.Run("empty repo list", func(t *testing.T) {
		cfg := &Config{
			Stacks: map[string]Stack{
				"test": {
					Repos: []Repo{},
				},
			},
		}
		if err := cfg.Validate(); err == nil {
			t.Error("expected validation error for empty repo list")
		}
	})

	t.Run("missing URL", func(t *testing.T) {
		cfg := &Config{
			Stacks: map[string]Stack{
				"test": {
					Repos: []Repo{
						{
							Description: "Missing URL",
						},
					},
				},
			},
		}
		if err := cfg.Validate(); err == nil {
			t.Error("expected validation error for missing URL")
		}
	})
}

func TestGetStack(t *testing.T) {
	cfg := &Config{
		Stacks: map[string]Stack{
			"test": {
				Repos: []Repo{
					{
						URL: "https://github.com/test/repo",
					},
				},
			},
		},
	}

	t.Run("existing stack", func(t *testing.T) {
		stack := cfg.GetStack("test")
		if stack == nil {
			t.Error("expected to find stack")
		}
	})

	t.Run("non-existing stack", func(t *testing.T) {
		stack := cfg.GetStack("nonexistent")
		if stack != nil {
			t.Error("expected nil for non-existing stack")
		}
	})
}

func TestGetAgentCommand(t *testing.T) {
	t.Run("default agent command", func(t *testing.T) {
		cfg := &Config{
			Defaults: Defaults{
				AgentCommand: "",
			},
		}
		if got := cfg.GetAgentCommand(); got != DefaultAgentCommand {
			t.Errorf("expected %q, got %q", DefaultAgentCommand, got)
		}
	})

	t.Run("configured agent command", func(t *testing.T) {
		cfg := &Config{
			Defaults: Defaults{
				AgentCommand: "custom-agent {workspace}",
			},
		}
		if got := cfg.GetAgentCommand(); got != "custom-agent {workspace}" {
			t.Errorf("expected %q, got %q", "custom-agent {workspace}", got)
		}
	})
}

func TestGetEditorCommand(t *testing.T) {
	t.Run("default editor command", func(t *testing.T) {
		cfg := &Config{
			Defaults: Defaults{
				EditorCommand: "",
			},
		}
		if got := cfg.GetEditorCommand(); got != DefaultEditorCommand {
			t.Errorf("expected %q, got %q", DefaultEditorCommand, got)
		}
	})

	t.Run("configured editor command", func(t *testing.T) {
		cfg := &Config{
			Defaults: Defaults{
				EditorCommand: "vim {workspaceDirectory}",
			},
		}
		if got := cfg.GetEditorCommand(); got != "vim {workspaceDirectory}" {
			t.Errorf("expected %q, got %q", "vim {workspaceDirectory}", got)
		}
	})
}

func TestSubstituteCommandTemplate(t *testing.T) {
	tests := []struct {
		command          string
		workspaceName    string
		workspacePath    string
		expected         string
	}{
		{"claude --name {workspace}", "my-ws", "/home/user/.muster/my-ws", "claude --name my-ws"},
		{"code {workspaceDirectory}", "my-ws", "/tmp/test", "code /tmp/test"},
		{"echo {workspace} {workspaceDirectory}", "my-ws", "/path", "echo my-ws /path"},
		{"no substitution", "my-ws", "/path", "no substitution"},
		{"{workspace}-{workspaceDirectory}", "test", "/home/test", "test-/home/test"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			if got := SubstituteCommandTemplate(tt.command, tt.workspaceName, tt.workspacePath); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
