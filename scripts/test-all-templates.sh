#!/usr/bin/env bash
# Aerostack CLI — E2E Template Verification
# Tests all templates by initializing and running their internal tests.

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

pass() { echo -e "${GREEN}✓${NC} $1"; }
fail() { echo -e "${RED}✗ FAIL${NC} $1"; }
info() { echo -e "${BOLD}$1${NC}"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }

REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
CLI_BIN=$(realpath "../../packages/cli/bin/aerostack")
export CLOUDFLARE_HYPERDRIVE_LOCAL_CONNECTION_STRING_PG="postgresql://user:pass@127.0.0.1:5432/db"
export CLOUDFLARE_HYPERDRIVE_LOCAL_CONNECTION_STRING_DB="postgresql://user:pass@127.0.0.1:5432/db"
TEMPLATES_DIR="$REPO_ROOT/packages/cli/internal/templates/templates"
TEST_ROOT="$REPO_ROOT/test-e2e-$(date +%H%M%S)"
# Clear old ones
rm -rf "$REPO_ROOT/test-e2e-*"

echo "=============================================="
echo "  Aerostack CLI — Template E2E Verification"
echo "=============================================="
echo ""

# Force build CLI
info "Building CLI..."
cd "$REPO_ROOT/packages/cli"
go build -o bin/aerostack ./cmd/aerostack

rm -rf "$TEST_ROOT"
mkdir -p "$TEST_ROOT"

# Get list of templates
TEMPLATES=$(ls "$TEMPLATES_DIR")

FAILED_TEMPLATES=()

for template in $TEMPLATES; do
    echo ""
    info "Testing template: $template"
    
    # 1. Aerostack Init
    echo "  [1/2] Initializing..."
    cd "$TEST_ROOT"
    if ! "$CLI_BIN" init "$template" --template="$template" --db=d1 > /dev/null 2>&1; then
        fail "Initialization failed for $template"
        FAILED_TEMPLATES+=("$template")
        continue
    fi
    pass "Initialized"

    cd "$template"

    # 2. Aerostack Test
    echo "  [2/2] Running tests..."
    if ! "$CLI_BIN" test > /dev/null 2>&1; then
        # Some templates might not have tests defined yet in their package.json or might need special setup
        # But we expect 'aerostack test' to at least not crash if it's a standard template.
        # Check if project has a test script
        if grep -q "\"test\":" package.json; then
            fail "Tests failed for $template"
            FAILED_TEMPLATES+=("$template")
            continue
        fi
        warn "No tests defined for $template, skipped verification"
    else
        pass "Tests passed"
    fi
done

echo ""
echo "=============================================="
if [ ${#FAILED_TEMPLATES[@]} -eq 0 ]; then
    echo -e "${GREEN}${BOLD}✅ ALL TEMPLATES PASSED E2E VERIFICATION${NC}"
else
    echo -e "${RED}${BOLD}✗ SOME TEMPLATES FAILED:${NC}"
    for failed in "${FAILED_TEMPLATES[@]}"; do
        echo -e "  - $failed"
    done
    exit 1
fi
echo "=============================================="
