package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var removeConfirm bool

var removeCmd = &cobra.Command{
	Use:   "remove <workspace>",
	Short: "Remove a workspace",
	Long: `Remove a workspace and all its contents.

By default, you will be prompted to confirm deletion.
Use --yes to skip the confirmation prompt.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace := args[0]

		if !removeConfirm {
			// TODO: Implement confirmation prompt
			cmd.Printf("Would delete workspace: %s\n", workspace)
			return nil
		}

		// TODO: Implement workspace deletion logic
		fmt.Printf("Deleted workspace: %s\n", workspace)
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&removeConfirm, "yes", "y", false, "Skip confirmation prompt and immediately delete the workspace")
}
