// Package persistence provides data persistence for Savanhi Shell.
// This file contains tests for preferences operations.
package persistence

import (
	"testing"
)

func TestHasPreferences(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Test no preferences initially
	hasPrefs, err := p.HasPreferences()
	if err != nil {
		t.Fatalf("HasPreferences() returned error: %v", err)
	}
	if hasPrefs {
		t.Error("HasPreferences() should return false initially")
	}
}

func TestSavePreferences(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	prefs := &Preferences{
		Theme: ThemePreferences{
			Name: "dark",
		},
		Shell: ShellPreferences{
			PreferredShell: "zsh",
		},
		Terminal: TerminalPreferences{
			FontSize: 12,
		},
		Advanced: AdvancedPreferences{
			BackupRetentionDays: 30,
		},
	}

	err := p.SavePreferences(prefs)
	if err != nil {
		t.Fatalf("SavePreferences() returned error: %v", err)
	}

	// Verify preferences saved
	hasPrefs, _ := p.HasPreferences()
	if !hasPrefs {
		t.Error("HasPreferences() should return true after save")
	}
}

func TestLoadPreferences(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Test loading non-existent preferences
	_, err := p.LoadPreferences()
	if err != ErrPreferencesNotFound {
		t.Errorf("LoadPreferences() should return ErrPreferencesNotFound, got: %v", err)
	}

	// Save preferences first with valid values
	prefs := &Preferences{
		Theme: ThemePreferences{
			Name: "test",
		},
		Terminal: TerminalPreferences{
			FontSize: 12,
		},
		Advanced: AdvancedPreferences{
			BackupRetentionDays: 30,
		},
	}
	p.SavePreferences(prefs)

	// Load preferences
	loaded, err := p.LoadPreferences()
	if err != nil {
		t.Fatalf("LoadPreferences() returned error: %v", err)
	}

	if loaded.Theme.Name != "test" {
		t.Errorf("Theme.Name = %s, want 'test'", loaded.Theme.Name)
	}
}

func TestResetPreferences(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	prefs, err := p.ResetPreferences()
	if err != nil {
		t.Fatalf("ResetPreferences() returned error: %v", err)
	}

	// Check default values
	if prefs.Theme.Name != "default" {
		t.Errorf("Default theme name = %s, want 'default'", prefs.Theme.Name)
	}

	if prefs.Shell.PreferredShell != "zsh" {
		t.Errorf("Default preferred shell = %s, want 'zsh'", prefs.Shell.PreferredShell)
	}

	if prefs.Terminal.FontFamily != "MesloLGM Nerd Font" {
		t.Errorf("Default font = %s, want 'MesloLGM Nerd Font'", prefs.Terminal.FontFamily)
	}
}

func TestValidatePreferences(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	tests := []struct {
		name    string
		prefs   *Preferences
		wantErr bool
	}{
		{
			name: "valid preferences",
			prefs: &Preferences{
				Terminal: TerminalPreferences{FontSize: 12},
				Advanced: AdvancedPreferences{UpdateChannel: "stable", BackupRetentionDays: 30},
			},
			wantErr: false,
		},
		{
			name: "invalid update channel",
			prefs: &Preferences{
				Advanced: AdvancedPreferences{UpdateChannel: "invalid", BackupRetentionDays: 30},
			},
			wantErr: true,
		},
		{
			name: "font size too small",
			prefs: &Preferences{
				Terminal: TerminalPreferences{FontSize: 4},
				Advanced: AdvancedPreferences{BackupRetentionDays: 30},
			},
			wantErr: true,
		},
		{
			name: "font size too large",
			prefs: &Preferences{
				Terminal: TerminalPreferences{FontSize: 100},
				Advanced: AdvancedPreferences{BackupRetentionDays: 30},
			},
			wantErr: true,
		},
		{
			name: "invalid cursor style",
			prefs: &Preferences{
				Terminal: TerminalPreferences{CursorStyle: "invalid", FontSize: 12},
				Advanced: AdvancedPreferences{BackupRetentionDays: 30},
			},
			wantErr: true,
		},
		{
			name: "invalid retention days",
			prefs: &Preferences{
				Advanced: AdvancedPreferences{BackupRetentionDays: 0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.validatePreferences(tt.prefs)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePreferences() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
