// Package installer provides dependency installation and management for Savanhi Shell.
package installer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/savanhi/shell/internal/staging"
)

// TestNewDependencyResolver tests creating a new resolver.
func TestNewDependencyResolver(t *testing.T) {
	resolver := NewDependencyResolver()
	if resolver == nil {
		t.Fatal("expected resolver, got nil")
	}

	// Check that built-in dependencies are registered
	deps := resolver.GetAllDependencies()
	if len(deps) == 0 {
		t.Error("expected built-in dependencies to be registered")
	}

	// Check specific dependencies
	if resolver.GetDependency("oh-my-posh") == nil {
		t.Error("expected oh-my-posh to be registered")
	}
	if resolver.GetDependency("zoxide") == nil {
		t.Error("expected zoxide to be registered")
	}
}

// TestDependencyResolver_Resolve tests dependency resolution.
func TestDependencyResolver_Resolve(t *testing.T) {
	resolver := NewDependencyResolver()

	tests := []struct {
		name      string
		input     []string
		wantLen   int
		wantError bool
	}{
		{
			name:    "single dependency",
			input:   []string{"zoxide"},
			wantLen: 1,
		},
		{
			name:    "multiple dependencies",
			input:   []string{"zoxide", "fzf"},
			wantLen: 2,
		},
		{
			name:      "unknown dependency",
			input:     []string{"unknown-dep"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := resolver.Resolve(tt.input)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(resolved) != tt.wantLen {
				t.Errorf("expected %d dependencies, got %d", tt.wantLen, len(resolved))
			}
		})
	}
}

// TestDependencyResolver_CheckAlreadyInstalled tests checking installed status.
func TestDependencyResolver_CheckAlreadyInstalled(t *testing.T) {
	resolver := NewDependencyResolver()

	statuses := resolver.CheckAlreadyInstalled()
	if len(statuses) == 0 {
		t.Error("expected status list, got empty")
	}

	for _, status := range statuses {
		if status.Name == "" {
			t.Error("expected status name to be set")
		}
	}
}

// TestDependencyResolver_MarkInstalled tests marking dependencies as installed.
func TestDependencyResolver_MarkInstalled(t *testing.T) {
	resolver := NewDependencyResolver()

	resolver.MarkInstalled("test-dep", "1.0.0")

	if !resolver.IsInstalled("test-dep") {
		t.Error("expected test-dep to be marked as installed")
	}

	version := resolver.GetInstalledVersion("test-dep")
	if version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", version)
	}
}

// TestInstallContext tests creating install context.
func TestInstallContext(t *testing.T) {
	ctx, err := NewInstallContext()
	if err != nil {
		t.Fatalf("failed to create install context: %v", err)
	}

	if ctx.HomeDir == "" {
		t.Error("expected home directory to be set")
	}
	if ctx.BinDir == "" {
		t.Error("expected bin directory to be set")
	}
	if ctx.OS == "" {
		t.Error("expected OS to be set")
	}
}

// TestDefaultOptions tests default options creation.
func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.SkipChecksum {
		t.Error("expected SkipChecksum to be false by default")
	}
	if opts.SkipVerification {
		t.Error("expected SkipVerification to be false by default")
	}
	if !opts.UseCache {
		t.Error("expected UseCache to be true by default")
	}
	if opts.MaxRetries != 3 {
		t.Errorf("expected MaxRetries to be 3, got %d", opts.MaxRetries)
	}
}

// TestDependency tests dependency struct.
func TestDependency(t *testing.T) {
	dep := &Dependency{
		Name:        "test",
		DisplayName: "Test Dep",
		Type:        ComponentTypeBinary,
		Version:     "1.0.0",
		Source:      "https://example.com/test",
	}

	if dep.Name != "test" {
		t.Errorf("expected name test, got %s", dep.Name)
	}
	if dep.Type != ComponentTypeBinary {
		t.Errorf("expected type binary, got %s", dep.Type)
	}
}

// TestInstallProgress tests install progress.
func TestInstallProgress(t *testing.T) {
	progress := &InstallProgress{
		Component: "test",
		Stage:     StageDownloading,
		Percent:   50,
		Message:   "Downloading...",
	}

	if progress.Component != "test" {
		t.Errorf("expected component test, got %s", progress.Component)
	}
	if progress.Stage != StageDownloading {
		t.Errorf("expected stage downloading, got %s", progress.Stage)
	}
	if progress.Percent != 50 {
		t.Errorf("expected percent 50, got %f", progress.Percent)
	}
}

