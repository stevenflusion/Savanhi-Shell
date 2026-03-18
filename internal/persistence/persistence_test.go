// Package persistence provides data persistence for Savanhi Shell.
// This file contains tests for persistence operations.
package persistence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFilePersister(t *testing.T) {
	p, err := NewFilePersister()
	if err != nil {
		t.Fatalf("NewFilePersister() returned error: %v", err)
	}
	if p == nil {
		t.Fatal("NewFilePersister() returned nil")
	}
	if p.configDir == "" {
		t.Error("configDir is empty")
	}
}

func TestNewFilePersisterWithPath(t *testing.T) {
	tmpDir := t.TempDir()

	p, err := NewFilePersisterWithPath(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePersisterWithPath() returned error: %v", err)
	}
	if p == nil {
		t.Fatal("NewFilePersisterWithPath() returned nil")
	}
	if p.configDir != tmpDir {
		t.Errorf("configDir = %s, want %s", p.configDir, tmpDir)
	}
}

func TestNewFilePersisterWithEmptyPath(t *testing.T) {
	_, err := NewFilePersisterWithPath("")
	if err == nil {
		t.Error("NewFilePersisterWithPath('') should return error")
	}
}

func TestGetConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	dir, err := p.GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() returned error: %v", err)
	}
	if dir != tmpDir {
		t.Errorf("GetConfigDir() = %s, want %s", dir, tmpDir)
	}

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}
}

func TestGetBackupDir(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	dir, err := p.GetBackupDir()
	if err != nil {
		t.Fatalf("GetBackupDir() returned error: %v", err)
	}

	expectedDir := filepath.Join(tmpDir, BackupsDir)
	if dir != expectedDir {
		t.Errorf("GetBackupDir() = %s, want %s", dir, expectedDir)
	}

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Backup directory was not created")
	}
}
