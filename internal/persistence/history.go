// Package persistence provides data persistence for Savanhi Shell.
// This file implements history operations.
package persistence

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// AppendHistory adds a new history entry.
func (p *FilePersister) AppendHistory(entry *HistoryEntry) error {
	if err := p.ensureConfigDir(); err != nil {
		return err
	}

	// Generate ID if not set
	if entry.ID == "" {
		entry.ID = generateID()
	}

	// Set timestamp if not set
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Load existing history
	entries, err := p.LoadHistory(0)
	if err != nil && err != ErrHistoryEmpty {
		return err
	}

	// Append new entry
	entries = append(entries, entry)

	// Trim to max entries
	if len(entries) > MaxHistoryEntries {
		entries = entries[len(entries)-MaxHistoryEntries:]
	}

	// Save history
	return p.saveHistory(entries)
}

// LoadHistory loads history entries, limited by the provided count.
// If limit <= 0, returns all entries.
func (p *FilePersister) LoadHistory(limit int) ([]*HistoryEntry, error) {
	historyPath := filepath.Join(p.configDir, HistoryFile)

	data, err := os.ReadFile(historyPath)
	if os.IsNotExist(err) {
		return []*HistoryEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read history: %w", err)
	}

	var entries []*HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	// Sort by timestamp (newest first)
	slices.SortFunc(entries, func(a, b *HistoryEntry) int {
		return cmp.Compare(b.Timestamp.Unix(), a.Timestamp.Unix())
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

// ClearHistory removes all history entries.
func (p *FilePersister) ClearHistory() error {
	historyPath := filepath.Join(p.configDir, HistoryFile)

	if err := os.Remove(historyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	return nil
}

// saveHistory saves history entries to disk.
func (p *FilePersister) saveHistory(entries []*HistoryEntry) error {
	// Sort by timestamp (oldest first for storage)
	slices.SortFunc(entries, func(a, b *HistoryEntry) int {
		return cmp.Compare(a.Timestamp.Unix(), b.Timestamp.Unix())
	})

	data, err := json.MarshalIndent(entries, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Write atomically
	historyPath := filepath.Join(p.configDir, HistoryFile)
	tempPath := historyPath + ".tmp"

	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write history: %w", err)
	}

	if err := os.Rename(tempPath, historyPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename history: %w", err)
	}

	return nil
}

// GetHistoryByAction returns history entries filtered by action type.
func (p *FilePersister) GetHistoryByAction(actionType ActionType) ([]*HistoryEntry, error) {
	entries, err := p.LoadHistory(0)
	if err != nil {
		return nil, err
	}

	filtered := make([]*HistoryEntry, 0)
	for _, entry := range entries {
		if entry.ActionType == actionType {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// GetHistoryByStatus returns history entries filtered by status.
func (p *FilePersister) GetHistoryByStatus(status ActionStatus) ([]*HistoryEntry, error) {
	entries, err := p.LoadHistory(0)
	if err != nil {
		return nil, err
	}

	filtered := make([]*HistoryEntry, 0)
	for _, entry := range entries {
		if entry.Status == status {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// GetRecentHistory returns history entries from the last n days.
func (p *FilePersister) GetRecentHistory(days int) ([]*HistoryEntry, error) {
	entries, err := p.LoadHistory(0)
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	filtered := make([]*HistoryEntry, 0)

	for _, entry := range entries {
		if entry.Timestamp.After(cutoff) {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// GetRollbackHistory returns history entries that can be rolled back.
func (p *FilePersister) GetRollbackHistory() ([]*HistoryEntry, error) {
	entries, err := p.LoadHistory(0)
	if err != nil {
		return nil, err
	}

	filtered := make([]*HistoryEntry, 0)
	for _, entry := range entries {
		if entry.RollbackAvailable && entry.Status == ActionStatusCompleted {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}
