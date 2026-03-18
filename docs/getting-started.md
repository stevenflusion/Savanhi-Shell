# Getting Started with Savanhi Shell

This guide will help you get up and running with Savanhi Shell quickly.

## Prerequisites

- **Operating System**: macOS or Linux (Windows via WSL)
- **Shell**: zsh or bash
- **Terminal**: Any modern terminal emulator
- **Go**: 1.21+ (only needed for building from source)

## Installation

### Option 1: Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/savanhi/shell/main/scripts/install.sh | bash
```

### Option 2: Homebrew (macOS)

```bash
brew tap savanhi/tap
brew install savanhi-shell
```

### Option 3: Manual Download

1. Download the latest release for your platform from [GitHub Releases](https://github.com/savanhi/shell/releases)
2. Make it executable:
   ```bash
   chmod +x savanhi-shell-*
   ```
3. Move to your PATH:
   ```bash
   sudo mv savanhi-shell-* /usr/local/bin/savanhi-shell
   ```

### Option 4: Build from Source

```bash
git clone https://github.com/savanhi/shell.git
cd shell
make build
sudo make install
```

## First Run

Launch Savanhi Shell in interactive mode:

```bash
savanhi-shell
```

You'll see an interactive TUI that guides you through:

### 1. System Detection

Savanhi Shell automatically detects:
- Your operating system and version
- Your current shell (zsh/bash)
- Your terminal emulator
- Installed Nerd Fonts

### 2. Theme Selection

Browse available oh-my-posh themes:
- Use `j`/`k` or arrow keys to navigate
- Press `Enter` to select
- Press `p` to preview the theme

### 3. Font Selection

Choose a Nerd Font to install:
- Recommended: MesloLGS NF (works best with most themes)
- Use arrow keys to browse
- Press `Enter` to select

### 4. Tool Installation

Select which productivity tools to install:
- **zoxide**: Smart `cd` command
- **fzf**: Fuzzy finder
- **bat**: Better `cat` with syntax highlighting
- **eza**: Modern `ls` replacement

### 5. Preview

See your configuration before committing:
- Live preview in a subshell
- Verify everything looks correct

### 6. Install

Apply your changes:
- Automatic backup created
- Changes committed to your shell RC file
- Verification of successful installation

## What Gets Installed

### Shell RC Modifications

Savanhi Shell adds sections to your `.zshrc` or `.bashrc`:

```bash
# >>> savanhi-oh-my-posh >>>
# Savanhi-managed oh-my-posh configuration
eval "$(oh-my-posh init zsh --config ~/.config/savanhi/themes/powerlevel10k.omp.json)"
# <<< savanhi-oh-my-posh <<<

# >>> savanhi-tools >>>
# Savanhi-managed tool configurations
eval "$(zoxide init zsh)"
eval "$(fzf --zsh)"
alias cat='bat --paging=never'
alias ls='eza'
# <<< savanhi-tools <<<
```

### Files Created

- `~/.config/savanhi/` - Configuration directory
- `~/.config/savanhi/original-backup.json` - Your original configuration
- `~/.config/savanhi/preferences.json` - Your preferences
- `~/.config/savanhi/themes/` - Downloaded themes

## Next Steps

- Learn about [Configuration Options](configuration.md)
- See [Troubleshooting](troubleshooting.md) if you encounter issues
- Check out [Advanced Usage](advanced.md) for more features

## Quick Commands

```bash
# Show version
savanhi-shell --version

# Detect system only
savanhi-shell --detect

# Verify installation
savanhi-shell --verify

# Rollback last installation
savanhi-shell --rollback

# Full uninstall (restore to original)
savanhi-shell --rollback-original
```

## Key Bindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` | Back |
| `l` / `→` | Select/Forward |
| `Enter` | Confirm/Select |
| `Esc` | Cancel/Back |
| `q` | Quit |
| `?` | Help |
| `r` | Refresh |
| `p` | Preview |

## Need Help?

- [Troubleshooting Guide](troubleshooting.md)
- [GitHub Issues](https://github.com/savanhi/shell/issues)
- [Discussions](https://github.com/savanhi/shell/discussions)