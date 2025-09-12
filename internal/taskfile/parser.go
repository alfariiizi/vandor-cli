package taskfile

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// TaskfileSchema represents the structure of a Taskfile
type TaskfileSchema struct {
	Version   string                 `yaml:"version"`
	Vars      map[string]interface{} `yaml:"vars,omitempty"`
	Env       map[string]interface{} `yaml:"env,omitempty"`
	Tasks     map[string]Task        `yaml:"tasks"`
	Includes  map[string]interface{} `yaml:"includes,omitempty"`
	Output    string                 `yaml:"output,omitempty"`
	Method    string                 `yaml:"method,omitempty"`
	Run       string                 `yaml:"run,omitempty"`
	Interval  string                 `yaml:"interval,omitempty"`
	Watch     bool                   `yaml:"watch,omitempty"`
	Dotenv    []string               `yaml:"dotenv,omitempty"`
	Silent    bool                   `yaml:"silent,omitempty"`
}

// Task represents a single task in the Taskfile
type Task struct {
	Desc         string                 `yaml:"desc,omitempty"`
	Summary      string                 `yaml:"summary,omitempty"`
	Aliases      []string               `yaml:"aliases,omitempty"`
	Cmds         []interface{}          `yaml:"cmds,omitempty"`
	Cmd          string                 `yaml:"cmd,omitempty"` // For simple string tasks
	Deps         []interface{}          `yaml:"deps,omitempty"`
	Preconditions []Precondition        `yaml:"preconditions,omitempty"`
	Requires     RequiredVars          `yaml:"requires,omitempty"`
	Vars         map[string]interface{} `yaml:"vars,omitempty"`
	Env          map[string]interface{} `yaml:"env,omitempty"`
	Dir          string                 `yaml:"dir,omitempty"`
	Sources      []string               `yaml:"sources,omitempty"`
	Generates    []string               `yaml:"generates,omitempty"`
	Status       []string               `yaml:"status,omitempty"`
	Method       string                 `yaml:"method,omitempty"`
	Prefix       string                 `yaml:"prefix,omitempty"`
	IgnoreError  bool                   `yaml:"ignore_error,omitempty"`
	Silent       bool                   `yaml:"silent,omitempty"`
	Interactive  bool                   `yaml:"interactive,omitempty"`
	Internal     bool                   `yaml:"internal,omitempty"`
	Platforms    []string               `yaml:"platforms,omitempty"`
	Label        string                 `yaml:"label,omitempty"`
	Prompt       string                 `yaml:"prompt,omitempty"`
	Run          string                 `yaml:"run,omitempty"`
	Watch        bool                   `yaml:"watch,omitempty"`
	Dotenv       []string               `yaml:"dotenv,omitempty"`
}

// Precondition represents a task precondition
type Precondition struct {
	Sh  string `yaml:"sh,omitempty"`
	Msg string `yaml:"msg,omitempty"`
}

// RequiredVars represents variables required by a task
type RequiredVars struct {
	Vars []string `yaml:"vars,omitempty"`
}

// Command represents a single command that can be a string or object
type Command struct {
	Cmd         string                 `yaml:"cmd,omitempty"`
	Silent      bool                   `yaml:"silent,omitempty"`
	Task        string                 `yaml:"task,omitempty"`
	Vars        map[string]interface{} `yaml:"vars,omitempty"`
	Ignore      bool                   `yaml:"ignore_error,omitempty"`
	Defer       bool                   `yaml:"defer,omitempty"`
	Platforms   []string               `yaml:"platforms,omitempty"`
	Dir         string                 `yaml:"dir,omitempty"`
}

// Variable represents a variable that can be a simple value or complex object
type Variable struct {
	Static   string   `yaml:"-"`
	Sh       string   `yaml:"sh,omitempty"`
	Ref      string   `yaml:"ref,omitempty"`
	Dir      string   `yaml:"dir,omitempty"`
}

// FindTaskfile searches for taskfiles in common naming patterns
func FindTaskfile() (string, error) {
	possibleNames := []string{
		"Taskfile.yaml",
		"Taskfile.yml", 
		"taskfile.yaml",
		"taskfile.yml",
		"taskfiles.yaml",
		"taskfiles.yml",
		"Task.yaml",
		"Task.yml",
	}

	for _, name := range possibleNames {
		if _, err := os.Stat(name); err == nil {
			return name, nil
		}
	}

	return "", fmt.Errorf("no taskfile found (checked: %s)", strings.Join(possibleNames, ", "))
}

