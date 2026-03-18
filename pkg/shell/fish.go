// Package shell provides Fish-specific RC file manipulation.
package shell

import (
	"os"
	"path/filepath"
	"strings"
)

// FishShell implements Shell for Fish.
type FishShell struct {
	BaseShell
}

// NewFishShell creates a new FishShell.
func NewFishShell() (*FishShell, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &FishShell{
		BaseShell: BaseShell{
			Type:    ShellTypeFish,
			Name:    "fish",
			RCFile:  filepath.Join(home, ".config", "fish", "config.fish"),
			HomeDir: home,
		},
	}, nil
}

// NewFishShellWithPath creates a new FishShell with a custom RC path.
// This is useful for testing.
func NewFishShellWithPath(rcPath string) (*FishShell, error) {
	home := filepath.Dir(rcPath)
	if home == "" {
		home = "."
	}

	return &FishShell{
		BaseShell: BaseShell{
			Type:    ShellTypeFish,
			Name:    "fish",
			RCFile:  rcPath,
			HomeDir: home,
		},
	}, nil
}

// GetConfigDir returns the Fish shell configuration directory.
// Fish uses ~/.config/fish for its configuration.
func (s *FishShell) GetConfigDir() string {
	return filepath.Join(s.HomeDir, ".config", "fish")
}

// InjectEnvVariable injects an environment variable into config.fish.
// Fish uses "set -x VAR_NAME value" syntax instead of "export VAR=value".
func (s *FishShell) InjectEnvVariable(key, value string) error {
	rcPath, _ := s.GetRCPath()

	// Check if file exists
	content := ""
	if data, err := os.ReadFile(rcPath); err == nil {
		content = string(data)
	}

	// Check if variable already exists
	lines := splitLines(content)
	found := false
	for i, line := range lines {
		if len(line) > 0 && line[0] != '#' {
			if hasFishEnvVar(line, key) {
				lines[i] = formatFishEnvVar(key, value)
				found = true
				break
			}
		}
	}

	if !found {
		lines = append(lines, formatFishEnvVar(key, value))
	}

	// Write atomically
	tempPath := rcPath + ".savanhi-tmp"
	if err := os.WriteFile(tempPath, []byte(joinLines(lines)), 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, rcPath)
}

// hasFishEnvVar checks if a line contains a Fish environment variable declaration.
// Matches both "set -x VAR value" and "set -gx VAR value" patterns.
// Returns false for commented lines (lines starting with #).
func hasFishEnvVar(line, key string) bool {
	// Skip commented lines
	trimmedLine := strings.TrimSpace(line)
	if len(trimmedLine) > 0 && trimmedLine[0] == '#' {
		return false
	}

	// Check for "set -x VAR" or "set -gx VAR" pattern
	patterns := []string{
		"set -x " + key + " ",
		"set -x " + key + "\t",
		"set -gx " + key + " ",
		"set -gx " + key + "\t",
		"set -x " + key + "\"",
		"set -gx " + key + "\"",
		"set -x " + key + "'",
		"set -gx " + key + "'",
	}

	lineLower := strings.ToLower(trimmedLine)
	for _, pattern := range patterns {
		if strings.Contains(lineLower, strings.ToLower(pattern)) {
			return true
		}
	}

	// Also check for exact match at end of line (no value)
	if strings.TrimSpace(lineLower) == "set -x "+strings.ToLower(key) ||
		strings.TrimSpace(lineLower) == "set -gx "+strings.ToLower(key) {
		return true
	}

	return false
}

// formatFishEnvVar formats a Fish environment variable declaration.
// Fish uses "set -x VAR_NAME "value"" syntax.
func formatFishEnvVar(key, value string) string {
	escaped := escapeFishValue(value)
	return "set -x " + key + " \"" + escaped + "\""
}

// escapeFishValue escapes special characters for Fish shell.
// Fish requires escaping backslashes and double quotes.
func escapeFishValue(value string) string {
	// Escape backslashes first (before escaping other chars)
	value = strings.ReplaceAll(value, `\`, `\\`)
	// Escape double quotes
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}
