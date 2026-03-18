// Package installer provides dependency installation and management.
// This file implements installation verification.
package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Verifier handles verification of installed components.
type Verifier struct {
	// context is the installation context.
	context *InstallContext

	// resolver is the dependency resolver.
	resolver *DependencyResolver
}

// NewVerifier creates a new verifier.
func NewVerifier(ctx *InstallContext, resolver *DependencyResolver) *Verifier {
	return &Verifier{
		context:  ctx,
		resolver: resolver,
	}
}

// VerifyComponent verifies a single component.
func (v *Verifier) VerifyComponent(ctx context.Context, name string) (*VerificationResult, error) {
	dep := v.resolver.GetDependency(name)
	if dep == nil {
		return nil, fmt.Errorf("unknown component: %s", name)
	}

	result := &VerificationResult{
		Component: name,
		Checks:    []VerificationCheck{},
	}

	// Run verification based on component type
	switch dep.Type {
	case ComponentTypeBinary:
		v.verifyBinary(ctx, dep, result)
	case ComponentTypeFont:
		v.verifyFont(ctx, dep, result)
	case ComponentTypePackage:
		v.verifyPackage(ctx, dep, result)
	default:
		return nil, fmt.Errorf("unsupported component type: %s", dep.Type)
	}

	// Determine overall installed status
	result.Installed = v.isInstalled(result)

	return result, nil
}

// verifyBinary verifies a binary component.
func (v *Verifier) verifyBinary(ctx context.Context, dep *Dependency, result *VerificationResult) {
	// Check 1: Binary exists in PATH
	check1 := VerificationCheck{
		Name:    "binary_in_path",
		Passed:  false,
		Message: "Binary not found in PATH",
	}

	path, err := exec.LookPath(dep.Name)
	if err == nil {
		check1.Passed = true
		check1.Message = fmt.Sprintf("Binary found at %s", path)
		result.Path = path
		result.InPATH = true
	} else {
		// Check bin directory
		binPath := filepath.Join(v.context.BinDir, dep.Name)
		if _, err := os.Stat(binPath); err == nil {
			check1.Passed = true
			check1.Message = fmt.Sprintf("Binary found at %s (not in PATH)", binPath)
			result.Path = binPath
			result.InPATH = false
			result.Issues = append(result.Issues, "Binary installed but not in PATH")
		}
	}
	result.Checks = append(result.Checks, check1)

	// Check 2: Binary is executable
	if result.Path != "" {
		check2 := VerificationCheck{
			Name:    "binary_executable",
			Passed:  false,
			Message: "Binary is not executable",
		}

		info, err := os.Stat(result.Path)
		if err == nil {
			if info.Mode()&0111 != 0 {
				check2.Passed = true
				check2.Message = "Binary is executable"
			}
		}
		result.Checks = append(result.Checks, check2)
	}

	// Check 3: Run verify command if specified
	if dep.VerifyCommand != "" {
		check3 := VerificationCheck{
			Name:    "verify_command",
			Passed:  false,
			Message: "Verify command failed",
		}

		cmdParts := strings.Fields(dep.VerifyCommand)
		if len(cmdParts) > 0 {
			// Use the installed path if available
			if result.Path != "" && cmdParts[0] == dep.Name {
				cmdParts[0] = result.Path
			}

			cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
			output, err := cmd.CombinedOutput()

			if err == nil {
				check3.Passed = true
				check3.Message = "Verify command passed"
				result.Working = true

				// Extract version from output
				outputStr := strings.TrimSpace(string(output))
				result.Version = v.extractVersion(outputStr)
			} else {
				check3.Message = fmt.Sprintf("Command failed: %v", err)
				result.Issues = append(result.Issues, fmt.Sprintf("Verify command failed: %v", err))
			}
		}
		result.Checks = append(result.Checks, check3)
	}
}