// ParseTaskfile parses a Taskfile from the given path
func ParseTaskfile(path string) (*TaskfileSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read taskfile: %w", err)
	}

	var taskfile TaskfileSchema
	if err := yaml.Unmarshal(data, &taskfile); err != nil {
		return nil, fmt.Errorf("failed to parse taskfile: %w", err)
	}

	// Normalize tasks - handle both string and object formats
	for name, task := range taskfile.Tasks {
		// If the task has no Cmds but has Cmd, convert it
		if len(task.Cmds) == 0 && task.Cmd != "" {
			task.Cmds = []interface{}{task.Cmd}
			task.Cmd = ""
			taskfile.Tasks[name] = task
		}
	}

	return &taskfile, nil
}

// GetTaskCommands extracts commands from a task, handling both string and object formats
func (t *Task) GetTaskCommands() ([]Command, error) {
	var commands []Command

	for _, cmd := range t.Cmds {
		switch v := cmd.(type) {
		case string:
			commands = append(commands, Command{Cmd: v})
		case map[string]interface{}:
			var command Command
			if cmdStr, ok := v["cmd"].(string); ok {
				command.Cmd = cmdStr
			}
			if silent, ok := v["silent"].(bool); ok {
				command.Silent = silent
			}
			if task, ok := v["task"].(string); ok {
				command.Task = task
			}
			if ignore, ok := v["ignore_error"].(bool); ok {
				command.Ignore = ignore
			}
			if defer_, ok := v["defer"].(bool); ok {
				command.Defer = defer_
			}
			if dir, ok := v["dir"].(string); ok {
				command.Dir = dir
			}
			if vars, ok := v["vars"].(map[string]interface{}); ok {
				command.Vars = vars
			}
			if platforms, ok := v["platforms"].([]interface{}); ok {
				for _, p := range platforms {
					if platform, ok := p.(string); ok {
						command.Platforms = append(command.Platforms, platform)
					}
				}
			}
			commands = append(commands, command)
		default:
			return nil, fmt.Errorf("unsupported command type: %T", v)
		}
	}

	return commands, nil
}

// GetRequiredVars extracts required variables from a task
func (t *Task) GetRequiredVars() []string {
	return t.Requires.Vars
}

// GetTaskVars extracts variables that should be prompted for
func (t *Task) GetTaskVars() []string {
	var vars []string
	
	// Add required vars
	vars = append(vars, t.GetRequiredVars()...)
	
	// Parse commands for template variables
	commands, _ := t.GetTaskCommands()
	for _, cmd := range commands {
		// Simple regex to find {{.VAR}} patterns
		// This is a basic implementation - could be enhanced
		if strings.Contains(cmd.Cmd, "{{") {
			// Extract variables from templates
			vars = append(vars, extractTemplateVars(cmd.Cmd)...)
		}
	}
	
	// Remove duplicates
	return uniqueStrings(vars)
}

// extractTemplateVars extracts variable names from template strings like {{.VAR}}
func extractTemplateVars(input string) []string {
	var vars []string
	parts := strings.Split(input, "{{")
	for _, part := range parts[1:] {
		if idx := strings.Index(part, "}}"); idx != -1 {
			varExpr := strings.TrimSpace(part[:idx])
			if strings.HasPrefix(varExpr, ".") {
				varName := strings.TrimPrefix(varExpr, ".")
				// Skip special variables
				if !isSpecialVariable(varName) {
					vars = append(vars, varName)
				}
			}
		}
	}
	return vars
}

// isSpecialVariable checks if a variable is a built-in Task variable
func isSpecialVariable(varName string) bool {
	specialVars := []string{
		"ROOT_DIR", "TASKFILE_DIR", "USER_WORKING_DIR", 
		"CLI_ARGS", "CLI_ARGS_LIST", "CLI_FORCE",
		"TASK", "CHECKSUM", "TIMESTAMP",
	}
	
	for _, special := range specialVars {
		if varName == special {
			return true
		}
	}
	return false
}

// uniqueStrings removes duplicate strings from a slice
func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, str := range input {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	
	return result
}