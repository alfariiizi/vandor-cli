package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/generators"
)

var addUsecaseCmd = &cobra.Command{
	Use:   "usecase [name]",
	Short: "Create a new usecase",
	Long:  `Create a new usecase for business logic.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Printf("Creating new usecase: %s\n", name)

		// Create new usecase using Jennifer generator
		if err := generators.GenerateUsecase(name); err != nil {
			er(fmt.Sprintf("Failed to create usecase: %v", err))
		}

		// Auto-sync usecase registry
		fmt.Println("Auto-syncing usecase registry...")
		if err := generators.GenerateUsecaseRegistry(); err != nil {
			er(fmt.Sprintf("Failed to sync usecase registry: %v", err))
		}

		fmt.Printf("✅ Usecase '%s' created and synced successfully!\n", name)
	},
}

func init() {
	addCmd.AddCommand(addUsecaseCmd)
}
