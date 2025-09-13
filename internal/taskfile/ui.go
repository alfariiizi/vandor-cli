package taskfile

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/alfariiizi/vandor-cli/internal/theme"
)

func getTaskfileStyles() *theme.Styles {
	return theme.GetCurrentStyles()
}

// TaskItem represents a task in the list
type TaskItem struct {
	name        string
	description string
	task        Task
}

func (i TaskItem) FilterValue() string { return i.name }

func (i TaskItem) Title() string { return i.name }
func (i TaskItem) Description() string {
	if i.description != "" {
		return i.description
	}
	if i.task.Desc != "" {
		return i.task.Desc
	}
	if i.task.Summary != "" {
		return i.task.Summary
	}
	return "No description"
}

// TaskSelector represents the state for task selection
type TaskSelector struct {
	list         list.Model
	choice       string
	quitting     bool
	selectedTask *Task
}

// TaskPrompt represents the state for prompting task variables
type TaskPrompt struct {
	inputs    []textinput.Model
	focused   int
	variables []string
	values    map[string]string
	task      *Task
	quitting  bool
	submitted bool
}

// NewTaskSelector creates a new task selector
func NewTaskSelector(taskfile *TaskfileSchema) *TaskSelector {
	items := make([]list.Item, 0, len(taskfile.Tasks))

	for name, task := range taskfile.Tasks {
		// Skip internal tasks
		if task.Internal {
			continue
		}

		items = append(items, TaskItem{
			name:        name,
			description: task.Desc,
			task:        task,
		})
	}

	const defaultWidth = 80
	const listHeight = 14

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select a task to run"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	styles := getTaskfileStyles()
	l.Styles.Title = styles.Title
	l.Styles.PaginationStyle = styles.Pagination
	l.Styles.HelpStyle = styles.Help

	return &TaskSelector{list: l}
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(TaskItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.name)

	styles := getTaskfileStyles()
	fn := styles.Item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return styles.SelectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	if _, err := fmt.Fprint(w, fn(str)); err != nil {
		// Error writing to output stream - not much we can do here
		return
	}
}

func (m TaskSelector) Init() tea.Cmd {
	return nil
}

func (m TaskSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(TaskItem)
			if ok {
				m.choice = i.name
				m.selectedTask = &i.task
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m TaskSelector) View() string {
	styles := getTaskfileStyles()
	if m.choice != "" {
		return styles.Quit.Render(fmt.Sprintf("Running task: %s", m.choice))
	}
	if m.quitting {
		return styles.Quit.Render("Task selection canceled.")
	}
	return "\n" + m.list.View()
}

// NewTaskPrompt creates a new task variable prompt
func NewTaskPrompt(task *Task, variables []string) *TaskPrompt {
	inputs := make([]textinput.Model, len(variables))

	for i, variable := range variables {
		t := textinput.New()
		t.Placeholder = fmt.Sprintf("Enter value for %s", variable)
		t.Focus()
		t.CharLimit = 256
		t.Width = 50

		if i == 0 {
			t.Focus()
		}

		inputs[i] = t
	}

	return &TaskPrompt{
		inputs:    inputs,
		variables: variables,
		values:    make(map[string]string),
		task:      task,
	}
}

func (m TaskPrompt) Init() tea.Cmd {
	return textinput.Blink
}

func (m TaskPrompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.focused == len(m.inputs)-1 {
				// Submit form
				for i, input := range m.inputs {
					m.values[m.variables[i]] = input.Value()
				}
				m.submitted = true
				return m, tea.Quit
			} else {
				// Move to next input
				m.inputs[m.focused].Blur()
				m.focused++
				m.inputs[m.focused].Focus()
			}
		case "shift+tab":
			if m.focused > 0 {
				m.inputs[m.focused].Blur()
				m.focused--
				m.inputs[m.focused].Focus()
			}
		case "tab":
			if m.focused < len(m.inputs)-1 {
				m.inputs[m.focused].Blur()
				m.focused++
				m.inputs[m.focused].Focus()
			}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m TaskPrompt) View() string {
	var b strings.Builder

	styles := getTaskfileStyles()
	b.WriteString(styles.Title.Render("Task Variables"))
	b.WriteString("\n\n")

	for i, input := range m.inputs {
		variable := m.variables[i]
		b.WriteString(fmt.Sprintf("Variable: %s\n", variable))
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	b.WriteString("Press Enter to continue to next field or submit\n")
	b.WriteString("Press Tab/Shift+Tab to navigate\n")
	b.WriteString("Press Ctrl+C or Esc to cancel\n")

	return b.String()
}

// RunTaskSelector runs the interactive task selector
func RunTaskSelector(taskfile *TaskfileSchema) (*Task, string, error) {
	selector := NewTaskSelector(taskfile)

	p := tea.NewProgram(selector)
	finalModel, err := p.Run()
	if err != nil {
		return nil, "", fmt.Errorf("error running task selector: %w", err)
	}

	if m, ok := finalModel.(TaskSelector); ok {
		if m.quitting && m.choice == "" {
			return nil, "", fmt.Errorf("task selection canceled")
		}
		return m.selectedTask, m.choice, nil
	}

	return nil, "", fmt.Errorf("unexpected model type")
}

// RunTaskPrompt runs the interactive variable prompt
func RunTaskPrompt(task *Task, variables []string) (map[string]string, error) {
	if len(variables) == 0 {
		return make(map[string]string), nil
	}

	prompt := NewTaskPrompt(task, variables)

	p := tea.NewProgram(prompt)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error running task prompt: %w", err)
	}

	if m, ok := finalModel.(TaskPrompt); ok {
		if m.quitting && !m.submitted {
			return nil, fmt.Errorf("variable input canceled")
		}
		return m.values, nil
	}

	return nil, fmt.Errorf("unexpected model type")
}
