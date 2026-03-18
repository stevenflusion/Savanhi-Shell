# PRD: Savanhi Shell Installer

**One command. Any theme. Any operating system. The Oh My Posh ecosystem: better and easier than ever.**

**Version**: 0.1.0-draft
**Author**: Steven Sanchez
**Date**: 2026-03-19
**Status**: Draft

---

## 1. Problem Statement

A configured terminal in 2026 is no longer optional—it's the standard. Every developer uses at least one CLI in their daily work. But here's the real problem:

**Currently, the installers for CLI themes and configurations are not very user-friendly. Trying to configure them is where everyone REALLY fails.**

Using Oh My Posh is still very difficult; it's like using a car without any adjustments—it works, but it's nowhere near its potential. To get the best settings, you need:

1. **Settings** — Configure a shell as your terminal's primary CLI.
2. **Themes** — Thousands of themes, but which ones actually provide a user-friendly interface?
3. **Fonts** — The same applies to fonts, but which one helps you in your day-to-day life as a developer?
4. **Icons** — Icons help you simplify directories and understand where you are located.
5. **Colors** — Undoubtedly, a developer should always have a good color palette in their daily work.
6. **Light or Dark** — Appearance is always important as a developer when you have different development schedules.
7. **Data Controls** — Keeping track of changes or modifications to our shell will save our skin.

Most developers either:

- Manually install a font, a theme, or simply configure your shell when you start the race.
- Spend DAYS manually configuring one settings, then can't replicate it on another machine or tool.
- Never set up memory, advanced settings or simply explore more topics for fear of damaging everything.

**This installer completely eliminates that gap.** Select your theme, font, or color, choose your configuration level, and the entire ecosystem of your favorite shell integrates seamlessly into your daily life, ready to use. From scratch to professional-level development in minutes.

---

## 2. Vision

**The Savanhi Shell ecosystem: installable by anyone, on any shell, on any operating system, with a single command.**

This is NOT a completely new CLI installer. Most terminals are already easy to install (`PowerShell`, `Zsh`, `Fish`, etc.). This is an **ecosystem configurator**: it takes any terminal shell you use and enhances it with the Savanhi toolset:

- **Data Controls**: memory management between terminals.

- **T.F.I.**: The best visual tweaks you can find to configure your shell.

- **Settings**: enhance and get the most out of your terminal shell while coding.

- **Autocomplete**: suggestions for previously used commands.

- **Persona & Settings**: security-prioritizing permissions, coding-oriented personality, themes.

**Before**: "I installed PowerShell / Zsh / Fish / whatever, but it's just a terminal shell."

**After**: `curl -sL get.savanhi.shell/shell | sh` → Select your font(s) → Select your configuration → Your shell terminal now has memory, theme, colors, fonts, and a utility that will actually help you. The same ecosystem regardless of the tool you use.

-

## 3. Target Users

### Primary

- **Professional developers** who want to seriously adopt configuration tools in their terminal, not just experiment with them.

- **Teams** that need a configured terminal for the development and work of all their members.

- **Developers who switch machines** and need to quickly replicate their shell terminal environment.

### Secondary

- **Students** who learn to program with terminal assistance.

- **DevOps/Platform Engineers** who automate the provisioning of CLI tools such as Git.

## 4. Supported Platforms

| Platform              | Package Manager        | Shells    | Priority |
| --------------------- | ---------------------- | --------- | -------- |
| macOS (Apple Silicon) | Homebrew               | zsh, fish | P0       |
| macOS (Intel)         | Homebrew               | zsh, fish | P0       |
| Linux - Ubuntu/Debian | apt + Homebrew         | bash, zsh | P0       |
| Linux - Arch          | pacman                 | bash, zsh | P0       |
| Linux - Fedora/RHEL   | dnf                    | bash, zsh | P1       |
| WSL 2 (Windows)       | apt + Homebrew         | bash, zsh | P1       |
| Windows (native)      | winget / scoop / choco | pwsh, cmd | P2       |
| Termux (Android)      | pkg                    | bash, zsh | P2       |

---

## 5. Prerequisites and Dependency Management

The installer MUST automatically install all prerequisites. A user with a **clean machine** should be able to run the installer and have everything work correctly, without needing to manually run `brew install node` beforehand.

### 5.0.1 Dependency Resolution Strategy

The installer follows a **dependency-first** approach:

1. **Detects** which terminal shells are already installed and their versions.

2. **Calculates** which themes and configurations can be applied based on the user's selections.

3. **Displays** the complete configuration tree BEFORE installing anything.

4. **Installs** dependencies first and then ecosystem configurations.

5. **Verifies** each configuration after installation.

```
┌──────────────────────────────────────────────────────────────────┐
│  CONFIGURATIONS TREE (shown to user before install)              │
│                                                                  │
│  Base tools:                                                     │
│    ✓ shell (already installed: *.**.*)                           │
│    ✓ settings (previous configurations)                          │
     ✓ Color scheme (Detect visual settings)                       │
│    ◌ Homebrew (will install)                                     │
│                                                                  │
│  Runtimes (needed by selected agents):                           │
│    ◌ Node.js 20 (needed by: for future installation)             │
│    ✓ Go 1.25 (already installed — not needed for binary installs)│
│                                                                  │
│  Detection and storage:                                          │
│    ◌ save (visual settings)                                      │
│    ◌ save (settings before installation)                         │
│                                                                  │
│  Configurations:                                                 │
│    ◌ THEMES (Choose your favorite theme)                         │
│    ◌ FONTS (Choose your favorite font)                           │
     ◌ COLORS (Choose your favorite color scheme)                  │
│    ◌ CONFIGURATIONS (Save all settings before installation)      │
│    ◌ ICONS (Choose from the five best icon libraries for shell)  │
│    ◌ DARK or LIGHT (Select the appearance)                       │
│    ◌ DATA CONTROL (Check your settings before installation)      │
└──────────────────────────────────────────────────────────────────┘
```

### 5.0.2 System-Level Dependencies

These are the base tools the installer itself and the ecosystem need.

#### Always Required

| Dependency | Min Version | Why | Install Method |
|-----------|-------------|-----|----------------|
| `bash` | 3.2+ | Install scripts, shell detection, preview execution | Pre-installed on all targets |
| `git` | 2.x | Clone themes, fonts, dotfiles, version control for configs | `brew`/`apt`/`pacman`/`dnf`/`pkg` |
| `curl` | Any | Binary downloads, font downloads, installer bootstrap | Pre-installed on most systems |
| `go` | 1.21+ | Build Savanhi Shell from source (optional binary install) | `brew`/`apt`/manual |

### Conditionally Required (based on user's selections)

| Dependency | Min Version | When Needed | Install Method |
|-----------|-------------|-------------|----------------|
| **oh-my-posh** | 19.0+ | User selects any theme (Agnoster, Paradox, etc.) | Direct binary download / `brew` / `winget` |
| **Nerd Fonts** | Latest | User wants icons, ligatures, or any theme with symbols | Font download + system font installer |
| **zoxide** | 0.9+ | User enables "smart cd" / autojump feature | `brew install zoxide` / `cargo install` / direct binary |
| **fzf** | 0.44+ | User enables fuzzy finder for history/command completion | `brew install fzf` / `apt install fzf` / git clone |
| **bat** | 0.24+ | User enables syntax-highlighted file preview | `brew install bat` / `apt install bat` / `cargo install` |
| **eza** | 0.18+ | User wants better `ls` with icons and git integration | `brew install eza` / `cargo install` / direct binary |
| **lsd** | 1.0+ | Alternative to eza, user choice in settings | `brew install lsd` / `cargo install` / direct binary |
| **zsh-autosuggestions** | Latest | Zsh users enable command autocomplete | `brew install zsh-autosuggestions` / git clone to `$ZSH_CUSTOM` |
| **zsh-syntax-highlighting** | Latest | Zsh users enable syntax highlighting | `brew install zsh-syntax-highlighting` / git clone |
| **Homebrew** | Any | macOS (primary package manager), Linux (optional) | Official install script from brew.sh |
| **PowerShell** | 7.4+ | Windows users or cross-platform PowerShell config | `winget install Microsoft.PowerShell` / `brew install powershell` |
| **starship** | 1.17+ | Alternative prompt engine (if user prefers over oh-my-posh) | `brew install starship` / `cargo install` / direct binary |
| **fig** / **Amazon Q** | Any | User wants IDE-style autocomplete in terminal | App Store / direct download |

### Platform-Specific Notes

| Platform | Pre-installed | Needs Installation | Special Handling |
|----------|--------------|-------------------|------------------|
| **macOS** | bash 3.2, curl, git (if Xcode CLT), shasum | Homebrew, oh-my-posh, Nerd Fonts | `xcode-select --install` for git; `shasum` (not `sha256sum`); BSD sed requires `-i ''`; Font Book for font validation |
| **Ubuntu/Debian** | bash, curl, git, sha256sum, fc-list | Homebrew (optional), Nerd Fonts, oh-my-posh | Fonts go to `~/.local/share/fonts/`; `fc-cache -fv` to refresh; apt versions of fzf may be outdated |
| **Arch** | bash, curl, git, python3, sha256sum, pacman | oh-my-posh (AUR), Nerd Fonts (AUR) | AUR packages (`yay -S oh-my-posh-bin`); rolling release keeps packages current; manual font installation |
| **Fedora/RHEL** | bash, curl, git, sha256sum | oh-my-posh (via install script), Nerd Fonts | May need `dnf copr` for some packages; SELinux context for fonts |
| **WSL 2** | Same as host Linux distro | Same as Linux + Windows Terminal integration | Windows-side fonts vs WSL fonts; Windows Terminal settings.json editing; Path translation (`/mnt/c/`) |
| **Windows native** | None guaranteed | Everything: Git for Windows, Nerd Fonts, PowerShell, oh-my-posh | Git Bash provides bash; Windows Terminal for modern experience; Font installation requires admin rights; Registry editing for default shell |
| **Termux (Android)** | bash, curl, git, pkg | oh-my-posh (if available), Nerd Fonts (manual) | No sudo available; Proot may be needed for Go compilation; Limited storage; Manual font installation to `~/.termux/font.ttf` |

### 5.0.3 Core Dependencies Version Management

Oh My Posh and Nerd Fonts are the most critical dependencies — themes won't render correctly without proper versions, and distro-packaged versions are often outdated or incomplete.

**Strategy:**

| Scenario | Action |
|----------|--------|
| Oh My Posh 19+ already installed | Use it. Skip download. |
| Oh My Posh installed but < 19 | Warn the user. Offer to upgrade to v19+ or use existing version with limited theme support. |
| Oh My Posh not installed | Download latest binary directly from GitHub releases (preferred) or install via package manager |
| Nerd Fonts already installed | Detect installed fonts via `fc-list` (Linux) or Font Book validation (macOS) |
| Nerd Fonts not installed | Download selected font from Nerd Fonts releases, install to user font directory |
| Go already installed (1.21+) | Offer "build from source" option for advanced users |
| Go not installed | Use pre-compiled Savanhi Shell binary (default) — no Go installation needed |

**Requirements:**

- **R-DEP-01**: The installer MUST detect all required dependencies and their versions BEFORE starting installation
- **R-DEP-02**: The installer MUST show the complete dependency tree to the user and get confirmation before installing anything
- **R-DEP-03**: The installer MUST install missing dependencies automatically (with user consent) using the platform's preferred method (Homebrew, direct binary download, package manager)
- **R-DEP-04**: The installer MUST handle oh-my-posh version requirements intelligently — v19+ required for modern themes, v18+ for basic themes
- **R-DEP-05**: The installer MUST NOT require Go installation unless the user explicitly chooses to build Savanhi Shell from source (pre-compiled binaries are the default)
- **R-DEP-06**: On Linux, the installer MUST NOT use distro-packaged Nerd Fonts if they are outdated or incomplete — prefer direct download from Nerd Fonts releases
- **R-DEP-07**: The installer MUST handle platform-specific differences transparently (BSD sed vs GNU sed, sha256sum vs shasum, fc-list vs Font Book)
- **R-DEP-08**: The installer MUST detect existing font managers and use appropriate installation methods (fontconfig on Linux, Font Book on macOS, registry on Windows)
- **R-DEP-09**: If a dependency installation fails, the installer MUST show a clear error with manual installation instructions and continue with other components
- **R-DEP-10**: The installer MUST NOT require root/sudo for dependency installation except when absolutely necessary (e.g., system-wide font installation), and MUST explain why when it does
- **R-DEP-11**: Homebrew MUST be offered as an option on macOS, NOT forced. On Linux, the installer SHOULD prefer native package managers (pacman, dnf, apt) where appropriate, falling back to Homebrew only when native packages are unavailable or outdated
- **R-DEP-12**: The installer MUST verify downloaded binaries using checksums (SHA256) from official sources
- **R-DEP-13**: The installer MUST cache downloaded dependencies in `~/.cache/savanhi/` to avoid re-downloading on subsequent runs
- **R-DEP-14**: The installer MUST provide a "dry-run" mode that shows what WOULD be installed without making any changes

