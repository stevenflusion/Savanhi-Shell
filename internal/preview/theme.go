// Package preview provides live preview capabilities for Savanhi Shell.
// This file implements theme preview functionality.
package preview

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Theme manifest constants.
const (
	// DefaultThemeTimeout is the default timeout for theme previews.
	DefaultThemeTimeout = 5 * time.Second

	// ThemeConfigFile is the standard name for theme config files.
	ThemeConfigFile = "theme.json"

	// BundledThemesDir is the directory for bundled themes.
	BundledThemesDir = "themes"
)

// bundledThemeNames is the list of themes bundled with Savanhi Shell.
var bundledThemeNames = []string{
	"paradox",
	"powerlevel10k_rainbow",
	"powerlevel10k_lean",
	"agnoster",
	"atomic",
	"captivity",
	"darkblood",
	"fish",
	"jandedobbeleer",
	"marcduiker",
	"microverse-power",
	"minimal",
	"negligible",
	"night-owl",
	"powerline",
}

// GetBundledThemes returns the list of bundled theme names.
// This function can be called without instantiating a ThemeProvider.
func GetBundledThemes() []string {
	return bundledThemeNames
}

// Common errors for theme preview operations.
var (
	// ErrThemeNotFound indicates the theme was not found.
	ErrThemeNotFound = errors.New("theme not found")

	// ErrThemeLoadFailed indicates failed to load theme.
	ErrThemeLoadFailed = errors.New("failed to load theme")

	// ErrOhMyPoshNotFound indicates oh-my-posh is not installed.
	ErrOhMyPoshNotFound = errors.New("oh-my-posh not found")

	// ErrThemeDownloadFailed indicates failed to download theme.
	ErrThemeDownloadFailed = errors.New("failed to download theme")
)

// ThemeInfo contains information about a theme.
type ThemeInfo struct {
	// Name is the theme name.
	Name string `json:"name"`

	// DisplayName is the human-readable theme name.
	DisplayName string `json:"display_name"`

	// Description is a brief theme description.
	Description string `json:"description"`

	// Author is the theme author.
	Author string `json:"author,omitempty"`

	// Version is the theme version.
	Version string `json:"version,omitempty"`

	// Path is the path to the theme file.
	Path string `json:"path"`

	// IsBundled indicates if the theme is bundled with Savanhi.
	IsBundled bool `json:"is_bundled"`

	// URL is the download URL for remote themes.
	URL string `json:"url,omitempty"`

	// Tags are theme category tags.
	Tags []string `json:"tags,omitempty"`
}

// DefaultThemeProvider implements ThemePreview interface.
type DefaultThemeProvider struct {
	// subsheller is the subshell spawner.
	subsheller SubshellSpawner

	// injector is the environment injector.
	injector EnvironmentInjector

	// bundledThemes is the embedded bundled themes.
	bundledThemes embed.FS

	// themesDir is the directory for cached themes.
	themesDir string

	// mu protects the themes cache.
	mu sync.RWMutex

	// themesCache is the cached list of themes.
	themesCache map[string]*ThemeInfo

	// ohMyPoshPath is the path to oh-my-posh binary.
	ohMyPoshPath string
}

// NewDefaultThemeProvider creates a new DefaultThemeProvider.
func NewDefaultThemeProvider(subsheller SubshellSpawner, injector EnvironmentInjector) (*DefaultThemeProvider, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	themesDir := filepath.Join(homeDir, ".config", "savanhi", "themes")

	provider := &DefaultThemeProvider{
		subsheller:   subsheller,
		injector:     injector,
		themesDir:    themesDir,
		themesCache:  make(map[string]*ThemeInfo),
		ohMyPoshPath: "", // Will be detected when needed
	}

	// Ensure themes directory exists
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create themes directory: %w", err)
	}

	return provider, nil
}

// PreviewTheme generates a preview for a specific theme.
func (t *DefaultThemeProvider) PreviewTheme(ctx context.Context, config *ThemePreviewConfig) (*PreviewResult, error) {
	// Validate configuration
	if config.ThemePath == "" && config.ThemeContent == "" {
		return nil, fmt.Errorf("theme path or content is required")
	}

	// Check for oh-my-posh
	ompPath := config.OhMyPoshPath
	if ompPath == "" {
		var err error
		ompPath, err = t.findOhMyPosh()
		if err != nil {
			return nil, ErrOhMyPoshNotFound
		}
	}

	// Create subshell config
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultThemeTimeout
	}

	// Build environment
	env := t.buildThemeEnv(config, ompPath)

	// Create subshell config
	subshellConfig := &SubshellConfig{
		ShellType:     config.Shell,
		Environment:   env,
		Timeout:       timeout,
		CaptureStdout: true,
		CaptureStderr: true,
		Command:       t.buildPreviewCommand(ompPath, config),
	}

	// Spawn subshell
	result, err := t.subsheller.Spawn(ctx, subshellConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to spawn preview: %w", err)
	}

	// Build preview result
	previewResult := &PreviewResult{
		ID:          generatePreviewID(),
		Status:      StatusCompleted,
		Output:      result.Stdout,
		ErrorOutput: result.Stderr,
		ExitCode:    result.ExitCode,
		Duration:    result.Duration,
		Config: PreviewConfig{
			Type:      PreviewTypeTheme,
			Shell:     config.Shell,
			ThemePath: config.ThemePath,
			ThemeName: config.ThemeName,
			Timeout:   config.Timeout,
		},
	}

	// Check for errors
	if result.ExitCode != 0 {
		previewResult.Status = StatusFailed
		previewResult.ErrorMessage = result.Stderr
	}

	// Note: Temp files are cleaned up by the DefaultSubsheller automatically
	// after the Spawn call completes

	return previewResult, nil
}

