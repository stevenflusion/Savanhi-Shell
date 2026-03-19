// Package shell provides tests for Zsh-specific functionality.
package shell

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasOhMyZsh(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       func(tmpDir string)
		wantHas        bool
		wantCustomPath string
	}{
		{
			name: "OMZ installed at default location",
			setupEnv: func(tmpDir string) {
				omzDir := filepath.Join(tmpDir, ".oh-my-zsh")
				os.MkdirAll(omzDir, 0755)
			},
			wantHas:        true,
			wantCustomPath: "",
		},
		{
			name: "OMZ installed with ZSH_CUSTOM",
			setupEnv: func(tmpDir string) {
				omzDir := filepath.Join(tmpDir, ".oh-my-zsh")
				os.MkdirAll(omzDir, 0755)
				os.Setenv("ZSH_CUSTOM", "/custom/path")
			},
			wantHas:        true,
			wantCustomPath: "/custom/path",
		},
		{
			name: "OMZ installed with ZSH env",
			setupEnv: func(tmpDir string) {
				omzDir := filepath.Join(tmpDir, ".oh-my-zsh")
				os.MkdirAll(omzDir, 0755)
				os.Setenv("ZSH", omzDir)
			},
			wantHas:        true,
			wantCustomPath: "",
		},
		{
			name: "OMZ not installed",
			setupEnv: func(tmpDir string) {
				// Nothing installed
			},
			wantHas:        false,
			wantCustomPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp home directory
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".zshrc")

			// Clear environment
			os.Unsetenv("ZSH")
			os.Unsetenv("ZSH_CUSTOM")

			// Setup environment
			tt.setupEnv(tmpDir)

			// Create ZshShell with custom home
			shell := &ZshShell{
				BaseShell: BaseShell{
					Type:    ShellTypeZsh,
					Name:    "zsh",
					RCFile:  rcPath,
					HomeDir: tmpDir,
				},
			}

			gotHas, gotCustomPath := shell.HasOhMyZsh()

			if gotHas != tt.wantHas {
				t.Errorf("HasOhMyZsh() gotHas = %v, want %v", gotHas, tt.wantHas)
			}
			if gotCustomPath != tt.wantCustomPath {
				t.Errorf("HasOhMyZsh() gotCustomPath = %v, want %v", gotCustomPath, tt.wantCustomPath)
			}
		})
	}
}

func TestGetOhMyZshPluginDir(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       func(tmpDir string)
		wantDirForHome func(tmpDir string) string
	}{
		{
			name: "default location",
			setupEnv: func(tmpDir string) {
				// No env vars set
			},
			wantDirForHome: func(tmpDir string) string {
				return filepath.Join(tmpDir, ".oh-my-zsh", "custom", "plugins")
			},
		},
		{
			name: "with ZSH_CUSTOM",
			setupEnv: func(tmpDir string) {
				os.Setenv("ZSH_CUSTOM", "/custom/zsh")
			},
			wantDirForHome: func(tmpDir string) string {
				return "/custom/zsh/plugins"
			},
		},
		{
			name: "with ZSH env",
			setupEnv: func(tmpDir string) {
				os.Setenv("ZSH", "/opt/oh-my-zsh")
			},
			wantDirForHome: func(tmpDir string) string {
				return "/opt/oh-my-zsh/custom/plugins"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".zshrc")

			// Clear environment
			os.Unsetenv("ZSH")
			os.Unsetenv("ZSH_CUSTOM")

			tt.setupEnv(tmpDir)

			shell := &ZshShell{
				BaseShell: BaseShell{
					Type:    ShellTypeZsh,
					Name:    "zsh",
					RCFile:  rcPath,
					HomeDir: tmpDir,
				},
			}

			gotDir := shell.GetOhMyZshPluginDir()
			wantDir := tt.wantDirForHome(tmpDir)

			if gotDir != wantDir {
				t.Errorf("GetOhMyZshPluginDir() = %v, want %v", gotDir, wantDir)
			}
		})
	}
}

func TestIsZshVersionCompatible(t *testing.T) {
	tests := []struct {
		currentVersion string
		minVersion     string
		wantCompatible bool
	}{
		{
			currentVersion: "5.8",
			minVersion:     "4.3.11",
			wantCompatible: true,
		},
		{
			currentVersion: "5.8.1",
			minVersion:     "5.8",
			wantCompatible: true,
		},
		{
			currentVersion: "4.3.10",
			minVersion:     "4.3.11",
			wantCompatible: false,
		},
		{
			currentVersion: "5.0",
			minVersion:     "4.3.11",
			wantCompatible: true,
		},
		{
			currentVersion: "4.3.11",
			minVersion:     "4.3.11",
			wantCompatible: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.currentVersion, func(t *testing.T) {
			result := compareVersions(tt.currentVersion, tt.minVersion)
			gotCompatible := result >= 0

			if gotCompatible != tt.wantCompatible {
				t.Errorf("compareVersions(%q, %q) = %d, want >= 0 for compatible",
					tt.currentVersion, tt.minVersion, result)
			}
		})
	}
}

