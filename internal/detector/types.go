// Package detector provides system detection capabilities.
// This file contains type definitions for all detection results.
package detector

import "time"

// OSType represents the operating system type.
type OSType string

const (
	// OSTypeMacOS represents macOS.
	OSTypeMacOS OSType = "macos"
	// OSTypeLinux represents Linux.
	OSTypeLinux OSType = "linux"
	// OSTypeWSL represents Windows Subsystem for Linux.
	OSTypeWSL OSType = "wsl"
	// OSTypeWindows represents native Windows.
	OSTypeWindows OSType = "windows"
	// OSTypeTermux represents Android Termux.
	OSTypeTermux OSType = "termux"
	// OSTypeUnknown represents an unknown OS.
	OSTypeUnknown OSType = "unknown"
)

// ShellType represents the shell type.
type ShellType string

const (
	// ShellTypeZsh represents zsh.
	ShellTypeZsh ShellType = "zsh"
	// ShellTypeBash represents bash.
	ShellTypeBash ShellType = "bash"
	// ShellTypeFish represents fish.
	ShellTypeFish ShellType = "fish"
	// ShellTypePwsh represents PowerShell.
	ShellTypePwsh ShellType = "pwsh"
	// ShellTypeUnknown represents an unknown shell.
	ShellTypeUnknown ShellType = "unknown"
)

// TerminalType represents the terminal emulator type.
type TerminalType string

const (
	// TerminalTypeITerm2 represents iTerm2 on macOS.
	TerminalTypeITerm2 TerminalType = "iterm2"
	// TerminalTypeAlacritty represents Alacritty terminal.
	TerminalTypeAlacritty TerminalType = "alacritty"
	// TerminalTypeKitty represents Kitty terminal.
	TerminalTypeKitty TerminalType = "kitty"
	// TerminalTypeWezTerm represents WezTerm terminal.
	TerminalTypeWezTerm TerminalType = "wezterm"
	// TerminalTypeWindowsTerminal represents Windows Terminal.
	TerminalTypeWindowsTerminal TerminalType = "windows-terminal"
	// TerminalTypeGNOMETerminal represents GNOME Terminal.
	TerminalTypeGNOMETerminal TerminalType = "gnome-terminal"
	// TerminalTypeKonsole represents KDE Konsole.
	TerminalTypeKonsole TerminalType = "konsole"
	// TerminalTypeVSCode represents VS Code integrated terminal.
	TerminalTypeVSCode TerminalType = "vscode"
	// TerminalTypeFoot represents Foot terminal (Wayland).
	TerminalTypeFoot TerminalType = "foot"
	// TerminalTypeUnknown represents an unknown terminal.
	TerminalTypeUnknown TerminalType = "unknown"
)

// OSInfo contains information about the operating system.
type OSInfo struct {
	// Type is the OS type (macOS, Linux, WSL, etc.).
	Type OSType `json:"type"`

	// Distro is the Linux distribution name (ubuntu, arch, fedora, etc.).
	// Empty for non-Linux systems.
	Distro string `json:"distro,omitempty"`

	// Version is the OS version string.
	Version string `json:"version"`

	// Arch is the CPU architecture (arm64, amd64, etc.).
	Arch string `json:"arch"`

	// PackageMgr is the primary package manager (brew, apt, pacman, dnf, etc.).
	PackageMgr string `json:"package_manager"`

	// PrettyName is a human-readable OS name.
	PrettyName string `json:"pretty_name"`

	// Codename is the distribution codename (e.g., "jammy" for Ubuntu 22.04).
	Codename string `json:"codename,omitempty"`
}

// ShellInfo contains information about a shell.
type ShellInfo struct {
	// Name is the shell name (zsh, bash, fish, pwsh).
	Name ShellType `json:"name"`

	// Version is the shell version string.
	Version string `json:"version"`

	// RCFile is the path to the shell's RC file (~/.zshrc, ~/.bashrc, etc.).
	RCFile string `json:"rc_file"`

	// RCFileExists indicates whether the RC file exists.
	RCFileExists bool `json:"rc_file_exists"`

	// ConfigDir is the path to the shell's config directory.
	ConfigDir string `json:"config_dir"`

	// IsDefault indicates whether this is the default login shell.
	IsDefault bool `json:"is_default"`

	// Path is the path to the shell executable.
	Path string `json:"path"`

	// AvailableShells is the list of available shells from /etc/shells.
	AvailableShells []string `json:"available_shells,omitempty"`
}

