package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version = "0.5.0"
	commit  = "dev"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print detailed version information for Vandor CLI including build details and system info.`,
	Run: func(cmd *cobra.Command, args []string) {
		detailed, _ := cmd.Flags().GetBool("detailed")
		
		if detailed {
			showDetailedVersion()
		} else {
			showSimpleVersion()
		}
	},
}

func showSimpleVersion() {
	fmt.Printf("Vandor CLI v%s\n", version)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("Built: %s\n", date)
}

func showDetailedVersion() {
	fmt.Println("🚀 Vandor CLI Version Information")
	fmt.Println("=================================")
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Commit:  %s\n", commit)
	fmt.Printf("Built:   %s\n", date)
	fmt.Println()
	
	// System information
	fmt.Println("💻 System Information")
	fmt.Println("--------------------")
	fmt.Printf("OS:           %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("Go Version:   %s\n", runtime.Version())
	fmt.Printf("Compiler:     %s\n", runtime.Compiler)
	fmt.Println()
	
	// Installation information
	fmt.Println("📍 Installation Information")
	fmt.Println("---------------------------")
	if exe, err := os.Executable(); err == nil {
		fmt.Printf("Binary Path: %s\n", exe)
	}
	fmt.Printf("User:        %s\n", os.Getenv("USER"))
	if pwd, err := os.Getwd(); err == nil {
		fmt.Printf("Working Dir: %s\n", pwd)
	}
	fmt.Println()
	
	// Update information
	fmt.Println("🔄 Update Information")
	fmt.Println("--------------------")
	fmt.Println("Check for updates: vandor upgrade check")
	fmt.Println("Upgrade to latest: vandor upgrade")
	fmt.Println("GitHub Releases:   https://github.com/alfariiizi/vandor-cli-cli/releases")
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("detailed", "d", false, "Show detailed version information")
}