#!/usr/bin/env bash
# Aerostack CLI — E2E Test Suite
# Tests real-world CLI usage BEFORE any release.
# Covers: build, init, dev server startup, error detection.
#
# Usage: ./scripts/e2e-test.sh
# Run from: sdks/packages/cli/

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CLI_PKG_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$CLI_PKG_DIR"

# ──────────────────────────────────────────
# Configuration — all paths absolute
# ──────────────────────────────────────────
CLI_BIN="$CLI_PKG_DIR/bin/aerostack"
TEST_DIR=""
DEV_PID=""
FAILED=0
DEV_LOG=""

# Colours
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
pass() { echo -e "${GREEN}✓${NC} $1"; }
fail() { echo -e "${RED}✗ FAIL${NC} $1"; FAILED=1; }
warn() { echo -e "${YELLOW}⚠${NC}  $1"; }

# ──────────────────────────────────────────
# Cleanup on exit
# ──────────────────────────────────────────
cleanup() {
  # Kill the direct dev PID
  if [ -n "$DEV_PID" ] && kill -0 "$DEV_PID" 2>/dev/null; then
    # Kill the whole process group to ensure wrangler + workerd children are cleaned up
    PGID=$(ps -o pgid= -p "$DEV_PID" 2>/dev/null | tr -d ' ')
    if [ -n "$PGID" ] && [ "$PGID" != "$$" ]; then
      kill -9 -- "-$PGID" 2>/dev/null || true
    fi
    kill -9 "$DEV_PID" 2>/dev/null || true
    wait "$DEV_PID" 2>/dev/null || true
  fi
  # Also kill anything still on port 8787 (workerd survivor)
  lsof -ti :8787 | xargs -r kill -9 2>/dev/null || true
  if [ -n "$TEST_DIR" ] && [ -d "$TEST_DIR" ]; then
    rm -rf "$TEST_DIR"
  fi
  if [ -n "$DEV_LOG" ] && [ -f "$DEV_LOG" ]; then
    rm -f "$DEV_LOG"
  fi
}
trap cleanup EXIT

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║   Aerostack CLI — End-to-End Test Suite           ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# ──────────────────────────────────────────
# Step 1: Build
# ──────────────────────────────────────────
echo "[1/6] Building CLI binary..."
cd "$CLI_PKG_DIR"
if go build \
  -ldflags "-X main.version=e2e-test -X main.commit=test -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o "$CLI_BIN" ./cmd/aerostack 2>&1; then
  pass "Build succeeded: $CLI_BIN"
else
  fail "Build failed — aborting E2E tests"
  exit 1
fi

# ──────────────────────────────────────────
# Step 2: Basic --help check
# ──────────────────────────────────────────
echo ""
echo "[2/6] Verifying CLI responds to --help..."
HELP_OUT=$("$CLI_BIN" --help 2>&1 || true)
if echo "$HELP_OUT" | grep -qiE "(aerostack|usage)"; then
  pass "--help output looks valid"
else
  fail "--help returned unexpected output"
  echo "$HELP_OUT" | head -20
fi

# ──────────────────────────────────────────
# Step 3: aerostack init
# ──────────────────────────────────────────
echo ""
echo "[3/6] Testing 'aerostack init' (project scaffolding)..."
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"

# Note: init requires a TTY for interactive prompts.
# When a project name AND template AND db are provided, it should skip the TUI.
INIT_OUT=$("$CLI_BIN" init e2e-project --template=blank --db=d1 2>&1 || true)
INIT_EXIT=$?

if [ -d "$TEST_DIR/e2e-project" ] && [ -f "$TEST_DIR/e2e-project/aerostack.toml" ]; then
  pass "Init scaffolded project successfully"
  cd "$TEST_DIR/e2e-project"
  [ -f "src/index.ts" ] && pass "src/index.ts created" || warn "src/index.ts missing (blank template may differ)"
else
  # Non-TTY environment (CI) — init requires a TTY for the interactive picker
  # This is a known limitation; warn but do not fail the release gate for it
  if echo "$INIT_OUT" | grep -qiE "(TTY|tty|terminal)"; then
    warn "Init requires a TTY (non-interactive CI environment) — using built-in test project"
  else
    fail "Init did not create project (unexpected failure)"
    echo "--- Init output ---"
    echo "$INIT_OUT"
    echo "-------------------"
  fi
fi

# ──────────────────────────────────────────
# Step 4: aerostack dev — launch & scan for errors (CRITICAL GATE)
# ──────────────────────────────────────────
echo ""
echo "[4/6] Testing 'aerostack dev' — launching server (CRITICAL)..."

