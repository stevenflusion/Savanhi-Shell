// Package installer provides tests for RC modification zsh plugin functionality.
package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/savanhi/shell/pkg/shell"
)

func TestRCModifier_InjectZshPlugin(t *testing.T) {
	tests := []struct {
		name       string
		plugin     Plugin
		sourcePath string
		wantErr    bool
	}{
		{
			name: "inject zsh-autosuggestions",
			plugin: Plugin{
				Name:        "zsh-autosuggestions",
				DisplayName: "Zsh Autosuggestions",
			},
			sourcePath: "/usr/local/share/zsh-autosuggestions/zsh-autosuggestions.zsh",
			wantErr:    false,
		},
		{
			name: "inject zsh-syntax-highlighting",
			plugin: Plugin{
				Name:        "zsh-syntax-highlighting",
				DisplayName: "Zsh Syntax Highlighting",
			},
			sourcePath: "/usr/local/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".zshrc")

			// Create initial .zshrc
			if err := os.WriteFile(rcPath, []byte("# Initial content\n"), 0644); err != nil {
				t.Fatalf("failed to create .zshrc: %v", err)
			}

			// Create ZshShell
			zshShell := &shell.ZshShell{
				BaseShell: shell.BaseShell{
					Type:    shell.ShellTypeZsh,
					Name:    "zsh",
					RCFile:  rcPath,
					HomeDir: tmpDir,
				},
			}

			modifier := NewRCModifier(zshShell, tmpDir)

			err := modifier.InjectZshPlugin(tt.plugin, tt.sourcePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("InjectZshPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the section was injected
				hasSection, err := modifier.HasZshPluginSection(tt.plugin.Name)
				if err != nil {
					t.Errorf("HasZshPluginSection() error = %v", err)
					return
				}
				if !hasSection {
					t.Error("Plugin section should exist after injection")
				}

				// Verify the content
				content, err := modifier.GetZshPluginSection(tt.plugin.Name)
				if err != nil {
					t.Errorf("GetZshPluginSection() error = %v", err)
					return
				}

				if content == "" {
					t.Error("Section content should not be empty")
				}
			}
		})
	}
}

func TestRCModifier_RemoveZshPlugin(t *testing.T) {
	tests := []struct {
		name       string
		plugin     Plugin
		sourcePath string
	}{
		{
			name: "remove zsh-autosuggestions",
			plugin: Plugin{
				Name:        "zsh-autosuggestions",
				DisplayName: "Zsh Autosuggestions",
			},
			sourcePath: "/usr/local/share/zsh-autosuggestions.zsh",
		},
		{
			name: "remove zsh-syntax-highlighting",
			plugin: Plugin{
				Name:        "zsh-syntax-highlighting",
				DisplayName: "Zsh Syntax Highlighting",
			},
			sourcePath: "/usr/local/share/zsh-syntax-highlighting.zsh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".zshrc")

			// Create ZshShell
			zshShell := &shell.ZshShell{
				BaseShell: shell.BaseShell{
					Type:    shell.ShellTypeZsh,
					Name:    "zsh",
					RCFile:  rcPath,
					HomeDir: tmpDir,
				},
			}

			modifier := NewRCModifier(zshShell, tmpDir)

			// First inject
			if err := modifier.InjectZshPlugin(tt.plugin, tt.sourcePath); err != nil {
				t.Fatalf("failed to inject plugin: %v", err)
			}

			// Verify it exists
			hasSection, _ := modifier.HasZshPluginSection(tt.plugin.Name)
			if !hasSection {
				t.Fatal("Plugin section should exist after injection")
			}

			// Now remove
			if err := modifier.RemoveZshPlugin(tt.plugin.Name); err != nil {
				t.Errorf("RemoveZshPlugin() error = %v", err)
				return
			}

			// Verify it's gone
			hasSection, err := modifier.HasZshPluginSection(tt.plugin.Name)
			if err != nil {
				t.Errorf("HasZshPluginSection() error = %v", err)
				return
			}
			if hasSection {
				t.Error("Plugin section should not exist after removal")
			}
		})
	}
}

