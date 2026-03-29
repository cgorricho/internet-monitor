#!/bin/bash

# Internet Monitor - Cross-platform Build Script
# Builds binaries for Linux, Windows, and macOS

set -e

# Configuration
APP_NAME="internet-monitor"
VERSION=${VERSION:-"dev"}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE} -s -w"

# Output directory
DIST_DIR="dist"
rm -rf ${DIST_DIR}
mkdir -p ${DIST_DIR}

echo "🚀 Building Internet Monitor v${VERSION}"
echo "📦 Commit: ${COMMIT}"
echo "📅 Date: ${DATE}"
echo ""

# Function to build for a platform
build_platform() {
    local goos=$1
    local goarch=$2
    local ext=$3
    local output_name="${APP_NAME}-${goos}-${goarch}${ext}"
    
    echo "🔨 Building for ${goos}/${goarch}..."
    
    CGO_ENABLED=1 \
    GOOS=${goos} \
    GOARCH=${goarch} \
    go build \
        -ldflags="${LDFLAGS}" \
        -o "${DIST_DIR}/${output_name}" \
        ./cmd/internet-monitor
    
    echo "✅ Built: ${DIST_DIR}/${output_name}"
}

# Build for Linux
echo "🐧 Building Linux binaries..."
build_platform "linux" "amd64" ""
build_platform "linux" "arm64" ""

# Build for Windows
echo "🪟 Building Windows binaries..."
build_platform "windows" "amd64" ".exe"
build_platform "windows" "arm64" ".exe"

# Build for macOS
echo "🍎 Building macOS binaries..."
build_platform "darwin" "amd64" ""
build_platform "darwin" "arm64" ""

echo ""
echo "🎉 Build completed successfully!"
echo ""
echo "📦 Built binaries:"
ls -la ${DIST_DIR}/
echo ""

# Calculate file sizes
echo "📊 File sizes:"
for file in ${DIST_DIR}/*; do
    size=$(du -h "$file" | cut -f1)
    filename=$(basename "$file")
    echo "  $filename: $size"
done
echo ""

# Create checksums
echo "🔐 Generating checksums..."
cd ${DIST_DIR}
sha256sum * > checksums.txt
echo "✅ Checksums saved to ${DIST_DIR}/checksums.txt"
cd ..

echo ""
echo "🚀 Ready for distribution!"
echo "Run './dist/internet-monitor-linux-amd64 --help' to get started"