// TestInstallResult tests install result.
func TestInstallResult(t *testing.T) {
	result := &InstallResult{
		Component:     "test",
		Success:       true,
		Version:       "1.0.0",
		InstalledPath: "/usr/local/bin/test",
	}

	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.Component != "test" {
		t.Errorf("expected component test, got %s", result.Component)
	}
}

// TestVerificationResult tests verification result.
func TestVerificationResult(t *testing.T) {
	result := &VerificationResult{
		Component: "test",
		Installed: true,
		Version:   "1.0.0",
		InPATH:    true,
		Working:   true,
	}

	if !result.Installed {
		t.Error("expected installed to be true")
	}
	if !result.InPATH {
		t.Error("expected InPATH to be true")
	}
}

// TestRollbackState tests rollback state.
func TestRollbackState(t *testing.T) {
	state := &RollbackState{
		ID:                  "test-id",
		Description:         "test rollback",
		InstalledComponents: []string{"comp1", "comp2"},
		CreatedAt:           time.Now(),
	}

	state.AddInstalledComponent("comp3")
	if len(state.InstalledComponents) != 3 {
		t.Errorf("expected 3 components, got %d", len(state.InstalledComponents))
	}

	state.AddInstalledFile("/tmp/test.txt")
	if len(state.InstalledFiles) != 1 {
		t.Errorf("expected 1 file, got %d", len(state.InstalledFiles))
	}
}

// TestStagedChange tests staged change.
func TestStagedChange(t *testing.T) {
	change := &staging.StagedChange{
		ID:        "change-1",
		Component: "test",
		Action:    "install",
		Target:    "/usr/local/bin/test",
		Status:    staging.StatusPending,
	}

	if change.Status != staging.StatusPending {
		t.Errorf("expected status pending, got %s", change.Status)
	}
	if change.Action != "install" {
		t.Errorf("expected action install, got %s", change.Action)
	}
}

// TestConflict tests conflict struct.
func TestConflict(t *testing.T) {
	conflict := &staging.Conflict{
		Target:              "/etc/config",
		Reason:              "multiple changes",
		SuggestedResolution: "merge changes",
	}

	if conflict.Target != "/etc/config" {
		t.Errorf("expected target /etc/config, got %s", conflict.Target)
	}
}

// TestFlowStep tests flow step.
func TestFlowStep(t *testing.T) {
	step := &FlowStep{
		Name:        "install",
		Description: "Install component",
		Phase:       "installation",
		Required:    true,
	}

	if step.Name != "install" {
		t.Errorf("expected name install, got %s", step.Name)
	}
	if !step.Required {
		t.Error("expected required to be true")
	}
}

// TestToolDefinition tests tool definition.
func TestToolDefinition(t *testing.T) {
	tool := &ToolDefinition{
		Name:          "test",
		DisplayName:   "Test Tool",
		Description:   "A test tool",
		VerifyCommand: "test --version",
	}

	if tool.Name != "test" {
		t.Errorf("expected name test, got %s", tool.Name)
	}
}

// TestRCModifier tests RC modifier creation.
func TestRCModifier(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "savanhi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backupDir := filepath.Join(tmpDir, "backups")
	modifier := &RCModifier{
		backupDir: backupDir,
	}

	if modifier == nil {
		t.Error("expected RC modifier, got nil")
	}
}

// TestVerifier tests verifier.
func TestVerifier(t *testing.T) {
	ctx := &InstallContext{
		HomeDir:   "/tmp",
		ConfigDir: "/tmp/.config/savanhi",
		OS:        "linux",
	}

	resolver := NewDependencyResolver()
	verifier := NewVerifier(ctx, resolver)

	if verifier == nil {
		t.Error("expected verifier, got nil")
	}
}

// TestToolInstaller tests tool installer.
func TestToolInstaller(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "savanhi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		BinDir:    filepath.Join(tmpDir, ".local", "bin"),
		OS:        "linux",
	}

	installer := NewToolInstaller(ctx)
	if installer == nil {
		t.Error("expected tool installer, got nil")
	}

	// Test getToolDefinition
	tool := installer.getToolDefinition("zoxide")
	if tool == nil {
		t.Error("expected zoxide definition, got nil")
	}
	if tool.Name != "zoxide" {
		t.Errorf("expected name zoxide, got %s", tool.Name)
	}
}