### 5.0.4 Component → Dependency Matrix

| Component | bash | git | curl | oh-my-posh | Nerd Fonts | Homebrew | zoxide | fzf | bat | eza |
|-----------|------|-----|------|------------|------------|----------|--------|-----|-----|-----|
| **Core TUI** (Bubble Tea) | — | — | — | — | — | — | — | — | — | — |
| **System Detection** | ✓ | ✓ | ✓ | ◌ (check) | ◌ (check) | ◌ | ◌ (check) | ◌ (check) | ◌ (check) | ◌ (check) |
| **Preview Engine** | ✓ | — | — | ✓ (preview) | ◌ (preview) | — | — | — | — | — |
| **Theme Installation** | — | — | ✓ (download) | ✓ (install) | — | ◌ | — | — | — | — |
| **Font Installation** | ✓ | — | ✓ (download) | — | ✓ (install) | ◌ (macOS) | — | — | — | — |
| **zoxide Integration** | ✓ | — | — | — | — | ◌ | ✓ (install) | — | — | — |
| **fzf Integration** | ✓ | — | — | — | — | ◌ | — | ✓ (install) | — | — |
| **bat Integration** | ✓ | — | — | — | — | ◌ | — | — | ✓ (install) | — |
| **eza/lsd Integration** | ✓ | — | — | — | — | ◌ | — | — | — | ✓ (install) |
| **Backup System** | ✓ | ✓ | — | — | — | — | — | — | — | — |
| **JSON Persistence** | — | — | — | — | — | — | — | — | — | — |

✓ = required, ◌ = optional/conditional, — = not needed

**Notes:**
- **Core TUI**: Pure Go application using Bubble Tea — no external dependencies
- **System Detection**: Uses bash scripts for deep system introspection; git for detecting dotfile repos; curl for checking online connectivity
- **Preview Engine**: Spawns bash subshells with temporary environments; requires oh-my-posh for theme preview; Nerd Fonts optional for icon preview
- **Theme Installation**: curl to download oh-my-posh binary or theme configs; oh-my-posh itself for applying themes
- **Font Installation**: bash scripts for font installation; curl to download font archives; Nerd Fonts for actual font files; Homebrew optional on macOS (`brew install --cask font-jetbrains-mono-nerd-font`)
- **Tool Integrations**: Each tool (zoxide, fzf, bat, eza) requires bash for shell integration scripts; Homebrew optional as install method
- **Backup System**: bash for file operations; git for optional dotfile backup to repository
- **JSON Persistence**: Pure Go implementation using standard library — no external dependencies

---

## 6. Architecture Overview

Savanhi Shell is built as a **Terminal User Interface (TUI)** application in Go, providing a visual, interactive experience for terminal configuration.

### 6.1 Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **TUI Framework** | Bubble Tea (Charm) | Interactive interface with keybindings, menus, forms |
| **Styling** | Lipgloss | Beautiful, consistent terminal UI theming |
| **Configuration** | JSON | User preferences and backup snapshots |
| **Preview Engine** | Subshell execution | Real-time preview without modifying actual shell |
| **Package Management** | Native OS tools | Homebrew, apt, pacman, dnf, winget, etc. |

### 6.2 Module Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Savanhi Shell (Go)                            │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │
│  │   Detector   │  │     TUI      │  │   Preview    │               │
│  │    Module    │→ │   Interface  │→ │    Engine    │               │
│  └──────────────┘  └──────────────┘  └──────────────┘               │
│         ↓                   ↓                ↓                      │
│  • Shell detection    • Theme selector   • Subshell spawn           │
│  • OS detection       • Font picker      • Config injection         │
│  • Current configs    • Color schemes    • Live preview pane        │
│  • Font inventory     • Live preview     • No system changes        │
│                                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │
│  │   Staging    │  │   Installer  │  │ Persistence  │               │
│  │    System    │→ │    Engine    │→ │    Layer     │               │
│  └──────────────┘  └──────────────┘  └──────────────┘               │
│         ↓                   ↓                ↓                      │
│  • Queue changes      • Lazy download    • ~/.config/savanhi/       │
│  • Validate config    • Install deps     • original-backup.json     │
│  • Backup original    • Apply configs    • preferences.json         │
│                       • Verify install   • history.json             │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 7. Real-Time Preview System

The core differentiator of Savanhi Shell is its **preview-before-install** approach. Users see exactly how their terminal will look BEFORE any changes are made to their system.

### 7.1 Preview Mechanism

```
User selects: Theme "Agnoster" + Font "JetBrainsMono Nerd Font"
         ↓
Preview Engine:
  1. Creates isolated subshell process
  2. Injects temporary environment:
     - OHMYPOSH_CONFIG=/tmp/preview/agnoster.omp.json
     - FONT_FAMILY="JetBrainsMono Nerd Font"
     - COLOR_SCHEME=selected_palette
  3. Renders sample prompt with simulated:
     - Git repository status
     - Directory path
     - Execution time
     - Exit codes
     - Icons and symbols
  4. Displays in TUI preview pane
         ↓
User sees live preview without touching their actual shell
```

### 7.2 Preview Pane Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  Savanhi Shell Preview                                    [?] Help  │
├─────────────────────────────────────────────────────────────────────┤
│  Theme: Agnoster                    Font: JetBrainsMono Nerd Font   │
│  Colors: Dracula                    Mode: Dark                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ PREVIEW (Live subshell) ────────────────────────────────────┐  │
│  │                                                               │  │
│  │  ~/projects/my-app on  feature/login via  v20.11.0         │  │
│  │  ❯ git status                                                │  │
│  │  M  src/auth.ts                                              │  │
│  │  A  src/login.tsx                                            │  │
│  │  ❯ _                                                         │  │
│  │                                                               │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  [↑/↓] Navigate  [Tab] Sections  [Enter] Apply  [Esc] Cancel       │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│  [A] Accept & Install    [C] Cancel    [R] Revert to Original      │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.3 Preview Actions

| Action | Key | Behavior |
|--------|-----|----------|
| **Accept** | `A` / `Enter` | Installs all selected components (fonts, themes, configs) |
| **Cancel** | `C` / `Esc` | Discards preview, returns to selector, nothing installed |
| **Revert** | `R` | Restores original configuration from backup |

---

## 8. JSON Persistence System

Savanhi Shell uses a **dual-snapshot JSON architecture** to ensure users never lose their original configuration and can track their customization history.

### 8.1 Configuration Directory Structure

```
~/.config/savanhi/
├── original-backup.json      # Snapshot of config BEFORE first use
├── preferences.json          # Current user preferences & selections
├── history.json              # Log of all modifications with timestamps
└── temp/
    ├── preview-*.json        # Temporary preview configs
    └── subshell-*/           # Isolated preview environments
```

### 8.2 original-backup.json (First-Run Snapshot)

Created automatically the first time Savanhi Shell runs:

```json
{
  "snapshot_version": "1.0",
  "created_at": "2026-03-19T14:30:00Z",
  "system": {
    "os": "macOS",
    "arch": "arm64",
    "shell": "zsh"
  },
  "original_config": {
    "shell_rc": "~/.zshrc",
    "shell_rc_content": "# Full original .zshrc content here...",
    "oh_my_posh": {
      "installed": false,
      "config_path": null
    },
    "fonts": {
      "installed": ["SF Mono", "Menlo"],
      "terminal_font": "SF Mono"
    },
    "environment_variables": {
      "PATH": "/usr/local/bin:/usr/bin...",
      "EDITOR": "vim"
    }
  },
  "checksums": {
    "shell_rc_sha256": "abc123...",
    "fonts_list_sha256": "def456..."
  }
}
```

### 8.3 preferences.json (User Preferences)

Stores current selections and customization state:

```json
{
  "version": "0.1.0",
  "last_modified": "2026-03-19T15:45:00Z",
  "user_profile": {
    "name": "Steven",
    "default_shell": "zsh"
  },
  "savanhi_config": {
    "theme": {
      "name": "agnoster",
      "source": "oh-my-posh",
      "custom_overrides": {}
    },
    "font": {
      "family": "JetBrainsMono Nerd Font",
      "size": 14,
      "ligatures": true
    },
    "colors": {
      "scheme": "dracula",
      "background": "dark",
      "custom_colors": {}
    },
    "icons": {
      "set": "nerd-fonts",
      "show_folder_icons": true
    },
    "features": {
      "zoxide": true,
      "fzf": true,
      "autocomplete": true
    }
  },
  "installed_components": [
    {
      "name": "oh-my-posh",
      "version": "19.0.0",
      "installed_at": "2026-03-19T15:30:00Z"
    },
    {
      "name": "JetBrainsMono Nerd Font",
      "version": "3.1.1",
      "installed_at": "2026-03-19T15:35:00Z"
    }
  ]
}
```

### 8.4 history.json (Modification Log)

Tracks every change made through Savanhi Shell:

```json
{
  "modifications": [
    {
      "timestamp": "2026-03-19T15:30:00Z",
      "action": "install",
      "component": "oh-my-posh",
      "previous_value": null,
      "new_value": "agnoster theme",
      "rollback_possible": true
    },
    {
      "timestamp": "2026-03-19T15:35:00Z",
      "action": "install",
      "component": "font",
      "previous_value": "SF Mono",
      "new_value": "JetBrainsMono Nerd Font",
      "rollback_possible": true
    },
    {
      "timestamp": "2026-03-19T16:00:00Z",
      "action": "change_theme",
      "component": "oh-my-posh",
      "previous_value": "agnoster",
      "new_value": "paradox",
      "rollback_possible": true
    }
  ]
}
```

---

## 9. Lazy Installation Strategy

Savanhi Shell follows a **download-only-on-accept** philosophy. Nothing is installed until the user explicitly accepts a preview.

### 9.1 Dependency Installation Matrix

| Component | Detection Method | Download Size | Install Trigger |
|-----------|-----------------|---------------|-----------------|
| **oh-my-posh** | `which oh-my-posh` | ~20MB binary | User clicks "Accept" |
| **Nerd Fonts** | `fc-list`, macOS Font Book | ~5-15MB per font | User clicks "Accept" |
| **zoxide** | `which zoxide` | ~3MB binary | User enables in features |
| **fzf** | `which fzf` | ~2MB binary | User enables in features |
| **eza/lsd** | `which eza` / `which lsd` | ~2MB binary | User enables icons |
| **bat** | `which bat` | ~5MB binary | User enables file preview |

### 9.2 Installation Flow

```
User Journey:
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│  Launch  │ → │ Detect   │ → │ Preview  │ → │  Accept  │
│  Savanhi │   │ System   │   │ Changes  │   │  Changes │
└──────────┘   └──────────┘   └──────────┘   └──────────┘
     ↓              ↓              ↓              ↓
Read JSON    Check what's    Show preview   Download &
configs      installed       in subshell    install all
             vs requested                   components
```

### 9.3 Revert Mechanism

The **Revert** action restores the system to its original state:

1. Reads `original-backup.json`
2. Uninstalls Savanhi-added components (optional)
3. Restores original shell RC files
4. Removes Savanhi-specific environment variables
5. Preserves user modifications made OUTSIDE Savanhi (merge strategy)

---

## 10. System Detection at Startup

Upon launch, Savanhi Shell performs a comprehensive system scan.

### 10.1 Detection Matrix

| Category | Detected Information | Method |
|----------|---------------------|--------|
| **Operating System** | macOS (Intel/ARM), Linux (distro), Windows, WSL | `uname`, `/etc/os-release`, `$OSTYPE` |
| **Shell** | bash, zsh, fish, pwsh version | `$SHELL`, `$0` |
| **Terminal Emulator** | iTerm2, Terminal.app, Windows Terminal, Alacritty | Environment vars, process tree |
| **Installed Fonts** | All system fonts, Nerd Font detection | `fc-list`, macOS `system_profiler` |
| **Current Theme** | oh-my-posh config, PS1, custom prompts | Config file parsing |
| **Package Managers** | brew, apt, pacman, dnf, winget, scoop | `which` checks |
| **Existing Tools** | git, fzf, zoxide, bat, eza | `which` checks |
| **Color Support** | True color, 256 color, basic | `tput`, `$COLORTERM` |

### 10.2 Detection Output Example

