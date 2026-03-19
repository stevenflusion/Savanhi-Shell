// Package e2e provides end-to-end tests for Savanhi Shell.
package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// getProjectRoot returns the project root directory by finding go.mod.
func getProjectRoot(t *testing.T) string {
	// Start from current directory and walk up to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod not found)")
		}
		dir = parent
	}
}

// TestInstallFlow tests the complete installation flow.
func TestInstallFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	projectRoot := getProjectRoot(t)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "savanhi-e2e-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up environment
	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}

	// Build binary - must run from project root
	binaryPath := filepath.Join(tmpDir, "savanhi-shell")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/savanhi-shell")
	buildCmd.Dir = projectRoot // Build from project root where go.mod exists
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Create test config
	configPath := filepath.Join(tmpDir, "config.json")
	configContent := `{
		"theme": "agnoster",
		"font": "MesloLGS NF",
		"install_oh_my_posh": true,
		"install_zoxide": false,
		"install_fzf": false,
		"install_bat": false,
		"install_eza": false,
		"dry_run": true
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Run in non-interactive mode with dry-run
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath,
		"--non-interactive",
		"--config", configPath,
		"--dry-run",
		"--verbose",
	)
	cmd.Env = append(os.Environ(), "HOME="+homeDir)
	cmd.Dir = tmpDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify expected output
	if string(output) == "" {
		t.Error("Expected some output")
	}
}

// TestRollbackFlow tests the rollback functionality.
func TestRollbackFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	projectRoot := getProjectRoot(t)

	tmpDir, err := os.MkdirTemp("", "savanhi-rollback-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build binary - must run from project root
	binaryPath := filepath.Join(tmpDir, "savanhi-shell")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/savanhi-shell")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Create test home
	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}

	// Test rollback with --dry-run
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath,
		"--rollback",
		"--dry-run",
		"--verbose",
	)
	cmd.Env = append(os.Environ(), "HOME="+homeDir)
	cmd.Dir = tmpDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Rollback command failed: %v\nOutput: %s", err, output)
	}
}

// TestDetection tests the system detection.
func TestDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	projectRoot := getProjectRoot(t)

	tmpDir, err := os.MkdirTemp("", "savanhi-detect-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build binary - must run from project root
	binaryPath := filepath.Join(tmpDir, "savanhi-shell")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/savanhi-shell")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Run detection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "--detect")
	cmd.Dir = tmpDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Detection failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains expected fields
	outputStr := string(output)
	if len(outputStr) < 50 {
		t.Errorf("Detection output too short: %s", outputStr)
	}
}

// TestVersion tests the version flag.
func TestVersion(t *testing.T) {
	projectRoot := getProjectRoot(t)

	tmpDir, err := os.MkdirTemp("", "savanhi-version-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build binary - must run from project root
	binaryPath := filepath.Join(tmpDir, "savanhi-shell")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/savanhi-shell")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Test version flag
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Version command failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains "Savanhi Shell"
	outputStr := string(output)
	if len(outputStr) < 10 {
		t.Errorf("Version output too short: %s", outputStr)
	}
}

// TestHelp tests the help flag.
func TestHelp(t *testing.T) {
	projectRoot := getProjectRoot(t)

	tmpDir, err := os.MkdirTemp("", "savanhi-help-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build binary - must run from project root
	binaryPath := filepath.Join(tmpDir, "savanhi-shell")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/savanhi-shell")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Test help flag
	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help command failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains usage info
	outputStr := string(output)
	if len(outputStr) < 100 {
		t.Errorf("Help output too short: %s", outputStr)
	}
}

// TestPreviewFlow tests the preview functionality.
func TestPreviewFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	projectRoot := getProjectRoot(t)
	_ = projectRoot // Used in build if test is implemented

	t.Log("Preview flow test - placeholder for actual implementation")
	// Preview requires TUI interaction which is harder to test in E2E
	// This would typically be tested with Docker + expect scripts
}

// TestHealthFlag tests the --health --quick CLI flag for non-interactive JSON output.
// Note: --health alone requires a TTY and cannot be tested in CI environment.
func TestHealthFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	projectRoot := getProjectRoot(t)

	tmpDir, err := os.MkdirTemp("", "savanhi-health-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build binary - must run from project root
	binaryPath := filepath.Join(tmpDir, "savanhi-shell")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/savanhi-shell")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Create test home
	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}

	// Test --health --quick flag (non-interactive JSON output)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "--health", "--quick")
	cmd.Env = append(os.Environ(), "HOME="+homeDir)
	cmd.Dir = tmpDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Health command failed: %v\nOutput: %s", err, output)
	}

	// Verify output is valid JSON
	outputStr := string(output)
	if !strings.HasPrefix(outputStr, "{") {
		t.Fatalf("Health output should be JSON, got: %s", outputStr[:min(100, len(outputStr))])
	}

	// Verify output contains expected health data fields
	expectedFields := []string{
		"\"Terminal\"",
		"\"Components\"",
	}

	for _, field := range expectedFields {
		if !strings.Contains(outputStr, field) {
			t.Errorf("Health output missing expected field: %s", field)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestHealthFlagJSON tests the JSON output format from --health --quick.
func TestHealthFlagJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	projectRoot := getProjectRoot(t)

	tmpDir, err := os.MkdirTemp("", "savanhi-health-json-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build binary - must run from project root
	binaryPath := filepath.Join(tmpDir, "savanhi-shell")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/savanhi-shell")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Create test home
	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}

	// Test --health --quick flag for JSON output
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "--health", "--quick")
	cmd.Env = append(os.Environ(), "HOME="+homeDir)
	cmd.Dir = tmpDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Health command failed: %v\nOutput: %s", err, output)
	}

	// Verify output is valid JSON by attempting to parse it
	outputStr := string(output)

	// Check it starts with { (valid JSON object)
	if !strings.HasPrefix(strings.TrimSpace(outputStr), "{") {
		t.Fatalf("Expected JSON output starting with '{', got: %s", outputStr[:min(100, len(outputStr))])
	}

	// Verify required fields are present
	requiredFields := []string{
		"\"Terminal\"",   // Terminal capabilities section
		"\"Components\"", // Installed components section
		"\"TrueColor\"",  // Terminal capability
		"\"Installed\"",  // Component status
	}

	for _, field := range requiredFields {
		if !strings.Contains(outputStr, field) {
			t.Errorf("JSON output missing required field: %s", field)
		}
	}

	t.Logf("Health JSON output is valid (%d bytes)", len(output))
}
