#!/bin/bash

# Vandor CLI Installation Script
# This script moves the vandor binary to ~/bin and updates PATH if needed

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Installing Vandor CLI...${NC}"

# Check if vandor binary exists
if [ ! -f "./vandor" ]; then
    echo -e "${RED}❌ Error: vandor binary not found in current directory${NC}"
    echo "Please run 'go build -o vandor main.go' first"
    exit 1
fi

# Create ~/bin directory if it doesn't exist
BIN_DIR="$HOME/bin"
if [ ! -d "$BIN_DIR" ]; then
    echo -e "${YELLOW}📁 Creating $BIN_DIR directory...${NC}"
    mkdir -p "$BIN_DIR"
fi

# Move vandor binary to ~/bin
echo -e "${YELLOW}📦 Moving vandor binary to $BIN_DIR...${NC}"
cp ./vandor "$BIN_DIR/vandor"
chmod +x "$BIN_DIR/vandor"

# Check if ~/bin is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo -e "${YELLOW}⚠️  $BIN_DIR is not in your PATH${NC}"
    echo -e "${YELLOW}📝 Adding $BIN_DIR to PATH in your shell profile...${NC}"
    
    # Detect shell and add to appropriate profile
    SHELL_NAME=$(basename "$SHELL")
    case "$SHELL_NAME" in
        bash)
            PROFILE_FILE="$HOME/.bashrc"
            if [ ! -f "$PROFILE_FILE" ]; then
                PROFILE_FILE="$HOME/.bash_profile"
            fi
            ;;
        zsh)
            PROFILE_FILE="$HOME/.zshrc"
            ;;
        fish)
            PROFILE_FILE="$HOME/.config/fish/config.fish"
            ;;
        *)
            PROFILE_FILE="$HOME/.profile"
            ;;
    esac
    
    # Add PATH export to profile if not already there
    if [ -f "$PROFILE_FILE" ]; then
        if ! grep -q "export PATH.*$BIN_DIR" "$PROFILE_FILE"; then
            echo "" >> "$PROFILE_FILE"
            echo "# Added by Vandor CLI installer" >> "$PROFILE_FILE"
            echo "export PATH=\"$BIN_DIR:\$PATH\"" >> "$PROFILE_FILE"
            echo -e "${GREEN}✅ Added $BIN_DIR to PATH in $PROFILE_FILE${NC}"
        else
            echo -e "${GREEN}✅ $BIN_DIR already in PATH configuration${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Could not find shell profile file${NC}"
        echo -e "${YELLOW}Please manually add the following to your shell profile:${NC}"
        echo "export PATH=\"$BIN_DIR:\$PATH\""
    fi
    
    echo -e "${YELLOW}🔄 Please restart your terminal or run: source $PROFILE_FILE${NC}"
else
    echo -e "${GREEN}✅ $BIN_DIR is already in your PATH${NC}"
fi

# Verify installation
if [ -x "$BIN_DIR/vandor" ]; then
    echo -e "${GREEN}✅ Vandor CLI installed successfully!${NC}"
    echo ""
    echo -e "${GREEN}🎉 You can now use Vandor CLI from anywhere:${NC}"
    echo "  vandor --help"
    echo "  vandor add domain User"
    echo "  vandor sync all"
    echo "  vandor task"
    echo ""
    
    # Show version if possible
    if [[ ":$PATH:" == *":$BIN_DIR:"* ]]; then
        echo -e "${GREEN}📋 Installed version:${NC}"
        "$BIN_DIR/vandor" --version 2>/dev/null || echo "Vandor CLI (latest)"
    else
        echo -e "${YELLOW}⚠️  Restart your terminal to use 'vandor' command globally${NC}"
        echo -e "${YELLOW}Or run: $BIN_DIR/vandor --help${NC}"
    fi
else
    echo -e "${RED}❌ Installation failed${NC}"
    exit 1
fi