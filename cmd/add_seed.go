package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addSeedCmd = &cobra.Command{
	Use:   "seed [name]",
	Short: "Create a new seed",
	Long:  `Create a new seed and generate seed code.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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