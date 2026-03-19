// Package tui provides the Bubble Tea TUI for Savanhi Shell.
// This file implements health dashboard data structures.
package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/savanhi/shell/internal/detector"
	"github.com/savanhi/shell/internal/installer"
)

// HealthData aggregates health check results for the health dashboard.
// It contains terminal capabilities, component status, and font/color test results.
type HealthData struct {
	// Terminal contains detected terminal capabilities.
	Terminal *TerminalCapabilities

	// Components maps component names to their installation status.
	Components map[string]*ComponentStatus

	// FontTest contains font rendering test results.
	FontTest *FontTestResult

	// ColorTest contains color capability test results.
	ColorTest *ColorTestResult

	// Errors contains any errors encountered during health checks.
	Errors []string

	// CheckedAt is the timestamp when health checks were performed.
	CheckedAt time.Time

	// ExportPath is the path where the health report will be exported.
	ExportPath string
}

// TerminalCapabilities wraps detector.TerminalInfo for health dashboard display.
// It provides a clean interface for terminal feature detection.
type TerminalCapabilities struct {
	// TrueColor indicates 24-bit true color support.
	TrueColor bool

	// Ligatures indicates font ligature support.
	Ligatures bool

	// Hyperlinks indicates OSC 8 hyperlink support.
	Hyperlinks bool

	// KittyGraphics indicates Kitty graphics protocol support.
	KittyGraphics bool

	// TerminalName is the detected terminal emulator name.
	TerminalName string
}

// ComponentStatus wraps installer.VerificationResult for health dashboard display.
// It provides a simplified view of component installation status.
type ComponentStatus struct {
	// Name is the component name.
	Name string

	// Installed indicates if the component is installed.
	Installed bool

	// Version is the installed version (if any).
	Version string

	// Issues contains any problems detected with the component.
	Issues []string
}

// FontTestResult contains font rendering test results.
// It validates Nerd Font glyph rendering capability.
type FontTestResult struct {
	// GlyphsRendered indicates if Nerd Font glyphs are visible.
	GlyphsRendered bool

	// FallbackUsed indicates if ASCII fallback is active.
	FallbackUsed bool

	// TestGlyphs are the glyphs used for testing.
	TestGlyphs []string
}

// ColorTestResult contains color capability test results.
// It validates terminal color mode support.
type ColorTestResult struct {
	// ColorMode is the detected color mode: "truecolor", "256", or "ansi16".
	ColorMode string

	// GradientOK indicates if the gradient test passed.
	GradientOK bool

	// PaletteOK indicates if the palette test passed.
	PaletteOK bool
}

// NewHealthData creates a new HealthData instance with default values.
func NewHealthData() *HealthData {
	return &HealthData{
		Components: make(map[string]*ComponentStatus),
		Errors:     []string{},
		CheckedAt:  time.Now(),
	}
}

// NewTerminalCapabilities creates a TerminalCapabilities from detector.TerminalInfo.
func NewTerminalCapabilities(info *detector.TerminalInfo) *TerminalCapabilities {
	if info == nil {
		return &TerminalCapabilities{}
	}

	return &TerminalCapabilities{
		TrueColor:     info.SupportsTrueColor,
		Ligatures:     info.SupportsLigatures,
		Hyperlinks:    info.SupportsHyperlinks,
		KittyGraphics: info.SupportsKittyGraphics,
		TerminalName:  info.Name,
	}
}

// NewComponentStatus creates a ComponentStatus from installer.VerificationResult.
func NewComponentStatus(result *installer.VerificationResult) *ComponentStatus {
	if result == nil {
		return &ComponentStatus{
			Issues: []string{},
		}
	}

	return &ComponentStatus{
		Name:      result.Component,
		Installed: result.Installed,
		Version:   result.Version,
		Issues:    result.Issues,
	}
}

// NewFontTestResult creates a FontTestResult with standard test glyphs.
func NewFontTestResult(glyphsRendered, fallbackUsed bool) *FontTestResult {
	return &FontTestResult{
		GlyphsRendered: glyphsRendered,
		FallbackUsed:   fallbackUsed,
		TestGlyphs:     []string{"", "", "", ""},
	}
}

// NewColorTestResult creates a ColorTestResult with detected color mode.
func NewColorTestResult(colorMode string, gradientOK, paletteOK bool) *ColorTestResult {
	return &ColorTestResult{
		ColorMode:  colorMode,
		GradientOK: gradientOK,
		PaletteOK:  paletteOK,
	}
}

