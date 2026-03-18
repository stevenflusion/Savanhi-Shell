// Package detector provides system detection capabilities for Savanhi Shell.
// It detects OS, shell, terminal, fonts, and existing configurations.
package detector

import "fmt"

// Detector is the main interface for system detection.
// Implementations provide detection capabilities for different platforms.
type Detector interface {
	// DetectOS returns information about the operating system.
	DetectOS() (*OSInfo, error)

	// DetectShell returns information about the current shell.
	DetectShell() (*ShellInfo, error)

	// DetectTerminal returns information about the terminal emulator.
	DetectTerminal() (*TerminalInfo, error)

	// DetectFonts returns an inventory of installed fonts, focusing on Nerd Fonts.
	DetectFonts() (*FontInventory, error)

	// DetectExistingConfigs returns information about existing configurations
	// like oh-my-posh, starship, etc.
	DetectExistingConfigs() (*ConfigSnapshot, error)

	// DetectAll runs all detections and returns a complete system snapshot.
	DetectAll() (*DetectorResult, error)
}

// DetectorResult contains the complete detection results.
type DetectorResult struct {
	OS              *OSInfo         `json:"os"`
	Shell           *ShellInfo      `json:"shell"`
	Terminal        *TerminalInfo   `json:"terminal"`
	Fonts           *FontInventory  `json:"fonts"`
	ExistingConfigs *ConfigSnapshot `json:"existing_configs"`
}

// DefaultDetector is the primary implementation of the Detector interface.
// It combines platform-specific detectors to provide comprehensive detection.
type DefaultDetector struct {
	osDetector       OSDetector
	shellDetector    ShellDetector
	terminalDetector TerminalDetector
	fontDetector     FontDetector
	configDetector   ConfigDetector
}

// NewDefaultDetector creates a new DefaultDetector with all sub-detectors.
func NewDefaultDetector() *DefaultDetector {
	return &DefaultDetector{
		osDetector:       NewOSDetector(),
		shellDetector:    NewShellDetector(),
		terminalDetector: NewTerminalDetector(),
		fontDetector:     NewFontDetector(),
		configDetector:   NewConfigDetector(),
	}
}

// DetectOS implements Detector.DetectOS.
func (d *DefaultDetector) DetectOS() (*OSInfo, error) {
	return d.osDetector.Detect()
}

// DetectShell implements Detector.DetectShell.
func (d *DefaultDetector) DetectShell() (*ShellInfo, error) {
	return d.shellDetector.Detect()
}

// DetectTerminal implements Detector.DetectTerminal.
func (d *DefaultDetector) DetectTerminal() (*TerminalInfo, error) {
	return d.terminalDetector.Detect()
}

// DetectFonts implements Detector.DetectFonts.
func (d *DefaultDetector) DetectFonts() (*FontInventory, error) {
	return d.fontDetector.Detect()
}

// DetectExistingConfigs implements Detector.DetectExistingConfigs.
func (d *DefaultDetector) DetectExistingConfigs() (*ConfigSnapshot, error) {
	return d.configDetector.Detect()
}

// DetectAll implements Detector.DetectAll.
func (d *DefaultDetector) DetectAll() (*DetectorResult, error) {
	result := &DetectorResult{}

	var err error

	result.OS, err = d.DetectOS()
	if err != nil {
		return nil, fmt.Errorf("failed to detect OS: %w", err)
	}

	result.Shell, err = d.DetectShell()
	if err != nil {
		return nil, fmt.Errorf("failed to detect shell: %w", err)
	}

	result.Terminal, err = d.DetectTerminal()
	if err != nil {
		return nil, fmt.Errorf("failed to detect terminal: %w", err)
	}

	result.Fonts, err = d.DetectFonts()
	if err != nil {
		return nil, fmt.Errorf("failed to detect fonts: %w", err)
	}

	result.ExistingConfigs, err = d.DetectExistingConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to detect existing configs: %w", err)
	}

	return result, nil
}
