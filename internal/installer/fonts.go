// Package installer provides dependency installation and management.
// This file implements Nerd Font installation.
package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FontInstaller handles Nerd Font installation.
type FontInstaller struct {
	// context is the installation context.
	context *InstallContext

	// fontDir is the font installation directory.
	fontDir string
}

// NewFontInstaller creates a new font installer.
func NewFontInstaller(ctx *InstallContext) *FontInstaller {
	fontDir := ctx.FontDir
	if fontDir == "" {
		homeDir, _ := os.UserHomeDir()
		if runtime.GOOS == "darwin" {
			fontDir = filepath.Join(homeDir, "Library", "Fonts")
		} else {
			fontDir = filepath.Join(homeDir, ".local", "share", "fonts")
		}
	}

	return &FontInstaller{
		context: ctx,
		fontDir: fontDir,
	}
}

// Install installs a Nerd Font.
func (i *FontInstaller) Install(ctx context.Context, fontName string, opts *Options) (*InstallResult, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	result := &InstallResult{
		Component: fontName,
	}

	// Check if already installed
	if i.IsInstalled(fontName) && !i.context.Force {
		result.Success = true
		result.InstalledPath = i.fontDir
		return result, nil
	}

	// Ensure font directory exists
	if err := os.MkdirAll(i.fontDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create font directory: %w", err)
	}

	// Get download URL
	downloadURL := i.getDownloadURL(fontName)

	// Download font archive
	cacheDir := filepath.Join(i.context.CacheDir, "fonts")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	installer := &DefaultInstaller{context: i.context}
	downloadResult, err := installer.downloadFile(ctx, downloadURL, cacheDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to download font: %w", err)
	}

	// Install fonts from archive
	if strings.HasSuffix(downloadResult.LocalPath, ".zip") {
		if err := i.installFromZip(downloadResult.LocalPath, fontName, result); err != nil {
			return nil, fmt.Errorf("failed to install font: %w", err)
		}
	} else if strings.HasSuffix(downloadResult.LocalPath, ".tar.xz") {
		if err := i.installFromTarXz(downloadResult.LocalPath, fontName, result); err != nil {
			return nil, fmt.Errorf("failed to install font: %w", err)
		}
	} else {
		// Single font file
		targetPath := filepath.Join(i.fontDir, filepath.Base(downloadResult.LocalPath))
		if err := copyFile(downloadResult.LocalPath, targetPath); err != nil {
			return nil, fmt.Errorf("failed to install font: %w", err)
		}
		result.InstalledPath = targetPath
	}

	// Clean up temp files if not caching
	if !opts.UseCache {
		os.Remove(downloadResult.LocalPath)
	}

	result.Success = true

	// Refresh font cache on Linux
	if runtime.GOOS != "darwin" {
		if err := i.refreshFontCache(); err != nil {
			result.Warnings = append(result.Warnings, "Failed to refresh font cache")
		}
	}

	return result, nil
}

// IsInstalled checks if a font is installed.
func (i *FontInstaller) IsInstalled(fontName string) bool {
	// Check for common font file patterns
	patterns := []string{
		fontName + "*.ttf",
		fontName + "*.otf",
		fontName + "NerdFont*",
		"*NerdFont*" + fontName + "*",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(i.fontDir, pattern))
		if err == nil && len(matches) > 0 {
			return true
		}
	}

	// Also check system font directories
	systemFontDirs := i.getSystemFontDirs()
	for _, dir := range systemFontDirs {
		for _, pattern := range patterns {
			matches, err := filepath.Glob(filepath.Join(dir, pattern))
			if err == nil && len(matches) > 0 {
				return true
			}
		}
	}

	return false
}

// getSystemFontDirs returns system font directories.
func (i *FontInstaller) getSystemFontDirs() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"/Library/Fonts", "/System/Library/Fonts"}
	case "linux":
		return []string{"/usr/share/fonts", "/usr/local/share/fonts"}
	default:
		return nil
	}
}

