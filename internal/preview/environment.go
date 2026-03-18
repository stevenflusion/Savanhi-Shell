// Package preview provides live preview capabilities for Savanhi Shell.
// This file implements environment variable injection for previews.
package preview

import (
	"fmt"
	"os"
	"strings"
)

// Environment variables used for preview configuration.
const (
	// EnvOhMyPoshTheme is the environment variable for Oh My Posh theme path.
	EnvOhMyPoshTheme = "POSH_THEME"

	// EnvOhMyPoshConfig is the alias for Oh My Posh config path.
	EnvOhMyPoshConfig = "POSH_CONFIG"

	// EnvTerm is the terminal type environment variable.
	EnvTerm = "TERM"

	// EnvTermProgram is the terminal program (iTerm2, etc.).
	EnvTermProgram = "TERM_PROGRAM"

	// EnvColorterm is the color terminal type.
	EnvColorterm = "COLORTERM"

	// EnvLcAll is for locale settings.
	EnvLcAll = "LC_ALL"

	// EnvLang is the language setting.
	EnvLang = "LANG"

	// EnvFontFamily is custom font family env var.
	EnvFontFamily = "FONT_FAMILY"

	// EnvFontSize is custom font size env var.
	EnvFontSize = "FONT_SIZE"

	// EnvColorScheme is custom color scheme env var.
	EnvColorScheme = "COLOR_SCHEME"

	// EnvPath is the PATH environment variable.
	EnvPath = "PATH"

	// EnvHome is the HOME environment variable.
	EnvHome = "HOME"

	// EnvUser is the USER environment variable.
	EnvUser = "USER"

	// EnvShell is the SHELL environment variable.
	EnvShell = "SHELL"
)

// DefaultEnvironmentInjector is the default implementation of EnvironmentInjector.
type DefaultEnvironmentInjector struct {
	// preserveEnv indicates whether to preserve parent environment.
	preserveEnv bool

	// extraEnv contains additional environment variables.
	extraEnv map[string]string
}

// NewDefaultEnvironmentInjector creates a new DefaultEnvironmentInjector.
func NewDefaultEnvironmentInjector() *DefaultEnvironmentInjector {
	return &DefaultEnvironmentInjector{
		preserveEnv: true,
		extraEnv:    make(map[string]string),
	}
}

// InjectEnvironment creates environment variables for a preview.
// It combines system-detected values with preview-specific values.
func (e *DefaultEnvironmentInjector) InjectEnvironment(config *PreviewConfig) (map[string]string, error) {
	env := make(map[string]string)

	// Preserve parent environment or start fresh
	if e.preserveEnv {
		for _, pair := range os.Environ() {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				env[parts[0]] = parts[1]
			}
		}
	}

	// Inject theme environment
	if config.ThemePath != "" {
		themeEnv := e.InjectThemeEnv(config.ThemePath)
		for k, v := range themeEnv {
			env[k] = v
		}
	}

	// Inject font environment
	if config.FontFamily != "" {
		fontEnv := e.InjectFontEnv(config.FontFamily, config.FontSize)
		for k, v := range fontEnv {
			env[k] = v
		}
	}

	// Inject color scheme environment
	if config.ColorScheme != "" {
		colorEnv := e.InjectColorSchemeEnv(config.ColorScheme)
		for k, v := range colorEnv {
			env[k] = v
		}
	}

	// Add custom environment variables from config
	for k, v := range config.Environment {
		env[k] = e.EscapeEnvValue(v)
	}

	// Add extra environment variables
	for k, v := range e.extraEnv {
		env[k] = e.EscapeEnvValue(v)
	}

	// Ensure essential variables are set
	e.ensureEssentialEnv(env)

	return env, nil
}

// InjectThemeEnv injects theme-specific environment variables.
func (e *DefaultEnvironmentInjector) InjectThemeEnv(themePath string) map[string]string {
	env := make(map[string]string)

	// Set Oh My Posh theme path
	env[EnvOhMyPoshTheme] = themePath
	env[EnvOhMyPoshConfig] = themePath

	return env
}

// InjectFontEnv injects font-specific environment variables.
func (e *DefaultEnvironmentInjector) InjectFontEnv(fontFamily string, fontSize int) map[string]string {
	env := make(map[string]string)

	// Set font family
	env[EnvFontFamily] = fontFamily

	// Set font size if specified
	if fontSize > 0 {
		env[EnvFontSize] = fmt.Sprintf("%d", fontSize)
	}

	return env
}

// InjectColorSchemeEnv injects color scheme environment variables.
func (e *DefaultEnvironmentInjector) InjectColorSchemeEnv(schemeName string) map[string]string {
	env := make(map[string]string)

	// Set color scheme
	env[EnvColorScheme] = schemeName

	// Ensure terminal supports colors
	if _, exists := env[EnvTerm]; !exists {
		env[EnvTerm] = "xterm-256color"
	}
	if _, exists := env[EnvColorterm]; !exists {
		env[EnvColorterm] = "truecolor"
	}

	return env
}

