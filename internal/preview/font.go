// Package preview provides live preview capabilities for Savanhi Shell.
// This file implements font preview functionality.
package preview

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Font preview constants.
const (
	// DefaultFontTimeout is the default timeout for font previews.
	DefaultFontTimeout = 3 * time.Second

	// DefaultFontSize is the default font size for previews.
	DefaultFontSize = 12

	// SampleTextNerdFont is the text to display for Nerd Font previews.
	SampleTextNerdFont = "  \uf31b \uf015 \uf013 \uf07b \uf008 \uf0e0 \uf07c \uf0e7 \uf002 \uf201"

	// SampleTextBasic is the basic text sample for font previews.
	SampleTextBasic = "The quick brown fox jumps over the lazy dog 0123456789"

	// SampleTextFull is the full sample with various characters.
	SampleTextFull = "ABCDEFabcdef © ® ™ € £ ¥ $ \uf31b → ← ↑ ↓"
)

// Common errors for font preview operations.
var (
	// ErrFontNotFound indicates the font is not installed.
	ErrFontNotFound = errors.New("font not found")

	// ErrFontPreviewFailed indicates the font preview failed.
	ErrFontPreviewFailed = errors.New("font preview failed")

	// ErrFontConfigNotAvailable indicates fontconfig is not available.
	ErrFontConfigNotAvailable = errors.New("fontconfig not available")
)

// FontInfo contains information about a font.
type FontInfo struct {
	// Name is the font family name.
	Name string `json:"name"`

	// Path is the path to the font file.
	Path string `json:"path"`

	// IsNerdFont indicates whether this is a Nerd Font.
	IsNerdFont bool `json:"is_nerd_font"`

	// IsMonospace indicates whether the font is monospaced.
	IsMonospace bool `json:"is_monospace"`
}

// FontPreviewResult contains the result of a font preview.
type FontPreviewResult struct {
	// FontFamily is the font family name.
	FontFamily string `json:"font_family"`

	// FontSize is the font size used.
	FontSize int `json:"font_size"`

	// IsNerdFont indicates if this is a Nerd Font.
	IsNerdFont bool `json:"is_nerd_font"`

	// SampleOutput is the preview sample output.
	SampleOutput string `json:"sample_output"`

	// Available indicates if the font is available.
	Available bool `json:"available"`

	// Path is the path to the font file.
	Path string `json:"path,omitempty"`
}

// DefaultFontPreviewHandler implements FontPreview interface.
type DefaultFontPreviewHandler struct {
	// subsheller is the subshell spawner.
	subsheller SubshellSpawner

	// injector is the environment injector.
	injector EnvironmentInjector

	// fontDirs are the directories to search for fonts.
	fontDirs []string

	// cachedFonts is the cache of detected fonts.
	cachedFonts map[string]*FontInfo
}

// NewDefaultFontPreviewHandler creates a new DefaultFontPreviewHandler.
func NewDefaultFontPreviewHandler(subsheller SubshellSpawner, injector EnvironmentInjector) (*DefaultFontPreviewHandler, error) {
	handler := &DefaultFontPreviewHandler{
		subsheller:  subsheller,
		injector:    injector,
		fontDirs:    getDefaultFontDirs(),
		cachedFonts: make(map[string]*FontInfo),
	}

	return handler, nil
}

