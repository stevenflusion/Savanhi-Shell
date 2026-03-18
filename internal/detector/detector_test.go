// Package detector provides system detection capabilities.
// This file contains tests for the detector module.
package detector

import (
	"testing"
)

func TestDefaultDetector(t *testing.T) {
	detector := NewDefaultDetector()
	if detector == nil {
		t.Error("NewDefaultDetector() returned nil")
	}
}

func TestOSDetectorInterface(t *testing.T) {
	// Verify OSDetector interface is implemented
	var _ OSDetector = NewOSDetector()
}

func TestShellDetectorInterface(t *testing.T) {
	// Verify ShellDetector interface is implemented
	var _ ShellDetector = NewShellDetector()
}

func TestTerminalDetectorInterface(t *testing.T) {
	// Verify TerminalDetector interface is implemented
	var _ TerminalDetector = NewTerminalDetector()
}

func TestFontDetectorInterface(t *testing.T) {
	// Verify FontDetector interface is implemented
	var _ FontDetector = NewFontDetector()
}

func TestConfigDetectorInterface(t *testing.T) {
	// Verify ConfigDetector interface is implemented
	var _ ConfigDetector = NewConfigDetector()
}
