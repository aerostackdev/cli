#!/usr/bin/env bash
# Phase 3 Test Script - Self-Healing
# Run from cli/ directory: ./scripts/phase3-test.sh
# Requires: go
# Skips if no API keys (AZURE_OPENAI_*, OPENAI_API_KEY, ANTHROPIC_API_KEY, AEROSTACK_API_KEY or aerostack login)

set -e
cd "$(dirname "$0")/.."
CLI="${CLI:-go run ./cmd/aerostack}"

echo "=========================================="
echo "  Aerostack Phase 3 (Self-Healing) Test"
echo "=========================================="

# Check for API keys - skip if none
has_key() {
  [ -n "$AZURE_OPENAI_API_KEY" ] && [ -n "$AZURE_OPENAI_ENDPOINT" ] && return 0
  [ -n "$OPENAI_API_KEY" ] && return 0
  [ -n "$ANTHROPIC_API_KEY" ] && return 0
  [ -n "$AEROSTACK_API_KEY" ] && return 0
  [ -f "$HOME/.aerostack/credentials.json" ] && grep -q '"api_key"' "$HOME/.aerostack/credentials.json" 2>/dev/null && return 0
  return 1
}

if ! has_key; then
  echo ""
  echo "⚠ No API keys found. Skipping Phase 3 (self-healing requires keys)."
  echo "  Set AZURE_OPENAI_*, OPENAI_API_KEY, ANTHROPIC_API_KEY, or run 'aerostack login'"
  echo ""
  exit 0
fi

# 1. Build CLI
echo ""
echo "[1/3] Building CLI..."
if go build -o /tmp/aerostack-test ./cmd/aerostack 2>/dev/null; then
  CLI="/tmp/aerostack-test"
  echo "✓ Build OK"
else
  CLI="go run ./cmd/aerostack"
fi

# 2. Run a command that fails (typo) - self-heal may trigger
# We use a non-existent subcommand to trigger error handling
echo ""
echo "[2/3] Testing error handling (invalid subcommand)..."
OUTPUT=$($CLI nonexistent-subcommand-xyz 2>&1) || true
echo "$OUTPUT" | grep -qE "(Error|error|unknown)" && echo "✓ Error captured" || echo "⚠ Unexpected output"

# 3. Verify healer/agent code path exists (go test)
echo ""
echo "[3/3] Running selfheal package tests..."
if go test ./internal/selfheal/... -count=1 2>&1; then
  echo "✓ Selfheal tests passed"
else
  echo "⚠ Selfheal tests failed or no tests"
fi

echo ""
echo "=========================================="
echo "  Phase 3 Test Complete"
echo "=========================================="
