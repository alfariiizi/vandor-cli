package main

import (
	"fmt"
	"os"

	"github.com/alfariiizi/vandor-cli/cmd"
	"github.com/alfariiizi/vandor-cli/internal/command"
)

// Build-time variables (injected via ldflags)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Register all unified commands
	if err := command.RegisterAllCommands(); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering commands: %v\n", err)
		os.Exit(1)
	}

	// Set version info for cmd package
	cmd.SetVersionInfo(version, commit, date)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
