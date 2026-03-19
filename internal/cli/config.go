// Package cli provides command-line interface functionality for Savanhi Shell.
// This file implements configuration management.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigManager manages configuration files.
type ConfigManager struct {
	// configPath is the path to the configuration file.
	configPath string

	// config is the current configuration.
	config *Config
}

// NewConfigManager creates a new configuration manager.
func NewConfigManager(configPath string) (*ConfigManager, error) {
	if configPath == "" {
		// Default config path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configPath = filepath.Join(homeDir, ".config", "savanhi", "config.json")
	}

	cm := &ConfigManager{
		configPath: configPath,
		config:     NewConfig(),
	}

	// Try to load existing config
	if _, err := os.Stat(configPath); err == nil {
		if err := cm.Load(); err != nil {
			// Non-fatal, use defaults
			fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		}
	}

	return cm, nil
}

// Load loads the configuration from file.
func (cm *ConfigManager) Load() error {
	config, err := LoadConfig(cm.configPath)
	if err != nil {
		return err
	}
	cm.config = config
	return nil
}

// Save saves the configuration to file.
func (cm *ConfigManager) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return SaveConfig(cm.configPath, cm.config)
}

// Get returns the current configuration.
func (cm *ConfigManager) Get() *Config {
	return cm.config
}

// Set updates the configuration.
func (cm *ConfigManager) Set(config *Config) {
	cm.config = config
}

// SetTheme sets the theme.
func (cm *ConfigManager) SetTheme(theme string) {
	cm.config.Theme = theme
}

// SetFont sets the font.
func (cm *ConfigManager) SetFont(font string) {
	cm.config.Font = font
}

// SetTools sets the tools.
func (cm *ConfigManager) SetTools(tools []string) {
	cm.config.Tools = tools
}

// InitConfig creates a default configuration file.
func InitConfig(path string) error {
	config := NewConfig()

	// Set some example values for generating a template
	config.Theme = "powerlevel10k"
	config.Font = "MesloLGS NF"
	config.Tools = []string{"zoxide", "fzf", "bat", "eza"}

	return SaveConfig(path, config)
}

// ShowConfig displays the current configuration.
func ShowConfig(config *Config) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding config: %v\n", err)
		return
	}
	fmt.Printf("%s\n", data)
}

// ValidateConfig validates the configuration.
func ValidateConfig(config *Config) error {
	// Validate theme (can be empty for default)
	if config.Theme != "" && len(config.Theme) > 100 {
		return fmt.Errorf("theme name too long (max 100 characters)")
	}

	// Validate font
	if config.Font != "" && len(config.Font) > 100 {
		return fmt.Errorf("font name too long (max 100 characters)")
	}

	// Validate tools
	validTools := map[string]bool{
		"zoxide": true,
		"fzf":    true,
		"bat":    true,
		"eza":    true,
	}
	for _, tool := range config.Tools {
		if !validTools[tool] {
			return fmt.Errorf("unknown tool: %s (valid: zoxide, fzf, bat, eza)", tool)
		}
	}

	// Validate timeout
	if config.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	return nil
}

// ValidatePlugins validates plugin names for installation.
func ValidatePlugins(plugins []string) error {
	validPlugins := map[string]bool{
		"zsh-autosuggestions":     true,
		"zsh-syntax-highlighting": true,
	}

	for _, plugin := range plugins {
		if !validPlugins[plugin] {
			return fmt.Errorf("unknown plugin: %s (valid: zsh-autosuggestions, zsh-syntax-highlighting)", plugin)
		}
	}

	return nil
}