// TerminalInfo contains information about the terminal emulator.
type TerminalInfo struct {
	// Type is the terminal emulator type.
	Type TerminalType `json:"type"`

	// Name is the terminal emulator name.
	Name string `json:"name"`

	// Version is the terminal version string.
	Version string `json:"version,omitempty"`

	// SupportsTrueColor indicates 24-bit true color support.
	SupportsTrueColor bool `json:"supports_true_color"`

	// SupportsLigatures indicates font ligature support.
	SupportsLigatures bool `json:"supports_ligatures"`

	// SupportsHyperlinks indicates OSC 8 hyperlink support.
	SupportsHyperlinks bool `json:"supports_hyperlinks"`

	// SupportsKittyGraphics indicates Kitty graphics protocol support.
	SupportsKittyGraphics bool `json:"supports_kitty_graphics"`

	// FontFamily is the detected font family name.
	FontFamily string `json:"font_family,omitempty"`

	// FontSize is the detected font size in points.
	FontSize int `json:"font_size,omitempty"`
}

// FontInfo contains information about a single font.
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

// FontInventory contains information about installed fonts.
type FontInventory struct {
	// Fonts is the list of detected fonts.
	Fonts []FontInfo `json:"fonts"`

	// NerdFonts is the list of detected Nerd Fonts.
	NerdFonts []FontInfo `json:"nerd_fonts"`

	// HasNerdFonts indicates whether any Nerd Fonts are installed.
	HasNerdFonts bool `json:"has_nerd_fonts"`

	// RecommendedFonts is the list of recommended but not installed fonts.
	RecommendedFonts []string `json:"recommended_fonts,omitempty"`
}

// ConfigSnapshot contains information about existing configurations.
type ConfigSnapshot struct {
	// HasOhMyPosh indicates whether oh-my-posh is configured.
	HasOhMyPosh bool `json:"has_oh_my_posh"`

	// OhMyPoshConfigPath is the path to the oh-my-posh config.
	OhMyPoshConfigPath string `json:"oh_my_posh_config_path,omitempty"`

	// HasStarship indicates whether starship is configured.
	HasStarship bool `json:"has_starship"`

	// StarshipConfigPath is the path to the starship config.
	StarshipConfigPath string `json:"starship_config_path,omitempty"`

	// HasZoxide indicates whether zoxide is installed.
	HasZoxide bool `json:"has_zoxide"`

	// HasFzf indicates whether fzf is installed.
	HasFzf bool `json:"has_fzf"`

	// HasBat indicates whether bat is installed.
	HasBat bool `json:"has_bat"`

	// HasEza indicates whether eza is installed.
	HasEza bool `json:"has_eza"`

	// HasSavanhiMarkers indicates whether Savanhi markers exist in RC files.
	HasSavanhiMarkers bool `json:"has_savanhi_markers"`

	// SavanhiMarkerContent contains the content between Savanhi markers (if found).
	SavanhiMarkerContent string `json:"savanhi_marker_content,omitempty"`

	// DetectedTheme is the detected theme name (if any).
	DetectedTheme string `json:"detected_theme,omitempty"`

	// DetectedFont is the detected font family (if any).
	DetectedFont string `json:"detected_font,omitempty"`

	// DetectedColorScheme is the detected color scheme (if any).
	DetectedColorScheme string `json:"detected_color_scheme,omitempty"`

	// DetectedAt is when this snapshot was taken.
	DetectedAt time.Time `json:"detected_at"`
}

// OS-specific detector interface.
type OSDetector interface {
	Detect() (*OSInfo, error)
}

// Shell-specific detector interface.
type ShellDetector interface {
	Detect() (*ShellInfo, error)
}

// Terminal-specific detector interface.
type TerminalDetector interface {
	Detect() (*TerminalInfo, error)
}

// Font-specific detector interface.
type FontDetector interface {
	Detect() (*FontInventory, error)
}

// Config-specific detector interface.
type ConfigDetector interface {
	Detect() (*ConfigSnapshot, error)
}
