#!/usr/bin/env bash
# Phase 1 Test Script - PKG (Project Knowledge Graph)
# Run from cli/ directory: ./scripts/phase1-test.sh
# Requires: go

set -e
cd "$(dirname "$0")/.."
CLI="${CLI:-go run ./cmd/aerostack}"

echo "=========================================="
echo "  Aerostack Phase 1 (PKG) Test Suite"
echo "=========================================="

# 1. Build CLI
echo ""
echo "[1/4] Building CLI..."
if go build -o /tmp/aerostack-test ./cmd/aerostack 2>/dev/null; then
  CLI="/tmp/aerostack-test"
  echo "✓ Build OK"
else
  echo "⚠ Build failed, using 'go run'"
  CLI="go run ./cmd/aerostack"
fi

# 2. Create minimal project (no init - faster, no npm install)
TEST_DIR=$(mktemp -d)
echo ""
echo "[2/4] Creating minimal project..."
cd "$TEST_DIR"
mkdir -p src
cat > src/index.ts << 'EOF'
export function hello(name: string): string {
  return `Hello, ${name}!`;
}
export const VERSION = "1.0.0";
EOF
echo "✓ Project created"

# 3. Run aerostack index
echo ""
echo "[3/4] Testing aerostack index..."
$CLI index 2>&1
test -f .aerostack/pkg.db && echo "✓ Index OK (pkg.db created)" || { echo "✗ Index failed (no pkg.db)"; exit 1; }

# 4. Verify we can query symbols
echo ""
echo "[4/4] Verifying symbols..."
if [ -f .aerostack/pkg.db ]; then
  SYM_COUNT=$(sqlite3 .aerostack/pkg.db "SELECT COUNT(*) FROM symbols" 2>/dev/null || echo "0")
  echo "  Symbols indexed: $SYM_COUNT"
fi

# Cleanup
cd /
rm -rf "$TEST_DIR"

echo ""
echo "=========================================="
echo "  Phase 1 Test Complete"
echo "=========================================="
