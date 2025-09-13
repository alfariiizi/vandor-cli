package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alfariiizi/vandor-cli/internal/command"
)

// CommandLauncher represents a TUI launcher with pre-selected command
type CommandLauncher struct {
	category string
	name     string
}

// NewCommandLauncher creates a new command launcher for specific command
func NewCommandLauncher(category, name string) *CommandLauncher {
	return &CommandLauncher{
		category: category,
		name:     name,
	}
}

// Run starts the TUI with a pre-selected command
func (cl *CommandLauncher) Run() error {
	model := NewEnhancedModel()

	// Pre-navigate to the command
	if cl.category != "" && cl.name != "" {
		// Validate command exists
		registry := command.GetGlobalRegistry()
		cmd, exists := registry.Get(cl.category, cl.name)
		if !exists {
			return fmt.Errorf("command %s/%s not found", cl.category, cl.name)
		}

		meta := cmd.GetMetadata()

		// Set up the model for direct command execution
		model.selectedCategory = cl.category
		model.selectedCommand = cl.name

		// If command requires arguments, go to input screen
		if len(meta.Args) > 0 {
			model.currentArgs = make([]string, len(meta.Args))
			model.argIndex = 0
			model.screen = ScreenInput
			model.textInput.Placeholder = fmt.Sprintf("Enter %s...", meta.Args[0])
			model.textInput.SetValue("")
			model.textInput.Focus()
		} else {
			// Execute immediately if no arguments needed
			model.screen = ScreenExecution
			// Execute the command right away
			executor := NewCommandExecutor()
			result := executor.ExecuteCommand(cl.category, cl.name, []string{})
			model.executionResult = result
			model.screen = ScreenResult
			if result.Success {
				model.resultMessage = result.Output
			} else {
				model.resultMessage = fmt.Sprintf("Error: %v", result.Error)
			}
		}
	}

	if _, err := tea.NewProgram(model).Run(); err != nil {
		return fmt.Errorf("error running command launcher TUI: %w", err)
	}

	return nil
}

// LaunchDirectCommand launches TUI directly for a specific command
func LaunchDirectCommand(category, name string) error {
	launcher := NewCommandLauncher(category, name)
	return launcher.Run()
}