```
┌─────────────────────────────────────────────────────────────┐
│  Savanhi Shell - System Detection                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  System Information:                                        │
│    OS: macOS 14.4 (Sonoma) - Apple Silicon                 │
│    Shell: zsh 5.9                                           │
│    Terminal: iTerm2 3.4.23                                  │
│                                                             │
│  Current Configuration:                                     │
│    Theme: No oh-my-posh detected                            │
│    Font: SF Mono (no Nerd Font)                             │
│    Colors: Basic terminal colors                            │
│                                                             │
│  Detected Tools:                                            │
│    ✓ git 2.42.0                                            │
│    ✓ brew 4.2.0                                            │
│    ✗ oh-my-posh (not installed)                            │
│    ✗ Nerd Fonts (not detected)                             │
│    ✗ fzf (not installed)                                   │
│    ✗ zoxide (not installed)                                │
│                                                             │
│  Recommendations:                                           │
│    → Your terminal could be much cooler!                    │
│    → 5 components available for installation                │
│                                                             │
│              [Continue to Setup]  [Exit]                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘

---

## 11. Components to Install & Configure

Savanhi Shell supports configuring a complete terminal ecosystem. The user selects which components to enhance their shell experience. **The primary job is CONFIGURATION with preview** — users see changes before committing. The installer CAN install missing components, but the core value is the seamless integration and preview system.

### 11.1 Theme Components

| Component | Config Location | What Gets Configured | Priority |
|-----------|-----------------|---------------------|----------|
| **Oh My Posh** | `~/.config/oh-my-posh/` | Full: themes, segments, colors, git integration, execution time, exit codes | P0 |
| **Starship** | `~/.config/starship.toml` | Alternative: cross-shell prompt with similar features | P1 |
| **Custom PS1** | Shell RC files | Fallback: manual prompt configuration for minimal setups | P2 |

### 11.2 Font Components

| Component | Config Location | What Gets Configured | Priority |
|-----------|-----------------|---------------------|----------|
| **Nerd Fonts** | System font directories | Font installation, terminal font configuration, ligatures | P0 |
| **Font Ligatures** | Terminal-specific | Enable/disable ligatures in terminal emulator settings | P1 |
| **Font Size** | Terminal-specific | Terminal font size adjustment | P1 |

### 11.3 Color & Appearance Components

| Component | Config Location | What Gets Configured | Priority |
|-----------|-----------------|---------------------|----------|
| **Color Schemes** | Terminal emulator config | Dracula, Nord, One Dark, Solarized, Catppuccin palettes | P0 |
| **Light/Dark Mode** | Terminal + Shell config | Automatic switching, manual toggle, OS sync | P1 |
| **Syntax Highlighting** | Shell RC files | Command highlighting, path validation, bracket matching | P1 |

### 11.4 Shell Enhancement Tools

| Component | Config Location | What Gets Configured | Priority |
|-----------|-----------------|---------------------|----------|
| **zoxide** | Shell RC files | Smart directory jumping, `z` command, interactive selection | P0 |
| **fzf** | Shell RC files | Fuzzy finder for history, files, git, processes | P0 |
| **zsh-autosuggestions** | `~/.zsh/zsh-autosuggestions/` | Command suggestions based on history | P1 |
| **zsh-syntax-highlighting** | `~/.zsh/zsh-syntax-highlighting/` | Real-time syntax highlighting for commands | P1 |
| **bat** | `~/.config/bat/` | Syntax-highlighted cat replacement, git integration | P1 |
| **eza** | Shell aliases | Modern ls replacement with icons and git support | P1 |
| **lsd** | Shell aliases | Alternative to eza, user preference | P2 |

### 11.5 Icon Components

| Component | Config Location | What Gets Configured | Priority |
|-----------|-----------------|---------------------|----------|
| **Nerd Font Icons** | Via Nerd Fonts | File type icons, folder icons, git status icons | P0 |
| **eza Icons** | Shell config | Directory listing with file type icons | P1 |
| **lsd Icons** | Shell config | Alternative icon display in directory listings | P2 |

### 11.6 Configuration Support Tiers

| Tier | What Gets Configured | Components |
|------|---------------------|------------|
| **Full** | Theme + Fonts + Colors + All Tools + Icons + Shell Enhancements | Oh My Posh, Nerd Fonts, zoxide, fzf, bat, eza, autosuggestions, syntax-highlighting |
| **Essential** | Theme + Fonts + Basic Tools | Oh My Posh, Nerd Fonts, zoxide, fzf |
| **Minimal** | Theme only | Oh My Posh or Starship with basic theme |
| **Custom** | User-selected subset | Any combination of the above |

> **Note:** The preview system works with ANY tier — users see exactly what they're getting before installation.

### 11.7 Shell-Specific Support

| Shell | Supported Components | Special Handling |
|-------|---------------------|------------------|
| **zsh** | Full support: themes, fonts, colors, all tools, plugins | Oh My Zsh integration, custom plugins |
| **bash** | Full support: themes, fonts, colors, most tools | bash-it or pure oh-my-posh |
| **fish** | Full support: themes, fonts, colors, all tools | Fisher plugin manager, fish-native configs |
| **PowerShell** | Partial support: oh-my-posh, fonts, limited tools | Windows Terminal integration, PSReadLine |

**Requirements:**

- R-COMP-01: The installer MUST detect already-installed components and offer configuration or upgrade only
- R-COMP-02: The installer MUST support configuring multiple components in a single session with dependency resolution
- R-COMP-03: The installer MUST show a live preview of each component BEFORE installation
- R-COMP-04: The installer MUST detect the user's existing configurations and offer to preserve, merge, or replace them
- R-COMP-05: The installer MUST clearly show the "Configuration Tier" for each selection so users understand the scope
- R-COMP-06: For components the user doesn't have installed, the installer SHOULD offer to install them with automatic dependency resolution
- R-COMP-07: The installer architecture MUST allow adding new components by implementing a single interface — no changes to TUI or core logic required
- R-COMP-08: The installer MUST be forward-compatible: when new terminal tools emerge, support can be added via the Component interface
- R-COMP-09: The installer MUST handle shell-specific differences transparently (zsh plugins vs fish functions vs bash scripts)
- R-COMP-10: The installer MUST provide sensible defaults based on detected system capabilities (e.g., disable icons if terminal doesn't support them)

### 11.8 Installation Methods

Savanhi Shell supports multiple installation methods to accommodate different user preferences and system constraints.

| Method | Command | Use Case | Priority |
|--------|---------|----------|----------|
| **One-liner** | `curl -sL get.savanhi.shell/shell | sh` | Quick start, clean machines | P0 |
| **Homebrew** | `brew install savanhi/tap/savanhi-shell` | macOS/Linux users who prefer Homebrew | P1 |
| **Go Install** | `go install github.com/savanhi/shell@latest` | Go developers, latest features | P1 |
| **Git Clone** | `git clone + make install` | Developers who want to modify/contribute | P2 |
| **Manual Download** | Download binary from releases | Air-gapped systems, specific versions | P2 |
| **Winget** | `winget install Savanhi.Shell` | Windows users | P2 |
| **Scoop** | `scoop install savanhi-shell` | Windows power users | P2 |

**Installation Flow:**

```
User runs: curl -sL get.savanhi.shell/shell | sh
         ↓
Bootstrap script:
  1. Detect OS and architecture
  2. Download appropriate binary (~15-20MB)
  3. Verify SHA256 checksum
  4. Install to ~/.local/bin or /usr/local/bin
  5. Add to PATH if needed
  6. Launch TUI immediately
         ↓
TUI opens with system detection
```

**Requirements:**

- R-INSTALL-01: The one-liner installer MUST work on all supported platforms without prerequisites (except curl)
- R-INSTALL-02: The installer MUST detect if Savanhi Shell is already installed and offer to update instead of reinstall
- R-INSTALL-03: The installer MUST add the binary to PATH automatically or provide clear instructions
- R-INSTALL-04: The installer MUST verify binary integrity using SHA256 checksums from official releases
- R-INSTALL-05: The installer SHOULD create a desktop entry or menu item on GUI systems (optional)
- R-INSTALL-06: The installer MUST NOT require sudo/root unless installing to system directories
- R-INSTALL-07: The uninstall process MUST be documented and reversible (`savanhi uninstall` or manual steps)

### 11.9 Terminal Emulator Integration

Savanhi Shell integrates with popular terminal emulators to ensure optimal display and functionality.

| Terminal | Integration Level | Configurable Aspects |
|----------|------------------|---------------------|
| **iTerm2** (macOS) | Full | Font, color scheme, status bar, shell integration, badge |
| **Windows Terminal** | Full | Font, color scheme, background, acrylic effects, profiles |
| **Terminal.app** (macOS) | Partial | Font, color scheme, basic profiles |
| **Alacritty** | Full | Font, colors.toml, shell config, hints |
| **Kitty** | Full | Font, colors, shell integration, kittens |
| **WezTerm** | Full | Font, color scheme, tab bar, mux |
| **Hyper** | Partial | Font, color scheme, plugins |
| **GNOME Terminal** | Partial | Font, color profile |
| **Konsole** | Partial | Font, color scheme, profiles |

**Integration Features:**

- **Font Configuration**: Automatically set terminal font to selected Nerd Font
- **Color Schemes**: Install color scheme files to terminal-specific locations
- **Shell Integration**: Enable terminal-specific features (e.g., iTerm2 shell integration, Windows Terminal quake mode)
- **Profile Creation**: Create dedicated "Savanhi" profile with optimized settings

**Requirements:**

- R-TERM-01: The installer MUST detect the active terminal emulator and offer emulator-specific configurations
- R-TERM-02: The installer MUST create a backup of existing terminal configurations before modification
- R-TERM-03: The installer MUST provide a "restore terminal config" option independent of shell config restore
- R-TERM-04: For unsupported terminals, the installer MUST provide manual configuration instructions
- R-TERM-05: The installer SHOULD offer to set the "Savanhi" profile as default (with user confirmation)

### 11.10 User Preferences & Profiles

Savanhi Shell supports user profiles to quickly apply predefined configurations.

#### Built-in Profiles

| Profile | Description | Components |
|---------|-------------|------------|
| **Developer** | Full-featured development environment | Oh My Posh, Nerd Fonts, zoxide, fzf, bat, eza, autosuggestions, syntax highlighting |
| **Minimalist** | Clean and simple | Oh My Posh (basic theme), Nerd Fonts (essential icons only) |
| **Power User** | Everything + advanced features | Full stack + custom aliases, advanced fzf bindings, tmux integration |
| **Corporate** | Security-focused, minimal external deps | Oh My Posh, basic colors, no external binaries |
| **Student** | Learning-friendly with documentation | Developer profile + inline help, command explanations |

#### Preference Categories

| Category | Options | Default |
|----------|---------|---------|
| **Appearance** | Dark / Light / Auto | Auto (follows OS) |
| **Prompt Style** | Powerline / Plain / Minimal | Powerline |
| **Icons** | Full / Minimal / None | Full |
| **Transparency** | Enabled / Disabled | Disabled |
| **Animations** | Enabled / Reduced / None | Reduced |
| **Language** | English / Spanish / Auto | Auto |

#### Custom Profiles

Users can create and save custom profiles:

```json
// ~/.config/savanhi/profiles/my-custom.json
{
  "name": "My Custom",
  "base": "developer",
  "overrides": {
    "theme": "paradox",
    "font": "FiraCode Nerd Font",
    "features": {
      "zoxide": true,
      "fzf": false,
      "bat": true
    },
    "aliases": {
      "ls": "eza --icons",
      "cat": "bat"
    }
  }
}
```

**Requirements:**

- R-PREF-01: The installer MUST offer built-in profiles during the first-run wizard
- R-PREF-02: Users MUST be able to create, save, and load custom profiles
- R-PREF-03: Profile changes MUST be previewable before application
- R-PREF-04: The installer MUST remember the last used profile and offer it as default on subsequent runs
- R-PREF-05: Profiles MUST be exportable/importable as JSON for sharing between machines
- R-PREF-06: The installer SHOULD detect the user's current setup and suggest the closest matching profile

## 12. User Experience

### 12.1 Installation Flow

```
curl -sL get.savanhi.shell/shell | sh
                  │
                  ▼
     ┌─────────────────────┐
     │   Download binary    │
     │   (detect OS/arch)   │
     └──────────┬──────────┘
                │
                ▼
     ┌─────────────────────────────────┐
     │   TUI: Welcome                   │
     │   "Savanhi Shell"                │
     │   Your terminal, supercharged.   │
     └──────────┬──────────────────────┘
                │
                ▼
     ┌─────────────────────────────────┐
     │  System Scan                     │
     │  OS: macOS (Apple Silicon)       │
     │  Shell: zsh 5.9 ✓                │
     │  Terminal: iTerm2 ✓              │
     │  oh-my-posh: not installed ✗     │
     │  Nerd Fonts: not detected ✗      │
     │  Current theme: default ✓        │
     └──────────┬──────────────────────┘
                │
                ▼
     ┌─────────────────────────────────┐
     │  Select Configuration Tier       │
     │                                  │
     │  ★ Developer                     │
     │    Full setup with all tools     │
     │    (Oh My Posh + Fonts + zoxide  │
     │     + fzf + bat + eza)           │
     │                                  │
     │  ○ Minimalist                    │
     │    Clean and simple              │
     │                                  │
     │  ○ Power User                    │
     │    Everything + advanced         │
     │                                  │
     │  ○ Custom                        │
     │    Pick each component           │
     └──────────┬──────────────────────┘
                │
        ┌───────┴───────┐
        │ If "Custom":  │
        │               ▼
        │  ┌──────────────────────┐
        │  │ Select Components:   │
        │  │ ☑ Oh My Posh         │
        │  │ ☑ Nerd Fonts         │
        │  │ ☑ zoxide             │
        │  │ ☑ fzf                │
        │  │ ☐ bat                │
        │  │ ☐ eza                │
        │  └────────┬─────────────┘
        │           │
        └───────┬───┘
                │
                ▼
     ┌─────────────────────────────────┐
     │  Theme Selection                 │
     │                                  │
     │  Select your theme:              │
     │                                  │
     │  ★ Agnoster                      │
     │  ○ Paradox                       │
     │  ○ Powerlevel10k                 │
     │  ○ Custom theme...               │
     │                                  │
     │  [Preview] ← Live preview pane   │
     └──────────┬──────────────────────┘
                │
                ▼
     ┌─────────────────────────────────┐
     │  Font Selection                  │
     │                                  │
     │  Select your font:               │
     │                                  │
     │  ★ JetBrainsMono Nerd Font       │
     │  ○ FiraCode Nerd Font            │
     │  ○ Hack Nerd Font                │
     │  ○ Meslo Nerd Font               │
     │                                  │
     │  [Preview] ← Shows icons & chars │
     └──────────┬──────────────────────┘
                │
                ▼
     ┌─────────────────────────────────┐
     │  Color Scheme                    │
     │                                  │
     │  Select color palette:           │
     │                                  │
     │  ★ Dracula                       │
     │  ○ Nord                          │
     │  ○ One Dark                      │
     │  ○ Catppuccin                    │
     │  ○ Solarized Dark                │
     │                                  │
     │  Mode: ○ Auto  ● Dark  ○ Light   │
     └──────────┬──────────────────────┘
                │
                ▼
     ┌─────────────────────────────────┐
     │  LIVE PREVIEW                    │
     │                                  │
     │  ┌───────────────────────────┐  │
     │  │ ~/projects/my-app         │  │
     │  │ on  feature/login via  │  │
     │  │ ❯ _                       │  │
     │  └───────────────────────────┘  │
     │                                  │
     │  This is how your terminal      │
     │  will look. What do you think?  │
     │                                  │
     │  [A] Accept & Install            │
     │  [C] Cancel & Go Back            │
     │  [R] Revert to Original          │
     └──────────┬──────────────────────┘
                │
                ▼
     ┌─────────────────────────────────┐
     │  Installing...                   │
     │                                  │
     │  ✓ Downloading oh-my-posh       │
     │  ✓ Installing oh-my-posh        │
     │  ✓ Downloading JetBrainsMono    │
     │  ✓ Installing Nerd Font          │
     │  ✓ Configuring zsh               │
     │  ◌ Configuring iTerm2...         │
     │    [████████░░] 85%              │
     └──────────┬──────────────────────┘
                │
                ▼
     ┌─────────────────────────────────┐
     │  Done! Your terminal is ready.   │
     │                                  │
     │  Configuration saved:            │
     │  • Theme: Agnoster               │
     │  • Font: JetBrainsMono Nerd Font │
     │  • Colors: Dracula (Dark)        │
     │  • Tools: zoxide, fzf            │
     │                                  │
     │  Next steps:                     │
     │  1. Restart your terminal        │
     │  2. Try: z Projects              │
     │  3. Try: Ctrl+R (fzf history)    │
     │                                  │
     │  Backup created: ~/.config/      │
     │  savanhi/original-backup.json    │
     └─────────────────────────────────┘
