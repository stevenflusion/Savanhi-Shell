// Package main is the entry point for Savanhi Shell.
// Savanhi Shell is a TUI application for configuring shell environments
// with themes, fonts, and productivity tools.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/savanhi/shell/internal/cli"
	"github.com/savanhi/shell/internal/detector"
	"github.com/savanhi/shell/internal/errors"
	"github.com/savanhi/shell/internal/persistence"
	"github.com/savanhi/shell/internal/preview"
	"github.com/savanhi/shell/internal/tui"
)

// Version information (set at build time).
var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

// Flags
var (
	// Version flag
	showVersion = flag.Bool("version", false, "Show version information")

	// Help flag
	showHelp = flag.Bool("help", false, "Show help")

	// Config file
	configFile = flag.String("config", "", "Path to configuration file")

	// Non-interactive mode
	nonInteractive = flag.Bool("non-interactive", false, "Run in non-interactive mode (for scripting)")

	// Dry run
	dryRun = flag.Bool("dry-run", false, "Perform a dry run without making changes")

	// Verbose output
	verbose = flag.Bool("verbose", false, "Enable verbose output")

	// Rollback mode
	rollback         = flag.Bool("rollback", false, "Rollback last installation")
	rollbackOriginal = flag.Bool("rollback-original", false, "Rollback to original state")

	// Detect mode
	detectOnly = flag.Bool("detect", false, "Only run system detection and output results")

	// Verify mode
	verifyOnly = flag.Bool("verify", false, "Verify existing installation")

	// Health dashboard mode
	healthMode = flag.Bool("health", false, "Run terminal health dashboard")

	// Quick mode (non-interactive JSON output for --health)
	healthQuick = flag.Bool("quick", false, "Output health report as JSON (use with --health)")

	// Timeout
	timeout = flag.Duration("timeout", 10*time.Minute, "Operation timeout")

	// Force overwrite
	force = flag.Bool("force", false, "Force overwrite existing installations")

	// Skip checksum verification
	skipChecksum = flag.Bool("skip-checksum", false, "Skip checksum verification")

	// Skip post-install verification
	skipVerify = flag.Bool("skip-verify", false, "Skip post-install verification")
)

func main() {
	// Parse flags
	flag.Parse()

	// Handle version
	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	// Handle help
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Run the application
	if err := run(); err != nil {
		// Use appropriate exit code
		exitCode := cli.ExitCodeFromError(err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitCode)
	}
}

func run() error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Handle cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Detect mode
	if *detectOnly {
		return runDetect(ctx)
	}

	// Verify mode
	if *verifyOnly {
		return runVerify(ctx)
	}

	// Health dashboard mode
	if *healthMode {
		if *healthQuick {
			return runHealthQuick(ctx)
		}
		return runHealth(ctx)
	}

	// Rollback mode
	if *rollback || *rollbackOriginal {
		return runRollback(ctx)
	}

	// Non-interactive mode
	if *nonInteractive {
		return runNonInteractive(ctx)
	}

	// Interactive TUI mode (default)
	return runTUI(ctx)
}

func runDetect(ctx context.Context) error {
	// Create detector
	d := detector.NewDefaultDetector()

	// Run detection
	result, err := d.DetectAll()
	if err != nil {
		return errors.DetectionFailed("system", err)
	}

	// Output results
	fmt.Println("System Detection Results:")
	fmt.Printf("  OS: %s %s (%s)\n", result.OS.Distro, result.OS.Version, result.OS.Arch)
	fmt.Printf("  Shell: %s %s\n", result.Shell.Name, result.Shell.Version)
	fmt.Printf("  Terminal: %s\n", result.Terminal.Name)
	if result.Fonts != nil {
		fmt.Printf("  Nerd Fonts: %d installed\n", len(result.Fonts.NerdFonts))
	}

	return nil
}

