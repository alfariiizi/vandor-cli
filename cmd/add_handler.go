package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/generators"
	"github.com/alfariiizi/vandor-cli/internal/tui"
	"github.com/alfariiizi/vandor-cli/internal/utils"
)

var addHandlerCmd = &cobra.Command{
	Use:   "handler [group] [name] [method]",
	Short: "Create a new HTTP handler",
	Long:  `Create a new HTTP handler with the specified group, name, and HTTP method. If no arguments are provided, opens TUI for interactive input.`,
	Args:  cobra.MaximumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, launch TUI for this specific command
		if len(args) == 0 {
			if err := tui.LaunchDirectCommand("add", "handler"); err != nil {
				er(fmt.Sprintf("Failed to launch TUI: %v", err))
			}
			return
		}

		// If not enough args provided, show error
		if len(args) < 3 {
			er("Group, name, and method are required. Usage: vandor add handler <group> <name> <method>")
		}

		group := args[0]
		name := args[1]
		method := strings.ToUpper(args[2])

		fmt.Printf("Creating new HTTP handler: %s in group %s with method %s\n", name, group, method)

		// Create new HTTP handler using Jennifer generator
		if err := generators.GenerateHandler(name, group, method); err != nil {
			er(fmt.Sprintf("Failed to create HTTP handler: %v", err))
		}

		// Auto-sync handler registry
		fmt.Println("Auto-syncing handler registry...")
		if err := generators.GenerateHandlerRegistry(); err != nil {
			er(fmt.Sprintf("Failed to sync handler registry: %v", err))
		}

		fmt.Printf("✅ HTTP handler '%s' created and synced successfully in group '%s' with method '%s'!\n", name, group, method)
	},
}

var addHandlerCrudCmd = &cobra.Command{
	Use:   "handler-crud [model]",
	Short: "Generate CRUD HTTP handlers for a model",
	Long:  `Generate CRUD (Create, Read, Update, Delete) HTTP handlers for the specified model. If no model is provided, opens TUI for interactive input.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, launch TUI for this specific command
		if len(args) == 0 {
			if err := tui.LaunchDirectCommand("add", "handler-crud"); err != nil {
				er(fmt.Sprintf("Failed to launch TUI: %v", err))
			}
			return
		}

		model := args[0]
		modelTitle := utils.ToPascalCase(model)

		fmt.Printf("Generating CRUD HTTP handlers for model: %s\n", modelTitle)

		// Generate CRUD handlers
		if err := runGoCommand("run", "./internal/cmd/http/crud/main.go", modelTitle); err != nil {
			er(fmt.Sprintf("Failed to generate CRUD handlers: %v", err))
		}

		// Auto-sync handler registry
		fmt.Println("Auto-syncing handler registry...")
		if err := generators.GenerateHandlerRegistry(); err != nil {
			er(fmt.Sprintf("Failed to sync handler registry: %v", err))
		}

		fmt.Printf("✅ CRUD HTTP handlers for model '%s' generated and synced successfully!\n", modelTitle)
	},
}

var addServiceHandlerCmd = &cobra.Command{
	Use:   "service-handler [group] [name] [method]",
	Short: "Create a new service and HTTP handler together",
	Long:  `Create both a service and its corresponding HTTP handler with the specified group, name, and HTTP method. If no arguments are provided, opens TUI for interactive input.`,
	Args:  cobra.MaximumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, launch TUI for this specific command
		if len(args) == 0 {
			if err := tui.LaunchDirectCommand("add", "service-handler"); err != nil {
				er(fmt.Sprintf("Failed to launch TUI: %v", err))
			}
			return
		}

		// If not enough args provided, show error
		if len(args) < 3 {
			er("Group, name, and method are required. Usage: vandor add service-handler <group> <name> <method>")
		}

		group := args[0]
		name := args[1]
		method := strings.ToUpper(args[2])

		fmt.Printf("Creating new service and HTTP handler: %s in group %s with method %s\n", name, group, method)

		// Create service using Jennifer generator
		if err := generators.GenerateService(name); err != nil {
			er(fmt.Sprintf("Failed to create service: %v", err))
		}

		// Create HTTP handler using Jennifer generator
		if err := generators.GenerateHandler(name, group, method); err != nil {
			er(fmt.Sprintf("Failed to create HTTP handler: %v", err))
		}

		// Auto-sync both service and handler registries
		fmt.Println("Auto-syncing service and handler registries...")
		if err := generators.GenerateServiceRegistry(); err != nil {
			er(fmt.Sprintf("Failed to sync service registry: %v", err))
		}
		if err := generators.GenerateHandlerRegistry(); err != nil {
			er(fmt.Sprintf("Failed to sync handler registry: %v", err))
		}

		fmt.Printf("✅ Service and HTTP handler '%s' created and synced successfully in group '%s' with method '%s'!\n", name, group, method)
	},
}

func init() {
	addCmd.AddCommand(addHandlerCmd)
	addCmd.AddCommand(addHandlerCrudCmd)
	addCmd.AddCommand(addServiceHandlerCmd)
}
