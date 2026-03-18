# Troubleshooting Guide

This guide covers common issues and their solutions.

## Installation Issues

### "Permission denied" errors

**Problem**: You see permission errors during installation.

**Solutions**:

1. Make sure you have write permissions:
   ```bash
   ls -la ~/.local/bin
   ls -la ~/.config/savanhi
   ```

2. If directories don't exist, create them:
   ```bash
   mkdir -p ~/.local/bin
   mkdir -p ~/.config/savanhi
   ```

3. If installing to system directories, use sudo:
   ```bash
   sudo savanhi-shell --non-interactive
   ```

### "Command not found: savanhi-shell"

**Problem**: The command isn't found after installation.

**Solutions**:

1. Check if it's in your PATH:
   ```bash
   which savanhi-shell
   echo $PATH
   ```

2. Add to PATH if needed (add to `.zshrc` or `.bashrc`):
   ```bash
   export PATH="$HOME/.local/bin:$PATH"
   ```

3. Rehash shell commands:
   ```bash
   hash -r  # for zsh/bash
   ```

### Download failures

**Problem**: Downloads fail with network errors.

**Solutions**:

1. Check your internet connection
2. Try a different mirror:
   ```bash
   export GITHUB_MIRROR="https://github.com"  # or your mirror
   ```
3. Use a proxy:
   ```bash
   export HTTPS_PROXY="http://your-proxy:port"
   ```
4. Increase timeout:
   ```bash
   savanhi-shell --non-interactive --timeout 30m
   ```

## Shell Issues

### RC file modifications not appearing

**Problem**: Changes aren't reflected in your shell.

**Solutions**:

1. Reload your RC file:
   ```bash
   source ~/.zshrc  # or ~/.bashrc
   ```

2. Start a new shell session:
   ```bash
   exec $SHELL
   ```

3. Check if the markers are present:
   ```bash
   grep "savanhi" ~/.zshrc
   ```

### Duplicate entries in RC file

**Problem**: Running Savanhi Shell multiple times creates duplicate entries.

**Solutions**:

1. The markers should prevent duplicates. Check for proper markers:
   ```bash
   grep -A2 "savanhi" ~/.zshrc
   ```

2. If duplicates exist, remove them manually or use rollback:
   ```bash
   savanhi-shell --rollback
   ```

### Shell prompts not loading

**Problem**: oh-my-posh prompt isn't showing.

**Solutions**:

1. Verify oh-my-posh is installed:
   ```bash
   which oh-my-posh
   oh-my-posh --version
   ```

2. Check the init command:
   ```bash
   echo $PROMPT_COMMAND  # bash
   echo $PROMPT  # zsh
   ```

3. Verify the theme file:
   ```bash
   cat ~/.config/savanhi/themes/$(cat ~/.config/savanhi/preferences.json | grep theme).omp.json
   ```

## Font Issues

### Icons showing as boxes or question marks

**Problem**: Nerd Font icons aren't rendering.

**Solutions**:

1. Verify font installation:
   ```bash
   ls ~/Library/Fonts/          # macOS
   ls ~/.local/share/fonts/      # Linux
   ```

2. Refresh font cache (Linux):
   ```bash
   fc-cache -fv
   ```

3. **Terminal configuration**: You must configure your terminal to use the Nerd Font:
   - **iTerm2**: Preferences → Profiles → Text → Font → Select Nerd Font
   - **Alacritty**: Add `family: "MesloLGS NF"` to your `alacritty.yml`
   - **WezTerm**: Set `font = wezterm.font("MesloLGS NF")`
   - **Kitty**: Add `font_family MesloLGS NF` to `kitty.conf`

4. Restart your terminal after font installation.

### Font looks wrong or distorted

**Problem**: Font appearance issues.

**Solutions**:

1. Try a different Nerd Font:
   ```bash
   savanhi-shell  # Select a different font
   ```

2. Check terminal font settings for conflicts.

3. Clear font cache:
   ```bash
   # macOS
   sudo atsutil databases -remove
   
   # Linux
   rm -rf ~/.cache/fontconfig
   fc-cache -fv
   ```

## Tool Issues

### zoxide not working

**Problem**: zoxide commands don't work.

**Solutions**:

1. Verify installation:
   ```bash
   which zoxide
   zoxide --version
   ```

2. Check if init command is in RC file:
   ```bash
   grep "zoxide init" ~/.zshrc
   ```

3. If missing, add manually:
   ```bash
   eval "$(zoxide init zsh)"  # zsh
   # or
   eval "$(zoxide init bash)"  # bash
   ```

### fzf not working

**Problem**: fzf keybindings don't work.

**Solutions**:

1. Verify installation:
   ```bash
   which fzf
   ```

2. Check for init in RC:
   ```bash
   grep "fzf" ~/.zshrc
   ```

3. Source fzf manually:
   ```bash
   source /usr/share/fzf/key-bindings.zsh
   source /usr/share/fzf/completion.zsh
   ```

### bat showing raw output

**Problem**: bat isn't coloring output.

**Solutions**:

1. Verify installation:
   ```bash
   bat --version
   ```

2. Check alias:
   ```bash
   alias cat
   ```

3. Test directly:
   ```bash
   bat --paging=never /etc/passwd
   ```

### eza not working

**Problem**: ls commands still use old ls.

**Solutions**:

1. Verify installation:
   ```bash
   which eza
   eza --version
   ```

2. Check alias:
   ```bash
   alias ls
   ```

3. Add alias manually:
   ```bash
   alias ls='eza'
   alias ll='eza -l'
   alias la='eza -la'
   ```

## Rollback Issues

### Rollback doesn't restore original state

**Problem**: Rollback doesn't fully restore your system.

**Solutions**:

1. Check for backup:
   ```bash
   ls ~/.config/savanhi/backups/
   ls ~/.config/savanhi/original-backup.json
   ```

2. Manual restore from backup:
   ```bash
   # View backup
   cat ~/.config/savanhi/original-backup.json
   
   # Restore RC file manually
   cp ~/.zshrc.backup ~/.zshrc
   ```

3. Full uninstall and reinstall if needed:
   ```bash
   savanhi-shell --rollback-original
   # Remove Savanhi completely
   rm -rf ~/.config/savanhi
   rm -rf ~/.local/share/savanhi
   ```

## Error Codes

| Code | Meaning | Solution |
|------|---------|----------|
| E001x | Configuration errors | Check config file syntax |
| E002x | Detection errors | Verify system is supported |
| E003x | Installation errors | Check logs, try --verbose |
| E004x | Rollback errors | Manual restore may be needed |
| E005x | Preview errors | Try --non-interactive mode |
| E006x | Persistence errors | Check file permissions |
| E007x | Shell errors | Verify shell is supported |
| E008x | TUI errors | Check terminal compatibility |
| E009x | Network errors | Check internet connection |
| E010x | System errors | Check permissions, disk space |

## Getting Help

If your issue isn't covered here:

1. **Verbose mode**: Run with `--verbose` for more details:
   ```bash
   savanhi-shell --non-interactive --verbose --config config.json
   ```

2. **Check logs**:
   ```bash
   cat ~/.config/savanhi/logs/latest.log
   ```

3. **GitHub Issues**: [Report a bug](https://github.com/savanhi/shell/issues/new?template=bug_report.md)

4. **Discussions**: [Ask a question](https://github.com/savanhi/shell/discussions)

## Debug Mode

Enable debug logging:

```bash
export SAVANHI_LOG_LEVEL=debug
savanhi-shell --non-interactive --verbose
```

Debug logs are saved to `~/.config/savanhi/logs/`.