// EscapeEnvValue escapes special characters in environment variable values.
func (e *DefaultEnvironmentInjector) EscapeEnvValue(value string) string {
	// Escape single quotes and backslashes for shell safety
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `'\''`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, `$`, `\$`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return value
}

// ensureEssentialEnv ensures essential environment variables are set.
func (e *DefaultEnvironmentInjector) ensureEssentialEnv(env map[string]string) {
	// Ensure HOME is set
	if _, exists := env[EnvHome]; !exists {
		if homeDir, err := os.UserHomeDir(); err == nil {
			env[EnvHome] = homeDir
		}
	}

	// Ensure USER is set
	if _, exists := env[EnvUser]; !exists {
		if user := os.Getenv(EnvUser); user != "" {
			env[EnvUser] = user
		}
	}

	// Ensure TERM is set for proper terminal handling
	if _, exists := env[EnvTerm]; !exists {
		env[EnvTerm] = "xterm-256color"
	}

	// Ensure locale is set
	if _, exists := env[EnvLcAll]; !exists {
		env[EnvLcAll] = "en_US.UTF-8"
	}
	if _, exists := env[EnvLang]; !exists {
		env[EnvLang] = "en_US.UTF-8"
	}
}

// SetPreserveEnv sets whether to preserve parent environment.
func (e *DefaultEnvironmentInjector) SetPreserveEnv(preserve bool) {
	e.preserveEnv = preserve
}

// SetExtraEnv adds additional environment variables.
func (e *DefaultEnvironmentInjector) SetExtraEnv(key, value string) {
	e.extraEnv[key] = value
}

// BuildEnvSlice converts environment map to slice for exec.Cmd.
func BuildEnvSlice(env map[string]string) []string {
	slice := make([]string, 0, len(env))
	for k, v := range env {
		slice = append(slice, fmt.Sprintf("%s=%s", k, v))
	}
	return slice
}

// MergeEnvironments merges multiple environment maps.
// Later maps override earlier ones for duplicate keys.
func MergeEnvironments(envs ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, env := range envs {
		for k, v := range env {
			result[k] = v
		}
	}
	return result
}

// GetThemeEnvValue extracts theme path from environment.
func GetThemeEnvValue(env map[string]string) string {
	if themePath, exists := env[EnvOhMyPoshTheme]; exists {
		return themePath
	}
	if themePath, exists := env[EnvOhMyPoshConfig]; exists {
		return themePath
	}
	return ""
}

// GetFontEnvValue extracts font family from environment.
func GetFontEnvValue(env map[string]string) (family string, size int) {
	family = env[EnvFontFamily]
	if sizeStr, exists := env[EnvFontSize]; exists {
		fmt.Sscanf(sizeStr, "%d", &size)
	}
	return family, size
}

// GetColorSchemeEnvValue extracts color scheme from environment.
func GetColorSchemeEnvValue(env map[string]string) string {
	return env[EnvColorScheme]
}

// CleanEnvForDisplay returns a clean copy of env for display (hides sensitive values).
func CleanEnvForDisplay(env map[string]string) map[string]string {
	sensitiveKeys := []string{
		"API_KEY", "SECRET", "PASSWORD", "TOKEN", "CREDENTIAL",
		"PRIVATE", "AUTH", "KEY", "PASS",
	}

	cleanEnv := make(map[string]string)
	for k, v := range env {
		// Check for sensitive keys
		upperKey := strings.ToUpper(k)
		isSensitive := false
		for _, sensitive := range sensitiveKeys {
			if strings.Contains(upperKey, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			cleanEnv[k] = "***REDACTED***"
		} else {
			cleanEnv[k] = v
		}
	}
	return cleanEnv
}

// ValidateEnv validates that required environment variables are set.
func ValidateEnv(env map[string]string, required []string) error {
	for _, key := range required {
		if _, exists := env[key]; !exists {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}
	return nil
}

// ExtractPathEnv extracts PATH environment variable safely.
func ExtractPathEnv(env map[string]string) string {
	return env[EnvPath]
}

// PrependPath prepends a directory to PATH in the environment.
func PrependPath(env map[string]string, dir string) map[string]string {
	result := make(map[string]string)
	for k, v := range env {
		result[k] = v
	}

	currentPath := result[EnvPath]
	if currentPath == "" {
		currentPath = os.Getenv(EnvPath)
	}

	if currentPath == "" {
		result[EnvPath] = dir
	} else {
		result[EnvPath] = dir + ":" + currentPath
	}

	return result
}
