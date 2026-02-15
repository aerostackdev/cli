#!/usr/bin/env bash
# Test Aerostack-first deploy flow
# Prerequisites:
#   1. API running: cd packages/api && npm run dev (or wrangler dev)
#   2. Create project + API key via admin dashboard at http://localhost:5173 (or similar)
#   3. Set AEROSTACK_API_URL=http://localhost:8787
#   4. Set AEROSTACK_API_KEY=ak_xxx (your project API key)
#
# Run: ./cli/scripts/test-aerostack-deploy.sh

set -e
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CLI="$REPO_ROOT/cli/bin/aerostack"
DEMO="$REPO_ROOT/demo-aerostack"

cd "$REPO_ROOT"

echo "=== Aerostack-first deploy test ==="
echo ""

# Build CLI
echo "[1] Building CLI..."
cd "$REPO_ROOT/cli" && go build -o bin/aerostack ./cmd/aerostack && cd "$REPO_ROOT"
echo "✓ CLI built"
echo ""

# Check API key
if [[ -z "$AEROSTACK_API_KEY" ]]; then
  echo "⚠ AEROSTACK_API_KEY not set. To test:"
  echo "  1. Start API: cd packages/api && npm run dev"
  echo "  2. Create project + API key in admin dashboard"
  echo "  3. export AEROSTACK_API_URL=http://localhost:8787"
  echo "  4. export AEROSTACK_API_KEY=ak_xxx"
  echo ""
  echo "Testing CLI commands without API..."
  $CLI whoami
  exit 0
fi

# Ensure demo project exists
if [[ ! -f "$DEMO/aerostack.toml" ]]; then
  echo "[2] Creating demo project..."
  $CLI init demo-aerostack --template=api
  cd "$DEMO"
else
  cd "$DEMO"
  echo "[2] Using existing demo project"
fi
echo ""

# Login (stores credentials)
echo "[3] Login..."
export AEROSTACK_API_KEY
$CLI login
echo ""

# Get project ID from validate (or use PROJECT_ID env)
echo "[4] Getting project info..."
if [[ -n "$PROJECT_ID" ]]; then
  echo "  Using PROJECT_ID from env: $PROJECT_ID"
else
  VALIDATE_RESP=$(curl -s -X POST "${AEROSTACK_API_URL:-http://localhost:8787}/api/v1/cli/validate" \
    -H "X-API-Key: $AEROSTACK_API_KEY")
  if echo "$VALIDATE_RESP" | grep -q '"projectId"'; then
    PROJECT_ID=$(echo "$VALIDATE_RESP" | grep -o '"projectId"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | cut -d'"' -f4)
  else
    echo "✗ Validate failed: $VALIDATE_RESP"
    echo "  Is the API running? Set AEROSTACK_API_URL (default http://localhost:8787)"
    exit 1
  fi
fi
echo "  Project ID: $PROJECT_ID"
echo ""

# Link
echo "[5] Link project..."
$CLI link "$PROJECT_ID"
echo ""

# Whoami
echo "[6] Whoami..."
$CLI whoami
echo ""

# Deploy
echo "[7] Deploy to Aerostack..."
$CLI deploy --env staging
echo ""
echo "✓ Aerostack-first deploy test complete!"