// HasErrors returns true if any errors were encountered.
func (h *HealthData) HasErrors() bool {
	return len(h.Errors) > 0
}

// AddError adds an error message to the health data.
func (h *HealthData) AddError(err string) {
	h.Errors = append(h.Errors, err)
}

// GetAllInstalled returns all installed components.
func (h *HealthData) GetAllInstalled() []*ComponentStatus {
	var installed []*ComponentStatus
	for _, status := range h.Components {
		if status.Installed {
			installed = append(installed, status)
		}
	}
	return installed
}

// GetAllMissing returns all missing components.
func (h *HealthData) GetAllMissing() []*ComponentStatus {
	var missing []*ComponentStatus
	for _, status := range h.Components {
		if !status.Installed {
			missing = append(missing, status)
		}
	}
	return missing
}

// GetHealthyCount returns the count of healthy components.
func (h *HealthData) GetHealthyCount() int {
	count := 0
	for _, status := range h.Components {
		if status.Installed && len(status.Issues) == 0 {
			count++
		}
	}
	return count
}

// GetIssueCount returns the total count of issues across all components.
func (h *HealthData) GetIssueCount() int {
	count := 0
	for _, status := range h.Components {
		count += len(status.Issues)
	}
	return count + len(h.Errors)
}

// CheckTerminalCapabilities detects terminal capabilities from environment and TerminalInfo.
// It examines COLORTERM, TERM_PROGRAM, KITTY_WINDOW_ID and other env vars to determine
// the terminal's feature support.
func CheckTerminalCapabilities(info *detector.TerminalInfo) *TerminalCapabilities {
	if info == nil {
		return &TerminalCapabilities{
			TerminalName: "unknown",
		}
	}

	caps := &TerminalCapabilities{
		TrueColor:     info.SupportsTrueColor,
		Ligatures:     info.SupportsLigatures,
		Hyperlinks:    info.SupportsHyperlinks,
		KittyGraphics: info.SupportsKittyGraphics,
		TerminalName:  info.Name,
	}

	// Additional detection from environment variables
	// COLORTERM indicates true color support
	if colorterm := os.Getenv("COLORTERM"); colorterm == "truecolor" || colorterm == "24bit" {
		caps.TrueColor = true
	}

	// TERM_PROGRAM helps identify specific terminals
	termProgram := os.Getenv("TERM_PROGRAM")
	switch termProgram {
	case "iTerm.app":
		caps.TrueColor = true
		caps.Ligatures = true
		caps.Hyperlinks = true
	case "WezTerm":
		caps.TrueColor = true
		caps.Ligatures = true
		caps.Hyperlinks = true
	case "vscode":
		caps.TrueColor = true
		caps.Ligatures = true
	}

	// Kitty-specific detection
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		caps.KittyGraphics = true
		caps.TrueColor = true
		caps.Ligatures = true
		caps.Hyperlinks = true
	}

	// Alacritty detection
	if os.Getenv("ALACRITTY_WINDOW_ID") != "" {
		caps.TrueColor = true
		caps.Ligatures = true
	}

	// Windows Terminal detection
	if os.Getenv("WT_SESSION") != "" {
		caps.TrueColor = true
	}

	return caps
}

// CheckInstalledComponents verifies all installed components using the installer's Verifier.
// It creates a Verifier with default context and resolver to check component status.
func CheckInstalledComponents() map[string]*ComponentStatus {
	ctx, err := installer.NewInstallContext()
	if err != nil {
		return map[string]*ComponentStatus{
			"error": {
				Name:      "error",
				Installed: false,
				Issues:    []string{"failed to create install context: " + err.Error()},
			},
		}
	}

	resolver := installer.NewDependencyResolver()
	verifier := installer.NewVerifier(ctx, resolver)

	// Run complete verification
	result, err := verifier.VerifyComplete(context.Background())
	if err != nil {
		return map[string]*ComponentStatus{
			"error": {
				Name:      "error",
				Installed: false,
				Issues:    []string{"verification failed: " + err.Error()},
			},
		}
	}

	// Transform results to ComponentStatus map
	components := make(map[string]*ComponentStatus)
	for name, vr := range result.Components {
		components[name] = NewComponentStatus(vr)
	}

	// Add RC file status
	if result.RCFile != nil {
		components["rc_file"] = &ComponentStatus{
			Name:      "rc_file",
			Installed: result.RCFile.Installed,
			Issues:    result.RCFile.Issues,
		}
	}

	return components
}

