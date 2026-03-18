// Package detector provides system detection capabilities.
// This file implements font detection for Nerd Fonts and system fonts.
package detector

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// fontDetector implements FontDetector interface.
type fontDetector struct{}

// NewFontDetector creates a new font detector.
func NewFontDetector() FontDetector {
	return &fontDetector{}
}

// Detect implements FontDetector.Detect.
func (d *fontDetector) Detect() (*FontInventory, error) {
	inventory := &FontInventory{
		Fonts:        []FontInfo{},
		NerdFonts:    []FontInfo{},
		HasNerdFonts: false,
	}

	// Use fc-list on Linux for more accurate detection
	if runtime.GOOS == "linux" {
		d.detectFontsWithFcList(inventory)
	}

	// Also scan font directories (works on all platforms)
	fontDirs := d.getFontDirectories()
	for _, dir := range fontDirs {
		d.scanFontDirectory(dir, inventory)
	}

	// Check if any Nerd Fonts were found
	for _, font := range inventory.Fonts {
		if font.IsNerdFont {
			inventory.NerdFonts = append(inventory.NerdFonts, font)
			inventory.HasNerdFonts = true
		}
	}

	// Set recommended fonts if no Nerd Fonts found
	if !inventory.HasNerdFonts {
		inventory.RecommendedFonts = []string{
			"MesloLGM Nerd Font",
			"JetBrainsMono Nerd Font",
			"FiraCode Nerd Font",
			"Hack Nerd Font",
		}
	}

	return inventory, nil
}

// detectFontsWithFcList uses fc-list command on Linux to detect fonts.
func (d *fontDetector) detectFontsWithFcList(inventory *FontInventory) {
	cmd := exec.Command("fc-list", ":family", "style", "file")
	output, err := cmd.Output()
	if err != nil {
		// fc-list not available, fall back to directory scanning
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse fc-list output: "Font Family:style=Regular:/path/to/font.ttf"
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}

		fontName := strings.TrimSpace(parts[0])
		fontPath := ""
		if len(parts) >= 3 {
			// The last part contains the file path
			fontPath = strings.TrimSpace(parts[len(parts)-1])
		}

		// Extract style if available
		isMonospace := d.isMonospaceFont(fontName)

		fontInfo := FontInfo{
			Name:        fontName,
			Path:        fontPath,
			IsNerdFont:  d.isNerdFont(fontName),
			IsMonospace: isMonospace,
		}

		// Avoid duplicates
		exists := false
		for _, f := range inventory.Fonts {
			if f.Name == fontInfo.Name {
				exists = true
				break
			}
		}

		if !exists {
			inventory.Fonts = append(inventory.Fonts, fontInfo)
		}
	}
}

// getFontDirectories returns the font directories for the current OS.
func (d *fontDetector) getFontDirectories() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var dirs []string

	switch runtime.GOOS {
	case "darwin":
		// macOS font directories
		dirs = append(dirs,
			filepath.Join(homeDir, "Library", "Fonts"),
			"/Library/Fonts",
			"/System/Library/Fonts",
		)
	case "linux":
		// Linux font directories
		dirs = append(dirs,
			filepath.Join(homeDir, ".local", "share", "fonts"),
			filepath.Join(homeDir, ".fonts"),
			"/usr/share/fonts",
			"/usr/local/share/fonts",
		)
	case "windows":
		// Windows font directory (when running on Windows/WSL)
		if windir := os.Getenv("WINDIR"); windir != "" {
			dirs = append(dirs, filepath.Join(windir, "Fonts"))
		}
	}

	return dirs
}

// scanFontDirectory scans a directory for fonts.
func (d *fontDetector) scanFontDirectory(dir string, inventory *FontInventory) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Recursively scan subdirectories
			d.scanFontDirectory(filepath.Join(dir, entry.Name()), inventory)
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		// Check for font file extensions
		if !d.isFontExtension(ext) {
			continue
		}

		// Create font info
		fontInfo := FontInfo{
			Name:        d.getFontName(name),
			Path:        filepath.Join(dir, name),
			IsNerdFont:  d.isNerdFont(name),
			IsMonospace: d.isMonospaceFont(name),
		}

		inventory.Fonts = append(inventory.Fonts, fontInfo)
	}
}

// isFontExtension checks if the file extension is a font file.
func (d *fontDetector) isFontExtension(ext string) bool {
	fontExts := map[string]bool{
		".ttf":   true,
		".otf":   true,
		".woff":  true,
		".woff2": true,
	}
	return fontExts[ext]
}

// getFontName extracts the font name from a file name.
func (d *fontDetector) getFontName(filename string) string {
	// Remove extension
	name := filename
	if idx := strings.LastIndex(filename, "."); idx > 0 {
		name = filename[:idx]
	}

	// Clean up common font name patterns
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")

	// Remove style suffixes like "Regular", "Bold", etc.
	styleSuffixes := []string{" Regular", " Bold", " Italic", " Bold Italic", " Light", " Medium"}
	for _, suffix := range styleSuffixes {
		name = strings.TrimSuffix(name, suffix)
	}

	return strings.TrimSpace(name)
}

// isNerdFont checks if the font is a Nerd Font by name.
func (d *fontDetector) isNerdFont(filename string) bool {
	name := strings.ToLower(filename)
	return strings.Contains(name, "nerd") ||
		strings.Contains(name, "nf") ||
		strings.Contains(name, "powerline")
}

// isMonospaceFont checks if the font is likely monospace.
func (d *fontDetector) isMonospaceFont(filename string) bool {
	name := strings.ToLower(filename)
	monospaceKeywords := []string{
		"mono", "monospace", "code", "console",
		"terminal", "fixed", "courier", "dejavu sans mono",
		"source code", "jetbrains", "fira code", "hack",
		"inconsolata", "droid sans mono", "roboto mono",
	}

	for _, keyword := range monospaceKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}

	return false
}

// HasRequiredSymbols checks if the font has the required Nerd Font symbols.
// This is a basic check - for accurate detection, use fontconfig or font parsing.
func (d *fontDetector) HasRequiredSymbols(fontPath string) bool {
	// Basic heuristic: if it's a Nerd Font, it has the symbols
	return d.isNerdFont(fontPath)
}
