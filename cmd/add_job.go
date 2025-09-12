package cmd

import (
	"fmt"

	"github.com/alfariiizi/vandor-cli/internal/generators"
	"github.com/spf13/cobra"
)

var addJobCmd = &cobra.Command{
	Use:   "job [name]",
	Short: "Create a new job",
	Long:  `Create a new job and regenerate job code.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Printf("Creating new job: %s\n", name)

		// Create new job using Jennifer generator
		if err := generators.GenerateJob(name); err != nil {
			er(fmt.Sprintf("Failed to create job: %v", err))
		}

		// Auto-sync job registry
		fmt.Println("Auto-syncing job registry...")
		if err := generators.GenerateJobRegistry(); err != nil {
			er(fmt.Sprintf("Failed to sync job registry: %v", err))
		}

		fmt.Printf("✅ Job '%s' created and synced successfully!\n", name)
	},
}

func init() {
	addCmd.AddCommand(addJobCmd)
}
