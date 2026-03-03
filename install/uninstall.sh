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
  printf "  %s%s◆%s %sAerostack CLI Uninstaller%s\n" "$COLOR_BOLD" "$COLOR_BRAND" "$COLOR_RESET" "$COLOR_BOLD" "$COLOR_RESET" >&2
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

header

# 1. Remove $HOME/.aerostack
AEROSTACK_DIR="$HOME/.aerostack"
if [ -d "$AEROSTACK_DIR" ]; then
    step "Removing Aerostack directory..."
    info "$AEROSTACK_DIR"
    rm -rf "$AEROSTACK_DIR"
    success "Directory removed."
else
    warn "Directory not found: $AEROSTACK_DIR"
fi

# 2. Clean up PATH in shell profiles
clean_path() {
    PROFILE="$1"
    if [ -f "$PROFILE" ]; then
        if grep -q ".aerostack/bin" "$PROFILE"; then
            step "Cleaning PATH..."
            info "$PROFILE"
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
            success "Removed Aerostack from $PROFILE (backup: ${PROFILE}.bak)"
        fi
    fi
}

echo "" >&2

clean_path "$HOME/.zshrc"
clean_path "$HOME/.bashrc"
clean_path "$HOME/.bash_profile"
clean_path "$HOME/.profile"

echo "" >&2
success "${COLOR_BOLD}Aerostack CLI uninstalled successfully!${COLOR_RESET}"
echo "" >&2
info "Note: You might need to restart your terminal or run 'hash -r' to clear the command cache."
echo "" >&2
