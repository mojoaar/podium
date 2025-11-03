#!/bin/bash
# Podium Build Script for Linux/macOS
# Builds binaries for all supported platforms

set -e

VERSION=${VERSION:-"1.0.0"}
BUILD_DIR="bin"
APP_NAME="podium"

echo "=========================================="
echo "  Building Podium v${VERSION}"
echo "=========================================="
echo ""

# Create bin directory if it doesn't exist
mkdir -p "$BUILD_DIR"

# Get the current git commit hash (if available)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS="-X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME}"

# Function to build for a specific platform
build_platform() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT=$3
    
    echo "Building for ${GOOS}/${GOARCH}..."
    
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi
    
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "$LDFLAGS" -o "${BUILD_DIR}/${OUTPUT}" .
    
    if [ $? -eq 0 ]; then
        echo "✓ Successfully built: ${BUILD_DIR}/${OUTPUT}"
        ls -lh "${BUILD_DIR}/${OUTPUT}"
    else
        echo "✗ Failed to build for ${GOOS}/${GOARCH}"
        return 1
    fi
    echo ""
}

# Parse command line arguments
BUILD_ALL=false
BUILD_CURRENT=false
BUILD_LINUX=false
BUILD_MACOS=false
BUILD_WINDOWS=false

if [ $# -eq 0 ]; then
    BUILD_CURRENT=true
else
    for arg in "$@"; do
        case $arg in
            all)
                BUILD_ALL=true
                ;;
            current)
                BUILD_CURRENT=true
                ;;
            linux)
                BUILD_LINUX=true
                ;;
            macos|darwin)
                BUILD_MACOS=true
                ;;
            windows)
                BUILD_WINDOWS=true
                ;;
            clean)
                echo "Cleaning build directory..."
                rm -rf "$BUILD_DIR"
                echo "✓ Cleaned"
                exit 0
                ;;
            *)
                echo "Unknown option: $arg"
                echo ""
                echo "Usage: $0 [all|current|linux|macos|windows|clean]"
                echo ""
                echo "Options:"
                echo "  all       - Build for all platforms"
                echo "  current   - Build for current platform only (default)"
                echo "  linux     - Build for Linux (amd64)"
                echo "  macos     - Build for macOS (amd64 and arm64)"
                echo "  windows   - Build for Windows (amd64)"
                echo "  clean     - Remove all built binaries"
                echo ""
                echo "Examples:"
                echo "  $0              # Build for current platform"
                echo "  $0 all          # Build for all platforms"
                echo "  $0 linux macos  # Build for Linux and macOS"
                exit 1
                ;;
        esac
    done
fi

# Build based on options
if [ "$BUILD_ALL" = true ]; then
    build_platform "linux" "amd64" "${APP_NAME}-linux-amd64"
    build_platform "linux" "arm64" "${APP_NAME}-linux-arm64"
    build_platform "darwin" "amd64" "${APP_NAME}-darwin-amd64"
    build_platform "darwin" "arm64" "${APP_NAME}-darwin-arm64"
    build_platform "windows" "amd64" "${APP_NAME}-windows-amd64"
elif [ "$BUILD_CURRENT" = true ]; then
    # Detect current OS
    CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    CURRENT_ARCH=$(uname -m)
    
    # Normalize architecture
    case $CURRENT_ARCH in
        x86_64)
            CURRENT_ARCH="amd64"
            ;;
        aarch64|arm64)
            CURRENT_ARCH="arm64"
            ;;
    esac
    
    # Normalize OS name
    case $CURRENT_OS in
        darwin)
            CURRENT_OS="darwin"
            ;;
        linux)
            CURRENT_OS="linux"
            ;;
        mingw*|msys*|cygwin*)
            CURRENT_OS="windows"
            ;;
    esac
    
    build_platform "$CURRENT_OS" "$CURRENT_ARCH" "${APP_NAME}"
else
    if [ "$BUILD_LINUX" = true ]; then
        build_platform "linux" "amd64" "${APP_NAME}-linux-amd64"
        build_platform "linux" "arm64" "${APP_NAME}-linux-arm64"
    fi
    
    if [ "$BUILD_MACOS" = true ]; then
        build_platform "darwin" "amd64" "${APP_NAME}-darwin-amd64"
        build_platform "darwin" "arm64" "${APP_NAME}-darwin-arm64"
    fi
    
    if [ "$BUILD_WINDOWS" = true ]; then
        build_platform "windows" "amd64" "${APP_NAME}-windows-amd64"
    fi
fi

echo "=========================================="
echo "  Build Complete!"
echo "=========================================="
echo ""
echo "Binaries are in the '${BUILD_DIR}/' directory"
ls -lh "$BUILD_DIR"
echo ""
