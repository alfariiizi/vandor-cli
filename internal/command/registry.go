package command

import (
	"fmt"
	"sync"
)

// DefaultRegistry implements the Registry interface
type DefaultRegistry struct {
	commands map[string]map[string]Command
	mu       sync.RWMutex
}

// NewRegistry creates a new command registry
func NewRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		commands: make(map[string]map[string]Command),
	}
}

// Register adds a command to the registry
func (r *DefaultRegistry) Register(cmd Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	meta := cmd.GetMetadata()
	if meta.Category == "" {
		return fmt.Errorf("command category cannot be empty")
	}
	if meta.Name == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	if r.commands[meta.Category] == nil {
		r.commands[meta.Category] = make(map[string]Command)
	}

	if _, exists := r.commands[meta.Category][meta.Name]; exists {
		return fmt.Errorf("command %s/%s already registered", meta.Category, meta.Name)
	}

	r.commands[meta.Category][meta.Name] = cmd
	return nil
}

// Get retrieves a command by category and name
func (r *DefaultRegistry) Get(category, name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if categoryMap, exists := r.commands[category]; exists {
		if cmd, exists := categoryMap[name]; exists {
			return cmd, true
		}
	}
	return nil, false
}

// List returns all commands in a category
func (r *DefaultRegistry) List(category string) []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var commands []Command
	if categoryMap, exists := r.commands[category]; exists {
		for _, cmd := range categoryMap {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// GetAll returns all registered commands grouped by category
func (r *DefaultRegistry) GetAll() map[string][]Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]Command)
	for category, categoryMap := range r.commands {
		var commands []Command
		for _, cmd := range categoryMap {
			commands = append(commands, cmd)
		}
		result[category] = commands
	}
	return result
}

// Global registry instance
var globalRegistry = NewRegistry()

// GetGlobalRegistry returns the global command registry
func GetGlobalRegistry() Registry {
	return globalRegistry
}
