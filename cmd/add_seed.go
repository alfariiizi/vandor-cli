package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/tui"
)

var addSeedCmd = &cobra.Command{
	Use:   "seed [name]",
	Short: "Create a new seed",
	Long:  `Create a new seed and generate seed code. If no name is provided, opens TUI for interactive input.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, launch TUI for this specific command
		if len(args) == 0 {
			if err := tui.LaunchDirectCommand("add", "seed"); err != nil {
				er(fmt.Sprintf("Failed to launch TUI: %v", err))
			}
			return
		}

		name := args[0]
		fmt.Printf("Creating new seed: %s\n", name)

		// Create new seed
		if err := runGoCommand("run", "./internal/cmd/seed/cmd-new-seed/main.go", name); err != nil {
			er(fmt.Sprintf("Failed to create seed: %v", err))
		}

		// Generate seed code
		if err := runGoCommand("run", "./internal/cmd/seed/cmd-generate/main.go"); err != nil {
			er(fmt.Sprintf("Failed to generate seed code: %v", err))
		}

		fmt.Printf("✅ Seed '%s' created successfully!\n", name)
	},
}

func init() {
	addCmd.AddCommand(addSeedCmd)
}
