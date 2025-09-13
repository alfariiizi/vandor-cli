package command

import (
	"fmt"

	"github.com/alfariiizi/vandor-cli/internal/generators"
)

// AddDomainCommand implements the add domain functionality
type AddDomainCommand struct{}

func NewAddDomainCommand() *AddDomainCommand {
	return &AddDomainCommand{}
}

func (c *AddDomainCommand) Execute(ctx *CommandContext) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("domain name is required")
	}

	name := ctx.Args[0]
	_, _ = fmt.Fprintf(ctx.Stdout, "Creating new domain: %s\n", name)

	// Create new domain using Jennifer generator
	if err := generators.GenerateDomain(name); err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}

	// Auto-sync domain registry
	_, _ = fmt.Fprintf(ctx.Stdout, "Auto-syncing domain registry...\n")
	if err := generators.GenerateDomainRegistry(); err != nil {
		return fmt.Errorf("failed to sync domain registry: %w", err)
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "✅ Domain '%s' created and synced successfully!\n", name)
	return nil
}

func (c *AddDomainCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "domain",
		Category:    "add",
		Description: "Create a new domain",
		Usage:       "vandor add domain <name>",
		Args:        []string{"name"},
	}
}

func (c *AddDomainCommand) Validate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("domain name is required")
	}
	if args[0] == "" {
		return fmt.Errorf("domain name cannot be empty")
	}
	return nil
}

// AddUsecaseCommand implements the add usecase functionality
type AddUsecaseCommand struct{}

func NewAddUsecaseCommand() *AddUsecaseCommand {
	return &AddUsecaseCommand{}
}

func (c *AddUsecaseCommand) Execute(ctx *CommandContext) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usecase name is required")
	}

	name := ctx.Args[0]
	_, _ = fmt.Fprintf(ctx.Stdout, "Creating new usecase: %s\n", name)

	// Create new usecase using Jennifer generator
	if err := generators.GenerateUsecase(name); err != nil {
		return fmt.Errorf("failed to create usecase: %w", err)
	}

	// Auto-sync usecase registry
	_, _ = fmt.Fprintf(ctx.Stdout, "Auto-syncing usecase registry...\n")
	if err := generators.GenerateUsecaseRegistry(); err != nil {
		return fmt.Errorf("failed to sync usecase registry: %w", err)
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "✅ Usecase '%s' created and synced successfully!\n", name)
	return nil
}

func (c *AddUsecaseCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "usecase",
		Category:    "add",
		Description: "Create a new usecase",
		Usage:       "vandor add usecase <name>",
		Args:        []string{"name"},
	}
}

func (c *AddUsecaseCommand) Validate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usecase name is required")
	}
	if args[0] == "" {
		return fmt.Errorf("usecase name cannot be empty")
	}
	return nil
}

// AddServiceCommand implements the add service functionality
type AddServiceCommand struct{}

func NewAddServiceCommand() *AddServiceCommand {
	return &AddServiceCommand{}
}

func (c *AddServiceCommand) Execute(ctx *CommandContext) error {
	if len(ctx.Args) < 2 {
		return fmt.Errorf("service group and name are required")
	}

	group := ctx.Args[0]
	name := ctx.Args[1]
	_, _ = fmt.Fprintf(ctx.Stdout, "Creating new service: %s in group %s\n", name, group)

	// Create new service using Jennifer generator (only pass the name as that's what the generator expects)
	if err := generators.GenerateService(name); err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Auto-sync service registry
	_, _ = fmt.Fprintf(ctx.Stdout, "Auto-syncing service registry...\n")
	if err := generators.GenerateServiceRegistry(); err != nil {
		return fmt.Errorf("failed to sync service registry: %w", err)
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "✅ Service '%s' created and synced successfully in group '%s'!\n", name, group)
	return nil
}

func (c *AddServiceCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "service",
		Category:    "add",
		Description: "Create a new service",
		Usage:       "vandor add service <group> <name>",
		Args:        []string{"group", "name"},
	}
}

func (c *AddServiceCommand) Validate(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("service group and name are required")
	}
	if args[0] == "" {
		return fmt.Errorf("service group cannot be empty")
	}
	if args[1] == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	return nil
}

// AddJobCommand implements the add job functionality
type AddJobCommand struct{}

func NewAddJobCommand() *AddJobCommand {
	return &AddJobCommand{}
}

func (c *AddJobCommand) Execute(ctx *CommandContext) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("job name is required")
	}

	name := ctx.Args[0]
	_, _ = fmt.Fprintf(ctx.Stdout, "Creating new job: %s\n", name)

	// Create new job using Jennifer generator
	if err := generators.GenerateJob(name); err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	// Auto-sync job registry
	_, _ = fmt.Fprintf(ctx.Stdout, "Auto-syncing job registry...\n")
	if err := generators.GenerateJobRegistry(); err != nil {
		return fmt.Errorf("failed to sync job registry: %w", err)
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "✅ Job '%s' created and synced successfully!\n", name)
	return nil
}

func (c *AddJobCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "job",
		Category:    "add",
		Description: "Create a new background job",
		Usage:       "vandor add job <name>",
		Args:        []string{"name"},
	}
}

func (c *AddJobCommand) Validate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("job name is required")
	}
	if args[0] == "" {
		return fmt.Errorf("job name cannot be empty")
	}
	return nil
}
