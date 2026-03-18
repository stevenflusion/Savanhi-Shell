// Package installer provides dependency installation and management.
// This file implements the installation flow orchestration.
package installer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/savanhi/shell/internal/persistence"
	"github.com/savanhi/shell/pkg/shell"
)

// InstallationFlow orchestrates the complete installation process.
type InstallationFlow struct {
	// context is the installation context.
	context *InstallContext

	// installer is the default installer.
	installer *DefaultInstaller

	// resolver is the dependency resolver.
	resolver *DependencyResolver

	// verifier is the verifier.
	verifier *Verifier

	// rcModifier is the RC file modifier.
	rcModifier *RCModifier

	// rollback is the rollback manager.
	rollback *RollbackManager

	// persister is the persistence layer.
	persister persistence.Persister

	// progress tracks installation progress.
	progress *InstallationProgress

	// options are installation options.
	options *Options

	// mu protects concurrent access.
	mu sync.Mutex
}

// InstallationProgress tracks the progress of an installation.
type InstallationProgress struct {
	// CurrentPhase is the current installation phase.
	CurrentPhase string `json:"current_phase"`

	// CurrentStep is the current step within the phase.
	CurrentStep string `json:"current_step"`

	// TotalSteps is the total number of steps.
	TotalSteps int `json:"total_steps"`

	// CompletedSteps is the number of completed steps.
	CompletedSteps int `json:"completed_steps"`

	// Percent is the overall completion percentage.
	Percent float64 `json:"percent"`

	// StartTime is when the installation started.
	StartTime time.Time `json:"start_time"`

	// Errors are any errors encountered.
	Errors []string `json:"errors,omitempty"`

	// Warnings are any warnings encountered.
	Warnings []string `json:"warnings,omitempty"`
}

// FlowStep represents a single step in the installation flow.
type FlowStep struct {
	// Name is the step name.
	Name string `json:"name"`

	// Description is a human-readable description.
	Description string `json:"description"`

	// Phase is the phase this step belongs to.
	Phase string `json:"phase"`

	// Required indicates if this step is required.
	Required bool `json:"required"`

	// Skipped indicates if this step was skipped.
	Skipped bool `json:"skipped,omitempty"`

	// Completed indicates if this step is completed.
	Completed bool `json:"completed"`

	// Error is any error from this step.
	Error string `json:"error,omitempty"`
}

// NewInstallationFlow creates a new installation flow.
func NewInstallationFlow(ctx *InstallContext, persister persistence.Persister, shellImpl shell.Shell) *InstallationFlow {
	installer, _ := NewInstaller()
	resolver := NewDependencyResolver()
	verifier := NewVerifier(ctx, resolver)
	rcModifier := NewRCModifier(shellImpl, ctx.ConfigDir)
	rollback := NewRollbackManager(persister, ctx, shellImpl)

	return &InstallationFlow{
		context:    ctx,
		installer:  installer,
		resolver:   resolver,
		verifier:   verifier,
		rcModifier: rcModifier,
		rollback:   rollback,
		persister:  persister,
		progress: &InstallationProgress{
			StartTime: time.Now(),
		},
		options: DefaultOptions(),
	}
}

// SetOptions sets installation options.
func (f *InstallationFlow) SetOptions(opts *Options) {
	f.options = opts
}

