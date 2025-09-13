package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/command"
)

var addDomainCmd = &cobra.Command{
	Use:   "domain [name]",
	Short: "Create a new domain",
	Long:  `Create a new domain and regenerate domain code.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Use unified command system
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
