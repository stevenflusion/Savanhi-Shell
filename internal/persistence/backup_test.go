// Package persistence provides data persistence for Savanhi Shell.
// This file contains tests for backup operations.
package persistence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/savanhi/shell/internal/detector"
)

func TestHasOriginalBackup_False(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	hasBackup, err := p.HasOriginalBackup()
	if err != nil {
		t.Fatalf("HasOriginalBackup() returned error: %v", err)
	}
	if hasBackup {
		t.Error("HasOriginalBackup() should return false for new installation")
	}
}

func TestSaveOriginalBackup(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	snapshot := &detector.DetectorResult{
		OS: &detector.OSInfo{
			Type:    detector.OSTypeLinux,
			Distro:  "ubuntu",
			Version: "22.04",
			Arch:    "amd64",
		},
		Shell: &detector.ShellInfo{
			Name:    detector.ShellTypeBash,
			Version: "5.1.16",
			RCFile:  "/home/test/.bashrc",
		},
	}

	rcContents := map[string]string{
		"/home/test/.bashrc": "# Original bashrc content\nexport PATH=$PATH:/usr/local/bin",
	}

	err := p.SaveOriginalBackup(snapshot, rcContents)
	if err != nil {
		t.Fatalf("SaveOriginalBackup() returned error: %v", err)
	}

	// Verify backup exists
	hasBackup, _ := p.HasOriginalBackup()
	if !hasBackup {
		t.Error("HasOriginalBackup() should return true after save")
	}

	// Verify backup cannot be overwritten
	err = p.SaveOriginalBackup(snapshot, rcContents)
	if err != ErrOriginalBackupExists {
		t.Errorf("Second SaveOriginalBackup() should return ErrOriginalBackupExists, got: %v", err)
	}
}

func TestLoadOriginalBackup(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Test loading non-existent backup
	_, err := p.LoadOriginalBackup()
	if err != ErrNoOriginalBackup {
		t.Errorf("LoadOriginalBackup() should return ErrNoOriginalBackup, got: %v", err)
	}

	// Create a backup
	snapshot := &detector.DetectorResult{
		OS: &detector.OSInfo{Type: detector.OSTypeMacOS},
	}
	rcContents := map[string]string{"/home/test/.zshrc": "original content"}

	if err := p.SaveOriginalBackup(snapshot, rcContents); err != nil {
		t.Fatalf("SaveOriginalBackup() returned error: %v", err)
	}

	// Load the backup
	backup, err := p.LoadOriginalBackup()
	if err != nil {
		t.Fatalf("LoadOriginalBackup() returned error: %v", err)
	}

	if backup == nil {
		t.Fatal("LoadOriginalBackup() returned nil backup")
	}

	if backup.DetectorSnapshot == nil {
		t.Error("DetectorSnapshot is nil")
	}

	if len(backup.RCFiles) != 1 {
		t.Errorf("RCFiles length = %d, want 1", len(backup.RCFiles))
	}
}

func TestCreateBackup(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Create test files to backup
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	backup, err := p.CreateBackup("test backup", []string{testFile})
	if err != nil {
		t.Fatalf("CreateBackup() returned error: %v", err)
	}

	if backup == nil {
		t.Fatal("CreateBackup() returned nil backup")
	}

	if backup.ID == "" {
		t.Error("Backup ID is empty")
	}

	if backup.Description != "test backup" {
		t.Errorf("Backup description = %s, want 'test backup'", backup.Description)
	}

	if len(backup.Files) != 1 {
		t.Errorf("Backup files length = %d, want 1", len(backup.Files))
	}
}

func TestListBackups(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// List empty backups
	backups, err := p.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() returned error: %v", err)
	}

	if len(backups) != 0 {
		t.Errorf("ListBackups() returned %d backups, want 0", len(backups))
	}

	// Create a backup
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	p.CreateBackup("first backup", []string{testFile})

	backups, err = p.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() returned error: %v", err)
	}

	if len(backups) != 1 {
		t.Errorf("ListBackups() returned %d backups, want 1", len(backups))
	}
}

func TestDeleteBackup(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Create a backup
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	backup, _ := p.CreateBackup("test", []string{testFile})

	// Delete the backup
	err := p.DeleteBackup(backup.ID)
	if err != nil {
		t.Fatalf("DeleteBackup() returned error: %v", err)
	}

	// Verify backup is gone
	_, err = p.LoadBackup(backup.ID)
	if err != ErrBackupNotFound {
		t.Errorf("LoadBackup() should return ErrBackupNotFound, got: %v", err)
	}
}

func TestDeleteOriginalBackup(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Should not be able to delete original backup
	err := p.DeleteBackup(OriginalBackupFile)
	if err != ErrBackupNotFound {
		t.Errorf("DeleteBackup(original) should return ErrBackupNotFound, got: %v", err)
	}
}
