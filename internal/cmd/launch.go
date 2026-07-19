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
	launchAgent    bool
	launchNoAgent  bool
	launchEditor   bool
	launchNoEditor bool
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
		workspacePath := filepath.Join(os.Getenv("HOME"), ".muster", workspaceName)

		// Check if workspace already exists
		_, err := os.Stat(workspacePath)
		isNew := os.IsNotExist(err)

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Launch or create workspace
		workspacePath, err = workspace.LaunchWorkspace(
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

		if isNew {
			fmt.Fprintf(os.Stderr, "\n✓ Workspace created successfully!\n")
		} else {
			fmt.Fprintf(os.Stderr, "\n✓ Opening existing workspace...\n")
		}

		fmt.Printf("Workspace ready at: %s\n", workspacePath)

		if workspace.ShouldLaunchAgent(cfg, launchAgent, launchNoAgent) {
			if err := workspace.LaunchAgent(cfg, workspacePath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to launch agent command: %v\n", err)
			}
		}

		if workspace.ShouldLaunchEditor(cfg, launchEditor, launchNoEditor) {
			if err := workspace.LaunchEditor(cfg, workspacePath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to launch editor command: %v\n", err)
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
	launchCmd.Flags().BoolVar(&launchAgent, "agent", false, "Launch the agent command after creating/opening the workspace")
	launchCmd.Flags().BoolVar(&launchNoAgent, "no-agent", false, "Do not launch the agent command")
	launchCmd.Flags().BoolVar(&launchEditor, "editor", false, "Launch the editor command after creating/opening the workspace")
	launchCmd.Flags().BoolVar(&launchNoEditor, "no-editor", false, "Do not launch the editor command")

	// Validate mutually exclusive flags in PreRunE
	launchCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if launchBranch != "" && launchNoBranch {
			return fmt.Errorf("--branch and --no-branch are mutually exclusive")
		}
		if launchAgent && launchNoAgent {
			return fmt.Errorf("--agent and --no-agent are mutually exclusive")
		}
		if launchEditor && launchNoEditor {
			return fmt.Errorf("--editor and --no-editor are mutually exclusive")
		}
		return nil
	}
}
