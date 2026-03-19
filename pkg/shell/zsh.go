// Package shell provides Zsh-specific RC file manipulation.
package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ZshShell implements Shell for Zsh.
type ZshShell struct {
	BaseShell
}

// NewZshShell creates a new ZshShell.
func NewZshShell() (*ZshShell, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &ZshShell{
		BaseShell: BaseShell{
			Type:    ShellTypeZsh,
			Name:    "zsh",
			RCFile:  filepath.Join(home, ".zshrc"),
			HomeDir: home,
		},
	}, nil
}

// NewZshShellWithPath creates a new ZshShell with a custom RC path.
// This is useful for testing.
func NewZshShellWithPath(rcPath string) (*ZshShell, error) {
	home := filepath.Dir(rcPath)
	if home == "" {
		home = "."
	}

	return &ZshShell{
		BaseShell: BaseShell{
			Type:    ShellTypeZsh,
			Name:    "zsh",
			RCFile:  rcPath,
			HomeDir: home,
		},
	}, nil
}

// GetConfigDir returns the Zsh config directory.
func (s *ZshShell) GetConfigDir() string {
	configDir := filepath.Join(s.HomeDir, ".config", "zsh")
	if _, err := os.Stat(configDir); err == nil {
		return configDir
	}
	// Fallback to .zsh dir
	return filepath.Join(s.HomeDir, ".zsh")
}

// GetEnvFile returns the path to .zshenv for environment settings.
func (s *ZshShell) GetEnvFile() string {
	return filepath.Join(s.HomeDir, ".zshenv")
}

// InjectEnvVariable injects an environment variable into .zshenv.
func (s *ZshShell) InjectEnvVariable(key, value string) error {
	envFile := s.GetEnvFile()

	// Check if file exists
	content := ""
	if data, err := os.ReadFile(envFile); err == nil {
		content = string(data)
	}

	// Check if variable already exists
	lines := splitLines(content)
	found := false
	for i, line := range lines {
		if len(line) > 0 && line[0] != '#' {
			if hasEnvVar(line, key) {
				lines[i] = formatEnvExport(key, value)
				found = true
				break
			}
		}
	}

	if !found {
		lines = append(lines, formatEnvExport(key, value))
	}

	// Write atomically
	tempPath := envFile + ".savanhi-tmp"
	if err := os.WriteFile(tempPath, []byte(joinLines(lines)), 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, envFile)
}

// Helper functions

