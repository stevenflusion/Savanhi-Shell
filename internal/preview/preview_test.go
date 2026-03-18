// Package preview provides live preview capabilities for Savanhi Shell.
// This file contains tests for the preview engine types and interfaces.
package preview

import (
	"testing"
	"time"

	"github.com/savanhi/shell/pkg/shell"
)

// Test types and constants
func TestTypes(t *testing.T) {
	t.Run("PreviewType", func(t *testing.T) {
		tests := []struct {
			name string
			pt   PreviewType
		}{
			{"theme", PreviewTypeTheme},
			{"font", PreviewTypeFont},
			{"color_scheme", PreviewTypeColorScheme},
			{"full", PreviewTypeFull},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.pt == "" {
					t.Errorf("PreviewType %s should not be empty", tt.name)
				}
			})
		}
	})

	t.Run("PreviewStatus", func(t *testing.T) {
		statuses := []PreviewStatus{
			StatusPending,
			StatusRunning,
			StatusCompleted,
			StatusFailed,
			StatusCancelled,
			StatusTimeout,
		}

		for _, status := range statuses {
			if status == "" {
				t.Error("PreviewStatus should not be empty")
			}
		}
	})
}

// Test PreviewConfig
func TestPreviewConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		config := &PreviewConfig{
			Shell: shell.ShellTypeZsh,
		}

		if config.Shell != shell.ShellTypeZsh {
			t.Errorf("expected ShellTypeZsh, got %s", config.Shell)
		}

		if config.Timeout == 0 {
			// Default timeout should be applied by the spawner
			t.Log("timeout is 0, will use default")
		}
	})

	t.Run("with environment", func(t *testing.T) {
		config := &PreviewConfig{
			Shell: shell.ShellTypeBash,
			Environment: map[string]string{
				"TEST_VAR": "test_value",
			},
		}

		if config.Environment["TEST_VAR"] != "test_value" {
			t.Error("environment variable not set correctly")
		}
	})
}

// Test PreviewResult
func TestPreviewResult(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := &PreviewResult{
			ID:        "test-123",
			Status:    StatusCompleted,
			Output:    "test output",
			ExitCode:  0,
			Duration:  time.Second,
			StartTime: time.Now(),
		}

		if result.Status != StatusCompleted {
			t.Errorf("expected StatusCompleted, got %s", result.Status)
		}

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
	})

	t.Run("failed result", func(t *testing.T) {
		result := &PreviewResult{
			ID:           "test-456",
			Status:       StatusFailed,
			ExitCode:     1,
			ErrorMessage: "preview failed",
		}

		if result.Status != StatusFailed {
			t.Errorf("expected StatusFailed, got %s", result.Status)
		}

		if result.ErrorMessage != "preview failed" {
			t.Errorf("expected error message 'preview failed', got %s", result.ErrorMessage)
		}
	})
}

// Test SubshellConfig
func TestSubshellConfig(t *testing.T) {
	t.Run("basic config", func(t *testing.T) {
		config := &SubshellConfig{
			ShellType:     shell.ShellTypeBash,
			CaptureStdout: true,
			CaptureStderr: true,
		}

		if config.ShellType != shell.ShellTypeBash {
			t.Errorf("expected ShellTypeBash, got %s", config.ShellType)
		}

		if !config.CaptureStdout {
			t.Error("expected CaptureStdout to be true")
		}

		if !config.CaptureStderr {
			t.Error("expected CaptureStderr to be true")
		}
	})

	t.Run("with environment", func(t *testing.T) {
		config := &SubshellConfig{
			ShellType: shell.ShellTypeZsh,
			Environment: map[string]string{
				"PATH": "/usr/bin",
				"HOME": "/home/test",
				"TERM": "xterm-256color",
			},
		}

		if len(config.Environment) != 3 {
			t.Errorf("expected 3 environment variables, got %d", len(config.Environment))
		}
	})
}

