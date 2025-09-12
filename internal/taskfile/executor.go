package taskfile

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/charmbracelet/lipgloss"
)

var (
	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("32")).
			Bold(true)

	outputStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("240"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
)

// TaskExecutor handles running tasks with variable substitution
type TaskExecutor struct {
	taskfile  *TaskfileSchema
	variables map[string]string
	workDir   string
}

// NewTaskExecutor creates a new task executor
func NewTaskExecutor(taskfile *TaskfileSchema, variables map[string]string) *TaskExecutor {
	workDir, _ := os.Getwd()
	
	return &TaskExecutor{
		taskfile:  taskfile,
		variables: variables,
		workDir:   workDir,
	}
}

// ExecuteTask runs a task with the given variables
func (e *TaskExecutor) ExecuteTask(taskName string, task *Task) error {
	fmt.Printf("\n%s\n", commandStyle.Render(fmt.Sprintf("🚀 Running task: %s", taskName)))
	
	if task.Desc != "" {
		fmt.Printf("%s\n", outputStyle.Render(fmt.Sprintf("Description: %s", task.Desc)))
	}
	
	// Get task commands
	commands, err := task.GetTaskCommands()
	if err != nil {
		return fmt.Errorf("failed to parse task commands: %w", err)
	}

	// Check preconditions
	if err := e.checkPreconditions(task); err != nil {
		return fmt.Errorf("precondition failed: %w", err)
	}

	// Execute dependencies first
	if err := e.executeDependencies(task); err != nil {
		return fmt.Errorf("dependency execution failed: %w", err)
	}

	// Build template context
	context := e.buildTemplateContext(task)

	// Execute commands
	for i, cmd := range commands {
		if cmd.Task != "" {
			// Execute another task
			if err := e.executeSubTask(cmd.Task, cmd.Vars); err != nil {
				if !cmd.Ignore {
					return fmt.Errorf("subtask '%s' failed: %w", cmd.Task, err)
				}
				fmt.Printf("%s\n", errorStyle.Render(fmt.Sprintf("Warning: subtask '%s' failed but was ignored", cmd.Task)))
			}
			continue
		}

		if cmd.Cmd == "" {
			continue
		}

		// Substitute variables in command
		expandedCmd, err := e.expandTemplate(cmd.Cmd, context)
		if err != nil {
			return fmt.Errorf("failed to expand command template: %w", err)
		}

		// Show command being executed (unless silent)
		if !cmd.Silent && !task.Silent && !e.taskfile.Silent {
			fmt.Printf("\n%s\n", commandStyle.Render(fmt.Sprintf("$ %s", expandedCmd)))
		}

		// Execute command
		if err := e.executeCommand(expandedCmd, cmd, task); err != nil {
			if !cmd.Ignore {
				return fmt.Errorf("command %d failed: %w", i+1, err)
			}
			fmt.Printf("%s\n", errorStyle.Render(fmt.Sprintf("Warning: command failed but was ignored: %s", expandedCmd)))
		}
	}

	fmt.Printf("\n%s\n", successStyle.Render(fmt.Sprintf("✅ Task '%s' completed successfully", taskName)))
	return nil
}

// checkPreconditions verifies task preconditions
func (e *TaskExecutor) checkPreconditions(task *Task) error {
	for _, precondition := range task.Preconditions {
		if precondition.Sh == "" {
			continue
		}

		context := e.buildTemplateContext(task)
		expandedCmd, err := e.expandTemplate(precondition.Sh, context)
		if err != nil {
			return fmt.Errorf("failed to expand precondition: %w", err)
		}

		cmd := exec.Command("sh", "-c", expandedCmd)
		if err := cmd.Run(); err != nil {
			msg := precondition.Msg
			if msg == "" {
				msg = fmt.Sprintf("precondition failed: %s", expandedCmd)
			}
			return fmt.Errorf(msg)
		}
	}
	return nil
}

// executeDependencies runs task dependencies
func (e *TaskExecutor) executeDependencies(task *Task) error {
	for _, dep := range task.Deps {
		var depTaskName string
		var depVars map[string]interface{}

		switch v := dep.(type) {
		case string:
			depTaskName = v
		case map[string]interface{}:
			if task, ok := v["task"].(string); ok {
				depTaskName = task
			}
			if vars, ok := v["vars"].(map[string]interface{}); ok {
				depVars = vars
			}
		default:
			return fmt.Errorf("unsupported dependency type: %T", v)
		}

		if _, exists := e.taskfile.Tasks[depTaskName]; exists {
			if err := e.executeSubTask(depTaskName, depVars); err != nil {
				return fmt.Errorf("dependency '%s' failed: %w", depTaskName, err)
			}
		} else {
			return fmt.Errorf("dependency task '%s' not found", depTaskName)
		}
	}
	return nil
}

