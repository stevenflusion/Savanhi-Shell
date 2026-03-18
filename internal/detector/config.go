// Package detector provides system detection capabilities.
// This file implements detection of existing configurations (oh-my-posh, starship, etc.).
package detector

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Savanhi marker constants for RC file sections.
const (
	// SavanhiStartMarker is the start marker for Savanhi-managed sections.
	SavanhiStartMarker = "# >>> savanhi-shell >>>"
	// SavanhiEndMarker is the end marker for Savanhi-managed sections.
	SavanhiEndMarker = "# <<< savanhi-shell <<<"
)

// configDetector implements ConfigDetector interface.
type configDetector struct{}

// NewConfigDetector creates a new config detector.
func NewConfigDetector() ConfigDetector {
	return &configDetector{}
}

// Detect implements ConfigDetector.Detect.
func (d *configDetector) Detect() (*ConfigSnapshot, error) {
	snapshot := &ConfigSnapshot{
		DetectedAt: time.Now(),
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return snapshot, nil // Return empty snapshot on error
	}

	// Detect oh-my-posh
	snapshot.HasOhMyPosh, snapshot.OhMyPoshConfigPath = d.detectOhMyPosh(homeDir)

	// Detect starship
	snapshot.HasStarship, snapshot.StarshipConfigPath = d.detectStarship(homeDir)

	// Detect installed tools
	snapshot.HasZoxide = d.isCommandAvailable("zoxide")
	snapshot.HasFzf = d.isCommandAvailable("fzf")
	snapshot.HasBat = d.isCommandAvailable("bat")
	snapshot.HasEza = d.isCommandAvailable("eza")

	// Detect Savanhi markers in RC files
	d.detectSavanhiMarkers(homeDir, snapshot)

	// Detect theme/font/color settings from existing configs
	d.detectThemeSettings(homeDir, snapshot)

	return snapshot, nil
}

// detectOhMyPosh checks if oh-my-posh is installed and configured.
func (d *configDetector) detectOhMyPosh(homeDir string) (bool, string) {
	// Check for oh-my-posh binary
	if !d.isCommandAvailable("oh-my-posh") {
		return false, ""
	}

	// Check common config locations
	configPaths := []string{
		filepath.Join(homeDir, ".config", "oh-my-posh", "config.json"),
		filepath.Join(homeDir, ".config", "oh-my-posh", "theme.json"),
		filepath.Join(homeDir, ".oh-my-posh", "config.json"),
		filepath.Join(homeDir, ".cache", "oh-my-posh", "config.json"),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return true, path
		}
	}

	// oh-my-posh is installed but config not found in standard locations
	return true, ""
}

// detectStarship checks if starship is installed and configured.
func (d *configDetector) detectStarship(homeDir string) (bool, string) {
	// Check for starship binary
	if !d.isCommandAvailable("starship") {
		return false, ""
	}

	// Check common config locations
	configPaths := []string{
		filepath.Join(homeDir, ".config", "starship.toml"),
		filepath.Join(homeDir, ".config", "starship", "starship.toml"),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return true, path
		}
	}

	// starship is installed but config not found in standard locations
	return true, ""
}

// detectSavanhiMarkers checks for Savanhi markers in RC files.
func (d *configDetector) detectSavanhiMarkers(homeDir string, snapshot *ConfigSnapshot) {
	// Check common RC files for Savanhi markers
	rcFiles := []string{
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".config", "fish", "config.fish"),
	}

	for _, rcFile := range rcFiles {
		if content, err := d.readFileContent(rcFile); err == nil {
			if hasMarkers, markerContent := d.parseSavanhiMarkers(content); hasMarkers {
				snapshot.HasSavanhiMarkers = true
				snapshot.SavanhiMarkerContent = markerContent
				return
			}
		}
	}
}

// parseSavanhiMarkers extracts content between Savanhi markers.
func (d *configDetector) parseSavanhiMarkers(content string) (bool, string) {
	startIdx := strings.Index(content, SavanhiStartMarker)
	endIdx := strings.Index(content, SavanhiEndMarker)

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return false, ""
	}

	// Extract content between markers
	markerContent := content[startIdx+len(SavanhiStartMarker) : endIdx]
	return true, strings.TrimSpace(markerContent)
}

// detectThemeSettings attempts to detect theme, font, and color settings.
func (d *configDetector) detectThemeSettings(homeDir string, snapshot *ConfigSnapshot) {
	// Try to detect from oh-my-posh config
	if snapshot.OhMyPoshConfigPath != "" {
		d.parseOhMyPoshConfig(snapshot.OhMyPoshConfigPath, snapshot)
	}

	// Try to detect from starship config
	if snapshot.StarshipConfigPath != "" {
		d.parseStarshipConfig(snapshot.StarshipConfigPath, snapshot)
	}

	// Try to detect from Savanhi markers
	if snapshot.HasSavanhiMarkers {
		d.parseSavanhiContent(snapshot.SavanhiMarkerContent, snapshot)
	}
}

// parseOhMyPoshConfig extracts theme info from oh-my-posh config.
func (d *configDetector) parseOhMyPoshConfig(configPath string, snapshot *ConfigSnapshot) {
	// Read and parse the config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	// Look for theme references
	contentStr := string(content)

	// Check for theme block
	// oh-my-posh themes reference: "theme": "name" or console title
	if strings.Contains(contentStr, "\"theme\"") {
		// Extract theme name (simplified parsing)
		lines := strings.Split(contentStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "\"theme\"") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					theme := strings.TrimSpace(strings.Trim(parts[1], "\","))
					snapshot.DetectedTheme = theme
					break
				}
			}
		}
	}
}

// parseStarshipConfig extracts theme info from starship config.
func (d *configDetector) parseStarshipConfig(configPath string, snapshot *ConfigSnapshot) {
	// Read and parse the TOML config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	contentStr := string(content)

	// Look for custom settings
	// Starship uses TOML format
	if strings.Contains(contentStr, "palette") {
		lines := strings.Split(contentStr, "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "palette") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					palette := strings.TrimSpace(strings.Trim(parts[1], "\""))
					snapshot.DetectedColorScheme = palette
					break
				}
			}
		}
	}
}

// parseSavanhiContent extracts settings from Savanhi marker content.
func (d *configDetector) parseSavanhiContent(content string, snapshot *ConfigSnapshot) {
	if content == "" {
		return
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for font settings
		if strings.Contains(line, "font_family") || strings.Contains(line, "FONT_FAMILY") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				font := strings.TrimSpace(strings.Trim(parts[1], "\"'"))
				if snapshot.DetectedFont == "" {
					snapshot.DetectedFont = font
				}
			}
		}

		// Look for theme settings
		if strings.Contains(line, "theme") && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				theme := strings.TrimSpace(strings.Trim(parts[1], "\"'"))
				if snapshot.DetectedTheme == "" {
					snapshot.DetectedTheme = theme
				}
			}
		}
	}
}

// readFileContent reads the content of a file.
func (d *configDetector) readFileContent(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	return content.String(), scanner.Err()
}

// isCommandAvailable checks if a command is available in PATH.
func (d *configDetector) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
