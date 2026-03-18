# Architecture Overview

This document provides a technical overview of Savanhi Shell's architecture, module structure, and data flow.

## Table of Contents

- [High-Level Architecture](#high-level-architecture)
- [Module Descriptions](#module-descriptions)
- [Data Flow](#data-flow)
- [Key Interfaces](#key-interfaces)
- [Design Decisions](#design-decisions)

## High-Level Architecture

Savanhi Shell follows a clean, layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Entry Point                                    │
│                    cmd/savanhi-shell/main.go                            │
│                    (CLI parsing, mode routing)                           │
└───────────────────────────┬─────────────────────────────────────────────┘
                            │
┌───────────────────────────┼─────────────────────────────────────────────┐
│                       Presentation Layer                                 │
├───────────────────────────┼─────────────────────────────────────────────┤
│   ┌───────────────────┐   │   ┌───────────────────────────────────┐     │
│   │   Interactive     │   │   │      Non-Interactive              │     │
│   │   (TUI Mode)       │   │   │      (CLI Mode)                   │     │
│   │   internal/tui    │   │   │      internal/cli                 │     │
│   └─────────┬─────────┘   │   └─────────────┬─────────────────────┘     │
│             │               │                 │                           │
└─────────────┼───────────────┼─────────────────┼───────────────────────────┘
              │               │                 │
              └───────────────┴────────┬────────┘
                                       │
┌──────────────────────────────────────┼──────────────────────────────────┐
│                          Core Layer                                    │
├──────────────────────────────────────┼──────────────────────────────────┤
│  ┌─────────────────┐    ┌────────────┴───────────┐    ┌──────────────┐  │
│  │   Detector      │    │      Installer          │    │   Preview    │  │
│  │ internal/detector│    │   internal/installer   │   │internal/preview│
│  └────────┬────────┘    └────────────┬────────────┘   └───────┬──────┘  │
│           │                        │                          │         │
│  ┌────────┴────────┐    ┌─────────┴───────────┐    ┌─────────┴─────┐   │
│  │    Staging       │    │      Rollback       │    │    Tools      │   │
│  │ internal/staging │    │ internal/installer  │    │internal/installer│
│  └──────────────────┘    └─────────────────────┘    └───────────────┘   │
└───────────────────────────────────────────────────────────────────────────┘
                                       │
┌──────────────────────────────────────┼──────────────────────────────────┐
│                       Infrastructure Layer                              │
├──────────────────────────────────────┼──────────────────────────────────┤
│  ┌─────────────────┐    ┌────────────┴───────────┐    ┌──────────────┐  │
│  │  Persistence     │    │     Shell Manipulation │   │   Errors     │  │
│  │internal/persistence│   │      pkg/shell        │   │internal/errors│
│  └─────────────────┘    └─────────────────────────┘   └──────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Module Descriptions

### Entry Point (`cmd/savanhi-shell/`)

**Purpose**: Application entry point and command-line interface handling.

**Responsibilities**:
- Parse command-line flags and arguments
- Route to appropriate execution mode (TUI, non-interactive, detect, verify, rollback)
- Handle version and help output
- Create dependency injection graph
- Manage graceful shutdown

**Key Files**:
- `main.go` - Entry point with flag parsing and mode routing

### Presentation Layer

#### Interactive TUI (`internal/tui/`)

**Purpose**: Bubble Tea-based terminal user interface for interactive configuration.

**Responsibilities**:
- Render screens (welcome, detection, theme selection, font selection, preview, install, complete)
- Handle user input and navigation
- Display progress and status updates
- Coordinate with detector and installer

**Key Components**:
- `model.go` - Bubble Tea model with screen state
- `view.go` - Screen rendering logic
- `update.go` - Message handling and state transitions
- `keys.go` - Key binding definitions
- `views/` - Individual screen implementations

#### Non-Interactive CLI (`internal/cli/`)

**Purpose**: Command-line interface for scripted/automated execution.

**Responsibilities**:
- Parse configuration files (JSON)
- Execute installation unattended
- Output progress and results
- Support CI/CD integration

**Key Components**:
- `config.go` - Configuration struct and loading
- `noninteractive.go` - Non-interactive execution engine
- `exitcodes.go` - Exit code definitions

### Core Layer

#### Detector (`internal/detector/`)

**Purpose**: System detection capabilities for OS, shell, terminal, and fonts.

**Responsibilities**:
- Detect operating system (macOS, Linux, Windows/WSL)
- Identify current shell (zsh, bash)
- Detect terminal emulator
- Inventory installed fonts (especially Nerd Fonts)
- Find existing configurations (oh-my-posh, starship, etc.)

**Key Components**:
- `detector.go` - Main detector interface and composition
- `os.go` - OS detection
- `shell.go` - Shell detection
- `terminal.go` - Terminal detection
- `fonts.go` - Font inventory
- `config.go` - Configuration detection

#### Installer (`internal/installer/`)

**Purpose**: Dependency installation and shell configuration.

**Responsibilities**:
- Download and install oh-my-posh
- Install Nerd Fonts
- Install productivity tools (zoxide, fzf, bat, eza)
- Modify shell RC files (`.zshrc`, `.bashrc`)
- Handle installation flow and progress

**Key Components**:
- `installer.go` - Core installer interface
- `flow.go` - Installation orchestration
- `tools.go` - Tool installation logic
- `fonts.go` - Font installation
- `ohmyposh.go` - oh-my-posh installation
- `rcmodifier.go` - RC file modification
- `verify.go` - Post-install verification
- `rollback.go` - Rollback management
- `resolver.go` - Dependency resolution

#### Preview (`internal/preview/`)

**Purpose**: Live preview functionality for themes and configurations.

**Responsibilities**:
- Create preview environment
- Spawn subshell with modified environment
- Apply theme temporarily
- Cleanup after preview exit

**Key Components**:
- `preview.go` - Preview orchestration
- `session.go` - Preview session management
- `environment.go` - Environment setup
- `cleanup.go` - Resource cleanup

#### Staging (`internal/staging/`)

**Purpose**: Staged change application with atomic commits.

**Responsibilities**:
- Queue pending changes
- Apply changes atomically
- Rollback failed changes
- Track change history

### Infrastructure Layer

#### Persistence (`internal/persistence/`)

**Purpose**: Data persistence for backups, preferences, and history.

**Responsibilities**:
- Manage configuration directory (`~/.config/savanhi/`)
- Store user preferences
- Create and manage backups
- Track installation history
- Handle preview sessions

**Key Components**:
- `persistence.go` - Main persistence interface
- `prefs.go` - Preferences management
- `backup.go` - Backup operations
- `history.go` - History tracking
- `types.go` - Data structures

#### Shell Manipulation (`pkg/shell/`)

**Purpose**: Cross-shell RC file manipulation.

**Responsibilities**:
- Detect and parse shell RC files
- Add managed sections with markers
- Remove managed sections
- Cross-shell compatibility (zsh, bash)

**Key Components**:
- `shell.go` - Shell interface
- `bash.go` - Bash implementation
- `zsh.go` - Zsh implementation
- `markers.go` - Section markers

#### Errors (`internal/errors/`)

**Purpose**: Structured error handling with codes and causes.

**Responsibilities**:
- Define error codes
- Wrap errors with context
- Provide user-friendly messages
- Track error origins

## Data Flow

### Interactive Installation Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     Interactive Installation Flow                        │
└─────────────────────────────────────────────────────────────────────────┘

User starts application
         │
         ▼
┌─────────────────┐
│  main.go:run()  │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│     runTUI()        │
│  Create detector    │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│   detector.Detect   │
│      All()          │──────► Detect OS, Shell, Terminal, Fonts
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Create TUI Model   │
│  With detector      │
│  result             │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐     ┌─────────────────────┐
│  Bubble Tea Event   │────►│  Screen Navigation │
│       Loop          │     │  Welcome → Detect   │
└─────────────────────┘     │  → Theme → Font    │
                            │  → Preview → Install│
                            └─────────────────────┘
         │
         ▼
┌─────────────────────┐
│  User Selection     │
│  Theme, Font, Tools │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│   Preview Screen    │
│  Create preview     │──────► preview.CreateSession()
│  session            │        Apply theme temporarily
└────────┬────────────┘        Spawn subshell
         │
         ▼
┌─────────────────────┐
│   Install Screen    │
│  Create backup      │──────► persistence.CreateBackup()
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Installation       │──────► installer.Install()
│   - oh-my-posh      │        └── download.go
│   - fonts           │        └── tools.go
│   - tools           │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Modify RC Files    │──────► rcmodifier.ModifyRC()
│  Add PATH           │        Add managed sections
│  Add evals          │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│     Verification    │──────► verifier.Verify()
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Complete Screen    │
│  Show results       │
└─────────────────────┘
```

### Non-Interactive Installation Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                   Non-Interactive Installation Flow                     │
└─────────────────────────────────────────────────────────────────────────┘

User runs: savanhi-shell --non-interactive --config config.json
         │
         ▼
┌─────────────────────┐
│  main.go:run()     │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ runNonInteractive() │
│ Load config.json    │──────► cli.LoadConfig()
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ NonInteractiveMode  │
│    Run()            │
└────────┬────────────┘
         │
         ├──────► Create backup (if enabled)
         │
         ├──────► Install oh-my-posh
         │
         ├──────► Install fonts
         │
         ├──────► Install tools
         │
         ▼
┌─────────────────────┐
│      Success/       │
│      Failure        │
└─────────────────────┘
```

### Rollback Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Rollback Flow                                   │
└─────────────────────────────────────────────────────────────────────────┘

User runs: savanhi-shell --rollback
         │
         ▼
┌─────────────────────┐
│   runRollback()     │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Load original       │──────► persistence.LoadOriginalBackup()
│ backup              │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Remove managed      │──────► rcmodifier.RemoveManagedSections()
│ sections from RC   │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Restore original    │──────► backup restoration
│ RC files            │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Update history      │──────► persistence.AppendHistory()
└─────────────────────┘
```

## Key Interfaces

### Detector Interface

```go
type Detector interface {
    DetectOS() (*OSInfo, error)
    DetectShell() (*ShellInfo, error)
    DetectTerminal() (*TerminalInfo, error)
    DetectFonts() (*FontInventory, error)
    DetectExistingConfigs() (*ConfigSnapshot, error)
    DetectAll() (*DetectorResult, error)
}
```

### Persister Interface

```go
type Persister interface {
    // Original backup operations
    HasOriginalBackup() (bool, error)
    SaveOriginalBackup(snapshot *detector.DetectorResult, rcContents map[string]string) error
    LoadOriginalBackup() (*OriginalBackup, error)

    // Preferences operations
    HasPreferences() (bool, error)
    SavePreferences(prefs *Preferences) error
    LoadPreferences() (*Preferences, error)
    ResetPreferences() (*Preferences, error)

    // History operations
    AppendHistory(entry *HistoryEntry) error
    LoadHistory(limit int) ([]*HistoryEntry, error)
    ClearHistory() error

    // Backup operations
    CreateBackup(description string, files []string) (*Backup, error)
    ListBackups() ([]*Backup, error)
    LoadBackup(id string) (*Backup, error)
    RestoreBackup(id string) error
    DeleteBackup(id string) error

    // Preview session operations
    CreatePreviewSession(theme string, rcBackup string) (*PreviewSession, error)
    GetPreviewSession() (*PreviewSession, error)
    EndPreviewSession() error
}
```

### Shell Interface

```go
type Shell interface {
    Name() string
    GetRCPath() (string, error)
    ReadRC() (string, error)
    WriteRC(content string) error
    BackupRC() (string, error)
    RestoreRC(backupPath string) error
    GetEnv() map[string]string
    GetPath() ([]string, error)
    SetPath(paths []string) error
}
```

## Design Decisions

### Why Bubble Tea for TUI?

Bubble Tea provides:
- **Declarative UI**: Clean separation of state and rendering
- **Event-driven**: Natural handling of user input and async operations
- **Composable**: Easy to compose complex UIs from simple components
- **Testable**: Pure functions make testing straightforward

### Why File-Based Persistence?

File-based persistence (vs. database) was chosen because:
- **No dependencies**: No external services required
- **Human-readable**: JSON files can be inspected and edited
- **Portable**: Easy to backup and transfer
- **Simple**: Adequate for single-user local data

### Why Managed Sections in RC Files?

Using marked sections (e.g., `# >>> savanhi-oh-my-posh >>>`) allows:
- **Safe updates**: Can modify without affecting user customizations
- **Clean removal**: Complete rollback without leaving artifacts
- **No merge conflicts**: Clear boundaries prevent conflicts

### Why Staged Installation?

The staging system (`internal/staging/`) provides:
- **Atomicity**: All-or-nothing changes
- **Preview**: Show what will change before applying
- **Rollback**: Easy recovery from failures

### Error Handling Strategy

Errors use structured error types (`internal/errors/`) with:
- **Error codes**: Machine-readable classification
- **Causes**: Wrapped underlying errors
- **Context**: User-friendly messages
- **Exit codes**: Consistent CLI exit codes

## Testing Strategy

The project follows a three-tier testing approach:

1. **Unit Tests**: Each module has `_test.go` files testing individual functions
2. **Integration Tests**: Test module interactions (internal package tests)
3. **E2E Tests**: Full workflow tests in `tests/e2e/`

## See Also

- [Getting Started](getting-started.md) - Quick start guide
- [Configuration](configuration.md) - Configuration options
- [Troubleshooting](troubleshooting.md) - Common issues