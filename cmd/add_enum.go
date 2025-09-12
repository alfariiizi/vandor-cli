package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var addEnumCmd = &cobra.Command{
	Use:   "enum [name]",
	Short: "Create a new enum",
	Long:  `Create a new enum type.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Printf("Creating new enum: %s\n", name)

		// Convert to lowercase as per the original task
		enumName := strings.ToLower(name)

		if err := runGoCommand("run", "./internal/cmd/enum/cmd/main.go", "add", enumName); err != nil {
			er(fmt.Sprintf("Failed to create enum: %v", err))
		}

		fmt.Printf("✅ Enum '%s' created successfully!\n", name)
	},
}

func init() {
	addCmd.AddCommand(addEnumCmd)
}
