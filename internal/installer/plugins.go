// Package installer provides dependency installation and management for Savanhi Shell.
// This file contains zsh plugin detection and management functionality.
package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/savanhi/shell/pkg/shell"
)

// PluginDetector handles detection of zsh plugin installation status.
type PluginDetector struct {
	// ctx is the installation context.
	ctx *InstallContext

	// shell is the shell interface for detection.
	shell shell.Shell
}

// NewPluginDetector creates a new plugin detector.
func NewPluginDetector(ctx *InstallContext, s shell.Shell) *PluginDetector {
	return &PluginDetector{
		ctx:   ctx,
		shell: s,
	}
}

// Detect detects the installation status of a single plugin.
func (d *PluginDetector) Detect(plugin Plugin) (*PluginStatus, error) {
	status := &PluginStatus{
		Plugin:    plugin,
		Installed: false,
		Method:    MethodNone,
	}

	// Try detection methods in order of preference
	// 1. Oh My Zsh
	if omzStatus := d.detectOhMyZsh(plugin); omzStatus != nil {
		return omzStatus, nil
	}

	// 2. Homebrew
	if brewStatus := d.detectHomebrew(plugin); brewStatus != nil {
		return brewStatus, nil
	}

	// 3. Git clone
	if gitStatus := d.detectGitClone(plugin); gitStatus != nil {
		return gitStatus, nil
	}

	// 4. Manual source line in .zshrc
	if manualStatus := d.detectManualSource(plugin); manualStatus != nil {
		return manualStatus, nil
	}

	return status, nil
}

// DetectAll detects the installation status of all supported plugins.
func (d *PluginDetector) DetectAll() ([]PluginStatus, error) {
	plugins := GetSupportedPlugins()
	statuses := make([]PluginStatus, len(plugins))

	for i, plugin := range plugins {
		status, err := d.Detect(plugin)
		if err != nil {
			return nil, fmt.Errorf("failed to detect plugin %s: %w", plugin.Name, err)
		}
		statuses[i] = *status
	}

	return statuses, nil
}

// detectOhMyZsh checks if a plugin is installed via Oh My Zsh.
func (d *PluginDetector) detectOhMyZsh(plugin Plugin) *PluginStatus {
	// Check if Oh My Zsh is installed
	zshShell, ok := d.shell.(*shell.ZshShell)
	if !ok {
		return nil
	}

	hasOMZ, zshCustom := zshShell.HasOhMyZsh()
	if !hasOMZ {
		return nil
	}

	// Determine the OMZ custom plugins directory
	var pluginDir string
	if zshCustom != "" {
		pluginDir = filepath.Join(zshCustom, "plugins", plugin.OhMyZshName)
	} else {
		pluginDir = filepath.Join(d.ctx.HomeDir, ".oh-my-zsh", "custom", "plugins", plugin.OhMyZshName)
	}

	// Check if the plugin directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return nil
	}

	// Check if it's in the plugins array in .zshrc
	inPluginsArray := d.isInPluginsArray(plugin.OhMyZshName)

	return &PluginStatus{
		Plugin:      plugin,
		Installed:   inPluginsArray,
		Method:      MethodOhMyZsh,
		InstallPath: pluginDir,
	}
}

