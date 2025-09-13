package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/command"
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
		// Use unified command system
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("theme", "list")
		if !exists {
			er("Theme list command not found in registry")
		}

		// Create command context
		ctx := command.NewCommandContext(args)

		// Execute the unified command
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute theme list command: %v", err))
		}
	},
}

var themeSetCmd = &cobra.Command{
	Use:   "set [theme-name]",
	Short: "Set the active theme",
	Long:  `Set the active theme for the Vandor CLI. Use 'vandor theme list' to see available themes.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Use unified command system
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("theme", "set")
		if !exists {
			er("Theme set command not found in registry")
		}

		// Create command context
		ctx := command.NewCommandContext(args)

		// Execute the unified command
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute theme set command: %v", err))
		}
	},
}

var themeInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current theme information",
	Long:  `Display information about the currently active theme.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use unified command system
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("theme", "info")
		if !exists {
			er("Theme info command not found in registry")
		}

		// Create command context
		ctx := command.NewCommandContext(args)

		// Execute the unified command
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute theme info command: %v", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(themeCmd)
	themeCmd.AddCommand(themeListCmd)
	themeCmd.AddCommand(themeSetCmd)
	themeCmd.AddCommand(themeInfoCmd)
}
