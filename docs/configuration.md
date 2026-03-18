# Configuration Reference

This document describes all configuration options for Savanhi Shell.

## Configuration File Location

Savanhi Shell looks for configuration in the following order:

1. `--config <FILE>` - Explicit command line flag
2. `~/.config/savanhi/config.json` - Default location
3. `$XDG_CONFIG_HOME/savanhi/config.json` - XDG standard

## Configuration File Format

Configuration files use JSON format:

```json
{
  "theme": "powerlevel10k",
  "font": "MesloLGS NF",
  "tools": ["zoxide", "fzf", "bat", "eza"],
  "install_oh_my_posh": true,
  "install_zoxide": true,
  "install_fzf": true,
  "install_bat": true,
  "install_eza": true,
  "skip_checksum": false,
  "skip_verification": false,
  "dry_run": false,
  "force": false,
  "timeout": "10m0s",
  "config_dir": "",
  "backup": true,
  "rollback": false,
  "rollback_to_original": false
}
```

## Configuration Options

### Theme Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `theme` | string | `"powerlevel10k"` | oh-my-posh theme to install |

Available themes:
- `powerlevel10k` - Feature-rich, highly customizable
- `agnoster` - Classic powerline style
- `paradox` - Simple and clean
- `pure` - Minimal prompt
- `robbyrussell` - Oh My Zsh default

### Font Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `font` | string | `"MesloLGS NF"` | Nerd Font to install |

Recommended fonts:
- `MesloLGS NF` - Best compatibility
- `JetBrainsMono Nerd Font` - Popular with developers
- `FiraCode Nerd Font` - With ligatures
- `Hack Nerd Font` - Classic terminal font

### Tool Installation

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `install_oh_my_posh` | bool | `true` | Install oh-my-posh |
| `install_zoxide` | bool | `true` | Install zoxide (smart cd) |
| `install_fzf` | bool | `true` | Install fzf (fuzzy finder) |
| `install_bat` | bool | `true` | Install bat (better cat) |
| `install_eza` | bool | `true` | Install eza (modern ls) |
| `tools` | array | `[]` | Additional tools to install |

### Installation Behavior

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `dry_run` | bool | `false` | Simulate installation without changes |
| `force` | bool | `false` | Overwrite existing installations |
| `skip_checksum` | bool | `false` | Skip SHA256 verification |
| `skip_verification` | bool | `false` | Skip post-install verification |
| `timeout` | duration | `"10m0s"` | Installation timeout |

### Backup and Rollback

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `backup` | bool | `true` | Create backup before installation |
| `rollback` | bool | `false` | Perform rollback operation |
| `rollback_to_original` | bool | `false` | Rollback to original state |

### Advanced Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `config_dir` | string | `"~/.config/savanhi"` | Configuration directory path |

## Example Configurations

### Minimal Configuration

```json
{
  "theme": "powerlevel10k",
  "font": "MesloLGS NF"
}
```

### Full Configuration

```json
{
  "theme": "powerlevel10k",
  "font": "JetBrainsMono Nerd Font",
  "tools": ["zoxide", "fzf", "bat", "eza"],
  "install_oh_my_posh": true,
  "install_zoxide": true,
  "install_fzf": true,
  "install_bat": true,
  "install_eza": true,
  "skip_checksum": false,
  "skip_verification": false,
  "dry_run": false,
  "force": false,
  "timeout": "15m0s",
  "backup": true
}
```

### CI/CD Configuration

For automated installations:

```json
{
  "theme": "agnoster",
  "font": "Hack Nerd Font",
  "install_oh_my_posh": true,
  "install_zoxide": true,
  "install_fzf": false,
  "install_bat": false,
  "install_eza": false,
  "skip_checksum": false,
  "skip_verification": false,
  "backup": true
}
```

## Environment Variables

Savanhi Shell respects these environment variables:

| Variable | Description |
|----------|-------------|
| `XDG_CONFIG_HOME` | Base configuration directory |
| `XDG_CACHE_HOME` | Cache directory |
| `SAVANHI_CONFIG_DIR` | Override configuration directory |
| `SAVANHI_NO_COLOR` | Disable colored output |
| `SAVANHI_LOG_LEVEL` | Set log level (debug, info, warn, error) |

## Using with Non-Interactive Mode

```bash
# Create configuration file
cat > config.json << 'EOF'
{
  "theme": "powerlevel10k",
  "font": "MesloLGS NF",
  "install_zoxide": true,
  "install_fzf": true,
  "dry_run": false
}
EOF

# Run non-interactively
savanhi-shell --non-interactive --config config.json
```

## Configuration Directory Structure

```
~/.config/savanhi/
├── config.json           # Main configuration
├── original-backup.json  # Original system state
├── preferences.json      # User preferences
├── history.json          # Installation history
├── themes/               # Downloaded themes
│   └── powerlevel10k.omp.json
├── backups/              # Timestamped backups
│   └── backup-20240115-120000/
└── cache/                # Download cache
    └── downloads/
```

## Merging Configurations

Command-line flags take precedence over configuration file:

```bash
# Override dry_run from config
savanhi-shell --non-interactive --config config.json --dry-run

# Force installation
savanhi-shell --non-interactive --config config.json --force
```

## See Also

- [Getting Started](getting-started.md)
- [Troubleshooting](troubleshooting.md)
- [Key Bindings](keybindings.md)