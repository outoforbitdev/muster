package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/outoforbitdev/muster/internal/config"
	"github.com/outoforbitdev/muster/internal/workspace"
)

var (
	launchStacks   []string
	launchRepos    []string
	launchBranch   string
	launchNoBranch bool
)

var launchCmd = &cobra.Command{
	Use:   "launch <workspace>",
	Short: "Launch or create a workspace",
	Long: `Launch an existing workspace or create a new one with the specified repos.

If the workspace already exists, this will open it in Claude Code.
If the workspace doesn't exist, this will:
  1. Create the workspace directory
  2. Clone all specified repos
  3. Checkout configured branches (with template substitution)
  4. Generate a CLAUDE.md file
  5. Launch Claude Code with the workspace`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceName := args[0]

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Launch or create workspace
		workspacePath, err := workspace.LaunchWorkspace(
			cfg,
			workspaceName,
			launchStacks,
			launchRepos,
			launchBranch,
			launchNoBranch,
		)
		if err != nil {
			return fmt.Errorf("failed to launch workspace: %w", err)
		}

		// Check if workspace is new or existing
		claudePath := filepath.Join(workspacePath, "CLAUDE.md")
		isNew := true
		if _, err := os.Stat(claudePath); err == nil {
			isNew = false
		}

		// Launch Claude Code
		if err := workspace.LaunchClaude(workspacePath, workspaceName); err != nil {
			// Don't fail if Claude is not installed, just print the path
			fmt.Printf("Workspace ready at: %s\n", workspacePath)
			if isNew {
				fmt.Printf("Note: Claude Code not found. You can manually open the workspace at: %s\n", workspacePath)
			}
		}

		return nil
	},
}

func init() {
	launchCmd.Flags().StringSliceVarP(&launchStacks, "stack", "s", []string{}, "Load repos from a named stack in config (repeatable)")
	launchCmd.Flags().StringSliceVarP(&launchRepos, "repo", "r", []string{}, "Add individual repos by git URL (repeatable)")
	launchCmd.Flags().StringVar(&launchBranch, "branch", "", "Default branch to checkout for all repos (supports {workspace} template)")
	launchCmd.Flags().BoolVar(&launchNoBranch, "no-branch", false, "Skip branch checkout; use default branches")

	// Validate mutually exclusive flags in PreRunE
	launchCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if launchBranch != "" && launchNoBranch {
			return fmt.Errorf("--branch and --no-branch are mutually exclusive")
		}
		return nil
	}
}
