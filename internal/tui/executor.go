package tui

import (
	"bytes"
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alfariiizi/vandor-cli/internal/command"
)

// ExecutionResult represents the result of a command execution
type ExecutionResult struct {
	Success bool
	Output  string
	Error   error
}

// CommandExecutor handles command execution within the TUI
type CommandExecutor struct {
	registry command.Registry
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{
		registry: command.GetGlobalRegistry(),
	}
}

// ExecuteCommand executes a command and returns the result
func (e *CommandExecutor) ExecuteCommand(category, name string, args []string) ExecutionResult {
	cmd, exists := e.registry.Get(category, name)
	if !exists {
		return ExecutionResult{
			Success: false,
			Output:  "",
			Error:   fmt.Errorf("command %s/%s not found", category, name),
		}
	}

	// Validate arguments
	if err := cmd.Validate(args); err != nil {
		return ExecutionResult{
			Success: false,
			Output:  "",
			Error:   fmt.Errorf("validation error: %w", err),
		}
	}

	// Create buffers to capture output
	var stdout, stderr bytes.Buffer

	// Create command context
	ctx := &command.CommandContext{
		Ctx:    context.Background(),
		Args:   args,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	// Execute the command
	err := cmd.Execute(ctx)

	// Combine stdout and stderr
	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	return ExecutionResult{
		Success: err == nil,
		Output:  output,
		Error:   err,
	}
}

// GetAvailableCommands returns all available commands grouped by category
func (e *CommandExecutor) GetAvailableCommands() map[string][]command.Command {
	return e.registry.GetAll()
}

// FormatCommandForDisplay formats a command for display in the TUI
func FormatCommandForDisplay(cmd command.Command) string {
	meta := cmd.GetMetadata()
	// Capitalize first letter manually to avoid deprecated strings.Title
	name := meta.Name
	if len(name) > 0 {
		name = string(name[0]-32) + name[1:] // Convert first char to uppercase
	}
	return fmt.Sprintf("%s - %s", name, meta.Description)
}

// ExecuteCommandMsg is a Bubble Tea message for command execution
type ExecuteCommandMsg struct {
	Category string
	Name     string
	Args     []string
}

// ExecutionCompleteMsg is sent when command execution is complete
type ExecutionCompleteMsg struct {
	Result ExecutionResult
}

// ExecuteCommandCmd returns a Bubble Tea command that executes a unified command
func ExecuteCommandCmd(category, name string, args []string) tea.Cmd {
	return func() tea.Msg {
		executor := NewCommandExecutor()
		result := executor.ExecuteCommand(category, name, args)
		return ExecutionCompleteMsg{Result: result}
	}
}
