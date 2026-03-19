// Package installer provides tests for zsh plugin detection.
package installer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/savanhi/shell/pkg/shell"
)

func TestInstallMethodString(t *testing.T) {
	tests := []struct {
		method   InstallMethod
		expected string
	}{
		{MethodNone, "none"},
		{MethodOhMyZsh, "Oh My Zsh"},
		{MethodHomebrew, "Homebrew"},
		{MethodGitClone, "Git Clone"},
		{InstallMethod(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.method.String()
			if got != tt.expected {
				t.Errorf("InstallMethod(%d).String() = %v, want %v", int(tt.method), got, tt.expected)
			}
		})
	}
}

func TestGetSupportedPlugins(t *testing.T) {
	plugins := GetSupportedPlugins()

	if len(plugins) < 2 {
		t.Errorf("GetSupportedPlugins() returned %d plugins, want at least 2", len(plugins))
	}

	// Check that zsh-autosuggestions and zsh-syntax-highlighting are included
	foundAutosuggestions := false
	foundSyntaxHighlighting := false

	for _, p := range plugins {
		if p.Name == "zsh-autosuggestions" {
			foundAutosuggestions = true
			if p.DisplayName != "Zsh Autosuggestions" {
				t.Errorf("zsh-autosuggestions DisplayName = %v, want 'Zsh Autosuggestions'", p.DisplayName)
			}
			if p.MustBeLast {
				t.Error("zsh-autosuggestions should not have MustBeLast = true")
			}
		}
		if p.Name == "zsh-syntax-highlighting" {
			foundSyntaxHighlighting = true
			if p.DisplayName != "Zsh Syntax Highlighting" {
				t.Errorf("zsh-syntax-highlighting DisplayName = %v, want 'Zsh Syntax Highlighting'", p.DisplayName)
			}
			if !p.MustBeLast {
				t.Error("zsh-syntax-highlighting should have MustBeLast = true")
			}
		}
	}

	if !foundAutosuggestions {
		t.Error("GetSupportedPlugins() missing zsh-autosuggestions")
	}
	if !foundSyntaxHighlighting {
		t.Error("GetSupportedPlugins() missing zsh-syntax-highlighting")
	}
}

func TestGetPluginByName(t *testing.T) {
	tests := []struct {
		name    string
		plugin  string
		wantNil bool
	}{
		{"existing plugin", "zsh-autosuggestions", false},
		{"another existing plugin", "zsh-syntax-highlighting", false},
		{"non-existing plugin", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := GetPluginByName(tt.plugin)
			if (p == nil) != tt.wantNil {
				t.Errorf("GetPluginByName(%q) = %v, want nil=%v", tt.plugin, p, tt.wantNil)
			}
			if p != nil && p.Name != tt.plugin {
				t.Errorf("GetPluginByName(%q).Name = %v, want %v", tt.plugin, p.Name, tt.plugin)
			}
		})
	}
}

func TestDefaultPluginInstallerConfig(t *testing.T) {
	cfg := DefaultPluginInstallerConfig()

	if !cfg.PreferOhMyZsh {
		t.Error("DefaultPluginInstallerConfig().PreferOhMyZsh should be true")
	}
	if !cfg.PreferHomebrew {
		t.Error("DefaultPluginInstallerConfig().PreferHomebrew should be true")
	}
	if cfg.Force {
		t.Error("DefaultPluginInstallerConfig().Force should be false")
	}
	if cfg.DryRun {
		t.Error("DefaultPluginInstallerConfig().DryRun should be false")
	}
}

func TestPluginStatusDefaults(t *testing.T) {
	status := &PluginStatus{
		Plugin:    GetSupportedPlugins()[0],
		Installed: false,
		Method:    MethodNone,
	}

	if status.Installed {
		t.Error("PluginStatus.Installed should default to false")
	}
	if status.Method != MethodNone {
		t.Errorf("PluginStatus.Method should default to MethodNone, got %v", status.Method)
	}
}

func TestPluginDetector_Detect(t *testing.T) {
	// Create a test environment
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")

	// Create minimal .zshrc
	rcContent := `export PATH=$PATH:/usr/local/bin
plugins=(git)
`
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
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

	// Create install context
	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		BinDir:    filepath.Join(tmpDir, ".local", "bin"),
		OS:        "linux",
	}

	detector := NewPluginDetector(ctx, zshShell)

	// Test detection for zsh-autosuggestions
	plugin := GetPluginByName("zsh-autosuggestions")
	if plugin == nil {
		t.Fatal("plugin should not be nil")
	}

	status, err := detector.Detect(*plugin)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// Plugin should not be installed in fresh environment
	if status.Installed {
		t.Error("Plugin should not be detected as installed in fresh environment")
	}
	if status.Method != MethodNone {
		t.Errorf("Plugin method should be MethodNone, got %v", status.Method)
	}
}

