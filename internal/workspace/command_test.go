package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/outoforbitdev/muster/internal/config"
)

func TestShouldLaunchAgent(t *testing.T) {
	tests := []struct {
		name              string
		launchAgentConfig *bool
		agentFlag         bool
		noAgentFlag       bool
		expected          bool
	}{
		{"no-agent flag takes precedence", nil, true, true, false},
		{"agent flag when config unset", nil, true, false, true},
		{"no-agent flag when config unset", nil, false, true, false},
		{"config true, no flags", boolPtr(true), false, false, true},
		{"config false, no flags", boolPtr(false), false, false, false},
		{"config nil defaults to true", nil, false, false, true},
		{"agent flag overrides config false", boolPtr(false), true, false, true},
		{"no-agent flag overrides config true", boolPtr(true), false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Defaults: config.Defaults{
					LaunchAgent: tt.launchAgentConfig,
				},
			}
			result := ShouldLaunchAgent(cfg, tt.agentFlag, tt.noAgentFlag)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestShouldLaunchEditor(t *testing.T) {
	tests := []struct {
		name         string
		launchEditor bool
		editorFlag   bool
		noEditorFlag bool
		expected     bool
	}{
		{"no-editor flag takes precedence", true, true, true, false},
		{"editor flag when config false", false, true, false, true},
		{"no-editor flag when config true", true, false, true, false},
		{"config true, no flags", true, false, false, true},
		{"config false, no flags", false, false, false, false},
		{"editor flag overrides config false", false, true, false, true},
		{"no-editor flag overrides config true", true, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Defaults: config.Defaults{
					LaunchEditor: tt.launchEditor,
				},
			}
			result := ShouldLaunchEditor(cfg, tt.editorFlag, tt.noEditorFlag)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestLaunchAgent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "muster-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	cfg := &config.Config{
		Defaults: config.Defaults{
			AgentCommand: "touch " + testFile,
		},
	}

	if err := LaunchAgent(cfg, "test-ws", tmpDir); err != nil {
		t.Errorf("LaunchAgent failed: %v", err)
	}

	if _, err := os.Stat(testFile); err != nil {
		t.Errorf("expected test file to be created, got error: %v", err)
	}
}

func TestLaunchAgentWithSubstitution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "muster-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "my-ws-test.txt")
	cfg := &config.Config{
		Defaults: config.Defaults{
			AgentCommand: "touch " + tmpDir + "/{workspace}-test.txt",
		},
	}

	if err := LaunchAgent(cfg, "my-ws", tmpDir); err != nil {
		t.Errorf("LaunchAgent failed: %v", err)
	}

	if _, err := os.Stat(testFile); err != nil {
		t.Errorf("expected test file to be created with workspace substitution, got error: %v", err)
	}
}
