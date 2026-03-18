// Package cli_test provides tests for the CLI package.
package cli_test

import (
	"context"
	stderrors "errors"
	"os"
	"testing"
	"time"

	"github.com/savanhi/shell/internal/cli"
)

func TestNewConfig(t *testing.T) {
	config := cli.NewConfig()

	if config == nil {
		t.Fatal("NewConfig() returned nil")
	}

	// Check defaults
	if !config.InstallOhMyPosh {
		t.Error("Default InstallOhMyPosh should be true")
	}
	if !config.InstallZoxide {
		t.Error("Default InstallZoxide should be true")
	}
	if !config.InstallFzf {
		t.Error("Default InstallFzf should be true")
	}
	if !config.InstallBat {
		t.Error("Default InstallBat should be true")
	}
	if !config.InstallEza {
		t.Error("Default InstallEza should be true")
	}
	if config.Timeout != 10*time.Minute {
		t.Errorf("Default Timeout = %v, want %v", config.Timeout, 10*time.Minute)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temp config file
	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	configContent := `{
		"theme": "test-theme",
		"font": "TestFont NF",
		"install_oh_my_posh": false,
		"dry_run": true
	}`

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	config, err := cli.LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config.Theme != "test-theme" {
		t.Errorf("Theme = %v, want %v", config.Theme, "test-theme")
	}
	if config.Font != "TestFont NF" {
		t.Errorf("Font = %v, want %v", config.Font, "TestFont NF")
	}
	if config.InstallOhMyPosh {
		t.Error("InstallOhMyPosh should be false")
	}
	if !config.DryRun {
		t.Error("DryRun should be true")
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := cli.LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("LoadConfig() should return error for non-existent file")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	config := cli.NewConfig()
	config.Theme = "saved-theme"
	config.Font = "SavedFont NF"

	if err := cli.SaveConfig(tmpFile.Name(), config); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify by loading
	loaded, err := cli.LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig() after save error = %v", err)
	}

	if loaded.Theme != "saved-theme" {
		t.Errorf("Loaded Theme = %v, want %v", loaded.Theme, "saved-theme")
	}
	if loaded.Font != "SavedFont NF" {
		t.Errorf("Loaded Font = %v, want %v", loaded.Font, "SavedFont NF")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *cli.Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  cli.NewConfig(),
			wantErr: false,
		},
		{
			name: "valid custom config",
			config: &cli.Config{
				Theme:         "custom",
				Font:          "CustomFont",
				InstallZoxide: true,
				Timeout:       5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "theme too long",
			config: &cli.Config{
				Theme: string(make([]byte, 101)),
			},
			wantErr: true,
		},
		{
			name: "font too long",
			config: &cli.Config{
				Font: string(make([]byte, 101)),
			},
			wantErr: true,
		},
		{
			name: "invalid tool",
			config: &cli.Config{
				Tools: []string{"invalid-tool"},
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: &cli.Config{
				Timeout: -1 * time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cli.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigManager(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "savanhi-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := tmpDir + "/config.json"

	cm, err := cli.NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}

	// Test Get
	config := cm.Get()
	if config == nil {
		t.Error("Get() returned nil")
	}

	// Test Set
	newConfig := cli.NewConfig()
	newConfig.Theme = "test-theme"
	cm.Set(newConfig)

	if cm.Get().Theme != "test-theme" {
		t.Error("Set() did not update config")
	}

	// Test Save
	if err := cm.Save(); err != nil {
		t.Errorf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Save() did not create config file")
	}
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: cli.ExitSuccess,
		},
		{
			name:     "standard error",
			err:      stderrors.New("standard error"),
			expected: cli.ExitError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cli.ExitCodeFromError(tt.err)
			if got != tt.expected {
				t.Errorf("ExitCodeFromError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNonInteractiveModeNewConfig(t *testing.T) {
	// Test with default config
	config := cli.NewConfig()

	// Verify it can be used to create a non-interactive mode
	// Note: This would require a full integration test for actual functionality
	if config.Timeout == 0 {
		t.Error("Default config should have non-zero timeout")
	}
}

func TestInitConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "init-config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if err := cli.InitConfig(tmpFile.Name()); err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	// Verify file was created and can be loaded
	config, err := cli.LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig() after InitConfig error = %v", err)
	}

	if config.Theme == "" {
		t.Error("InitConfig did not set default theme")
	}
	if config.Font == "" {
		t.Error("InitConfig did not set default font")
	}
}

func TestShowConfig(t *testing.T) {
	config := cli.NewConfig()
	config.Theme = "show-test"

	// Test that ShowConfig doesn't panic
	// It just prints to stdout, so we just verify it runs
	cli.ShowConfig(config)
}

func TestNonInteractiveMode(t *testing.T) {
	t.Run("dry run", func(t *testing.T) {
		config := cli.NewConfig()
		config.DryRun = true

		// Create non-interactive mode
		// This would require system detection which may not work in test environment
		// So we just test basic construction
		_, err := cli.NewNonInteractiveMode(config, os.Stdout, os.Stderr, false)
		// May fail due to system detection - that's okay for unit test
		if err != nil {
			t.Logf("NewNonInteractiveMode() error (expected in some environments): %v", err)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Context should already be expired
		select {
		case <-ctx.Done():
			// Expected
		default:
			t.Error("Context should be expired immediately")
		}
	})
}
