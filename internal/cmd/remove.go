package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		workspaceName := args[0]
		workspacePath := filepath.Join(os.Getenv("HOME"), ".workspaces", workspaceName)

		// Check if workspace exists
		if _, err := os.Stat(workspacePath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("workspace %q not found at %s", workspaceName, workspacePath)
			}
			return fmt.Errorf("failed to check workspace: %w", err)
		}

		// Ask for confirmation if not --yes
		if !removeConfirm {
			fmt.Printf("Are you sure you want to delete workspace %q? This cannot be undone.\n", workspaceName)
			fmt.Print("Type 'yes' to confirm: ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			if strings.TrimSpace(response) != "yes" {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}

		// Delete the workspace
		if err := os.RemoveAll(workspacePath); err != nil {
			return fmt.Errorf("failed to delete workspace: %w", err)
		}

		fmt.Printf("Deleted workspace %q\n", workspaceName)
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&removeConfirm, "yes", "y", false, "Skip confirmation prompt and immediately delete the workspace")
}