// getDownloadURL returns the download URL for a Nerd Font.
func (i *FontInstaller) getDownloadURL(fontName string) string {
	// Nerd Fonts are hosted on GitHub releases
	// Format: https://github.com/ryanoasis/nerd-fonts/releases/latest/download/{FontName}.zip

	// Normalize font name
	normalizedName := fontName
	if !strings.Contains(normalizedName, "NerdFont") {
		// Common font name mappings
		fontMap := map[string]string{
			"MesloLGM":    "Meslo",
			"MesloLGM-NF": "Meslo",
			"MesloLGS":    "Meslo",
			"FiraCode":    "FiraCode",
			"JetBrains":   "JetBrainsMono",
			"SourceCode":  "SourceCodePro",
			"Ubuntu":      "UbuntuMono",
		}

		if mapped, ok := fontMap[normalizedName]; ok {
			normalizedName = mapped
		}
	}

	return fmt.Sprintf("https://github.com/ryanoasis/nerd-fonts/releases/latest/download/%s.zip", normalizedName)
}

// installFromZip installs fonts from a zip archive.
func (i *FontInstaller) installFromZip(zipPath, fontName string, result *InstallResult) error {
	// Use shared implementation from download.go
	installer := &DefaultInstaller{context: i.context}
	return installer.installFontFromZip(zipPath, &Dependency{Name: fontName, Type: ComponentTypeFont}, result)
}

// installFromTarXz installs fonts from a tar.xz archive.
func (i *FontInstaller) installFromTarXz(archivePath, fontName string, result *InstallResult) error {
	// tar.xz is less common for fonts, but handle it
	// For now, we'll skip this as most Nerd Fonts come as zip
	return fmt.Errorf("tar.xz format not supported for fonts, please use zip format")
}

// refreshFontCache refreshes the font cache.
func (i *FontInstaller) refreshFontCache() error {
	// Use fc-cache if available
	if _, err := exec.LookPath("fc-cache"); err != nil {
		// fc-cache not available, skip
		return nil
	}

	cmd := exec.Command("fc-cache", "-fv", i.fontDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to refresh font cache: %w\n%s", err, string(output))
	}

	return nil
}

// Uninstall removes a font.
func (i *FontInstaller) Uninstall(fontName string) error {
	// Find and remove all matching font files
	patterns := []string{
		fontName + "*.ttf",
		fontName + "*.otf",
		fontName + "NerdFont*",
		"*" + fontName + "*",
	}

	removed := false
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(i.fontDir, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			if err := os.Remove(match); err != nil {
				return fmt.Errorf("failed to remove font %s: %w", match, err)
			}
			removed = true
		}
	}

	if !removed {
		return fmt.Errorf("font %s not found", fontName)
	}

	// Refresh font cache
	if runtime.GOOS != "darwin" {
		if err := i.refreshFontCache(); err != nil {
			// Non-fatal
		}
	}

	return nil
}

// ListInstalled lists installed Nerd Fonts.
func (i *FontInstaller) ListInstalled() ([]string, error) {
	fonts := make([]string, 0)

	// Search for Nerd Font files
	pattern := filepath.Join(i.fontDir, "*NerdFont*")
	matches, err := filepath.Glob(pattern)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	seen := make(map[string]bool)
	for _, match := range matches {
		// Extract font name from filename
		base := filepath.Base(match)
		// Common patterns: FontName-NF-Regular.ttf, FontNameNerdFont-Regular.ttf
		if idx := strings.Index(base, "NerdFont"); idx > 0 {
			name := base[:idx]
			// Clean up name
			name = strings.TrimSuffix(name, "-")
			name = strings.TrimSuffix(name, "_")
			if !seen[name] {
				fonts = append(fonts, name)
				seen[name] = true
			}
		}
	}

	// Also check system directories
	for _, dir := range i.getSystemFontDirs() {
		matches, _ = filepath.Glob(filepath.Join(dir, "*NerdFont*"))
		for _, match := range matches {
			base := filepath.Base(match)
			if idx := strings.Index(base, "NerdFont"); idx > 0 {
				name := base[:idx]
				name = strings.TrimSuffix(name, "-")
				name = strings.TrimSuffix(name, "_")
				if !seen[name] {
					fonts = append(fonts, name+" (system)")
					seen[name] = true
				}
			}
		}
	}

	return fonts, nil
}

// GetRecommendedFonts returns recommended Nerd Fonts.
func (i *FontInstaller) GetRecommendedFonts() []string {
	return []string{
		"MesloLGM-NF",
		"JetBrainsMono-NF",
		"FiraCode-NF",
		"SourceCodePro-NF",
		"UbuntuMono-NF",
	}
}
