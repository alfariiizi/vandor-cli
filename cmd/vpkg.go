package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/alfariiizi/vandor-cli/internal/vpkg"
	"github.com/spf13/cobra"
)

var vpkgCmd = &cobra.Command{
	Use:   "vpkg",
	Short: "Manage Vandor packages",
	Long: `Manage Vandor packages (vpkg) - add, remove, list, and execute packages for your project.

The vpkg system allows you to install reusable components into your project:
- fx-module packages: Library modules that integrate with Fx dependency injection
- cli-command packages: Executable CLI tools that can be run via 'vpkg exec'`,
}

var (
	vpkgRegistry string
	vpkgForce    bool
	vpkgDryRun   bool
	vpkgDest     string
	vpkgVersion  string
	vpkgBackup   bool
	vpkgTags     []string
	vpkgType     string
)

var vpkgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available packages from registry",
	Long:  `List all available packages from the Vandor package registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := vpkg.NewRegistryClient(vpkgRegistry)
		
		opts := vpkg.ListOptions{
			Registry: vpkgRegistry,
			Tags:     vpkgTags,
			Type:     vpkgType,
		}
		
		packages, err := client.ListPackages(opts)
		if err != nil {
			er(fmt.Sprintf("Failed to list packages: %v", err))
		}

		if len(packages) == 0 {
			fmt.Println("No packages found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tVERSION\tDESCRIPTION")
		fmt.Fprintln(w, "----\t----\t-------\t-----------")
		
		for _, pkg := range packages {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
				pkg.Name, pkg.Type, pkg.Version, truncate(pkg.Description, 50))
		}
		w.Flush()

		fmt.Printf("\nFound %d package(s)\n", len(packages))
		if len(vpkgTags) > 0 || vpkgType != "" {
			fmt.Printf("Filters: ")
			if vpkgType != "" {
				fmt.Printf("type=%s ", vpkgType)
			}
			if len(vpkgTags) > 0 {
				fmt.Printf("tags=%s", strings.Join(vpkgTags, ","))
			}
			fmt.Println()
		}
	},
}

var vpkgAddCmd = &cobra.Command{
	Use:   "add [package-name][@version]",
	Short: "Add a Vandor package",
	Long: `Add a Vandor package to your project by downloading and installing its templates.

Examples:
  vandor vpkg add vandor/redis-cache
  vandor vpkg add vandor/redis-cache@v0.2.0
  vandor vpkg add acme/migrate-db --dest internal/tools/migrate`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]

		installer := vpkg.NewInstaller(vpkgRegistry)
		opts := vpkg.InstallOptions{
			Registry: vpkgRegistry,
			Dest:     vpkgDest,
			Force:    vpkgForce,
			DryRun:   vpkgDryRun,
			Version:  vpkgVersion,
		}

		fmt.Printf("Installing package: %s\n", packageName)
		if vpkgDryRun {
			fmt.Println("(Dry run - no changes will be made)")
		}

		if err := installer.Install(packageName, opts); err != nil {
			er(fmt.Sprintf("Failed to install package %s: %v", packageName, err))
		}

		if !vpkgDryRun {
			fmt.Printf("✅ Package '%s' installed successfully!\n", packageName)
		} else {
			fmt.Printf("✅ Package '%s' would be installed successfully\n", packageName)
		}
	},
}

var vpkgRemoveCmd = &cobra.Command{
	Use:   "remove [package-name]",
	Short: "Remove an installed Vandor package",
	Long:  `Remove an installed Vandor package from your project and clean up its files.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]

		installer := vpkg.NewInstaller(vpkgRegistry)
		
		fmt.Printf("Removing package: %s\n", packageName)
		if vpkgBackup {
			fmt.Println("Creating backup before removal...")
		}

		if err := installer.Remove(packageName, vpkgBackup); err != nil {
			er(fmt.Sprintf("Failed to remove package %s: %v", packageName, err))
		}
	},
}

