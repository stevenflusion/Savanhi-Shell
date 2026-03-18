// Package shell provides Zsh-specific RC file manipulation.
package shell

import (
	"os"
	"path/filepath"
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