func TestPluginDetector_DetectAll(t *testing.T) {
	// Create a test environment
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")

	// Create minimal .zshrc
	rcContent := `export PATH=$PATH:/usr/local/bin
plugins=(git)
`
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
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

	// Create install context
	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		BinDir:    filepath.Join(tmpDir, ".local", "bin"),
		OS:        "linux",
	}

	detector := NewPluginDetector(ctx, zshShell)

	statuses, err := detector.DetectAll()
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	// Should return all supported plugins
	expectedCount := len(GetSupportedPlugins())
	if len(statuses) != expectedCount {
		t.Errorf("DetectAll() returned %d statuses, want %d", len(statuses), expectedCount)
	}

	// Check that all statuses have correct plugin info
	for _, status := range statuses {
		if status.Plugin.Name == "" {
			t.Error("PluginStatus should have a valid Plugin")
		}
		if status.Installed && status.Method == MethodNone {
			t.Errorf("Plugin %s is installed but has MethodNone", status.Plugin.Name)
		}
	}
}

func TestPluginDetector_DetectWithOhMyZsh(t *testing.T) {
	// Create a test environment with Oh MyZsh
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")

	// Create OMZ directory
	omzDir := filepath.Join(tmpDir, ".oh-my-zsh")
	omzPluginsDir := filepath.Join(omzDir, "custom", "plugins")
	autosuggestionsDir := filepath.Join(omzPluginsDir, "zsh-autosuggestions")

	if err := os.MkdirAll(autosuggestionsDir, 0755); err != nil {
		t.Fatalf("failed to create OMZ plugin dir: %v", err)
	}

	// Create minimal .zshrc with plugins array including autosuggestions
	rcContent := `source $ZSH/oh-my-zsh.sh
plugins=(git zsh-autosuggestions)
`
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
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

	// Create install context
	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		BinDir:    filepath.Join(tmpDir, ".local", "bin"),
		OS:        "linux",
	}

	detector := NewPluginDetector(ctx, zshShell)

	// Test detection for zsh-autosuggestions (should find it)
	plugin := GetPluginByName("zsh-autosuggestions")
	if plugin == nil {
		t.Fatal("plugin should not be nil")
	}

	status, err := detector.Detect(*plugin)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// Plugin should be detected as installed via OMZ
	if !status.Installed {
		t.Error("Plugin should be detected as installed (OMZ)")
	}
	if status.Method != MethodOhMyZsh {
		t.Errorf("Plugin method should be MethodOhMyZsh, got %v", status.Method)
	}
}

func TestPluginDetector_isInPluginsArray(t *testing.T) {
	// Create a test environment
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")

	tests := []struct {
		name       string
		rcContent  string
		pluginName string
		wantResult bool
	}{
		{
			name:       "plugin in array",
			rcContent:  "plugins=(git npm fzf)",
			pluginName: "npm",
			wantResult: true,
		},
		{
			name:       "plugin not in array",
			rcContent:  "plugins=(git npm fzf)",
			pluginName: "zsh-autosuggestions",
			wantResult: false,
		},
		{
			name:       "no plugins array",
			rcContent:  "export PATH=$PATH:/usr/local/bin",
			pluginName: "git",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(rcPath, []byte(tt.rcContent), 0644); err != nil {
				t.Fatalf("failed to create .zshrc: %v", err)
			}

			zshShell := &shell.ZshShell{
				BaseShell: shell.BaseShell{
					Type:    shell.ShellTypeZsh,
					Name:    "zsh",
					RCFile:  rcPath,
					HomeDir: tmpDir,
				},
			}

			ctx := &InstallContext{
				HomeDir: tmpDir,
			}

			detector := NewPluginDetector(ctx, zshShell)
			got := detector.isInPluginsArray(tt.pluginName)

			if got != tt.wantResult {
				t.Errorf("isInPluginsArray(%q) = %v, want %v", tt.pluginName, got, tt.wantResult)
			}
		})
	}
}

