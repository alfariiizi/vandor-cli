#!/bin/bash
# Vandor CLI Installation Script
# This script downloads and installs the latest version of Vandor CLI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="alfariiizi/vandor-cli"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="vandor"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    i386|i686) ARCH="386" ;;
esac

case $OS in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *) echo -e "${RED}Unsupported OS: $OS${NC}"; exit 1 ;;
esac

echo -e "${BLUE}🚀 Vandor CLI Installer${NC}"
echo -e "${BLUE}=========================${NC}"
echo ""

# Check if curl or wget is available
if command -v curl >/dev/null 2>&1; then
    DOWNLOADER="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
    DOWNLOADER="wget -qO-"
else
    echo -e "${RED}Error: curl or wget is required${NC}"
    exit 1
fi

# Get latest release info
echo -e "${YELLOW}🔍 Getting latest release information...${NC}"
LATEST_RELEASE=$($DOWNLOADER "https://api.github.com/repos/$REPO/releases/latest")

if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to get release information${NC}"
    exit 1
fi

# Extract version and download URL
VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
DOWNLOAD_URL=""

# Look for the appropriate asset
if echo "$LATEST_RELEASE" | grep -q "vandor-$OS-$ARCH"; then
    DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url.*vandor-$OS-$ARCH" | sed -E 's/.*"([^"]+)".*/\1/')
elif echo "$LATEST_RELEASE" | grep -q "vandor.*$OS.*$ARCH"; then
    DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url.*vandor.*$OS.*$ARCH" | sed -E 's/.*"([^"]+)".*/\1/' | head -1)
fi

if [ -z "$DOWNLOAD_URL" ]; then
    echo -e "${RED}No compatible binary found for $OS/$ARCH${NC}"
    echo -e "${YELLOW}Available assets:${NC}"
    echo "$LATEST_RELEASE" | grep "browser_download_url" | sed -E 's/.*"([^"]+)".*/\1/'
    exit 1
fi

echo -e "${GREEN}✅ Found Vandor CLI $VERSION for $OS/$ARCH${NC}"
echo -e "${YELLOW}📦 Download URL: $DOWNLOAD_URL${NC}"

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download
echo -e "${YELLOW}⬇️  Downloading...${NC}"
BINARY_FILE="$TMP_DIR/$BINARY_NAME"

if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$DOWNLOAD_URL" -o "$BINARY_FILE"
else
    wget -q "$DOWNLOAD_URL" -O "$BINARY_FILE"
fi

if [ $? -ne 0 ]; then
    echo -e "${RED}Download failed${NC}"
    exit 1
fi

# Extract if it's a tar.gz
if [[ "$DOWNLOAD_URL" == *.tar.gz ]]; then
    echo -e "${YELLOW}📂 Extracting archive...${NC}"
    cd "$TMP_DIR"
    tar -xzf "$BINARY_NAME"
    # Find the vandor binary in extracted files
    BINARY_FILE=$(find . -name "vandor" -o -name "vandor.exe" | head -1)
    if [ -z "$BINARY_FILE" ]; then
        echo -e "${RED}Binary not found in archive${NC}"
        exit 1
    fi
    BINARY_FILE="$TMP_DIR/$BINARY_FILE"
fi

# Make executable
chmod +x "$BINARY_FILE"

# Test the binary
echo -e "${YELLOW}🧪 Testing binary...${NC}"
if ! "$BINARY_FILE" version >/dev/null 2>&1; then
    echo -e "${RED}Downloaded binary is not working${NC}"
    exit 1
fi

# Install
echo -e "${YELLOW}📥 Installing to $INSTALL_DIR...${NC}"

# Check if we need sudo
if [ ! -w "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}🔐 Sudo access required for installation${NC}"
    sudo cp "$BINARY_FILE" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    cp "$BINARY_FILE" "$INSTALL_DIR/$BINARY_NAME"
fi

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ Vandor CLI $VERSION installed successfully!${NC}"
    echo ""
    echo -e "${BLUE}🎉 Quick Start:${NC}"
    echo -e "   ${YELLOW}vandor --help${NC}     # Show all commands"
    echo -e "   ${YELLOW}vandor init${NC}       # Initialize a new project"  
    echo -e "   ${YELLOW}vandor tui${NC}        # Launch interactive TUI"
    echo -e "   ${YELLOW}vandor upgrade${NC}    # Update to latest version"
    echo ""
    echo -e "${BLUE}📚 Documentation: https://github.com/$REPO${NC}"
    echo ""
else
    echo -e "${RED}Installation failed${NC}"
    exit 1
fi
