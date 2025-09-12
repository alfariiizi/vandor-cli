package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new components to your Vandor project",
	Long:  `Add various components like schemas, domains, usecases, services, jobs, etc. to your Vandor project.`,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

// Helper function to run Go commands
func runGoCommand(args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Helper function to run shell scripts
func runScript(script string, args ...string) error {
	fullArgs := append([]string{script}, args...)
	cmd := exec.Command("bash", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

