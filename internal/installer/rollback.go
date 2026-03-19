// Package installer provides dependency installation and management.
// This file implements rollback functionality.
package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/savanhi/shell/internal/persistence"
	"github.com/savanhi/shell/pkg/shell"
)

// RollbackManager handles rollback of installations.
type RollbackManager struct {
	// persister is the persistence layer.
	persister persistence.Persister

	// context is the installation context.
	context *InstallContext

	// shell is the shell interface.
	shell shell.Shell
}

// NewRollbackManager creates a new rollback manager.
func NewRollbackManager(p persistence.Persister, ctx *InstallContext, s shell.Shell) *RollbackManager {
	return &RollbackManager{
		persister: p,
		context:   ctx,
		shell:     s,
	}
}

// RollbackState represents the state before an installation for rollback.
type RollbackState struct {
	// ID is the unique identifier for this state.
	ID string `json:"id"`

	// CreatedAt is when this state was created.
	CreatedAt time.Time `json:"created_at"`

	// Description is a human-readable description.
	Description string `json:"description"`

	// RCBackupPath is the path to the RC file backup.
	RCBackupPath string `json:"rc_backup_path"`

	// InstalledComponents are components that were installed.
	InstalledComponents []string `json:"installed_components"`

	// InstalledFiles are files that were created.
	InstalledFiles []string `json:"installed_files"`

	// ModifiedFiles are files that were modified.
	ModifiedFiles []ModifiedFile `json:"modified_files"`

	// Checksums are checksums of modified files before modification.
	Checksums map[string]string `json:"checksums"`
}

// ModifiedFile represents a file that was modified.
type ModifiedFile struct {
	// Path is the file path.
	Path string `json:"path"`

	// BackupPath is the path to the backup.
	BackupPath string `json:"backup_path"`

	// Checksum is the original checksum.
	Checksum string `json:"checksum"`
}

// CreateRollbackState creates a rollback state before installation.
func (r *RollbackManager) CreateRollbackState(description string) (*RollbackState, error) {
	state := &RollbackState{
		ID:                  fmt.Sprintf("%d", time.Now().UnixNano()),
		CreatedAt:           time.Now(),
		Description:         description,
		InstalledComponents: []string{},
		InstalledFiles:      []string{},
		ModifiedFiles:       []ModifiedFile{},
		Checksums:           make(map[string]string),
	}

	// Backup RC file
	rcPath, err := r.shell.GetRCPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get RC path: %w", err)
	}

	if _, err := os.Stat(rcPath); err == nil {
		// File exists, create backup
		backupPath := filepath.Join(r.context.ConfigDir, "backups", fmt.Sprintf("rc-%s.backup", state.ID))
		if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create backup directory: %w", err)
		}

		content, err := os.ReadFile(rcPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read RC file: %w", err)
		}

		if err := os.WriteFile(backupPath, content, 0644); err != nil {
			return nil, fmt.Errorf("failed to write backup: %w", err)
		}

		state.RCBackupPath = backupPath
		state.Checksums[rcPath] = simpleChecksum(content)
	}

	return state, nil
}

// AddInstalledComponent adds a component to the rollback state.
func (s *RollbackState) AddInstalledComponent(name string) {
	s.InstalledComponents = append(s.InstalledComponents, name)
}

// AddInstalledFile adds a file to the rollback state.
func (s *RollbackState) AddInstalledFile(path string) {
	s.InstalledFiles = append(s.InstalledFiles, path)
}

// AddModifiedFile adds a modified file to the rollback state.
func (s *RollbackState) AddModifiedFile(path, backupPath, checksum string) {
	s.ModifiedFiles = append(s.ModifiedFiles, ModifiedFile{
		Path:       path,
		BackupPath: backupPath,
		Checksum:   checksum,
	})
}

// Rollback performs a rollback to the given state.
func (r *RollbackManager) Rollback(state *RollbackState) (*RollbackResult, error) {
	result := &RollbackResult{
		Components: []string{},
		Files:      []string{},
	}

	// Remove installed components
	for _, component := range state.InstalledComponents {
		if err := r.uninstallComponent(component); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("failed to uninstall %s: %v", component, err))
		} else {
			result.Components = append(result.Components, component)
		}
	}

	// Remove installed files
	for _, file := range state.InstalledFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("failed to remove %s: %v", file, err))
		} else {
			result.Files = append(result.Files, file)
		}
	}

	// Restore modified files
	for _, modified := range state.ModifiedFiles {
		if modified.BackupPath != "" {
			content, err := os.ReadFile(modified.BackupPath)
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("failed to read backup %s: %v", modified.BackupPath, err))
				continue
			}

			if err := os.WriteFile(modified.Path, content, 0644); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("failed to restore %s: %v", modified.Path, err))
				continue
			}

			result.Files = append(result.Files, modified.Path)
		}
	}

	// Restore RC file
	if state.RCBackupPath != "" {
		rcPath, err := r.shell.GetRCPath()
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("failed to get RC path: %v", err))
		} else {
			if err := r.restoreRCFile(state.RCBackupPath, rcPath); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("failed to restore RC file: %v", err))
			} else {
				result.Files = append(result.Files, rcPath)
			}
		}
	}

	// Remove Savanhi sections from RC file
	rcModifier := NewRCModifier(r.shell, r.context.ConfigDir)
	if err := rcModifier.RemoveAllSections(); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("failed to remove Savanhi sections: %v", err))
	}

	result.Success = len(result.Warnings) == 0
	return result, nil
}

