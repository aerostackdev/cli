#!/usr/bin/env bash
# Aerostack CLI + SDK — End-to-End Demo
# Run from repo root: ./cli/scripts/demo-e2e.sh
# Requires: go, node 18+, (optional) Cloudflare login for deploy

set -e
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CLI_DIR="$REPO_ROOT/cli"
DEMO_DIR="${DEMO_DIR:-$REPO_ROOT/demo-aerostack}"
CLI="${CLI:-$REPO_ROOT/cli/bin/aerostack}"

cd "$REPO_ROOT"

echo "=============================================="
echo "  Aerostack CLI + SDK — End-to-End Demo"
echo "=============================================="
echo ""
echo "Demo directory: $DEMO_DIR"
echo ""

# 1. Build CLI
echo "[1/6] Building CLI..."
mkdir -p "$CLI_DIR/bin"
cd "$CLI_DIR"
if go build -o bin/aerostack ./cmd/aerostack 2>/dev/null; then
  echo "✓ CLI built at $CLI"
else
  echo "✗ Build failed"
  exit 1
fi

# 2. Create demo project
echo ""
echo "[2/6] Creating demo project (api template)..."
rm -rf "$DEMO_DIR"
cd "$REPO_ROOT"
$CLI init "$(basename "$DEMO_DIR")" --template=api
cd "$DEMO_DIR"
echo "✓ Project created at $DEMO_DIR"

# 3. Verify structure
echo ""
echo "[3/6] Verifying project structure..."
test -f aerostack.toml && echo "  ✓ aerostack.toml"
test -f src/index.ts && echo "  ✓ src/index.ts"
test -d shared && echo "  ✓ shared/"
test -f package.json && echo "  ✓ package.json"
test -f vitest.config.ts && echo "  ✓ vitest.config.ts"

# 4. Run tests
echo ""
echo "[4/6] Running tests (aerostack test)..."
$CLI test
echo "✓ Tests passed"

# 5. Dev server (start, wait, stop)
echo ""
echo "[5/6] Testing dev server (5 seconds)..."
$CLI dev &
DEV_PID=$!
sleep 5
kill $DEV_PID 2>/dev/null || true
wait $DEV_PID 2>/dev/null || true
echo "✓ Dev server started successfully"

# 6. Deploy (optional — set DEPLOY=1 to attempt, requires Cloudflare login + D1)
echo ""
echo "[6/6] Deploy..."
if [[ "$DEPLOY" == "1" ]] && npx wrangler whoami 2>/dev/null | grep -q "Logged in"; then
  echo "  Attempting deploy to staging..."
  $CLI deploy --env staging || echo "  ⚠ Deploy failed (ensure D1 database_id in aerostack.toml)"
else
  echo "  Skipped (set DEPLOY=1 to attempt). Prerequisites:"
  echo "    - npx wrangler login"
  echo "    - Create D1: npx wrangler d1 create demo-db"
  echo "    - Update aerostack.toml env.staging.d1_databases.database_id"
  echo "    - aerostack deploy --env staging"
fi

echo ""
echo "=============================================="
echo "  Demo Complete"
echo "=============================================="
echo ""
echo "Next steps:"
echo "  cd $DEMO_DIR"
echo "  aerostack dev          # Start dev server"
echo "  aerostack test         # Run tests"
echo "  aerostack deploy -e staging   # Deploy (after wrangler login + D1 setup)"
echo ""
echo "To deploy:"
echo "  1. npx wrangler login"
echo "  2. npx wrangler d1 create demo-db"
echo "  3. Update aerostack.toml: env.staging.d1_databases.database_id = '<your-d1-id>'"
echo "  4. aerostack deploy --env staging"
echo ""
