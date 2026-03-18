// Package persistence provides data persistence for Savanhi Shell.
// This file implements backup operations.
package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/savanhi/shell/internal/detector"
)

// HasOriginalBackup checks if an original backup exists.
func (p *FilePersister) HasOriginalBackup() (bool, error) {
	if err := p.ensureConfigDir(); err != nil {
		return false, err
	}

	backupPath := filepath.Join(p.configDir, OriginalBackupFile)
	_, err := os.Stat(backupPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, nil
}

// SaveOriginalBackup creates the original backup with the initial state.
// Returns ErrOriginalBackupExists if backup already exists.
func (p *FilePersister) SaveOriginalBackup(snapshot *detector.DetectorResult, rcContents map[string]string) error {
	// Check if original backup already exists
	hasBackup, err := p.HasOriginalBackup()
	if err != nil {
		return err
	}
	if hasBackup {
		return ErrOriginalBackupExists
	}

	if err := p.ensureConfigDir(); err != nil {
		return err
	}

	// Create backup structure
	backup := &OriginalBackup{
		CreatedAt:        time.Now(),
		Version:          "1.0.0", // TODO: Get from build info
		DetectorSnapshot: snapshot,
		RCFiles:          rcContents,
	}

	// Populate shell backup
	if snapshot.Shell != nil {
		backup.Shell = ShellBackup{
			Name: string(snapshot.Shell.Name),
		}
		if rcContent, ok := rcContents[snapshot.Shell.RCFile]; ok {
			backup.Shell.RCContent = rcContent
		}
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}

	// Write atomically
	backupPath := filepath.Join(p.configDir, OriginalBackupFile)
	tempPath := backupPath + ".tmp"

	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	if err := os.Rename(tempPath, backupPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename backup: %w", err)
	}

	return nil
}

// LoadOriginalBackup retrieves the original backup.
// Returns ErrNoOriginalBackup if no backup exists.
func (p *FilePersister) LoadOriginalBackup() (*OriginalBackup, error) {
	backupPath := filepath.Join(p.configDir, OriginalBackupFile)

	data, err := os.ReadFile(backupPath)
	if os.IsNotExist(err) {
		return nil, ErrNoOriginalBackup
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read backup: %w", err)
	}

	var backup OriginalBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidBackup, err)
	}

	return &backup, nil
}

