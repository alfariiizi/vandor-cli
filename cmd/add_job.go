package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/generators"
	"github.com/alfariiizi/vandor-cli/internal/tui"
)

var addJobCmd = &cobra.Command{
	Use:   "job [name]",
	Short: "Create a new job",
	Long:  `Create a new job and regenerate job code. If no name is provided, opens TUI for interactive input.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, launch TUI for this specific command
		if len(args) == 0 {
			if err := tui.LaunchDirectCommand("add", "job"); err != nil {
				er(fmt.Sprintf("Failed to launch TUI: %v", err))
			}
			return
		}

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
