// Package installer provides dependency installation and management.
// This file implements tool installation (zoxide, fzf, bat, eza).
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

// ToolInstaller handles installation of shell tools.
type ToolInstaller struct {
	// context is the installation context.
	context *InstallContext

	// binDir is the binary installation directory.
	binDir string
}

// NewToolInstaller creates a new tool installer.
func NewToolInstaller(ctx *InstallContext) *ToolInstaller {
	binDir := ctx.BinDir
	if binDir == "" {
		homeDir, _ := os.UserHomeDir()
		binDir = filepath.Join(homeDir, ".local", "bin")
	}

	return &ToolInstaller{
		context: ctx,
		binDir:  binDir,
	}
}

// Install installs a tool by name.
func (i *ToolInstaller) Install(ctx context.Context, toolName string, opts *Options) (*InstallResult, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Get tool definition
	tool := i.getToolDefinition(toolName)
	if tool == nil {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	result := &InstallResult{
		Component: toolName,
	}

	// Check if already installed
	if i.IsInstalled(toolName) && !i.context.Force {
		version, _ := i.GetVersion(toolName)
		result.Success = true
		result.Version = version
		result.InstalledPath, _ = i.GetPath(toolName)
		return result, nil
	}

	// Try package manager first (faster and more reliable)
	if tool.PackageManager != nil && len(tool.PackageManager) > 0 && i.context.PackageMgr != "" {
		if err := i.installViaPackageManager(ctx, toolName, tool, result); err == nil {
			result.Success = true
			return result, nil
		}
		// Fall back to binary download
	}

	// Install via binary download
	if err := i.installBinary(ctx, toolName, tool, opts, result); err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	return result, nil
}

// ToolDefinition defines a shell tool.
type ToolDefinition struct {
	// Name is the tool name.
	Name string

	// DisplayName is a human-readable name.
	DisplayName string

	// Description is what the tool does.
	Description string

	// PackageManager is the package manager package name.
	PackageManager map[string]string // package_manager -> package name

	// BinaryURL is the download URL template for binary download.
	BinaryURL string

	// VerifyCommand is the command to verify installation.
	VerifyCommand string

	// PostInstall is code to add to RC file.
	PostInstall string
}

// getToolDefinition returns the tool definition.
func (i *ToolInstaller) getToolDefinition(name string) *ToolDefinition {
	tools := map[string]*ToolDefinition{
		"zoxide": {
			Name:           "zoxide",
			DisplayName:    "Zoxide",
			Description:    "Fast directory jumping (z command)",
			PackageManager: map[string]string{"brew": "zoxide", "apt": "zoxide", "pacman": "zoxide", "dnf": "zoxide"},
			BinaryURL:      "https://github.com/ajeetdsouza/zoxide/releases/latest/download/zoxide-{os}-{arch}",
			VerifyCommand:  "zoxide --version",
			PostInstall:    `eval "$(zoxide init {shell})"`,
		},
		"fzf": {
			Name:           "fzf",
			DisplayName:    "FZF",
			Description:    "Command-line fuzzy finder",
			PackageManager: map[string]string{"brew": "fzf", "apt": "fzf", "pacman": "fzf", "dnf": "fzf"},
			BinaryURL:      "https://github.com/junegunn/fzf/releases/latest/download/fzf-{os}-{arch}",
			VerifyCommand:  "fzf --version",
			PostInstall:    `[ -f ~/.fzf.{shell} ] && source ~/.fzf.{shell}`,
		},
		"bat": {
			Name:           "bat",
			DisplayName:    "Bat",
			Description:    "Cat clone with syntax highlighting",
			PackageManager: map[string]string{"brew": "bat", "apt": "bat", "pacman": "bat", "dnf": "bat"},
			BinaryURL:      "https://github.com/sharkdp/bat/releases/latest/download/bat-{os}-{arch}.{ext}",
			VerifyCommand:  "bat --version",
			PostInstall:    "",
		},
		"eza": {
			Name:           "eza",
			DisplayName:    "Eza",
			Description:    "Modern ls replacement",
			PackageManager: map[string]string{"brew": "eza", "apt": "eza", "pacman": "eza", "dnf": "eza"},
			BinaryURL:      "https://github.com/eza-community/eza/releases/latest/download/eza_{os}_{arch}",
			VerifyCommand:  "eza --version",
			PostInstall:    "",
		},
	}

	return tools[name]
}

// installViaPackageManager installs a tool using the system package manager.
func (i *ToolInstaller) installViaPackageManager(ctx context.Context, toolName string, tool *ToolDefinition, result *InstallResult) error {
	packageName, ok := tool.PackageManager[i.context.PackageMgr]
	if !ok {
		return fmt.Errorf("no package mapping for %s on %s", toolName, i.context.PackageMgr)
	}

	var cmd *exec.Cmd
	switch i.context.PackageMgr {
	case "brew":
		cmd = exec.CommandContext(ctx, "brew", "install", packageName)
	case "apt", "apt-get":
		cmd = exec.CommandContext(ctx, "sudo", "apt-get", "install", "-y", packageName)
	case "pacman":
		cmd = exec.CommandContext(ctx, "sudo", "pacman", "-S", "--noconfirm", packageName)
	case "dnf", "yum":
		cmd = exec.CommandContext(ctx, "sudo", i.context.PackageMgr, "install", "-y", packageName)
	case "apk":
		cmd = exec.CommandContext(ctx, "apk", "add", packageName)
	default:
		return fmt.Errorf("unsupported package manager: %s", i.context.PackageMgr)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("package installation failed: %w\n%s", err, string(output))
	}

	// Find installed path
	if path, err := exec.LookPath(toolName); err == nil {
		result.InstalledPath = path
	}

	return nil
}

// installBinary installs a tool via binary download.
func (i *ToolInstaller) installBinary(ctx context.Context, toolName string, tool *ToolDefinition, opts *Options, result *InstallResult) error {
	// Ensure bin directory exists
	if err := os.MkdirAll(i.binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Build download URL
	url := tool.BinaryURL
	url = strings.ReplaceAll(url, "{os}", runtime.GOOS)
	url = strings.ReplaceAll(url, "{arch}", runtime.GOARCH)
	url = strings.ReplaceAll(url, "{shell}", i.context.Shell)
	url = strings.ReplaceAll(url, "{ext}", i.getArchiveExtension())

	// Download
	cacheDir := filepath.Join(i.context.CacheDir, "tools")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	installer := &DefaultInstaller{context: i.context}
	downloadResult, err := installer.downloadFile(ctx, url, cacheDir, opts)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", toolName, err)
	}

	// Determine target path
	targetPath := filepath.Join(i.binDir, toolName)
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	// Handle archive if needed
	if strings.HasSuffix(downloadResult.LocalPath, ".tar.gz") || strings.HasSuffix(downloadResult.LocalPath, ".tgz") {
		return i.installFromTarGz(downloadResult.LocalPath, toolName, targetPath, result)
	} else if strings.HasSuffix(downloadResult.LocalPath, ".zip") {
		return i.installFromZip(downloadResult.LocalPath, toolName, targetPath, result)
	} else if strings.HasSuffix(downloadResult.LocalPath, ".tar.xz") {
		return i.installFromTarXz(downloadResult.LocalPath, toolName, targetPath, result)
	}

	// Single binary
	if err := copyFile(downloadResult.LocalPath, targetPath); err != nil {
		return fmt.Errorf("failed to install %s: %w", toolName, err)
	}

	// Make executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to make %s executable: %w", toolName, err)
	}

	result.InstalledPath = targetPath
	return nil
}

