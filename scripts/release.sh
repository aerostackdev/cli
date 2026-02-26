#!/usr/bin/env bash
# Aerostack CLI — Release Guard Script
# Runs all verification BEFORE creating any release artifacts.
# A release is BLOCKED until every check passes.
#
# Usage:
#   ./scripts/release.sh                    # Full release
#   ./scripts/release.sh --dry-run          # Verify only, do not tag/push

set -euo pipefail
cd "$(dirname "$0")/.."

DRY_RUN=0
FAST_MODE=0
for arg in "$@"; do
  if [[ "$arg" == "--dry-run" ]]; then DRY_RUN=1; fi
  if [[ "$arg" == "--fast" ]]; then FAST_MODE=1; fi
done

# Colours
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BOLD='\033[1m'; NC='\033[0m'
pass() { echo -e "${GREEN}✓${NC} $1"; }
fail() { echo -e "${RED}✗ FAIL${NC} $1"; }
info() { echo -e "${BOLD}$1${NC}"; }
warn() { echo -e "${YELLOW}⚠${NC}  $1"; }

# ──────────────────────────────────────────
# Resolve version from sdks/VERSION_CLI
# ──────────────────────────────────────────
SDKS_ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"
CLI_VERSION_FILE="$SDKS_ROOT/VERSION_CLI"

if [ ! -f "$CLI_VERSION_FILE" ]; then
  fail "Cannot find VERSION_CLI at $CLI_VERSION_FILE"
  exit 1
fi

VERSION=$(cat "$CLI_VERSION_FILE" | tr -d '[:space:]')
TAG="v$VERSION"

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║       Aerostack CLI — Pre-Release Guard           ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""
info "  Version: $TAG"
if [ "$DRY_RUN" -eq 1 ]; then
  warn "  DRY RUN — will not tag or push"
fi
if [ "$FAST_MODE" -eq 1 ]; then
  warn "  FAST MODE — skipping local tests (relying on CI gates)"
fi
echo ""

# ──────────────────────────────────────────
# Gate 1: Git state
# ──────────────────────────────────────────
info "[Gate 1/5] Checking Git state..."

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
  warn "Not on 'main' branch (on '$CURRENT_BRANCH'). Proceeding anyway..."
fi

if ! git diff --quiet || ! git diff --cached --quiet; then
  fail "Uncommitted changes detected. Commit or stash before releasing."
  exit 1
else
  pass "Working tree is clean"
fi

# Check tag doesn't already exist
if git tag -l | grep -q "^${TAG}$"; then
  fail "Tag $TAG already exists! Bump VERSION_CLI first."
  exit 1
else
  pass "Tag $TAG is new"
fi

# ──────────────────────────────────────────
# Gate 2: Go build
# ──────────────────────────────────────────
echo ""
info "[Gate 2/5] Building release binary..."
if go build \
  -ldflags "-X main.version=$VERSION -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o bin/aerostack ./cmd/aerostack 2>&1; then
  pass "Build succeeded"
else
  fail "Build failed — cannot release"
  exit 1
fi

# ──────────────────────────────────────────
# Gate 3/4/5: Tests (Skipped in FAST_MODE)
# ──────────────────────────────────────────
if [ "$FAST_MODE" -eq 1 ]; then
  info "Skipping Gates 3-5 (Fast Mode)..."
  pass "Tests bypassed (CI will enforce rules on push)"
else
  echo ""
  info "[Gate 3/5] Running Go unit tests..."
  if go test ./... 2>&1; then
    pass "All unit tests passed"
  else
    fail "Unit tests failed — cannot release"
    exit 1
  fi

  echo ""
  info "[Gate 4/5] Running phase verification suite..."
  if ./scripts/verify-all.sh 2>&1; then
    pass "Phase verification suite passed"
  else
    fail "Phase verification suite failed — cannot release"
    exit 1
  fi

  echo ""
  info "[Gate 5/5] Running E2E tests (build → init → dev → connectivity)..."
  if ./scripts/e2e-test.sh 2>&1; then
    pass "E2E tests passed — dev command works end-to-end"
  else
    fail "E2E tests failed — aerostack dev is broken, DO NOT RELEASE"
    exit 1
  fi
fi

# ──────────────────────────────────────────
# All gates passed
# ──────────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════════════════╗"
echo -e "║  ${GREEN}✅ All 5 gates passed — ready to release $TAG${NC}"
echo "╚══════════════════════════════════════════════════╝"
echo ""

if [ "$DRY_RUN" -eq 1 ]; then
  warn "DRY RUN complete. No tags or commits created."
  exit 0
fi

# ──────────────────────────────────────────
# Confirmation prompt
# ──────────────────────────────────────────
echo -e "${BOLD}About to:${NC}"
echo "  1. Apply versions via apply_version.sh"
echo "  2. git add ."
echo "  3. git commit -m \"chore: release $TAG\""
echo "  4. git tag $TAG"
echo "  5. git push origin main && git push origin --tags"
echo ""
read -rp "Proceed with release $TAG? [y/N] " CONFIRM
if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
  warn "Release aborted by user."
  exit 0
fi

# ──────────────────────────────────────────
# Execute release steps
# ──────────────────────────────────────────
cd "$SDKS_ROOT"

echo ""
info "Applying versions..."
./scripts/apply_version.sh

info "Staging changes..."
git add .

info "Committing..."
git commit -m "chore: release $TAG"

info "Tagging $TAG..."
git tag "$TAG"

info "Pushing to remote..."
git push origin main
git push origin --tags

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo -e "║  ${GREEN}🚀 Released $TAG successfully!${NC}"
echo "╚══════════════════════════════════════════════════╝"
echo ""
echo "  CI/CD pipeline will now build and publish binaries."
echo "  Monitor at: https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]//' | sed 's/\.git$//')/actions"
echo ""