// detectHomebrew checks if a plugin is installed via Homebrew.
func (d *PluginDetector) detectHomebrew(plugin Plugin) *PluginStatus {
	// Check if Homebrew is available
	if _, err := exec.LookPath("brew"); err != nil {
		return nil
	}

	// Check if the package is installed
	if plugin.BrewPackage == "" {
		return nil
	}

	cmd := exec.Command("brew", "--prefix", plugin.BrewPackage)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	installPath := strings.TrimSpace(string(output))
	if installPath == "" {
		return nil
	}

	// Verify the installation path actually exists
	// brew --prefix may return a path even for uninstalled packages
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return nil
	}

	// Verify the source file exists in the installation path
	sourcePath := filepath.Join(installPath, plugin.SourceFile)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		// Try alternate locations
		altPaths := []string{
			filepath.Join(installPath, plugin.Name+".zsh"),
			filepath.Join(installPath, "share", plugin.Name+".zsh"),
		}
		found := false
		for _, altPath := range altPaths {
			if _, err := os.Stat(altPath); err == nil {
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	// Verify the plugin is sourced in .zshrc
	sourcedInRC := d.isSourcedInRC(plugin, installPath)

	return &PluginStatus{
		Plugin:      plugin,
		Installed:   sourcedInRC,
		Method:      MethodHomebrew,
		InstallPath: installPath,
	}
}

// detectGitClone checks if a plugin is installed via git clone.
func (d *PluginDetector) detectGitClone(plugin Plugin) *PluginStatus {
	// Check common git clone locations
	possiblePaths := []string{
		filepath.Join(d.ctx.HomeDir, ".zsh", plugin.Name),
		filepath.Join(d.ctx.HomeDir, ".local", "share", "zsh", "plugins", plugin.Name),
		filepath.Join(d.ctx.HomeDir, ".config", "zsh", "plugins", plugin.Name),
	}

	for _, pluginPath := range possiblePaths {
		if _, err := os.Stat(pluginPath); err == nil {
			// Check if it's a git repository
			gitDir := filepath.Join(pluginPath, ".git")
			if _, err := os.Stat(gitDir); err == nil {
				// Verify the plugin is sourced in .zshrc
				sourcedInRC := d.isSourcedInRC(plugin, pluginPath)

				return &PluginStatus{
					Plugin:      plugin,
					Installed:   sourcedInRC,
					Method:      MethodGitClone,
					InstallPath: pluginPath,
				}
			}
		}
	}

	return nil
}

// detectManualSource checks if a plugin is manually sourced in .zshrc.
func (d *PluginDetector) detectManualSource(plugin Plugin) *PluginStatus {
	rcPath, err := d.shell.GetRCPath()
	if err != nil {
		return nil
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		return nil
	}

	rcContent := string(content)

	// Look for source lines that reference the plugin
	// Common patterns:
	// source /path/to/plugin.zsh
	// . /path/to/plugin.zsh
	// source ~/.zsh/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh
	searchPatterns := []string{
		fmt.Sprintf("source ~/.zsh/%s/%s", plugin.Name, plugin.SourceFile),
		fmt.Sprintf(". ~/.zsh/%s/%s", plugin.Name, plugin.SourceFile),
		fmt.Sprintf("source ~/.local/share/zsh/plugins/%s/%s", plugin.Name, plugin.SourceFile),
		fmt.Sprintf("source ~/.config/zsh/plugins/%s/%s", plugin.Name, plugin.SourceFile),
	}

	for _, pattern := range searchPatterns {
		if strings.Contains(rcContent, pattern) {
			return &PluginStatus{
				Plugin:      plugin,
				Installed:   true,
				Method:      MethodGitClone,
				InstallPath: "", // Unknown - user custom path
			}
		}
	}

	return nil
}

// isInPluginsArray checks if a plugin is in the Oh My Zsh plugins array.
func (d *PluginDetector) isInPluginsArray(pluginName string) bool {
	rcPath, err := d.shell.GetRCPath()
	if err != nil {
		return false
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		return false
	}

	rcContent := string(content)

	// Look for plugins=(... plugin ...)
	// This is a simplified check - ParsePluginsArray in zsh.go does more thorough parsing
	return strings.Contains(rcContent, "plugins=(") &&
		strings.Contains(rcContent, pluginName)
}

// isSourcedInRC checks if a plugin is sourced in .zshrc.
func (d *PluginDetector) isSourcedInRC(plugin Plugin, installPath string) bool {
	rcPath, err := d.shell.GetRCPath()
	if err != nil {
		return false
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		return false
	}

	rcContent := string(content)

	// Check for source line to the plugin
	sourceFile := filepath.Join(installPath, plugin.SourceFile)
	sourcePatterns := []string{
		fmt.Sprintf("source %s", sourceFile),
		fmt.Sprintf(". %s", sourceFile),
	}

	for _, pattern := range sourcePatterns {
		if strings.Contains(rcContent, pattern) {
			return true
		}
	}

	// Check for Savanhi section markers
	marker := fmt.Sprintf("savanhi-%s", strings.TrimPrefix(plugin.Name, "zsh-"))
	startMarker := fmt.Sprintf("# >>> %s >>>", marker)
	if strings.Contains(rcContent, startMarker) {
		return true
	}

	return false
}

// DetectPluginManagers checks for installed plugin managers (antigen, zinit, zplug).
// Returns a list of detected plugin managers that might conflict with direct installation.
func (d *PluginDetector) DetectPluginManagers() []string {
	rcPath, err := d.shell.GetRCPath()
	if err != nil {
		return nil
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		return nil
	}

	rcContent := string(content)
	var managers []string

	// Check for antigen
	if strings.Contains(rcContent, "antigen") || strings.Contains(rcContent, "antibundle") {
		managers = append(managers, "antigen")
	}

	// Check for zinit
	if strings.Contains(rcContent, "zinit") || strings.Contains(rcContent, "zplugin") {
		managers = append(managers, "zinit")
	}

	// Check for zplug
	if strings.Contains(rcContent, "zplug") {
		managers = append(managers, "zplug")
	}

	return managers
}

// PluginInstaller handles zsh plugin installation.
type PluginInstaller struct {
	// ctx is the installation context.
	ctx *InstallContext

	// shell is the shell interface.
	shell shell.Shell

	// detector is the plugin detector.
	detector *PluginDetector
}

// NewPluginInstaller creates a new plugin installer.
func NewPluginInstaller(ctx *InstallContext, s shell.Shell) *PluginInstaller {
	return &PluginInstaller{
		ctx:      ctx,
		shell:    s,
		detector: NewPluginDetector(ctx, s),
	}
}

// Detect is a convenience method that calls the detector.
func (p *PluginInstaller) Detect(plugin Plugin) (*PluginStatus, error) {
	return p.detector.Detect(plugin)
}

// DetectAll is a convenience method that calls the detector.
func (p *PluginInstaller) DetectAll() ([]PluginStatus, error) {
	return p.detector.DetectAll()
}

// Install installs a plugin using the specified method.
// If method is MethodNone, it will auto-select the best method based on detection.
func (p *PluginInstaller) Install(ctx context.Context, plugin Plugin, method InstallMethod) (*PluginInstallResult, error) {
	result := &PluginInstallResult{
		Plugin:     plugin,
		Success:    false,
		Method:     method,
		RCModified: false,
		Warnings:   []string{},
	}

	// Validate zsh version compatibility
	if plugin.MinZshVersion != "" {
		zshShell, ok := p.shell.(*shell.ZshShell)
		if !ok {
			return nil, fmt.Errorf("plugin installation requires zsh shell")
		}

		compatible, err := zshShell.IsZshVersionCompatible(plugin.MinZshVersion)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("could not verify zsh version: %v", err))
		} else if !compatible {
			return nil, fmt.Errorf("zsh version incompatible: plugin %s requires zsh %s or later", plugin.Name, plugin.MinZshVersion)
		}
	}

	// Check if plugin is a zsh plugin (require ZshShell)
	zshShell, ok := p.shell.(*shell.ZshShell)
	if !ok {
		return nil, fmt.Errorf("plugin installation requires zsh shell")
	}

	// Auto-select method if not specified
	if method == MethodNone {
		method = p.selectInstallMethod(zshShell, plugin)
	}

	result.Method = method

	// Track installation steps for rollback
	var installSteps []string
	var cleanupNeeded = true

	// Defer rollback on failure
	defer func() {
		if !result.Success && cleanupNeeded {
			p.rollbackInstallation(installSteps, method, plugin)
		}
	}()

	// Install based on method
	var installPath string
	var err error

	switch method {
	case MethodOhMyZsh:
		installPath, err = p.installOhMyZsh(ctx, zshShell, plugin)
		installSteps = append(installSteps, "clone-plugin", "add-to-plugins-array")
	case MethodHomebrew:
		installPath, err = p.installHomebrew(ctx, plugin)
		installSteps = append(installSteps, "brew-install", "inject-source-line")
	case MethodGitClone:
		installPath, err = p.installGitClone(ctx, plugin)
		installSteps = append(installSteps, "clone-plugin", "inject-source-line")
	default:
		return nil, fmt.Errorf("unsupported installation method: %v", method)
	}

	if err != nil {
		result.Error = err
		return result, fmt.Errorf("installation failed: %w", err)
	}

	result.InstallPath = installPath
	result.Success = true

	// Ensure correct sourcing order (syntax-highlighting last)
	if err := p.ensureCorrectOrder(plugin, zshShell); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("failed to ensure correct order: %v", err))
	}

	result.RCModified = true
	cleanupNeeded = false

	return result, nil
}

