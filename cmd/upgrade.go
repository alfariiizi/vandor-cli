package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	GitHubAPIURL = "https://api.github.com/repos/alfariiizi/vandor-cli/releases/latest"
)

func getCurrentVersion() string {
	version, _, _ := getVersionInfo()
	return version
}

type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	Assets      []GitHubAsset `json:"assets"`
	CreatedAt   string        `json:"created_at"`
	PublishedAt string        `json:"published_at"`
}

type GitHubAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Vandor CLI to the latest version",
	Long: `Check for and install the latest version of Vandor CLI.
This command will:
- Check GitHub releases for the latest version
- Download the appropriate binary for your system
- Replace the current binary with the new version
- Verify the installation was successful`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := performUpgrade(); err != nil {
			er(fmt.Sprintf("Upgrade failed: %v", err))
		}
	},
}

var upgradeCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if a newer version is available",
	Long:  `Check GitHub releases to see if a newer version of Vandor CLI is available.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, _, err := checkForUpdates(true); err != nil {
			er(fmt.Sprintf("Failed to check for updates: %v", err))
		}
	},
}

var upgradeSourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Upgrade by building from source code",
	Long: `Build and install the latest version from GitHub source code.
This is useful when no releases are available yet.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := performSourceUpgrade(); err != nil {
			er(fmt.Sprintf("Source upgrade failed: %v", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.AddCommand(upgradeCheckCmd)
	upgradeCmd.AddCommand(upgradeSourceCmd)
}

func performUpgrade() error {
	fmt.Println("🔍 Checking for the latest version...")

	release, hasUpdate, err := checkForUpdates(false)
	if err != nil {
		// If releases are not available, offer source upgrade as fallback
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "releases") {
			fmt.Println("⚠️  No GitHub releases found.")
			fmt.Println("💡 Would you like to upgrade from source code instead? (y/N)")
			var response string
			if _, scanErr := fmt.Scanln(&response); scanErr != nil {
				// Handle scan error gracefully
				response = "n"
			}
			if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
				return performSourceUpgrade()
			}
			return fmt.Errorf("upgrade canceled - no releases available")
		}
		return fmt.Errorf("failed to check for updates: %v", err)
	}

	if !hasUpdate {
		fmt.Printf("✅ You're already running the latest version (%s)!\n", getCurrentVersion())
		return nil
	}

	fmt.Printf("📦 New version available: %s -> %s\n", getCurrentVersion(), release.TagName)
	fmt.Printf("📅 Released: %s\n", formatDate(release.PublishedAt))

	if release.Body != "" {
		fmt.Println("\n📋 Release Notes:")
		fmt.Println(release.Body)
		fmt.Println()
	}

	// Find the appropriate asset for the current platform
	asset, err := findAssetForPlatform(release.Assets)
	if err != nil {
		return fmt.Errorf("no compatible binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("⬇️  Downloading %s (%s)...\n", asset.Name, formatSize(asset.Size))

	// Download the new binary
	tempFile, err := downloadAsset(asset)
	if err != nil {
		return fmt.Errorf("failed to download: %v", err)
	}
	defer func() {
		if rmErr := os.Remove(tempFile); rmErr != nil {
			fmt.Printf("Warning: failed to remove temp file: %v\n", rmErr)
		}
	}()

	// Extract if it's a compressed archive
	binaryPath := tempFile
	if strings.HasSuffix(asset.Name, ".tar.gz") {
		extractedPath, err := extractTarGz(tempFile)
		if err != nil {
			return fmt.Errorf("failed to extract archive: %v", err)
		}
		binaryPath = extractedPath
		defer func() {
			if rmErr := os.Remove(extractedPath); rmErr != nil {
				fmt.Printf("Warning: failed to remove extracted file: %v\n", rmErr)
			}
		}()
	}

	// Replace the current binary
	if err := replaceCurrentBinary(binaryPath); err != nil {
		return fmt.Errorf("failed to replace binary: %v", err)
	}

	fmt.Printf("✅ Successfully upgraded to Vandor CLI %s!\n", release.TagName)
	fmt.Println("🎉 Run 'vandor version' to verify the installation.")

	return nil
}

func checkForUpdates(verbose bool) (*GitHubRelease, bool, error) {
	if verbose {
		fmt.Printf("🔍 Checking for updates (current version: %s)...\n", getCurrentVersion())
	}

	resp, err := http.Get(GitHubAPIURL)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch release info: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, false, fmt.Errorf("failed to parse release info: %v", err)
	}

	// Clean version strings for comparison
	currentVer := strings.TrimPrefix(getCurrentVersion(), "v")
	latestVer := strings.TrimPrefix(release.TagName, "v")

	hasUpdate := currentVer != latestVer

	if verbose {
		if hasUpdate {
			fmt.Printf("🆕 New version available: %s -> %s\n", getCurrentVersion(), release.TagName)
			fmt.Printf("📅 Released: %s\n", formatDate(release.PublishedAt))
			fmt.Println("💡 Run 'vandor upgrade' to install the latest version.")
		} else {
			fmt.Printf("✅ You're running the latest version (%s)!\n", getCurrentVersion())
		}
	}

	return &release, hasUpdate, nil
}

func findAssetForPlatform(assets []GitHubAsset) (*GitHubAsset, error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Map Go architecture names to common naming conventions
	archMap := map[string][]string{
		"amd64": {"amd64", "x86_64", "64bit"},
		"386":   {"386", "i386", "32bit"},
		"arm64": {"arm64", "aarch64"},
		"arm":   {"arm", "armv7"},
	}

	// Map Go OS names to common naming conventions
	osMap := map[string][]string{
		"linux":   {"linux"},
		"darwin":  {"darwin", "macos", "osx"},
		"windows": {"windows", "win"},
	}

	archVariants := archMap[archName]
	osVariants := osMap[osName]

	// Look for exact matches first
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)

		// Check if this asset matches our platform
		osMatch := false
		archMatch := false

		for _, osVar := range osVariants {
			if strings.Contains(name, osVar) {
				osMatch = true
				break
			}
		}

		for _, archVar := range archVariants {
			if strings.Contains(name, archVar) {
				archMatch = true
				break
			}
		}

		if osMatch && archMatch {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("no binary found for %s/%s", osName, archName)
}

func downloadAsset(asset *GitHubAsset) (string, error) {
	resp, err := http.Get(asset.DownloadURL)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "vandor-upgrade-*")
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := tempFile.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close temp file: %v\n", closeErr)
		}
	}()

	// Copy with progress indication
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		if rmErr := os.Remove(tempFile.Name()); rmErr != nil {
			fmt.Printf("Warning: failed to remove temp file: %v\n", rmErr)
		}
		return "", err
	}

	return tempFile.Name(), nil
}