// Glyph represents a single test glyph for font rendering.
type Glyph struct {
	// Symbol is the Nerd Font character (if supported).
	Symbol string
	// ASCII is the ASCII fallback character(s).
	ASCII string
	// Name is a human-readable description.
	Name string
	// Category groups glyphs (file, git, status, security, etc.).
	Category string
}

// GenerateFontTest returns a slice of test glyphs for font rendering validation.
// Includes Nerd Font characters with ASCII fallbacks for terminals without Nerd Font support.
func GenerateFontTest() []Glyph {
	return []Glyph{
		// File/Folder icons
		{Symbol: "", ASCII: "[DIR]", Name: "folder", Category: "files"},
		{Symbol: "", ASCII: "[FILE]", Name: "file", Category: "files"},
		{Symbol: "", ASCII: "[PY]", Name: "python", Category: "files"},
		{Symbol: "", ASCII: "[JS]", Name: "javascript", Category: "files"},
		{Symbol: "", ASCII: "[GO]", Name: "go", Category: "files"},

		// Git icons
		{Symbol: "", ASCII: "[GIT]", Name: "git", Category: "git"},
		{Symbol: "", ASCII: "[BR]", Name: "branch", Category: "git"},
		{Symbol: "", ASCII: "[MR]", Name: "merge", Category: "git"},
		{Symbol: "", ASCII: "[+]", Name: "add", Category: "git"},
		{Symbol: "", ASCII: "[-]", Name: "remove", Category: "git"},

		// Status icons
		{Symbol: "✓", ASCII: "[OK]", Name: "check", Category: "status"},
		{Symbol: "✗", ASCII: "[X]", Name: "cross", Category: "status"},
		{Symbol: "", ASCII: "[!]", Name: "warning", Category: "status"},
		{Symbol: "", ASCII: "[i]", Name: "info", Category: "status"},
		{Symbol: "", ASCII: "[?]", Name: "question", Category: "status"},

		// Security icons
		{Symbol: "", ASCII: "[LOCK]", Name: "lock", Category: "security"},
		{Symbol: "", ASCII: "[UNLOCK]", Name: "unlock", Category: "security"},
		{Symbol: "", ASCII: "[KEY]", Name: "key", Category: "security"},
		{Symbol: "", ASCII: "[SHIELD]", Name: "shield", Category: "security"},

		// Misc icons
		{Symbol: "", ASCII: "[>]", Name: "arrow-right", Category: "misc"},
		{Symbol: "", ASCII: "[*]", Name: "star", Category: "misc"},
		{Symbol: "", ASCII: "[#]", Name: "hash", Category: "misc"},
		{Symbol: "", ASCII: "[@]", Name: "at", Category: "misc"},
	}
}

// GenerateFontTestResult creates a FontTestResult from terminal capabilities.
// Returns glyphs with ASCII fallback if Nerd Fonts are not detected.
func GenerateFontTestResult(caps *TerminalCapabilities) *FontTestResult {
	if caps == nil {
		return &FontTestResult{
			GlyphsRendered: false,
			FallbackUsed:   true,
			TestGlyphs:     []string{"[OK]", "[X]", "[DIR]", "[FILE]", "[GIT]"},
		}
	}

	// Check if terminal likely supports Nerd Fonts
	// Terminals with ligature support usually have Nerd Font rendering
	glyphsRendered := caps.Ligatures || caps.TrueColor

	glyphs := GenerateFontTest()
	testGlyphs := make([]string, len(glyphs))
	fallbackUsed := false

	if glyphsRendered {
		for i, g := range glyphs {
			testGlyphs[i] = g.Symbol
		}
	} else {
		fallbackUsed = true
		for i, g := range glyphs {
			testGlyphs[i] = g.ASCII
		}
	}

	return &FontTestResult{
		GlyphsRendered: glyphsRendered,
		FallbackUsed:   fallbackUsed,
		TestGlyphs:     testGlyphs,
	}
}

// ColorGradient represents a color test gradient.
type ColorGradient struct {
	// Mode is the color mode ("truecolor", "256", or "ansi16").
	Mode string
	// Gradient is the ANSI escape sequence for the gradient.
	Gradient string
	// Blocks is a simple block representation for terminals without color support.
	Blocks string
}

