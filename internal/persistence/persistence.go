// Package persistence provides data persistence for Savanhi Shell.
// It handles backups, preferences, history, and preview sessions.
package persistence

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/savanhi/shell/internal/detector"
)

// Common errors returned by the persistence package.
var (
	// ErrBackupNotFound indicates the requested backup does not exist.
	ErrBackupNotFound = errors.New("backup not found")

	// ErrPreferencesNotFound indicates preferences haven't been created yet.
	ErrPreferencesNotFound = errors.New("preferences not found")

	// ErrOriginalBackupExists indicates an original backup already exists.
	ErrOriginalBackupExists = errors.New("original backup already exists")

	// ErrNoOriginalBackup indicates no original backup exists for restore.
	ErrNoOriginalBackup = errors.New("no original backup exists")

	// ErrInvalidBackup indicates the backup is corrupted or invalid.
	ErrInvalidBackup = errors.New("invalid or corrupted backup")

	// ErrHistoryEmpty indicates no history entries exist.
	ErrHistoryEmpty = errors.New("no history entries")

	// ErrPreviewSessionNotFound indicates no active preview session.
	ErrPreviewSessionNotFound = errors.New("preview session not found")

	// ErrPreviewSessionActive indicates a preview session is already active.
	ErrPreviewSessionActive = errors.New("preview session already active")
)

// File paths and constants for the persistence layer.
const (
	// ConfigDirName is the name of the Savanhi config directory.
	ConfigDirName = "savanhi"

	// OriginalBackupFile is the filename for the first-run backup.
	OriginalBackupFile = "original-backup.json"

	// PreferencesFile is the filename for user preferences.
	PreferencesFile = "preferences.json"

	// HistoryFile is the filename for action history.
	HistoryFile = "history.json"

	// BackupsDir is the directory name for timestamped backups.
	BackupsDir = "backups"

	// PreviewDir is the directory name for preview sessions.
	PreviewDir = "preview"

	// ThemesDir is the directory name for cached themes.
	ThemesDir = "themes"

	// LogsDir is the directory name for log files.
	LogsDir = "logs"

	// MaxHistoryEntries is the maximum number of history entries to keep.
	MaxHistoryEntries = 1000

	// MaxBackupRetentionDays is the default backup retention period.
	MaxBackupRetentionDays = 30

	// PreferencesVersion is the current preferences file format version.
	PreferencesVersion = 1
)

// Persister is the interface for persistence operations.
// Implementations handle backup, preferences, history, and preview sessions.
type Persister interface {
	// === Original Backup Operations ===

	// HasOriginalBackup checks if an original backup exists.
	HasOriginalBackup() (bool, error)

	// SaveOriginalBackup creates the original backup with the initial state.
	// Returns ErrOriginalBackupExists if backup already exists.
	SaveOriginalBackup(snapshot *detector.DetectorResult, rcContents map[string]string) error

	// LoadOriginalBackup retrieves the original backup.
	// Returns ErrNoOriginalBackup if no backup exists.
	LoadOriginalBackup() (*OriginalBackup, error)

	// === Preferences Operations ===

	// HasPreferences checks if preferences exist.
	HasPreferences() (bool, error)

	// SavePreferences saves user preferences.
	SavePreferences(prefs *Preferences) error

	// LoadPreferences loads user preferences.
	// Returns ErrPreferencesNotFound if no preferences exist.
	LoadPreferences() (*Preferences, error)

	// ResetPreferences resets preferences to defaults.
	ResetPreferences() (*Preferences, error)

	// === History Operations ===

	// AppendHistory adds a new history entry.
	AppendHistory(entry *HistoryEntry) error

	// LoadHistory loads history entries, limited by the provided count.
	// If limit <= 0, returns all entries.
	LoadHistory(limit int) ([]*HistoryEntry, error)

	// ClearHistory removes all history entries.
	ClearHistory() error

	// === Backup Operations ===

	// CreateBackup creates a timestamped backup of current state.
	CreateBackup(description string, files []string) (*Backup, error)

	// ListBackups lists all available backups.
	ListBackups() ([]*Backup, error)

	// LoadBackup loads a backup by ID.
	// Returns ErrBackupNotFound if backup doesn't exist.
	LoadBackup(id string) (*Backup, error)

	// RestoreBackup restores from a backup by ID.
	// Returns ErrBackupNotFound if backup doesn't exist.
	RestoreBackup(id string) error

	// DeleteBackup removes a backup by ID.
	// Original backup cannot be deleted (returns ErrBackupNotFound).
	DeleteBackup(id string) error

	// CleanupOldBackups removes backups older than retention days.
	CleanupOldBackups(retentionDays int) (int, error)

	// === Preview Session Operations ===

	// CreatePreviewSession creates a new preview session.
	CreatePreviewSession(theme string, rcBackup string) (*PreviewSession, error)

	// GetPreviewSession retrieves the active preview session.
	// Returns ErrPreviewSessionNotFound if no active session.
	GetPreviewSession() (*PreviewSession, error)

	// EndPreviewSession ends the active preview session.
	EndPreviewSession() error

	// === Utility Operations ===

	// GetConfigDir returns the path to the Savanhi config directory.
	GetConfigDir() (string, error)

	// GetBackupDir returns the path to the backups directory.
	GetBackupDir() (string, error)
}

// FilePersister is the default file-based implementation of Persister.
type FilePersister struct {
	// configDir is the path to the Savanhi config directory.
	configDir string

	// xdgConfigHome is XDG_CONFIG_HOME for Linux compatibility.
	xdgConfigHome string

	// userHome is the user's home directory.
	userHome string
}

// NewFilePersister creates a new FilePersister.
func NewFilePersister() (*FilePersister, error) {
	p := &FilePersister{}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	p.userHome = homeDir

	// Get XDG config directory
	p.xdgConfigHome = os.Getenv("XDG_CONFIG_HOME")
	if p.xdgConfigHome == "" {
		p.xdgConfigHome = filepath.Join(homeDir, ".config")
	}

	// Set config directory path
	p.configDir = filepath.Join(p.xdgConfigHome, ConfigDirName)

	return p, nil
}

// NewFilePersisterWithPath creates a FilePersister with a custom config path.
// This is useful for testing.
func NewFilePersisterWithPath(configDir string) (*FilePersister, error) {
	if configDir == "" {
		return nil, errors.New("config directory cannot be empty")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &FilePersister{
		configDir:     configDir,
		xdgConfigHome: filepath.Dir(configDir),
		userHome:      homeDir,
	}, nil
}

// GetConfigDir returns the path to the Savanhi config directory.
func (p *FilePersister) GetConfigDir() (string, error) {
	// Ensure directory exists
	if err := os.MkdirAll(p.configDir, 0755); err != nil {
		return "", err
	}
	return p.configDir, nil
}

// GetBackupDir returns the path to the backups directory.
func (p *FilePersister) GetBackupDir() (string, error) {
	backupDir := filepath.Join(p.configDir, BackupsDir)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}
	return backupDir, nil
}

// ensureConfigDir creates the config directory if it doesn't exist.
func (p *FilePersister) ensureConfigDir() error {
	dirs := []string{
		p.configDir,
		filepath.Join(p.configDir, BackupsDir),
		filepath.Join(p.configDir, PreviewDir),
		filepath.Join(p.configDir, ThemesDir),
		filepath.Join(p.configDir, LogsDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
