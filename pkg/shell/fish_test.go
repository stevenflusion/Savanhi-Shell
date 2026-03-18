// Package shell provides tests for Fish shell RC file manipulation.
package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFishShell(t *testing.T) {
	shell, err := NewFishShell()
	if err != nil {
		t.Fatalf("NewFishShell() error = %v", err)
	}

	if shell.Type != ShellTypeFish {
		t.Errorf("NewFishShell().Type = %s, want %s", shell.Type, ShellTypeFish)
	}

	if shell.Name != "fish" {
		t.Errorf("NewFishShell().Name = %s, want fish", shell.Name)
	}

	path, err := shell.GetRCPath()
	if err != nil {
		t.Errorf("GetRCPath() error = %v", err)
		return
	}

	// Should end with config.fish
	if filepath.Base(path) != "config.fish" {
		t.Errorf("GetRCPath() = %s, want path ending with config.fish", path)
	}

	// Should contain .config/fish in path
	if !strings.Contains(path, ".config/fish") && !strings.Contains(path, ".config\\fish") {
		t.Errorf("GetRCPath() = %s, want path containing .config/fish", path)
	}
}

func TestNewFishShellWithPath(t *testing.T) {
	tests := []struct {
		name   string
		rcPath string
	}{
		{
			name:   "custom rc path",
			rcPath: "/tmp/test/config.fish",
		},
		{
			name:   "relative path",
			rcPath: "config.fish",
		},
		{
			name:   "nested path",
			rcPath: "/home/user/.config/fish/config.fish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell, err := NewFishShellWithPath(tt.rcPath)
			if err != nil {
				t.Fatalf("NewFishShellWithPath(%s) error = %v", tt.rcPath, err)
			}

			if shell.RCFile != tt.rcPath {
				t.Errorf("RCFile = %s, want %s", shell.RCFile, tt.rcPath)
			}

			if shell.Type != ShellTypeFish {
				t.Errorf("Type = %s, want %s", shell.Type, ShellTypeFish)
			}

			if shell.Name != "fish" {
				t.Errorf("Name = %s, want fish", shell.Name)
			}
		})
	}
}

func TestFishShell_GetRCPath(t *testing.T) {
	shell, err := NewFishShell()
	if err != nil {
		t.Fatalf("NewFishShell() error = %v", err)
	}

	path, err := shell.GetRCPath()
	if err != nil {
		t.Errorf("GetRCPath() error = %v", err)
		return
	}

	if path == "" {
		t.Error("GetRCPath() returned empty path")
		return
	}

	// Should end with config.fish
	if filepath.Base(path) != "config.fish" {
		t.Errorf("GetRCPath() = %s, want path ending with config.fish", path)
	}
}

func TestFishShell_GetConfigDir(t *testing.T) {
	shell, err := NewFishShell()
	if err != nil {
		t.Fatalf("NewFishShell() error = %v", err)
	}

	configDir := shell.GetConfigDir()

	if configDir == "" {
		t.Error("GetConfigDir() returned empty path")
		return
	}

	// Should end with fish directory
	if filepath.Base(configDir) != "fish" {
		t.Errorf("GetConfigDir() = %s, want path ending with fish", configDir)
	}

	// Should contain .config
	if !strings.Contains(configDir, ".config") && !strings.Contains(configDir, ".config") {
		t.Errorf("GetConfigDir() = %s, want path containing .config", configDir)
	}
}

func TestFishShell_EnsureRCFile(t *testing.T) {
	t.Run("creates config directory and file", func(t *testing.T) {
		tmpDir := t.TempDir()
		rcPath := filepath.Join(tmpDir, ".config", "fish", "config.fish")

		shell, err := NewFishShellWithPath(rcPath)
		if err != nil {
			t.Fatalf("NewFishShellWithPath() error = %v", err)
		}

		err = shell.EnsureRCFile()
		if err != nil {
			t.Fatalf("EnsureRCFile() error = %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(rcPath); os.IsNotExist(err) {
			t.Error("RC file was not created")
		}

		// Verify config directory exists
		configDir := filepath.Dir(rcPath)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			t.Error("Config directory was not created")
		}
	})

	t.Run("does not overwrite existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		rcPath := filepath.Join(tmpDir, ".config", "fish", "config.fish")

		// Create file with content
		os.MkdirAll(filepath.Dir(rcPath), 0755)
		originalContent := "# original content\n"
		os.WriteFile(rcPath, []byte(originalContent), 0644)

		shell, err := NewFishShellWithPath(rcPath)
		if err != nil {
			t.Fatalf("NewFishShellWithPath() error = %v", err)
		}

		err = shell.EnsureRCFile()
		if err != nil {
			t.Fatalf("EnsureRCFile() error = %v", err)
		}

		// Verify content unchanged
		content, err := os.ReadFile(rcPath)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}

		if string(content) != originalContent {
			t.Errorf("Content changed from %q to %q", originalContent, string(content))
		}
	})
}