// executeSubTask executes a subtask with optional variable overrides
func (e *TaskExecutor) executeSubTask(taskName string, vars map[string]interface{}) error {
	task, exists := e.taskfile.Tasks[taskName]
	if !exists {
		return fmt.Errorf("task '%s' not found", taskName)
	}

	// Create new executor with merged variables
	subVars := make(map[string]string)
	for k, v := range e.variables {
		subVars[k] = v
	}
	
	// Override with subtask vars
	for k, v := range vars {
		if str, ok := v.(string); ok {
			subVars[k] = str
		} else {
			subVars[k] = fmt.Sprintf("%v", v)
		}
	}

	subExecutor := NewTaskExecutor(e.taskfile, subVars)
	return subExecutor.ExecuteTask(taskName, &task)
}

// executeCommand runs a single command
func (e *TaskExecutor) executeCommand(cmdStr string, cmd Command, task *Task) error {
	// Determine working directory
	workDir := e.workDir
	if cmd.Dir != "" {
		workDir = cmd.Dir
	} else if task.Dir != "" {
		workDir = task.Dir
	}

	// Check platform compatibility
	if len(cmd.Platforms) > 0 {
		currentPlatform := runtime.GOOS
		supported := false
		for _, platform := range cmd.Platforms {
			if platform == currentPlatform {
				supported = true
				break
			}
		}
		if !supported {
			fmt.Printf("%s\n", outputStyle.Render(fmt.Sprintf("Skipping command (platform %s not supported)", currentPlatform)))
			return nil
		}
	}

	// Create command
	var execCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		execCmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		execCmd = exec.Command("sh", "-c", cmdStr)
	}

	execCmd.Dir = workDir

	// Set environment variables
	execCmd.Env = os.Environ()
	for k, v := range e.buildEnvironmentVars(task) {
		execCmd.Env = append(execCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Handle interactive tasks
	if task.Interactive {
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		return execCmd.Run()
	}

	// Capture output for non-interactive tasks
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Stream output in real-time
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if !cmd.Silent && !task.Silent && !e.taskfile.Silent {
				fmt.Printf("%s\n", outputStyle.Render(scanner.Text()))
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if !cmd.Silent && !task.Silent && !e.taskfile.Silent {
				fmt.Printf("%s\n", errorStyle.Render(scanner.Text()))
			}
		}
	}()

	return execCmd.Wait()
}

// buildTemplateContext creates template context with all available variables
func (e *TaskExecutor) buildTemplateContext(task *Task) map[string]interface{} {
	context := make(map[string]interface{})

	// Add global variables
	for k, v := range e.taskfile.Vars {
		context[k] = v
	}

	// Add task variables
	for k, v := range task.Vars {
		context[k] = v
	}

	// Add user-provided variables
	for k, v := range e.variables {
		context[k] = v
	}

	// Add special Task variables
	rootDir, _ := filepath.Abs(".")
	context["ROOT_DIR"] = rootDir
	context["TASKFILE_DIR"] = filepath.Dir("Taskfile.yaml") // This could be improved
	context["USER_WORKING_DIR"] = e.workDir

	return context
}

// buildEnvironmentVars builds environment variables for command execution
func (e *TaskExecutor) buildEnvironmentVars(task *Task) map[string]string {
	env := make(map[string]string)

	// Add global env vars
	for k, v := range e.taskfile.Env {
		if str, ok := v.(string); ok {
			env[k] = str
		} else {
			env[k] = fmt.Sprintf("%v", v)
		}
	}

	// Add task env vars
	for k, v := range task.Env {
		if str, ok := v.(string); ok {
			env[k] = str
		} else {
			env[k] = fmt.Sprintf("%v", v)
		}
	}

	return env
}

// expandTemplate expands Go template syntax in strings
func (e *TaskExecutor) expandTemplate(input string, context map[string]interface{}) (string, error) {
	tmpl, err := template.New("task").Parse(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, context); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}