func splitLines(s string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func joinLines(lines []string) string {
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}

func hasEnvVar(line, key string) bool {
	// Check for "export KEY=" pattern
	prefix := "export " + key + "="
	if len(line) >= len(prefix) && line[:len(prefix)] == prefix {
		return true
	}
	// Check for "KEY=" pattern
	prefix2 := key + "="
	if len(line) >= len(prefix2) && line[:len(prefix2)] == prefix2 {
		return true
	}
	return false
}

func formatEnvExport(key, value string) string {
	return "export " + key + "=\"" + value + "\""
}

// HasOhMyZsh detects if Oh My Zsh is installed.
// Returns (true, customPath) if installed, (false, "") otherwise.
// The customPath is the value of $ZSH_CUSTOM if set, otherwise empty string.
func (s *ZshShell) HasOhMyZsh() (bool, string) {
	// Check for ZSH environment variable (set by OMZ)
	zshEnv := os.Getenv("ZSH")
	if zshEnv != "" {
		// ZSH is set, check if it points to a valid OMZ installation
		if _, err := os.Stat(zshEnv); err == nil {
			// Return the ZSH_CUSTOM path if set
			customPath := os.Getenv("ZSH_CUSTOM")
			return true, customPath
		}
	}

	// Check for default OMZ location
	omzPath := filepath.Join(s.HomeDir, ".oh-my-zsh")
	if _, err := os.Stat(omzPath); err == nil {
		// OMZ found at default location
		customPath := os.Getenv("ZSH_CUSTOM")
		return true, customPath
	}

	return false, ""
}

// GetOhMyZshPluginDir returns the Oh My Zsh custom plugins directory.
// This is where custom plugins should be installed.
func (s *ZshShell) GetOhMyZshPluginDir() string {
	// Check for ZSH_CUSTOM environment variable first
	customPath := os.Getenv("ZSH_CUSTOM")
	if customPath != "" {
		return filepath.Join(customPath, "plugins")
	}

	// Check for ZSH environment variable
	zshEnv := os.Getenv("ZSH")
	if zshEnv != "" {
		return filepath.Join(zshEnv, "custom", "plugins")
	}

	// Default location
	return filepath.Join(s.HomeDir, ".oh-my-zsh", "custom", "plugins")
}

// GetZshVersion returns the installed zsh version.
// Returns the version string and an error if zsh is not installed or version cannot be parsed.
func (s *ZshShell) GetZshVersion() (string, error) {
	cmd := exec.Command("zsh", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get zsh version: %w", err)
	}

	// Parse version from output like "zsh 5.8 (x86_64-apple-darwin21.0)"
	// or "zsh 5.8.1"
	outputStr := string(output)
	outputStr = strings.TrimSpace(outputStr)

	// Remove "zsh " prefix
	if strings.HasPrefix(outputStr, "zsh ") {
		outputStr = strings.TrimPrefix(outputStr, "zsh ")
	}

	// Extract version number (everything before space or parenthesis)
	version := ""
	for i, ch := range outputStr {
		if ch == ' ' || ch == '(' {
			version = outputStr[:i]
			break
		}
	}
	if version == "" {
		version = outputStr
	}

	return version, nil
}

// IsZshVersionCompatible checks if the installed zsh version meets the minimum requirement.
// The minimum required version for zsh-autosuggestions and zsh-syntax-highlighting is 4.3.11.
func (s *ZshShell) IsZshVersionCompatible(minVersion string) (bool, error) {
	currentVersion, err := s.GetZshVersion()
	if err != nil {
		return false, err
	}

	return compareVersions(currentVersion, minVersion) >= 0, nil
}

// ParsePluginsArray parses the plugins=() array from .zshrc.
// Returns a slice of plugin names or an empty slice if not found.
func (s *ZshShell) ParsePluginsArray() ([]string, error) {
	rcPath, err := s.GetRCPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read .zshrc: %w", err)
	}

	return parsePluginsFromArray(string(content)), nil
}

// parsePluginsFromArray extracts plugin names from .zshrc content.
func parsePluginsFromArray(content string) []string {
	var plugins []string
	lines := strings.Split(content, "\n")

	// Find the plugins=() line
	inPluginsArray := false
	var pluginsLine strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inPluginsArray {
			// Look for plugins=( start
			if strings.HasPrefix(trimmed, "plugins=(") {
				inPluginsArray = true
				// Check if it's a single line: plugins=(git npm)
				rest := strings.TrimPrefix(trimmed, "plugins=(")

				// Check if closing paren is on same line
				if idx := strings.Index(rest, ")"); idx >= 0 {
					// Single line - extract plugins and we're done
					rest = rest[:idx]
					pluginsLine.WriteString(rest)
					break
				}

				pluginsLine.WriteString(rest)
				pluginsLine.WriteString(" ")
			}
		} else {
			// Continue collecting until )
			if strings.Contains(trimmed, ")") {
				// End of array - extract content before )
				idx := strings.Index(trimmed, ")")
				pluginsLine.WriteString(trimmed[:idx])
				break
			}
			pluginsLine.WriteString(trimmed)
			pluginsLine.WriteString(" ")
		}
	}

	if pluginsLine.Len() == 0 {
		return []string{}
	}

	// Parse plugin names from the collected content
	// They are space-separated in the array
	content = pluginsLine.String()
	// Remove any trailing whitespace and quotes
	content = strings.TrimSpace(content)

	// Split by whitespace
	fields := strings.Fields(content)
	for _, field := range fields {
		field = strings.TrimSpace(field)
		// Remove any remaining parentheses or quotes
		field = strings.TrimSuffix(field, ")")
		field = strings.Trim(field, "\"'")
		if field != "" && field != ")" {
			plugins = append(plugins, field)
		}
	}

	return plugins
}

