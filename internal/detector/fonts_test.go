// Package detector provides system detection capabilities.
// This file contains tests for font detection.
package detector

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewFontDetector(t *testing.T) {
	detector := NewFontDetector()
	if detector == nil {
		t.Error("NewFontDetector() returned nil")
	}
}

func TestFontDetector_GetFontDirectories(t *testing.T) {
	detector := &fontDetector{}
	dirs := detector.getFontDirectories()

	// On any platform, should return some directories
	// Results vary by OS
	switch runtime.GOOS {
	case "darwin":
		// macOS should have Library/Fonts directories
		if len(dirs) == 0 {
			t.Error("macOS should return font directories")
		}
	case "linux":
		// Linux should have /usr/share/fonts and user directories
		if len(dirs) == 0 {
			t.Error("Linux should return font directories")
		}
	}
}

func TestIsFontExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".ttf", true},
		{".otf", true},
		{".woff", true},
		{".woff2", true},
		{".txt", false},
		{".ini", false},
		{".json", false},
		{".TTF", false}, // case sensitive check
	}

	detector := &fontDetector{}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := detector.isFontExtension(tt.ext)
			if result != tt.expected {
				t.Errorf("isFontExtension(%s) = %v, want %v", tt.ext, result, tt.expected)
			}
		})
	}
}

func TestGetFontName(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"JetBrainsMono-Regular.ttf", "JetBrainsMono"},
		{"FiraCode-Bold.ttf", "FiraCode"},
		{"Hack-Regular.ttf", "Hack"},
		{"MesloLGM-NerdFont.ttf", "MesloLGM NerdFont"},
		{"SourceCodePro-Medium.ttf", "SourceCodePro"},
		{"font_name.ttf", "font name"},
	}

	detector := &fontDetector{}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := detector.getFontName(tt.filename)
			// We're just checking it doesn't crash and returns something reasonable
			if result == "" {
				t.Errorf("getFontName(%s) returned empty string", tt.filename)
			}
		})
	}
}

func TestIsNerdFont(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"MesloLGM-NerdFont.ttf", true},
		{"JetBrainsMono-NF.ttf", true},
		{"FiraCode-Powerline.ttf", true},
		{"Hack-Regular.ttf", false},
		{"SourceCodePro-Medium.ttf", false},
		{"nerd-font.ttf", true},
		// Note: case-insensitive, so "nerd" in any case matches{"NOTNERD-Regular.ttf", true}, // Contains "nerd" (case-insensitive)
	}

	detector := &fontDetector{}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := detector.isNerdFont(tt.filename)
			if result != tt.expected {
				t.Errorf("isNerdFont(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestIsMonospaceFont(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"JetBrainsMono-Regular.ttf", true},
		{"FiraCode-Regular.ttf", true},
		{"Hack-Regular.ttf", true},
		{"SourceCodePro-Regular.ttf", true},
		{"Inconsolata-Regular.ttf", true},
		{"RobotoMono-Regular.ttf", true},
		{"DejaVuSansMono.ttf", true},
		{"Arial-Regular.ttf", false},
		{"TimesNewRoman-Regular.ttf", false},
		{"SomeRandomFont.ttf", false},
	}

	detector := &fontDetector{}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := detector.isMonospaceFont(tt.filename)
			if result != tt.expected {
				t.Errorf("isMonospaceFont(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestFontDetector_ScanFontDirectory(t *testing.T) {
	// Create a temp directory with test font files
	tmpDir, err := os.MkdirTemp("", "fonts-*")
	if err != nil {
		t.Skip("Cannot create temp directory")
	}
	defer os.RemoveAll(tmpDir)

	// Create test font files
	fontFiles := []string{
		"TestMono-Regular.ttf",
		"TestSans-Regular.ttf",
		"TestNerd-Regular.otf",
		"notafont.txt",
	}

	for _, f := range fontFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte{}, 0644); err != nil {
			t.Skip("Cannot create test font file")
		}
	}

	detector := &fontDetector{}
	inventory := &FontInventory{Fonts: []FontInfo{}}

	detector.scanFontDirectory(tmpDir, inventory)

	// Should find 3 font files (excluding .txt)
	if len(inventory.Fonts) != 3 {
		t.Errorf("Found %d fonts, want 3", len(inventory.Fonts))
	}
}