```

### 12.2 Non-Interactive Mode

For CI, automation, and team provisioning:

```bash
# Install with a preset
savanhi install --profile developer --non-interactive

# Custom installation
savanhi install \
  --shell zsh \
  --theme agnoster \
  --font "JetBrainsMono Nerd Font" \
  --colors dracula \
  --tools zoxide,fzf,bat,eza \
  --non-interactive

# Apply saved profile
savanhi apply --profile my-team-config

# Revert to original
savanhi revert

# Export current config
savanhi export --output my-config.json

# Import and apply config
savanhi import --input team-config.json
```

**Requirements:**

- R-UX-01: The installer MUST support both interactive TUI and non-interactive CLI modes
- R-UX-02: The TUI MUST use the Bubbletea framework with Lipgloss styling for consistent UI
- R-UX-03: Installation progress MUST stream real-time logs to the TUI
- R-UX-04: The installer MUST show a live preview BEFORE applying any changes
- R-UX-05: The installer MUST show clear "Next Steps" after completion (restart terminal, first commands)
- R-UX-06: The TUI MUST support vim-style navigation (j/k, Enter, Esc, q to quit)
- R-UX-07: Every step that modifies the system MUST be reversible via backup
- R-UX-08: Non-interactive mode MUST validate all arguments before execution and fail fast with clear errors
- R-UX-09: Non-interactive mode MUST support `--dry-run` to show what would be installed
- R-UX-10: All operations MUST be idempotent (running twice produces the same result)

### 12.3 Screens

| Screen | Purpose | Key Features |
|--------|---------|--------------|
| **Welcome** | Branding, version, introduction | Logo animation, version info, quick start tip |
| **System Detection** | Show detected OS, shell, terminal, current config | Visual indicators (✓/✗), expandable details, recommendations |
| **Profile Selection** | Choose from built-in or custom profiles | Preview icons, profile descriptions, "Create Custom" option |
| **Component Selection** | Pick individual components | Checkboxes with descriptions, dependency highlighting |
| **Theme Selection** | Select oh-my-posh theme | Live preview pane, theme thumbnails, search/filter |
| **Font Selection** | Select Nerd Font | Live preview of icons, font samples, ligature preview |
| **Color Selection** | Choose color scheme and mode | Color swatches, light/dark toggle, terminal preview |
| **Live Preview** | See final result before install | Interactive subshell, Accept/Cancel/Revert buttons |
| **Review** | Summary of all changes | Complete dependency tree, disk space, estimated time |
| **Installing** | Real-time progress | Per-component progress, log output, cancel option |
| **Complete** | Success and next steps | Configuration summary, helpful commands, backup info |
| **Backup Management** | Manage previous configurations | List backups, restore, delete, export |
| **Settings** | Savanhi Shell preferences | Default profile, auto-update, language |

---



## 13. Technical Architecture

### 13.0 Ecosystem Architecture — How Everything Connects

This section describes how all Savanhi Shell components interact with each other, both at **install time** (what the installer does) and at **runtime** (what the user experiences daily).

#### 13.0.1 The Big Picture

**Key architectural components:**

1. **Detection Engine**: Scans OS, shell, terminal, fonts, and existing configs
2. **Preview Engine**: Creates isolated subshells to show live previews without modifying the system
3. **Install Engine**: Downloads, installs, and configures components based on user selections
4. **Persistence Layer**: JSON-based backup and preferences system
5. **Shell Integration**: Configures RC files and shell-specific plugins
6. **Terminal Integration**: Configures terminal emulator settings (fonts, colors, themes)

#### 13.0.2 Preview System — Live Preview Without Installation

This is what happens BEFORE the user commits to any changes:

**Flow:**
1. User selects theme, font, colors, tools
2. Preview Engine creates temporary environment
3. Spawns isolated subshell with selected configs injected
4. Renders live preview in TUI preview pane
5. User sees result and decides: Accept / Cancel / Revert

**Key insight:** The preview uses a REAL shell with REAL configurations, just in an isolated temporary environment. This is not a mock-up — it's the actual experience.

#### 13.0.3 Installation Pipeline — Dependency Resolution Order

**Phase 1: System Detection**
- Detect OS, architecture, shell, terminal emulator
- Scan installed dependencies (git, curl, Homebrew)
- Scan existing shell configs (.zshrc, .bashrc)
- Load existing backup if present

**Phase 2: User Choices**
- Select profile (Developer, Minimalist, Power User, Custom)
- Or pick individual components (theme, font, colors, tools)

**Phase 3: Live Preview**
- Generate live preview in isolated subshell
- Show user the final result
- Wait for decision (Accept / Cancel / Revert)

**Phase 4: Pre-Install Backup**
- Backup existing configs to original-backup.json
- Save current preferences to preferences.json

**Phase 5: Dependencies**
- Check for missing dependencies
- Install Homebrew if needed (macOS/Linux)
- Install git if missing

**Phase 6: Component Installation**
- Download components (oh-my-posh, fonts, tools)
- Install Oh My Posh
- Install Nerd Fonts
- Install tools (zoxide, fzf, bat, eza)
- Configure shell RC files
- Configure terminal emulator

**Phase 7: Verification**
- Health checks for all installed components
- Verify configs are valid
- Confirm installation success

#### 13.0.4 Component Configuration Matrix — What Gets Configured Where

| Component | Configuration Location | What Gets Injected |
|-----------|------------------------|-------------------|
| **Oh My Posh** | `~/.config/oh-my-posh/` | Theme JSON, custom segments |
| **Shell RC** | `~/.zshrc`, `~/.bashrc`, `~/.config/fish/config.fish` | Aliases, exports, tool initialization |
| **Nerd Fonts** | System font directories (`~/.local/share/fonts/`, `~/Library/Fonts/`, `C:\Windows\Fonts\`) | Font files (.ttf, .otf) |
| **zoxide** | Shell RC files | `eval "$(zoxide init zsh)"` |
| **fzf** | Shell RC files | Key bindings, completion |
| **bat** | `~/.config/bat/` | Theme, pager config |
| **Terminal Emulator** | App-specific (iTerm2 plist, Windows Terminal settings.json, Alacritty.yml) | Font, color scheme, profile |
| **Savanhi Configs** | `~/.config/savanhi/` | original-backup.json, preferences.json, history.json |

#### 13.0.5 Runtime Experience — How the User Uses Savanhi Daily

After installation, this is the daily workflow:

1. User opens terminal
2. Shell loads RC file with Savanhi configurations
3. Oh My Posh renders custom prompt with theme
4. Nerd Fonts display icons and ligatures
5. zoxide provides smart directory jumping
6. fzf provides fuzzy finder for history
7. bat provides syntax-highlighted file viewing
8. eza provides icon-enhanced directory listings
9. User can run `savanhi` to reconfigure or revert

**Key architectural principle:** Once installed, Savanhi configs are standard shell configurations. No runtime daemon, no background service — just pure shell customization that works everywhere.

#### 13.0.6 Cross-Shell Synchronization

When a user switches shells (e.g., from bash to zsh), Savanhi maintains consistency:

1. **Shared Preferences**: `preferences.json` is shell-agnostic
2. **Per-Shell Configs**: Each shell gets its own RC file modifications
3. **Component Reuse**: Tools (zoxide, fzf) are installed once, initialized per-shell
4. **Font & Colors**: Terminal-level settings work across all shells

**Backup Strategy:**
- Each shell's original RC is backed up separately
- Restoring can be done per-shell or globally
- JSON configs are portable across machines

---

### 13.1 Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Language | Go 1.21+ | Single binary, cross-compile, no runtime deps, excellent for CLI tools |
| TUI | Bubbletea (Charm) | Elm architecture, proven in production, excellent terminal support |
| Styling | Lipgloss | Beautiful, consistent terminal styling |
| Configuration | JSON | Human-readable, native Go support, easy to backup/restore |
| Shell Scripting | bash 3.2+ | Universal shell support, compatible across platforms |
| Distribution | Homebrew tap + direct binary + curl installer | Same as modern CLI tools |
| Version Control | git | For cloning themes, fonts, updates |

---

### 13.2 Package Structure (Proposed)

```
savanhi-shell/
├── cmd/
│   └── savanhi/
│       └── main.go                 # CLI entrypoint
├── internal/
│   ├── system/
│   │   ├── detect.go               # OS, arch, shell, terminal detection
│   │   ├── exec.go                 # Command execution with logging
│   │   └── deps.go                 # Dependency detection and installation
│   ├── shell/
│   │   ├── shell.go                # Shell interface
│   │   ├── zsh.go                  # Zsh-specific configuration
│   │   ├── bash.go                 # Bash-specific configuration
│   │   ├── fish.go                 # Fish-specific configuration
│   │   └── powershell.go           # PowerShell configuration
│   ├── components/
│   │   ├── component.go            # Component interface
│   │   ├── ohmyposh.go             # Oh My Posh installation and config
│   │   ├── fonts.go                # Nerd Fonts installation
│   │   ├── zoxide.go               # zoxide installation
│   │   ├── fzf.go                  # fzf installation
│   │   ├── bat.go                  # bat installation
│   │   └── eza.go                  # eza installation
│   ├── terminal/
│   │   ├── terminal.go             # Terminal emulator interface
│   │   ├── iterm2.go               # iTerm2 configuration
│   │   ├── windowsterminal.go      # Windows Terminal configuration
│   │   ├── alacritty.go            # Alacritty configuration
│   │   └── kitty.go                # Kitty configuration
│   ├── preview/
│   │   ├── preview.go              # Preview engine
│   │   └── subshell.go             # Isolated subshell management
│   ├── persistence/
│   │   ├── backup.go               # Backup and restore system
│   │   ├── preferences.go          # Preferences JSON management
│   │   └── history.go              # Modification history
│   ├── profiles/
│   │   ├── profile.go              # Profile interface
│   │   ├── developer.go            # Developer profile
│   │   ├── minimalist.go           # Minimalist profile
│   │   └── poweruser.go            # Power User profile
│   └── tui/
│       ├── model.go                # Bubbletea state model
│       ├── update.go               # Message handling
│       ├── view.go                 # Rendering
│       ├── styles.go               # Lipgloss styles
│       └── screens/
│           ├── welcome.go
│           ├── detection.go
│           ├── profile.go
│           ├── components.go
│           ├── theme.go
│           ├── font.go
│           ├── colors.go
│           ├── preview.go
│           ├── installing.go
│           ├── complete.go
│           └── backup.go
├── pkg/
│   └── utils/                      # Shared utilities
├── e2e/
│   ├── Dockerfile.*                # Per-OS test containers
│   └── e2e_test.sh
├── scripts/
│   └── install.sh                  # curl-able installer script
├── themes/
│   └── *.omp.json                  # Bundled Oh My Posh themes
├── fonts/
│   └── manifest.json               # Font metadata and download URLs
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── .goreleaser.yaml
```

---

### 13.3 Component Interface

Every terminal component MUST implement a common interface. Methods return `ErrNotSupported` for capabilities the component doesn't have on a specific platform. The installer handles this gracefully — it skips unsupported steps and shows the user what WAS configured vs what COULDN'T be.

```go
type Component interface {
    // Identity
    Name() string
    Description() string
    Category() ComponentCategory  // Theme, Font, Tool, Enhancement

    // Detection
    Detect() (*DetectionResult, error)     // Is it installed? What version?
    IsInstalled() bool                      // Quick check
    
    // Installation
    Install(ctx context.Context) error     // Download and install
    Download(ctx context.Context) error    // Download only (for cache)
    
    // Configuration
    Configure(shell Shell) error           // Configure for specific shell
    GetConfigFiles() []string              // Files this component modifies
    
    // Preview support
    SupportsPreview() bool                 // Can this component be previewed?
    GeneratePreviewConfig() (string, error) // Generate temp config for preview
    
    // Validation
    Verify() error                         // Post-install health check
    
    // Metadata
    Dependencies() []string                // Other components required
    Platforms() []Platform                 // Supported platforms
}
```

This interface is the **extension point** for community contributions. Adding support for a new terminal tool means implementing this interface — nothing else changes.

### 13.4 Shell Interface

Each shell implements a common interface for configuration:

```go
type Shell interface {
    // Identity
    Name() string                          // zsh, bash, fish, pwsh
    Version() string                       // Shell version
    
    // Detection
    Detect() (*ShellInfo, error)           // Is this shell installed?
    IsDefault() bool                       // Is this the user's default shell?
    
    // Configuration
    GetRCFile() string                     // ~/.zshrc, ~/.bashrc, etc.
    GetConfigDir() string                  // ~/.config/fish, etc.
    
    // Modification
    AddToRC(content string) error          // Append to RC file
    RemoveFromRC(marker string) error      // Remove marked section
    BackupRC() (string, error)             // Backup RC file
    RestoreRC(backupPath string) error     // Restore from backup
    
    // Tool integration
    InitTool(tool string) string           // Get initialization command
    AddAlias(alias, command string) error  // Add shell alias
    
    // Preview
    SpawnSubshell(env map[string]string) (*Subshell, error)
}
```

### 13.5 Profile System

```go
type Profile struct {
    ID          string
    Name        string
    Description string
    Components  []ComponentID           // Components to install
    Theme       ThemeConfig             // Default theme
    Font        FontConfig              // Default font
    Colors      ColorConfig             // Color scheme and mode
    Tools       []ToolID                // Which tools to enable
    Aliases     map[string]string       // Shell aliases
}
```

**Predefined profiles:**

| Profile | What's Included | Description |
|---------|----------------|-------------|
| `developer` | Oh My Posh + Nerd Fonts + all tools + plugins | The complete experience for developers |
| `minimalist` | Oh My Posh + essential fonts only | Clean and simple, minimal overhead |
| `power-user` | Everything + custom aliases + advanced configs | Maximum productivity setup |
| `corporate` | Oh My Posh + basic colors + no external binaries | Security-focused, minimal dependencies |
| `student` | Developer profile + inline help + explanations | Learning-friendly with documentation |
| `custom` | User-defined | Full control over every aspect |

---


## 14. Distribution & Installation

### 14.1 Install Methods

| Method | Command | Priority | Notes |
|--------|---------|----------|-------|
| curl (recommended) | `curl -sL get.savanhi.shell/shell \| sh` | P0 | Universal, works everywhere with curl |
| Homebrew | `brew install savanhi/tap/savanhi-shell` | P0 | macOS/Linux users who prefer Homebrew |
| Go install | `go install github.com/savanhi/shell/cmd/savanhi@latest` | P1 | Go developers, latest features |
| Direct binary | Download from GitHub Releases | P1 | Air-gapped systems, specific versions |
| winget (Windows) | `winget install Savanhi.Shell` | P2 | Windows native package manager |
| Scoop (Windows) | `scoop install savanhi-shell` | P2 | Windows power users |

### 14.2 Cross-Compilation Targets

| Target | GOOS/GOARCH | Priority | Notes |
|--------|-------------|----------|-------|
| macOS Apple Silicon | darwin/arm64 | P0 | M1/M2/M3 Macs |
| macOS Intel | darwin/amd64 | P0 | Intel Macs |
| Linux x86_64 | linux/amd64 | P0 | Primary Linux target |
| Linux ARM64 | linux/arm64 | P1 | Raspberry Pi 4, cloud ARM instances |
| Linux ARM | linux/arm | P1 | Raspberry Pi 3, older ARM devices |
| Windows x86_64 | windows/amd64 | P2 | Native Windows support |
| Windows ARM64 | windows/arm64 | P2 | Windows on ARM devices |
| Android ARM64 | android/arm64 | P2 | Termux on Android |

### 14.3 Release Automation

**Build Pipeline:**

1. **GoReleaser** for cross-compilation and release packaging
   - Generates binaries for all targets
   - Creates tar.gz/zip archives
   - Generates checksums (SHA256)
   - Creates Homebrew formula

2. **GitHub Actions** for CI/CD
   - Run tests on PR
   - Build on tag push
   - Create GitHub Release
   - Upload artifacts

3. **Homebrew Tap** auto-update on new release
   - Repository: `savanhi/homebrew-tap`
   - Formula: `savanhi-shell.rb`
   - Auto-generated by GoReleaser

4. **Checksum Verification** in curl installer script
   - Downloads binary + checksums.txt
   - Verifies SHA256 before execution
   - Fails gracefully on mismatch

**Release Checklist:**

- [ ] Update version in `main.go`
- [ ] Update CHANGELOG.md
- [ ] Tag release: `git tag -a v0.1.0 -m "Release v0.1.0"`
- [ ] Push tag: `git push origin v0.1.0`
- [ ] GitHub Actions builds and releases
- [ ] Verify Homebrew tap updated
- [ ] Test curl installer on clean VM
- [ ] Announce release

### 14.4 Installer Script Details

The curl installer (`scripts/install.sh`) performs the following:

1. **Detect OS and Architecture**
   ```bash
   OS=$(uname -s | tr '[:upper:]' '[:lower:]')
   ARCH=$(uname -m)
   case $ARCH in
       x86_64) ARCH="amd64" ;;
       arm64|aarch64) ARCH="arm64" ;;
       armv7l) ARCH="arm" ;;
   esac
   ```

2. **Determine Download URL**
   ```
   https://github.com/savanhi/shell/releases/download/v${VERSION}/savanhi_${VERSION}_${OS}_${ARCH}.tar.gz
   ```

3. **Download and Verify**
   - Download binary archive
   - Download checksums.txt
   - Verify SHA256 checksum
   - Extract binary

4. **Install**
   - Check for existing installation
   - Copy binary to `~/.local/bin/` or `/usr/local/bin/`
   - Add to PATH if needed
   - Verify installation: `savanhi --version`

5. **Launch**
   - Optionally launch TUI immediately: `savanhi`

### 14.5 Versioning Strategy

**Semantic Versioning:** `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes (new config format, incompatible changes)
- **MINOR**: New features, new components, new profiles (backwards compatible)
- **PATCH**: Bug fixes, security updates, component version bumps