func extractTarGz(srcPath string) (string, error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close source file: %v\n", closeErr)
		}
	}()

	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := gzReader.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close gzip reader: %v\n", closeErr)
		}
	}()

	tarReader := tar.NewReader(gzReader)

	// Look for the vandor binary in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Typeflag == tar.TypeReg && (strings.HasSuffix(header.Name, "vandor") || strings.HasSuffix(header.Name, "vandor.exe")) {
			// Extract this file
			tempFile, err := os.CreateTemp("", "vandor-binary-*")
			if err != nil {
				return "", err
			}
			defer func() {
				if closeErr := tempFile.Close(); closeErr != nil {
					fmt.Printf("Warning: failed to close temp file: %v\n", closeErr)
				}
			}()

			_, err = io.Copy(tempFile, tarReader)
			if err != nil {
				if rmErr := os.Remove(tempFile.Name()); rmErr != nil {
					fmt.Printf("Warning: failed to remove temp file: %v\n", rmErr)
				}
				return "", err
			}

			// Make executable
			if err := os.Chmod(tempFile.Name(), 0755); err != nil {
				if rmErr := os.Remove(tempFile.Name()); rmErr != nil {
					fmt.Printf("Warning: failed to remove temp file: %v\n", rmErr)
				}
				return "", err
			}

			return tempFile.Name(), nil
		}
	}

	return "", fmt.Errorf("vandor binary not found in archive")
}

func replaceCurrentBinary(newBinaryPath string) error {
	// Get the path of the current executable
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %v", err)
	}

	// Resolve symlinks
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %v", err)
	}

	// Create a temporary script to perform the replacement after we exit
	scriptPath, err := createReplacementScript(newBinaryPath, currentExe)
	if err != nil {
		return fmt.Errorf("failed to create replacement script: %v", err)
	}

	// Execute the replacement script in the background and exit
	fmt.Println("📥 Starting binary replacement...")
	fmt.Println("⏳ The process will restart automatically...")

	return executeReplacementScript(scriptPath)
}

