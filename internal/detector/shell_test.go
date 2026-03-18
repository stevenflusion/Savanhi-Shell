// Package detector provides system detection capabilities.
// This file contains tests for shell detection.
package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewShellDetector(t *testing.T) {
	detector := NewShellDetector()
	if detector == nil {
		t.Error("NewShellDetector() returned nil")
	}
}

func TestShellDetector_Detect(t *testing.T) {
	detector := NewShellDetector()
	info, err := detector.Detect()

	if err != nil {
		t.Errorf("Detect() returned error: %v", err)
	}

	if info == nil {
		t.Fatal("Detect() returned nil ShellInfo")
	}

	// Verify shell type is valid
	validTypes := map[ShellType]bool{
		ShellTypeZsh:     true,
		ShellTypeBash:    true,
		ShellTypeFish:    true,
		ShellTypePwsh:    true,
		ShellTypeUnknown: true,
	}

	if !validTypes[info.Name] {
		t.Errorf("Invalid shell type: %s", info.Name)
	}
}

func TestGetShellTypeFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected ShellType
	}{
		{"zsh path", "/bin/zsh", ShellTypeZsh},
		{"bash path", "/bin/bash", ShellTypeBash},
		{"fish path", "/usr/bin/fish", ShellTypeFish},
		{"pwsh path", "/usr/bin/pwsh", ShellTypePwsh},
		{"powershell path", "/usr/bin/powershell", ShellTypePwsh},
		{"unknown path", "/usr/bin/unknown", ShellTypeUnknown},
	}

	detector := &shellDetector{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.getShellTypeFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("getShellTypeFromPath(%s) = %s, want %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGetRCFilePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		name      string
		shellType ShellType
		expected  string
	}{
		{"zsh RC", ShellTypeZsh, filepath.Join(homeDir, ".zshrc")},
		{"bash RC", ShellTypeBash, filepath.Join(homeDir, ".bashrc")},
		{"fish RC", ShellTypeFish, filepath.Join(homeDir, ".config", "fish", "config.fish")},
		{"unknown RC", ShellTypeUnknown, ""},
	}

	detector := &shellDetector{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.getRCFilePath(tt.shellType)
			if result != tt.expected {
				t.Errorf("getRCFilePath(%s) = %s, want %s", tt.shellType, result, tt.expected)
			}
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name      string
		shellType ShellType
		contains  string
	}{
		{"zsh config", ShellTypeZsh, "zsh"},
		{"bash config", ShellTypeBash, "bash"},
		{"fish config", ShellTypeFish, "fish"},
		{"unknown config", ShellTypeUnknown, ".config"},
	}

	detector := &shellDetector{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.getConfigDir(tt.shellType)
			if !filepath.IsAbs(result) && result != "" {
				t.Errorf("getConfigDir(%s) = %s, want absolute path", tt.shellType, result)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Skip("Cannot create temp file")
	}
	defer os.Remove(tmpFile.Name())

	detector := &shellDetector{}

	// Test existing file
	if !detector.fileExists(tmpFile.Name()) {
		t.Error("fileExists() returned false for existing file")
	}

	// Test non-existing file
	if detector.fileExists("/non/existent/file") {
		t.Error("fileExists() returned true for non-existing file")
	}
}
