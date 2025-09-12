package cmd

import (
	"fmt"
	"strings"

	"github.com/alfariiizi/vandor-cli/internal/vpkg"
	"github.com/spf13/cobra"
)

var vpkgCmd = &cobra.Command{
	Use:   "vpkg",
	Short: "Manage Vandor packages",
	Long:  `Manage Vandor packages (vpkg) - add, remove, list, and update packages for your project.`,
}

var vpkgAddCmd = &cobra.Command{
	Use:   "add [package-name]",
	Short: "Add a Vandor package",
	Long: `Add a Vandor package to your project. The package can be from official Vandor repository or community packages.
	
Examples:
  vandor vpkg add redis-cache
  vandor vpkg add audit-logger
  vandor vpkg add kafka-bus`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]

		fmt.Printf("Adding Vandor package: %s\n", packageName)

		manager, err := vpkg.NewManager()
		if err != nil {
			er(fmt.Sprintf("Failed to initialize package manager: %v", err))
		}

		if err := manager.AddPackage(packageName); err != nil {
			er(fmt.Sprintf("Failed to add package %s: %v", packageName, err))
		}

		fmt.Printf("✅ Package '%s' added successfully!\n", packageName)
	},
}

var vpkgRemoveCmd = &cobra.Command{
	Use:   "remove [package-name]",
	Short: "Remove a Vandor package",
	Long:  `Remove a Vandor package from your project and clean up associated files.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]

		fmt.Printf("Removing Vandor package: %s\n", packageName)

		manager, err := vpkg.NewManager()
		if err != nil {
			er(fmt.Sprintf("Failed to initialize package manager: %v", err))
		}

		if err := manager.RemovePackage(packageName); err != nil {
			er(fmt.Sprintf("Failed to remove package %s: %v", packageName, err))
		}

		fmt.Printf("✅ Package '%s' removed successfully!\n", packageName)
	},
}

var vpkgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Vandor packages",
	Long:  `List all installed Vandor packages in your project.`,
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := vpkg.NewManager()
		if err != nil {
			er(fmt.Sprintf("Failed to initialize package manager: %v", err))
		}

		packages, err := manager.ListPackages()
		if err != nil {
			er(fmt.Sprintf("Failed to list packages: %v", err))
		}

		if len(packages) == 0 {
			fmt.Println("No Vandor packages installed.")
			return
		}

		fmt.Println("Installed Vandor packages:")
		for _, pkg := range packages {
			fmt.Printf("  - %s (v%s) [%s]\n", pkg.Name, pkg.Version, strings.Join(pkg.Tags, ", "))
		}
	},
}

var vpkgUpdateCmd = &cobra.Command{
	Use:   "update [package-name]",
	Short: "Update a Vandor package",
	Long:  `Update a Vandor package to the latest version. If no package name is provided, updates all packages.`,
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := vpkg.NewManager()
		if err != nil {
			er(fmt.Sprintf("Failed to initialize package manager: %v", err))
		}

		if len(args) == 0 {
			fmt.Println("Updating all Vandor packages...")
			if err := manager.UpdateAllPackages(); err != nil {
				er(fmt.Sprintf("Failed to update packages: %v", err))
			}
			fmt.Println("✅ All packages updated successfully!")
		} else {
			packageName := args[0]
			fmt.Printf("Updating Vandor package: %s\n", packageName)
			if err := manager.UpdatePackage(packageName); err != nil {
				er(fmt.Sprintf("Failed to update package %s: %v", packageName, err))
			}
			fmt.Printf("✅ Package '%s' updated successfully!\n", packageName)
		}
	},
}

var vpkgSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for available Vandor packages",
	Long:  `Search for available Vandor packages in the official and community repositories.`,
	Run: func(cmd *cobra.Command, args []string) {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		manager, err := vpkg.NewManager()
		if err != nil {
			er(fmt.Sprintf("Failed to initialize package manager: %v", err))
		}

		packages, err := manager.SearchPackages(query)
		if err != nil {
			er(fmt.Sprintf("Failed to search packages: %v", err))
		}

		if len(packages) == 0 {
			fmt.Println("No packages found.")
			return
		}

		fmt.Println("Available Vandor packages:")
		for _, pkg := range packages {
			fmt.Printf("  - %s (v%s) - %s\n", pkg.Name, pkg.Version, pkg.Description)
			if len(pkg.Tags) > 0 {
				fmt.Printf("    Tags: [%s]\n", strings.Join(pkg.Tags, ", "))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(vpkgCmd)

	vpkgCmd.AddCommand(vpkgAddCmd)
	vpkgCmd.AddCommand(vpkgRemoveCmd)
	vpkgCmd.AddCommand(vpkgListCmd)
	vpkgCmd.AddCommand(vpkgUpdateCmd)
	vpkgCmd.AddCommand(vpkgSearchCmd)
}
