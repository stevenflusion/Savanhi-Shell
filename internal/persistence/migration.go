// Package persistence provides data persistence for Savanhi Shell.
// This file implements migration support for future version compatibility.
package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MigrationVersion represents the current schema version.
const MigrationVersion = 1

// Migration errors.
var (
	// ErrMigrationFailed indicates a migration failed.
	ErrMigrationFailed = fmt.Errorf("migration failed")
	// ErrUnsupportedVersion indicates an unsupported schema version.
	ErrUnsupportedVersion = fmt.Errorf("unsupported schema version")
)

// MigrationFunc is a function that migrates data from one version to the next.
type MigrationFunc func([]byte) ([]byte, error)

// Migration represents a single migration step.
type Migration struct {
	// FromVersion is the version to migrate from.
	FromVersion int
	// ToVersion is the version to migrate to.
	ToVersion int
	// Migrate performs the migration.
	Migrate MigrationFunc
}

// Migrator handles schema migrations for persistence files.
type Migrator struct {
	// migrations is a list of available migrations.
	migrations []Migration
}

// NewMigrator creates a new Migrator with all registered migrations.
func NewMigrator() *Migrator {
	m := &Migrator{
		migrations: make([]Migration, 0),
	}

	// Register migrations as needed
	// Example: m.RegisterMigration(migrateV0ToV1)
	// Future versions would add more migrations here

	return m
}

// RegisterMigration registers a migration.
func (m *Migrator) RegisterMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// NeedsMigration checks if a file needs migration.
func (m *Migrator) NeedsMigration(version int) bool {
	return version < MigrationVersion
}

// Migrate migrates data from the current version to the latest version.
func (m *Migrator) Migrate(data []byte, currentVersion int) ([]byte, int, error) {
	// Already at latest version
	if currentVersion >= MigrationVersion {
		return data, currentVersion, nil
	}

	// No migrations registered yet
	if len(m.migrations) == 0 && currentVersion < MigrationVersion {
		// For v1, we don't need migrations yet - just update version
		if currentVersion == 0 {
			currentVersion = MigrationVersion
		}
	}

	// Apply migrations in order
	result := data
	version := currentVersion

	for _, migration := range m.migrations {
		if version == migration.FromVersion {
			var err error
			result, err = migration.Migrate(result)
			if err != nil {
				return nil, version, fmt.Errorf("%w: %v", ErrMigrationFailed, err)
			}
			version = migration.ToVersion
		}
	}

	return result, version, nil
}

// MigratePreferences migrates preferences from an older version.
func (p *FilePersister) MigratePreferences() error {
	prefsPath := filepath.Join(p.configDir, PreferencesFile)

	// Check if preferences file exists
	data, err := os.ReadFile(prefsPath)
	if os.IsNotExist(err) {
		return nil // No preferences file, nothing to migrate
	}
	if err != nil {
		return fmt.Errorf("failed to read preferences: %w", err)
	}

	// Parse version from preferences
	var rawPrefs map[string]interface{}
	if err := json.Unmarshal(data, &rawPrefs); err != nil {
		return fmt.Errorf("failed to parse preferences: %w", err)
	}

	// Get current version
	currentVersion := 0
	if v, ok := rawPrefs["version"].(float64); ok {
		currentVersion = int(v)
	}

	// Check if migration needed
	migrator := NewMigrator()
	if !migrator.NeedsMigration(currentVersion) {
		return nil // Already at latest version
	}

	// Migrate
	newData, newVersion, err := migrator.Migrate(data, currentVersion)
	if err != nil {
		return fmt.Errorf("failed to migrate preferences: %w", err)
	}

	// Write migrated preferences atomically
	tempPath := prefsPath + ".tmp"
	if err := os.WriteFile(tempPath, newData, 0600); err != nil {
		return fmt.Errorf("failed to write migrated preferences: %w", err)
	}

	if err := os.Rename(tempPath, prefsPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename migrated preferences: %w", err)
	}

	// Update version in the new data
	var newPrefs map[string]interface{}
	if err := json.Unmarshal(newData, &newPrefs); err == nil {
		newPrefs["version"] = newVersion
		if newData, err = json.MarshalIndent(newPrefs, "", " "); err == nil {
			tempPath := prefsPath + ".tmp"
			if err := os.WriteFile(tempPath, newData, 0600); err == nil {
				os.Rename(tempPath, prefsPath)
			}
		}
	}

	return nil
}

// MigrateOriginalBackup migrates original backup from an older version.
func (p *FilePersister) MigrateOriginalBackup() error {
	backupPath := filepath.Join(p.configDir, OriginalBackupFile)

	// Check if backup file exists
	data, err := os.ReadFile(backupPath)
	if os.IsNotExist(err) {
		return nil // No backup file, nothing to migrate
	}
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Original backup doesn't have a version field yet
	// For v1, we just ensure it parses correctly
	var backup OriginalBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidBackup, err)
	}

	// For future versions, add migration logic here
	return nil
}

// GetSchemaVersion returns the current schema version for a file type.
func GetSchemaVersion() int {
	return MigrationVersion
}

// BackupMetadata contains metadata about a backup for migration purposes.
type BackupMetadata struct {
	SchemaVersion int    `json:"schema_version"`
	CreatedBy     string `json:"created_by"`
	CreatedAt     string `json:"created_at"`
}

// ValidateSchemaVersion validates that the schema version is supported.
func ValidateSchemaVersion(version int) error {
	if version < 0 {
		return fmt.Errorf("%w: version %d is invalid", ErrUnsupportedVersion, version)
	}
	if version > MigrationVersion {
		return fmt.Errorf("%w: version %d is newer than supported %d", ErrUnsupportedVersion, version, MigrationVersion)
	}
	return nil
}
