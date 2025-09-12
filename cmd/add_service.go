package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vandordev/vandor/internal/generators"
)

var addServiceCmd = &cobra.Command{
	Use:   "service [group] [name]",
	Short: "Create a new service",
	Long:  `Create a new service in the specified group.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		group := args[0]
		name := args[1]
		fmt.Printf("Creating new service: %s in group %s\n", name, group)
		
		// Create new service using Jennifer generator
		if err := generators.GenerateService(name); err != nil {
			er(fmt.Sprintf("Failed to create service: %v", err))
		}
		
		// Auto-sync service registry
		fmt.Println("Auto-syncing service registry...")
		if err := generators.GenerateServiceRegistry(); err != nil {
			er(fmt.Sprintf("Failed to sync service registry: %v", err))
		}
		
		fmt.Printf("✅ Service '%s' created and synced successfully in group '%s'!\n", name, group)
	},
}

func init() {
	addCmd.AddCommand(addServiceCmd)
}