// Test EnvironmentInjection
func TestEnvironmentInjection(t *testing.T) {
	t.Run("inject theme environment", func(t *testing.T) {
		injector := NewDefaultEnvironmentInjector()
		env := injector.InjectThemeEnv("/path/to/theme.omp.json")

		if env[EnvOhMyPoshTheme] != "/path/to/theme.omp.json" {
			t.Errorf("expected POSH_THEME to be set, got %s", env[EnvOhMyPoshTheme])
		}
	})

	t.Run("inject font environment", func(t *testing.T) {
		injector := NewDefaultEnvironmentInjector()
		env := injector.InjectFontEnv("MesloLGM Nerd Font", 14)

		if env[EnvFontFamily] != "MesloLGM Nerd Font" {
			t.Errorf("expected FONT_FAMILY to be set, got %s", env[EnvFontFamily])
		}

		if env[EnvFontSize] != "14" {
			t.Errorf("expected FONT_SIZE to be 14, got %s", env[EnvFontSize])
		}
	})

	t.Run("inject color scheme environment", func(t *testing.T) {
		injector := NewDefaultEnvironmentInjector()
		env := injector.InjectColorSchemeEnv("dracula")

		if env[EnvColorScheme] != "dracula" {
			t.Errorf("expected COLOR_SCHEME to be set, got %s", env[EnvColorScheme])
		}
	})

	t.Run("full environment injection", func(t *testing.T) {
		injector := NewDefaultEnvironmentInjector()
		config := &PreviewConfig{
			Shell:       shell.ShellTypeZsh,
			ThemePath:   "/path/to/theme.json",
			FontFamily:  "JetBrainsMono Nerd Font",
			ColorScheme: "nord",
			Environment: map[string]string{
				"CUSTOM_VAR": "custom_value",
			},
		}

		env, err := injector.InjectEnvironment(config)
		if err != nil {
			t.Fatalf("failed to inject environment: %v", err)
		}

		// Check theme
		if env[EnvOhMyPoshTheme] != "/path/to/theme.json" {
			t.Errorf("expected POSH_THEME to be set")
		}

		// Check font
		if env[EnvFontFamily] != "JetBrainsMono Nerd Font" {
			t.Errorf("expected FONT_FAMILY to be set")
		}

		// Check color scheme
		if env[EnvColorScheme] != "nord" {
			t.Errorf("expected COLOR_SCHEME to be set")
		}

		// Check custom var
		if env["CUSTOM_VAR"] != "custom_value" {
			t.Errorf("expected CUSTOM_VAR to be set")
		}
	})
}

// Test SubshellSpawner
func TestDefaultSubsheller(t *testing.T) {
	t.Run("create subsheller", func(t *testing.T) {
		subsheller, err := NewDefaultSubsheller()
		if err != nil {
			t.Fatalf("failed to create subsheller: %v", err)
		}

		if subsheller == nil {
			t.Fatal("subsheller should not be nil")
		}
	})

	t.Run("validate config - valid", func(t *testing.T) {
		subsheller, _ := NewDefaultSubsheller()
		config := &SubshellConfig{
			ShellType: shell.ShellTypeBash,
		}

		err := subsheller.validateConfig(config)
		if err != nil {
			t.Errorf("unexpected error for valid config: %v", err)
		}
	})

	t.Run("validate config - missing shell type", func(t *testing.T) {
		subsheller, _ := NewDefaultSubsheller()
		config := &SubshellConfig{}

		err := subsheller.validateConfig(config)
		if err == nil {
			t.Error("expected error for missing shell type")
		}
	})

	t.Run("validate config - unsupported shell", func(t *testing.T) {
		subsheller, _ := NewDefaultSubsheller()
		config := &SubshellConfig{
			ShellType: shell.ShellType("unsupported"),
		}

		err := subsheller.validateConfig(config)
		if err == nil {
			t.Error("expected error for unsupported shell type")
		}
	})
}

