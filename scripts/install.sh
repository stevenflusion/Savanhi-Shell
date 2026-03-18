#!/bin/bash

# Savanhi Shell Install Script
# https://github.com/savanhi/shell
#
# This script installs Savanhi Shell on macOS and Linux.
# Usage: curl -fsSL https://raw.githubusercontent.com/savanhi/shell/main/scripts/install.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="savanhi/shell"
BINARY_NAME="savanhi-shell"
INSTALL_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.config/savanhi"

# Print functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*)    echo "darwin" ;;
        Linux*)     echo "linux" ;;
        CYGWIN*)    echo "windows" ;;
        MINGW*)     echo "windows" ;;
        *)          error "Unsupported OS: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        armv7l)        echo "arm" ;;
        *)             error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest version
get_latest_version() {
    info "Fetching latest version..."
    
    if command -v curl &> /dev/null; then
        curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget &> /dev/null; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "curl or wget required"
    fi
}

# Download binary
download_binary() {
    local version="$1"
    local os="$2"
    local arch="$3"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${os}-${arch}"
    
    if [ "$os" = "windows" ]; then
        download_url="${download_url}.exe"
    fi
    
    info "Downloading ${BINARY_NAME} ${version} for ${os}/${arch}..."
    
    local tmp_file="/tmp/${BINARY_NAME}"
    
    if command -v curl &> /dev/null; then
        curl -sSL -o "$tmp_file" "$download_url"
    elif command -v wget &> /dev/null; then
        wget -qO "$tmp_file" "$download_url"
    else
        error "curl or wget required"
    fi
    
    if [ ! -f "$tmp_file" ]; then
        error "Failed to download binary"
    fi
    
    echo "$tmp_file"
}

# Verify checksum
verify_checksum() {
    local binary_path="$1"
    local os="$2"
    local arch="$3"
    local version="$4"
    
    info "Verifying checksum..."
    
    local checksum_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${version}-checksums.txt"
    local tmp_checksum="/tmp/checksums.txt"
    
    if command -v curl &> /dev/null; then
        curl -sSL -o "$tmp_checksum" "$checksum_url" 2>/dev/null || true
    elif command -v wget &> /dev/null; then
        wget -qO "$tmp_checksum" "$checksum_url" 2>/dev/null || true
    fi
    
    if [ -f "$tmp_checksum" ]; then
        if command -v sha256sum &> /dev/null; then
            local expected_checksum=$(grep "${BINARY_NAME}-${os}-${arch}" "$tmp_checksum" | cut -d' ' -f1)
            local actual_checksum=$(sha256sum "$binary_path" | cut -d' ' -f1)
            
            if [ "$expected_checksum" = "$actual_checksum" ]; then
                success "Checksum verified"
            else
                warn "Checksum mismatch! Expected: $expected_checksum, Got: $actual_checksum"
                warn "Proceeding without checksum verification"
            fi
        else
            warn "sha256sum not found, skipping checksum verification"
        fi
    else
        warn "Checksums file not found, skipping verification"
    fi
}

# Install binary
install_binary() {
    local binary_path="$1"
    
    info "Installing ${BINARY_NAME}..."
    
    # Create installation directory
    mkdir -p "$INSTALL_DIR"
    
    # Move binary
    mv "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    
    success "Binary installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

# Add to PATH
add_to_path() {
    # Check if already in PATH
    if [[ ":$PATH:" == *":${INSTALL_DIR}:"* ]]; then
        info "${INSTALL_DIR} is already in PATH"
        return
    fi
    
    info "Adding ${INSTALL_DIR} to PATH..."
    
    # Detect shell
    local shell_config=""
    if [ -n "$ZSH_VERSION" ]; then
        shell_config="${HOME}/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        if [ -f "${HOME}/.bashrc" ]; then
            shell_config="${HOME}/.bashrc"
        else
            shell_config="${HOME}/.bash_profile"
        fi
    fi
    
    if [ -n "$shell_config" ] && [ -f "$shell_config" ]; then
        # Check if already added
        if grep -q 'export PATH="${HOME}/.local/bin:$PATH"' "$shell_config" 2>/dev/null; then
            success "PATH already configured in ${shell_config}"
        else
            echo '' >> "$shell_config"
            echo '# Added by Savanhi Shell installer' >> "$shell_config"
            echo 'export PATH="${HOME}/.local/bin:$PATH"' >> "$shell_config"
            success "Added ${INSTALL_DIR} to PATH in ${shell_config}"
            warn "Please restart your shell or run: source ${shell_config}"
        fi
    else
        warn "Could not determine shell config file"
        warn "Please add to PATH manually: export PATH=\"\${HOME}/.local/bin:\$PATH\""
    fi
}

# Create config directory
setup_config() {
    if [ ! -d "$CONFIG_DIR" ]; then
        info "Creating config directory..."
        mkdir -p "$CONFIG_DIR"
        success "Config directory created at ${CONFIG_DIR}"
    fi
}

# Verify installation
verify_installation() {
    info "Verifying installation..."
    
    if "${INSTALL_DIR}/${BINARY_NAME}" --version &>/dev/null; then
        success "Installation successful!"
        echo ""
        "${INSTALL_DIR}/${BINARY_NAME}" --version
        echo ""
        info "Run 'savanhi-shell' to start the interactive TUI"
        info "Run 'savanhi-shell --help' for more options"
    else
        error "Installation verification failed"
    fi
}

# Main installation
main() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║     Savanhi Shell Installer           ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
    echo ""
    
    # Check prerequisites
    if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
        error "curl or wget required. Please install one and try again."
    fi
    
    # Detect system
    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=$(get_latest_version)
    
    info "System: ${OS}/${ARCH}"
    info "Version: ${VERSION}"
    echo ""
    
    # Download
    BINARY_PATH=$(download_binary "$VERSION" "$OS" "$ARCH")
    
    # Verify checksum
    verify_checksum "$BINARY_PATH" "$OS" "$ARCH" "$VERSION"
    
    # Install
    install_binary "$BINARY_PATH"
    
    # Setup
    add_to_path
    setup_config
    
    # Verify
    verify_installation
    
    echo ""
    success "Installation complete!"
    echo ""
    echo -e "Next steps:"
    echo -e "  ${GREEN}1.${NC} Restart your shell or run: ${YELLOW}exec \$SHELL${NC}"
    echo -e "  ${GREEN}2.${NC} Start Savanhi Shell: ${YELLOW}savanhi-shell${NC}"
    echo -e "  ${GREEN}3.${NC} For help: ${YELLOW}savanhi-shell --help${NC}"
    echo ""
}

# Run main
main "$@"