// RollbackToOriginal rolls back to the original state.
func (r *RollbackManager) RollbackToOriginal() (*RollbackResult, error) {
	// Get original backup
	backup, err := r.persister.LoadOriginalBackup()
	if err != nil {
		return nil, fmt.Errorf("failed to load original backup: %w", err)
	}

	result := &RollbackResult{
		Components: []string{},
		Files:      []string{},
	}

	// Restore RC files from original backup
	for rcPath, content := range backup.RCFiles {
		// Expand path
		expandedPath := expandPath(rcPath, r.context.HomeDir)

		if err := os.WriteFile(expandedPath, []byte(content), 0644); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("failed to restore %s: %v", rcPath, err))
			continue
		}

		result.Files = append(result.Files, expandedPath)
	}

	// Remove installed components
	for _, component := range backup.Tools.InstalledBySavanhi {
		if err := r.uninstallComponent(component); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("failed to uninstall %s: %v", component, err))
		} else {
			result.Components = append(result.Components, component)
		}
	}

	// Remove installed fonts
	for _, font := range backup.Fonts.NerdFontsInstalled {
		if err := r.uninstallFont(font); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("failed to uninstall font %s: %v", font, err))
		}
	}

	// Clean up ~/.config/savanhi directory
	configDir := r.context.ConfigDir
	if err := r.cleanupConfigDir(configDir); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("failed to clean config dir: %v", err))
	}

	result.Success = len(result.Warnings) == 0
	return result, nil
}

// restoreRCFile restores an RC file from backup.
func (r *RollbackManager) restoreRCFile(backupPath, targetPath string) error {
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Write atomically
	tempPath := targetPath + ".tmp"
	if err := os.WriteFile(tempPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempPath, targetPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to restore RC file: %w", err)
	}

	return nil
}

// uninstallComponent uninstalls a component.
func (r *RollbackManager) uninstallComponent(name string) error {
	switch name {
	case "oh-my-posh":
		return os.Remove(filepath.Join(r.context.BinDir, "oh-my-posh"))
	case "zoxide":
		return os.Remove(filepath.Join(r.context.BinDir, "zoxide"))
	case "fzf":
		return os.Remove(filepath.Join(r.context.BinDir, "fzf"))
	case "bat":
		// bat is usually installed via package manager, skip
		return nil
	case "eza":
		// eza is usually installed via package manager, skip
		return nil
	case "zsh-autosuggestions", "zsh-syntax-highlighting":
		// Zsh plugins are handled by PluginInstaller
		pluginInstaller := NewPluginInstaller(r.context, r.shell)
		return pluginInstaller.Uninstall(name)
	default:
		return fmt.Errorf("unknown component: %s", name)
	}
}

// uninstallFont uninstalls a font.
func (r *RollbackManager) uninstallFont(name string) error {
	fontInstaller := NewFontInstaller(r.context)
	return fontInstaller.Uninstall(name)
}

// cleanupConfigDir cleans up the Savanhi config directory.
func (r *RollbackManager) cleanupConfigDir(configDir string) error {
	// Preserve original-backup.json
	backupPath := filepath.Join(configDir, "original-backup.json")
	var backupData []byte
	if data, err := os.ReadFile(backupPath); err == nil {
		backupData = data
	}

	// Remove everything
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(configDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			// Continue even on error
		}
	}

	// Restore backup file
	if len(backupData) > 0 {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(backupPath, backupData, 0644); err != nil {
			return err
		}
	}

	return nil
}

// simpleChecksum creates a simple checksum of content.
func simpleChecksum(data []byte) string {
	var h uint32 = 0
	for _, b := range data {
		h = h*31 + uint32(b)
	}
	return fmt.Sprintf("%08x", h)
}

// expandPath expands ~ and environment variables in a path.
func expandPath(path, homeDir string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	if path == "~" {
		return homeDir
	}
	return os.ExpandEnv(path)
}
