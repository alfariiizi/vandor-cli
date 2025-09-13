package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/generators"
)

var addDomainCmd = &cobra.Command{
	Use:   "domain [name]",
	Short: "Create a new domain",
	Long:  `Create a new domain and regenerate domain code.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Printf("Creating new domain: %s\n", name)

		// Create new domain using Jennifer generator
		if err := generators.GenerateDomain(name); err != nil {
			er(fmt.Sprintf("Failed to create domain: %v", err))
		}

		// Auto-sync domain registry
		fmt.Println("Auto-syncing domain registry...")
		if err := generators.GenerateDomainRegistry(); err != nil {
			er(fmt.Sprintf("Failed to sync domain registry: %v", err))
		}

		fmt.Printf("✅ Domain '%s' created and synced successfully!\n", name)
	},
}

func init() {
	addCmd.AddCommand(addDomainCmd)
}
