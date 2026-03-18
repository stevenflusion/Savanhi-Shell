<div align="center">

# Savanhi Shell

**One command. Any theme. Any operating system.**

*The Oh My Posh ecosystem: better and easier than ever.*

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20WSL-lightgrey?style=for-the-badge)](https://github.com/stevenflusion/Savanhi-Shell)

[Features](#-features) • [Installation](#-installation) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Contributing](#-contributing)

</div>

---

<img src="docs/assets/demo.png" alt="Savanhi Shell Demo" width="100%">

> **Preview your terminal configuration before installing.** No more "try and cry" — see exactly how your terminal will look, then decide.

## 🚀 Features

### Core Features

| Feature | Description |
|---------|-------------|
| **🔍 Live Preview** | See your theme, font, and colors in real-time before committing |
| **📦 Smart Installation** | Automatic dependency resolution with oh-my-posh, Nerd Fonts, and tools |
| **🔄 Safe Rollback** | One command to restore your original configuration |
| **🖥️ Cross-Platform** | Works on macOS (Intel & Apple Silicon), Linux, and WSL |
| **🐚 Multi-Shell** | Supports both zsh and bash |

### What Gets Configured

| Component | What You Get |
|-----------|--------------|
| **oh-my-posh** | Beautiful prompt themes (Agnoster, Paradox, Powerlevel10k, etc.) |
| **Nerd Fonts** | Patched fonts with icons and symbols |
| **zoxide** | Smart `cd` command (like `z` but better) |
| **fzf** | Fuzzy finder for files, history, and more |
| **bat** | Syntax-highlighted `cat` replacement |
| **eza** | Modern `ls` with icons and git status |

## 📋 Requirements

| Requirement | Minimum Version |
|-------------|-----------------|
| **Go** | 1.21+ (for building from source) |
| **Shell** | zsh or bash |
| **Terminal** | Any modern terminal (iTerm2, Alacritty, Kitty, etc.) |

## 📦 Installation

### Quick Install (Recommended)

**macOS / Linux / WSL:**

```bash
curl -fsSL https://raw.githubusercontent.com/stevenflusion/Savanhi-Shell/main/scripts/install.sh | bash
```

This will:
1. Detect your OS and architecture
2. Download the appropriate binary
3. Verify the checksum
4. Install to `~/.local/bin` (or `/usr/local/bin` with sudo)

### Homebrew (macOS)

```bash
brew tap stevenflusion/tap
brew install savanhi-shell
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/stevenflusion/Savanhi-Shell.git
cd Savanhi-Shell

# Build
make build

# Install (optional)
sudo make install
```

### Manual Download

Download the latest binary from [Releases](https://github.com/stevenflusion/Savanhi-Shell/releases):

| Platform | Architecture | Download |
|----------|--------------|----------|
| macOS | Intel (amd64) | `savanhi-shell-darwin-amd64` |
| macOS | Apple Silicon (arm64) | `savanhi-shell-darwin-arm64` |
| Linux | amd64 | `savanhi-shell-linux-amd64` |
| Linux | arm64 | `savanhi-shell-linux-arm64` |

```bash
# Make executable
chmod +x savanhi-shell-*

# Move to PATH
mv savanhi-shell-* /usr/local/bin/savanhi-shell
```

## ⚡ Quick Start

### 1. Launch Interactive Mode

```bash
savanhi-shell
```

### 2. Follow the TUI Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  Savanhi Shell                                                  │
│  Your terminal, supercharged.                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  🔍 Detecting your system...                                    │
│                                                                 │
│  ✓ macOS 14.4 (Sonoma) - Apple Silicon                         │
│  ✓ zsh 5.9                                                      │
│  ✓ iTerm2 3.4.23                                               │
│  ✗ oh-my-posh (not installed)                                   │
│  ✗ Nerd Fonts (not detected)                                    │
│                                                                 │
│  [Press Enter to continue]                                      │
└─────────────────────────────────────────────────────────────────┘
```

### 3. Select Your Configuration

Choose your theme, font, and tools with live preview:

```
┌─────────────────────────────────────────────────────────────────┐
│  Select Theme                          [Preview Active]          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ★ Agnoster                                                   │
│  ○ Paradox                                                    │
│  ○ Powerlevel10k                                              │
│  ○ Pure                                                       │
│                                                                 │
│  ┌─ PREVIEW ─────────────────────────────────────────────────┐ │
│  │  ~/projects/my-app on main via  v20.11.0                  │ │
│  │  ❯ _                                                    │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                 │
│  [↑/↓] Navigate  [Enter] Select  [q] Quit                      │
└─────────────────────────────────────────────────────────────────┘
```

### 4. Install

Review your selections and install with one key:

```
┌─────────────────────────────────────────────────────────────────┐
│  Ready to Install                                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Theme:        Agnoster                                        │
│  Font:         JetBrainsMono Nerd Font                         │
│  Tools:        zoxide, fzf, bat, eza                           │
│                                                                 │
│  [A] Accept & Install    [C] Cancel    [R] View Rollback      │
└─────────────────────────────────────────────────────────────────┘
```

## 🎨 Command Line Options

```bash
savanhi-shell [options]

Options:
  --version              Show version information
  --help                 Show help message
  --config <FILE>        Path to configuration file (JSON)
  --non-interactive      Run without TUI (for scripting/CI)
  --dry-run              Preview changes without applying
  --detect               Only run system detection
  --verify               Verify existing installation
  --rollback             Rollback last installation
  --rollback-original    Restore to original state
  --verbose              Enable verbose output
  --timeout <DURATION>   Operation timeout (default: 10m)
```

### Non-Interactive Mode

For scripting or CI/CD pipelines:

```bash
# Create config file
cat > config.json << EOF
{
  "theme": "agnoster",
  "font": "JetBrainsMono Nerd Font",
  "tools": ["zoxide", "fzf", "bat", "eza"],
  "dry_run": false
}
EOF

# Run non-interactive
savanhi-shell --non-interactive --config config.json
```

## ⌨️ Key Bindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` | Move left / back |
| `l` / `→` | Move right / select |
| `Enter` | Select / confirm |
| `Esc` | Back / cancel |
| `q` | Quit |
| `?` | Help |
| `r` | Refresh |

## 🔄 Rollback

### Undo Last Installation

```bash
savanhi-shell --rollback
```

### Full Restore

Restore your shell to its original state (before Savanhi made any changes):

```bash
savanhi-shell --rollback-original
```

This removes:
- All Savanhi modifications from your RC files
- Installed components (oh-my-posh, fonts, tools)
- Configuration directory (`~/.config/savanhi/`)

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [Getting Started](docs/getting-started.md) | Quick start guide |
| [Configuration](docs/configuration.md) | Detailed configuration options |
| [Architecture](docs/architecture.md) | Technical architecture overview |
| [Troubleshooting](docs/troubleshooting.md) | Common issues and solutions |
| [Contributing](CONTRIBUTING.md) | Contribution guidelines |

## 🏗️ Project Structure

```
savanhi-shell/
├── cmd/savanhi-shell/          # Entry point + CLI
├── internal/
│   ├── cli/                    # Non-interactive CLI
│   ├── detector/               # OS/Shell/Terminal detection
│   ├── errors/                 # Error handling
│   ├── installer/              # Installation engine
│   ├── persistence/           # JSON backup & preferences
│   ├── preview/                # Live preview (subshell)
│   ├── staging/                # Change staging
│   └── tui/                    # Bubble Tea interface
├── pkg/shell/                  # Public shell interface
├── configs/bundled/            # Bundled themes
├── scripts/                    # Install scripts
├── tests/e2e/                  # End-to-end tests
└── docs/                       # Documentation
```

## 🛠️ Development

### Quick Start (Development)

```bash
# Clone and build
git clone https://github.com/stevenflusion/Savanhi-Shell.git
cd Savanhi-Shell
make build

# Run the TUI
./savanhi-shell

# Or just detect your system (no changes)
./savanhi-shell --detect

# Run tests
make test
```

### Prerequisites

- Go 1.21+
- Make (optional, for Makefile commands)

### Commands

```bash
make build          # Build for current platform
make build-all      # Build for all platforms (cross-compile)
make test           # Run tests
make coverage       # Run tests with coverage
make lint           # Run linter
make clean          # Clean build artifacts
make install        # Install to /usr/local/bin (requires sudo)
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run E2E tests
go test ./tests/e2e/... -v

# Run specific package
go test ./internal/detector/... -v
```

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'feat: add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [oh-my-posh](https://ohmyposh.dev/) - Prompt theme engine
- [Nerd Fonts](https://www.nerdfonts.com/) - Patched fonts
- [zoxide](https://github.com/ajeetdsouza/zoxide) - Smart cd
- [fzf](https://github.com/junegunn/fzf) - Command-line fuzzy finder
- [bat](https://github.com/sharkdp/bat) - A cat clone with syntax highlighting
- [eza](https://github.com/eza-community/eza) - A modern alternative to ls

---

<div align="center">

**Made with ❤️ by [Steven](https://github.com/stevenflusion)**

[Report Bug](https://github.com/stevenflusion/Savanhi-Shell/issues) · [Request Feature](https://github.com/stevenflusion/Savanhi-Shell/issues) · [Discussions](https://github.com/stevenflusion/Savanhi-Shell/discussions)

</div>