// PreviewFont generates a preview for a specific font.
func (f *DefaultFontPreviewHandler) PreviewFont(ctx context.Context, config *FontPreviewConfig) (*PreviewResult, error) {
	// Check font availability first
	available, err := f.CheckFontAvailability(config.FontFamily)
	if err != nil {
		return nil, fmt.Errorf("failed to check font availability: %w", err)
	}

	if !available {
		return nil, ErrFontNotFound
	}

	// Set default timeout
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultFontTimeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build environment
	env := f.buildFontEnv(config)

	// Determine sample text
	sampleText := config.SampleText
	if sampleText == "" {
		if config.ShowNerdFontIcons {
			sampleText = SampleTextNerdFont
		} else {
			sampleText = SampleTextBasic
		}
	}

	// Create subshell config
	subshellConfig := &SubshellConfig{
		ShellType:     config.Shell,
		Environment:   env,
		Timeout:       timeout,
		CaptureStdout: true,
		CaptureStderr: true,
		Command:       fmt.Sprintf("echo '%s'", escapeForShell(sampleText)),
	}

	// Spawn subshell
	result, err := f.subsheller.Spawn(ctx, subshellConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFontPreviewFailed, err)
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
			Type:       PreviewTypeFont,
			Shell:      config.Shell,
			FontFamily: config.FontFamily,
			FontSize:   config.FontSize,
			Timeout:    config.Timeout,
		},
	}

	if result.ExitCode != 0 {
		previewResult.Status = StatusFailed
		previewResult.ErrorMessage = result.Stderr
	}

	return previewResult, nil
}

// CheckFontAvailability checks if a font is installed on the system.
func (f *DefaultFontPreviewHandler) CheckFontAvailability(fontFamily string) (bool, error) {
	// Check cache first
	if info, exists := f.cachedFonts[fontFamily]; exists {
		return info != nil && info.Path != "", nil
	}

	// Find font
	info, err := f.findFont(fontFamily)
	if err != nil {
		return false, nil
	}

	if info != nil {
		f.cachedFonts[fontFamily] = info
		return true, nil
	}

	return false, nil
}

// GetInstalledFonts returns a list of installed fonts.
func (f *DefaultFontPreviewHandler) GetInstalledFonts(nerdFontsOnly bool) ([]string, error) {
	fonts, err := f.listAllFonts()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for _, font := range fonts {
		if nerdFontsOnly {
			// Check if font name contains "Nerd"
			if strings.Contains(strings.ToLower(font), "nerd") {
				result = append(result, font)
			}
		} else {
			result = append(result, font)
		}
	}

	return result, nil
}

// findFont finds a font by name on the system.
func (f *DefaultFontPreviewHandler) findFont(fontFamily string) (*FontInfo, error) {
	// Use fontconfig on Linux/macOS
	if hasFontConfig() {
		return f.findFontViaFontConfig(fontFamily)
	}

	// Fall back to searching font directories
	return f.findFontInDirs(fontFamily)
}

// findFontViaFontConfig finds a font using fontconfig.
func (f *DefaultFontPreviewHandler) findFontViaFontConfig(fontFamily string) (*FontInfo, error) {
	// Use fc-list to find the font
	cmd := exec.Command("fc-list", fontFamily, "file")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return nil, nil
	}

	// Return first match
	return &FontInfo{
		Name:        fontFamily,
		Path:        strings.TrimSpace(lines[0]),
		IsNerdFont:  strings.Contains(strings.ToLower(fontFamily), "nerd"),
		IsMonospace: true, // Assume monospace for terminal fonts
	}, nil
}

// findFontInDirs finds a font by searching directories.
func (f *DefaultFontPreviewHandler) findFontInDirs(fontFamily string) (*FontInfo, error) {
	fontName := strings.ToLower(fontFamily)

	for _, dir := range f.fontDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			// Check font file name
			fileName := strings.ToLower(entry.Name())
			if strings.Contains(fileName, fontName) {
				return &FontInfo{
					Name:        fontFamily,
					Path:        filepath.Join(dir, entry.Name()),
					IsNerdFont:  strings.Contains(fileName, "nerd"),
					IsMonospace: true,
				}, nil
			}
		}
	}

	return nil, nil
}

// listAllFonts lists all available fonts on the system.
func (f *DefaultFontPreviewHandler) listAllFonts() ([]string, error) {
	if hasFontConfig() {
		return f.listFontsViaFontConfig()
	}

	return f.listFontsInDirs()
}

// listFontsViaFontConfig lists fonts using fontconfig.
func (f *DefaultFontPreviewHandler) listFontsViaFontConfig() ([]string, error) {
	cmd := exec.Command("fc-list", ":", "family")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	fonts := make([]string, 0, len(lines))
	seen := make(map[string]bool)

	for _, line := range lines {
		// Extract family name
		parts := strings.Split(line, ",")
		if len(parts) > 0 {
			family := strings.TrimSpace(parts[0])
			if !seen[family] {
				seen[family] = true
				fonts = append(fonts, family)
			}
		}
	}

	return fonts, nil
}