func TestFishShell_InjectEnvVariable(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		key      string
		value    string
		want     string
	}{
		{
			name:     "inject into empty file",
			existing: "",
			key:      "EDITOR",
			value:    "vim",
			want:     "set -x EDITOR \"vim\"\n",
		},
		{
			name:     "inject into file with content",
			existing: "# existing comment\n",
			key:      "EDITOR",
			value:    "vim",
			want:     "# existing comment\nset -x EDITOR \"vim\"\n",
		},
		{
			name:     "update existing variable",
			existing: "set -x EDITOR \"nano\"\n",
			key:      "EDITOR",
			value:    "vim",
			want:     "set -x EDITOR \"vim\"\n",
		},
		{
			name:     "inject with spaces in value",
			existing: "",
			key:      "PATH",
			value:    "/usr/local/bin /usr/bin",
			want:     "set -x PATH \"/usr/local/bin /usr/bin\"\n",
		},
		{
			name:     "inject value with quotes",
			existing: "",
			key:      "GREETING",
			value:    "hello \"world\"",
			want:     "set -x GREETING \"hello \\\"world\\\"\"\n",
		},
		{
			name:     "inject value with backslashes",
			existing: "",
			key:      "PATH",
			value:    "C:\\Users\\test",
			want:     "set -x PATH \"C:\\\\Users\\\\test\"\n",
		},
		{
			name:     "inject empty value",
			existing: "",
			key:      "EMPTY",
			value:    "",
			want:     "set -x EMPTY \"\"\n",
		},
		{
			name:     "update variable with gx flag",
			existing: "set -gx PATH \"/usr/bin\"\n",
			key:      "PATH",
			value:    "/usr/local/bin",
			want:     "set -x PATH \"/usr/local/bin\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, "config.fish")

			if tt.existing != "" {
				os.WriteFile(rcPath, []byte(tt.existing), 0644)
			}

			shell, err := NewFishShellWithPath(rcPath)
			if err != nil {
				t.Fatalf("NewFishShellWithPath() error = %v", err)
			}

			err = shell.InjectEnvVariable(tt.key, tt.value)
			if err != nil {
				t.Fatalf("InjectEnvVariable() error = %v", err)
			}

			content, err := os.ReadFile(rcPath)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}

			if string(content) != tt.want {
				t.Errorf("Content = %q, want %q", string(content), tt.want)
			}
		})
	}
}

func TestFishShell_InjectEnvVariable_MultipleVariables(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, "config.fish")

	shell, err := NewFishShellWithPath(rcPath)
	if err != nil {
		t.Fatalf("NewFishShellWithPath() error = %v", err)
	}

	// Inject first variable
	err = shell.InjectEnvVariable("EDITOR", "vim")
	if err != nil {
		t.Fatalf("InjectEnvVariable(EDITOR) error = %v", err)
	}

	// Inject second variable
	err = shell.InjectEnvVariable("PAGER", "less")
	if err != nil {
		t.Fatalf("InjectEnvVariable(PAGER) error = %v", err)
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	contentStr := string(content)

	// Both variables should be present
	if !strings.Contains(contentStr, "set -x EDITOR \"vim\"") {
		t.Error("EDITOR variable not found in content")
	}

	if !strings.Contains(contentStr, "set -x PAGER \"less\"") {
		t.Error("PAGER variable not found in content")
	}
}

func TestHasFishEnvVar(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		key   string
		found bool
	}{
		{
			name:  "simple set -x",
			line:  "set -x EDITOR vim",
			key:   "EDITOR",
			found: true,
		},
		{
			name:  "set -x with quotes",
			line:  "set -x EDITOR \"vim\"",
			key:   "EDITOR",
			found: true,
		},
		{
			name:  "set -gx with quotes",
			line:  "set -gx PATH \"/usr/bin\"",
			key:   "PATH",
			found: true,
		},
		{
			name:  "set -x with value at end",
			line:  "set -x FOO",
			key:   "FOO",
			found: true,
		},
		{
			name:  "different variable",
			line:  "set -x EDITOR vim",
			key:   "PAGER",
			found: false,
		},
		{
			name:  "commented line",
			line:  "# set -x EDITOR vim",
			key:   "EDITOR",
			found: false,
		},
		{
			name:  "not a set command",
			line:  "echo hello",
			key:   "EDITOR",
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasFishEnvVar(tt.line, tt.key)
			if got != tt.found {
				t.Errorf("hasFishEnvVar(%q, %q) = %v, want %v", tt.line, tt.key, got, tt.found)
			}
		})
	}
}

