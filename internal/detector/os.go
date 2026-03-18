// Package detector provides system detection capabilities.
// This file implements OS detection for macOS, Linux, WSL, and Termux.
package detector

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// osDetector implements OSDetector for various platforms.
type osDetector struct{}

// NewOSDetector creates a new OS detector.
func NewOSDetector() OSDetector {
	return &osDetector{}
}

// Detect implements OSDetector.Detect.
func (d *osDetector) Detect() (*OSInfo, error) {
	info := &OSInfo{
		Arch:    runtime.GOARCH,
		Version: runtime.GOOS,
	}

	// Detect OS type
	switch runtime.GOOS {
	case "darwin":
		info.Type = OSTypeMacOS
		if err := d.detectMacOSDetails(info); err != nil {
			return nil, fmt.Errorf("failed to detect macOS details: %w", err)
		}
		info.PackageMgr = "brew"

	case "linux":
		// Check if running under WSL
		if isWSL() {
			info.Type = OSTypeWSL
		} else if isTermux() {
			info.Type = OSTypeTermux
		} else {
			info.Type = OSTypeLinux
		}
		if err := d.detectLinuxDetails(info); err != nil {
			return nil, fmt.Errorf("failed to detect Linux details: %w", err)
		}

	case "windows":
		info.Type = OSTypeWindows
		// Windows detection would go here
		// For now, we don't support native Windows
		return nil, fmt.Errorf("native Windows is not yet supported, use WSL")

	default:
		info.Type = OSTypeUnknown
		return info, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return info, nil
}

// detectMacOSDetails fills in macOS-specific information.
func (d *osDetector) detectMacOSDetails(info *OSInfo) error {
	// GetmacOS version from sw_vers
	version, err := d.runCommand("sw_vers", "-productVersion")
	if err == nil {
		info.Version = version
	}

	// Get macOS name
	name, err := d.runCommand("sw_vers", "-productName")
	if err == nil {
		info.PrettyName = name
	}

	// Get build version
	build, err := d.runCommand("sw_vers", "-buildVersion")
	if err == nil {
		info.Codename = build
	}

	return nil
}

// detectLinuxDetails fills in Linux-specific information.
func (d *osDetector) detectLinuxDetails(info *OSInfo) error {
	// Parse /etc/os-release
	osRelease, err := d.parseOSRelease()
	if err != nil {
		// If /etc/os-release doesn't exist, try lsb_release
		info.Distro = "unknown"
		info.Version = "unknown"
		return nil
	}

	info.Distro = osRelease["ID"]
	info.Version = osRelease["VERSION_ID"]
	info.PrettyName = osRelease["PRETTY_NAME"]

	if name, ok := osRelease["VERSION_CODENAME"]; ok {
		info.Codename = name
	}

	// Determine package manager based on distribution
	info.PackageMgr = d.detectPackageManager(info.Distro)

	return nil
}

// parseOSRelease reads and parses /etc/os-release.
func (d *osDetector) parseOSRelease() (map[string]string, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/os-release: %w", err)
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Parse KEY="VALUE" or KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := strings.Trim(parts[1], "\"'")
		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read /etc/os-release: %w", err)
	}

	return result, nil
}

// detectPackageManager returns the package manager for a given distribution.
func (d *osDetector) detectPackageManager(distro string) string {
	switch strings.ToLower(distro) {
	case "ubuntu", "debian", "linuxmint", "pop":
		return "apt"
	case "arch", "manjaro", "endeavouros":
		return "pacman"
	case "fedora", "rhel", "centos", "rocky", "almalinux":
		return "dnf"
	case "opensuse", "opensuse-leap", "opensuse-tumbleweed":
		return "zypper"
	case "alpine":
		return "apk"
	case "gentoo":
		return "emerge"
	default:
		return "unknown"
	}
}

// runCommand executes a command and returns its output.
func (d *osDetector) runCommand(name string, args ...string) (string, error) {
	// This will be replaced with a proper exec call in production
	// For now, return empty to avoid import cycles
	return "", fmt.Errorf("command execution not implemented")
}

// isWSL checks if the current environment is Windows Subsystem for Linux.
func isWSL() bool {
	// Check /proc/version for Microsoft or WSL indicators
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "Microsoft") ||
		strings.Contains(content, "WSL") ||
		strings.Contains(content, "microsoft")
}

// isTermux checks if the current environment is Android Termux.
func isTermux() bool {
	// Check for TERMUX_VERSION environment variable
	return os.Getenv("TERMUX_VERSION") != ""
}