// getArchiveExtension returns the archive extension for the current platform.
func (i *ToolInstaller) getArchiveExtension() string {
	switch runtime.GOOS {
	case "windows":
		return "zip"
	case "darwin":
		return "tar.gz"
	default:
		return "tar.gz"
	}
}

// installFromTarGz installs from a tar.gz archive.
func (i *ToolInstaller) installFromTarGz(archivePath, toolName, targetPath string, result *InstallResult) error {
	// Use tar command (simpler than implementing tar parsing)
	cmd := exec.Command("tar", "-xzf", archivePath, "-C", i.binDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to extract archive: %w\n%s", err, string(output))
	}

	// Find the extracted binary
	extractedPath := filepath.Join(i.binDir, toolName)
	if _, err := os.Stat(extractedPath); err != nil {
		// Try to find it
		matches, _ := filepath.Glob(filepath.Join(i.binDir, "*"))
		for _, match := range matches {
			if strings.Contains(match, toolName) && !strings.HasSuffix(match, ".tar.gz") {
				extractedPath = match
				break
			}
		}
	}

	// Move to target path if needed
	if extractedPath != targetPath {
		if err := os.Rename(extractedPath, targetPath); err != nil {
			return fmt.Errorf("failed to move binary: %w", err)
		}
	}

	// Make executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	result.InstalledPath = targetPath
	return nil
}

// installFromZip installs from a zip archive.
func (i *ToolInstaller) installFromZip(archivePath, toolName, targetPath string, result *InstallResult) error {
	// Use unzip command or Go archive/zip
	cmd := exec.Command("unzip", "-o", archivePath, "-d", i.binDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to extract archive: %w\n%s", err, string(output))
	}

	// Find the extracted binary
	extractedPath := filepath.Join(i.binDir, toolName)
	if runtime.GOOS == "windows" {
		extractedPath += ".exe"
	}

	// Move to target path if needed
	if extractedPath != targetPath {
		if err := os.Rename(extractedPath, targetPath); err != nil {
			return fmt.Errorf("failed to move binary: %w", err)
		}
	}

	// Make executable (Unix only)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	result.InstalledPath = targetPath
	return nil
}