func TestRCModifier_EnsurePluginOrder(t *testing.T) {
	tests := []struct {
		name           string
		setupContent   string
		plugins        []Plugin
		wantErr        bool
		wantEndContent string
	}{
		{
			name: "syntax-highlighting already last",
			setupContent: `# >>> savanhi-zsh-autosuggestions >>>
source /path/to/zsh-autosuggestions.zsh
# <<< savanhi-zsh-autosuggestions <<<
# >>> savanhi-zsh-syntax-highlighting >>>
source /path/to/zsh-syntax-highlighting.zsh
# <<< savanhi-zsh-syntax-highlighting <<<
`,
			plugins: []Plugin{
				{Name: "zsh-autosuggestions"},
				{Name: "zsh-syntax-highlighting", MustBeLast: true},
			},
			wantErr: false,
		},
		{
			name: "syntax-highlighting not last - needs reorder",
			setupContent: `# >>> savanhi-zsh-syntax-highlighting >>>
source /path/to/zsh-syntax-highlighting.zsh
# <<< savanhi-zsh-syntax-highlighting <<<
# >>> savanhi-zsh-autosuggestions >>>
source /path/to/zsh-autosuggestions.zsh
# <<< savanhi-zsh-autosuggestions <<<
`,
			plugins: []Plugin{
				{Name: "zsh-autosuggestions"},
				{Name: "zsh-syntax-highlighting", MustBeLast: true},
			},
			wantErr: false,
		},
		{
			name: "no syntax-highlighting plugin",
			setupContent: `# >>> savanhi-zsh-autosuggestions >>>
source /path/to/zsh-autosuggestions.zsh
# <<< savanhi-zsh-autosuggestions <<<
`,
			plugins: []Plugin{
				{Name: "zsh-autosuggestions"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".zshrc")

			// Create .zshrc with setup content
			if err := os.WriteFile(rcPath, []byte(tt.setupContent), 0644); err != nil {
				t.Fatalf("failed to create .zshrc: %v", err)
			}

			// Create ZshShell
			zshShell := &shell.ZshShell{
				BaseShell: shell.BaseShell{
					Type:    shell.ShellTypeZsh,
					Name:    "zsh",
					RCFile:  rcPath,
					HomeDir: tmpDir,
				},
			}

			modifier := NewRCModifier(zshShell, tmpDir)

			err := modifier.EnsurePluginOrder(tt.plugins)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsurePluginOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify syntax-highlighting is at the end
				content, err := os.ReadFile(rcPath)
				if err != nil {
					t.Fatalf("failed to read .zshrc: %v", err)
				}

				// Find the last occurrence of section markers
				syntaxMarker := "# >>> savanhi-zsh-syntax-highlighting >>>"
				lastSectionIdx := -1
				syntaxIdx := -1

				for i := 0; i < len(content); i++ {
					// Check for any section start marker
					if i+10 < len(content) && string(content[i:i+10]) == "# >>> sav" {
						// Find the section name
						for j := i; j < len(content); j++ {
							if content[j] == '>' && j+2 < len(content) && content[j+1] == '>' && content[j+2] == '\n' {
								lastSectionIdx = j + 3
								break
							}
						}
					}
				}

				// If syntax-highlighting section exists, verify it's last
				syntaxIdx = findLastSectionIndex(content, syntaxMarker)

				if syntaxIdx != -1 && lastSectionIdx != -1 {
					// Check if there are other sections after syntax-highlighting
					autosuggestionsMarker := "# >>> savanhi-zsh-autosuggestions >>>"
					autosuggestionsIdx := findLastSectionIndex(content, autosuggestionsMarker)

					// If both exist, syntax-highlighting should be after autosuggestions
					if autosuggestionsIdx != -1 && syntaxIdx < autosuggestionsIdx {
						t.Error("zsh-syntax-highlighting should be after zsh-autosuggestions in .zshrc")
					}
				}
			}
		})
	}
}

