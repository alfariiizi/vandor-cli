package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	GitHubAPIURL = "https://api.github.com/repos/alfariiizi/vandor-cli/releases/latest"
	CurrentVersion = "0.5.0" // This should match the version in version.go
)

type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Assets     []GitHubAsset `json:"assets"`
	CreatedAt  string `json:"created_at"`
	PublishedAt string `json:"published_at"`
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

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.AddCommand(upgradeCheckCmd)
}

func performUpgrade() error {
	fmt.Println("🔍 Checking for the latest version...")
	
	release, hasUpdate, err := checkForUpdates(false)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %v", err)
	}

	if !hasUpdate {
		fmt.Printf("✅ You're already running the latest version (%s)!\n", CurrentVersion)
		return nil
	}

	fmt.Printf("📦 New version available: %s -> %s\n", CurrentVersion, release.TagName)
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
	defer os.Remove(tempFile)

	// Extract if it's a compressed archive
	binaryPath := tempFile
	if strings.HasSuffix(asset.Name, ".tar.gz") {
		extractedPath, err := extractTarGz(tempFile)
		if err != nil {
			return fmt.Errorf("failed to extract archive: %v", err)
		}
		binaryPath = extractedPath
		defer os.Remove(extractedPath)
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
		fmt.Printf("🔍 Checking for updates (current version: %s)...\n", CurrentVersion)
	}

	resp, err := http.Get(GitHubAPIURL)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch release info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, false, fmt.Errorf("failed to parse release info: %v", err)
	}

	// Clean version strings for comparison
	currentVer := strings.TrimPrefix(CurrentVersion, "v")
	latestVer := strings.TrimPrefix(release.TagName, "v")

	hasUpdate := currentVer != latestVer

	if verbose {
		if hasUpdate {
			fmt.Printf("🆕 New version available: %s -> %s\n", CurrentVersion, release.TagName)
			fmt.Printf("📅 Released: %s\n", formatDate(release.PublishedAt))
			fmt.Println("💡 Run 'vandor upgrade' to install the latest version.")
		} else {
			fmt.Printf("✅ You're running the latest version (%s)!\n", CurrentVersion)
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "vandor-upgrade-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Copy with progress indication
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

func extractTarGz(srcPath string) (string, error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return "", err
	}
	defer gzReader.Close()

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
			defer tempFile.Close()

			_, err = io.Copy(tempFile, tarReader)
			if err != nil {
				os.Remove(tempFile.Name())
				return "", err
			}

			// Make executable
			if err := os.Chmod(tempFile.Name(), 0755); err != nil {
				os.Remove(tempFile.Name())
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

	// Create backup of current binary
	backupPath := currentExe + ".backup"
	if err := copyFile(currentExe, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Replace the binary
	if err := copyFile(newBinaryPath, currentExe); err != nil {
		// Restore backup if replacement fails
		copyFile(backupPath, currentExe)
		os.Remove(backupPath)
		return fmt.Errorf("failed to replace binary: %v", err)
	}

	// Make executable
	if err := os.Chmod(currentExe, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %v", err)
	}

	// Clean up backup
	os.Remove(backupPath)

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
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