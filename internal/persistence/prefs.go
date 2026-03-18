// Package persistence provides data persistence for Savanhi Shell.
// This file implements preferences operations.
package persistence

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HasPreferences checks if preferences exist.
func (p *FilePersister) HasPreferences() (bool, error) {
	if err := p.ensureConfigDir(); err != nil {
		return false, err
	}

	prefsPath := filepath.Join(p.configDir, PreferencesFile)
	_, err := os.Stat(prefsPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, nil
}

// SavePreferences saves user preferences.
func (p *FilePersister) SavePreferences(prefs *Preferences) error {
	if err := p.ensureConfigDir(); err != nil {
		return err
	}

	// Update timestamp and version
	prefs.LastUpdated = time.Now()
	prefs.Version = PreferencesVersion

	// Validate preferences
	if err := p.validatePreferences(prefs); err != nil {
		return err
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(prefs, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	// Write atomically
	prefsPath := filepath.Join(p.configDir, PreferencesFile)
	tempPath := prefsPath + ".tmp"

	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write preferences: %w", err)
	}

	if err := os.Rename(tempPath, prefsPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename preferences: %w", err)
	}

	return nil
}

// LoadPreferences loads user preferences.
// Returns ErrPreferencesNotFound if no preferences exist.
func (p *FilePersister) LoadPreferences() (*Preferences, error) {
	prefsPath := filepath.Join(p.configDir, PreferencesFile)

	data, err := os.ReadFile(prefsPath)
	if os.IsNotExist(err) {
		return nil, ErrPreferencesNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read preferences: %w", err)
	}

	var prefs Preferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
	}

	return &prefs, nil
}

// ResetPreferences resets preferences to defaults.
func (p *FilePersister) ResetPreferences() (*Preferences, error) {
	defaults := p.getDefaultPreferences()

	if err := p.SavePreferences(defaults); err != nil {
		return nil, err
	}

	return defaults, nil
}

// getDefaultPreferences returns the default preferences.
func (p *FilePersister) getDefaultPreferences() *Preferences {
	return &Preferences{
		Version:     PreferencesVersion,
		LastUpdated: time.Now(),
		Theme: ThemePreferences{
			Name:           "default",
			AutoUpdate:     true,
			Variant:        "dark",
			CustomSettings: make(map[string]interface{}),
		},
		Shell: ShellPreferences{
			PreferredShell:           "zsh",
			EnableSyntaxHighlighting: true,
			EnableAutosuggestions:    true,
			EnableHistorySettings:    true,
			Aliases:                  make(map[string]string),
		},
		Terminal: TerminalPreferences{
			FontFamily:      "MesloLGM Nerd Font",
			FontSize:        12,
			EnableLigatures: true,
			CursorStyle:     "block",
		},
		Fonts: FontPreferences{
			PrimaryNerdFont:     "MesloLGM Nerd Font",
			FallbackFont:        "Monaco",
			EnableNerdFontIcons: true,
		},
		Tools: ToolPreferences{
			EnableZoxide: true,
			EnableFzf:    true,
			EnableBat:    true,
			EnableEza:    true,
			CustomTools:  make(map[string]ToolConfig),
		},
		Advanced: AdvancedPreferences{
			CreateBackup:        true,
			BackupRetentionDays: MaxBackupRetentionDays,
			EnableTelemetry:     false,
			AutoUpdate:          true,
			UpdateChannel:       "stable",
			Verbose:             false,
			Experimental:        make(map[string]bool),
		},
	}
}

// validatePreferences validates user preferences.
func (p *FilePersister) validatePreferences(prefs *Preferences) error {
	// Validate update channel
	validChannels := map[string]bool{"stable": true, "beta": true}
	if prefs.Advanced.UpdateChannel != "" && !validChannels[prefs.Advanced.UpdateChannel] {
		return fmt.Errorf("invalid update channel: %s", prefs.Advanced.UpdateChannel)
	}

	// Validate font size
	if prefs.Terminal.FontSize < 6 || prefs.Terminal.FontSize > 72 {
		return fmt.Errorf("font size must be between 6 and 72, got %d", prefs.Terminal.FontSize)
	}

	// Validate cursor style
	validCursors := map[string]bool{"block": true, "bar": true, "underline": true}
	if prefs.Terminal.CursorStyle != "" && !validCursors[prefs.Terminal.CursorStyle] {
		return fmt.Errorf("invalid cursor style: %s", prefs.Terminal.CursorStyle)
	}

	// Validate retention days
	if prefs.Advanced.BackupRetentionDays < 1 || prefs.Advanced.BackupRetentionDays > 365 {
		return fmt.Errorf("backup retention must be between 1 and 365 days")
	}

	return nil
}

// generateID generates a random ID for history entries.
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
