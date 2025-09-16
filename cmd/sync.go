package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/command"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync and regenerate code for various components",
	Long:  `Sync and regenerate code for domains, usecases, services, handlers, jobs, schedulers, seeds, and database models.`,
}

var syncAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Sync all code",
	Long:  `Sync all code components including domains, usecases, services, handlers, jobs, schedulers, seeds, and database models.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use unified command system
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("sync", "all")
		if !exists {
			er("Sync all command not found in registry")
		}

		// Create command context
		ctx := command.NewCommandContext(args)

		// Execute the unified command
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute sync all command: %v", err))
		}
	},
}

var syncCoreCmd = &cobra.Command{
	Use:   "core",
	Short: "Sync core code",
	Long:  `Sync core code including usecases, services, and domains.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use unified command system
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("sync", "core")
		if !exists {
			er("Sync core command not found in registry")
		}

		// Create command context
		ctx := command.NewCommandContext(args)

		// Execute the unified command
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute sync core command: %v", err))
		}
	},
}

// Individual generate commands
var syncDomainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Sync domain code",
	Run: func(cmd *cobra.Command, args []string) {
		// Use unified command system
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("sync", "domain")
		if !exists {
			er("Sync domain command not found in registry")
		}

		// Create command context
		ctx := command.NewCommandContext(args)

		// Execute the unified command
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute sync domain command: %v", err))
		}
	},
}

var syncUsecaseCmd = &cobra.Command{
	Use:   "usecase",
	Short: "Sync usecase code",
	Run: func(cmd *cobra.Command, args []string) {
		// Use unified command system
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("sync", "usecase")
		if !exists {
			er("Sync usecase command not found in registry")
		}

		// Create command context
		ctx := command.NewCommandContext(args)

		// Execute the unified command
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute sync usecase command: %v", err))
		}
	},
}

var syncServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Sync service code",
	Run: func(cmd *cobra.Command, args []string) {
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("sync", "service")
		if !exists {
			er("Sync service command not found in registry")
		}
		ctx := command.NewCommandContext(args)
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute sync service command: %v", err))
		}
	},
}

var syncJobCmd = &cobra.Command{
	Use:   "job",
	Short: "Sync job code",
	Run: func(cmd *cobra.Command, args []string) {
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("sync", "job")
		if !exists {
			er("Sync job command not found in registry")
		}
		ctx := command.NewCommandContext(args)
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute sync job command: %v", err))
		}
	},
}

var syncSchedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Sync scheduler code",
	Run: func(cmd *cobra.Command, args []string) {
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("sync", "scheduler")
		if !exists {
			er("Sync scheduler command not found in registry")
		}
		ctx := command.NewCommandContext(args)
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute sync scheduler command: %v", err))
		}
	},
}

var syncEnumCmd = &cobra.Command{
	Use:   "enum",
	Short: "Sync enum code",
	Run: func(cmd *cobra.Command, args []string) {
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("sync", "enum")
		if !exists {
			er("Sync enum command not found in registry")
		}
		ctx := command.NewCommandContext(args)
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute sync enum command: %v", err))
		}
	},
}

var syncSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Sync seed code",
	Run: func(cmd *cobra.Command, args []string) {
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("sync", "seed")
		if !exists {
			er("Sync seed command not found in registry")
		}
		ctx := command.NewCommandContext(args)
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute sync seed command: %v", err))
		}
	},
}

// var syncHandlerCmd = &cobra.Command{
// 	Use:   "handler",
// 	Short: "Sync HTTP handler code",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		registry := command.GetGlobalRegistry()
// 		unifiedCmd, exists := registry.Get("sync", "handler")
// 		if !exists {
// 			er("Sync handler command not found in registry")
// 		}
// 		ctx := command.NewCommandContext(args)
// 		if err := unifiedCmd.Execute(ctx); err != nil {
// 			er(fmt.Sprintf("Failed to execute sync handler command: %v", err))
// 		}
// 	},
// }

var syncDbModelCmd = &cobra.Command{
	Use:   "db-model",
	Short: "Sync DB Model code",
	Run: func(cmd *cobra.Command, args []string) {
		registry := command.GetGlobalRegistry()
		unifiedCmd, exists := registry.Get("sync", "db-model")
		if !exists {
			er("Sync db-model command not found in registry")
		}
		ctx := command.NewCommandContext(args)
		if err := unifiedCmd.Execute(ctx); err != nil {
			er(fmt.Sprintf("Failed to execute sync db-model command: %v", err))
		}
	},
}

// runCommand function moved to internal/command/sync_commands.go

func init() {
	rootCmd.AddCommand(syncCmd)

	// Add subcommands to sync
	syncCmd.AddCommand(syncAllCmd)
	syncCmd.AddCommand(syncCoreCmd)
	syncCmd.AddCommand(syncDomainCmd)
	syncCmd.AddCommand(syncUsecaseCmd)
	syncCmd.AddCommand(syncServiceCmd)
	syncCmd.AddCommand(syncJobCmd)
	syncCmd.AddCommand(syncSchedulerCmd)
	syncCmd.AddCommand(syncEnumCmd)
	syncCmd.AddCommand(syncSeedCmd)
	// Note: syncHandlerCmd removed - now managed by http-huma vpkg package
	syncCmd.AddCommand(syncDbModelCmd)
}