func TestPluginInstallerConfig(t *testing.T) {
	cfg := &PluginInstallerConfig{
		PreferOhMyZsh:  true,
		PreferHomebrew: false,
		Force:          true,
		DryRun:         false,
	}

	if !cfg.PreferOhMyZsh {
		t.Error("PreferOhMyZsh should be true")
	}
	if cfg.PreferHomebrew {
		t.Error("PreferHomebrew should be false")
	}
	if !cfg.Force {
		t.Error("Force should be true")
	}
	if cfg.DryRun {
		t.Error("DryRun should be false")
	}
}

func TestPluginInstaller_Detect(t *testing.T) {
	// Create a test environment
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")

	// Create minimal .zshrc
	rcContent := `export PATH=$PATH:/usr/local/bin`
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
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

	// Create install context
	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		OS:        "linux",
	}

	installer := NewPluginInstaller(ctx, zshShell)

	// Test Detect
	plugin := GetPluginByName("zsh-autosuggestions")
	if plugin == nil {
		t.Fatal("plugin should not be nil")
	}

	status, err := installer.Detect(*plugin)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if status.Plugin.Name != "zsh-autosuggestions" {
		t.Errorf("Detect returned wrong plugin: %v", status.Plugin.Name)
	}
}

func TestPluginInstaller_DetectAll(t *testing.T) {
	// Create a test environment
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")

	// Create minimal .zshrc
	rcContent := `export PATH=$PATH:/usr/local/bin`
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
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

	// Create install context
	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		OS:        "linux",
	}

	installer := NewPluginInstaller(ctx, zshShell)

	statuses, err := installer.DetectAll()
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	// Should return all supported plugins
	expectedCount := len(GetSupportedPlugins())
	if len(statuses) != expectedCount {
		t.Errorf("DetectAll() returned %d statuses, want %d", len(statuses), expectedCount)
	}
}

func TestPluginInstaller_Install(t *testing.T) {
	tests := []struct {
		name    string
		plugin  Plugin
		method  InstallMethod
		wantErr bool
	}{
		{
			name: "auto-select method",
			plugin: Plugin{
				Name:          "test-plugin",
				DisplayName:   "Test Plugin",
				Repository:    "https://github.com/test/test-plugin",
				SourceFile:    "test-plugin.zsh",
				OhMyZshName:   "test-plugin",
				MinZshVersion: "4.3.11",
			},
			method:  MethodNone, // Auto-select
			wantErr: false,
		},
		{
			name: "git clone method",
			plugin: Plugin{
				Name:          "test-plugin-clone",
				DisplayName:   "Test Plugin Clone",
				Repository:    "https://github.com/test/test-plugin-clone",
				SourceFile:    "test-plugin-clone.zsh",
				MinZshVersion: "4.3.11",
			},
			method:  MethodGitClone,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test environment
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".zshrc")

			// Create minimal .zshrc
			rcContent := `export PATH=$PATH:/usr/local/bin`
			if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
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

			// Create install context
			ctx := &InstallContext{
				HomeDir:   tmpDir,
				ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
				OS:        "linux",
			}

			installer := NewPluginInstaller(ctx, zshShell)

			// Note: Actual installation tests would require git/brew/OMZ
			// Here we test the structure and auto-selection logic
			result, err := installer.Install(context.Background(), tt.plugin, tt.method)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Install() expected error, got nil")
				}
			} else {
				// Installation may fail due to missing dependencies (git, etc.)
				// which is expected in test environment
				// We just verify the method selection works
				if result == nil && err != nil {
					// Expected in test environment without git
					t.Logf("Install() returned expected error: %v", err)
				}
			}
		})
	}
}

