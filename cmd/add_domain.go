package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/command"
	"github.com/alfariiizi/vandor-cli/internal/tui"
)

var addDomainCmd = &cobra.Command{
	Use:   "domain [name]",
	Short: "Create a new domain",
	Long:  `Create a new domain and regenerate domain code. If no name is provided, opens TUI for interactive input.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, launch TUI for this specific command
		if len(args) == 0 {
			if err := tui.LaunchDirectCommand("add", "domain"); err != nil {
				er(fmt.Sprintf("Failed to launch TUI: %v", err))
			}
			return
		}

		// Use unified command system for direct execution
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("add", "domain")
		if !exists {
			er("Domain command not found in registry")
		}

		// Create command context
		ctx := command.NewCommandContext(args)

		// Execute the unified command
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute domain command: %v", err))
		}
	},
}

func init() {
	addCmd.AddCommand(addDomainCmd)
}
