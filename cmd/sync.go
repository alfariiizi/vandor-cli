package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/alfariiizi/vandor-cli/internal/generators"
	"github.com/spf13/cobra"
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
		fmt.Println("Syncing all code...")

		// Use Jennifer-based generators
		generators := []func() error{
			generators.GenerateDomainRegistry,
			generators.GenerateUsecaseRegistry,
			generators.GenerateServiceRegistry,
			generators.GenerateHandlerRegistry,
			generators.GenerateJobRegistry,
		}

		// Additional commands that still use old approach
		commands := [][]string{
			{"go", "run", "./internal/cmd/scheduler/cmd-regenerate-scheduler/main.go"},
			{"go", "run", "./internal/cmd/seed/cmd-generate/main.go"},
			{"go", "run", "./internal/cmd/entgo/main.go"},
			{"goimports", "-w", "."},
		}

		// Run Jennifer-based generators first
		for _, gen := range generators {
			if err := gen(); err != nil {
				er(fmt.Sprintf("Failed to run generator: %v", err))
			}
		}

		// Run remaining commands
		for _, cmdArgs := range commands {
			if err := runCommand(cmdArgs[0], cmdArgs[1:]...); err != nil {
				er(fmt.Sprintf("Failed to run %s: %v", cmdArgs[0], err))
			}
		}

		fmt.Println("✅ All code synced successfully!")
	},
}

var syncCoreCmd = &cobra.Command{
	Use:   "core",
	Short: "Sync core code",
	Long:  `Sync core code including usecases, services, and domains.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing core code...")

		// Use Jennifer-based generators
		generators := []func() error{
			generators.GenerateUsecaseRegistry,
			generators.GenerateServiceRegistry,
			generators.GenerateDomainRegistry,
		}

		// Run Jennifer-based generators
		for _, gen := range generators {
			if err := gen(); err != nil {
				er(fmt.Sprintf("Failed to run generator: %v", err))
			}
		}

		fmt.Println("✅ Core code synced successfully!")
	},
}

// Individual generate commands
var syncDomainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Sync domain code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing domain code...")
		if err := generators.GenerateDomainRegistry(); err != nil {
			er(fmt.Sprintf("Failed to generate domain code: %v", err))
		}
		fmt.Println("✅ Domain code synced successfully!")
	},
}

var syncUsecaseCmd = &cobra.Command{
	Use:   "usecase",
	Short: "Sync usecase code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing usecases...")
		if err := generators.GenerateUsecaseRegistry(); err != nil {
			er(fmt.Sprintf("Failed to generate usecase code: %v", err))
		}
		fmt.Println("✅ Usecase code synced successfully!")
	},
}

var syncServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Sync service code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing services...")
		if err := generators.GenerateServiceRegistry(); err != nil {
			er(fmt.Sprintf("Failed to generate service code: %v", err))
		}
		fmt.Println("✅ Service code synced successfully!")
	},
}

var syncJobCmd = &cobra.Command{
	Use:   "job",
	Short: "Sync job code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing jobs...")
		if err := generators.GenerateJobRegistry(); err != nil {
			er(fmt.Sprintf("Failed to generate job code: %v", err))
		}
		fmt.Println("✅ Job code synced successfully!")
	},
}

var syncSchedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Sync scheduler code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing schedulers...")
		if err := runGoCommand("run", "./internal/cmd/scheduler/cmd-regenerate-scheduler/main.go"); err != nil {
			er(fmt.Sprintf("Failed to generate scheduler code: %v", err))
		}
		fmt.Println("✅ Scheduler code synced successfully!")
	},
}

var syncEnumCmd = &cobra.Command{
	Use:   "enum",
	Short: "Sync enum code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing enums...")
		if err := runGoCommand("run", "./internal/cmd/enum/cmd/main.go", "generate"); err != nil {
			er(fmt.Sprintf("Failed to generate enum code: %v", err))
		}
		fmt.Println("✅ Enum code synced successfully!")
	},
}

var syncSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Sync seed code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing seeds...")
		if err := runGoCommand("run", "./internal/cmd/seed/cmd-generate/main.go"); err != nil {
			er(fmt.Sprintf("Failed to generate seed code: %v", err))
		}
		fmt.Println("✅ Seed code synced successfully!")
	},
}

var syncHandlerCmd = &cobra.Command{
	Use:   "handler",
	Short: "Sync HTTP handler code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing HTTP handlers...")
		if err := generators.GenerateHandlerRegistry(); err != nil {
			er(fmt.Sprintf("Failed to generate handler code: %v", err))
		}
		fmt.Println("✅ Handler code synced successfully!")
	},
}

var syncDbModelCmd = &cobra.Command{
	Use:   "db-model",
	Short: "Sync DB Model code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing DB Model...")

		// Generate ent code
		if err := runGoCommand("run", "./internal/cmd/entgo/main.go"); err != nil {
			er(fmt.Sprintf("Failed to generate DB model: %v", err))
		}

		// Run goimports on the generated code
		if err := runCommand("goimports", "-w", "./internal/infrastructure/db/rest/."); err != nil {
			er(fmt.Sprintf("Failed to run goimports: %v", err))
		}

		fmt.Println("✅ DB Model synced successfully!")
	},
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

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
	syncCmd.AddCommand(syncHandlerCmd)
	syncCmd.AddCommand(syncDbModelCmd)
}
