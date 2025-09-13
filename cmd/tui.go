package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch Vandor TUI (Terminal User Interface)",
	Long: `Launch the interactive Terminal User Interface for Vandor.
This provides a user-friendly interface to manage your project without remembering commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		app := tui.NewApp()
		if err := app.Run(); err != nil {
			er(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