func TestFormatFishEnvVar(t *testing.T) {
	tests := []struct {
		key   string
		value string
		want  string
	}{
		{
			key:   "EDITOR",
			value: "vim",
			want:  "set -x EDITOR \"vim\"",
		},
		{
			key:   "PATH",
			value: "/usr/local/bin:/usr/bin",
			want:  "set -x PATH \"/usr/local/bin:/usr/bin\"",
		},
		{
			key:   "EMPTY",
			value: "",
			want:  "set -x EMPTY \"\"",
		},
		{
			key:   "WITH_SPACES",
			value: "value with spaces",
			want:  "set -x WITH_SPACES \"value with spaces\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := formatFishEnvVar(tt.key, tt.value)
			if got != tt.want {
				t.Errorf("formatFishEnvVar(%q, %q) = %q, want %q", tt.key, tt.value, got, tt.want)
			}
		})
	}
}

func TestEscapeFishValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "no escaping needed",
			value: "simple",
			want:  "simple",
		},
		{
			name:  "escape backslash",
			value: "path\\to\\file",
			want:  "path\\\\to\\\\file",
		},
		{
			name:  "escape double quote",
			value: "value \"quoted\"",
			want:  "value \\\"quoted\\\"",
		},
		{
			name:  "escape both",
			value: "path\\to \"quoted\"",
			want:  "path\\\\to \\\"quoted\\\"",
		},
		{
			name:  "empty string",
			value: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeFishValue(tt.value)
			if got != tt.want {
				t.Errorf("escapeFishValue(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestFishShell_ReadRC(t *testing.T) {
	t.Run("read existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		rcPath := filepath.Join(tmpDir, "config.fish")
		content := "# Fish config\nset -x EDITOR vim\n"

		shell, err := NewFishShellWithPath(rcPath)
		if err != nil {
			t.Fatalf("NewFishShellWithPath() error = %v", err)
		}

		// Create file
		os.WriteFile(rcPath, []byte(content), 0644)

		got, err := shell.ReadRC()
		if err != nil {
			t.Fatalf("ReadRC() error = %v", err)
		}

		if got != content {
			t.Errorf("ReadRC() = %q, want %q", got, content)
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		rcPath := filepath.Join(tmpDir, "config.fish")

		shell, err := NewFishShellWithPath(rcPath)
		if err != nil {
			t.Fatalf("NewFishShellWithPath() error = %v", err)
		}

		_, err = shell.ReadRC()
		if err != ErrRCNotFound {
			t.Errorf("ReadRC() error = %v, want ErrRCNotFound", err)
		}
	})
}

func TestFishShell_WriteRC(t *testing.T) {
	t.Run("write to new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		rcPath := filepath.Join(tmpDir, ".config", "fish", "config.fish")

		shell, err := NewFishShellWithPath(rcPath)
		if err != nil {
			t.Fatalf("NewFishShellWithPath() error = %v", err)
		}

		content := "# Fish config\nset -x EDITOR vim\n"
		err = shell.WriteRC(content)
		if err != nil {
			t.Fatalf("WriteRC() error = %v", err)
		}

		// Verify file was created
		got, err := os.ReadFile(rcPath)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}

		if string(got) != content {
			t.Errorf("Content = %q, want %q", string(got), content)
		}
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		rcPath := filepath.Join(tmpDir, "config.fish")

		os.WriteFile(rcPath, []byte("old content"), 0644)

		shell, err := NewFishShellWithPath(rcPath)
		if err != nil {
			t.Fatalf("NewFishShellWithPath() error = %v", err)
		}

		newContent := "new content\n"
		err = shell.WriteRC(newContent)
		if err != nil {
			t.Fatalf("WriteRC() error = %v", err)
		}

		got, err := os.ReadFile(rcPath)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}

		if string(got) != newContent {
			t.Errorf("Content = %q, want %q", string(got), newContent)
		}
	})
}

func TestFishShell_Backup(t *testing.T) {
	t.Run("backup existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		rcPath := filepath.Join(tmpDir, "config.fish")
		content := "# Fish config\n"

		os.WriteFile(rcPath, []byte(content), 0644)

		shell, err := NewFishShellWithPath(rcPath)
		if err != nil {
			t.Fatalf("NewFishShellWithPath() error = %v", err)
		}

		backupPath, err := shell.Backup()
		if err != nil {
			t.Fatalf("Backup() error = %v", err)
		}

		// Verify backup exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Error("Backup file was not created")
		}

		// Verify content
		backupContent, err := os.ReadFile(backupPath)
		if err != nil {
			t.Fatalf("ReadFile(backup) error = %v", err)
		}

		if string(backupContent) != content {
			t.Errorf("Backup content = %q, want %q", string(backupContent), content)
		}
	})

	t.Run("backup non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		rcPath := filepath.Join(tmpDir, "config.fish")

		shell, err := NewFishShellWithPath(rcPath)
		if err != nil {
			t.Fatalf("NewFishShellWithPath() error = %v", err)
		}

		_, err = shell.Backup()
		if err != ErrRCNotFound {
			t.Errorf("Backup() error = %v, want ErrRCNotFound", err)
		}
	})
}