**Pre-releases:**
- `v0.1.0-alpha.1` - Early testing
- `v0.1.0-beta.1` - Feature complete, testing
- `v0.1.0-rc.1` - Release candidate

**LTS Releases:**
- Even minor versions (v1.0.x, v1.2.x) receive critical patches
- Odd minor versions (v1.1.x, v1.3.x) are latest features

### 14.6 Update Mechanism

**Automatic Update Checks:**
- Savanhi checks for updates on startup (can be disabled)
- Shows notification if new version available
- Prompts user to update

**Update Command:**
```bash
savanhi update
```

**Update Flow:**
1. Check current version against GitHub releases
2. Download latest installer script
3. Run installer with `--upgrade` flag
4. Preserve existing configs
5. Update complete

**Requirements:**

- R-DIST-01: The curl installer MUST work on all supported platforms without prerequisites (except curl)
- R-DIST-02: All binaries MUST be signed and checksums verified
- R-DIST-03: The installer MUST detect existing installations and offer upgrade paths
- R-DIST-04: Homebrew formula MUST be automatically updated on release
- R-DIST-05: The installer MUST support offline/air-gapped installation via direct binary download
- R-DIST-06: Version information MUST be embedded in the binary and accessible via `savanhi --version`
- R-DIST-07: Update checks MUST be optional and respect user privacy (no tracking)
- R-DIST-08: Pre-release versions MUST be clearly marked and opt-in only

---


## 15. Update & Maintenance

### 15.1 Self-Update

Savanhi Shell supports updating itself and its components to ensure users always have the latest features and bug fixes.

**Update Commands:**

```bash
# Check for updates
savanhi update check

# Update Savanhi Shell itself
savanhi update self

# Update all installed components to latest versions
savanhi update --components

# Update specific component
savanhi update --component oh-my-posh
savanhi update --component fonts
savanhi update --component zoxide

# Update everything (Savanhi + all components)
savanhi update --all
```

**Update Flow:**

1. **Check for Updates**
   - Query GitHub API for latest release
   - Compare with installed version
   - List available updates

2. **Download Update**
   - Download new binary for the appropriate platform
   - Verify checksum
   - Backup current binary

3. **Install Update**
   - Replace binary atomically
   - Verify new version works
   - Clean up backup

**Requirements:**

- R-UPDATE-01: The installer MUST support `savanhi update` to check for and install newer versions of itself
- R-UPDATE-02: The installer MUST support `savanhi update --components` to update all installed terminal components (oh-my-posh, fonts, zoxide, fzf, etc.)
- R-UPDATE-03: The installer MUST support updating individual components via `savanhi update --component <name>`
- R-UPDATE-04: The installer SHOULD check for updates on launch and notify the user (NOT auto-update without consent)
- R-UPDATE-05: The installer MUST preserve user configurations during self-updates
- R-UPDATE-06: The installer MUST support rollback to previous version if update fails

### 15.2 Configuration Sync & Portability

Savanhi Shell supports exporting and importing configurations for backup, team sharing, and multi-machine synchronization.

**Export Configuration:**

```bash
# Export current configuration
savanhi export --output my-terminal-config.json

# Export with profiles
savanhi export --output team-config.json --include-profiles

# Export minimal config (no history)
savanhi export --output minimal-config.json --no-history
```

