package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/theme"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage CLI themes",
	Long:  `Manage and switch between different visual themes for the Vandor CLI.`,
}

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available themes",
	Long:  `List all available themes for the Vandor CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available themes:")
		fmt.Println()
		fmt.Println("Catppuccin themes:")
		fmt.Println("  mocha    - Catppuccin Mocha (dark, recommended for dark terminals)")
		fmt.Println("  latte    - Catppuccin Latte (light, recommended for light terminals)")
		fmt.Println("  frappe   - Catppuccin Frappe (medium contrast)")
		fmt.Println()
		fmt.Println("Other themes:")
		fmt.Println("  default  - Default adaptive theme")
		fmt.Println("  dracula  - Dracula inspired theme")
		fmt.Println("  auto     - Auto-detect based on system theme (default)")
		fmt.Println()
		fmt.Println("Usage: vandor theme set <theme-name>")
	},
}

var themeSetCmd = &cobra.Command{
	Use:   "set [theme-name]",
	Short: "Set the active theme",
	Long:  `Set the active theme for the Vandor CLI. Use 'vandor theme list' to see available themes.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		themeName := args[0]

		switch themeName {
		case "mocha":
			theme.SetCatppuccinMochaTheme()
			fmt.Println("✅ Theme set to Catppuccin Mocha (dark)")
		case "latte":
			theme.SetCatppuccinLatteTheme()
			fmt.Println("✅ Theme set to Catppuccin Latte (light)")
		case "frappe":
			theme.SetCatppuccinFrappeTheme()
			fmt.Println("✅ Theme set to Catppuccin Frappe (medium)")
		case "default":
			theme.SetDefaultTheme()
			fmt.Println("✅ Theme set to Default")
		case "dracula":
			theme.SetDraculaTheme()
			fmt.Println("✅ Theme set to Dracula")
		case "auto":
			theme.SetAutoTheme()
			fmt.Println("✅ Theme set to Auto (system-detected)")
		default:
			fmt.Printf("❌ Unknown theme: %s\n", themeName)
			fmt.Println("Use 'vandor theme list' to see available themes.")
		}

		fmt.Println("Run 'vandor tui' to see the new theme in action!")
	},
}

var themeInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current theme information",
	Long:  `Display information about the currently active theme.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🎨 Vandor CLI Theme System")
		fmt.Println()
		fmt.Println("Current theme: Auto-detected Catppuccin")
		fmt.Println("- Dark terminals: Catppuccin Mocha")
		fmt.Println("- Light terminals: Catppuccin Latte")
		fmt.Println()
		fmt.Println("The theme system automatically adapts to your terminal's appearance.")
		fmt.Println("You can override this with 'vandor theme set <theme-name>'")
		fmt.Println()
		fmt.Println("For best experience:")
		fmt.Println("- Use 'mocha' theme with dark terminal backgrounds")
		fmt.Println("- Use 'latte' theme with light terminal backgrounds")
		fmt.Println("- Use 'auto' to let the system choose automatically")
	},
}

func init() {
	rootCmd.AddCommand(themeCmd)
	themeCmd.AddCommand(themeListCmd)
	themeCmd.AddCommand(themeSetCmd)
	themeCmd.AddCommand(themeInfoCmd)
}
