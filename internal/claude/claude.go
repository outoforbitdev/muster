package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/outoforbitdev/muster/internal/config"
)

// RepoInfo holds info about a cloned repo for CLAUDE.md generation.
type RepoInfo struct {
	Name        string
	Path        string
	Description string
}

// GenerateCLAUDE generates the content of a CLAUDE.md file for a workspace.
// repos is a slice of RepoInfo with paths and descriptions.
func GenerateCLAUDE(workspace string, stack *config.Stack, repos []RepoInfo) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Workspace: %s\n\n", workspace)

	if stack != nil && stack.Description != "" {
		sb.WriteString(stack.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Repos\n\n")
	for _, repo := range repos {
		fmt.Fprintf(&sb, "- **%s**: `%s`", repo.Name, repo.Path)
		if repo.Description != "" {
			fmt.Fprintf(&sb, " — %s", repo.Description)
		}
		sb.WriteString("\n")
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