**Import Configuration:**

```bash
# Import configuration
savanhi import --input my-terminal-config.json

# Import and apply immediately
savanhi import --input team-config.json --apply

# Preview before applying
savanhi import --input team-config.json --preview
```

**Configuration Bundle Structure:**

```json
{
  "savanhi_version": "0.1.0",
  "exported_at": "2026-03-19T15:30:00Z",
  "preferences": {
    // Full preferences.json content
  },
  "profiles": [
    // Custom profiles
  ],
  "components": {
    "oh-my-posh": "19.0.0",
    "zoxide": "0.9.0",
    "fzf": "0.44.0"
  },
  "theme_config": {
    // Theme-specific settings
  },
  "aliases": {
    "ls": "eza --icons",
    "cat": "bat"
  }
}
```

**Dotfiles Integration:**

Savanhi Shell is designed to work seamlessly with dotfiles repositories:

```bash
# Link Savanhi config to dotfiles repo
savanhi link --to ~/dotfiles/savanhi

# This creates symlinks:
# ~/.config/savanhi/ -> ~/dotfiles/savanhi/
```

**Requirements:**

- R-SYNC-01: The installer MUST support `savanhi export` to create a portable configuration bundle
- R-SYNC-02: The installer MUST support `savanhi import` to apply a configuration bundle
- R-SYNC-03: The installer SHOULD keep component versions synchronized with the versions specified in preferences
- R-SYNC-04: The installer MUST support importing configurations on a fresh machine and recreating the exact same setup
- R-SYNC-05: The installer SHOULD support dotfiles integration via symlinks
- R-SYNC-06: The installer MUST NOT export sensitive information (API keys, tokens, passwords) in configuration bundles
- R-SYNC-07: The installer SHOULD support team configuration profiles that can be shared across an organization
- R-SYNC-08: The installer MUST validate imported configurations and reject incompatible or corrupted bundles

### 15.3 Backup Management

Savanhi Shell maintains a history of configuration backups for easy rollback.

**Backup Commands:**

```bash
# List all backups
savanhi backup list

# Show backup details
savanhi backup show <backup-id>

# Restore from backup
savanhi backup restore <backup-id>

# Delete old backup
savanhi backup delete <backup-id>

# Export backup
savanhi backup export <backup-id> --output backup.tar.gz
```

**Backup Storage:**

```
~/.config/savanhi/backups/
├── 2026-03-19-143000-original/     # First-run backup
│   ├── .zshrc
│   ├── .bashrc
│   └── savanhi-backup.json
├── 2026-03-20-100500-before-theme-change/
│   ├── .zshrc
│   └── savanhi-backup.json
└── 2026-03-21-153000-before-font-update/
    ├── .zshrc
│   └── savanhi-backup.json
```

**Automatic Cleanup:**
- Keep last 10 backups by default
- Keep first-run backup forever
- Configurable retention policy

**Requirements:**

- R-BACKUP-01: The installer MUST create a backup before every configuration change
- R-BACKUP-02: The installer MUST support listing, viewing, and restoring from backups
- R-BACKUP-03: The installer MUST keep the original first-run backup permanently (unless explicitly deleted)
- R-BACKUP-04: The installer SHOULD automatically clean up old backups based on retention policy
- R-BACKUP-05: The installer MUST support exporting backups for external storage

---


## 16. Post-Install Experience

### 16.1 What the User Gets After Installation

When Savanhi Shell completes installation with the "Developer" profile:

**Oh My Posh:**
- `~/.config/oh-my-posh/config.json` — Selected theme (Agnoster, Paradox, etc.)
- Binary in `~/.local/bin/oh-my-posh` or via Homebrew
- Prompt displaying: git status, execution time, exit codes, directory, icons

**Nerd Fonts:**
- `~/.local/share/fonts/JetBrainsMonoNerdFont-Regular.ttf` (Linux)
- `~/Library/Fonts/JetBrainsMonoNerdFont-Regular.ttf` (macOS)
- Terminal font configured to use Nerd Font
- Icons and ligatures working in prompt and terminal

**Shell Configuration:**
- `~/.zshrc` (or `.bashrc`, `config.fish`) modified with:
  - Oh My Posh initialization
  - zoxide: `eval "$(zoxide init zsh)"`
  - fzf: Key bindings and completion
  - Aliases: `alias ls='eza --icons'`, `alias cat='bat'`
  - Custom PATH if needed

**Tools Installed:**
- `zoxide` — Smart directory jumping via `z` command
- `fzf` — Fuzzy finder with `Ctrl+R` for history, `Ctrl+T` for files
- `bat` — Syntax-highlighted `cat` replacement
- `eza` — Modern `ls` with icons and git integration
- `zsh-autosuggestions` — Command suggestions (zsh only)
- `zsh-syntax-highlighting` — Real-time syntax highlighting (zsh only)

**Terminal Emulator:**
- iTerm2: Profile "Savanhi" with Dracula colors, JetBrainsMono font
- Windows Terminal: Profile with color scheme and font configured
- Alacritty: `alacritty.yml` updated with colors and font

**Savanhi Configuration:**
- `~/.config/savanhi/original-backup.json` — Snapshot of pre-Savanhi configs
- `~/.config/savanhi/preferences.json` — Current Savanhi settings
- `~/.config/savanhi/history.json` — Log of all modifications

**Verification:**
- The installer runs health checks: oh-my-posh responds, fonts are detected, tools are in PATH
- Clear output: "Your terminal is ready. Restart your terminal to see the changes."

### 16.2 First-Run Experience

After restarting the terminal, the user sees:

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  ~/projects/my-awesome-project on  main via  v20.11.0        │
│  ❯ _                                                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**What changed:**
- Beautiful prompt with git branch, Node.js version, icons
- Directory jumping with `z` — type `z proj` to jump to `~/projects`
- Fuzzy history with `Ctrl+R` — type to search command history
- Syntax-highlighted files with `bat README.md`
- Icon-enhanced directories with `eza` or `ls`

### 16.3 Next Steps Guide

The completion screen MUST show:

**1. Restart Your Terminal**
```bash
# Close and reopen your terminal, or run:
source ~/.zshrc  # or ~/.bashrc, ~/.config/fish/config.fish
```

**2. Try It Out**
```bash
# Smart directory jumping
z Projects          # Jump to ~/Projects
z shell            # Jump to most frequented "shell" directory

# Fuzzy finder
Ctrl+R              # Search command history
Ctrl+T              # Fuzzy find files

# Better file viewing
cat README.md       # Now uses bat with syntax highlighting
ls                  # Now uses eza with icons

# Check what's installed
savanhi status      # Show current configuration
```

**3. Learn the Tools**
- `zoxide --help` — Learn about smart cd
- `fzf --help` — Master fuzzy finding
- `bat --help` — Better cat
- `eza --help` — Modern ls

**4. Customize Further**
```bash
savanhi             # Re-run TUI to change theme, font, or tools
savanhi revert      # Go back to original configuration
savanhi backup list # See all configuration backups
```

**5. Join the Community**
- GitHub: github.com/savanhi/shell
- Discord: discord.gg/savanhi
- Documentation: docs.savanhi.shell

### 16.4 Daily Usage Patterns

**Typical developer workflow with Savanhi:**

```bash
# Morning — open terminal, already configured
❯ z myproject        # Jump to project in 3 keystrokes

# Work — git operations with visual feedback
❯ git status         # Oh My Posh shows branch, changes in prompt
❯ git commit -m "..." # Exit code shown in prompt if fails

# Navigation — never type full paths
❯ z dow              # Jumps to ~/Downloads
❯ z co               # Jumps to ~/Code (most used)

# History — find that command from yesterday
❯ Ctrl+R → "docker"  # Fuzzy search, hit Enter

# Files — view with syntax highlighting
❯ bat src/main.go    # Syntax highlighted, git changes marked

# Directories — see icons and git status
❯ eza --icons --git  # Beautiful directory listing
```

### 16.5 Troubleshooting Quick Reference

If something doesn't work:

```bash
# Check if component is installed
which oh-my-posh
which zoxide

# Check Savanhi configuration
savanhi status      # Show current config
savanhi doctor      # Diagnose issues

# Revert if needed
savanhi revert      # Go back to original
savanhi backup restore <id>  # Restore specific backup

# Get help
savanhi --help
savanhi docs        # Open documentation
```

**Requirements:**

- R-POST-01: The installer MUST show a clear summary of what was installed and configured
- R-POST-02: The completion screen MUST include exact commands to try first
- R-POST-03: The installer MUST remind users to restart their terminal
- R-POST-04: The installer SHOULD verify installation health and report any issues
- R-POST-05: The installer MUST provide clear documentation links for learning the tools
- R-POST-06: The installer SHOULD suggest next steps based on the selected profile
- R-POST-07: The completion screen MUST show how to revert changes if needed

---


## 17. Non-Functional Requirements

### 17.1 Performance

Savanhi Shell is designed to be fast and lightweight, ensuring a smooth user experience across all supported platforms.

- **R-PERF-01**: Full installation (all tools + fonts + configuration) MUST complete in under 3 minutes on a standard broadband connection (25 Mbps)
- **R-PERF-02**: The Savanhi binary MUST be under 20MB (single static binary)
- **R-PERF-03**: The TUI MUST render at 60fps minimum for smooth animations and interactions
- **R-PERF-04**: System detection MUST complete in under 2 seconds
- **R-PERF-05**: Live preview generation MUST complete in under 1 second after selection changes
- **R-PERF-06**: Component downloads SHOULD use concurrent connections where possible
- **R-PERF-07**: The installer SHOULD cache downloaded components to avoid re-downloading on subsequent runs

### 17.2 Security

Security is paramount. Savanhi Shell handles system configurations and must do so safely.

- **R-SEC-01**: The installer MUST NOT request, store, or transmit any credentials, API keys, or tokens
- **R-SEC-02**: The curl installer script MUST be verifiable via SHA256 checksum before execution
- **R-SEC-03**: All binary downloads MUST use HTTPS with certificate validation
- **R-SEC-04**: The installer MUST NOT require root/sudo except for specific system operations (e.g., installing system-wide fonts), and MUST explain WHY when it does
- **R-SEC-05**: The installer MUST validate all downloaded binaries using SHA256 checksums from official sources
- **R-SEC-06**: The installer MUST NOT modify system directories outside of user space without explicit user consent
- **R-SEC-07**: Configuration files MUST be created with appropriate permissions (user-readable only where sensitive)
- **R-SEC-08**: The installer MUST scan for and warn about potential security issues (e.g., world-writable directories in PATH)

### 17.3 Reliability

Savanhi Shell must be reliable and handle failures gracefully.

