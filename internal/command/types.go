package command

import (
	"context"
	"io"
	"os"
)

// CommandContext provides execution context for commands
type CommandContext struct {
	Ctx    context.Context
	Args   []string
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

// NewCommandContext creates a new command context with default values
func NewCommandContext(args []string) *CommandContext {
	return &CommandContext{
		Ctx:    context.Background(),
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}
}

// Command represents a unified command that can be executed by both CLI and TUI
type Command interface {
	// Execute runs the command with the given context
	Execute(ctx *CommandContext) error

	// Metadata returns command information
	GetMetadata() CommandMetadata

	// Validate checks if the arguments are valid for this command
	Validate(args []string) error
}

// CommandMetadata contains information about the command
type CommandMetadata struct {
	Name        string   // e.g., "domain", "sync-all"
	Category    string   // e.g., "add", "sync", "vpkg", "theme"
	Description string   // Short description
	Usage       string   // Usage example
	Args        []string // Expected arguments
	Flags       []Flag   // Available flags
}

// Flag represents a command flag
type Flag struct {
	Name        string
	Short       string
	Description string
	Type        string // "string", "bool", "stringSlice"
	Default     interface{}
}

// Result represents the result of a command execution
type Result struct {
	Success bool
	Message string
	Error   error
	Data    interface{} // Optional data for programmatic use
}

// Registry manages all available commands
type Registry interface {
	// Register adds a command to the registry
	Register(cmd Command) error

	// Get retrieves a command by category and name
	Get(category, name string) (Command, bool)

	// List returns all commands in a category
	List(category string) []Command

	// GetAll returns all registered commands grouped by category
	GetAll() map[string][]Command
}