// Install performs the complete installation.
func (f *InstallationFlow) Install(ctx context.Context, components []string) error {
	f.mu.Lock()
	f.progress.CurrentPhase = "preparing"
	f.mu.Unlock()

	// Step 1: Create backup
	backupPath, err := f.createBackup()
	if err != nil {
		f.progress.Warnings = append(f.progress.Warnings, fmt.Sprintf("Backup warning: %v", err))
	}
	_ = backupPath // Used for logging if needed

	// Track state for rollback
	rollbackState := &RollbackState{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		CreatedAt:   time.Now(),
		Description: "pre-installation",
	}

	// Track progress
	steps := f.buildSteps(components)
	f.progress.TotalSteps = len(steps)

	// Step 2: Resolve dependencies
	f.updateProgress("resolving", "Resolving dependencies", 10)
	plan, err := f.resolver.GeneratePlan(components, Platform(f.context.OS))
	if err != nil {
		f.progress.Errors = append(f.progress.Errors, err.Error())
		f.performRollback(ctx, rollbackState)
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Check for conflicts
	if len(plan.Conflicts) > 0 {
		f.progress.Warnings = append(f.progress.Warnings, plan.Conflicts...)
	}

	// Step 3: Create backup
	f.updateProgress("backup", "Creating backup", 15)
	if !f.context.DryRun {
		backup, err := f.persister.CreateBackup("pre-install", f.getRCFiles())
		if err != nil {
			f.progress.Warnings = append(f.progress.Warnings, fmt.Sprintf("Backup warning: %v", err))
		} else {
			rollbackState.RCBackupPath = backup.ID
		}
	}

	// Step 4: Stage changes
	f.updateProgress("staging", "Staging changes", 20)
	if err := f.stageChanges(plan); err != nil {
		f.progress.Errors = append(f.progress.Errors, err.Error())
		f.performRollback(ctx, rollbackState)
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Step 5: Install dependencies
	currentStep := 0
	for _, dep := range plan.Components {
		currentStep++
		percent := 25 + float64(currentStep)*50/float64(len(plan.Components))
		f.updateProgress("installing", fmt.Sprintf("Installing %s", dep.DisplayName), percent)

		if dep.IsInstalled && !f.context.Force {
			f.progress.Warnings = append(f.progress.Warnings, fmt.Sprintf("%s already installed, skipping", dep.DisplayName))
			continue
		}

		if !f.context.DryRun {
			result, err := f.installer.Install(ctx, dep, f.options)
			if err != nil {
				f.progress.Errors = append(f.progress.Errors, err.Error())
				if !dep.Optional {
					f.performRollback(ctx, rollbackState)
					return fmt.Errorf("failed to install %s: %w", dep.Name, err)
				}
				f.progress.Warnings = append(f.progress.Warnings, fmt.Sprintf("Failed to install optional %s: %v", dep.DisplayName, err))
				continue
			}

			// Track installed component
			rollbackState.AddInstalledComponent(dep.Name)
			if result.InstalledPath != "" {
				rollbackState.AddInstalledFile(result.InstalledPath)
			}
		}
	}

	// Step 6: Modify RC files
	f.updateProgress("configuring", "Configuring shell", 80)
	if !f.context.DryRun {
		if err := f.configureShell(plan); err != nil {
			f.progress.Errors = append(f.progress.Errors, err.Error())
			f.performRollback(ctx, rollbackState)
			return fmt.Errorf("failed to configure shell: %w", err)
		}
	}

	// Step 7: Verify installation
	f.updateProgress("verifying", "Verifying installation", 90)
	if !f.context.DryRun && !f.options.SkipVerification {
		if err := f.verifyInstallation(ctx, components); err != nil {
			f.progress.Errors = append(f.progress.Errors, err.Error())
			f.progress.Warnings = append(f.progress.Warnings, "Verification had issues")
		}
	}

	// Step 8: Commit staging
	f.updateProgress("finalizing", "Finalizing", 95)
	if !f.context.DryRun {
		// Clear staging after successful installation
	}

	// Completion
	f.updateProgress("completed", "Installation complete", 100)
	f.progress.CurrentPhase = "completed"

	return nil
}

// buildSteps builds the installation steps.
func (f *InstallationFlow) buildSteps(components []string) []FlowStep {
	steps := []FlowStep{
		{Name: "prepare", Description: "Prepare installation", Phase: "preparation", Required: true},
		{Name: "backup", Description: "Create backup", Phase: "preparation", Required: true},
		{Name: "stage", Description: "Stage changes", Phase: "preparation", Required: true},
	}

	for _, comp := range components {
		steps = append(steps, FlowStep{
			Name:        fmt.Sprintf("install_%s", comp),
			Description: fmt.Sprintf("Install %s", comp),
			Phase:       "installation",
			Required:    true,
		})
	}

	steps = append(steps,
		FlowStep{Name: "configure", Description: "Configure shell", Phase: "configuration", Required: true},
		FlowStep{Name: "verify", Description: "Verify installation", Phase: "verification", Required: true},
		FlowStep{Name: "finalize", Description: "Finalize installation", Phase: "finalization", Required: true},
	)

	return steps
}

// createBackup creates a backup before installation.
func (f *InstallationFlow) createBackup() (string, error) {
	backupPath, err := f.rcModifier.Backup()
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}
	return backupPath, nil
}

// stageChanges stages the planned changes.
func (f *InstallationFlow) stageChanges(plan *InstallPlan) error {
	// Staging is handled by the staging system
	// For flow purposes, we validate the plan here
	if len(plan.Conflicts) > 0 {
		for _, conflict := range plan.Conflicts {
			f.progress.Warnings = append(f.progress.Warnings, conflict)
		}
	}

	return nil
}

// configureShell configures the shell RC files.
func (f *InstallationFlow) configureShell(plan *InstallPlan) error {
	// Prepare RC file for modification
	if err := f.rcModifier.PrepareForInstall(); err != nil {
		return fmt.Errorf("failed to prepare RC file: %w", err)
	}

	// Add PATH modification
	binDir := f.context.BinDir
	if binDir != "" {
		if err := f.rcModifier.InjectPath(binDir); err != nil {
			f.progress.Warnings = append(f.progress.Warnings, fmt.Sprintf("Failed to add PATH: %v", err))
		}
	}

	// Configure installed components
	for _, dep := range plan.Components {
		switch dep.Name {
		case "oh-my-posh":
			// Oh-my-posh configuration is handled by theme selection
		case "zoxide":
			if err := f.rcModifier.InjectZoxide(); err != nil {
				f.progress.Warnings = append(f.progress.Warnings, fmt.Sprintf("Failed to configure zoxide: %v", err))
			}
		case "fzf":
			if err := f.rcModifier.InjectFZF(); err != nil {
				f.progress.Warnings = append(f.progress.Warnings, fmt.Sprintf("Failed to configure fzf: %v", err))
			}
		case "bat":
			if err := f.rcModifier.InjectBatAliases(); err != nil {
				f.progress.Warnings = append(f.progress.Warnings, fmt.Sprintf("Failed to configure bat: %v", err))
			}
		case "eza":
			if err := f.rcModifier.InjectEzaAliases(); err != nil {
				f.progress.Warnings = append(f.progress.Warnings, fmt.Sprintf("Failed to configure eza: %v", err))
			}
		}
	}

	return nil
}

// verifyInstallation verifies the installation.
func (f *InstallationFlow) verifyInstallation(ctx context.Context, components []string) error {
	results, err := f.verifier.QuickVerify(ctx, components)
	if err != nil {
		return err
	}

	for _, result := range results {
		if !result.Installed {
			f.progress.Warnings = append(f.progress.Warnings, fmt.Sprintf("%s not verified correctly", result.Component))
		}
	}

	return nil
}

// performRollback rolls back the installation.
func (f *InstallationFlow) performRollback(ctx context.Context, state *RollbackState) {
	f.progress.CurrentPhase = "rollback"

	result, err := f.rollback.Rollback(state)
	if err != nil {
		f.progress.Errors = append(f.progress.Errors, fmt.Sprintf("Rollback failed: %v", err))
		return
	}

	f.progress.Warnings = append(f.progress.Warnings, result.Warnings...)
}

// getRCFiles returns the list of RC files to backup.
func (f *InstallationFlow) getRCFiles() []string {
	files := []string{}

	// Add shell RC files
	rcPath, err := f.rcModifier.shell.GetRCPath()
	if err == nil {
		files = append(files, rcPath)
	}

	return files
}

// updateProgress updates the installation progress.
func (f *InstallationFlow) updateProgress(phase, step string, percent float64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.progress.CurrentPhase = phase
	f.progress.CurrentStep = step
	f.progress.Percent = percent
}

// GetProgress returns the current installation progress.
func (f *InstallationFlow) GetProgress() *InstallationProgress {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := *f.progress
	return &result
}

// GetProgressChannel returns a channel for progress updates.
func (f *InstallationFlow) GetProgressChannel() <-chan *InstallProgress {
	return f.installer.GetProgress()
}

// IsComplete checks if the installation is complete.
func (f *InstallationFlow) IsComplete() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.progress.CurrentPhase == "completed" || f.progress.CurrentPhase == "failed"
}

// HasErrors checks if there were errors.
func (f *InstallationFlow) HasErrors() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.progress.Errors) > 0
}

// GetErrors returns the list of errors.
func (f *InstallationFlow) GetErrors() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.progress.Errors
}

// GetWarnings returns the list of warnings.
func (f *InstallationFlow) GetWarnings() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.progress.Warnings
}

// Rollback performs a rollback to the original state.
func (f *InstallationFlow) Rollback(ctx context.Context) error {
	_, err := f.rollback.RollbackToOriginal()
	return err
}
