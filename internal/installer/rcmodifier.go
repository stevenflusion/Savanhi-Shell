// Package installer provides dependency installation and management.
// This file implements RC file modification for shell configuration.
package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/savanhi/shell/pkg/shell"
)

// RCModifier handles shell RC file modifications.
type RCModifier struct {
	// shell is the shell interface.
	shell shell.Shell

	// backupDir is the directory for RC file backups.
	backupDir string
}

// NewRCModifier creates a new RC modifier.
func NewRCModifier(s shell.Shell, backupDir string) *RCModifier {
	return &RCModifier{
		shell:     s,
		backupDir: backupDir,
	}
}

// SectionMarker represents a marked section in an RC file.
type SectionMarker struct {
	// Name is the section identifier.
	Name string

	// Content is the section content.
	Content string

	// Description is a human-readable description.
	Description string
}

// Common Savanhi section markers.
const (
	// Main Savanhi configuration section.
	SectionSavanhiMain = "savanhi-main"

	// Oh-my-posh initialization.
	SectionOhMyPosh = "savanhi-omp"

	// Zoxide initialization.
	SectionZoxide = "savanhi-zoxide"

	// FZF initialization.
	SectionFZF = "savanhi-fzf"

	// Bat aliases.
	SectionBat = "savanhi-bat"

	// Eza aliases.
	SectionEza = "savanhi-eza"

	// Aliases section.
	SectionAliases = "savanhi-aliases"

	// Path modifications.
	SectionPath = "savanhi-path"
)

// Backup creates a backup of the RC file.
func (m *RCModifier) Backup() (string, error) {
	rcPath, err := m.shell.GetRCPath()
	if err != nil {
		return "", fmt.Errorf("failed to get RC path: %w", err)
	}

	// Read current content
	content, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, no backup needed
			return "", nil
		}
		return "", fmt.Errorf("failed to read RC file: %w", err)
	}

	// Create backup
	backupPath := filepath.Join(m.backupDir, filepath.Base(rcPath)+".backup")
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}

// InjectSection injects a marked section into the RC file.
func (m *RCModifier) InjectSection(marker, content string) error {
	// Ensure RC file exists
	if err := m.shell.EnsureRCFile(); err != nil {
		return fmt.Errorf("failed to ensure RC file: %w", err)
	}

	// Inject using shell interface
	if err := m.shell.InjectSection(marker, content); err != nil {
		return fmt.Errorf("failed to inject section: %w", err)
	}

	return nil
}

// RemoveSection removes a marked section from the RC file.
func (m *RCModifier) RemoveSection(marker string) error {
	if err := m.shell.RemoveSection(marker); err != nil {
		return fmt.Errorf("failed to remove section: %w", err)
	}

	return nil
}

// HasSection checks if a marked section exists.
func (m *RCModifier) HasSection(marker string) (bool, error) {
	return m.shell.HasSection(marker)
}

// GetSection returns the content of a marked section.
func (m *RCModifier) GetSection(marker string) (string, error) {
	return m.shell.GetSection(marker)
}

// InjectOhMyPosh injects oh-my-posh initialization.
func (m *RCModifier) InjectOhMyPosh(themePath string) error {
	content := fmt.Sprintf(`# Initialize Oh My Posh
eval "$(oh-my-posh init %s --config '%s')"`+"\n", m.shell.GetName(), themePath)

	return m.InjectSection(SectionOhMyPosh, content)
}

// InjectZoxide injects zoxide initialization.
func (m *RCModifier) InjectZoxide() error {
	shellName := m.shell.GetName()
	content := fmt.Sprintf(`# Initialize zoxide (smart cd command)
eval "$(%s init %s)"`+"\n", "zoxide", shellName)

	return m.InjectSection(SectionZoxide, content)
}

