package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultAgentCommand  = "claude --name {workspace}"
	DefaultEditorCommand = "code {workspaceDirectory}"
)

type Repo struct {
	URL                  string `json:"url"`
	TemplateBranchSyntax string `json:"templateBranchSyntax,omitempty"`
	Description          string `json:"description,omitempty"`
	Directory            string `json:"directory,omitempty"`
}

type Stack struct {
	Repos       []Repo `json:"repos"`
	Description string `json:"description,omitempty"`
}

type Defaults struct {
	CheckoutBranchOnLaunch bool   `json:"checkoutBranchOnLaunch"`
	TemplateBranchSyntax   string `json:"templateBranchSyntax,omitempty"`
	AgentCommand           string `json:"agentCommand,omitempty"`
	EditorCommand          string `json:"editorCommand,omitempty"`
	LaunchAgent            *bool  `json:"launchAgent,omitempty"`
	LaunchEditor           bool   `json:"launchEditor,omitempty"`
}

type Config struct {
	Stacks   map[string]Stack `json:"stacks"`
	Defaults Defaults         `json:"defaults"`
}

// Load reads the config from ~/.config/muster/config.json.
func Load() (*Config, error) {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "muster", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s: use 'muster init' to create it", configPath)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that the config is valid.
func (c *Config) Validate() error {
	if c.Stacks == nil {
		c.Stacks = make(map[string]Stack)
	}

	for stackName, stack := range c.Stacks {
		if len(stack.Repos) == 0 {
			return fmt.Errorf("stack %q has no repos", stackName)
		}

		for i, repo := range stack.Repos {
			if repo.URL == "" {
				return fmt.Errorf("stack %q repo %d has no URL", stackName, i)
			}

			if err := validateGitURL(repo.URL); err != nil {
				return fmt.Errorf("stack %q repo %d: invalid URL: %w", stackName, i, err)
			}
		}
	}

	return nil
}

// validateGitURL checks that the URL is a valid Git URL (SSH or HTTPS).
func validateGitURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL is empty")
	}

	// Check for git@ SSH format
	if len(urlStr) >= 4 && urlStr[0:4] == "git@" {
		return nil
	}

	// Check for git:// scheme
	if len(urlStr) >= 6 && urlStr[0:6] == "git://" {
		return nil
	}

	// Try to parse as HTTPS/HTTP URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("not a valid URL: %w", err)
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("URL scheme must be https, http, or SSH (git@...)")
	}

	if u.Host == "" {
		return fmt.Errorf("URL has no host")
	}

	return nil
}

// GetStack returns a stack by name, or nil if not found.
func (c *Config) GetStack(name string) *Stack {
	stack, ok := c.Stacks[name]
	if !ok {
		return nil
	}
	return &stack
}

// GetAgentCommand returns the agent command, using the configured value or the default.
func (c *Config) GetAgentCommand() string {
	if c.Defaults.AgentCommand != "" {
		return c.Defaults.AgentCommand
	}
	return DefaultAgentCommand
}

// GetEditorCommand returns the editor command, using the configured value or the default.
func (c *Config) GetEditorCommand() string {
	if c.Defaults.EditorCommand != "" {
		return c.Defaults.EditorCommand
	}
	return DefaultEditorCommand
}

// SubstituteCommandTemplate replaces {workspace} and {workspaceDirectory} in a command template.
func SubstituteCommandTemplate(command, workspaceName, workspacePath string) string {
	command = strings.ReplaceAll(command, "{workspace}", workspaceName)
	command = strings.ReplaceAll(command, "{workspaceDirectory}", workspacePath)
	return command
}