// InstallAll installs multiple plugins in the correct order.
// Plugins with MustBeLast=true (like zsh-syntax-highlighting) are installed last.
func (p *PluginInstaller) InstallAll(ctx context.Context, plugins []Plugin) ([]PluginInstallResult, error) {
	// Separate plugins into regular and must-be-last
	var regularPlugins []Plugin
	var lastPlugins []Plugin

	for _, plugin := range plugins {
		if plugin.MustBeLast {
			lastPlugins = append(lastPlugins, plugin)
		} else {
			regularPlugins = append(regularPlugins, plugin)
		}
	}

	// Sort must-be-last plugins to ensure syntax-highlighting is truly last
	// (currently only one, but future-proof)
	sort.Slice(lastPlugins, func(i, j int) bool {
		return lastPlugins[i].Name < lastPlugins[j].Name
	})

	// Install regular plugins first
	results := make([]PluginInstallResult, 0, len(plugins))

	for _, plugin := range regularPlugins {
		result, err := p.Install(ctx, plugin, MethodNone)
		if err != nil {
			return results, fmt.Errorf("failed to install %s: %w", plugin.Name, err)
		}
		results = append(results, *result)
	}

	// Install must-be-last plugins after all others
	for _, plugin := range lastPlugins {
		result, err := p.Install(ctx, plugin, MethodNone)
		if err != nil {
			return results, fmt.Errorf("failed to install %s: %w", plugin.Name, err)
		}
		results = append(results, *result)
	}

	return results, nil
}

