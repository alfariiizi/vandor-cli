package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addSchemaCmd = &cobra.Command{
	Use:   "schema [name]",
	Short: "Create a new schema",
	Long:  `Create a new schema using ent-tools.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Printf("Creating new schema: %s\n", name)
		
		if err := runScript("./internal/scripts/ent-tools.sh", "new", name); err != nil {
			er(fmt.Sprintf("Failed to create schema: %v", err))
		}
		
		fmt.Printf("✅ Schema '%s' created successfully!\n", name)
	},
}

func init() {
	addCmd.AddCommand(addSchemaCmd)
}