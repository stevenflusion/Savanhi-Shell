// Package persistence provides data persistence for Savanhi Shell.
// This file implements preview session operations.
package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CreatePreviewSession creates a new preview session.
func (p *FilePersister) CreatePreviewSession(theme string, rcBackup string) (*PreviewSession, error) {
	if err := p.ensureConfigDir(); err != nil {
		return nil, err
	}

	// Check if a session already exists
	activeSession, err := p.GetPreviewSession()
	if err == nil && activeSession != nil && activeSession.Active {
		return nil, ErrPreviewSessionActive
	}

	session := &PreviewSession{
		ID:        generateID(),
		CreatedAt: time.Now(),
		Theme:     theme,
		RCBackup:  rcBackup,
		Active:    true,
	}

	// Save session
	if err := p.savePreviewSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

// GetPreviewSession retrieves the active preview session.
// Returns ErrPreviewSessionNotFound if no active session.
func (p *FilePersister) GetPreviewSession() (*PreviewSession, error) {
	sessionPath := filepath.Join(p.configDir, PreviewDir, "session.json")

	data, err := os.ReadFile(sessionPath)
	if os.IsNotExist(err) {
		return nil, ErrPreviewSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read preview session: %w", err)
	}

	var session PreviewSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preview session: %w", err)
	}

	if !session.Active {
		return nil, ErrPreviewSessionNotFound
	}

	return &session, nil
}

// EndPreviewSession ends the active preview session.
func (p *FilePersister) EndPreviewSession() error {
	sessionPath := filepath.Join(p.configDir, PreviewDir, "session.json")

	// Check if session exists
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return ErrPreviewSessionNotFound
	}

	// Remove session file
	if err := os.Remove(sessionPath); err != nil {
		return fmt.Errorf("failed to end preview session: %w", err)
	}

	return nil
}

// UpdatePreviewSession updates the active preview session.
func (p *FilePersister) UpdatePreviewSession(session *PreviewSession) error {
	if err := p.ensureConfigDir(); err != nil {
		return err
	}

	return p.savePreviewSession(session)
}

// SetPreviewSessionPID sets the subshell PID for the active session.
func (p *FilePersister) SetPreviewSessionPID(pid int) error {
	session, err := p.GetPreviewSession()
	if err != nil {
		return err
	}

	session.SubshellPID = pid
	return p.savePreviewSession(session)
}

// savePreviewSession saves the preview session to disk.
func (p *FilePersister) savePreviewSession(session *PreviewSession) error {
	previewDir := filepath.Join(p.configDir, PreviewDir)
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		return fmt.Errorf("failed to create preview directory: %w", err)
	}

	data, err := json.MarshalIndent(session, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal preview session: %w", err)
	}

	// Write atomically
	sessionPath := filepath.Join(previewDir, "session.json")
	tempPath := sessionPath + ".tmp"

	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write preview session: %w", err)
	}

	if err := os.Rename(tempPath, sessionPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename preview session: %w", err)
	}

	return nil
}

// HasActivePreviewSession checks if there's an active preview session.
func (p *FilePersister) HasActivePreviewSession() (bool, error) {
	session, err := p.GetPreviewSession()
	if err == ErrPreviewSessionNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return session != nil && session.Active, nil
}

// GetPreviewRCBackup retrieves the RC file backup from an active session.
func (p *FilePersister) GetPreviewRCBackup() (string, error) {
	session, err := p.GetPreviewSession()
	if err != nil {
		return "", err
	}

	return session.RCBackup, nil
}

// SavePreviewRCBackup saves the current RC file content for preview restoration.
func (p *FilePersister) SavePreviewRCBackup(rcContent string) (string, error) {
	if err := p.ensureConfigDir(); err != nil {
		return "", err
	}

	previewDir := filepath.Join(p.configDir, PreviewDir)
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create preview directory: %w", err)
	}

	// Generate unique backup filename
	backupName := fmt.Sprintf("rc-backup-%d.json", time.Now().Unix())
	backupPath := filepath.Join(previewDir, backupName)

	data, err := json.MarshalIndent(map[string]string{
		"content":   rcContent,
		"timestamp": time.Now().Format(time.RFC3339),
	}, "", " ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal RC backup: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write RC backup: %w", err)
	}

	return backupPath, nil
}

// RestorePreviewRCBackup restores the RC file from a backup.
func (p *FilePersister) RestorePreviewRCBackup(backupPath string) (string, error) {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to read RC backup: %w", err)
	}

	var backup struct {
		Content   string `json:"content"`
		Timestamp string `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &backup); err != nil {
		return "", fmt.Errorf("failed to unmarshal RC backup: %w", err)
	}

	return backup.Content, nil
}

// CleanupPreviewDirectory removes all preview session files.
func (p *FilePersister) CleanupPreviewDirectory() error {
	previewDir := filepath.Join(p.configDir, PreviewDir)

	if err := os.RemoveAll(previewDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to cleanup preview directory: %w", err)
	}

	return nil
}