// Uninstall uninstalls a plugin.
func (p *PluginInstaller) Uninstall(pluginName string) error {
	plugin := GetPluginByName(pluginName)
	if plugin == nil {
		return fmt.Errorf("unknown plugin: %s", pluginName)
	}

	// Detect how it was installed
	status, err := p.Detect(*plugin)
	if err != nil {
		return fmt.Errorf("failed to detect plugin: %w", err)
	}

	if !status.Installed {
		// Plugin not installed, nothing to do
		return nil
	}

	// Get zsh shell
	zshShell, ok := p.shell.(*shell.ZshShell)
	if !ok {
		return fmt.Errorf("plugin uninstall requires zsh shell")
	}

	// Remove based on installation method
	switch status.Method {
	case MethodOhMyZsh:
		return p.uninstallOhMyZsh(*plugin, zshShell, status.InstallPath)
	case MethodHomebrew:
		return p.uninstallHomebrew(*plugin)
	case MethodGitClone:
		return p.uninstallGitClone(*plugin, status.InstallPath)
	default:
		return fmt.Errorf("unknown installation method for uninstall")
	}
}

// selectInstallMethod selects the best installation method based on environment.
func (p *PluginInstaller) selectInstallMethod(zshShell *shell.ZshShell, plugin Plugin) InstallMethod {
	// Priority: Oh My Zsh > Homebrew > Git Clone

	// Check if Oh My Zsh is installed
	hasOMZ, _ := zshShell.HasOhMyZsh()
	if hasOMZ {
		return MethodOhMyZsh
	}

	// Check if Homebrew is available
	if _, err := exec.LookPath("brew"); err == nil {
		if plugin.BrewPackage != "" {
			return MethodHomebrew
		}
	}

	// Fallback to git clone
	return MethodGitClone
}

// installOhMyZsh installs a plugin via Oh My Zsh.
func (p *PluginInstaller) installOhMyZsh(ctx context.Context, zshShell *shell.ZshShell, plugin Plugin) (string, error) {
	// Get OMZ custom plugins directory
	pluginDir := zshShell.GetOhMyZshPluginDir()
	targetDir := filepath.Join(pluginDir, plugin.OhMyZshName)

	// Check if plugin directory already exists
	if _, err := os.Stat(targetDir); err == nil {
		// Directory exists, check if it's a valid repo
		gitDir := filepath.Join(targetDir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			// Update existing repo
			if err := p.updateGitRepo(ctx, targetDir); err != nil {
				return "", fmt.Errorf("failed to update plugin: %w", err)
			}
		}
	} else {
		// Clone the plugin
		if err := p.cloneRepo(ctx, plugin.Repository, targetDir); err != nil {
			return "", fmt.Errorf("failed to clone plugin: %w", err)
		}
	}

	// Add to plugins array in .zshrc
	if err := zshShell.AddToPluginsArray(plugin.OhMyZshName); err != nil {
		return "", fmt.Errorf("failed to add to plugins array: %w", err)
	}

	return targetDir, nil
}

