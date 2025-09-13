package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alfariiizi/vandor-cli/internal/theme"
)

// Screen represents different TUI screens
type Screen string

const (
	ScreenMain      Screen = "main"
	ScreenAdd       Screen = "add"
	ScreenSync      Screen = "sync"
	ScreenTheme     Screen = "theme"
	ScreenVpkg      Screen = "vpkg"
	ScreenInput     Screen = "input"
	ScreenExecution Screen = "execution"
	ScreenResult    Screen = "result"
)

// EnhancedModel represents the enhanced TUI model
type EnhancedModel struct {
	list      list.Model
	textInput textinput.Model
	executor  *CommandExecutor
	screen    Screen
	quitting  bool

	// Command execution state
	selectedCategory string
	selectedCommand  string
	currentArgs      []string
	argIndex         int
	executionResult  ExecutionResult
	resultMessage    string
}

// NewEnhancedModel creates a new enhanced TUI model
func NewEnhancedModel() *EnhancedModel {
	// Initialize text input
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 100
	ti.Width = 50

	model := &EnhancedModel{
		textInput: ti,
		executor:  NewCommandExecutor(),
		screen:    ScreenMain,
		list:      createMainList(),
	}

	return model
}

func (m *EnhancedModel) Init() tea.Cmd {
	return nil
}

func (m *EnhancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		width := msg.Width
		if width < 40 {
			width = 40
		}
		if width > 120 {
			width = 120
		}
		m.list.SetWidth(width)
		return m, nil

	case tea.KeyMsg:
		// Handle special keys first
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			if m.screen != ScreenMain {
				return m.navigateBack(), nil
			}
			m.quitting = true
			return m, tea.Quit
		case "esc":
			return m.navigateBack(), nil
		case "enter":
			return m.handleEnterKey()
		}

		// For list screens, let the list handle other keys (navigation, filtering, etc.)
		if m.screen == ScreenMain || m.screen == ScreenAdd || m.screen == ScreenSync || m.screen == ScreenTheme {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		// For input screen, let the text input handle the keys
		if m.screen == ScreenInput {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

	case ExecutionCompleteMsg:
		m.executionResult = msg.Result
		m.screen = ScreenResult
		if m.executionResult.Success {
			m.resultMessage = m.executionResult.Output
		} else {
			m.resultMessage = fmt.Sprintf("Error: %v", m.executionResult.Error)
		}
		return m, nil
	}

	return m, nil
}

func (m *EnhancedModel) handleEnterKey() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenMain:
		return m.handleMainMenu()
	case ScreenAdd, ScreenSync, ScreenTheme:
		return m.handleCommandSelection()
	case ScreenInput:
		return m.handleInputSubmission()
	case ScreenResult, ScreenExecution:
		return m.navigateToMain(), nil
	}
	return m, nil
}

func (m *EnhancedModel) handleMainMenu() (tea.Model, tea.Cmd) {
	i, ok := m.list.SelectedItem().(item)
	if !ok {
		return m, nil
	}

	choice := string(i)
	switch choice {
	case "Add Component":
		m.screen = ScreenAdd
		m.list = m.createCommandList("add")
	case "Sync Code":
		m.screen = ScreenSync
		m.list = m.createCommandList("sync")
	case "Manage Themes":
		m.screen = ScreenTheme
		m.list = m.createCommandList("theme")
	case "Quit":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *EnhancedModel) handleCommandSelection() (tea.Model, tea.Cmd) {
	i, ok := m.list.SelectedItem().(item)
	if !ok {
		return m, nil
	}

	// Extract command name from the display text
	commandDisplay := string(i)
	parts := strings.Split(commandDisplay, " - ")
	if len(parts) < 1 {
		return m, nil
	}

	commandName := strings.ToLower(parts[0])
	m.selectedCategory = string(m.screen)
	m.selectedCommand = commandName

	// Get command metadata to check required arguments
	cmd, exists := m.executor.registry.Get(m.selectedCategory, m.selectedCommand)
	if !exists {
		return m, nil
	}

	meta := cmd.GetMetadata()

	// If no arguments required, execute immediately
	if len(meta.Args) == 0 {
		m.screen = ScreenExecution
		return m, ExecuteCommandCmd(m.selectedCategory, m.selectedCommand, []string{})
	}

	// Start collecting arguments
	m.currentArgs = make([]string, len(meta.Args))
	m.argIndex = 0
	m.screen = ScreenInput
	m.textInput.Placeholder = fmt.Sprintf("Enter %s...", meta.Args[0])
	m.textInput.SetValue("")
	m.textInput.Focus()

	return m, nil
}

func (m *EnhancedModel) handleInputSubmission() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.textInput.Value())
	if value == "" {
		return m, nil // Don't submit empty values
	}

	// Store the current argument
	m.currentArgs[m.argIndex] = value
	m.argIndex++

	// Get command metadata to check if we need more arguments
	cmd, exists := m.executor.registry.Get(m.selectedCategory, m.selectedCommand)
	if !exists {
		return m.navigateToMain(), nil
	}

	meta := cmd.GetMetadata()

	// If we have all arguments, execute the command
	if m.argIndex >= len(meta.Args) {
		m.screen = ScreenExecution
		return m, ExecuteCommandCmd(m.selectedCategory, m.selectedCommand, m.currentArgs)
	}

	// Prepare for next argument
	m.textInput.Placeholder = fmt.Sprintf("Enter %s...", meta.Args[m.argIndex])
	m.textInput.SetValue("")

	return m, nil
}