func TestRCModifier_AddPluginToSection(t *testing.T) {
	tests := []struct {
		name            string
		existingContent string
		plugin          Plugin
		sourcePath      string
		wantErr         bool
	}{
		{
			name:            "add to new section",
			existingContent: "# Initial content\n",
			plugin: Plugin{
				Name:        "zsh-autosuggestions",
				DisplayName: "Zsh Autosuggestions",
			},
			sourcePath: "/path/to/zsh-autosuggestions.zsh",
			wantErr:    false,
		},
		{
			name: "add to existing section",
			existingContent: `# Initial content
# >>> savanhi-zsh-autosuggestions >>>
source /existing/path.zsh
# <<< savanhi-zsh-autosuggestions <<<
`,
			plugin: Plugin{
				Name:        "zsh-autosuggestions",
				DisplayName: "Zsh Autosuggestions",
			},
			sourcePath: "/new/path.zsh",
			wantErr:    false,
		},
		{
			name: "add duplicate - should be idempotent",
			existingContent: `# Initial content
# >>> savanhi-zsh-autosuggestions >>>
source /path/to/zsh-autosuggestions.zsh
# <<< savanhi-zsh-autosuggestions <<<
`,
			plugin: Plugin{
				Name:        "zsh-autosuggestions",
				DisplayName: "Zsh Autosuggestions",
			},
			sourcePath: "/path/to/zsh-autosuggestions.zsh",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".zshrc")

			// Create .zshrc with existing content
			if err := os.WriteFile(rcPath, []byte(tt.existingContent), 0644); err != nil {
				t.Fatalf("failed to create .zshrc: %v", err)
			}

			// Create ZshShell
			zshShell := &shell.ZshShell{
				BaseShell: shell.BaseShell{
					Type:    shell.ShellTypeZsh,
					Name:    "zsh",
					RCFile:  rcPath,
					HomeDir: tmpDir,
				},
			}

			modifier := NewRCModifier(zshShell, tmpDir)

			err := modifier.AddPluginToSection(tt.plugin, tt.sourcePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddPluginToSection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the source path is in the content
				content, err := os.ReadFile(rcPath)
				if err != nil {
					t.Fatalf("failed to read .zshrc: %v", err)
				}

				if !containsString(string(content), tt.sourcePath) {
					t.Errorf("Source path %s should be in .zshrc", tt.sourcePath)
				}
			}
		})
	}
}

func TestRCModifier_RemoveAllPluginSections(t *testing.T) {
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")

	// Create .zshrc with both plugins
	content := `# Initial content
# >>> savanhi-zsh-autosuggestions >>>
source /path/to/zsh-autosuggestions.zsh
# <<< savanhi-zsh-autosuggestions <<<
# >>> savanhi-zsh-syntax-highlighting >>>
source /path/to/zsh-syntax-highlighting.zsh
# <<< savanhi-zsh-syntax-highlighting <<<
# More content
`
	if err := os.WriteFile(rcPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .zshrc: %v", err)
	}

	// Create ZshShell
	zshShell := &shell.ZshShell{
		BaseShell: shell.BaseShell{
			Type:    shell.ShellTypeZsh,
			Name:    "zsh",
			RCFile:  rcPath,
			HomeDir: tmpDir,
		},
	}

	modifier := NewRCModifier(zshShell, tmpDir)

	// Remove all plugin sections
	if err := modifier.RemoveAllPluginSections(); err != nil {
		t.Errorf("RemoveAllPluginSections() error = %v", err)
		return
	}

	// Verify both sections are gone
	hasAuto, _ := modifier.HasZshPluginSection("zsh-autosuggestions")
	hasSyntax, _ := modifier.HasZshPluginSection("zsh-syntax-highlighting")

	if hasAuto {
		t.Error("zsh-autosuggestions section should be removed")
	}
	if hasSyntax {
		t.Error("zsh-syntax-highlighting section should be removed")
	}

	// Verify initial content is preserved
	rcContent, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("failed to read .zshrc: %v", err)
	}

	if !containsString(string(rcContent), "# Initial content") {
		t.Error("Initial content should be preserved")
	}
	if !containsString(string(rcContent), "# More content") {
		t.Error("Trailing content should be preserved")
	}
}

func TestRCModifier_GetPluginSectionMarker(t *testing.T) {
	modifier := &RCModifier{}

	tests := []struct {
		name       string
		pluginName string
		wantMarker string
	}{
		{
			name:       "zsh-autosuggestions",
			pluginName: "zsh-autosuggestions",
			wantMarker: SectionZshAutosuggestions,
		},
		{
			name:       "zsh-syntax-highlighting",
			pluginName: "zsh-syntax-highlighting",
			wantMarker: SectionZshSyntaxHighlighting,
		},
		{
			name:       "custom plugin",
			pluginName: "zsh-custom",
			wantMarker: "savanhi-custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := modifier.getPluginSectionMarker(tt.pluginName)
			if got != tt.wantMarker {
				t.Errorf("getPluginSectionMarker(%q) = %q, want %q", tt.pluginName, got, tt.wantMarker)
			}
		})
	}
}

// Helper functions

func findLastSectionIndex(content []byte, marker string) int {
	idx := -1
	startIdx := 0

	for {
		found := findMarker(content[startIdx:], marker)
		if found == -1 {
			break
		}
		idx = startIdx + found
		startIdx = idx + 1
	}

	return idx
}

func findMarker(content []byte, marker string) int {
	for i := 0; i < len(content)-len(marker)+1; i++ {
		if string(content[i:i+len(marker)]) == marker {
			return i
		}
	}
	return -1
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
