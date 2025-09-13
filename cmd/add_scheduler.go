package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/generators"
	"github.com/alfariiizi/vandor-cli/internal/tui"
)

var addSchedulerCmd = &cobra.Command{
	Use:   "scheduler [name]",
	Short: "Create a new scheduler",
	Long:  `Create a new scheduler and regenerate scheduler code. If no name is provided, opens TUI for interactive input.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, launch TUI for this specific command
		if len(args) == 0 {
			if err := tui.LaunchDirectCommand("add", "scheduler"); err != nil {
				er(fmt.Sprintf("Failed to launch TUI: %v", err))
			}
			return
		}

		name := args[0]
		fmt.Printf("Creating new scheduler: %s\n", name)

		// Create new scheduler using Jennifer generator
		if err := generators.GenerateScheduler(name); err != nil {
			er(fmt.Sprintf("Failed to create scheduler: %v", err))
		}

		// Auto-sync scheduler registry (keeping old approach for now)
		fmt.Println("Auto-syncing scheduler registry...")
		if err := runGoCommand("run", "./internal/cmd/scheduler/cmd-regenerate-scheduler/main.go"); err != nil {
			er(fmt.Sprintf("Failed to sync scheduler registry: %v", err))
		}

		fmt.Printf("✅ Scheduler '%s' created and synced successfully!\n", name)
	},
}

func init() {
	addCmd.AddCommand(addSchedulerCmd)
}