// InjectFZF injects FZF initialization.
func (m *RCModifier) InjectFZF() error {
	content := `# Initialize FZF (fuzzy finder)
[ -f ~/.fzf.%s ] && source ~/.fzf.%s
` + "\n"

	shellName := m.shell.GetName()
	content = fmt.Sprintf(content, shellName, shellName)

	return m.InjectSection(SectionFZF, content)
}

// InjectBatAliases injects bat aliases.
func (m *RCModifier) InjectBatAliases() error {
	content := `# Bat aliases (better cat)
alias cat='bat --paging=never'
alias catp='bat --paging=auto'
` + "\n"

	return m.InjectSection(SectionBat, content)
}

// InjectEzaAliases injects eza aliases.
func (m *RCModifier) InjectEzaAliases() error {
	content := `# Eza aliases (better ls)
alias ls='eza --icons=auto'
alias ll='eza --icons=auto --long'
alias la='eza --icons=auto --long --all'
alias lt='eza --icons=auto --tree'
` + "\n"

	return m.InjectSection(SectionEza, content)
}

// InjectPath injects PATH modifications.
func (m *RCModifier) InjectPath(binDir string) error {
	// Expand ~ to home directory
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(binDir, "~/") {
		binDir = filepath.Join(home, binDir[2:])
	}

	content := fmt.Sprintf(`# Add Savanhi binaries to PATH
export PATH="%s:$PATH"
`, binDir)

	return m.InjectSection(SectionPath, content)
}

// InjectMainSection injects the main Savanhi configuration.
func (m *RCModifier) InjectMainSection() error {
	content := `# Savanhi Shell Configuration
# This section is managed by Savanhi Shell.
# Do not edit this section manually; changes may be overwritten.
# See 'savanhi config' for configuration options.
` + "\n"

	return m.InjectSection(SectionSavanhiMain, content)
}

// RemoveAllSections removes all Savanhi sections from the RC file.
func (m *RCModifier) RemoveAllSections() error {
	sections := []string{
		SectionSavanhiMain,
		SectionOhMyPosh,
		SectionZoxide,
		SectionFZF,
		SectionBat,
		SectionEza,
		SectionPath,
	}

	for _, section := range sections {
		if err := m.RemoveSection(section); err != nil {
			// Log but continue
			fmt.Printf("Warning: failed to remove section %s: %v\n", section, err)
		}
	}

	return nil
}

// GetSavanhiSections returns all Savanhi sections in the RC file.
func (m *RCModifier) GetSavanhiSections() (map[string]string, error) {
	sections := []string{
		SectionSavanhiMain,
		SectionOhMyPosh,
		SectionZoxide,
		SectionFZF,
		SectionBat,
		SectionEza,
		SectionPath,
	}

	result := make(map[string]string)
	for _, section := range sections {
		content, err := m.GetSection(section)
		if err != nil {
			continue
		}
		if content != "" {
			result[section] = content
		}
	}

	return result, nil
}

// Restore restores the RC file from backup.
func (m *RCModifier) Restore(backupPath string) error {
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	rcPath, err := m.shell.GetRCPath()
	if err != nil {
		return fmt.Errorf("failed to get RC path: %w", err)
	}

	// Write atomically
	tempPath := rcPath + ".tmp"
	if err := os.WriteFile(tempPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempPath, rcPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to restore RC file: %w", err)
	}

	return nil
}

// GetRCContent returns the current RC file content.
func (m *RCModifier) GetRCContent() (string, error) {
	return m.shell.ReadRC()
}

// SetRCContent sets the RC file content.
func (m *RCModifier) SetRCContent(content string) error {
	return m.shell.WriteRC(content)
}

// PrepareForInstall prepares the RC file for installation.
func (m *RCModifier) PrepareForInstall() error {
	// Create backup first
	if _, err := m.Backup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Ensure RC file exists
	if err := m.shell.EnsureRCFile(); err != nil {
		return fmt.Errorf("failed to ensure RC file: %w", err)
	}

	// Inject main section
	return m.InjectMainSection()
}
