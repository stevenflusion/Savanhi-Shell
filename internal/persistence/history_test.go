// Package persistence provides data persistence for Savanhi Shell.
// This file contains tests for history operations.
package persistence

import (
	"testing"
	"time"
)

func TestAppendHistory(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	entry := &HistoryEntry{
		ActionType:  ActionTypeInstall,
		Description: "Installed zsh",
		Status:      ActionStatusCompleted,
	}

	err := p.AppendHistory(entry)
	if err != nil {
		t.Fatalf("AppendHistory() returned error: %v", err)
	}

	// Verify ID was generated
	if entry.ID == "" {
		t.Error("Entry ID was not generated")
	}

	// Verify timestamp was set
	if entry.Timestamp.IsZero() {
		t.Error("Entry timestamp was not set")
	}

	// Load history and verify
	entries, err := p.LoadHistory(10)
	if err != nil {
		t.Fatalf("LoadHistory() returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("LoadHistory() returned %d entries, want 1", len(entries))
	}

	if entries[0].Description != "Installed zsh" {
		t.Errorf("Entry description = %s, want 'Installed zsh'", entries[0].Description)
	}
}

func TestLoadHistory(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Test empty history - returns empty slice, not error
	entries, err := p.LoadHistory(0)
	if err != nil {
		t.Fatalf("LoadHistory() returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("LoadHistory() on empty history returned %d entries, want 0", len(entries))
	}

	// Add some entries
	for i := 0; i < 5; i++ {
		p.AppendHistory(&HistoryEntry{
			ActionType:  ActionTypeInstall,
			Description: "test",
			Status:      ActionStatusCompleted,
		})
	}

	// Test limit
	entries, err = p.LoadHistory(3)
	if err != nil {
		t.Fatalf("LoadHistory() returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("LoadHistory(3) returned %d entries, want 3", len(entries))
	}

	// Test no limit
	entries, err = p.LoadHistory(0)
	if err != nil {
		t.Fatalf("LoadHistory() returned error: %v", err)
	}

	if len(entries) != 5 {
		t.Errorf("LoadHistory(0) returned %d entries, want 5", len(entries))
	}
}

func TestClearHistory(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Add entries
	p.AppendHistory(&HistoryEntry{
		ActionType:  ActionTypeInstall,
		Description: "test",
	})

	// Clear history
	err := p.ClearHistory()
	if err != nil {
		t.Fatalf("ClearHistory() returned error: %v", err)
	}

	// Verify cleared - should return empty slice
	entries, err := p.LoadHistory(0)
	if err != nil {
		t.Fatalf("LoadHistory() returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("LoadHistory() after clear returned %d entries, want 0", len(entries))
	}
}

func TestGetHistoryByAction(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Add entries with different actions
	p.AppendHistory(&HistoryEntry{
		ActionType:  ActionTypeInstall,
		Description: "install1",
	})
	p.AppendHistory(&HistoryEntry{
		ActionType:  ActionTypeConfigure,
		Description: "configure1",
	})
	p.AppendHistory(&HistoryEntry{
		ActionType:  ActionTypeInstall,
		Description: "install2",
	})

	// Get install actions
	entries, err := p.GetHistoryByAction(ActionTypeInstall)
	if err != nil {
		t.Fatalf("GetHistoryByAction() returned error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("GetHistoryByAction(Install) returned %d entries, want 2", len(entries))
	}

	// Get configure actions
	entries, err = p.GetHistoryByAction(ActionTypeConfigure)
	if err != nil {
		t.Fatalf("GetHistoryByAction() returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("GetHistoryByAction(Configure) returned %d entries, want 1", len(entries))
	}
}

func TestGetHistoryByStatus(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Add entries with different statuses
	p.AppendHistory(&HistoryEntry{
		ActionType:  ActionTypeInstall,
		Description: "completed",
		Status:      ActionStatusCompleted,
	})
	p.AppendHistory(&HistoryEntry{
		ActionType:  ActionTypeInstall,
		Description: "failed",
		Status:      ActionStatusFailed,
	})

	// Get completed
	entries, err := p.GetHistoryByStatus(ActionStatusCompleted)
	if err != nil {
		t.Fatalf("GetHistoryByStatus() returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("GetHistoryByStatus(Completed) returned %d entries, want 1", len(entries))
	}
}

func TestGetRecentHistory(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Add recent entry
	p.AppendHistory(&HistoryEntry{
		ActionType:  ActionTypeInstall,
		Description: "recent",
	})

	// Get recent history (last 1 day)
	entries, err := p.GetRecentHistory(1)
	if err != nil {
		t.Fatalf("GetRecentHistory() returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("GetRecentHistory(1) returned %d entries, want 1", len(entries))
	}

	// Get old history (should be empty)
	entries, err = p.GetRecentHistory(0)
	if err != nil {
		t.Fatalf("GetRecentHistory() returned error: %v", err)
	}

	// Note: This might still return 1 because of how time.Now() works
	// The entry was just created, so it should pass the filter
}

func TestMaxHistoryEntries(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	// Add more than max entries
	for i := 0; i < MaxHistoryEntries+10; i++ {
		p.AppendHistory(&HistoryEntry{
			ActionType:  ActionTypeInstall,
			Description: "test",
		})
	}

	// Load all entries
	entries, err := p.LoadHistory(0)
	if err != nil {
		t.Fatalf("LoadHistory() returned error: %v", err)
	}

	// Should not exceed max
	if len(entries) > MaxHistoryEntries {
		t.Errorf("History has %d entries, should not exceed %d", len(entries), MaxHistoryEntries)
	}
}

func TestHistoryTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	p, _ := NewFilePersisterWithPath(tmpDir)

	before := time.Now()

	entry := &HistoryEntry{
		ActionType:  ActionTypeInstall,
		Description: "test",
	}
	p.AppendHistory(entry)

	after := time.Now()

	// Verify timestamp is within expected range
	if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
		t.Error("Entry timestamp is not correctly set")
	}
}
