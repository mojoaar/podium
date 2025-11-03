#!/bin/bash
# Podium Service Installation Script for Linux/macOS
# Run with sudo on Linux, regular user on macOS

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BIN_DIR="$SCRIPT_DIR/bin"
BINARY_NAME="podium"
BINARY_PATH="$BIN_DIR/$BINARY_NAME"

echo "=========================================="
echo "  Podium Service Installation"
echo "=========================================="
echo ""

# Check if binary exists in bin directory
if [ ! -f "$BINARY_PATH" ]; then
    echo "Binary not found in bin/. Building Podium binary..."
    cd "$SCRIPT_DIR"
    mkdir -p "$BIN_DIR"
    go build -o "$BINARY_PATH"
    echo "✓ Binary built successfully"
    echo ""
fi

# Make binary executable
chmod +x "$BINARY_PATH"

echo "Installing Podium as a system service..."
cd "$SCRIPT_DIR"
"$BINARY_PATH" -service install

echo ""
echo "✓ Service installed successfully!"
echo ""
echo "To start the service:"
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "  $BINARY_PATH -service start"
    echo ""
    echo "Or use launchctl:"
    echo "  launchctl load ~/Library/LaunchAgents/Podium.plist"
else
    echo "  sudo $BINARY_PATH -service start"
    echo ""
    echo "Or use systemctl:"
    echo "  sudo systemctl start Podium"
    echo "  sudo systemctl enable Podium  # Enable on boot"
fi
echo ""
echo "To access your website, visit: http://localhost:8080"
echo ""