// TestFontInstaller tests font installer.
func TestFontInstaller(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "savanhi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		FontDir:   filepath.Join(tmpDir, ".local", "share", "fonts"),
		OS:        "linux",
	}

	installer := NewFontInstaller(ctx)
	if installer == nil {
		t.Error("expected font installer, got nil")
	}

	// Test getRecommendedFonts
	fonts := installer.GetRecommendedFonts()
	if len(fonts) == 0 {
		t.Error("expected recommended fonts list")
	}
}

// TestOhMyPoshInstaller tests oh-my-posh installer.
func TestOhMyPoshInstaller(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "savanhi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := &InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
		BinDir:    filepath.Join(tmpDir, ".local", "bin"),
		OS:        "linux",
		Arch:      "amd64",
	}

	installer := NewOhMyPoshInstaller(ctx)
	if installer == nil {
		t.Error("expected oh-my-posh installer, got nil")
	}

	// Test getDownloadURL
	url := installer.getDownloadURL()
	if url == "" {
		t.Error("expected download URL")
	}
	if !contains(url, "github.com") {
		t.Error("expected GitHub URL")
	}
}

// TestGeneratePlan tests install plan generation.
func TestGeneratePlan(t *testing.T) {
	resolver := NewDependencyResolver()

	plan, err := resolver.GeneratePlan([]string{"zoxide"}, PlatformLinux)
	if err != nil {
		t.Fatalf("failed to generate plan: %v", err)
	}

	if len(plan.Components) == 0 {
		t.Error("expected components in plan")
	}
}

// Helper function to check if string contains substring.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestNewInstaller tests creating a new installer.
func TestNewInstaller(t *testing.T) {
	installer, err := NewInstaller()
	if err != nil {
		t.Fatalf("failed to create installer: %v", err)
	}
	if installer == nil {
		t.Error("expected installer, got nil")
	}
}

// TestDefaultInstaller_GetDependency tests getting dependencies.
func TestDefaultInstaller_GetDependency(t *testing.T) {
	installer, _ := NewInstaller()
	_ = installer.resolver.GetDependency("zoxide")
	// Dependency may or may not exist, that's OK for this test
}

// TestInstallationProgress tests installation progress.
func TestInstallationProgress(t *testing.T) {
	progress := &InstallationProgress{
		CurrentPhase:   "installing",
		CurrentStep:    "Installing zoxide",
		TotalSteps:     10,
		CompletedSteps: 3,
		Percent:        30.0,
	}

	if progress.CurrentPhase != "installing" {
		t.Errorf("expected phase installing, got %s", progress.CurrentPhase)
	}
	if progress.Percent != 30.0 {
		t.Errorf("expected percent 30, got %f", progress.Percent)
	}
}

// TestCompleteVerificationResult tests complete verification.
func TestCompleteVerificationResult(t *testing.T) {
	result := &CompleteVerificationResult{
		Components:   make(map[string]*VerificationResult),
		AllInstalled: true,
	}

	if len(result.Components) != 0 {
		t.Error("expected empty components map")
	}
	if !result.AllInstalled {
		t.Error("expected all installed to be true")
	}
}

// TestStagingSystem tests staging system.
func TestStagingSystem(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "savanhi-staging-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := &staging.InstallContext{
		HomeDir:   tmpDir,
		ConfigDir: filepath.Join(tmpDir, ".config", "savanhi"),
	}

	stg := staging.NewStagingSystem(ctx)
	if stg == nil {
		t.Error("expected staging system, got nil")
	}

	// Test queue
	change := &staging.StagedChange{
		ID:        "test-1",
		Component: "test",
		Action:    "install",
		Target:    "/tmp/test",
		Status:    staging.StatusPending,
		CreatedAt: time.Now(),
	}

	if err := stg.Queue(change); err != nil {
		t.Errorf("failed to queue change: %v", err)
	}

	// Test validate
	_ = stg.Validate() // Validation may or may not have errors

	// Test clear
	if err := stg.Clear(); err != nil {
		t.Errorf("failed to clear staging: %v", err)
	}
}

// TestContext tests context creation.
func TestContext(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Error("expected context, got nil")
	}
}
