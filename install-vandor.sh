#!/bin/bash
set -e

# Vandor CLI Installer
echo "🚀 Vandor CLI Installer"
echo "========================="
echo

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case $OS in
    linux*)     OS="linux";;
    darwin*)    OS="darwin";;
    msys*|cygwin*|mingw*) OS="windows";;
    *)          echo "❌ Unsupported operating system: $OS"; exit 1;;
esac

# Detect Architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64|amd64)   ARCH="amd64";;
    aarch64|arm64)  ARCH="arm64";;
    armv7l)         ARCH="arm";;
    i386|i686)      ARCH="386";;
    *)              echo "❌ Unsupported architecture: $ARCH"; exit 1;;
esac

echo "🔍 Getting latest release information..."

# Get latest release info from GitHub API
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/alfariiizi/vandor-cli/releases/latest")

if [ $? -ne 0 ]; then
    echo "❌ Failed to fetch release information"
    exit 1
fi

VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name"' | sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "❌ Could not determine latest version"
    exit 1
fi

echo "✅ Found Vandor CLI $VERSION for $OS/$ARCH"

# Determine preferred file extensions and patterns
if [ "$OS" = "windows" ]; then
    PREFERRED_PATTERNS=(
        "vandor-$OS-$ARCH\.zip"
        "vandor-$OS-$ARCH\.exe"
    )
else
    PREFERRED_PATTERNS=(
        "vandor-$OS-$ARCH\.tar\.gz"
        "vandor-$OS-$ARCH"
    )
fi

# Find the best matching download URL
DOWNLOAD_URL=""
for pattern in "${PREFERRED_PATTERNS[@]}"; do
    DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep -o '"browser_download_url": *"[^"]*"' | grep -E "/$pattern\"" | sed 's/"browser_download_url": *"//;s/"//' | head -1)
    if [ -n "$DOWNLOAD_URL" ]; then
        echo "📦 Download URL: $DOWNLOAD_URL"
        break
    fi
done

if [ -z "$DOWNLOAD_URL" ]; then
    echo "❌ No compatible binary found for $OS/$ARCH"
    echo "Available assets:"
    echo "$LATEST_RELEASE" | grep '"name"' | sed 's/.*"name": *"\([^"]*\)".*/\1/' | sed 's/^/  /'
    exit 1
fi

# Determine installation directory
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

# Download and install
echo "⬇️  Downloading..."
TEMP_DIR=$(mktemp -d)
FILENAME=$(basename "$DOWNLOAD_URL")

if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_DIR/$FILENAME"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$DOWNLOAD_URL" -O "$TEMP_DIR/$FILENAME"
else
    echo "❌ Neither curl nor wget is available"
    exit 1
fi

echo "📦 Installing..."

# Extract and install based on file type
cd "$TEMP_DIR"
case "$FILENAME" in
    *.tar.gz)
        tar -xzf "$FILENAME"
        BINARY_NAME=$(tar -tzf "$FILENAME" | head -1)
        mv "$BINARY_NAME" "$INSTALL_DIR/vandor"
        ;;
    *.zip)
        unzip -q "$FILENAME"
        BINARY_NAME=$(unzip -l "$FILENAME" | tail -n +4 | head -1 | awk '{print $4}')
        mv "$BINARY_NAME" "$INSTALL_DIR/vandor"
        ;;
    *.exe)
        mv "$FILENAME" "$INSTALL_DIR/vandor.exe"
        ;;
    *)
        mv "$FILENAME" "$INSTALL_DIR/vandor"
        ;;
esac

# Make executable
chmod +x "$INSTALL_DIR/vandor"* 2>/dev/null || true

# Clean up
rm -rf "$TEMP_DIR"

echo "✅ Vandor CLI installed successfully!"
echo
echo "📍 Installation location: $INSTALL_DIR/vandor"

# Check if install directory is in PATH
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo "⚠️  Warning: $INSTALL_DIR is not in your PATH"
    echo "   Add the following line to your ~/.bashrc, ~/.zshrc, or ~/.profile:"
    echo "   export PATH=\"$INSTALL_DIR:\$PATH\""
    echo
fi

# Test installation
if command -v vandor >/dev/null 2>&1; then
    echo "🎉 Installation verified!"
    vandor --version
else
    echo "⚠️  Installation complete, but 'vandor' command not found in PATH"
    echo "   You may need to restart your shell or add $INSTALL_DIR to your PATH"
fi

echo
echo "🚀 Get started with: vandor --help"
echo "📖 Documentation: https://github.com/alfariiizi/vandor-cli"