// Test Helper Functions
func TestHelperFunctions(t *testing.T) {
	t.Run("IsShellAvailable", func(t *testing.T) {
		// Bash should be available on most systems
		if !IsShellAvailable(shell.ShellTypeBash) {
			t.Log("bash not available (this may be expected in some environments)")
		}

		// Zsh may or may not be available
		t.Logf("zsh available: %v", IsShellAvailable(shell.ShellTypeZsh))
	})

	t.Run("GetDefaultShell", func(t *testing.T) {
		defaultShell := GetDefaultShell()
		if defaultShell == "" {
			t.Error("expected non-empty default shell")
		}
		t.Logf("default shell: %s", defaultShell)
	})

	t.Run("GetSubshellPath", func(t *testing.T) {
		path, err := GetSubshellPath(shell.ShellTypeBash)
		if err != nil {
			t.Logf("bash path not found (may be expected): %v", err)
		} else {
			t.Logf("bash path: %s", path)
		}
	})
}

// Test Font Preview Handler
func TestFontPreviewHandler(t *testing.T) {
	t.Run("GetNerdFontIcons", func(t *testing.T) {
		icons := GetNerdFontIcons()
		if len(icons) == 0 {
			t.Error("expected non-empty Nerd Font icons list")
		}
	})

	t.Run("IsNerdFont", func(t *testing.T) {
		tests := []struct {
			name     string
			fontName string
			expected bool
		}{
			{"nerd font explicit", "MesloLGM Nerd Font", true},
			{"nerd font suffix", "JetBrainsMono NF", true},
			{"nerd font words", "Fira Code Nerd Font", true},
			{"regular font", "Arial", false},
			{"system font", "Monaco", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsNerdFont(tt.fontName)
				if result != tt.expected {
					t.Errorf("IsNerdFont(%s) = %v, expected %v", tt.fontName, result, tt.expected)
				}
			})
		}
	})

	t.Run("GetRecommendedFonts", func(t *testing.T) {
		fonts := GetRecommendedFonts()
		if len(fonts) == 0 {
			t.Error("expected non-empty recommended fonts list")
		}

		// Check that MesloLGM is in the list (it's the most recommended)
		found := false
		for _, font := range fonts {
			if font == "MesloLGM Nerd Font" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected MesloLGM Nerd Font in recommended fonts")
		}
	})
}

