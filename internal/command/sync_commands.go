package command

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/alfariiizi/vandor-cli/internal/generators"
)

// SyncAllCommand implements the sync all functionality
type SyncAllCommand struct{}

func NewSyncAllCommand() *SyncAllCommand {
	return &SyncAllCommand{}
}

func (c *SyncAllCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing all code...\n")

	// Use Jennifer-based generators
	generatorFuncs := []func() error{
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
	for _, gen := range generatorFuncs {
		if err := gen(); err != nil {
			return fmt.Errorf("failed to run generator: %w", err)
		}
	}

	// Run remaining commands
	for _, cmdArgs := range commands {
		if err := runCommand(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return fmt.Errorf("failed to run %s: %w", cmdArgs[0], err)
		}
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "✅ All code synced successfully!\n")
	return nil
}

func (c *SyncAllCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "all",
		Category:    "sync",
		Description: "Sync all code components",
		Usage:       "vandor sync all",
		Args:        []string{},
	}
}

func (c *SyncAllCommand) Validate(args []string) error {
	return nil // No arguments required
}

// SyncCoreCommand implements the sync core functionality
type SyncCoreCommand struct{}

func NewSyncCoreCommand() *SyncCoreCommand {
	return &SyncCoreCommand{}
}

func (c *SyncCoreCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing core code...\n")

	// Use Jennifer-based generators
	generatorFuncs := []func() error{
		generators.GenerateUsecaseRegistry,
		generators.GenerateServiceRegistry,
		generators.GenerateDomainRegistry,
	}

	// Run Jennifer-based generators
	for _, gen := range generatorFuncs {
		if err := gen(); err != nil {
			return fmt.Errorf("failed to run generator: %w", err)
		}
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "✅ Core code synced successfully!\n")
	return nil
}

func (c *SyncCoreCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "core",
		Category:    "sync",
		Description: "Sync core code (domains, usecases, services)",
		Usage:       "vandor sync core",
		Args:        []string{},
	}
}

func (c *SyncCoreCommand) Validate(args []string) error {
	return nil // No arguments required
}

// SyncDomainCommand implements the sync domain functionality
type SyncDomainCommand struct{}

func NewSyncDomainCommand() *SyncDomainCommand {
	return &SyncDomainCommand{}
}

func (c *SyncDomainCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing domain code...\n")
	if err := generators.GenerateDomainRegistry(); err != nil {
		return fmt.Errorf("failed to generate domain code: %w", err)
	}
	_, _ = fmt.Fprintf(ctx.Stdout, "✅ Domain code synced successfully!\n")
	return nil
}

func (c *SyncDomainCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "domain",
		Category:    "sync",
		Description: "Sync domain code",
		Usage:       "vandor sync domain",
		Args:        []string{},
	}
}

func (c *SyncDomainCommand) Validate(args []string) error {
	return nil // No arguments required
}

// SyncUsecaseCommand implements the sync usecase functionality
type SyncUsecaseCommand struct{}

func NewSyncUsecaseCommand() *SyncUsecaseCommand {
	return &SyncUsecaseCommand{}
}

func (c *SyncUsecaseCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing usecases...\n")
	if err := generators.GenerateUsecaseRegistry(); err != nil {
		return fmt.Errorf("failed to generate usecase code: %w", err)
	}
	_, _ = fmt.Fprintf(ctx.Stdout, "✅ Usecase code synced successfully!\n")
	return nil
}

func (c *SyncUsecaseCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "usecase",
		Category:    "sync",
		Description: "Sync usecase code",
		Usage:       "vandor sync usecase",
		Args:        []string{},
	}
}

func (c *SyncUsecaseCommand) Validate(args []string) error {
	return nil // No arguments required
}

// SyncServiceCommand implements the sync service functionality
type SyncServiceCommand struct{}

func NewSyncServiceCommand() *SyncServiceCommand {
	return &SyncServiceCommand{}
}

func (c *SyncServiceCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing services...\n")
	if err := generators.GenerateServiceRegistry(); err != nil {
		return fmt.Errorf("failed to generate service code: %w", err)
	}
	_, _ = fmt.Fprintf(ctx.Stdout, "✅ Service code synced successfully!\n")
	return nil
}

func (c *SyncServiceCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "service",
		Category:    "sync",
		Description: "Sync service code",
		Usage:       "vandor sync service",
		Args:        []string{},
	}
}

func (c *SyncServiceCommand) Validate(args []string) error {
	return nil // No arguments required
}

// SyncDbModelCommand implements the sync db-model functionality
type SyncDbModelCommand struct{}

func NewSyncDbModelCommand() *SyncDbModelCommand {
	return &SyncDbModelCommand{}
}

func (c *SyncDbModelCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing DB Model...\n")

	// Generate ent code
	if err := runCommand("go", "run", "./internal/cmd/entgo/main.go"); err != nil {
		return fmt.Errorf("failed to generate DB model: %w", err)
	}

	// Run goimports on the generated code
	if err := runCommand("goimports", "-w", "./internal/infrastructure/db/rest/."); err != nil {
		return fmt.Errorf("failed to run goimports: %w", err)
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "✅ DB Model synced successfully!\n")
	return nil
}

func (c *SyncDbModelCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "db-model",
		Category:    "sync",
		Description: "Sync database models using Ent",
		Usage:       "vandor sync db-model",
		Args:        []string{},
	}
}

func (c *SyncDbModelCommand) Validate(args []string) error {
	return nil // No arguments required
}

// Helper function to run external commands
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
