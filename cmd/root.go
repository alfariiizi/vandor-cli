package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vandor",
	Short: "Vandor CLI - A code generation and package management tool",
	Long: `Vandor CLI is a powerful tool for managing Go projects with hexagonal architecture.
It provides code generation, package management, and TUI interfaces for streamlined development.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Config initialization will be handled here
}

func er(msg interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}