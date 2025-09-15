package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/vpkg"
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
		_, _ = fmt.Fprintln(w, "NAME\tTYPE\tVERSION\tDESCRIPTION")
		_, _ = fmt.Fprintln(w, "----\t----\t-------\t-----------")

		for _, pkg := range packages {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				pkg.Name, pkg.Type, pkg.Version, truncate(pkg.Description, 50))
		}
		_ = w.Flush()

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

		opts := vpkg.InstallOptions{
			Registry: vpkgRegistry,
			Dest:     vpkgDest,
			Force:    vpkgForce,
			DryRun:   vpkgDryRun,
			Version:  vpkgVersion,
		}

		// Check if we should use progress UI
		useProgress, _ := cmd.Flags().GetBool("progress")
		forceTUI, _ := cmd.Flags().GetBool("force-tui")

		// Set environment variable for forcing TUI if flag is set
		if forceTUI {
			_ = os.Setenv("FORCE_TUI", "1")
		}

		if useProgress {
			// Use progress installer with enhanced text progress
			progressInstaller := vpkg.NewProgressInstaller(vpkgRegistry, packageName)

			// Always use simple text progress (no TUI)
			if err := progressInstaller.InstallWithSimpleProgress(packageName, opts); err != nil {
				er(fmt.Sprintf("Failed to install package %s: %v", packageName, err))
			}
		} else {
			// Use simple installer without progress
			installer := vpkg.NewInstaller(vpkgRegistry)

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
		_, _ = fmt.Fprintln(w, "NAME\tTYPE\tVERSION\tPATH\tINSTALLED")
		_, _ = fmt.Fprintln(w, "----\t----\t-------\t----\t---------")

		for _, pkg := range packages {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				pkg.Name, pkg.Type, pkg.Version, pkg.Path,
				pkg.InstalledAt.Format("2006-01-02 15:04"))
		}
		_ = w.Flush()

		fmt.Printf("\nTotal: %d package(s) installed\n", len(packages))
	},
}

var vpkgGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate .tmpl files from .go files",
	Long: `Generate template files from Go source files by intelligently replacing common patterns with template variables.

This tool analyzes Go source files and converts them to vpkg templates by:
- Replacing package names with {{.Package}} variables
- Converting import paths to use {{.ImportPath}}
- Replacing struct/function names with template variables
- Adding common template patterns

Examples:
  # Process single file
  vandor vpkg generate --input redis.go --output redis.go.tmpl --pkg-name redis-cache

  # Process entire directory
  vandor vpkg generate --input-dir packages/redis-cache/files --output packages/redis-cache/templates
  vandor vpkg generate --input-dir ./files --output ./templates --pkg-name redis-cache

⚠️  Warning: This tool provides a starting point but may require manual review and adjustment.
Always verify the generated templates work correctly before using them.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputPath, _ := cmd.Flags().GetString("input")
		inputDir, _ := cmd.Flags().GetString("input-dir")
		outputPath, _ := cmd.Flags().GetString("output")
		packageName, _ := cmd.Flags().GetString("pkg-name")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Validate input flags
		if inputPath == "" && inputDir == "" {
			er("Either --input (for single file) or --input-dir (for directory) is required")
		}
		if inputPath != "" && inputDir != "" {
			er("Cannot use both --input and --input-dir. Choose one.")
		}
		if outputPath == "" {
			er("Output path is required (use --output)")
		}

		// Use inputDir if specified, otherwise use inputPath
		finalInputPath := inputPath
		if inputDir != "" {
			finalInputPath = inputDir
		}

		generator := vpkg.NewTemplateGenerator()
		opts := vpkg.GenerateOptions{
			InputPath:   finalInputPath,
			OutputPath:  outputPath,
			PackageName: packageName,
			DryRun:      dryRun,
			Verbose:     verbose,
		}

		if err := generator.Generate(opts); err != nil {
			er(fmt.Sprintf("Failed to generate templates: %v", err))
		}
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

		// Find entry point by trying common CLI entry patterns
		entryPoint := findCLIEntryPoint(targetPkg.Path)
		if entryPoint == "" {
			er(fmt.Sprintf("No CLI entry point found in package %s. Tried: main.go, cmd/main.go", packageName))
		}

		entryPath := filepath.Join(targetPkg.Path, entryPoint)

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
	vpkgCmd.AddCommand(vpkgGenerateCmd)
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
	vpkgAddCmd.Flags().Bool("progress", true, "Show installation progress with TUI (default: true)")
	vpkgAddCmd.Flags().Bool("force-tui", false, "Force TUI mode even in non-TTY environments (for testing)")

	// Remove flags
	vpkgRemoveCmd.Flags().BoolVar(&vpkgBackup, "backup", false, "Create backup before removing")

	// Generate flags
	vpkgGenerateCmd.Flags().String("input", "", "Input file path (single .go file)")
	vpkgGenerateCmd.Flags().String("input-dir", "", "Input directory path (processes all .go files)")
	vpkgGenerateCmd.Flags().String("output", "", "Output path (file or directory for .tmpl files)")
	vpkgGenerateCmd.Flags().String("pkg-name", "", "Package name for template context (e.g., 'redis-cache')")
	vpkgGenerateCmd.Flags().Bool("dry-run", false, "Show what would be generated without creating files")
	vpkgGenerateCmd.Flags().Bool("verbose", false, "Show detailed generation process")
}

// truncate truncates a string to the specified length
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

// findCLIEntryPoint finds the entry point for a CLI package by trying common patterns
func findCLIEntryPoint(packagePath string) string {
	// Common CLI entry point patterns to try
	entryPatterns := []string{
		"main.go",      // Root level main.go
		"cmd/main.go",  // cmd directory main.go
		"cmd/cli.go",   // cmd directory cli.go
		"cmd/root.go",  // cmd directory root.go (common with Cobra)
		"main/main.go", // main directory
		"cli/main.go",  // cli directory
	}

	for _, pattern := range entryPatterns {
		entryPath := filepath.Join(packagePath, pattern)
		if _, err := os.Stat(entryPath); err == nil {
			return pattern // Return relative path
		}
	}

	return "" // No entry point found
}