// installFromTarXz installs from a tar.xz archive.
func (i *ToolInstaller) installFromTarXz(archivePath, toolName, targetPath string, result *InstallResult) error {
	cmd := exec.Command("tar", "-xJf", archivePath, "-C", i.binDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to extract archive: %w\n%s", err, string(output))
	}

	// Similar to tar.gz handling
	extractedPath := filepath.Join(i.binDir, toolName)
	if _, err := os.Stat(extractedPath); err != nil {
		matches, _ := filepath.Glob(filepath.Join(i.binDir, "*"))
		for _, match := range matches {
			if strings.Contains(match, toolName) {
				extractedPath = match
				break
			}
		}
	}

	if extractedPath != targetPath {
		if err := os.Rename(extractedPath, targetPath); err != nil {
			return fmt.Errorf("failed to move binary: %w", err)
		}
	}

	if err := os.Chmod(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	result.InstalledPath = targetPath
	return nil
}

// IsInstalled checks if a tool is installed.
func (i *ToolInstaller) IsInstalled(toolName string) bool {
	// Check PATH
	if _, err := exec.LookPath(toolName); err == nil {
		return true
	}

	// Check bin directory
	targetPath := filepath.Join(i.binDir, toolName)
	if _, err := os.Stat(targetPath); err == nil {
		return true
	}

	return false
}

// GetVersion returns the installed version.
func (i *ToolInstaller) GetVersion(toolName string) (string, error) {
	tool := i.getToolDefinition(toolName)
	if tool == nil || tool.VerifyCommand == "" {
		return "", fmt.Errorf("tool not found or no verify command")
	}

	cmdParts := strings.Fields(tool.VerifyCommand)
	if len(cmdParts) == 0 {
		return "", fmt.Errorf("invalid verify command")
	}

	// Use installed path if not in PATH
	if _, err := exec.LookPath(cmdParts[0]); err != nil {
		cmdParts[0] = filepath.Join(i.binDir, cmdParts[0])
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetPath returns the path to a tool.
func (i *ToolInstaller) GetPath(toolName string) (string, error) {
	// Check PATH first
	if path, err := exec.LookPath(toolName); err == nil {
		return path, nil
	}

	// Check bin directory
	targetPath := filepath.Join(i.binDir, toolName)
	if _, err := os.Stat(targetPath); err == nil {
		return targetPath, nil
	}

	return "", fmt.Errorf("tool %s not found", toolName)
}

// Uninstall removes a tool.
func (i *ToolInstaller) Uninstall(toolName string) error {
	// Try package manager first
	tool := i.getToolDefinition(toolName)
	if tool != nil && tool.PackageManager[i.context.PackageMgr] != "" {
		if i.uninstallViaPackageManager(toolName, tool) == nil {
			return nil
		}
	}

	// Remove binary
	targetPath := filepath.Join(i.binDir, toolName)
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", toolName, err)
	}

	return nil
}

// uninstallViaPackageManager removes a tool via package manager.
func (i *ToolInstaller) uninstallViaPackageManager(toolName string, tool *ToolDefinition) error {
	var cmd *exec.Cmd
	switch i.context.PackageMgr {
	case "brew":
		cmd = exec.Command("brew", "uninstall", tool.PackageManager["brew"])
	case "apt", "apt-get":
		cmd = exec.Command("sudo", "apt-get", "remove", "-y", tool.PackageManager["apt"])
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-R", "--noconfirm", tool.PackageManager["pacman"])
	case "dnf", "yum":
		cmd = exec.Command("sudo", i.context.PackageMgr, "remove", "-y", tool.PackageManager["dnf"])
	default:
		return fmt.Errorf("unsupported package manager: %s", i.context.PackageMgr)
	}

	return cmd.Run()
}

// GetPostInstallRC returns the RC file content to add for a tool.
func (i *ToolInstaller) GetPostInstallRC(toolName, shell string) string {
	tool := i.getToolDefinition(toolName)
	if tool == nil || tool.PostInstall == "" {
		return ""
	}

	content := tool.PostInstall
	content = strings.ReplaceAll(content, "{shell}", shell)
	return content
}

// ListSupportedTools returns all supported tools.
func (i *ToolInstaller) ListSupportedTools() []*ToolDefinition {
	return []*ToolDefinition{
		i.getToolDefinition("zoxide"),
		i.getToolDefinition("fzf"),
		i.getToolDefinition("bat"),
		i.getToolDefinition("eza"),
	}
}