// Test Color Scheme Handler
func TestColorSchemeHandler(t *testing.T) {
	t.Run("GetDefaultColorSchemes", func(t *testing.T) {
		schemes := GetDefaultColorSchemes()
		if len(schemes) == 0 {
			t.Error("expected non-empty color schemes list")
		}

		// Check for expected schemes
		expectedSchemes := []string{"dracula", "nord", "gruvbox"}
		for _, expected := range expectedSchemes {
			found := false
			for _, scheme := range schemes {
				if scheme.Name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected %s in default color schemes", expected)
			}
		}
	})

	t.Run("ValidateColorScheme", func(t *testing.T) {
		tests := []struct {
			name    string
			scheme  *ColorScheme
			wantErr bool
		}{
			{
				name: "valid scheme",
				scheme: &ColorScheme{
					Name: "test",
					Colors: ColorDefinitions{
						Foreground: "#FFFFFF",
						Background: "#000000",
					},
				},
				wantErr: false,
			},
			{
				name: "missing name",
				scheme: &ColorScheme{
					Name: "",
					Colors: ColorDefinitions{
						Foreground: "#FFFFFF",
						Background: "#000000",
					},
				},
				wantErr: true,
			},
			{
				name: "missing foreground",
				scheme: &ColorScheme{
					Name: "test",
					Colors: ColorDefinitions{
						Background: "#000000",
					},
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateColorScheme(tt.scheme)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateColorScheme() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

// Test Preview Safety Checker
func TestPreviewSafetyChecker(t *testing.T) {
	t.Run("validate config - timeout", func(t *testing.T) {
		checker := NewPreviewSafetyChecker()

		tests := []struct {
			name    string
			timeout time.Duration
			wantErr bool
		}{
			{"valid timeout", 5 * time.Second, false},
			{"minimum timeout", MinPreviewTimeout, false},
			{"below minimum", MinPreviewTimeout - 1, true},
			{"maximum timeout", MaxPreviewTimeout, false},
			{"above maximum", MaxPreviewTimeout + 1, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &PreviewConfig{
					Shell:   shell.ShellTypeBash,
					Timeout: tt.timeout,
				}

				err := checker.ValidateConfig(config)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("validate config - shell type", func(t *testing.T) {
		checker := NewPreviewSafetyChecker()

		tests := []struct {
			name      string
			shellType shell.ShellType
			wantErr   bool
		}{
			{"bash", shell.ShellTypeBash, false},
			{"zsh", shell.ShellTypeZsh, false},
			{"fish", shell.ShellTypeFish, false},
			{"empty", shell.ShellType(""), true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &PreviewConfig{
					Shell: tt.shellType,
				}

				err := checker.ValidateConfig(config)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

// Test Cleanup
func TestCleanup(t *testing.T) {
	t.Run("create cleaner", func(t *testing.T) {
		cleaner := NewPreviewCleaner()
		if cleaner == nil {
			t.Fatal("expected non-nil cleaner")
		}
	})

	t.Run("track temp files", func(t *testing.T) {
		cleaner := NewPreviewCleaner()

		cleaner.TrackTempFile("/tmp/test-file")
		cleaner.TrackTempDir("/tmp/test-dir")
		cleaner.TrackProcess(12345)

		// Verify tracking (internal state)
		// These should not panic
	})

	t.Run("cleanup all", func(t *testing.T) {
		cleaner := NewPreviewCleaner()

		result := cleaner.CleanupAll()
		if result == nil {
			t.Fatal("expected non-nil cleanup result")
		}

		// Empty cleanup should succeed
		if !result.Success {
			t.Log("cleanup result had errors (acceptable for empty cleaner)")
		}
	})
}

// Test Environment Builder Functions
func TestEnvironmentBuilderFunctions(t *testing.T) {
	t.Run("BuildEnvSlice", func(t *testing.T) {
		env := map[string]string{
			"PATH": "/usr/bin",
			"HOME": "/home/test",
			"TERM": "xterm",
		}

		slice := BuildEnvSlice(env)

		if len(slice) != 3 {
			t.Errorf("expected 3 env vars, got %d", len(slice))
		}
	})

	t.Run("MergeEnvironments", func(t *testing.T) {
		env1 := map[string]string{"A": "1", "B": "2"}
		env2 := map[string]string{"B": "3", "C": "4"}

		merged := MergeEnvironments(env1, env2)

		if merged["A"] != "1" {
			t.Error("expected A=1")
		}
		if merged["B"] != "3" {
			t.Error("expected B=3 (env2 should override)")
		}
		if merged["C"] != "4" {
			t.Error("expected C=4")
		}
	})

	t.Run("CleanEnvForDisplay", func(t *testing.T) {
		env := map[string]string{
			"PATH":     "/usr/bin",
			"API_KEY":  "secret123",
			"PASSWORD": "pass123",
			"HOME":     "/home/test",
		}

		clean := CleanEnvForDisplay(env)

		if clean["PATH"] != "/usr/bin" {
			t.Error("PATH should not be redacted")
		}
		if clean["API_KEY"] != "***REDACTED***" {
			t.Error("API_KEY should be redacted")
		}
		if clean["PASSWORD"] != "***REDACTED***" {
			t.Error("PASSWORD should be redacted")
		}
	})
}

// Test Guidelines for running integration tests
// These tests would require actual shell processes, so they're marked as integration tests
func TestIntegrationGuidelines(t *testing.T) {
	t.Skip("integration tests require actual shell processes")

	// Example of what an integration test would look like:
	// t.Run("spawn subshell", func(t *testing.T) {
	//     if testing.Short() {
	//         t.Skip("skipping integration test in short mode")
	//     }
	//
	//     subsheller, _ := NewDefaultSubsheller()
	//     config := &SubshellConfig{
	//         ShellType:     shell.ShellTypeBash,
	//         Timeout:       5 * time.Second,
	//         CaptureStdout: true,
	//         Command:       "echo 'Hello, World!'",
	//     }
	//
	//     ctx := context.Background()
	//     result, err := subsheller.Spawn(ctx, config)
	//     if err != nil {
	//         t.Fatalf("failed to spawn subshell: %v", err)
	//     }
	//
	//     if strings.TrimSpace(result.Stdout) != "Hello, World!" {
	//         t.Errorf("unexpected output: %s", result.Stdout)
	//     }
	// })
}
