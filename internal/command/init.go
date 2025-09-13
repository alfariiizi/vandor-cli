package command

// RegisterAllCommands registers all available commands with the global registry
func RegisterAllCommands() error {
	registry := GetGlobalRegistry()

	// Add commands
	commands := []Command{
		// Add category
		NewAddDomainCommand(),
		NewAddUsecaseCommand(),
		NewAddServiceCommand(),
		NewAddJobCommand(),

		// Sync category
		NewSyncAllCommand(),
		NewSyncCoreCommand(),
		NewSyncDomainCommand(),
		NewSyncUsecaseCommand(),
		NewSyncServiceCommand(),
		NewSyncDbModelCommand(),

		// Theme category
		NewThemeListCommand(),
		NewThemeSetCommand(),
		NewThemeInfoCommand(),
	}

	// Register all commands
	for _, cmd := range commands {
		if err := registry.Register(cmd); err != nil {
			return err
		}
	}

	return nil
}