// listFontsInDirs lists fonts by searching directories.
func (f *DefaultFontPreviewHandler) listFontsInDirs() ([]string, error) {
	fonts := make([]string, 0)
	seen := make(map[string]bool)

	fontExts := []string{".ttf", ".otf", ".ttc", ".woff", ".woff2"}

	for _, dir := range f.fontDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			name := entry.Name()
			ext := strings.ToLower(filepath.Ext(name))
			for _, fontExt := range fontExts {
				if ext == fontExt {
					// Extract font name from filename
					fontName := strings.TrimSuffix(name, filepath.Ext(name))
					fontName = strings.ReplaceAll(fontName, "-", " ")
					fontName = strings.ReplaceAll(fontName, "_", " ")

					if !seen[fontName] {
						seen[fontName] = true
						fonts = append(fonts, fontName)
					}
					break
				}
			}
		}
	}

	return fonts, nil
}

// buildFontEnv builds environment variables for font preview.
func (f *DefaultFontPreviewHandler) buildFontEnv(config *FontPreviewConfig) map[string]string {
	env := make(map[string]string)

	// Preserve current environment
	for _, pair := range os.Environ() {
		if idx := strings.Index(pair, "="); idx > 0 {
			env[pair[:idx]] = pair[idx+1:]
		}
	}

	// Add font environment
	fontEnv := f.injector.InjectFontEnv(config.FontFamily, config.FontSize)
	for k, v := range fontEnv {
		env[k] = v
	}

	return env
}

// hasFontConfig checks if fontconfig is available.
func hasFontConfig() bool {
	_, err := exec.LookPath("fc-list")
	return err == nil
}

// getDefaultFontDirs returns default font directories for the current platform.
func getDefaultFontDirs() []string {
	homeDir, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(homeDir, "Library", "Fonts"),
			"/Library/Fonts",
			"/System/Library/Fonts",
		}
	case "linux":
		return []string{
			filepath.Join(homeDir, ".local", "share", "fonts"),
			filepath.Join(homeDir, ".fonts"),
			"/usr/share/fonts",
			"/usr/local/share/fonts",
		}
	case "windows":
		return []string{
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Windows", "Fonts"),
			filepath.Join(os.Getenv("WINDIR"), "Fonts"),
		}
	default:
		return []string{
			filepath.Join(homeDir, ".fonts"),
		}
	}
}

// escapeForShell escapes a string for shell output.
func escapeForShell(s string) string {
	// Escape single quotes and backslashes
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `'\''`)
	return s
}

// GetNerdFontIcons returns common Nerd Font icons for display.
func GetNerdFontIcons() []string {
	return []string{
		"\uf31b", // Folder
		"\uf015", // Home
		"\uf013", // Gear
		"\uf07b", // Folder open
		"\uf008", // Git branch
		"\uf0e0", // Document
		"\uf07c", // Folder
		"\uf0e7", // Lightning
		"\uf002", // Search
		"\uf201", // Chart
	}
}

// IsNerdFont checks if a font is likely a Nerd Font based on name.
func IsNerdFont(fontName string) bool {
	name := strings.ToLower(fontName)
	return strings.Contains(name, "nerd") ||
		strings.Contains(name, "nf") ||
		strings.Contains(name, "nerd font")
}

// GetRecommendedFonts returns recommended Nerd Fonts for terminal use.
func GetRecommendedFonts() []string {
	return []string{
		"MesloLGM Nerd Font",
		"JetBrainsMono Nerd Font",
		"FiraCode Nerd Font",
		"Hack Nerd Font",
		"SourceCodePro Nerd Font",
		"Mononoki Nerd Font",
		"DejaVuSansMono Nerd Font",
		"RobotoMono Nerd Font",
		"UbuntuMono Nerd Font",
		"Consolas Nerd Font",
	}
}
