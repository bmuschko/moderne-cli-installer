#!/bin/bash

# Build script for cross-compiling the Moderne CLI installer
# Produces binaries for Windows, Linux, and macOS

set -e

VERSION="${1:-dev}"
OUTPUT_DIR="dist"
BINARY_NAME="moderne-cli-installer"

echo "Building Moderne CLI Installer (version: $VERSION)"
echo "=============================================="

# Clean and create output directory
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Build targets
declare -a TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for TARGET in "${TARGETS[@]}"; do
    OS=$(echo "$TARGET" | cut -d'/' -f1)
    ARCH=$(echo "$TARGET" | cut -d'/' -f2)

    OUTPUT_NAME="${BINARY_NAME}-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    echo "Building for $OS/$ARCH..."

    GOOS=$OS GOARCH=$ARCH go build \
        -ldflags="-s -w" \
        -o "${OUTPUT_DIR}/${OUTPUT_NAME}" \
        .

    echo "  -> ${OUTPUT_DIR}/${OUTPUT_NAME}"
done

echo ""
echo "Build complete! Binaries are in the '$OUTPUT_DIR' directory:"
ls -lh "$OUTPUT_DIR"