// verifyFont verifies a font component.
func (v *Verifier) verifyFont(ctx context.Context, dep *Dependency, result *VerificationResult) {
	// Check 1: Font files exist
	check1 := VerificationCheck{
		Name:    "font_installed",
		Passed:  false,
		Message: "Font files not found",
	}

	fontPatterns := []string{
		dep.Name + "*.ttf",
		dep.Name + "*.otf",
		dep.Name + "*NerdFont*",
		"*NerdFont*" + dep.Name + "*",
	}

	fontFound := false
	for _, pattern := range fontPatterns {
		matches, err := filepath.Glob(filepath.Join(v.context.FontDir, pattern))
		if err == nil && len(matches) > 0 {
			fontFound = true
			result.Path = v.context.FontDir
			break
		}
	}

	if fontFound {
		check1.Passed = true
		check1.Message = "Font files found"
		result.Installed = true
	} else {
		// Check system font directories
		systemDirs := []string{
			"/Library/Fonts",
			"/System/Library/Fonts",
			"/usr/share/fonts",
			"/usr/local/share/fonts",
		}

		for _, dir := range systemDirs {
			for _, pattern := range fontPatterns {
				matches, err := filepath.Glob(filepath.Join(dir, pattern))
				if err == nil && len(matches) > 0 {
					fontFound = true
					check1.Passed = true
					check1.Message = "Font found in system directory"
					result.Path = dir
					break
				}
			}
			if fontFound {
				break
			}
		}
	}

	result.Checks = append(result.Checks, check1)

	// Check 2: Font cache is updated (Linux only)
	if v.context.OS != "darwin" {
		check2 := VerificationCheck{
			Name:    "font_cache",
			Passed:  false,
			Message: "Font cache not available",
		}

		if _, err := exec.LookPath("fc-cache"); err == nil {
			// fc-cache exists
			check2.Passed = true
			check2.Message = "Font cache tool available"
		}
		result.Checks = append(result.Checks, check2)
	}
}

// verifyPackage verifies a package installed via package manager.
func (v *Verifier) verifyPackage(ctx context.Context, dep *Dependency, result *VerificationResult) {
	// Check 1: Binary/command exists
	check1 := VerificationCheck{
		Name:    "binary_exists",
		Passed:  false,
		Message: "Binary not found",
	}

	path, err := exec.LookPath(dep.Name)
	if err == nil {
		check1.Passed = true
		check1.Message = fmt.Sprintf("Binary found at %s", path)
		result.Path = path
		result.InPATH = true
	}
	result.Checks = append(result.Checks, check1)

	// Check 2: Run verify command if specified
	if dep.VerifyCommand != "" {
		check2 := VerificationCheck{
			Name:    "verify_command",
			Passed:  false,
			Message: "Verify command failed",
		}

		cmdParts := strings.Fields(dep.VerifyCommand)
		if len(cmdParts) > 0 {
			cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
			output, err := cmd.CombinedOutput()

			if err == nil {
				check2.Passed = true
				check2.Message = "Verify command passed"
				result.Working = true
				result.Version = v.extractVersion(strings.TrimSpace(string(output)))
			} else {
				check2.Message = fmt.Sprintf("Command failed: %v", err)
			}
		}
		result.Checks = append(result.Checks, check2)
	}
}

// isInstalled determines if all checks pass.
func (v *Verifier) isInstalled(result *VerificationResult) bool {
	for _, check := range result.Checks {
		if !check.Passed {
			return false
		}
	}
	return len(result.Checks) > 0
}