// GenerateColorTest creates color test content based on terminal capabilities.
// Returns appropriate gradient for true-color, 256-color, or ANSI-16 terminals.
func GenerateColorTest(caps *TerminalCapabilities) *ColorTestResult {
	if caps == nil {
		return &ColorTestResult{
			ColorMode:  "ansi16",
			GradientOK: false,
			PaletteOK:  false,
		}
	}

	var colorMode string
	var gradientOK bool

	switch {
	case caps.TrueColor:
		colorMode = "truecolor"
		gradientOK = true
	case has256ColorSupport():
		colorMode = "256"
		gradientOK = true
	default:
		colorMode = "ansi16"
		gradientOK = false
	}

	// Basic terminals always have ANSI 16 support
	paletteOK := true

	return &ColorTestResult{
		ColorMode:  colorMode,
		GradientOK: gradientOK,
		PaletteOK:  paletteOK,
	}
}

// has256ColorSupport checks if the terminal supports 256 colors.
func has256ColorSupport() bool {
	// Check COLORTERM for256-color support
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return true
	}

	// Check TERM for256-color support
	term := os.Getenv("TERM")
	if len(term) >= 3 {
		// Common 256-color TERM values
		if term == "xterm-256color" || term == "screen-256color" ||
			term == "tmux-256color" || term == "rxvt-unicode-256color" {
			return true
		}
	}

	// Modern terminals typically support256 colors
	termProgram := os.Getenv("TERM_PROGRAM")
	if termProgram == "iTerm.app" || termProgram == "WezTerm" || termProgram == "vscode" {
		return true
	}

	return false
}

// HealthCheckCompleteMsg is sent when health checks complete.
type HealthCheckCompleteMsg struct {
	Data *HealthData
	Err  error
}

// RunHealthCheck returns a tea.Cmd that performs all health checks asynchronously.
// It detects terminal capabilities, verifies installed components, and generates test results.
func RunHealthCheck() tea.Cmd {
	return func() tea.Msg {
		data := NewHealthData()

		// 1. Detect terminal capabilities
		td := detector.NewTerminalDetector()
		termInfo, err := td.Detect()
		if err != nil {
			data.AddError("terminal detection failed: " + err.Error())
		} else {
			data.Terminal = CheckTerminalCapabilities(termInfo)
		}

		//2. Check installed components
		data.Components = CheckInstalledComponents()

		// 3. Generate font test result
		data.FontTest = GenerateFontTestResult(data.Terminal)

		// 4. Generate color test result
		data.ColorTest = GenerateColorTest(data.Terminal)

		// 5. Set export path
		homeDir, err := os.UserHomeDir()
		if err == nil {
			data.ExportPath = homeDir + "/.config/savanhi-shell/health-report.json"
		}

		return HealthCheckCompleteMsg{
			Data: data,
			Err:  nil,
		}
	}
}

// RunHealthCheckWithDetector returns a tea.Cmd that performs health checks with a custom detector.
// This is useful for testing or when detector result is already available.
func RunHealthCheckWithDetector(detResult *detector.DetectorResult) tea.Cmd {
	return func() tea.Msg {
		data := NewHealthData()

		// Use provided detector result for terminal info
		if detResult != nil && detResult.Terminal != nil {
			data.Terminal = CheckTerminalCapabilities(detResult.Terminal)
		} else {
			// Fallback to direct detection
			td := detector.NewTerminalDetector()
			termInfo, err := td.Detect()
			if err != nil {
				data.AddError("terminal detection failed: " + err.Error())
			} else {
				data.Terminal = CheckTerminalCapabilities(termInfo)
			}
		}

		// Check installed components
		data.Components = CheckInstalledComponents()

		// Generate font test result
		data.FontTest = GenerateFontTestResult(data.Terminal)

		// Generate color test result
		data.ColorTest = GenerateColorTest(data.Terminal)

		// Set export path
		homeDir, err := os.UserHomeDir()
		if err == nil {
			data.ExportPath = homeDir + "/.config/savanhi-shell/health-report.json"
		}

		return HealthCheckCompleteMsg{
			Data: data,
			Err:  nil,
		}
	}
}

// ExportHealthReport saves the health data to a JSON file at the specified path.
// It creates the parent directories if they don't exist.
func ExportHealthReport(data *HealthData, path string) error {
	if data == nil {
		return fmt.Errorf("health data is nil")
	}

	// Create parent directories if needed
	dir := path[:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal health data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write health report: %w", err)
	}

	return nil
}