var vpkgListInstalledCmd = &cobra.Command{
	Use:   "list-installed",
	Short: "List installed packages",
	Long:  `List all packages currently installed in your project.`,
	Run: func(cmd *cobra.Command, args []string) {
		installer := vpkg.NewInstaller(vpkgRegistry)

		packages, err := installer.ListInstalled()
		if err != nil {
			er(fmt.Sprintf("Failed to list installed packages: %v", err))
		}

		if len(packages) == 0 {
			fmt.Println("No packages installed.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tVERSION\tPATH\tINSTALLED")
		fmt.Fprintln(w, "----\t----\t-------\t----\t---------")
		
		for _, pkg := range packages {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", 
				pkg.Name, pkg.Type, pkg.Version, pkg.Path, 
				pkg.InstalledAt.Format("2006-01-02 15:04"))
		}
		w.Flush()

		fmt.Printf("\nTotal: %d package(s) installed\n", len(packages))
	},
}

var vpkgExecCmd = &cobra.Command{
	Use:   "exec [package-name] [args...]",
	Short: "Execute a CLI package",
	Long: `Execute a CLI package in exec mode. This runs the package's main command 
without requiring it to be built into your application.

Examples:
  vandor vpkg exec acme/migrate-db status
  vandor vpkg exec acme/migrate-db up`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]
		packageArgs := args[1:]

		installer := vpkg.NewInstaller(vpkgRegistry)

		// Find installed package
		packages, err := installer.ListInstalled()
		if err != nil {
			er(fmt.Sprintf("Failed to list installed packages: %v", err))
		}

		var targetPkg *vpkg.InstalledPackage
		for _, pkg := range packages {
			if pkg.Name == packageName {
				targetPkg = &pkg
				break
			}
		}

		if targetPkg == nil {
			er(fmt.Sprintf("Package %s is not installed. Install it with 'vandor vpkg add %s'", packageName, packageName))
		}

		if targetPkg.Type != "cli-command" {
			er(fmt.Sprintf("Package %s is not a CLI command (type: %s). Only cli-command packages can be executed.", packageName, targetPkg.Type))
		}

		// Find entry point
		entryPoint := targetPkg.Meta.Entry
		if entryPoint == "" {
			entryPoint = "cmd/main.go"
		}

		entryPath := filepath.Join(targetPkg.Path, entryPoint)
		if _, err := os.Stat(entryPath); os.IsNotExist(err) {
			er(fmt.Sprintf("Entry point %s not found in package %s", entryPath, packageName))
		}

		// Execute the package
		fmt.Printf("Executing %s with args: %v\n", packageName, packageArgs)
		
		cmdArgs := append([]string{"run", entryPath}, packageArgs...)
		execCmd := exec.Command("go", cmdArgs...)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin

		if err := execCmd.Run(); err != nil {
			er(fmt.Sprintf("Failed to execute package %s: %v", packageName, err))
		}
	},
}

func init() {
	rootCmd.AddCommand(vpkgCmd)

	// Add subcommands
	vpkgCmd.AddCommand(vpkgListCmd)
	vpkgCmd.AddCommand(vpkgAddCmd)
	vpkgCmd.AddCommand(vpkgRemoveCmd)
	vpkgCmd.AddCommand(vpkgListInstalledCmd)
	vpkgCmd.AddCommand(vpkgExecCmd)

	// Global flags
	vpkgCmd.PersistentFlags().StringVar(&vpkgRegistry, "registry", "", "Alternative registry URL")

	// List flags
	vpkgListCmd.Flags().StringSliceVar(&vpkgTags, "tags", []string{}, "Filter by tags (comma-separated)")
	vpkgListCmd.Flags().StringVar(&vpkgType, "type", "", "Filter by package type (fx-module, cli-command)")

	// Add flags
	vpkgAddCmd.Flags().StringVar(&vpkgDest, "dest", "", "Override destination path")
	vpkgAddCmd.Flags().BoolVar(&vpkgForce, "force", false, "Overwrite existing files")
	vpkgAddCmd.Flags().BoolVar(&vpkgDryRun, "dry-run", false, "Show what would be done without making changes")
	vpkgAddCmd.Flags().StringVar(&vpkgVersion, "version", "", "Specific version to install")

	// Remove flags
	vpkgRemoveCmd.Flags().BoolVar(&vpkgBackup, "backup", false, "Create backup before removing")
}

// truncate truncates a string to the specified length
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}