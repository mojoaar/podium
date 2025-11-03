#!/bin/bash
# Podium Service Uninstallation Script for Linux/macOS
# Run with sudo on Linux, regular user on macOS

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BIN_DIR="$SCRIPT_DIR/bin"
BINARY_NAME="podium"
BINARY_PATH="$BIN_DIR/$BINARY_NAME"

echo "=========================================="
echo "  Podium Service Uninstallation"
echo "=========================================="
echo ""

if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: $BINARY_NAME binary not found in bin/ directory!"
    exit 1
fi

echo "Uninstalling Podium service..."
cd "$SCRIPT_DIR"
"$BINARY_PATH" -service uninstall

echo ""
echo "✓ Service uninstalled successfully!"
echo ""
