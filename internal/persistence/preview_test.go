// Package persistence provides data persistence for Savanhi Shell.
// This file contains tests for preview session operations.
package persistence

import (
	"testing"
)

func TestCreatePreviewSession(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	session, err := p.CreatePreviewSession("powerlevel10k", "backup content")
	if err != nil {
		t.Fatalf("CreatePreviewSession() returned error: %v", err)
	}

	if session == nil {
		t.Fatal("CreatePreviewSession() returned nil session")
	}

	if session.ID == "" {
		t.Error("Session ID is empty")
	}

	if session.Theme != "powerlevel10k" {
		t.Errorf("Session theme = %s, want 'powerlevel10k'", session.Theme)
	}

	if !session.Active {
		t.Error("Session should be active")
	}
}

func TestGetPreviewSession(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Test no session
	_, err := p.GetPreviewSession()
	if err != ErrPreviewSessionNotFound {
		t.Errorf("GetPreviewSession() should return ErrPreviewSessionNotFound, got: %v", err)
	}

	// Create session
	p.CreatePreviewSession("test", "backup")

	// Get session
	session, err := p.GetPreviewSession()
	if err != nil {
		t.Fatalf("GetPreviewSession() returned error: %v", err)
	}

	if session.Theme != "test" {
		t.Errorf("Session theme = %s, want 'test'", session.Theme)
	}
}

func TestEndPreviewSession(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Test ending non-existent session
	err := p.EndPreviewSession()
	if err != ErrPreviewSessionNotFound {
		t.Errorf("EndPreviewSession() should return ErrPreviewSessionNotFound, got: %v", err)
	}

	// Create and end session
	p.CreatePreviewSession("test", "backup")
	err = p.EndPreviewSession()
	if err != nil {
		t.Fatalf("EndPreviewSession() returned error: %v", err)
	}

	// Verify session is gone
	_, err = p.GetPreviewSession()
	if err != ErrPreviewSessionNotFound {
		t.Error("Session should be gone after EndPreviewSession()")
	}
}

func TestHasActivePreviewSession(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// No session initially
	hasSession, err := p.HasActivePreviewSession()
	if err != nil {
		t.Fatalf("HasActivePreviewSession() returned error: %v", err)
	}
	if hasSession {
		t.Error("HasActivePreviewSession() should return false initially")
	}

	// Create session
	p.CreatePreviewSession("test", "backup")

	hasSession, err = p.HasActivePreviewSession()
	if err != nil {
		t.Fatalf("HasActivePreviewSession() returned error: %v", err)
	}
	if !hasSession {
		t.Error("HasActivePreviewSession() should return true after creating session")
	}
}

func TestCreatePreviewSessionAlreadyActive(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Create first session
	_, err := p.CreatePreviewSession("first", "backup1")
	if err != nil {
		t.Fatalf("First CreatePreviewSession() returned error: %v", err)
	}

	// Try to create second session while first is active
	_, err = p.CreatePreviewSession("second", "backup2")
	if err != ErrPreviewSessionActive {
		t.Errorf("CreatePreviewSession() should return ErrPreviewSessionActive, got: %v", err)
	}
}

func TestSetPreviewSessionPID(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Test with no session
	err := p.SetPreviewSessionPID(12345)
	if err != ErrPreviewSessionNotFound {
		t.Errorf("SetPreviewSessionPID() should return ErrPreviewSessionNotFound, got: %v", err)
	}

	// Create session and set PID
	p.CreatePreviewSession("test", "backup")
	err = p.SetPreviewSessionPID(12345)
	if err != nil {
		t.Fatalf("SetPreviewSessionPID() returned error: %v", err)
	}

	// Verify PID was set
	session, _ := p.GetPreviewSession()
	if session.SubshellPID != 12345 {
		t.Errorf("Session PID = %d, want 12345", session.SubshellPID)
	}
}

func TestSaveAndRestorePreviewRCBackup(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	rcContent := "# Original zshrc\nexport PATH=$PATH:/usr/local/bin"

	// Save backup
	backupPath, err := p.SavePreviewRCBackup(rcContent)
	if err != nil {
		t.Fatalf("SavePreviewRCBackup() returned error: %v", err)
	}

	if backupPath == "" {
		t.Error("Backup path is empty")
	}

	// Restore backup
	restored, err := p.RestorePreviewRCBackup(backupPath)
	if err != nil {
		t.Fatalf("RestorePreviewRCBackup() returned error: %v", err)
	}

	if restored != rcContent {
		t.Errorf("Restored content = %s, want %s", restored, rcContent)
	}
}

func TestCleanupPreviewDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Create session
	p.CreatePreviewSession("test", "backup")

	// Cleanup
	err := p.CleanupPreviewDirectory()
	if err != nil {
		t.Fatalf("CleanupPreviewDirectory() returned error: %v", err)
	}

	// Verify session is gone
	_, err = p.GetPreviewSession()
	if err != ErrPreviewSessionNotFound {
		t.Error("Session should be gone after cleanup")
	}
}