// AddToPluginsArray adds a plugin to the plugins array in .zshrc.
// Returns an error if the plugin is already in the array.
func (s *ZshShell) AddToPluginsArray(plugin string) error {
	rcPath, err := s.GetRCPath()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create a new .zshrc with plugins array
			newContent := fmt.Sprintf("plugins=(%s)\n", plugin)
			return os.WriteFile(rcPath, []byte(newContent), 0644)
		}
		return fmt.Errorf("failed to read .zshrc: %w", err)
	}

	rcContent := string(content)

	// Check if already in plugins array
	plugins := parsePluginsFromArray(rcContent)
	for _, p := range plugins {
		if p == plugin {
			return nil // Already added
		}
	}

	// Add to existing plugins array
	lines := strings.Split(rcContent, "\n")
	var newLines []string
	added := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !added && strings.HasPrefix(trimmed, "plugins=(") {
			// Check if single-line plugins array
			if strings.Contains(trimmed, ")") {
				// Single line: plugins=(git npm)
				// Insert plugin before )
				idx := strings.Index(trimmed, ")")
				before := trimmed[:idx]
				after := trimmed[idx:]
				newLine := before + " " + plugin + " " + after
				newLines = append(newLines, newLine)
				added = true
			} else {
				// Multi-line plugins array
				newLines = append(newLines, line)
			}
		} else if !added && inPluginsArrayContext(lines, i) {
			// Inside multi-line plugins array, check if this is where we should insert
			if strings.Contains(trimmed, ")") {
				// Insert before closing paren
				newLines = append(newLines, "  "+plugin)
				newLines = append(newLines, line)
				added = true
			} else {
				newLines = append(newLines, line)
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	// If no plugins array found, add one
	if !added {
		// Add at the end of the file
		newLines = append(newLines, "")
		newLines = append(newLines, fmt.Sprintf("plugins=(%s)", plugin))
	}

	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(rcPath, []byte(newContent), 0644)
}

// inPluginsArrayContext checks if line index is inside a multi-line plugins array.
func inPluginsArrayContext(lines []string, currentIdx int) bool {
	// Look backwards for plugins=(
	for i := 0; i < currentIdx; i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "plugins=(") {
			// Check if it was closed on same line
			if strings.Contains(lines[i], ")") {
				return false
			}
			// Check if we've seen the closing )
			for j := i + 1; j < currentIdx; j++ {
				if strings.Contains(lines[j], ")") {
					return false
				}
			}
			return true
		}
	}
	return false
}

// RemoveFromPluginsArray removes a plugin from the plugins array in .zshrc.
func (s *ZshShell) RemoveFromPluginsArray(plugin string) error {
	rcPath, err := s.GetRCPath()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		return fmt.Errorf("failed to read .zshrc: %w", err)
	}

	rcContent := string(content)
	plugins := parsePluginsFromArray(rcContent)

	// Check if plugin is in array
	found := false
	for _, p := range plugins {
		if p == plugin {
			found = true
			break
		}
	}

	if !found {
		return nil // Not in array, nothing to remove
	}

	// Remove from .zshrc
	lines := strings.Split(rcContent, "\n")
	var newLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "plugins=(") {
			// Single line plugin removal
			if strings.Contains(trimmed, ")") {
				// Parse and rebuild without the plugin
				newPlugins := removePluginFromList(plugins, plugin)
				newLine := "plugins=(" + strings.Join(newPlugins, " ") + ")"
				newLines = append(newLines, newLine)
			} else {
				newLines = append(newLines, line)
			}
		} else if strings.Contains(trimmed, plugin) {
			// Skip this line (it's the plugin entry in multi-line array)
			continue
		} else {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(rcPath, []byte(newContent), 0644)
}

// removePluginFromList removes a plugin from the list.
func removePluginFromList(plugins []string, plugin string) []string {
	var result []string
	for _, p := range plugins {
		if p != plugin {
			result = append(result, p)
		}
	}
	return result
}

// compareVersions compares two semantic versions.
// Returns:
//
//	-1 if v1 < v2
//	0 if v1 == v2
//	1 if v1 > v2
func compareVersions(v1, v2 string) int {
	v1Parts := parseVersionParts(v1)
	v2Parts := parseVersionParts(v2)

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(v1Parts) {
			n1 = v1Parts[i]
		}
		if i < len(v2Parts) {
			n2 = v2Parts[i]
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// parseVersionParts parses version string into numeric parts.
func parseVersionParts(version string) []int {
	// Remove any trailing text (e.g., "5.8.1 (x86_64...)")
	if idx := strings.Index(version, " "); idx > 0 {
		version = version[:idx]
	}
	if idx := strings.Index(version, "("); idx > 0 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n := 0
		for _, ch := range part {
			if ch >= '0' && ch <= '9' {
				n = n*10 + int(ch-'0')
			} else {
				break
			}
		}
		result = append(result, n)
	}

	return result
}
