// Package detector provides system detection capabilities.
// This file contains tests for terminal detection.
package detector

import (
	"os"
	"testing"
)

func TestNewTerminalDetector(t *testing.T) {
	detector := NewTerminalDetector()
	if detector == nil {
		t.Error("NewTerminalDetector() returned nil")
	}
}

func TestTerminalDetector_Detect(t *testing.T) {
	detector := NewTerminalDetector()
	info, err := detector.Detect()

	if err != nil {
		t.Errorf("Detect() returned error: %v", err)
	}

	if info == nil {
		t.Fatal("Detect() returned nil TerminalInfo")
	}

	// Terminal type should be set (even if unknown)
	if info.Type == "" {
		t.Error("Terminal type is empty")
	}
}

func TestDetectTerminalType(t *testing.T) {
	// Save original env vars
	origTermProgram := os.Getenv("TERM_PROGRAM")
	origAlacritty := os.Getenv("ALACRITTY_WINDOW_ID")
	origKitty := os.Getenv("KITTY_WINDOW_ID")
	origWT := os.Getenv("WT_SESSION")

	defer func() {
		os.Setenv("TERM_PROGRAM", origTermProgram)
		os.Setenv("ALACRITTY_WINDOW_ID", origAlacritty)
		os.Setenv("KITTY_WINDOW_ID", origKitty)
		os.Setenv("WT_SESSION", origWT)
	}()

	tests := []struct {
		name         string
		setupEnv     func()
		expectedType TerminalType
	}{
		{
			name: "iTerm2 detection",
			setupEnv: func() {
				os.Setenv("TERM_PROGRAM", "iTerm.app")
			},
			expectedType: TerminalTypeITerm2,
		},
		{
			name: "Alacritty detection",
			setupEnv: func() {
				os.Setenv("ALACRITTY_WINDOW_ID", "12345")
			},
			expectedType: TerminalTypeAlacritty,
		},
		{
			name: "Kitty detection",
			setupEnv: func() {
				os.Setenv("KITTY_WINDOW_ID", "12345")
			},
			expectedType: TerminalTypeKitty,
		},
		{
			name: "Windows Terminal detection",
			setupEnv: func() {
				os.Setenv("WT_SESSION", "abc123")
			},
			expectedType: TerminalTypeWindowsTerminal,
		},
		{
			name: "VS Code terminal detection",
			setupEnv: func() {
				os.Setenv("TERM_PROGRAM", "vscode")
			},
			expectedType: TerminalTypeVSCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars
			os.Unsetenv("TERM_PROGRAM")
			os.Unsetenv("ALACRITTY_WINDOW_ID")
			os.Unsetenv("KITTY_WINDOW_ID")
			os.Unsetenv("WT_SESSION")
			os.Unsetenv("GNOME_TERMINAL_SCREEN")
			os.Unsetenv("KONSOLE_VERSION")

			// Setup test env
			tt.setupEnv()

			detector := &terminalDetector{}
			termType, _ := detector.detectTerminalType()

			if termType != tt.expectedType {
				t.Errorf("detectTerminalType() = %s, want %s", termType, tt.expectedType)
			}
		})
	}
}

func TestDetectTrueColorSupport(t *testing.T) {
	// Save original env vars
	origColorterm := os.Getenv("COLORTERM")
	origTerm := os.Getenv("TERM")
	origTermProgram := os.Getenv("TERM_PROGRAM")

	defer func() {
		os.Setenv("COLORTERM", origColorterm)
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origTermProgram)
	}()

	tests := []struct {
		name       string
		setupEnv   func()
		expectTrue bool
	}{
		{
			name: "COLORTERM=truecolor",
			setupEnv: func() {
				os.Setenv("COLORTERM", "truecolor")
			},
			expectTrue: true,
		},
		{
			name: "COLORTERM=24bit",
			setupEnv: func() {
				os.Setenv("COLORTERM", "24bit")
			},
			expectTrue: true,
		},
		{
			name: "iTerm2 supports true color",
			setupEnv: func() {
				os.Setenv("TERM_PROGRAM", "iTerm.app")
			},
			expectTrue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env
			os.Unsetenv("COLORTERM")
			os.Unsetenv("TERM_PROGRAM")

			tt.setupEnv()

			detector := &terminalDetector{}
			result := detector.detectTrueColorSupport()

			if result != tt.expectTrue {
				t.Errorf("detectTrueColorSupport() = %v, want %v", result, tt.expectTrue)
			}
		})
	}
}

func TestDetectLigatureSupport(t *testing.T) {
	tests := []struct {
		name       string
		termType   TerminalType
		expectTrue bool
	}{
		{"iTerm2 supports ligatures", TerminalTypeITerm2, true},
		{"Alacritty supports ligatures", TerminalTypeAlacritty, true},
		{"Kitty supports ligatures", TerminalTypeKitty, true},
		{"VS Code supports ligatures", TerminalTypeVSCode, true},
		{"Unknown terminal", TerminalTypeUnknown, false},
		{"GNOME Terminal may not", TerminalTypeGNOMETerminal, false},
	}

	detector := &terminalDetector{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.detectLigatureSupport(tt.termType)
			if result != tt.expectTrue {
				t.Errorf("detectLigatureSupport(%s) = %v, want %v", tt.termType, result, tt.expectTrue)
			}
		})
	}
}