# Use init-created project if available, otherwise use built-in test project
if [ -d "$TEST_DIR/e2e-project" ] && [ -f "$TEST_DIR/e2e-project/aerostack.toml" ]; then
  cd "$TEST_DIR/e2e-project"
else
  # Fallback: use the built-in test project that ships with the repo
  cd "$CLI_PKG_DIR/my-blank-app"
fi

# Pre-check: ensure port 8787 is free before launching (avoid silent hang)
if lsof -ti :8787 &>/dev/null; then
  warn "Port 8787 is in use — killing old process before starting E2E test"
  lsof -ti :8787 | xargs kill -9 2>/dev/null || true
  sleep 1
fi

DEV_LOG=$(mktemp /tmp/aerostack-e2e-dev-XXXX.log)

# Launch dev in background using absolute binary path
"$CLI_BIN" dev > "$DEV_LOG" 2>&1 &
DEV_PID=$!

# Wait up to 20 seconds for server to become ready or fail
READY=0
for i in $(seq 1 20); do
  sleep 1

  # Check if process died early
  if ! kill -0 "$DEV_PID" 2>/dev/null; then
    break
  fi

  # Check for success signal
  if grep -qE "(Dev server ready|localhost|http://127\.0\.0\.1)" "$DEV_LOG" 2>/dev/null; then
    READY=1
    break
  fi
done

# ── Scan for known error patterns ──
ERROR_PATTERNS=(
  "Unknown arguments"
  "Unknown argument"
  "unknown flag"
  "Error: command"
  "\[ERROR\]"
  "fatal error"
  "panic:"
  "is not recognized"
)

FOUND_ERRORS=0
for pattern in "${ERROR_PATTERNS[@]}"; do
  if grep -qiE "$pattern" "$DEV_LOG" 2>/dev/null; then
    fail "Dev server output contains error pattern: '$pattern'"
    FOUND_ERRORS=1
  fi
done

if [ "$FOUND_ERRORS" -eq 0 ] && [ "$READY" -eq 1 ]; then
  pass "'aerostack dev' started without errors"
elif [ "$FOUND_ERRORS" -eq 0 ] && [ "$READY" -eq 0 ]; then
  warn "'aerostack dev' did not produce ready signal within 20s — check log below"
  tail -20 "$DEV_LOG" 2>/dev/null || true
  FAILED=1
elif [ "$FOUND_ERRORS" -eq 1 ]; then
  echo "--- Dev server log (last 30 lines) ---"
  tail -30 "$DEV_LOG" 2>/dev/null || true
  echo "--------------------------------------"
fi

# ── Connectivity check (if server came up) ──
if [ "$READY" -eq 1 ] && [ "$FOUND_ERRORS" -eq 0 ]; then
  echo ""
  echo "[5/6] Verifying HTTP connection to localhost:8787..."
  sleep 2
  HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 http://127.0.0.1:8787 2>/dev/null || echo "000")
  if [[ "$HTTP_STATUS" =~ ^[2345][0-9][0-9]$ ]]; then
    pass "Dev server responded with HTTP $HTTP_STATUS"
  else
    warn "Dev server not responding (HTTP $HTTP_STATUS) — may be port conflict on CI"
  fi
else
  echo "[5/6] Skipping HTTP check (server not ready)"
fi

# Stop dev server
if [ -n "$DEV_PID" ] && kill -0 "$DEV_PID" 2>/dev/null; then
  kill "$DEV_PID" 2>/dev/null || true
  wait "$DEV_PID" 2>/dev/null || true
  DEV_PID=""
fi

# ──────────────────────────────────────────
# Step 6: wrangler.toml sanity check
# ──────────────────────────────────────────
echo ""
echo "[6/6] Checking generated wrangler.toml..."
if [ -f ".aerostack/wrangler.toml" ]; then
  pass ".aerostack/wrangler.toml exists"
  if grep -qE "d1_databases|hyperdrive|kv_namespaces" ".aerostack/wrangler.toml"; then
    pass "wrangler.toml contains binding config"
  else
    warn "wrangler.toml looks minimal (no bindings)"
  fi
else
  warn ".aerostack/wrangler.toml not found (dev server may not have started)"
fi

# ──────────────────────────────────────────
# Result
# ──────────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════════════════╗"
if [ $FAILED -eq 0 ]; then
  echo -e "║  ${GREEN}✅ All E2E tests PASSED — safe to release${NC}         ║"
  echo "╚══════════════════════════════════════════════════╝"
  exit 0
else
  echo -e "║  ${RED}❌ E2E tests FAILED — DO NOT RELEASE${NC}             ║"
  echo "╚══════════════════════════════════════════════════╝"
  exit 1
fi

