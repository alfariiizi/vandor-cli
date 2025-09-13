package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alfariiizi/vandor-cli/internal/taskfile"
)

var taskCmd = &cobra.Command{
	Use:   "task [task-name]",
	Short: "Run tasks from Taskfile with interactive prompts",
	Long: `Run tasks from Taskfile.yaml (or similar) with interactive task selection and variable prompting.

Supports various Taskfile naming patterns:
- Taskfile.yaml / Taskfile.yml
- taskfile.yaml / taskfile.yml  
- taskfiles.yaml / taskfiles.yml
- Task.yaml / Task.yml

If no task name is provided, an interactive selector will be shown.
If a task requires variables, you'll be prompted to enter them interactively.`,
	Example: `  # Interactive task selection
  vandor task

  # Run specific task directly
  vandor task build

  # Both modes support variable prompting if needed`,
	Run: func(cmd *cobra.Command, args []string) {
		// Find taskfile
		taskfilePath, err := taskfile.FindTaskfile()
		if err != nil {
			er(fmt.Sprintf("Failed to find taskfile: %v", err))
		}

		fmt.Printf("📋 Found taskfile: %s\n", taskfilePath)

		// Parse taskfile
		tf, err := taskfile.ParseTaskfile(taskfilePath)
		if err != nil {
			er(fmt.Sprintf("Failed to parse taskfile: %v", err))
		}

		var selectedTask *taskfile.Task
		var taskName string

		// Determine task to run
		if len(args) > 0 {
			// Task specified via command line
			taskName = args[0]
			if task, exists := tf.Tasks[taskName]; exists {
				selectedTask = &task
			} else {
				er(fmt.Sprintf("Task '%s' not found in taskfile", taskName))
			}
		} else {
			// Interactive task selection
			fmt.Println("🎯 Select a task to run:")
			var err error
			selectedTask, taskName, err = taskfile.RunTaskSelector(tf)
			if err != nil {
				er(fmt.Sprintf("Task selection failed: %v", err))
			}
		}

		// Get required variables for the task
		requiredVars := selectedTask.GetTaskVars()

		var variables map[string]string
		if len(requiredVars) > 0 {
			fmt.Printf("\n📝 Task '%s' requires variables\n", taskName)
			var err error
			variables, err = taskfile.RunTaskPrompt(selectedTask, requiredVars)
			if err != nil {
				er(fmt.Sprintf("Variable input failed: %v", err))
			}
		} else {
			variables = make(map[string]string)
		}

		// Execute the task
		executor := taskfile.NewTaskExecutor(tf, variables)
		if err := executor.ExecuteTask(taskName, selectedTask); err != nil {
			er(fmt.Sprintf("Task execution failed: %v", err))
		}
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available tasks",
	Long:  `List all available tasks from the Taskfile with their descriptions.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Find taskfile
		taskfilePath, err := taskfile.FindTaskfile()
		if err != nil {
			er(fmt.Sprintf("Failed to find taskfile: %v", err))
		}

		// Parse taskfile
		tf, err := taskfile.ParseTaskfile(taskfilePath)
		if err != nil {
			er(fmt.Sprintf("Failed to parse taskfile: %v", err))
		}

		fmt.Printf("📋 Available tasks in %s:\n\n", taskfilePath)

		if len(tf.Tasks) == 0 {
			fmt.Println("No tasks found.")
			return
		}

		// Find longest task name for formatting
		maxNameLen := 0
		for name, task := range tf.Tasks {
			if !task.Internal && len(name) > maxNameLen {
				maxNameLen = len(name)
			}
		}

		// List tasks
		for name, task := range tf.Tasks {
			if task.Internal {
				continue // Skip internal tasks
			}

			description := task.Desc
			if description == "" {
				description = task.Summary
			}
			if description == "" {
				description = "No description"
			}

			// Format output
			padding := strings.Repeat(" ", maxNameLen-len(name)+2)
			fmt.Printf("  %s%s%s\n", name, padding, description)

			// Show required variables if any
			requiredVars := task.GetTaskVars()
			if len(requiredVars) > 0 {
				fmt.Printf("    Variables: %s\n", strings.Join(requiredVars, ", "))
			}

			// Show aliases if any
			if len(task.Aliases) > 0 {
				fmt.Printf("    Aliases: %s\n", strings.Join(task.Aliases, ", "))
			}

			fmt.Println()
		}
	},
}

var taskInfoCmd = &cobra.Command{
	Use:   "info [task-name]",
	Short: "Show detailed information about a task",
	Long:  `Show detailed information about a specific task including description, commands, variables, and dependencies.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]

		// Find taskfile
		taskfilePath, err := taskfile.FindTaskfile()
		if err != nil {
			er(fmt.Sprintf("Failed to find taskfile: %v", err))
		}

		// Parse taskfile
		tf, err := taskfile.ParseTaskfile(taskfilePath)
		if err != nil {
			er(fmt.Sprintf("Failed to parse taskfile: %v", err))
		}

		// Find task
		task, exists := tf.Tasks[taskName]
		if !exists {
			er(fmt.Sprintf("Task '%s' not found in taskfile", taskName))
		}

		// Display task information
		fmt.Printf("📋 Task Information: %s\n\n", taskName)

		if task.Desc != "" {
			fmt.Printf("Description: %s\n", task.Desc)
		}
		if task.Summary != "" {
			fmt.Printf("Summary: %s\n", task.Summary)
		}

		if len(task.Aliases) > 0 {
			fmt.Printf("Aliases: %s\n", strings.Join(task.Aliases, ", "))
		}

		// Show variables
		requiredVars := task.GetTaskVars()
		if len(requiredVars) > 0 {
			fmt.Printf("Required Variables: %s\n", strings.Join(requiredVars, ", "))
		}

		// Show dependencies
		if len(task.Deps) > 0 {
			fmt.Printf("Dependencies: ")
			var deps []string
			for _, dep := range task.Deps {
				if str, ok := dep.(string); ok {
					deps = append(deps, str)
				}
			}
			fmt.Printf("%s\n", strings.Join(deps, ", "))
		}

		// Show commands
		commands, err := task.GetTaskCommands()
		if err != nil {
			fmt.Printf("Error parsing commands: %v\n", err)
		} else if len(commands) > 0 {
			fmt.Printf("\nCommands:\n")
			for i, cmd := range commands {
				if cmd.Task != "" {
					fmt.Printf("  %d. Run task: %s\n", i+1, cmd.Task)
				} else if cmd.Cmd != "" {
					fmt.Printf("  %d. %s\n", i+1, cmd.Cmd)
				}
			}
		}

		// Show preconditions
		if len(task.Preconditions) > 0 {
			fmt.Printf("\nPreconditions:\n")
			for i, precond := range task.Preconditions {
				fmt.Printf("  %d. %s\n", i+1, precond.Sh)
				if precond.Msg != "" {
					fmt.Printf("     Message: %s\n", precond.Msg)
				}
			}
		}

		// Show platforms
		if len(task.Platforms) > 0 {
			fmt.Printf("Supported Platforms: %s\n", strings.Join(task.Platforms, ", "))
		}

		// Show flags
		var flags []string
		if task.Internal {
			flags = append(flags, "internal")
		}
		if task.Silent {
			flags = append(flags, "silent")
		}
		if task.Interactive {
			flags = append(flags, "interactive")
		}
		if len(flags) > 0 {
			fmt.Printf("Flags: %s\n", strings.Join(flags, ", "))
		}
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)

	// Add subcommands
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskInfoCmd)
}