func TestParsePluginsArray(t *testing.T) {
	tests := []struct {
		name        string
		rcContent   string
		wantPlugins []string
		wantErr     bool
	}{
		{
			name:        "single line plugins",
			rcContent:   `plugins=(git npm fzf)`,
			wantPlugins: []string{"git", "npm", "fzf"},
			wantErr:     false,
		},
		{
			name: "multi-line plugins",
			rcContent: `plugins=(
  git
  npm
  fzf
)`,
			wantPlugins: []string{"git", "npm", "fzf"},
			wantErr:     false,
		},
		{
			name:        "plugins with hyphen",
			rcContent:   `plugins=(git zsh-autosuggestions zsh-syntax-highlighting)`,
			wantPlugins: []string{"git", "zsh-autosuggestions", "zsh-syntax-highlighting"},
			wantErr:     false,
		},
		{
			name:        "no plugins array",
			rcContent:   `# some config\nexport PATH=$PATH:/usr/local/bin`,
			wantPlugins: []string{},
			wantErr:     false,
		},
		{
			name:        "empty plugins array",
			rcContent:   `plugins=()`,
			wantPlugins: []string{},
			wantErr:     false,
		},
		{
			name: "plugins with comments",
			rcContent: `# Oh My Zsh plugins
plugins=(git npm fzf)
# more config`,
			wantPlugins: []string{"git", "npm", "fzf"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".zshrc")

			if err := os.WriteFile(rcPath, []byte(tt.rcContent), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			shell := &ZshShell{
				BaseShell: BaseShell{
					Type:    ShellTypeZsh,
					Name:    "zsh",
					RCFile:  rcPath,
					HomeDir: tmpDir,
				},
			}

			gotPlugins, err := shell.ParsePluginsArray()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePluginsArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(gotPlugins) != len(tt.wantPlugins) {
				t.Errorf("ParsePluginsArray() got %d plugins, want %d", len(gotPlugins), len(tt.wantPlugins))
				return
			}

			for i, plugin := range gotPlugins {
				if plugin != tt.wantPlugins[i] {
					t.Errorf("ParsePluginsArray()[%d] = %v, want %v", i, plugin, tt.wantPlugins[i])
				}
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1     string
		v2     string
		result int
	}{
		{"5.8", "5.8", 0},
		{"5.8", "5.7", 1},
		{"5.8", "5.9", -1},
		{"5.8.1", "5.8", 1},
		{"4.3.11", "4.3.10", 1},
		{"4.3", "4.3.11", -1},
		{"5.8 (x86_64)", "5.8", 0},
	}

	for _, tt := range tests {
		t.Run(tt.v1+" vs "+tt.v2, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.result {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.result)
			}
		})
	}
}

func TestParseVersionParts(t *testing.T) {
	tests := []struct {
		version string
		want    []int
	}{
		{"5.8", []int{5, 8}},
		{"5.8.1", []int{5, 8, 1}},
		{"4.3.11", []int{4, 3, 11}},
		{"5.8 (x86_64)", []int{5, 8}},
		{"5.8.1 (aarch64-apple-darwin21.0)", []int{5, 8, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := parseVersionParts(tt.version)
			if len(got) != len(tt.want) {
				t.Errorf("parseVersionParts(%q) = %v, want %v", tt.version, got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("parseVersionParts(%q)[%d] = %d, want %d", tt.version, i, v, tt.want[i])
				}
			}
		})
	}
}

func TestRemovePluginFromList(t *testing.T) {
	tests := []struct {
		name    string
		plugins []string
		remove  string
		want    []string
	}{
		{
			name:    "remove from middle",
			plugins: []string{"git", "npm", "fzf"},
			remove:  "npm",
			want:    []string{"git", "fzf"},
		},
		{
			name:    "remove from start",
			plugins: []string{"git", "npm", "fzf"},
			remove:  "git",
			want:    []string{"npm", "fzf"},
		},
		{
			name:    "remove from end",
			plugins: []string{"git", "npm", "fzf"},
			remove:  "fzf",
			want:    []string{"git", "npm"},
		},
		{
			name:    "remove non-existent",
			plugins: []string{"git", "npm"},
			remove:  "fzf",
			want:    []string{"git", "npm"},
		},
		{
			name:    "empty list",
			plugins: []string{},
			remove:  "git",
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removePluginFromList(tt.plugins, tt.remove)

			if len(got) != len(tt.want) {
				t.Errorf("removePluginFromList() got %v, want %v", got, tt.want)
				return
			}

			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("removePluginFromList()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestParsePluginsFromArray(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantNames []string
	}{
		{
			name:      "single line",
			content:   "plugins=(git npm fzf)",
			wantNames: []string{"git", "npm", "fzf"},
		},
		{
			name: "multi-line",
			content: `plugins=(
    git
    npm
    fzf
)`,
			wantNames: []string{"git", "npm", "fzf"},
		},
		{
			name:      "empty",
			content:   "plugins=()",
			wantNames: []string{},
		},
		{
			name:      "no plugins",
			content:   "export PATH=$PATH:/usr/local/bin",
			wantNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePluginsFromArray(tt.content)

			if len(got) != len(tt.wantNames) {
				t.Errorf("parsePluginsFromArray() got %d plugins, want %d", len(got), len(tt.wantNames))
				return
			}

			for i, name := range got {
				if name != tt.wantNames[i] {
					t.Errorf("parsePluginsFromArray()[%d] = %v, want %v", i, name, tt.wantNames[i])
				}
			}
		})
	}
}
