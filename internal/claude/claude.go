package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/outoforbitdev/muster/internal/config"
)

// GenerateCLAUDE generates the content of a CLAUDE.md file for a workspace.
// repos maps repo names to their paths relative to the workspace root.
func GenerateCLAUDE(workspace string, stack *config.Stack, repos map[string]string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Workspace: %s\n\n", workspace))

	if stack != nil && stack.Description != "" {
		sb.WriteString(stack.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Repos\n\n")
	for repoName, repoPath := range repos {
		sb.WriteString(fmt.Sprintf("- **%s**: `%s`\n", repoName, repoPath))
	}

	return sb.String()
}

// WriteCLAUDE writes a CLAUDE.md file to the workspace root.
func WriteCLAUDE(workspacePath, content string) error {
	claudePath := filepath.Join(workspacePath, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write CLAUDE.md: %w", err)
	}
	return nil
}
