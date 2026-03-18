// Package shell provides Bash-specific RC file manipulation.
package shell

import (
	"os"
	"path/filepath"
)

// BashShell implements Shell for Bash.
type BashShell struct {
	BaseShell
}

// NewBashShell creates a new BashShell.
func NewBashShell() (*BashShell, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &BashShell{
		BaseShell: BaseShell{
			Type:    ShellTypeBash,
			Name:    "bash",
			RCFile:  filepath.Join(home, ".bashrc"),
			HomeDir: home,
		},
	}, nil
}

// NewBashShellWithPath creates a new BashShell with a custom RC path.
// This is useful for testing.
func NewBashShellWithPath(rcPath string) (*BashShell, error) {
	home := filepath.Dir(rcPath)
	if home == "" {
		home = "."
	}

	return &BashShell{
		BaseShell: BaseShell{
			Type:    ShellTypeBash,
			Name:    "bash",
			RCFile:  rcPath,
			HomeDir: home,
		},
	}, nil
}

// GetProfileFile returns the path to .bash_profile for login shell settings.
func (s *BashShell) GetProfileFile() string {
	return filepath.Join(s.HomeDir, ".bash_profile")
}

// InjectEnvVariable injects an environment variable into .bashrc or .bash_profile.
func (s *BashShell) InjectEnvVariable(key, value string) error {
	rcPath, _ := s.GetRCPath()

	// Check if file exists
	content := ""
	if data, err := os.ReadFile(rcPath); err == nil {
		content = string(data)
	}

	// Check if variable already exists
	lines := splitLines(content)
	found := false
	for i, line := range lines {
		if len(line) > 0 && line[0] != '#' {
			if hasEnvVar(line, key) {
				lines[i] = formatBashEnvExport(key, value)
				found = true
				break
			}
		}
	}

	if !found {
		lines = append(lines, formatBashEnvExport(key, value))
	}

	// Write atomically
	tempPath := rcPath + ".savanhi-tmp"
	if err := os.WriteFile(tempPath, []byte(joinLines(lines)), 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, rcPath)
}

func formatBashEnvExport(key, value string) string {
	return "export " + key + "=\"" + value + "\""
}
