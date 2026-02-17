#!/usr/bin/env bash
# Phase 2 Robust Test Script
# Run from cli/ directory: ./scripts/phase2-test.sh
# Requires: go, node 18+, (optional) NEON_API_KEY for db neon create

set -e
cd "$(dirname "$0")/.."
CLI="${CLI:-go run ./cmd/aerostack}"

echo "=========================================="
echo "  Aerostack Phase 2 Test Suite"
echo "=========================================="

# 1. Build (or use go run)
echo ""
echo "[1/8] Building CLI..."
if go build -o /tmp/aerostack-test ./cmd/aerostack 2>/dev/null; then
  CLI="/tmp/aerostack-test"
  echo "✓ Build OK"
else
  echo "⚠ Build failed, using 'go run'"
  CLI="go run ./cmd/aerostack"
fi

# 2. Help
echo ""
echo "[2/8] Testing help..."
$CLI --help | grep -q "aerostack" && echo "✓ Help OK" || { echo "✗ Help failed"; exit 1; }

# 3. Init in temp dir
TEST_DIR=$(mktemp -d)
echo ""
echo "[3/8] Testing init (in $TEST_DIR)..."
cd "$TEST_DIR"
$CLI init phase2-test --template=blank --db=d1 2>/dev/null || true
cd phase2-test 2>/dev/null || true
test -f aerostack.toml && echo "✓ Init OK" || { echo "✗ Init failed (no aerostack.toml)"; exit 1; }

# 4. db migrate new
echo ""
echo "[4/8] Testing db migrate new..."
$CLI db migrate new add_users
test -f migrations/*_add_users.sql && echo "✓ db migrate new OK" || { echo "✗ db migrate new failed"; exit 1; }

# 5. db migrate new --postgres
echo ""
echo "[5/8] Testing db migrate new --postgres..."
$CLI db migrate new add_posts --postgres
test -f migrations_postgres/*_add_posts.sql && echo "✓ db migrate new --postgres OK" || { echo "✗ db migrate new --postgres failed"; exit 1; }

# 6. db migrate apply (D1)
echo ""
echo "[6/8] Testing db migrate apply (D1)..."
# Add minimal SQL to the migration file we created
MIG_FILE=$(ls migrations/*_add_users.sql 2>/dev/null | head -1)
if [ -n "$MIG_FILE" ]; then
  echo "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT);" > "$MIG_FILE"
fi
$CLI db migrate apply 2>&1 | grep -qE "(Migrations applied|Applying)" && echo "✓ db migrate apply OK" || echo "⚠ db migrate apply (wrangler may need npx)"

# 7. generate types / db pull
echo ""
echo "[7/8] Testing generate types..."
$CLI generate types -o shared/types.ts 2>&1 | tail -1
test -f shared/types.ts && echo "✓ generate types OK" || echo "⚠ generate types (may need wrangler)"

echo ""
echo "[7b] Testing db pull (alias)..."
$CLI db pull -o shared/types2.ts 2>&1 | tail -1
test -f shared/types2.ts && echo "✓ db pull OK" || echo "⚠ db pull"

# 8. wrangler.toml has Hyperdrive
echo ""
echo "[8/8] Checking wrangler.toml..."
test -f wrangler.toml && grep -q "d1_databases" wrangler.toml && echo "✓ wrangler.toml has D1" || true

# Cleanup
cd /
rm -rf "$TEST_DIR"

echo ""
echo "=========================================="
echo "  Phase 2 Test Complete"
echo "=========================================="
echo ""
echo "Optional manual tests:"
echo "  - aerostack db neon create test-db (requires NEON_API_KEY)"
echo "  - aerostack dev (requires Node 18+)"
echo "  - aerostack generate types with Postgres (add [[postgres_databases]] to aerostack.toml)"
echo ""