// CreateBackup creates a timestamped backup of current state.
func (p *FilePersister) CreateBackup(description string, files []string) (*Backup, error) {
	if err := p.ensureConfigDir(); err != nil {
		return nil, err
	}

	backupDir, err := p.GetBackupDir()
	if err != nil {
		return nil, err
	}

	// Generate unique ID
	id := generateBackupID()
	timestamp := time.Now()

	// Create backup directory for files
	backupFilesDir := filepath.Join(backupDir, id)
	if err := os.MkdirAll(backupFilesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup files
	backupFiles := make([]BackupFile, 0, len(files))
	var totalSize int64

	for _, originalPath := range files {
		// Read original file
		content, err := os.ReadFile(originalPath)
		if err != nil {
			// Skip files that don't exist
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to read %s: %w", originalPath, err)
		}

		// Create backup file with same relative name
		relPath := filepath.Base(originalPath)
		backupPath := filepath.Join(backupFilesDir, relPath)

		if err := os.WriteFile(backupPath, content, 0600); err != nil {
			return nil, fmt.Errorf("failed to write backup %s: %w", relPath, err)
		}

		// Calculate hash
		hash := fmt.Sprintf("%x", simpleHash(content))

		backupFiles = append(backupFiles, BackupFile{
			OriginalPath: originalPath,
			BackupPath:   backupPath,
			Hash:         hash,
			Size:         int64(len(content)),
		})

		totalSize += int64(len(content))
	}

	// Create backup metadata
	backup := &Backup{
		ID:          id,
		CreatedAt:   timestamp,
		Type:        BackupTypeAuto,
		Description: description,
		Size:        totalSize,
		Files:       backupFiles,
		Metadata:    make(map[string]string),
	}

	// If description contains "pre-update", mark as pre-update backup
	if len(description) > 0 && (description == "pre-update" || len(description) >= 10 && description[:10] == "pre-update") {
		backup.Type = BackupTypePreUpdate
	}

	// Save metadata
	if err := p.saveBackupMetadata(backup); err != nil {
		os.RemoveAll(backupFilesDir)
		return nil, err
	}

	return backup, nil
}

// ListBackups lists all available backups.
func (p *FilePersister) ListBackups() ([]*Backup, error) {
	backupDir, err := p.GetBackupDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(backupDir)
	if os.IsNotExist(err) {
		return []*Backup{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	backups := make([]*Backup, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		backup, err := p.loadBackupMetadata(entry.Name())
		if err != nil {
			// Skip corrupted backups
			continue
		}

		backups = append(backups, backup)
	}

	// Sort by creation time (newest first)
	sortBackups(backups)

	return backups, nil
}

// LoadBackup loads a backup by ID.
func (p *FilePersister) LoadBackup(id string) (*Backup, error) {
	return p.loadBackupMetadata(id)
}

// RestoreBackup restores from a backup by ID.
func (p *FilePersister) RestoreBackup(id string) error {
	backup, err := p.LoadBackup(id)
	if err != nil {
		return err
	}

	// Restore each file
	for _, bf := range backup.Files {
		content, err := os.ReadFile(bf.BackupPath)
		if err != nil {
			return fmt.Errorf("failed to read backup %s: %w", bf.BackupPath, err)
		}

		if err := os.WriteFile(bf.OriginalPath, content, 0644); err != nil {
			return fmt.Errorf("failed to restore %s: %w", bf.OriginalPath, err)
		}
	}

	return nil
}

// DeleteBackup removes a backup by ID.
func (p *FilePersister) DeleteBackup(id string) error {
	// Cannot delete original backup
	if id == "original" || id == OriginalBackupFile {
		return ErrBackupNotFound
	}

	backupDir := filepath.Join(p.configDir, BackupsDir)
	backupPath := filepath.Join(backupDir, id)

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return ErrBackupNotFound
	}

	// Remove backup directory
	return os.RemoveAll(backupPath)
}

// CleanupOldBackups removes backups older than retention days.
func (p *FilePersister) CleanupOldBackups(retentionDays int) (int, error) {
	backups, err := p.ListBackups()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	deleted := 0

	for _, backup := range backups {
		// Skip original backup
		if backup.Type == BackupTypeOriginal {
			continue
		}

		if backup.CreatedAt.Before(cutoff) {
			if err := p.DeleteBackup(backup.ID); err != nil {
				continue // Skip on error
			}
			deleted++
		}
	}

	return deleted, nil
}

// saveBackupMetadata saves backup metadata to disk.
func (p *FilePersister) saveBackupMetadata(backup *Backup) error {
	backupDir := filepath.Join(p.configDir, BackupsDir)
	metadataPath := filepath.Join(backupDir, backup.ID, "metadata.json")

	data, err := json.MarshalIndent(backup, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup metadata: %w", err)
	}

	// Write atomically
	tempPath := metadataPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return os.Rename(tempPath, metadataPath)
}

// loadBackupMetadata loads backup metadata from disk.
func (p *FilePersister) loadBackupMetadata(id string) (*Backup, error) {
	backupDir := filepath.Join(p.configDir, BackupsDir)
	metadataPath := filepath.Join(backupDir, id, "metadata.json")

	data, err := os.ReadFile(metadataPath)
	if os.IsNotExist(err) {
		return nil, ErrBackupNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read backup metadata: %w", err)
	}

	var backup Backup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidBackup, err)
	}

	return &backup, nil
}

// generateBackupID generates a unique backup ID.
func generateBackupID() string {
	return fmt.Sprintf("%d-%s", time.Now().Unix(), uuid.New().String()[:8])
}

// simpleHash creates a simple hash of the content.
func simpleHash(data []byte) uint32 {
	var h uint32
	for _, b := range data {
		h = h*31 + uint32(b)
	}
	return h
}

// sortBackups sorts backups by creation time (newest first).
func sortBackups(backups []*Backup) {
	// Simple insertion sort for small lists
	for i := 1; i < len(backups); i++ {
		for j := i; j > 0 && backups[j].CreatedAt.After(backups[j-1].CreatedAt); j-- {
			backups[j], backups[j-1] = backups[j-1], backups[j]
		}
	}
}