func createReplacementScript(newBinaryPath, currentExe string) (string, error) {
	backupPath := currentExe + ".backup"

	// Create shell script content
	scriptContent := fmt.Sprintf(`#!/bin/bash
set -e

echo "🔄 Performing binary replacement..."

# Wait for current process to exit
sleep 1

# Create backup
if [ -f "%s" ]; then
    echo "📦 Creating backup..."
    cp "%s" "%s" || exit 1
fi

# Replace binary
echo "📥 Installing new binary..."
cp "%s" "%s" || {
    echo "❌ Failed to replace binary, restoring backup..."
    if [ -f "%s" ]; then
        cp "%s" "%s"
    fi
    exit 1
}

# Set permissions
chmod +x "%s"

# Clean up
rm -f "%s"
rm -f "%s"

echo "✅ Binary replacement completed successfully!"
echo "🎉 Vandor CLI has been upgraded!"
echo "   Run 'vandor version' to verify the new version."
echo "Hit enter to exit..."
`, currentExe, currentExe, backupPath, newBinaryPath, currentExe, backupPath, backupPath, currentExe, currentExe, backupPath, "$0")

	// Create temporary script file
	scriptFile, err := os.CreateTemp("", "vandor-upgrade-*.sh")
	if err != nil {
		return "", err
	}
	defer func() { _ = scriptFile.Close() }()

	if _, err := scriptFile.WriteString(scriptContent); err != nil {
		return "", err
	}

	// Make script executable
	if err := os.Chmod(scriptFile.Name(), 0755); err != nil {
		return "", err
	}

	return scriptFile.Name(), nil
}

func executeReplacementScript(scriptPath string) error {
	// Execute the script in the background
	cmd := exec.Command("bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the script
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start replacement script: %v", err)
	}

	// Don't wait for it to complete - we need to exit so the binary can be replaced
	fmt.Println("🚀 Replacement process started. Exiting...")
	os.Exit(0)
	return nil // This line will never be reached
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDate(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("January 2, 2006")
}

func performSourceUpgrade() error {
	fmt.Println("🔧 Building from source code...")
	fmt.Println("📋 This will:")
	fmt.Println("   1. Clone the latest code from GitHub")
	fmt.Println("   2. Build the binary using Go")
	fmt.Println("   3. Replace your current installation")
	fmt.Println()

	// Check prerequisites
	if !commandExists("git") {
		return fmt.Errorf("git is required for source installation")
	}
	if !commandExists("go") {
		return fmt.Errorf("go is required for source installation (install from https://golang.org/)")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "vandor-source-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer func() {
		if rmErr := os.RemoveAll(tempDir); rmErr != nil {
			fmt.Printf("Warning: failed to clean up temp directory: %v\n", rmErr)
		}
	}()

	fmt.Printf("📁 Working directory: %s\n", tempDir)

	// Clone repository
	fmt.Println("📥 Cloning repository...")
	if err := executeCommand("git", "clone", "https://github.com/alfariiizi/vandor-cli.git", tempDir+"/vandor-cli"); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	// Change to project directory
	projectDir := filepath.Join(tempDir, "vandor-cli")

	// Build binary
	fmt.Println("🔨 Building binary...")
	buildCmd := []string{"go", "build", "-o", "vandor", "main.go"}
	if err := executeCommandInDir(projectDir, buildCmd[0], buildCmd[1:]...); err != nil {
		return fmt.Errorf("failed to build: %v", err)
	}

	// Test the binary
	builtBinary := filepath.Join(projectDir, "vandor")
	fmt.Println("🧪 Testing built binary...")
	if err := executeCommand(builtBinary, "version"); err != nil {
		return fmt.Errorf("built binary failed to run: %v", err)
	}

	// Replace current binary
	fmt.Println("📥 Installing new binary...")
	if err := replaceCurrentBinary(builtBinary); err != nil {
		return fmt.Errorf("failed to replace binary: %v", err)
	}

	fmt.Println("✅ Successfully upgraded Vandor CLI from source!")
	fmt.Println("🎉 Run 'vandor version' to verify the installation.")

	return nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func executeCommand(name string, args ...string) error {
	return executeCommandInDir("", name, args...)
}

func executeCommandInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
