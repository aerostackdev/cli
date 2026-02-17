#!/usr/bin/env bash
# Phase Verification - runs all implemented phase test scripts
# Run from cli/ directory: ./scripts/verify-all.sh
# Requires: go, node 18+

set -e
cd "$(dirname "$0")/.."

echo "=========================================="
echo "  Aerostack Phase Verification Suite"
echo "=========================================="

FAILED=0

# Phase 2: D1, migrations, generate types (always run)
if [ -f "./scripts/phase2-test.sh" ]; then
  echo ""
  if ./scripts/phase2-test.sh; then
    echo "✓ Phase 2 passed"
  else
    echo "✗ Phase 2 failed"
    FAILED=1
  fi
fi

# Phase 1: PKG index + search
if [ -f "./scripts/phase1-test.sh" ]; then
  echo ""
  if ./scripts/phase1-test.sh; then
    echo "✓ Phase 1 passed"
  else
    echo "✗ Phase 1 failed"
    FAILED=1
  fi
fi

# Phase 3: Self-healing (skip if no API keys)
if [ -f "./scripts/phase3-test.sh" ]; then
  echo ""
  if ./scripts/phase3-test.sh; then
    echo "✓ Phase 3 passed"
  else
    echo "✗ Phase 3 failed"
    FAILED=1
  fi
fi

# Phase 6: Migration converter
if [ -f "./scripts/phase6-test.sh" ]; then
  echo ""
  if ./scripts/phase6-test.sh; then
    echo "✓ Phase 6 passed"
  else
    echo "✗ Phase 6 failed"
    FAILED=1
  fi
fi

# Go unit tests
echo ""
echo "[Go tests] Running go test ./..."
if go test ./... 2>&1; then
  echo "✓ Go tests passed"
else
  echo "✗ Go tests failed"
  FAILED=1
fi

echo ""
echo "=========================================="
if [ $FAILED -eq 0 ]; then
  echo "  All verifications passed"
else
  echo "  Some verifications failed"
  exit 1
fi
echo "=========================================="
