package cmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/outoforbitdev/muster/internal/config"
)

//go:embed config.example.json
var configExampleFS embed.FS

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize muster configuration",
	Long: `Initialize the muster configuration by creating ~/.config/muster/config.json.

This creates a new config file with an example configuration that you can customize.
If a config file already exists, this command will fail to prevent accidental overwrites.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine config path
		homeDir := os.Getenv("HOME")
		configDir := filepath.Join(homeDir, ".config", "muster")
		configPath := filepath.Join(configDir, "config.json")

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists at %s\n\nTo reconfigure, edit the file directly or delete it and run 'muster init' again", configPath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check config file: %w", err)
		}

		// Create config directory
		fmt.Fprintf(os.Stderr, "Creating config directory at %s...\n", configDir)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Read embedded example config
		exampleData, err := configExampleFS.ReadFile("config.example.json")
		if err != nil {
			return fmt.Errorf("failed to read example config: %w", err)
		}

		// Validate the example config by parsing it
		var cfg config.Config
		if err := json.Unmarshal(exampleData, &cfg); err != nil {
			return fmt.Errorf("failed to parse example config: %w", err)
		}

		// Validate the config structure
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("example config validation failed: %w", err)
		}

		// Write config file
		fmt.Fprintf(os.Stderr, "Writing config file...\n")
		if err := os.WriteFile(configPath, exampleData, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Fprintf(os.Stderr, "\n✓ Config initialized successfully!\n")
		fmt.Printf("Config file created at: %s\n\n", configPath)
		fmt.Printf("Next steps:\n")
		fmt.Printf("1. Edit the config file to customize your stacks and repositories\n")
		fmt.Printf("2. Run 'muster launch <workspace> --stack <stack-name>' to create a workspace\n")

		return nil
	},
}
