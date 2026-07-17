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
