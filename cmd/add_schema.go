package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/tui"
)

var addSchemaCmd = &cobra.Command{
	Use:   "schema [name]",
	Short: "Create a new schema",
	Long:  `Create a new schema using ent-tools. If no name is provided, opens TUI for interactive input.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, launch TUI for this specific command
		if len(args) == 0 {
			if err := tui.LaunchDirectCommand("add", "schema"); err != nil {
				er(fmt.Sprintf("Failed to launch TUI: %v", err))
			}
			return
		}

		name := args[0]
		fmt.Printf("Creating new schema: %s\n", name)

		if err := runScript("./internal/scripts/ent-tools.sh", "new", name); err != nil {
			er(fmt.Sprintf("Failed to create schema: %v", err))
		}

		fmt.Printf("✅ Schema '%s' created successfully!\n", name)
	},
}

func init() {
	addCmd.AddCommand(addSchemaCmd)
}
