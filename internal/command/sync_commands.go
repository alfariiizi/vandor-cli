package command

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/alfariiizi/vandor-cli/internal/regenerate/domain"
	"github.com/alfariiizi/vandor-cli/internal/regenerate/entgo"
	"github.com/alfariiizi/vandor-cli/internal/regenerate/enum"
	"github.com/alfariiizi/vandor-cli/internal/regenerate/handler"
	"github.com/alfariiizi/vandor-cli/internal/regenerate/job"
	"github.com/alfariiizi/vandor-cli/internal/regenerate/scheduler"
	"github.com/alfariiizi/vandor-cli/internal/regenerate/seed"
	"github.com/alfariiizi/vandor-cli/internal/regenerate/service"
	"github.com/alfariiizi/vandor-cli/internal/regenerate/usecase"
	"github.com/alfariiizi/vandor-cli/internal/vpkg"
)

// SyncAllCommand implements the sync all functionality
type SyncAllCommand struct{}

func NewSyncAllCommand() *SyncAllCommand {
	return &SyncAllCommand{}
}

func (c *SyncAllCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing all code...\n")

	// Run all generation functions in the order specified by taskfile gen:all
	// Note: handler is now managed by http-huma vpkg package
	generators := []struct {
		name string
		fn   func() error
	}{
		{"domain", domain.RegenerateDomain},
		{"usecase", usecase.RegenerateUsecase},
		{"service", service.RegenerateService},
		{"job", job.RegenerateJob},
		{"scheduler", scheduler.RegenerateScheduler},
		{"seed", seed.RegenerateSeed},
		{"entgo", entgo.RegenerateEntgo},
	}

	// Run all generators in sequence
	for _, gen := range generators {
		_, _ = fmt.Fprintf(ctx.Stdout, "Regenerating %s...\n", gen.name)
		if err := gen.fn(); err != nil {
			return fmt.Errorf("failed to regenerate %s: %w", gen.name, err)
		}
	}

	// Run vpkg sync capabilities
	_, _ = fmt.Fprintf(ctx.Stdout, "Checking for vpkg sync capabilities...\n")
	vpkgSyncManager := vpkg.NewVpkgSyncManager()
	if err := vpkgSyncManager.ExecuteSyncCapabilities(); err != nil {
		return fmt.Errorf("failed to execute vpkg sync capabilities: %w", err)
	}

	// Run goimports to clean up
	if err := runCommand("goimports", "-w", "."); err != nil {
		return fmt.Errorf("failed to run goimports: %w", err)
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

	// Run core generation functions in the order specified by taskfile gen:core
	generators := []struct {
		name string
		fn   func() error
	}{
		{"usecase", usecase.RegenerateUsecase},
		{"service", service.RegenerateService},
		{"domain", domain.RegenerateDomain},
	}

	// Run all generators in sequence
	for _, gen := range generators {
		_, _ = fmt.Fprintf(ctx.Stdout, "Regenerating %s...\n", gen.name)
		if err := gen.fn(); err != nil {
			return fmt.Errorf("failed to regenerate %s: %w", gen.name, err)
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
	if err := domain.RegenerateDomain(); err != nil {
		return fmt.Errorf("failed to generate domain code: %w", err)
	}
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
	if err := usecase.RegenerateUsecase(); err != nil {
		return fmt.Errorf("failed to generate usecase code: %w", err)
	}
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
	if err := service.RegenerateService(); err != nil {
		return fmt.Errorf("failed to generate service code: %w", err)
	}
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

// SyncJobCommand implements the sync job functionality
type SyncJobCommand struct{}

func NewSyncJobCommand() *SyncJobCommand {
	return &SyncJobCommand{}
}

func (c *SyncJobCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing jobs...\n")
	if err := job.RegenerateJob(); err != nil {
		return fmt.Errorf("failed to generate job code: %w", err)
	}
	return nil
}

func (c *SyncJobCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "job",
		Category:    "sync",
		Description: "Sync job code",
		Usage:       "vandor sync job",
		Args:        []string{},
	}
}

func (c *SyncJobCommand) Validate(args []string) error {
	return nil // No arguments required
}

// SyncSchedulerCommand implements the sync scheduler functionality
type SyncSchedulerCommand struct{}

func NewSyncSchedulerCommand() *SyncSchedulerCommand {
	return &SyncSchedulerCommand{}
}

func (c *SyncSchedulerCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing schedulers...\n")
	if err := scheduler.RegenerateScheduler(); err != nil {
		return fmt.Errorf("failed to generate scheduler code: %w", err)
	}
	return nil
}

func (c *SyncSchedulerCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "scheduler",
		Category:    "sync",
		Description: "Sync scheduler code",
		Usage:       "vandor sync scheduler",
		Args:        []string{},
	}
}

func (c *SyncSchedulerCommand) Validate(args []string) error {
	return nil // No arguments required
}

// SyncEnumCommand implements the sync enum functionality
type SyncEnumCommand struct{}

func NewSyncEnumCommand() *SyncEnumCommand {
	return &SyncEnumCommand{}
}

func (c *SyncEnumCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing enums...\n")
	if err := enum.RegenerateEnum(); err != nil {
		return fmt.Errorf("failed to generate enum code: %w", err)
	}
	return nil
}

func (c *SyncEnumCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "enum",
		Category:    "sync",
		Description: "Sync enum code",
		Usage:       "vandor sync enum",
		Args:        []string{},
	}
}

func (c *SyncEnumCommand) Validate(args []string) error {
	return nil // No arguments required
}

// SyncSeedCommand implements the sync seed functionality
type SyncSeedCommand struct{}

func NewSyncSeedCommand() *SyncSeedCommand {
	return &SyncSeedCommand{}
}

func (c *SyncSeedCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing seeds...\n")
	if err := seed.RegenerateSeed(); err != nil {
		return fmt.Errorf("failed to generate seed code: %w", err)
	}
	return nil
}

func (c *SyncSeedCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "seed",
		Category:    "sync",
		Description: "Sync seed code",
		Usage:       "vandor sync seed",
		Args:        []string{},
	}
}

func (c *SyncSeedCommand) Validate(args []string) error {
	return nil // No arguments required
}

// SyncHandlerCommand implements the sync handler functionality
type SyncHandlerCommand struct{}

func NewSyncHandlerCommand() *SyncHandlerCommand {
	return &SyncHandlerCommand{}
}

func (c *SyncHandlerCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing HTTP handlers...\n")
	if err := handler.RegenerateHandler(); err != nil {
		return fmt.Errorf("failed to generate handler code: %w", err)
	}
	return nil
}

func (c *SyncHandlerCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "handler",
		Category:    "sync",
		Description: "Sync HTTP handler code",
		Usage:       "vandor sync handler",
		Args:        []string{},
	}
}

func (c *SyncHandlerCommand) Validate(args []string) error {
	return nil // No arguments required
}

// SyncDbModelCommand implements the sync db-model functionality
type SyncDbModelCommand struct{}

func NewSyncDbModelCommand() *SyncDbModelCommand {
	return &SyncDbModelCommand{}
}

func (c *SyncDbModelCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Syncing DB Model...\n")
	if err := entgo.RegenerateEntgo(); err != nil {
		return fmt.Errorf("failed to generate DB model: %w", err)
	}
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
