#!/bin/sh
# Aerostack CLI uninstall script
# Usage: curl -fsSL https://get.aerostack.dev/uninstall.sh | sh

set -e

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
  printf "%s%s%s\n" "$COLOR_BRAND" "$1" "$COLOR_RESET"
}

info() {
  printf "%s%s%s\n" "$COLOR_INFO" "$1" "$COLOR_RESET"
}

success() {
  printf "%s%s%s\n" "$COLOR_SUCCESS" "$1" "$COLOR_RESET"
}

warn() {
  printf "%s%s%s\n" "$COLOR_WARN" "$1" "$COLOR_RESET" >&2
}

error() {
  printf "%s%s%s\n" "$COLOR_ERROR" "$1" "$COLOR_RESET" >&2
}

brand "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
brand "  Aerostack CLI Uninstaller"
brand "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 1. Remove $HOME/.aerostack
AEROSTACK_DIR="$HOME/.aerostack"
if [ -d "$AEROSTACK_DIR" ]; then
    info "Removing Aerostack directory ($AEROSTACK_DIR)..."
    rm -rf "$AEROSTACK_DIR"
    success "Aerostack directory removed."
else
    warn "Aerostack directory not found at $AEROSTACK_DIR."
fi

# 2. Clean up PATH in shell profiles
clean_path() {
    PROFILE="$1"
    if [ -f "$PROFILE" ]; then
        if grep -q ".aerostack/bin" "$PROFILE"; then
            info "Removing Aerostack from PATH in $PROFILE..."
            # Create a backup
            cp "$PROFILE" "${PROFILE}.bak"
            # Remove the lines
            if [ "$(uname)" = "Darwin" ]; then
                # macOS sed requires an empty extension for -i
                sed -i '' '/# Aerostack CLI/d' "$PROFILE"
                sed -i '' '/.aerostack\/bin/d' "$PROFILE"
            else
                sed -i '/# Aerostack CLI/d' "$PROFILE"
                sed -i '/.aerostack\/bin/d' "$PROFILE"
            fi
            success "Updated $PROFILE (backup created at ${PROFILE}.bak)."
        fi
    fi
}

clean_path "$HOME/.zshrc"
clean_path "$HOME/.bashrc"
clean_path "$HOME/.bash_profile"
clean_path "$HOME/.profile"

echo ""
brand "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
success "Aerostack CLI uninstalled successfully ✅"
brand "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
info "Note: You might need to restart your terminal or run 'hash -r' to clear the command cache."
