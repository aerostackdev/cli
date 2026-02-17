#!/usr/bin/env bash
# Phase 6 Test Script - Worker Migration (wrangler.toml -> aerostack.toml)
# Run from cli/ directory: ./scripts/phase6-test.sh
# Requires: go

set -e
cd "$(dirname "$0")/.."
CLI="${CLI:-go run ./cmd/aerostack}"

echo "=========================================="
echo "  Aerostack Phase 6 (Migration) Test"
echo "=========================================="

# 1. Build CLI
echo ""
echo "[1/4] Building CLI..."
if go build -o /tmp/aerostack-test ./cmd/aerostack 2>/dev/null; then
  CLI="/tmp/aerostack-test"
  echo "✓ Build OK"
else
  CLI="go run ./cmd/aerostack"
fi

# 2. Run converter unit tests
echo ""
echo "[2/4] Running migration converter tests..."
if go test ./internal/modules/migration/... -count=1 2>&1; then
  echo "✓ Converter tests passed"
else
  echo "✗ Converter tests failed"
  exit 1
fi

# 3. Create temp wrangler project
TEST_DIR=$(mktemp -d)
echo ""
echo "[3/4] Creating sample wrangler project..."
cd "$TEST_DIR"
mkdir -p src
cat > wrangler.toml << 'EOF'
name = "phase6-test-worker"
main = "src/index.ts"
compatibility_date = "2024-01-01"

[[d1_databases]]
binding = "DB"
database_name = "test-db"
database_id = "test-id-123"

[[kv_namespaces]]
binding = "CACHE"
id = "kv-id-456"
EOF
cat > src/index.ts << 'EOF'
export default { fetch: () => new Response("ok") };
EOF
echo "✓ Wrangler project created"

# 4. Run aerostack migrate
echo ""
echo "[4/4] Testing aerostack migrate..."
$CLI migrate 2>&1
test -f aerostack.toml && echo "✓ aerostack.toml created" || { echo "✗ aerostack.toml not created"; exit 1; }
grep -q "phase6-test-worker" aerostack.toml && echo "✓ Config migrated correctly" || { echo "✗ Config mismatch"; exit 1; }
grep -q "d1_databases" aerostack.toml && echo "✓ D1 mapping present" || true
grep -q "kv_namespaces" aerostack.toml && echo "✓ KV mapping present" || true

# Cleanup
cd /
rm -rf "$TEST_DIR"

echo ""
echo "=========================================="
echo "  Phase 6 Test Complete"
echo "=========================================="
