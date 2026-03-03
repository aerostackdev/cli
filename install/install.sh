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
  COLOR_BRAND="${ESC}[36m"   # cyan
  COLOR_SUCCESS="${ESC}[32m" # emerald
  COLOR_WARN="${ESC}[33m"    # amber
  COLOR_ERROR="${ESC}[31m"   # coral
  COLOR_MUTED="${ESC}[90m"   # slate
  COLOR_BOLD="${ESC}[1m"
  COLOR_RESET="${ESC}[0m"
else
  COLOR_BRAND=""
  COLOR_SUCCESS=""
  COLOR_WARN=""
  COLOR_ERROR=""
  COLOR_MUTED=""
  COLOR_BOLD=""
  COLOR_RESET=""
fi

# Design System Printers (stderr to avoid subshell capture)
header() {
  echo "" >&2
  printf "%s──────────────────────────────────────────────────%s\n" "$COLOR_MUTED" "$COLOR_RESET" >&2
  printf "  %s%s◆%s %sAerostack CLI%s\n" "$COLOR_BOLD" "$COLOR_BRAND" "$COLOR_RESET" "$COLOR_BOLD" "$COLOR_RESET" >&2
  printf "%s──────────────────────────────────────────────────%s\n" "$COLOR_MUTED" "$COLOR_RESET" >&2
  echo "" >&2
}

info() {
  printf "  %s%s%s\n" "$COLOR_MUTED" "$1" "$COLOR_RESET" >&2
}

step() {
  printf "  %s%s◆%s %s\n" "$COLOR_BOLD" "$COLOR_BRAND" "$COLOR_RESET" "$1" >&2
}

success() {
  printf "  %s✓%s %s\n" "$COLOR_SUCCESS" "$COLOR_RESET" "$1" >&2
}

warn() {
  printf "  %s⚠%s %s\n" "$COLOR_WARN" "$COLOR_RESET" "$1" >&2
}

error() {
  printf "  %s✗%s %s\n" "$COLOR_ERROR" "$COLOR_RESET" "$1" >&2
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

# Get latest version from GitHub API (silent operation)
get_latest_version() {
  if [ -n "$VERSION" ]; then
    echo "$VERSION"
    return
  fi

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

  header

  step "Version     ${COLOR_MUTED}v${VERSION}${UPGRADE}${COLOR_RESET}"
  step "Platform    ${COLOR_MUTED}${OS} • ${ARCH}${COLOR_RESET}"
  
  TMPDIR=$(mktemp -d)
  trap "rm -rf $TMPDIR" EXIT

  echo "" >&2
  step "Downloading..."
  info "$DOWNLOAD_URL"

  CURL_FLAGS="-fL"
  if is_tty; then
    CURL_FLAGS="$CURL_FLAGS -#"
  else
    CURL_FLAGS="$CURL_FLAGS -s"
  fi

  if ! curl $CURL_FLAGS -o "$TMPDIR/$ARCHIVE" "$DOWNLOAD_URL"; then
    error "Download failed. Check GitHub releases at: ${RELEASES_URL}"
    exit 1
  fi

  mkdir -p "$INSTALL_DIR"

  echo "" >&2
  step "Installing..."
  
  if [ "$OS" = "windows" ]; then
    unzip -o -q "$TMPDIR/$ARCHIVE" -d "$TMPDIR"
    mv "$TMPDIR/${BINARY}.exe" "$INSTALL_DIR/${BINARY}.exe"
    chmod +x "$INSTALL_DIR/${BINARY}.exe"
    success "Installed to $INSTALL_DIR/${BINARY}.exe"
  else
    tar -xzf "$TMPDIR/$ARCHIVE" -C "$TMPDIR"
    mv "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"
    chmod +x "$INSTALL_DIR/$BINARY"
    success "Installed to $INSTALL_DIR/$BINARY"
    add_to_path
  fi

  echo "" >&2
  success "${COLOR_BOLD}Aerostack CLI installed successfully!${COLOR_RESET}"
  echo "" >&2
  info "Next steps:"
  info "1) Verify the install by running:"
  printf "     %s%saerostack --version%s\n" "$COLOR_BOLD" "$COLOR_BRAND" "$COLOR_RESET" >&2
  info "2) If 'command not found', restart your terminal or source your shell profile."
  echo "" >&2
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
    success "Added Aerostack CLI to PATH in $SHELL_RC"
  fi
}

detect_platform
install
