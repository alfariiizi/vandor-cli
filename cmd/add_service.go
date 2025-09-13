package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/command"
	"github.com/alfariiizi/vandor-cli/internal/tui"
)

var addServiceCmd = &cobra.Command{
	Use:   "service [group] [name]",
	Short: "Create a new service",
	Long:  `Create a new service in the specified group. If no arguments are provided, opens TUI for interactive input.`,
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, launch TUI for this specific command
		if len(args) == 0 {
			if err := tui.LaunchDirectCommand("add", "service"); err != nil {
				er(fmt.Sprintf("Failed to launch TUI: %v", err))
			}
			return
		}

		// If not enough args provided, show error
		if len(args) < 2 {
			er("Both service group and name are required. Usage: vandor add service <group> <name>")
		}

		// Use unified command system for direct execution
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("add", "service")
		if !exists {
			er("Service command not found in registry")
		}

		// Create command context
		ctx := command.NewCommandContext(args)

		// Execute the unified command
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute service command: %v", err))
		}
	},
}

func init() {
	addCmd.AddCommand(addServiceCmd)
}
