#!/bin/bash

# Usage: ./build.sh [os] [arch]
# Examples:
#   ./build.sh              # Default: windows amd64
#   ./build.sh linux amd64
#   ./build.sh darwin arm64

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

echo "Building DataLane PDF Generator Service..."
echo "Target: $TARGET_OS/$TARGET_ARCH"
echo "Output: $OUTPUT_DIR"
echo ""

# Set environment for cross-compilation
export GOOS=$TARGET_OS
export GOARCH=$TARGET_ARCH

# Build API Server
echo "Building API Server (datalane)..."
go build -o $OUTPUT_DIR/datalane$EXT ./cmd/main.go
if [ $? -eq 0 ]; then
    echo "✅ API Server built successfully: $OUTPUT_DIR/datalane$EXT"
else
    echo "❌ Failed to build API Server"
    exit 1
fi

# Build Worker Service
echo "Building PDF Worker (pdf-generator)..."
go build -o $OUTPUT_DIR/pdf-generator$EXT ./cmd/pdf-generator/main.go
if [ $? -eq 0 ]; then
    echo "✅ PDF Worker built successfully: $OUTPUT_DIR/pdf-generator$EXT"
else
    echo "❌ Failed to build PDF Worker"
    exit 1
fi

echo ""
echo "🎉 Build completed for $TARGET_OS/$TARGET_ARCH!"