- **R-REL-01**: Every installation step MUST be idempotent (safe to re-run multiple times)
- **R-REL-02**: If a step fails, the installer MUST continue with remaining steps and report all failures at the end
- **R-REL-03**: The installer MUST support `savanhi repair` to re-run failed steps or fix broken configurations
- **R-REL-04**: The backup system MUST create timestamped snapshots before ANY configuration modification
- **R-REL-05**: The installer MUST verify disk space availability before starting installation
- **R-REL-06**: Network operations MUST have timeouts and retry logic with exponential backoff
- **R-REL-07**: Partial installations MUST be resumable (don't re-download already installed components)
- **R-REL-08**: The installer MUST handle signals gracefully (SIGINT, SIGTERM) and clean up temporary files

### 17.4 Extensibility

Savanhi Shell is designed to grow. New components, shells, and terminal emulators can be added easily.

- **R-EXT-01**: Adding a new terminal component (e.g., a new tool like `ripgrep`) MUST only require implementing the Component interface
- **R-EXT-02**: Adding support for a new shell (e.g., `nu` shell) MUST only require implementing the Shell interface
- **R-EXT-03**: Adding support for a new terminal emulator MUST only require implementing the Terminal interface
- **R-EXT-04**: Profiles MUST be declarative (JSON data, not code) to allow user-created profiles
- **R-EXT-05**: Themes MUST be swappable without code changes (Oh My Posh theme files)
- **R-EXT-06**: Color schemes MUST be defined in standard formats (JSON, YAML) for easy addition
- **R-EXT-07**: The TUI MUST dynamically adapt to available components without code changes

### 17.5 Accessibility

Savanhi Shell should be usable by everyone, regardless of their setup or abilities.

- **R-ACC-01**: The TUI MUST work in terminals with 80x24 minimum dimensions
- **R-ACC-02**: The TUI MUST support both mouse and keyboard navigation
- **R-ACC-03**: All interactive elements MUST be accessible via keyboard (Tab, Enter, Arrow keys, Esc)
- **R-ACC-04**: The non-interactive mode MUST provide equivalent functionality for screen readers and CI environments
- **R-ACC-05**: Color schemes MUST have sufficient contrast ratios (WCAG 2.1 AA minimum)
- **R-ACC-06**: The TUI MUST support high-contrast mode for visually impaired users
- **R-ACC-07**: All text in the TUI SHOULD respect terminal font size settings
- **R-ACC-08**: The installer MUST work in restricted environments (e.g., Docker containers, CI systems)

### 17.6 Compatibility

Savanhi Shell must work across a wide range of systems and configurations.

- **R-COMP-01**: The binary MUST run on all supported platforms without recompilation (static linking)
- **R-COMP-02**: The installer MUST detect and respect existing user configurations
- **R-COMP-03**: The installer MUST work alongside existing dotfiles managers (chezmoi, stow, yadm)
- **R-COMP-04**: The installer MUST handle Unicode and emoji correctly across all supported terminals
- **R-COMP-05**: The installer MUST detect and warn about incompatible terminal emulators or shell versions
- **R-COMP-06**: Configuration files MUST be forward and backward compatible where possible

### 17.7 Maintainability

The codebase must be maintainable for long-term development.

- **R-MAINT-01**: All code MUST have unit tests with >80% coverage
- **R-MAINT-02**: All public functions MUST have documentation comments
- **R-MAINT-03**: The codebase MUST follow Go best practices and pass `go vet` and `golint`
- **R-MAINT-04**: Dependencies MUST be minimized; prefer standard library where possible
- **R-MAINT-05**: The project MUST have comprehensive documentation for contributors
- **R-MAINT-06**: Error messages MUST be clear, actionable, and include context

---


## 18. Relationship to Other Tools

### 18.1 Compatibility Matrix

Savanhi Shell is designed to work alongside existing terminal configuration tools and dotfiles managers.

| Tool | Relationship | Compatibility Notes |
|------|--------------|---------------------|
| **Oh My Posh** | Core dependency | Savanhi configures and manages Oh My Posh themes |
| **Starship** | Alternative | Savanhi can use Starship instead of Oh My Posh if preferred |
| **Powerlevel10k** | Alternative theme | Supported as Oh My Posh alternative |
| **Nerd Fonts** | Core dependency | Savanhi downloads and installs selected Nerd Fonts |
| **Homebrew** | Preferred package manager | Used on macOS and Linux when available |
| **zoxide** | Optional enhancement | Installed and configured if selected |
| **fzf** | Optional enhancement | Installed and configured if selected |
| **bat** | Optional enhancement | Installed and aliased to `cat` if selected |
| **eza/lsd** | Optional enhancement | Installed and aliased to `ls` if selected |

### 18.2 Dotfiles Managers

Savanhi Shell is designed to coexist with popular dotfiles management tools.

| Dotfiles Manager | Compatibility | Notes |
|-----------------|---------------|-------|
| **chezmoi** | Full support | Savanhi respects chezmoi-managed files; use `savanhi link` to integrate |
| **GNU Stow** | Full support | Works with stow-symlinked configurations |
| **yadm** | Full support | Detects yadm-managed files and preserves them |
| **bare git repo** | Full support | Standard approach; Savanhi adds to existing setup |
| **Ansible/Puppet** | Compatible | Savanhi can be run after configuration management tools |

**Integration Strategy:**

```bash
# If using chezmoi:
chezmoi init --apply username/dotfiles  # Apply your dotfiles first
savanhi install --profile developer     # Then enhance with Savanhi

# Savanhi will:
# 1. Detect chezmoi-managed files
# 2. Ask whether to integrate with chezmoi or manage separately
# 3. Create backups before any modifications
# 4. Document changes for you to commit to your dotfiles repo
```

### 18.3 Terminal Emulators

Savanhi Shell configures terminal emulators when possible, but respects user preferences.

| Terminal | Configuration Level | Notes |
|----------|-------------------|-------|
| **iTerm2** | Full | Can set font, colors, profile automatically |
| **Windows Terminal** | Full | Can set font, color scheme, profile |
| **Alacritty** | Full | Direct config file modification |
| **Kitty** | Full | Direct config file modification |
| **Terminal.app** | Partial | Limited configurability; manual instructions provided |
| **GNOME Terminal** | Partial | Profile-based; manual steps may be needed |
| **Konsole** | Partial | Profile-based; manual steps may be needed |
| **Custom/Exotic** | Detection only | Shows manual configuration instructions |

**Requirements:**

- R-TERM-01: Savanhi Shell MUST detect existing terminal configurations and offer to preserve them
- R-TERM-02: Savanhi Shell MUST work independently — no other tool is a prerequisite
- R-TERM-03: Savanhi Shell SHOULD detect popular dotfiles managers and offer integration options
- R-TERM-04: Savanhi Shell MUST NOT overwrite dotfiles-managed files without explicit user consent
- R-TERM-05: Savanhi Shell SHOULD provide documentation on how to integrate its configs into dotfiles repos
- R-TERM-06: Terminal emulator configurations MUST be optional (user can skip if they prefer manual setup)

### 18.4 Oh My Posh Ecosystem

Savanhi Shell is built on top of the Oh My Posh ecosystem and enhances it.

**What Savanhi adds to Oh My Posh:**

| Feature | Oh My Posh Alone | Oh My Posh + Savanhi |
|---------|-----------------|---------------------|
| Installation | Manual download and setup | One command, automatic |
| Theme selection | Edit JSON manually | Interactive TUI with preview |
| Font installation | Manual download | Automatic with font preview |
| Color schemes | Manual configuration | Pre-configured palettes |
| Shell integration | Manual RC editing | Automatic, safe edits |
| Revert changes | Manual restoration | One-command revert |
| Cross-shell sync | Manual config per shell | Automatic synchronization |

**Requirements:**

- R-OMP-01: Savanhi MUST support all official Oh My Posh themes
- R-OMP-02: Savanhi MUST support custom user themes (from file or URL)
- R-OMP-03: Savanhi MUST keep Oh My Posh updated to latest stable version
- R-OMP-04: Savanhi SHOULD support theme hot-reloading for live preview

### 18.5 Version Managers

Savanhi Shell detects and respects version managers.

| Version Manager | Detection | Integration |
|----------------|-----------|-------------|
| **fnm** (Node.js) | Detected | No conflict; separate concern |
| **nvm** (Node.js) | Detected | No conflict; separate concern |
| **pyenv** (Python) | Detected | No conflict; separate concern |
| **rbenv** (Ruby) | Detected | No conflict; separate concern |
| **goenv** (Go) | Detected | No conflict; separate concern |
| **rustup** (Rust) | Detected | No conflict; separate concern |

**Note:** Savanhi Shell doesn't manage language versions — it focuses on terminal enhancement. These tools can coexist without conflict.

---


## 19. Future Considerations (Out of Scope for v1)

These are NOT requirements for v1 but should inform architectural decisions and guide future development.

### 19.1 Team & Enterprise Features

1. **Team Profiles** — Shareable configuration profiles for standardizing terminal setup across development teams
   - Centralized profile repository
   - Organization-wide defaults
   - Onboarding automation for new team members

2. **Enterprise Integration** — Support for managed environments
   - Active Directory / LDAP integration
   - Group Policy support on Windows
   - Corporate proxy configuration
   - Custom CA certificate handling

### 19.2 Community & Marketplace

3. **Theme Marketplace** — Browse and install community-created themes
   - Curated theme gallery in TUI
   - User ratings and previews
   - One-click installation
   - Theme sharing via GitHub Gist

4. **Plugin Ecosystem** — Extensible plugin system for custom components
   - Third-party component support
   - Plugin API and documentation
   - Community plugin registry
   - Safe plugin sandboxing

### 19.3 Enhanced User Experience

5. **Terminal Health Dashboard** — TUI screen showing real-time status
   - Component health indicators
   - Font rendering test
   - Color support verification
   - Performance metrics
   - Update availability

6. **Project-Aware Configuration** — Auto-detection of project type
   - Detect Node.js, Python, Go, Rust projects
   - Suggest relevant tools and aliases
   - Project-specific profiles
   - Per-directory environment variables

7. **Wizard for First-Time Users** — Interactive tutorial mode
   - Step-by-step feature introduction
   - Interactive tutorials for each tool
   - Tip of the day system
   - Video integration for complex features

### 19.4 Migration & Portability

8. **Cross-Platform Migration** — Transfer configuration between OS
   - Export from macOS, import on Linux
   - Windows ↔ Unix configuration translation
   - Path and environment variable mapping

9. **Shell Migration Tool** — Switch between shells seamlessly
   - Migrate zsh config to fish
   - Convert bash aliases to PowerShell
   - Preserve functionality across shells

10. **Prompt Engine Migration** — Switch between Oh My Posh and alternatives
    - Convert Oh My Posh themes to Starship
    - Migrate custom segments
    - Feature parity analysis

### 19.5 Advanced Installation Methods

11. **Remote Provisioning** — Configure remote servers via SSH
    - `savanhi remote user@server` command
    - Batch configuration of multiple servers
    - Ansible/Chef/Puppet integration

12. **Nix/Home Manager Support** — Declarative alternative
    - Nix flake for reproducible installation
    - Home Manager module
    - Purely functional configuration
    - Rollback support via Nix generations

13. **Container Integration** — DevContainer and Docker support
    - DevContainer feature for VS Code
    - Docker image with pre-configured terminal
    - docker-compose integration

### 19.6 AI & Smart Features

14. **Smart Recommendations** — AI-powered suggestions
    - Analyze command history for patterns
    - Suggest aliases based on frequent commands
    - Recommend tools based on workflow
    - Performance optimization tips

15. **Natural Language Configuration** — Configure with plain English
    - "Make my terminal look like a hacker movie"
    - "Optimize for Python development"
    - "I want a minimal setup"

16. **Configuration Analytics** — Insights into terminal usage
    - Most used commands
    - Time saved with zoxide
    - Tool usage statistics
    - Export usage reports

### 19.7 Accessibility & Inclusion

17. **Screen Reader Optimization** — Enhanced accessibility
    - Full screen reader support
    - Audio cues for actions
    - Braille display compatibility
    - High-contrast themes

18. **Localization** — Multi-language support
    - Spanish, Portuguese, French, German, Japanese
    - RTL language support
    - Cultural adaptation of examples

### 19.8 Integration with Development Ecosystem

19. **IDE Integration** — Seamless IDE experience
    - VS Code extension
    - JetBrains plugin
    - Cursor integration
    - Neovim plugin

20. **Cloud Shell Support** — Native cloud environment support
    - GitHub Codespaces integration
    - Gitpod configuration
    - AWS CloudShell
    - Google Cloud Shell

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| **Oh My Posh** | Cross-shell prompt customization tool |
| **Nerd Font** | Patched fonts with thousands of extra icons and glyphs |
| **TUI** | Terminal User Interface — interactive terminal application |
| **Bubbletea** | Go framework for building TUIs using The Elm Architecture |
| **Lipgloss** | Go styling library for terminal applications |
| **zoxide** | Smart directory jumper — remembers frequently used directories |
| **fzf** | General-purpose command-line fuzzy finder |
| **bat** | Syntax-highlighting cat clone with Git integration |
| **eza** | Modern replacement for ls with icons and git support |
| **RC file** | Run Commands file — shell startup script (.zshrc, .bashrc) |
| **Subshell** | Child shell process spawned from parent shell |
| **Prompt** | Text displayed by shell indicating ready for input |
| **Theme** | Visual configuration for terminal prompt and colors |
| **Profile** | Predefined set of configurations for specific use cases |
| **Backup** | Snapshot of original configuration before modification |
| **Dotfiles** | Configuration files for Unix-like systems (start with .) |
| **Nerd Font Icons** | Special glyphs for files, folders, git status, etc. |
| **Ligatures** | Font feature combining character sequences into single glyphs |
| **Idempotent** | Operation that produces the same result if executed multiple times |

---

## Appendix B: Changelog

### v0.1.0-draft (2026-03-19)
- Initial PRD draft
- Core concept and architecture defined
- All 19 sections completed

---


## 20. Success Metrics

These metrics define what success looks like for Savanhi Shell v1.0 and beyond.

### 20.1 Installation Performance

| Metric | Target | Measurement |
|--------|--------|-------------|
| Time from curl to configured terminal | < 3 minutes | From running install command to "Done!" message |
| First-run detection speed | < 2 seconds | Time to detect OS, shell, terminal, and current config |
| Live preview generation | < 1 second | Time from selection to rendered preview |
| Binary size | < 20 MB | Single static binary |
| Installation success rate | > 95% | Percentage of successful installations on first attempt |

### 20.2 Platform & Compatibility Coverage

| Metric | Target | Measurement |
|--------|--------|-------------|
| Supported OS at launch | macOS + 3 Linux distros + WSL | macOS, Ubuntu, Arch, Fedora, WSL2 |
| Shell coverage | 4 shells at launch | zsh, bash, fish, PowerShell |
| Terminal emulator detection | 6+ emulators | iTerm2, Terminal.app, Windows Terminal, Alacritty, Kitty, WezTerm |
| Component coverage | 8+ components | Oh My Posh, Nerd Fonts, zoxide, fzf, bat, eza, zsh-autosuggestions, zsh-syntax-highlighting |
| Idempotency | 100% | Re-running produces same result, no duplicate entries |

### 20.3 User Experience Quality

| Metric | Target | Measurement |
|--------|--------|-------------|
| User needs to manually edit configs | 0 files | Everything configured automatically (except optional tweaks) |
| Preview accuracy | 100% | Preview matches actual post-install appearance |
| Revert success rate | > 99% | Successfully restore original configuration |
| TUI responsiveness | 60 FPS | Smooth animations, no lag |
| Non-interactive mode parity | 100% | CLI provides same functionality as TUI |

### 20.4 Reliability & Robustness

| Metric | Target | Measurement |
|--------|--------|-------------|
| Backup creation reliability | 100% | Every modification backed up before change |
| Error recovery rate | > 90% | Successful recovery from partial failures |
| Network timeout handling | Graceful | Exponential backoff, clear error messages |
| Disk space check | Always | Verify space before downloading |
| Signal handling | Graceful | Clean shutdown on SIGINT/SIGTERM |

### 20.5 Community & Adoption

| Metric | Target (6 months) | Measurement |
|--------|-------------------|-------------|
| GitHub stars | 1,000+ | Community interest |
| Active installations | 5,000+ | Based on update checks (opt-in) |
| Community themes | 10+ | User-contributed themes |
| Bug reports resolved | > 90% | Issue resolution rate |
| Documentation completeness | 100% | All features documented |

### 20.6 Developer Experience

| Metric | Target | Measurement |
|--------|--------|-------------|
| Code coverage | > 80% | Unit test coverage |
| Build time | < 30 seconds | From clean to binary |
| Cross-compilation | All targets | Single command builds all platforms |
| Release automation | Fully automated | Tag → Release in < 10 minutes |
| Contributor onboarding | < 1 hour | Time for new contributor to first PR |

### 20.7 Long-Term Vision Metrics

| Metric | 1 Year Target | 3 Year Target |
|--------|---------------|---------------|
| Active users | 10,000+ | 100,000+ |
| Platform support | 8+ OS/platforms | All major platforms |
| Component ecosystem | 15+ tools | 30+ tools |
| Community plugins | 5+ | 50+ |
| Enterprise adoption | 10+ companies | 100+ companies |
| Nix/Home Manager support | Available | Mature |

---


## 21. Open Questions

These questions need resolution before or during v1 development:

### 21.1 Naming & Branding

1. **Binary name**: `savanhi`, `savanhi-shell`, `sav`, or `sh`? Should it be short and memorable or descriptive?
2. **Domain**: `get.savanhi.shell`, `savanhi.dev`, `savanhi.sh`, or something else?
3. **Organization**: Should this be under a GitHub organization (`savanhi/shell`) or personal account initially?
4. **Logo & branding**: Do we need a logo? Should it relate to shells, terminals, or the name Savanhi?

### 21.2 Technical Architecture

5. **Theme distribution**: Should themes be bundled in the binary, fetched from GitHub at runtime, or downloaded on-demand?
6. **Font distribution**: Should we bundle font metadata or fetch from Nerd Fonts API? How to handle font versioning?
7. **Configuration format**: Is JSON the right choice for preferences, or should we use YAML/TOML for better human readability?
8. **State storage**: Should we use SQLite instead of JSON files for better querying and performance?

### 21.3 Platform Support

9. **Windows native**: How much effort to invest in native Windows (not WSL) support for v1? PowerShell is widely used but Oh My Posh works best on Unix-like systems.
10. **Termux/Android**: Is mobile terminal configuration a priority, or should we focus on desktop first?
11. **ChromeOS**: Should we support ChromeOS Linux (Crostini) explicitly?

### 21.4 User Experience

12. **Default profile**: Should "Developer" be the default, or should we ask users on first run?
13. **Auto-preview**: Should preview generate automatically on selection change, or require explicit action (performance vs convenience)?
14. **Update frequency**: Should we check for updates on every launch, weekly, or only on explicit request?
15. **Telemetry**: Should we collect anonymous usage data to improve the tool? How to make it opt-in and transparent?

### 21.5 Integration & Ecosystem

16. **Oh My Posh convergence**: Should we eventually merge more closely with Oh My Posh, or remain a separate configurator?
17. **Version pinning**: Should we pin specific versions of tools (reproducible) or always install latest (fresh)?
18. **Rollback granularity**: Should we support rolling back individual components, or only full configuration restore?
19. **Team/enterprise features**: Should v1 include team profiles, or is that a v2 feature?

### 21.6 Business & Sustainability

20. **Monetization**: Should this remain completely free, or offer paid features (team/enterprise)?
21. **Sponsorship**: Should we accept GitHub Sponsors or similar to fund development?
22. **Commercial support**: Should we offer commercial support for enterprise users?

---

## Appendix C: Competitive Landscape

| Tool | What it does | Savanhi Differentiation |
|------|-------------|------------------------|
| **Oh My Posh install scripts** | Installs Oh My Posh only | Savanhi does full ecosystem: fonts, colors, tools, shell integration |
| **Nerd Fonts install scripts** | Installs fonts only | Savanhi integrates fonts with terminal and shell automatically |
| **Homebrew** | Package manager | Savanhi is opinionated terminal configurator, not generic package manager |
| **Ansible/Dotfiles** | Configuration management | Savanhi is interactive with live preview, not declarative only |
| **Fig / Amazon Q** | IDE-style autocomplete | Savanhi configures underlying terminal, works everywhere |
| **Starship** | Cross-shell prompt | Savanhi can configure Starship OR Oh My Posh, plus full ecosystem |
| **Shell install scripts** | Install shell only | Savanhi enhances existing shells, doesn't replace them |
| **Terminal theme repos** | Manual theme installation | Savanhi automates with detection and preview |
| **Prez/Oh My Zsh** | Shell frameworks | Savanhi works alongside them, doesn't conflict |
| **chezmoi/stow/yadm** | Dotfiles managers | Savanhi integrates with them, can export to them |

**Key insight:** No existing tool provides the full Savanhi experience — interactive preview, comprehensive terminal configuration, and seamless integration across shells, terminals, and tools. Savanhi orchestrates ALL of them into a coherent, working terminal ecosystem.

---

## Appendix D: Example Commands Reference

### Installation Commands

```bash
# Quick install with curl (recommended)
curl -sL get.savanhi.shell/shell | sh

# Install via Homebrew
brew install savanhi/tap/savanhi-shell

# Install via Go
go install github.com/savanhi/shell/cmd/savanhi@latest
```

### Interactive Mode

```bash
# Launch TUI
savanhi

# Launch with specific profile
savanhi --profile developer

# First-time setup wizard
savanhi init
```

### Non-Interactive Mode

```bash
# Install with a preset
savanhi install --profile developer --non-interactive

# Custom installation
savanhi install \
  --shell zsh \
  --theme agnoster \
  --font "JetBrainsMono Nerd Font" \
  --colors dracula \
  --tools zoxide,fzf,bat,eza \
  --non-interactive

# Dry run (show what would be installed)
savanhi install --profile developer --dry-run
```

### Update Commands

```bash
# Check for updates
savanhi update check

# Update Savanhi itself
savanhi update self

# Update all components
savanhi update --components

# Update specific component
savanhi update --component oh-my-posh

# Update everything
savanhi update --all
```

### Configuration Management

```bash
# Export current configuration
savanhi export --output my-config.json

# Import configuration
savanhi import --input my-config.json

# Apply saved profile
savanhi apply --profile my-custom-profile

# Link to dotfiles repo
savanhi link --to ~/dotfiles/savanhi
```

### Backup & Restore

```bash
# List all backups
savanhi backup list

# Show backup details
savanhi backup show <backup-id>

# Restore from backup
savanhi backup restore <backup-id>

# Revert to original (first-run) configuration
savanhi revert

# Delete old backup
savanhi backup delete <backup-id>
```

### Information & Diagnostics

```bash
# Show current status
savanhi status

# Run diagnostics
savanhi doctor

# Show version
savanhi --version

# Show help
savanhi --help

# Open documentation
savanhi docs
```

### Profile Management

```bash
# List available profiles
savanhi profile list

# Create new profile
savanhi profile create --name my-profile --base developer

# Delete profile
savanhi profile delete my-profile

# Export profile
savanhi profile export my-profile --output my-profile.json
```

### Repair & Maintenance

```bash
# Repair broken installation
savanhi repair

# Verify installation
savanhi verify

# Clean cache
savanhi cache clean

# Uninstall Savanhi
savanhi uninstall
```

---

## Appendix E: File Locations Reference

### Savanhi Configuration

| File/Directory | Purpose |
|----------------|---------|
| `~/.config/savanhi/` | Main configuration directory |
| `~/.config/savanhi/preferences.json` | Current user preferences |
| `~/.config/savanhi/original-backup.json` | First-run system snapshot |
| `~/.config/savanhi/history.json` | Modification history |
| `~/.config/savanhi/backups/` | Configuration backups |
| `~/.config/savanhi/profiles/` | Custom user profiles |
| `~/.config/savanhi/cache/` | Downloaded components cache |
| `~/.config/savanhi/logs/` | Installation logs |

### Shell Configuration (Modified by Savanhi)

| Shell | RC File | Savanhi Section Marker |
|-------|---------|----------------------|
| zsh | `~/.zshrc` | `# Savanhi Shell Configuration` |
| bash | `~/.bashrc` | `# Savanhi Shell Configuration` |
| fish | `~/.config/fish/config.fish` | `# Savanhi Shell Configuration` |
| PowerShell | `~/.config/powershell/Microsoft.PowerShell_profile.ps1` | `# Savanhi Shell Configuration` |

### Component Configuration

| Component | Configuration Location |
|-----------|----------------------|
| Oh My Posh | `~/.config/oh-my-posh/config.json` |
| zoxide | Shell RC file (eval statement) |
| fzf | Shell RC file (key bindings) |
| bat | `~/.config/bat/config` |
| Nerd Fonts | System font directories |

### Terminal Emulator Configuration

| Terminal | Configuration Location |
|----------|----------------------|
| iTerm2 | `~/Library/Preferences/com.googlecode.iterm2.plist` |
| Windows Terminal | `%LOCALAPPDATA%\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\settings.json` |
| Alacritty | `~/.config/alacritty/alacritty.yml` or `alacritty.toml` |
| Kitty | `~/.config/kitty/kitty.conf` |

---

## Document Information

**Document**: Savanhi Shell Product Requirements Document  
**Version**: 0.1.0-draft  
**Author**: Steven Sanchez  
**Date**: 2026-03-19  
**Status**: Draft - Complete  

### Sections Summary

| Section | Title | Status |
|---------|-------|--------|
| 1 | Problem Statement | ✅ Complete |
| 2 | Vision | ✅ Complete |
| 3 | Target Users | ✅ Complete |
| 4 | Supported Platforms | ✅ Complete |
| 5 | Prerequisites and Dependency Management | ✅ Complete |
| 6 | Architecture Overview | ✅ Complete |
| 7 | Real-Time Preview System | ✅ Complete |
| 8 | JSON Persistence System | ✅ Complete |
| 9 | Lazy Installation Strategy | ✅ Complete |
| 10 | System Detection at Startup | ✅ Complete |
| 11 | Components to Install & Configure | ✅ Complete |
| 12 | User Experience | ✅ Complete |
| 13 | Technical Architecture | ✅ Complete |
| 14 | Distribution & Installation | ✅ Complete |
| 15 | Update & Maintenance | ✅ Complete |
| 16 | Post-Install Experience | ✅ Complete |
| 17 | Non-Functional Requirements | ✅ Complete |
| 18 | Relationship to Other Tools | ✅ Complete |
| 19 | Future Considerations | ✅ Complete |
| 20 | Success Metrics | ✅ Complete |
| 21 | Open Questions | ✅ Complete |
| A | Glossary | ✅ Complete |
| B | Changelog | ✅ Complete |
| C | Competitive Landscape | ✅ Complete |
| D | Example Commands Reference | ✅ Complete |
| E | File Locations Reference | ✅ Complete |

**Total Sections**: 21 main sections + 5 appendices  
**Total Pages**: ~40-50 pages (estimated)  
**Word Count**: ~15,000+ words  

---