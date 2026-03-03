#!/bin/sh
# Aerostack CLI install script
# Usage: curl -fsSL https://get.aerostack.dev | sh
# Or with version: curl -fsSL https://get.aerostack.dev | VERSION=v1.0.0 sh

set -e

REPO="aerostackdev/cli"
BINARY="aerostack"
INSTALL_DIR="${AEROSTACK_INSTALL_DIR:-$HOME/.aerostack/bin}"
RELEASES_URL="https://github.com/${REPO}/releases"

# Basic TTY-aware styling
is_tty() {
  [ -t 1 ]
}

if is_tty; then
  ESC="$(printf '\033')"
  COLOR_BRAND="${ESC}[95m"   # bright magenta
  COLOR_INFO="${ESC}[36m"    # cyan
  COLOR_SUCCESS="${ESC}[92m" # bright green
  COLOR_WARN="${ESC}[93m"    # yellow
  COLOR_ERROR="${ESC}[91m"   # red
  COLOR_RESET="${ESC}[0m"
else
  COLOR_BRAND=""
  COLOR_INFO=""
  COLOR_SUCCESS=""
  COLOR_WARN=""
  COLOR_ERROR=""
  COLOR_RESET=""
fi

brand() {
  printf "%s%s%s\n" "$COLOR_BRAND" "$1" "$COLOR_RESET" >&2
}

info() {
  printf "%s%s%s\n" "$COLOR_INFO" "$1" "$COLOR_RESET" >&2
}

success() {
  printf "%s%s%s\n" "$COLOR_SUCCESS" "$1" "$COLOR_RESET" >&2
}

warn() {
  printf "%s%s%s\n" "$COLOR_WARN" "$1" "$COLOR_RESET" >&2
}

error() {
  printf "%s%s%s\n" "$COLOR_ERROR" "$1" "$COLOR_RESET" >&2
}

# Detect OS and architecture
detect_platform() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)

  case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *)
      error "Unsupported OS: $OS"
      exit 1
      ;;
  esac

  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    i386|i686) ARCH="386" ;;
    *)
      error "Unsupported architecture: $ARCH"
      exit 1
      ;;
  esac
}

# Get latest version from GitHub API
get_latest_version() {
  if [ -n "$VERSION" ]; then
    echo "$VERSION"
    return
  fi
  info "Detecting latest Aerostack CLI version..."

  RESPONSE=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null || true)
  TAG_LINE=$(printf '%s\n' "$RESPONSE" | grep '"tag_name":' | head -n 1 || true)
  VERSION=$(printf '%s\n' "$TAG_LINE" | sed -E 's/.*"v?([^"]*)".*/\1/' 2>/dev/null || true)


  if [ -z "$VERSION" ]; then
    error "Unable to detect latest version from GitHub releases."
    error "Please check your network connection or visit:"
    error "  ${RELEASES_URL}"
    exit 1
  fi
  echo "$VERSION"
}

# Download and install
install() {
  VERSION=$(get_latest_version)
  # Strip 'v' prefix if present for asset name
  VER="${VERSION#v}"
  
  if [ "$OS" = "windows" ]; then
    EXT="zip"
    ARCHIVE="${BINARY}_${VER}_${OS}_${ARCH}.${EXT}"
  else
    EXT="tar.gz"
    ARCHIVE="${BINARY}_${VER}_${OS}_${ARCH}.${EXT}"
  fi

  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VER}/${ARCHIVE}"

  UPGRADE=""
  if [ "$OS" != "windows" ] && [ -f "$INSTALL_DIR/$BINARY" ]; then
    UPGRADE=" (upgrade)"
  fi
  if [ "$OS" = "windows" ] && [ -f "$INSTALL_DIR/${BINARY}.exe" ]; then
    UPGRADE=" (upgrade)"
  fi

  brand "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  brand "  Aerostack CLI Installer"
  brand "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  if [ -n "$UPGRADE" ]; then
    info "Upgrading existing install..."
  fi
  info "Version: v${VERSION}"
  info "Platform: $OS • $ARCH"
  info "Download:"
  info "  $DOWNLOAD_URL"

  TMPDIR=$(mktemp -d)
  trap "rm -rf $TMPDIR" EXIT

  if is_tty; then
    info ""
    info "Downloading Aerostack CLI binary..."
  else
    echo "Downloading Aerostack CLI binary..."
  fi

  CURL_FLAGS="-fL"
  if is_tty; then
    # Show progress for humans; avoid in non-TTY (e.g. CI logs)
    CURL_FLAGS="$CURL_FLAGS -#"
  else
    CURL_FLAGS="$CURL_FLAGS -s"
  fi

  if ! curl $CURL_FLAGS -o "$TMPDIR/$ARCHIVE" "$DOWNLOAD_URL"; then
    error "Download failed. The release may not exist yet or network is unavailable."
    error "Check GitHub releases at:"
    error "  ${RELEASES_URL}"
    exit 1
  fi

  mkdir -p "$INSTALL_DIR"

  if [ "$OS" = "windows" ]; then
    if is_tty; then
      info ""
      info "Installing Aerostack CLI..."
    fi
    unzip -o -q "$TMPDIR/$ARCHIVE" -d "$TMPDIR"
    mv "$TMPDIR/${BINARY}.exe" "$INSTALL_DIR/${BINARY}.exe"
    chmod +x "$INSTALL_DIR/${BINARY}.exe"
    echo ""
    success "Installed to $INSTALL_DIR/${BINARY}.exe"
    echo "Add to PATH: $INSTALL_DIR"
  else
    if is_tty; then
      info ""
      info "Installing Aerostack CLI..."
    fi
    tar -xzf "$TMPDIR/$ARCHIVE" -C "$TMPDIR"
    mv "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"
    chmod +x "$INSTALL_DIR/$BINARY"
    echo ""
    success "Installed to $INSTALL_DIR/$BINARY"
    add_to_path
  fi

  echo ""
  brand "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  success "Aerostack CLI installed successfully ✅"
  brand "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
  echo "Next steps:"
  echo "  1) Verify the install:"
  echo "       aerostack --version"
  echo "  2) If 'command not found', restart your terminal"
  echo "     or 'source' your shell config file."
}

add_to_path() {
  SHELL_RC=""
  if [ -f "$HOME/.zshrc" ]; then
    SHELL_RC="$HOME/.zshrc"
  elif [ -f "$HOME/.bashrc" ]; then
    SHELL_RC="$HOME/.bashrc"
  elif [ -f "$HOME/.bash_profile" ]; then
    SHELL_RC="$HOME/.bash_profile"
  elif [ -f "$HOME/.profile" ]; then
    SHELL_RC="$HOME/.profile"
  else
    SHELL_RC="$HOME/.profile"
    touch "$SHELL_RC"
  fi
  if grep -q ".aerostack/bin" "$SHELL_RC" 2>/dev/null; then
    info "PATH already configured in $SHELL_RC"
  else
    echo "" >> "$SHELL_RC"
    echo "# Aerostack CLI" >> "$SHELL_RC"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"
    info "Added Aerostack CLI to PATH in $SHELL_RC"
    info "Run 'source $SHELL_RC' or restart your terminal to use 'aerostack'."
  fi
}

detect_platform
install
