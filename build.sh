#!/bin/bash

# Define output directory
OUTPUT_DIR="build"
mkdir -p $OUTPUT_DIR

echo "Building DataLane PDF Generator Service..."

# Build API Server
echo "Building API Server (datalane)..."
go build -o $OUTPUT_DIR/datalane.exe ./cmd/main.go
if [ $? -eq 0 ]; then
    echo "✅ API Server built successfully: $OUTPUT_DIR/datalane.exe"
else
    echo "❌ Failed to build API Server"
    exit 1
fi

# Build Worker Service
echo "Building PDF Worker (pdf-generator)..."
go build -o $OUTPUT_DIR/pdf-generator.exe ./cmd/pdf-generator/main.go
if [ $? -eq 0 ]; then
    echo "✅ PDF Worker built successfully: $OUTPUT_DIR/pdf-generator.exe"
else
    echo "❌ Failed to build PDF Worker"
    exit 1
fi

echo "🎉 Build completed!"
