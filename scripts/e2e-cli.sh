#!/usr/bin/env bash
# Aerostack CLI — E2E Entry Point for CI
# Invokes the full E2E test suite (e2e-test.sh).
#
# Usage: ./scripts/e2e-cli.sh
# Run from: sdks/packages/cli/

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CLI_PKG_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$CLI_PKG_DIR"

if [ -f "$SCRIPT_DIR/e2e-test.sh" ]; then
  exec "$SCRIPT_DIR/e2e-test.sh"
else
  echo "Error: e2e-test.sh not found at $SCRIPT_DIR/e2e-test.sh"
  exit 1
fi
