#!/bin/bash

# Usage: ./build.sh [os] [arch]
# Examples:
#   ./build.sh              # Default: windows amd64
#   ./build.sh linux amd64
#   ./build.sh darwin arm64

set -e

# Parse arguments with defaults
TARGET_OS="${1:-windows}"
TARGET_ARCH="${2:-amd64}"

# Determine file extension
EXT=""
if [ "$TARGET_OS" = "windows" ]; then
    EXT=".exe"
fi

# Define output directory
OUTPUT_DIR="build/${TARGET_OS}_${TARGET_ARCH}"
mkdir -p $OUTPUT_DIR

echo "========================================"
echo "  DataLane Build Script"
echo "========================================"
echo "Target: $TARGET_OS/$TARGET_ARCH"
echo "Output: $OUTPUT_DIR"
echo ""

# ========================================
# Step 1: Check prerequisites
# ========================================
echo "📋 Checking prerequisites..."

if ! command -v bun &> /dev/null; then
    echo "❌ bun is not installed. Please install bun first."
    echo "   Visit: https://bun.sh/"
    exit 1
fi
echo "✅ bun is available"

if ! command -v go &> /dev/null; then
    echo "❌ go is not installed. Please install Go first."
    exit 1
fi
echo "✅ go is available"
echo ""

# ========================================
# Step 2: Build UI (Next.js Static Export)
# ========================================
echo "🎨 Building UI (Next.js)..."

cd ui

echo "   Installing dependencies..."
bun install --frozen-lockfile

echo "   Building static export..."
bun run build

cd ..

echo "✅ UI built successfully"
echo ""

# ========================================
# Step 3: Copy UI assets to embed directory
# ========================================
echo "📦 Copying UI assets to embed directory..."

# Clean existing dist
rm -rf internal/assets/dist/*

# Copy new dist
cp -r ui/dist/* internal/assets/dist/

echo "✅ UI assets copied to internal/assets/dist/"
echo ""

# ========================================
# Step 4: Build Go binaries
# ========================================
echo "🔨 Building Go binaries..."

# Set environment for cross-compilation
export GOOS=$TARGET_OS
export GOARCH=$TARGET_ARCH

# Build API Server
echo "   Building API Server (datalane)..."
go build -o $OUTPUT_DIR/datalane$EXT ./cmd/main.go
if [ $? -eq 0 ]; then
    echo "   ✅ datalane$EXT"
else
    echo "   ❌ Failed to build datalane"
    exit 1
fi

# Build Worker Service
echo "   Building PDF Worker (datalane_gen_pdf)..."
go build -o $OUTPUT_DIR/datalane_gen_pdf$EXT ./cmd/pdf-generator/main.go
if [ $? -eq 0 ]; then
    echo "   ✅ datalane_gen_pdf$EXT"
else
    echo "   ❌ Failed to build datalane_gen_pdf"
    exit 1
fi

echo ""
echo "========================================"
echo "🎉 Build completed for $TARGET_OS/$TARGET_ARCH!"
echo "========================================"
echo ""
echo "Output files:"
echo "  - $OUTPUT_DIR/datalane$EXT"
echo "  - $OUTPUT_DIR/datalane_gen_pdf$EXT"
