// Package persistence provides data persistence for Savanhi Shell.
// This file contains tests for migration operations.
package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewMigrator(t *testing.T) {
	m := NewMigrator()
	if m == nil {
		t.Fatal("NewMigrator() returned nil")
	}

	if len(m.migrations) != 0 {
		t.Logf("Migrator has %d migrations registered", len(m.migrations))
	}
}

func TestMigrator_NeedsMigration(t *testing.T) {
	m := NewMigrator()

	tests := []struct {
		name     string
		version  int
		expected bool
	}{
		{"version 0 needs migration", 0, true},
		{"version 1 is current", MigrationVersion, false},
		{"future version doesn't need migration", MigrationVersion + 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.NeedsMigration(tt.version)
			if result != tt.expected {
				t.Errorf("NeedsMigration(%d) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestMigrator_Migrate(t *testing.T) {
	m := NewMigrator()

	tests := []struct {
		name           string
		currentVersion int
		expectError    bool
	}{
		{"migrate from version 0", 0, false},
		{"migrate from version 1", 1, false},
		{"migrate from future version", MigrationVersion + 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte(`{"test": "data"}`)
			result, newVersion, err := m.Migrate(data, tt.currentVersion)

			if tt.expectError {
				if err == nil {
					t.Error("Migrate() should return error, but didn't")
				}
				return
			}

			if err != nil {
				t.Errorf("Migrate() returned unexpected error: %v", err)
				return
			}

			if string(result) != string(data) {
				t.Errorf("Migrate() modified data unexpectedly")
			}

			// For v0 or current version, should be at latest
			t.Logf("Migrated from version %d to %d", tt.currentVersion, newVersion)
		})
	}
}

func TestMigratePreferences_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Should not error when no preferences file exists
	err := p.MigratePreferences()
	if err != nil {
		t.Errorf("MigratePreferences() returned error for non-existent file: %v", err)
	}
}

func TestMigratePreferences_CurrentVersion(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Create preferences with current version
	prefs := &Preferences{
		Version:     MigrationVersion,
		LastUpdated: parseTestTime("2024-01-01T00:00:00Z"),
	}
	data, _ := json.MarshalIndent(prefs, "", " ")

	prefsPath := filepath.Join(tmpDir, PreferencesFile)
	if err := os.WriteFile(prefsPath, data, 0644); err != nil {
		t.Fatalf("Failed to write preferences: %v", err)
	}

	// Should not need migration
	err := p.MigratePreferences()
	if err != nil {
		t.Errorf("MigratePreferences() returned error: %v", err)
	}
}

func TestMigrateOriginalBackup_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Should not error when no backup file exists
	err := p.MigrateOriginalBackup()
	if err != nil {
		t.Errorf("MigrateOriginalBackup() returned error for non-existent file: %v", err)
	}
}

func TestMigrateOriginalBackup_ValidBackup(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Create a valid backup
	backup := &OriginalBackup{
		Version: "1.0.0",
	}
	data, _ := json.MarshalIndent(backup, "", " ")

	backupPath := filepath.Join(tmpDir, OriginalBackupFile)
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		t.Fatalf("Failed to write backup: %v", err)
	}

	// Should validate successfully
	err := p.MigrateOriginalBackup()
	if err != nil {
		t.Errorf("MigrateOriginalBackup() returned error for valid backup: %v", err)
	}
}

func TestMigrateOriginalBackup_InvalidBackup(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Create an invalid backup (corrupted JSON)
	backupPath := filepath.Join(tmpDir, OriginalBackupFile)
	if err := os.WriteFile(backupPath, []byte(`{"invalid json`), 0644); err != nil {
		t.Fatalf("Failed to write backup: %v", err)
	}

	// Should return error for invalid backup
	err := p.MigrateOriginalBackup()
	if err == nil {
		t.Error("MigrateOriginalBackup() should return error for invalid backup")
	}
}

func TestGetSchemaVersion(t *testing.T) {
	version := GetSchemaVersion()
	if version < 1 {
		t.Errorf("GetSchemaVersion() = %d, want >= 1", version)
	}
}

func TestValidateSchemaVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     int
		expectError bool
	}{
		{"version 0 is valid", 0, false},
		{"version 1 is valid", 1, false},
		{"negative version is invalid", -1, true},
		{"future version is invalid", MigrationVersion + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchemaVersion(tt.version)
			if tt.expectError && err == nil {
				t.Error("ValidateSchemaVersion() should return error, but didn't")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateSchemaVersion() returned unexpected error: %v", err)
			}
		})
	}
}

func TestMigrationRegistration(t *testing.T) {
	m := NewMigrator()

	// Register a test migration
	testMigration := Migration{
		FromVersion: 0,
		ToVersion:   1,
		Migrate: func(data []byte) ([]byte, error) {
			// Simple pass-through for testing
			return data, nil
		},
	}

	m.RegisterMigration(testMigration)

	if len(m.migrations) != 1 {
		t.Errorf("Migration not registered, count = %d", len(m.migrations))
	}
}

// parseTestTime parses a time string for testing.
func parseTestTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