// installHomebrew installs a plugin via Homebrew.
func (p *PluginInstaller) installHomebrew(ctx context.Context, plugin Plugin) (string, error) {
	// Check if brew is available
	if _, err := exec.LookPath("brew"); err != nil {
		return "", fmt.Errorf("homebrew not found")
	}

	// Install via brew
	cmd := exec.CommandContext(ctx, "brew", "install", plugin.BrewPackage)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("brew install failed: %w\n%s", err, string(output))
	}

	// Get install path using brew --prefix
	cmd = exec.CommandContext(ctx, "brew", "--prefix", plugin.BrewPackage)
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get brew prefix: %w", err)
	}

	installPath := strings.TrimSpace(string(output))

	// Create RC modifier and inject source line
	rcModifier := NewRCModifier(p.shell, p.ctx.ConfigDir)

	// Determine source file path
	// Homebrew installs to <prefix>/share/<plugin>/<plugin>.zsh or similar
	sourcePath := filepath.Join(installPath, plugin.SourceFile)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		// Try alternate locations
		altPaths := []string{
			filepath.Join(installPath, "share", plugin.Name, plugin.SourceFile),
			filepath.Join(installPath, plugin.Name+".zsh"),
		}
		for _, altPath := range altPaths {
			if _, err := os.Stat(altPath); err == nil {
				sourcePath = altPath
				break
			}
		}
	}

	if err := rcModifier.InjectZshPlugin(plugin, sourcePath); err != nil {
		return "", fmt.Errorf("failed to inject source line: %w", err)
	}

	return installPath, nil
}

// installGitClone installs a plugin via git clone.
func (p *PluginInstaller) installGitClone(ctx context.Context, plugin Plugin) (string, error) {
	// Determine target directory
	pluginDir := filepath.Join(p.ctx.HomeDir, ".zsh", plugin.Name)

	// Check if directory already exists
	if _, err := os.Stat(pluginDir); err == nil {
		gitDir := filepath.Join(pluginDir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			// Update existing repo
			if err := p.updateGitRepo(ctx, pluginDir); err != nil {
				return "", fmt.Errorf("failed to update plugin: %w", err)
			}
		} else {
			// Directory exists but not a git repo, backup and re-clone
			backupDir := pluginDir + ".backup"
			if err := os.Rename(pluginDir, backupDir); err != nil {
				return "", fmt.Errorf("failed to backup existing directory: %w", err)
			}
		}
	}

	// Create plugin directory
	if err := os.MkdirAll(filepath.Dir(pluginDir), 0755); err != nil {
		return "", fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Clone the plugin
	if err := p.cloneRepo(ctx, plugin.Repository, pluginDir); err != nil {
		return "", fmt.Errorf("failed to clone plugin: %w", err)
	}

	// Create RC modifier and inject source line
	rcModifier := NewRCModifier(p.shell, p.ctx.ConfigDir)

	// Determine source file path
	sourcePath := filepath.Join(pluginDir, plugin.SourceFile)
	if err := rcModifier.InjectZshPlugin(plugin, sourcePath); err != nil {
		return "", fmt.Errorf("failed to inject source line: %w", err)
	}

	return pluginDir, nil
}

// uninstallOhMyZsh uninstalls an Oh My Zsh plugin.
func (p *PluginInstaller) uninstallOhMyZsh(plugin Plugin, zshShell *shell.ZshShell, installPath string) error {
	// Remove from plugins array
	if err := zshShell.RemoveFromPluginsArray(plugin.OhMyZshName); err != nil {
		// Non-fatal, continue
		fmt.Printf("Warning: failed to remove from plugins array: %v\n", err)
	}

	// Remove plugin directory
	if installPath != "" {
		if err := os.RemoveAll(installPath); err != nil {
			return fmt.Errorf("failed to remove plugin directory: %w", err)
		}
	}

	return nil
}

// uninstallHomebrew uninstalls a Homebrew-installed plugin.
func (p *PluginInstaller) uninstallHomebrew(plugin Plugin) error {
	// Remove source line from RC
	rcModifier := NewRCModifier(p.shell, p.ctx.ConfigDir)
	if err := rcModifier.RemoveZshPlugin(plugin.Name); err != nil {
		// Non-fatal, continue
		fmt.Printf("Warning: failed to remove source line: %v\n", err)
	}

	// Uninstall via brew
	cmd := exec.Command("brew", "uninstall", plugin.BrewPackage)
	if err := cmd.Run(); err != nil {
		// Non-fatal, may not have been installed via brew
		fmt.Printf("Warning: failed to uninstall via brew: %v\n", err)
	}

	return nil
}

// uninstallGitClone uninstalls a git-cloned plugin.
func (p *PluginInstaller) uninstallGitClone(plugin Plugin, installPath string) error {
	// Remove source line from RC
	rcModifier := NewRCModifier(p.shell, p.ctx.ConfigDir)
	if err := rcModifier.RemoveZshPlugin(plugin.Name); err != nil {
		fmt.Printf("Warning: failed to remove source line: %v\n", err)
	}

	// Remove plugin directory
	if installPath != "" {
		if err := os.RemoveAll(installPath); err != nil {
			return fmt.Errorf("failed to remove plugin directory: %w", err)
		}
	} else {
		// Try common locations
		commonPaths := []string{
			filepath.Join(p.ctx.HomeDir, ".zsh", plugin.Name),
			filepath.Join(p.ctx.HomeDir, ".local", "share", "zsh", "plugins", plugin.Name),
			filepath.Join(p.ctx.HomeDir, ".config", "zsh", "plugins", plugin.Name),
		}
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				if err := os.RemoveAll(path); err != nil {
					fmt.Printf("Warning: failed to remove %s: %v\n", path, err)
				}
			}
		}
	}

	return nil
}