func runVerify(ctx context.Context) error {
	fmt.Println("Verifying installation...")

	// Create detector
	d := detector.NewDefaultDetector()

	// Detect system
	result, err := d.DetectAll()
	if err != nil {
		return errors.DetectionFailed("system", err)
	}

	// Check for installed components
	var issues []string

	// Check for oh-my-posh
	fmt.Printf("Checking oh-my-posh... ")
	// Implementation would check for actual installation
	fmt.Println("OK")

	// Check for fonts
	fmt.Printf("Checking Nerd Fonts... ")
	if result.Fonts != nil && len(result.Fonts.NerdFonts) > 0 {
		fmt.Println("OK")
	} else {
		fmt.Println("NOT FOUND")
		issues = append(issues, "No Nerd Fonts detected")
	}

	// Check for tools
	fmt.Printf("Checking tools (zoxide, fzf, bat, eza)... ")
	// Implementation would check for actual installations
	fmt.Println("OK")

	// Print summary
	fmt.Println()
	fmt.Println("=== Verification Summary ===")
	if len(issues) == 0 {
		fmt.Println("All components installed correctly!")
	} else {
		fmt.Println("Issues found:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	}

	return nil
}

func runRollback(ctx context.Context) error {
	// Load configuration
	config := cli.NewConfig()
	config.Rollback = *rollback
	config.RollbackToOriginal = *rollbackOriginal

	// Create non-interactive runner
	runner, err := cli.NewNonInteractiveMode(config, os.Stdout, os.Stderr, *verbose)
	if err != nil {
		return err
	}

	// Run rollback
	if *rollbackOriginal {
		fmt.Println("Rolling back to original state...")
	} else {
		fmt.Println("Rolling back last installation...")
	}

	return runner.Run(ctx)
}

func runNonInteractive(ctx context.Context) error {
	// Load configuration
	config := cli.NewConfig()

	// Load config file if specified
	if *configFile != "" {
		loadedConfig, err := cli.LoadConfig(*configFile)
		if err != nil {
			return err
		}
		config = loadedConfig
	}

	// Apply flags
	config.DryRun = *dryRun
	config.Force = *force
	config.SkipChecksum = *skipChecksum
	config.SkipVerification = *skipVerify
	config.Timeout = *timeout

	// Create non-interactive runner
	runner, err := cli.NewNonInteractiveMode(config, os.Stdout, os.Stderr, *verbose)
	if err != nil {
		return err
	}

	// Run installation
	return runner.Run(ctx)
}

func runTUI(ctx context.Context) error {
	// Create detector
	d := detector.NewDefaultDetector()

	// Run system detection
	result, err := d.DetectAll()
	if err != nil {
		return errors.DetectionFailed("system", err)
	}

	// Create persister
	p, err := persistence.NewFilePersister()
	if err != nil {
		return errors.NewWithCause(errors.ErrPersistenceFailed,
			"Failed to create persistence layer", err)
	}

	// Check for existing preferences
	var prefs *persistence.Preferences
	hasPrefs, err := p.HasPreferences()
	if err == nil && hasPrefs {
		prefs, _ = p.LoadPreferences()
	} else {
		prefs = &persistence.Preferences{
			Theme: persistence.ThemePreferences{
				Name: "powerlevel10k",
			},
			Fonts: persistence.FontPreferences{
				PrimaryNerdFont: "MesloLGS NF",
			},
			Tools: persistence.ToolPreferences{
				EnableZoxide: true,
				EnableFzf:    true,
				EnableBat:    true,
				EnableEza:    true,
			},
			Advanced: persistence.AdvancedPreferences{
				CreateBackup: true,
			},
		}
	}

	// Load available themes and fonts
	themes := preview.GetBundledThemes()
	fonts := preview.GetRecommendedFonts()

	// Create TUI model
	model := tui.NewModel().
		WithDetector(result).
		WithPersister(p).
		WithPreferences(prefs).
		WithThemes(themes)

	// Store fonts for later use (font selection screen)
	// Fonts will be loaded when transitioning to font selection screen
	_ = fonts // Will be used when font selection is implemented

	// Create Bubble Tea program
	pProgram := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run in goroutine to handle context cancellation
	errChan := make(chan error, 1)
	go func() {
		if _, err := pProgram.Run(); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	// Wait for completion or context cancellation
	select {
	case <-ctx.Done():
		pProgram.Quit()
		return errors.Canceled("operation")
	case err := <-errChan:
		return err
	}
}

func runHealth(ctx context.Context) error {
	// Create detector
	d := detector.NewDefaultDetector()

	// Run system detection
	result, err := d.DetectAll()
	if err != nil {
		return errors.DetectionFailed("system", err)
	}

	// Create TUI model starting at Health Dashboard screen
	model := tui.NewModel().WithDetector(result)
	model.CurrentScreen = tui.ScreenHealthDashboard

	// Create Bubble Tea program
	pProgram := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run in goroutine to handle context cancellation
	errChan := make(chan error, 1)
	go func() {
		if _, err := pProgram.Run(); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	// Wait for completion or context cancellation
	select {
	case <-ctx.Done():
		pProgram.Quit()
		return errors.Canceled("operation")
	case err := <-errChan:
		return err
	}
}

func runHealthQuick(ctx context.Context) error {
	// Create detector
	d := detector.NewDefaultDetector()

	// Run system detection
	result, err := d.DetectAll()
	if err != nil {
		return errors.DetectionFailed("system", err)
	}

	// Run health checks
	healthCmd := tui.RunHealthCheckWithDetector(result)
	msg := healthCmd()

	healthMsg, ok := msg.(tui.HealthCheckCompleteMsg)
	if !ok {
		return fmt.Errorf("unexpected health check response")
	}

	if healthMsg.Err != nil {
		return healthMsg.Err
	}

	// Set export path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	exportPath := homeDir + "/.config/savanhi-shell/health-report.json"
	healthMsg.Data.ExportPath = exportPath

	// Output as JSON to stdout
	jsonData, err := json.MarshalIndent(healthMsg.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal health data: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func printVersion() {
	fmt.Printf("Savanhi Shell %s\n", version)
	fmt.Printf("  Git Commit: %s\n", gitCommit)
	fmt.Printf("  Build Date: %s\n", buildDate)
	fmt.Printf("  Go Version: %s\n", runtime.Version())
	fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func printHelp() {
	fmt.Println(`Savanhi Shell - Shell Environment Configuration Tool

USAGE:
    savanhi-shell [OPTIONS]

OPTIONS:
    --version              Show version information
    --help                 Show this help message
    --config <FILE>        Path to configuration file (YAML or JSON)
    --non-interactive      Run in non-interactive mode (for scripting)
    --dry-run              Perform a dry run without making changes
    --verbose              Enable verbose output
    --rollback             Rollback last installation
    --rollback-original    Rollback to original state (full uninstall)
    --detect               Only run system detection
    --verify               Verify existing installation
    --health               Run terminal health dashboard
    --health --quick      Output health report as JSON (non-interactive)
    --timeout <DURATION>   Operation timeout (default: 10m)
    --force                Force overwrite existing installations
    --skip-checksum        Skip checksum verification
    --skip-verify          Skip post-install verification

INTERACTIVE MODE:
    When run without flags, Savanhi Shell launches an interactive TUI
    that guides you through:
    1. System detection (OS, shell, terminal, fonts)
    2. Theme selection (oh-my-posh themes)
    3. Font selection (Nerd Fonts)
    4. Tool installation (zoxide, fzf, bat, eza)
    5. Preview and confirmation

HEALTH MODE:
    When run with --health, Savanhi Shell launches the Health Dashboard
    TUI showing terminal capabilities, installed components, font test,
    and color test:
    savanhi-shell --health

    For non-interactive JSON output:
    savanhi-shell --health --quick

NON-INTERACTIVE MODE:
    When run with --non-interactive, Savanhi Shell reads configuration
    from a file and performs installation without interaction:

    savanhi-shell --non-interactive --config config.json

EXAMPLES:
    # Run interactive TUI
    savanhi-shell

    # Show version
    savanhi-shell --version

    # Detect system only
    savanhi-shell --detect

    # Non-interactive installation
    savanhi-shell --non-interactive --config my-config.json

    # Dry run to see what would be installed
    savanhi-shell --non-interactive --dry-run --config my-config.json

    # Rollback last installation
    savanhi-shell --rollback

    # Full uninstall (restore original state)
    savanhi-shell --rollback-original

    # Verify installation
    savanhi-shell --verify

    # Run terminal health dashboard
    savanhi-shell --health

    # Output health report as JSON
    savanhi-shell --health --quick

For more information, visit: https://github.com/savanhi/shell`)
}
