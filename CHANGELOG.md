# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2026-03-18

### Added
- Fish shell support for theme previews and RC file manipulation
- `FishShell` implementation in `pkg/shell/fish.go` with proper `set -x VAR value` syntax
- Fish shell detection in preview engine validation
- Fish-specific environment variable escaping in preview sessions
- Dynamic RC file path messages in TUI install success screen
- Preview screen now shows selected theme, font, and system info
- Foot terminal detection for Wayland users
- Detection of current shell session (not just default shell) via $FISH_VERSION, $ZSH_VERSION, $BASH_VERSION

### Changed
- Preview engine now supports Fish shell with proper config.fish temp files
- Preview session RC generation includes Fish-specific system config sourcing
- Installer verification now correctly identifies Fish RC file paths (~/.config/fish/config.fish)
- TUI install screen displays shell-appropriate RC source commands
- Preview screen displays actual configuration instead of placeholder text
- Shell detection now detects current session shell instead of only default shell
- Terminal detection now recognizes Wayland terminals (foot, etc.)

### Fixed
- Preview validation now accepts Fish as a valid shell type
- Subshell spawning now uses correct Fish shell flags (-l for login, no -i)
- Theme selector now displays bundled themes (was empty before)
- Font selector now displays recommended Nerd Fonts

## [1.0.0] - 2024-01-XX

### Added
- Initial release of Savanhi Shell
- Interactive TUI for shell configuration
- System detection (OS, shell, terminal, fonts)
- oh-my-posh theme selection and installation
- Nerd Font installation with preview
- Productivity tool installation (zoxide, fzf, bat, eza)
- Live preview of configuration changes
- Automatic backup and rollback functionality
- Non-interactive mode for scripting and CI/CD
- Cross-platform support (macOS, Linux)
- zsh and bash shell support
- Multiple terminal emulator support
- CLI flags for various operation modes
- Configuration file support
- Comprehensive error handling
- Detailed logging

### Changed
- N/A (Initial release)

### Deprecated
- N/A (Initial release)

### Removed
- N/A (Initial release)

### Fixed
- N/A (Initial release)

### Security
- SHA256 checksum verification for downloads
- Safe file operations with atomic writes
- Permission checks before modifications

## [0.9.0-beta] - 2024-01-XX

### Added
- Beta release for testing
- Core TUI functionality
- Basic installation flow
- System detection module
- Persistence layer for backups

## [0.1.0-alpha] - 2024-01-XX

### Added
- Initial development release
- Project structure setup
- Basic tests

---

## Version History

| Version | Date | Description |
|---------|------|-------------|
| 1.0.1 | 2026-03-18 | Fish shell support |
| 1.0.0 | 2024-01-XX | Initial public release |
| 0.9.0-beta | 2024-01-XX | Beta testing release |
| 0.1.0-alpha | 2024-01-XX | Initial development release |

## Upgrade Guide

### From 0.9.0-beta to 1.0.0

1. Backup your configuration:
   ```bash
   cp ~/.config/savanhi/config.json ~/.config/savanhi/config.json.bak
   ```

2. Download the new version

3. Run verification:
   ```bash
   savanhi-shell --verify
   ```

4. Restore configuration if needed:
   ```bash
   mv ~/.config/savanhi/config.json.bak ~/.config/savanhi/config.json
   ```

### From 0.1.0-alpha to 0.9.0-beta

Complete reinstall recommended:
```bash
savanhi-shell --rollback-original
# Install new version
savanhi-shell
```

---

[1.0.1]: https://github.com/savanhi/shell/releases/tag/v1.0.1
[1.0.0]: https://github.com/savanhi/shell/releases/tag/v1.0.0
[0.9.0-beta]: https://github.com/savanhi/shell/releases/tag/v0.9.0-beta
[0.1.0-alpha]: https://github.com/savanhi/shell/releases/tag/v0.1.0-alpha