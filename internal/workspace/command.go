package workspace

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/outoforbitdev/muster/internal/config"
)

// ShouldLaunchAgent determines whether to run the agent command.
// Precedence: --no-agent > --agent > cfg.Defaults.LaunchAgent (default true).
func ShouldLaunchAgent(cfg *config.Config, agentFlag, noAgentFlag bool) bool {
	if noAgentFlag {
		return false
	}
	if agentFlag {
		return true
	}
	return cfg.Defaults.LaunchAgent == nil || *cfg.Defaults.LaunchAgent
}

// ShouldLaunchEditor determines whether to run the editor command.
// Precedence: --no-editor > --editor > cfg.Defaults.LaunchEditor (default false).
func ShouldLaunchEditor(cfg *config.Config, editorFlag, noEditorFlag bool) bool {
	if noEditorFlag {
		return false
	}
	if editorFlag {
		return true
	}
	return cfg.Defaults.LaunchEditor
}

// runCommand substitutes the template and executes it via the shell so
// user-configured commands can include arguments/flags naturally.
func runCommand(commandTemplate, workspacePath string) error {
	command := config.SubstituteWorkspaceDirectory(commandTemplate, workspacePath)
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workspacePath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

// LaunchAgent runs the configured agentCommand for the workspace.
func LaunchAgent(cfg *config.Config, workspacePath string) error {
	return runCommand(cfg.GetAgentCommand(), workspacePath)
}

// LaunchEditor runs the configured editorCommand for the workspace.
func LaunchEditor(cfg *config.Config, workspacePath string) error {
	return runCommand(cfg.GetEditorCommand(), workspacePath)
}