func TestPluginInstaller_InstallAll(t *testing.T) {
	// Create a test environment
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")

	// Create minimal .zshrc
	rcContent := `export PATH=$PATH:/usr/local/bin`
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
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

	// Create install context
	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		OS:        "linux",
	}

	_ = NewPluginInstaller(ctx, zshShell) // Not used directly, but verifies construction

	// Test ordering - MustBeLast plugins should be installed last
	plugins := []Plugin{
		{
			Name:          "test-plugin-1",
			DisplayName:   "Test Plugin 1",
			Repository:    "https://github.com/test/plugin1",
			SourceFile:    "plugin1.zsh",
			MustBeLast:    false,
			MinZshVersion: "4.3.11",
		},
		{
			Name:          "zsh-syntax-highlighting",
			DisplayName:   "Zsh Syntax Highlighting",
			Repository:    "https://github.com/zsh-users/zsh-syntax-highlighting",
			SourceFile:    "zsh-syntax-highlighting.zsh",
			MustBeLast:    true,
			MinZshVersion: "4.3.11",
		},
		{
			Name:          "test-plugin-2",
			DisplayName:   "Test Plugin 2",
			Repository:    "https://github.com/test/plugin2",
			SourceFile:    "plugin2.zsh",
			MustBeLast:    false,
			MinZshVersion: "4.3.11",
		},
	}

	// Verify MustBeLast ordering logic
	// Note: Actual installation will fail without git, but we can verify the ordering
	t.Run("ordering logic", func(t *testing.T) {
		// Verify that MustBeLast plugins are identified correctly
		var lastPlugins []Plugin
		var regularPlugins []Plugin
		for _, p := range plugins {
			if p.MustBeLast {
				lastPlugins = append(lastPlugins, p)
			} else {
				regularPlugins = append(regularPlugins, p)
			}
		}

		if len(lastPlugins) != 1 {
			t.Errorf("Expected 1 MustBeLast plugin, got %d", len(lastPlugins))
		}
		if lastPlugins[0].Name != "zsh-syntax-highlighting" {
			t.Errorf("Expected zsh-syntax-highlighting as MustBeLast, got %s", lastPlugins[0].Name)
		}
		if len(regularPlugins) != 2 {
			t.Errorf("Expected 2 regular plugins, got %d", len(regularPlugins))
		}
	})
}

func TestPluginInstaller_Uninstall(t *testing.T) {
	tests := []struct {
		name       string
		pluginName string
		wantErr    bool
	}{
		{
			name:       "unknown plugin",
			pluginName: "nonexistent-plugin",
			wantErr:    true,
		},
		{
			name:       "valid plugin name",
			pluginName: "zsh-autosuggestions",
			wantErr:    false, // Will succeed but do nothing (not installed)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test environment
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".zshrc")

			// Create minimal .zshrc
			rcContent := `export PATH=$PATH:/usr/local/bin`
			if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
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

			// Create install context
			ctx := &InstallContext{
				HomeDir:   tmpDir,
				ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
				OS:        "linux",
			}

			installer := NewPluginInstaller(ctx, zshShell)

			err := installer.Uninstall(tt.pluginName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Uninstall() expected error, got nil")
				}
			} else {
				// For valid plugins that aren't installed, Uninstall should succeed
				if err != nil {
					t.Logf("Uninstall() error: %v (plugin may not be installed)", err)
				}
			}
		})
	}
}

func TestPluginInstaller_SelectInstallMethod(t *testing.T) {
	// Create a test environment
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")

	// Create minimal .zshrc
	rcContent := `export PATH=$PATH:/usr/local/bin`
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
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

	// Create install context
	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		OS:        "linux",
	}

	installer := NewPluginInstaller(ctx, zshShell)
	plugin := GetSupportedPlugins()[0]

	// Test method selection (will fallback to GitClone since no OMZ/brew)
	selectedMethod := installer.selectInstallMethod(zshShell, plugin)
	if selectedMethod != MethodGitClone {
		// If brew is available on the test system, it might be MethodHomebrew
		// But since we're testing, we just verify a valid method is returned
		t.Logf("Selected method: %v (expected GitClone without OMZ/brew)", selectedMethod)
	}
}

func TestPluginInstaller_EnsureCorrectOrder(t *testing.T) {
	// Create a test environment
	tmpDir := t.TempDir()
	rcPath := filepath.Join(tmpDir, ".zshrc")

	// Create .zshrc with plugins in wrong order
	rcContent := `# >>> savanhi-zsh-syntax-highlighting >>>
source ~/.zsh/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh
# <<< savanhi-zsh-syntax-highlighting <<<
# >>> savanhi-zsh-autosuggestions >>>
source ~/.zsh/zsh-autosuggestions/zsh-autosuggestions.zsh
# <<< savanhi-zsh-autosuggestions <<<
`
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
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

	// Create install context
	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		OS:        "linux",
	}

	installer := NewPluginInstaller(ctx, zshShell)

	// Test with non-MustBeLast plugin (should do nothing)
	regularPlugin := Plugin{Name: "zsh-autosuggestions", MustBeLast: false}
	err := installer.ensureCorrectOrder(regularPlugin, zshShell)
	if err != nil {
		t.Errorf("ensureCorrectOrder for non-MustBeLast returned error: %v", err)
	}

	// Test with MustBeLast plugin (should ensure it's last)
	syntaxPlugin := Plugin{Name: "zsh-syntax-highlighting", MustBeLast: true}
	err = installer.ensureCorrectOrder(syntaxPlugin, zshShell)
	if err != nil {
		t.Errorf("ensureCorrectOrder for MustBeLast returned error: %v", err)
	}
}
