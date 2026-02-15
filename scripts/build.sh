#!/bin/bash
set -e

echo "🔨 Building Aerostack CLI..."

# Build for current platform
go build -o bin/aerostack \
  -ldflags "-X main.version=dev -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  cmd/aerostack/main.go

echo "✅ Build complete: bin/aerostack"
echo ""
echo "To install globally:"
echo "  sudo mv bin/aerostack /usr/local/bin/"