// ListAvailableThemes returns a list of available themes.
func (t *DefaultThemeProvider) ListAvailableThemes() ([]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return cached themes if available
	if len(t.themesCache) > 0 {
		themes := make([]string, 0, len(t.themesCache))
		for name := range t.themesCache {
			themes = append(themes, name)
		}
		return themes, nil
	}

	// Load bundled themes
	themes := make([]string, 0)

	// Scan bundled themes directory
	bundledThemes, err := t.listBundledThemes()
	if err == nil {
		themes = append(themes, bundledThemes...)
	}

	// Scan cached themes directory
	cachedThemes, err := t.listCachedThemes()
	if err == nil {
		themes = append(themes, cachedThemes...)
	}

	return themes, nil
}

// GetThemePath returns the path to a theme file.
func (t *DefaultThemeProvider) GetThemePath(themeName string) (string, error) {
	// Check bundled themes first
	bundledPath := t.getBundledThemePath(themeName)
	if bundledPath != "" {
		return bundledPath, nil
	}

	// Check cached themes
	cachedPath := filepath.Join(t.themesDir, themeName, ThemeConfigFile)
	if _, err := os.Stat(cachedPath); err == nil {
		return cachedPath, nil
	}

	return "", ErrThemeNotFound
}

// findOhMyPosh finds the oh-my-posh binary.
func (t *DefaultThemeProvider) findOhMyPosh() (string, error) {
	// Use cached path if available
	if t.ohMyPoshPath != "" {
		return t.ohMyPoshPath, nil
	}

	// Check common locations
	locations := []string{
		"oh-my-posh",
		"/usr/local/bin/oh-my-posh",
		"/usr/bin/oh-my-posh",
		"/opt/homebrew/bin/oh-my-posh",
	}

	// Check user's home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		locations = append(locations,
			filepath.Join(homeDir, ".local", "bin", "oh-my-posh"),
			filepath.Join(homeDir, ".oh-my-posh", "bin", "oh-my-posh"),
		)
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			t.ohMyPoshPath = loc
			return loc, nil
		}
	}

	return "", ErrOhMyPoshNotFound
}

// buildThemeEnv builds environment variables for theme preview.
func (t *DefaultThemeProvider) buildThemeEnv(config *ThemePreviewConfig, ompPath string) map[string]string {
	env := make(map[string]string)

	// Preserve current environment
	for _, pair := range os.Environ() {
		if idx := findEqual(pair); idx > 0 {
			env[pair[:idx]] = pair[idx+1:]
		}
	}

	// Set oh-my-posh theme
	if config.ThemePath != "" {
		env[EnvOhMyPoshTheme] = config.ThemePath
	}

	// Add theme environment from injector
	themeEnv := t.injector.InjectThemeEnv(config.ThemePath)
	for k, v := range themeEnv {
		env[k] = v
	}

	// Ensure PATH includes oh-my-posh location
	ompDir := filepath.Dir(ompPath)
	if currentPath, exists := env[EnvPath]; exists {
		env[EnvPath] = ompDir + ":" + currentPath
	} else {
		env[EnvPath] = ompDir
	}

	return env
}

// buildPreviewCommand builds the command to run for theme preview.
func (t *DefaultThemeProvider) buildPreviewCommand(ompPath string, config *ThemePreviewConfig) string {
	// The preview command runs oh-my-posh and captures the output
	// We use 'print' to get the prompt without actually changing the shell
	if config.ThemeContent != "" {
		// For inline theme content, we need to write it to a temp file first
		// This is handled by the subshell spawner which creates a temp RC file
		return fmt.Sprintf("%s print", ompPath)
	}

	return fmt.Sprintf("%s print", ompPath)
}

// listBundledThemes lists themes bundled with the application.
func (t *DefaultThemeProvider) listBundledThemes() ([]string, error) {
	// This would normally read from embedded FS
	// For now, return a list of known bundled themes
	// In production, this would read from the embedded themes directory
	bundledThemes := []string{
		"paradox",
		"powerlevel10k_rainbow",
		"powerlevel10k_lean",
		"agnoster",
		"atomic",
		" captivity",
		"darkblood",
		"fish",
		"jandedobbeleer",
		"marcduiker",
		"microverse-power",
		"minimal",
		"negligible",
		"night-owl",
		"powerline",
	}

	return bundledThemes, nil
}