func (m *EnhancedModel) createCommandList(category string) list.Model {
	commands := m.executor.registry.List(category)
	items := make([]list.Item, len(commands))

	for i, cmd := range commands {
		items[i] = item(FormatCommandForDisplay(cmd))
	}

	styles := getStyles()
	const defaultWidth = 60
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = fmt.Sprintf("Available %s commands:", category)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styles.Title
	l.Styles.PaginationStyle = styles.Pagination
	l.Styles.HelpStyle = styles.Help

	return l
}

func (m *EnhancedModel) navigateBack() *EnhancedModel {
	switch m.screen {
	case ScreenAdd, ScreenSync, ScreenTheme, ScreenVpkg:
		m.screen = ScreenMain
		m.list = createMainList()
	case ScreenInput, ScreenExecution, ScreenResult:
		// Go back to the command list for the current category
		m.screen = Screen(m.selectedCategory)
		m.list = m.createCommandList(m.selectedCategory)
		m.textInput.Blur()
	default:
		m.screen = ScreenMain
		m.list = createMainList()
	}
	return m
}

func (m *EnhancedModel) navigateToMain() *EnhancedModel {
	m.screen = ScreenMain
	m.list = createMainList()
	m.textInput.Blur()
	m.selectedCategory = ""
	m.selectedCommand = ""
	m.currentArgs = nil
	m.argIndex = 0
	return m
}

func (m *EnhancedModel) View() string {
	styles := getStyles()

	if m.quitting {
		return styles.Quit.Render("Goodbye!")
	}

	switch m.screen {
	case ScreenInput:
		return m.renderInputScreen(styles)
	case ScreenExecution:
		return m.renderExecutionScreen(styles)
	case ScreenResult:
		return m.renderResultScreen(styles)
	default:
		return m.renderListScreen(styles)
	}
}

func (m *EnhancedModel) renderListScreen(styles *theme.Styles) string {
	var title string
	var helpText string

	switch m.screen {
	case ScreenAdd:
		title = "🔧 Add Components"
		helpText = "↑/↓: navigate • enter: select • esc: back • q: quit"
	case ScreenSync:
		title = "⚙️ Sync Code"
		helpText = "↑/↓: navigate • enter: select • esc: back • q: quit"
	case ScreenTheme:
		title = "🎨 Manage Themes"
		helpText = "↑/↓: navigate • enter: select • esc: back • q: quit"
	case ScreenVpkg:
		title = "📦 Manage Packages"
		helpText = "↑/↓: navigate • enter: select • esc: back • q: quit"
	default:
		title = "🚀 Vandor CLI"
		helpText = "↑/↓: navigate • enter: select • q: quit"
	}

	header := styles.Title.Render(title)
	help := styles.Help.Render(helpText)

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", header, m.list.View(), help)
}

func (m *EnhancedModel) renderInputScreen(styles *theme.Styles) string {
	// Get command metadata for context
	cmd, exists := m.executor.registry.Get(m.selectedCategory, m.selectedCommand)
	if !exists {
		return "Error: Command not found"
	}

	meta := cmd.GetMetadata()
	title := fmt.Sprintf("📝 %s", meta.Description)

	progress := fmt.Sprintf("Argument %d of %d", m.argIndex+1, len(meta.Args))

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n\n%s\n\n%s",
		styles.Title.Render(title),
		styles.Info.Render(progress),
		styles.Info.Render(fmt.Sprintf("Required: %s", meta.Args[m.argIndex])),
		m.textInput.View(),
		styles.Help.Render("Press 'enter' to continue, 'esc' to cancel"))
}

func (m *EnhancedModel) renderExecutionScreen(styles *theme.Styles) string {
	title := "⏳ Executing Command..."
	command := fmt.Sprintf("%s %s %s", m.selectedCategory, m.selectedCommand, strings.Join(m.currentArgs, " "))

	return fmt.Sprintf("\n%s\n\n%s\n\n%s",
		styles.Title.Render(title),
		styles.Info.Render(command),
		styles.Help.Render("Please wait..."))
}

func (m *EnhancedModel) renderResultScreen(styles *theme.Styles) string {
	var title string
	var messageStyle lipgloss.Style

	if m.executionResult.Success {
		title = "✅ Command Completed Successfully"
		messageStyle = styles.Success
	} else {
		title = "❌ Command Failed"
		messageStyle = styles.Error
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s",
		styles.Title.Render(title),
		messageStyle.Render(m.resultMessage),
		styles.Help.Render("Press 'enter' or 'esc' to continue"))
}

func createMainList() list.Model {
	items := []list.Item{
		item("Add Component"),
		item("Sync Code"),
		item("Manage Themes"),
		item("Quit"),
	}

	styles := getStyles()
	const defaultWidth = 60
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "What would you like to do?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styles.Title
	l.Styles.PaginationStyle = styles.Pagination
	l.Styles.HelpStyle = styles.Help

	return l
}

// EnhancedApp represents the enhanced TUI application
type EnhancedApp struct{}

// NewEnhancedApp creates a new enhanced TUI application
func NewEnhancedApp() *EnhancedApp {
	return &EnhancedApp{}
}

// Run starts the enhanced TUI application
func (a *EnhancedApp) Run() error {
	// Commands are already registered in main.go, no need to register again
	model := NewEnhancedModel()
	if _, err := tea.NewProgram(model).Run(); err != nil {
		return fmt.Errorf("error running enhanced TUI: %w", err)
	}

	return nil
}