// cloneRepo clones a git repository with retry support.
func (p *PluginInstaller) cloneRepo(ctx context.Context, repoURL, targetPath string) error {
	maxRetries := 3
	retryDelay := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repoURL, targetPath)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}

		if attempt < maxRetries {
			fmt.Printf("Clone attempt %d/%d failed: %v, retrying...\n", attempt, maxRetries, err)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		} else {
			return fmt.Errorf("git clone failed after %d attempts: %w\n%s", maxRetries, err, string(output))
		}
	}

	return nil
}

// updateGitRepo updates an existing git repository.
func (p *PluginInstaller) updateGitRepo(ctx context.Context, targetPath string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", targetPath, "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w\n%s", err, string(output))
	}
	return nil
}

// ensureCorrectOrder ensures that plugins with MustBeLast are sourced last.
func (p *PluginInstaller) ensureCorrectOrder(plugin Plugin, zshShell *shell.ZshShell) error {
	if !plugin.MustBeLast {
		return nil
	}

	// Get all supported plugins
	plugins := GetSupportedPlugins()

	// Create RC modifier to check and fix order
	rcModifier := NewRCModifier(p.shell, p.ctx.ConfigDir)

	return rcModifier.EnsurePluginOrder(plugins)
}

// rollbackInstallation rolls back an installation in case of failure.
func (p *PluginInstaller) rollbackInstallation(steps []string, method InstallMethod, plugin Plugin) {
	fmt.Printf("Rolling back installation of %s...\n", plugin.Name)

	// Create RC modifier for cleanup
	rcModifier := NewRCModifier(p.shell, p.ctx.ConfigDir)

	// Remove any added source lines
	if err := rcModifier.RemoveZshPlugin(plugin.Name); err != nil {
		fmt.Printf("Warning: failed to remove plugin section: %v\n", err)
	}

	// Method-specific rollback
	zshShell, ok := p.shell.(*shell.ZshShell)
	if !ok {
		return
	}

	switch method {
	case MethodOhMyZsh:
		// Remove from plugins array
		if err := zshShell.RemoveFromPluginsArray(plugin.OhMyZshName); err != nil {
			fmt.Printf("Warning: failed to remove from plugins array: %v\n", err)
		}

		// Remove plugin directory
		pluginDir := zshShell.GetOhMyZshPluginDir()
		targetDir := filepath.Join(pluginDir, plugin.OhMyZshName)
		if err := os.RemoveAll(targetDir); err != nil {
			fmt.Printf("Warning: failed to remove plugin directory: %v\n", err)
		}

	case MethodGitClone:
		// Remove cloned directory
		pluginDir := filepath.Join(p.ctx.HomeDir, ".zsh", plugin.Name)
		if err := os.RemoveAll(pluginDir); err != nil {
			fmt.Printf("Warning: failed to remove plugin directory: %v\n", err)
		}
	}

	fmt.Printf("Rollback complete for %s\n", plugin.Name)
}