// listCachedThemes lists themes cached locally.
func (t *DefaultThemeProvider) listCachedThemes() ([]string, error) {
	themes := make([]string, 0)

	entries, err := os.ReadDir(t.themesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return themes, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it has a theme.json file
			themeFile := filepath.Join(t.themesDir, entry.Name(), ThemeConfigFile)
			if _, err := os.Stat(themeFile); err == nil {
				themes = append(themes, entry.Name())
			}
		}
	}

	return themes, nil
}

// getBundledThemePath returns the path for a bundled theme.
func (t *DefaultThemeProvider) getBundledThemePath(themeName string) string {
	// In production, this would extract from embedded FS
	// For now, check if theme name is in known bundled themes
	bundledThemes, _ := t.listBundledThemes()
	for _, name := range bundledThemes {
		if name == themeName {
			// Return a path that would be extracted to temp
			return filepath.Join(t.themesDir, "bundled", themeName+".omp.json")
		}
	}
	return ""
}

// DownloadTheme downloads a theme from a URL.
func (t *DefaultThemeProvider) DownloadTheme(ctx context.Context, url string, name string) error {
	// Create theme directory
	themeDir := filepath.Join(t.themesDir, name)
	if err := os.MkdirAll(themeDir, 0755); err != nil {
		return fmt.Errorf("%w: failed to create theme directory: %v", ErrThemeDownloadFailed, err)
	}

	// Download would happen here in production
	// For now, this is a placeholder
	return fmt.Errorf("%w: theme download not implemented", ErrThemeDownloadFailed)
}

// RemoveTheme removes a cached theme.
func (t *DefaultThemeProvider) RemoveTheme(name string) error {
	themeDir := filepath.Join(t.themesDir, name)

	// Don't remove bundled themes
	if t.isBundledTheme(name) {
		return fmt.Errorf("cannot remove bundled theme: %s", name)
	}

	if err := os.RemoveAll(themeDir); err != nil {
		return fmt.Errorf("failed to remove theme: %w", err)
	}

	t.mu.Lock()
	delete(t.themesCache, name)
	t.mu.Unlock()

	return nil
}

// isBundledTheme checks if a theme is bundled.
func (t *DefaultThemeProvider) isBundledTheme(name string) bool {
	bundledThemes, _ := t.listBundledThemes()
	for _, bname := range bundledThemes {
		if bname == name {
			return true
		}
	}
	return false
}

// findEqual finds the position of '=' in a string.
func findEqual(s string) int {
	for i, c := range s {
		if c == '=' {
			return i
		}
	}
	return -1
}

// generatePreviewID generates a unique preview ID.
func generatePreviewID() string {
	return fmt.Sprintf("preview-%d", time.Now().UnixNano())
}

// CreateTempThemeFile creates a temporary file for theme content.
func CreateTempThemeFile(content string) (string, error) {
	tempDir, err := os.MkdirTemp("", "savanhi-theme-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	themeFile := filepath.Join(tempDir, "theme.omp.json")
	if err := os.WriteFile(themeFile, []byte(content), 0644); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to write theme file: %w", err)
	}

	return themeFile, nil
}

// ValidateTheme validates a theme file.
func ValidateTheme(themePath string) error {
	// Read theme file
	content, err := os.ReadFile(themePath)
	if err != nil {
		return fmt.Errorf("%w: failed to read theme file: %v", ErrThemeLoadFailed, err)
	}

	// Basic validation - check if it's valid JSON
	// Oh-my-posh themes are JSON files
	if len(content) == 0 {
		return fmt.Errorf("%w: theme file is empty", ErrThemeLoadFailed)
	}

	// Check for common JSON structure indicators
	contentStr := string(content)
	if contentStr[0] != '{' && contentStr[0] != '[' {
		return fmt.Errorf("%w: theme file is not valid JSON", ErrThemeLoadFailed)
	}

	return nil
}

// GetThemeDisplayName returns a human-readable theme name.
func GetThemeDisplayName(themeName string) string {
	// Convert snake_case to Title Case
	words := splitWords(themeName)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = string(word[0]-32) + word[1:]
		}
	}
	return joinWords(words)
}

// splitWords splits a snake_case or kebab-case string into words.
func splitWords(s string) []string {
	var words []string
	var currentWord []rune

	for _, c := range s {
		if c == '_' || c == '-' {
			if len(currentWord) > 0 {
				words = append(words, string(currentWord))
				currentWord = nil
			}
		} else {
			currentWord = append(currentWord, c)
		}
	}

	if len(currentWord) > 0 {
		words = append(words, string(currentWord))
	}

	return words
}

// joinWords joins words with spaces.
func joinWords(words []string) string {
	result := ""
	for i, word := range words {
		if i > 0 {
			result += " "
		}
		result += word
	}
	return result
}

// fs.FS interface for embedded themes
var _ fs.FS = (*embed.FS)(nil)
