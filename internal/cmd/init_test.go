package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/outoforbitdev/muster/internal/config"
)

func TestInitCommand(t *testing.T) {
	origHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", origHome); err != nil {
			t.Fatalf("failed to restore HOME: %v", err)
		}
	}()

	t.Run("successful init creates config", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.Setenv("HOME", tempDir); err != nil {
			t.Fatalf("failed to set HOME: %v", err)
		}

		cmd := initCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		configPath := filepath.Join(tempDir, ".config", "muster", "config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("config file was not created at %s", configPath)
		}

		// Verify config can be parsed
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		var cfg config.Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Errorf("config is not valid JSON: %v", err)
		}

		// Verify config passes validation
		if err := cfg.Validate(); err != nil {
			t.Errorf("config failed validation: %v", err)
		}
	})

	t.Run("error when config already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.Setenv("HOME", tempDir); err != nil {
			t.Fatalf("failed to set HOME: %v", err)
		}

		// First init should succeed
		cmd := initCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("first init failed: %v", err)
		}

		// Second init should fail
		err = cmd.RunE(cmd, []string{})
		if err == nil {
			t.Error("expected error on second init, got none")
		}
	})

	t.Run("creates directory if it doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.Setenv("HOME", tempDir); err != nil {
			t.Fatalf("failed to set HOME: %v", err)
		}

		cmd := initCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		configDir := filepath.Join(tempDir, ".config", "muster")
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			t.Errorf("config directory was not created at %s", configDir)
		}
	})

	t.Run("embedded config is valid", func(t *testing.T) {
		data, err := configExampleFS.ReadFile("config.example.json")
		if err != nil {
			t.Fatalf("failed to read embedded config: %v", err)
		}

		var cfg config.Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Errorf("embedded config is not valid JSON: %v", err)
		}

		if err := cfg.Validate(); err != nil {
			t.Errorf("embedded config failed validation: %v", err)
		}
	})
}