// extractVersion extracts version from command output.
func (v *Verifier) extractVersion(output string) string {
	// Similar to extractVersion in installer.go but as a method
	if strings.Contains(strings.ToLower(output), "version") {
		parts := strings.Fields(output)
		for i, part := range parts {
			if strings.ToLower(part) == "version" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}

	// Look for version pattern
	fields := strings.Fields(output)
	for _, field := range fields {
		if strings.Count(field, ".") >= 1 {
			if len(field) >= 3 && (field[0] >= '0' && field[0] <= '9' || (field[0] == 'v' && len(field) >= 4)) {
				return strings.TrimPrefix(field, "v")
			}
		}
	}

	return output
}

// VerifyAll verifies all installed components.
func (v *Verifier) VerifyAll(ctx context.Context) (map[string]*VerificationResult, error) {
	results := make(map[string]*VerificationResult)

	// Get all registered dependencies
	deps := v.resolver.GetAllDependencies()

	for _, dep := range deps {
		result, err := v.VerifyComponent(ctx, dep.Name)
		if err != nil {
			results[dep.Name] = &VerificationResult{
				Component: dep.Name,
				Installed: false,
				Issues:    []string{err.Error()},
			}
			continue
		}
		results[dep.Name] = result
	}

	return results, nil
}

// QuickVerify performs a quick verification of essential components.
func (v *Verifier) QuickVerify(ctx context.Context, components []string) ([]*VerificationResult, error) {
	results := make([]*VerificationResult, 0, len(components))

	for _, name := range components {
		result, err := v.VerifyComponent(ctx, name)
		if err != nil {
			results = append(results, &VerificationResult{
				Component: name,
				Installed: false,
				Issues:    []string{err.Error()},
			})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// IsAllInstalled checks if all components are installed.
func (v *Verifier) IsAllInstalled(ctx context.Context, components []string) (bool, error) {
	for _, name := range components {
		result, err := v.VerifyComponent(ctx, name)
		if err != nil {
			return false, err
		}
		if !result.Installed {
			return false, nil
		}
	}
	return true, nil
}

// GetInstallReport generates a human-readable installation report.
func (v *Verifier) GetInstallReport(results map[string]*VerificationResult) string {
	var sb strings.Builder

	sb.WriteString("Installation Verification Report\n")
	sb.WriteString("================================\n\n")

	installed := 0
	failed := 0

	for name, result := range results {
		if result.Installed {
			installed++
			sb.WriteString(fmt.Sprintf("✓ %s: installed", name))
			if result.Version != "" {
				sb.WriteString(fmt.Sprintf(" (v%s)", result.Version))
			}
			sb.WriteString("\n")
		} else {
			failed++
			sb.WriteString(fmt.Sprintf("✗ %s: NOT installed", name))
			if len(result.Issues) > 0 {
				sb.WriteString(fmt.Sprintf(" - %s", result.Issues[0]))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(fmt.Sprintf("\nSummary: %d installed, %d missing\n", installed, failed))

	return sb.String()
}

// VerifyRCFile verifies RC file modifications.
func (v *Verifier) VerifyRCFile(ctx context.Context, shellType string) (*VerificationResult, error) {
	result := &VerificationResult{
		Component: "rc_file",
		Checks:    []VerificationCheck{},
	}

	// RC file check depends on shell type
	rcPath := filepath.Join(v.context.HomeDir, ".zshrc")
	if shellType == "bash" {
		rcPath = filepath.Join(v.context.HomeDir, ".bashrc")
	}

	// Check 1: RC file exists
	check1 := VerificationCheck{
		Name:    "rc_exists",
		Passed:  false,
		Message: "RC file does not exist",
	}

	if _, err := os.Stat(rcPath); err == nil {
		check1.Passed = true
		check1.Message = "RC file exists"
	}
	result.Checks = append(result.Checks, check1)

	// Check 2: Savanhi markers present
	check2 := VerificationCheck{
		Name:    "savanhi_markers",
		Passed:  false,
		Message: "Savanhi markers not found",
	}

	if content, err := os.ReadFile(rcPath); err == nil {
		if strings.Contains(string(content), "savanhi") {
			check2.Passed = true
			check2.Message = "Savanhi configuration found"
		}
	}
	result.Checks = append(result.Checks, check2)

	// Check 3: PATH includes Savanhi bin
	check3 := VerificationCheck{
		Name:    "bin_in_path",
		Passed:  false,
		Message: "Savanhi bin directory not in PATH",
	}

	if content, err := os.ReadFile(rcPath); err == nil {
		if strings.Contains(string(content), v.context.BinDir) {
			check3.Passed = true
			check3.Message = "Savanhi bin directory in PATH"
		}
	}
	result.Checks = append(result.Checks, check3)

	result.Installed = v.isInstalled(result)
	return result, nil
}

// VerifyComplete performs a complete system verification.
func (v *Verifier) VerifyComplete(ctx context.Context) (*CompleteVerificationResult, error) {
	result := &CompleteVerificationResult{
		Components: make(map[string]*VerificationResult),
		Timestamp:  time.Now(),
	}

	// Verify all components
	allResults, err := v.VerifyAll(ctx)
	if err != nil {
		return nil, err
	}
	result.Components = allResults

	// Verify RC file
	rcResult, err := v.VerifyRCFile(ctx, v.context.Shell)
	if err != nil {
		result.RCFile = &VerificationResult{
			Component: "rc_file",
			Installed: false,
			Issues:    []string{err.Error()},
		}
	} else {
		result.RCFile = rcResult
	}

	// Calculate overall status
	allInstalled := true
	for _, r := range result.Components {
		if !r.Installed {
			allInstalled = false
			break
		}
	}
	result.AllInstalled = allInstalled && result.RCFile.Installed

	return result, nil
}

// CompleteVerificationResult represents a complete verification result.
type CompleteVerificationResult struct {
	// Components contains verification results for all components.
	Components map[string]*VerificationResult `json:"components"`

	// RCFile contains the RC file verification result.
	RCFile *VerificationResult `json:"rc_file"`

	// AllInstalled indicates if everything is installed.
	AllInstalled bool `json:"all_installed"`

	// Timestamp is when the verification was performed.
	Timestamp time.Time `json:"timestamp"`
}
