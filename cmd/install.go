package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Generate installation scripts and instructions",
	Long:  `Generate installation scripts for different platforms and provide installation instructions.`,
}

var installScriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Generate installation script",
	Long:  `Generate a shell script for installing Vandor CLI from GitHub releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := generateInstallScript(); err != nil {
			er(fmt.Sprintf("Failed to generate install script: %v", err))
		}
	},
}

var installInstructionsCmd = &cobra.Command{
	Use:   "instructions",
	Short: "Show installation instructions",
	Long:  `Display detailed installation instructions for different platforms.`,
	Run: func(cmd *cobra.Command, args []string) {
		showInstallationInstructions()
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Vandor CLI from the system",
	Long:  `Remove Vandor CLI from the system including the binary and any configuration files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := performUninstall(); err != nil {
			er(fmt.Sprintf("Failed to uninstall: %v", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(installScriptCmd)
	installCmd.AddCommand(installInstructionsCmd)
	rootCmd.AddCommand(uninstallCmd)
}

func generateInstallScript() error {
	script := `#!/bin/bash
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
`

	filename := "install-vandor.sh"
	if err := os.WriteFile(filename, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to write script: %v", err)
	}

	fmt.Printf("✅ Installation script generated: %s\n", filename)
	fmt.Println()
	fmt.Println("📋 Next Steps:")
	fmt.Println("1. Host this script on GitHub or your web server")
	fmt.Println("2. Commit it to your repository as 'install-vandor.sh'")
	fmt.Println("3. Users can then install with:")
	fmt.Println("   curl -fsSL https://raw.githubusercontent.com/vandordev/vandor-cli/main/install-vandor.sh | bash")
	fmt.Println()
	fmt.Println("🧪 Local testing:")
	fmt.Printf("   chmod +x %s && ./%s\n", filename, filename)
	fmt.Println()
	fmt.Println("💡 Pro tip: Create a short URL like https://install.vandor.dev for easier sharing!")

	return nil
}

func showInstallationInstructions() {
	fmt.Println("🚀 Vandor CLI Installation Instructions")
	fmt.Println("=======================================")
	fmt.Println()

	// Auto-install methods
	fmt.Println("📦 Quick Install (Recommended)")
	fmt.Println("------------------------------")
	fmt.Println("# Download and install latest version automatically:")
	fmt.Println("curl -fsSL https://raw.githubusercontent.com/vandordev/vandor-cli/main/install-vandor.sh | bash")
	fmt.Println()
	fmt.Println("# Or with wget:")
	fmt.Println("wget -qO- https://raw.githubusercontent.com/vandordev/vandor-cli/main/install-vandor.sh | bash")
	fmt.Println()
	fmt.Println("# Alternative one-liner (if hosted on your domain):")
	fmt.Println("curl -fsSL https://install.vandor.dev | bash")
	fmt.Println()

	// Manual download
	fmt.Println("💾 Manual Download")
	fmt.Println("------------------")
	fmt.Println("1. Visit: https://github.com/alfariiizi/vandor-cli-cli/releases/latest")
	fmt.Printf("2. Download the binary for your platform (%s/%s)\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println("3. Extract and move to PATH:")
	fmt.Println()
	
	if runtime.GOOS == "windows" {
		fmt.Println("   # Windows (PowerShell)")
		fmt.Println("   Expand-Archive vandor-windows-amd64.zip")
		fmt.Println("   Move-Item vandor.exe C:\\Windows\\System32\\")
	} else {
		fmt.Println("   # Linux/macOS")
		fmt.Println("   tar -xzf vandor-linux-amd64.tar.gz")
		fmt.Println("   sudo mv vandor /usr/local/bin/")
		fmt.Println("   chmod +x /usr/local/bin/vandor")
	}
	fmt.Println()

	// From source
	fmt.Println("🔧 Build from Source")
	fmt.Println("--------------------")
	fmt.Println("# Requires Go 1.21+")
	fmt.Println("git clone https://github.com/alfariiizi/vandor-cli-cli.git")
	fmt.Println("cd vandor-cli")
	fmt.Println("go build -o vandor main.go")
	fmt.Println("sudo mv vandor /usr/local/bin/")
	fmt.Println()

	// Package managers (future)
	fmt.Println("📋 Package Managers (Coming Soon)")
	fmt.Println("----------------------------------")
	fmt.Println("# Homebrew (macOS/Linux)")
	fmt.Println("brew install vandor-cli")
	fmt.Println()
	fmt.Println("# Snap (Linux)")
	fmt.Println("sudo snap install vandor")
	fmt.Println()
	fmt.Println("# Chocolatey (Windows)")
	fmt.Println("choco install vandor-cli")
	fmt.Println()

	// Verification
	fmt.Println("✅ Verify Installation")
	fmt.Println("----------------------")
	fmt.Println("vandor version        # Check installed version")
	fmt.Println("vandor --help         # Show all commands")
	fmt.Println("vandor upgrade check  # Check for updates")
	fmt.Println()

	// Quick start
	fmt.Println("🎯 Quick Start")
	fmt.Println("--------------")
	fmt.Println("vandor init           # Initialize new project")
	fmt.Println("vandor tui            # Launch interactive UI")
	fmt.Println("vandor theme set mocha # Set beautiful theme")
	fmt.Println()

	// Support
	fmt.Println("🆘 Support")
	fmt.Println("----------")
	fmt.Println("GitHub: https://github.com/alfariiizi/vandor-cli-cli/issues")
	fmt.Println("Docs:   https://docs.vandor.dev")
}

func performUninstall() error {
	fmt.Println("🗑️  Vandor CLI Uninstall")
	fmt.Println("=======================")
	fmt.Println()

	// Get the current executable path
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Resolve symlinks
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %v", err)
	}

	fmt.Printf("📍 Found Vandor CLI at: %s\n", exe)
	fmt.Println()

	// Check if this is a system installation
	systemPaths := []string{
		"/usr/local/bin/vandor",
		"/usr/bin/vandor",
		"/opt/vandor/vandor",
	}

	needsSudo := false
	for _, path := range systemPaths {
		if exe == path {
			needsSudo = true
			break
		}
	}

	// Show what will be removed
	fmt.Println("🔍 The following will be removed:")
	fmt.Printf("   • Binary: %s\n", exe)
	
	// Check for backup files
	backupPath := exe + ".backup"
	if _, err := os.Stat(backupPath); err == nil {
		fmt.Printf("   • Backup: %s\n", backupPath)
	}

	fmt.Println()

	// Confirmation prompt
	fmt.Print("❓ Are you sure you want to uninstall Vandor CLI? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("❌ Uninstall cancelled.")
		return nil
	}

	fmt.Println()
	fmt.Println("🗑️  Removing Vandor CLI...")

	// Remove the binary
	if needsSudo {
		fmt.Println("🔐 Sudo access required for system-wide removal")
		// For system paths, we need to instruct the user to use sudo
		fmt.Printf("Please run: sudo rm -f %s\n", exe)
		if _, err := os.Stat(backupPath); err == nil {
			fmt.Printf("Please run: sudo rm -f %s\n", backupPath)
		}
		fmt.Println()
		fmt.Println("⚠️  Manual removal required for system installation.")
		fmt.Println("    Run the commands above to complete the uninstallation.")
		return nil
	} else {
		// Remove binary
		if err := os.Remove(exe); err != nil {
			return fmt.Errorf("failed to remove binary: %v", err)
		}

		// Remove backup if it exists
		if _, err := os.Stat(backupPath); err == nil {
			os.Remove(backupPath)
		}
	}

	fmt.Println("✅ Vandor CLI has been successfully uninstalled!")
	fmt.Println()
	fmt.Println("📋 What was removed:")
	fmt.Printf("   • %s\n", exe)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		fmt.Printf("   • %s (if existed)\n", backupPath)
	}
	fmt.Println()
	fmt.Println("🙏 Thank you for using Vandor CLI!")
	fmt.Println("   If you need to reinstall, visit: https://github.com/alfariiizi/vandor-cli-cli")

	return nil
}