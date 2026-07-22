package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "muster",
	Short: "Muster: Manage multi-repo workspaces with coordinated branches",
	Long: `Muster is a CLI tool for creating and managing workspaces that contain multiple cloned repositories with coordinated branch checkout and configuration.

Use 'muster init' to initialize your configuration.
Use 'muster launch' to create a new workspace or open an existing one.
Use 'muster remove' to delete a workspace.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(launchCmd)
	rootCmd.AddCommand(removeCmd)
}
