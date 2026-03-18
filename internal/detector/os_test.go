// Package detector provides system detection capabilities.
// This file contains tests for OS detection.
package detector

import (
	"runtime"
	"testing"
)

func TestNewOSDetector(t *testing.T) {
	detector := NewOSDetector()
	if detector == nil {
		t.Error("NewOSDetector() returned nil")
	}
}

func TestOSDetector_Detect(t *testing.T) {
	detector := NewOSDetector()
	info, err := detector.Detect()

	if err != nil {
		t.Errorf("Detect() returned error: %v", err)
	}

	if info == nil {
		t.Fatal("Detect() returned nil OSInfo")
	}

	// Verify architecture is set
	if info.Arch == "" {
		t.Error("Arch is empty")
	}

	// Verify OS type is valid
	validTypes := map[OSType]bool{
		OSTypeMacOS:  true,
		OSTypeLinux:  true,
		OSTypeWSL:    true,
		OSTypeTermux: true,
	}

	if !validTypes[info.Type] && info.Type != OSTypeWindows && info.Type != OSTypeUnknown {
		t.Errorf("Invalid OS type: %s", info.Type)
	}

	// Verify architecture matches runtime
	expectedArch := runtime.GOARCH
	if info.Arch != expectedArch {
		t.Errorf("Arch = %s, want %s", info.Arch, expectedArch)
	}
}

func TestOSDetector_DetectMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS test on non-darwin platform")
	}

	detector := NewOSDetector()
	info, err := detector.Detect()

	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}

	if info.Type != OSTypeMacOS {
		t.Errorf("Type = %s, want %s", info.Type, OSTypeMacOS)
	}

	if info.PackageMgr != "brew" {
		t.Errorf("PackageMgr = %s, want brew", info.PackageMgr)
	}
}

func TestOSDetector_DetectLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux test on non-linux platform")
	}

	detector := NewOSDetector()
	info, err := detector.Detect()

	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}

	// On Linux, should be either Linux, WSL, or Termux
	validTypes := []OSType{OSTypeLinux, OSTypeWSL, OSTypeTermux}
	valid := false
	for _, vt := range validTypes {
		if info.Type == vt {
			valid = true
			break
		}
	}

	if !valid {
		t.Errorf("Type = %s, want one of %v", info.Type, validTypes)
	}
}

func TestParseOSRelease(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping /etc/os-release test on non-linux platform")
	}

	detector := &osDetector{}
	result, err := detector.parseOSRelease()

	if err != nil {
		t.Skipf("Cannot parse /etc/os-release: %v", err)
	}

	// Verify basic fields exist
	if result["ID"] == "" {
		t.Error("ID field is empty")
	}
}

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		name     string
		distro   string
		expected string
	}{
		{"Ubuntu uses apt", "ubuntu", "apt"},
		{"Debian uses apt", "debian", "apt"},
		{"Arch uses pacman", "arch", "pacman"},
		{"Manjaro uses pacman", "manjaro", "pacman"},
		{"Fedora uses dnf", "fedora", "dnf"},
		{"CentOS uses dnf", "centos", "dnf"},
		{"Alpine uses apk", "alpine", "apk"},
		{"Unknown distro", "unknown", "unknown"},
	}

	detector := &osDetector{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.detectPackageManager(tt.distro)
			if result != tt.expected {
				t.Errorf("detectPackageManager(%s) = %s, want %s", tt.distro, result, tt.expected)
			}
		})
	}
}

func TestIsWSL(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The actual result depends on the environment
	_ = isWSL()
}

func TestIsTermux(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The actual result depends on the environment
	_ = isTermux()
}
