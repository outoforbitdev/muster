package cmd

import (
	"github.com/spf13/cobra"
)

var (
	launchStacks []string
	launchRepos  []string
	launchBranch string
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
		workspace := args[0]
		// TODO: Implement workspace launch logic
		cmd.Printf("Launching workspace: %s\n", workspace)
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
			return cmd.Usage()
		}
		return nil
	}
}
