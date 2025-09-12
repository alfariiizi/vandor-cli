package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/alfariiizi/vandor-cli/internal/theme"
)

const listHeight = 14

func getStyles() *theme.Styles {
	return theme.GetCurrentStyles()
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	styles := getStyles()
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

type model struct {
	list     list.Model
	choice   string
	quitting bool
	screen   string // "main", "add", "generate", "vpkg"
	input    textinput.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Set responsive width with reasonable bounds
		width := msg.Width
		if width < 40 {
			width = 40 // minimum width for readability
		}
		if width > 120 {
			width = 120 // maximum width to prevent overly wide display
		}
		m.list.SetWidth(width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				return m.handleMenuChoice()
			}

		case "esc":
			if m.screen != "main" {
				m.screen = "main"
				m.list = m.createMainMenu()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) handleMenuChoice() (tea.Model, tea.Cmd) {
	switch m.choice {
	case "Initialize Project":
		return m, tea.Sequence(
			tea.Printf("🚀 Initializing project..."),
			tea.Quit,
		)
	case "Add Component":
		m.screen = "add"
		m.list = m.createAddMenu()
		return m, nil
	case "Generate Code":
		m.screen = "generate"
		m.list = m.createGenerateMenu()
		return m, nil
	case "Manage Packages":
		m.screen = "vpkg"
		m.list = m.createVpkgMenu()
		return m, nil
	case "Quit":
		m.quitting = true
		return m, tea.Quit

	// Add menu items
	case "Add Schema":
		return m, tea.Sequence(
			tea.Printf("📋 Adding schema... (Use: vandor add schema <name>)"),
			tea.Quit,
		)
	case "Add Domain":
		return m, tea.Sequence(
			tea.Printf("🏛️ Adding domain... (Use: vandor add domain <name>)"),
			tea.Quit,
		)
	case "Add Usecase":
		return m, tea.Sequence(
			tea.Printf("⚙️ Adding usecase... (Use: vandor add usecase <name>)"),
			tea.Quit,
		)
	case "Add Service":
		return m, tea.Sequence(
			tea.Printf("🔧 Adding service... (Use: vandor add service <group> <name>)"),
			tea.Quit,
		)
	case "Add HTTP Handler":
		return m, tea.Sequence(
			tea.Printf("🌐 Adding HTTP handler... (Use: vandor add handler <group> <name> <method>)"),
			tea.Quit,
		)
	case "Add Job":
		return m, tea.Sequence(
			tea.Printf("🔄 Adding job... (Use: vandor add job <name>)"),
			tea.Quit,
		)
	case "Add Scheduler":
		return m, tea.Sequence(
			tea.Printf("⏰ Adding scheduler... (Use: vandor add scheduler <name>)"),
			tea.Quit,
		)

	// Generate menu items
	case "Generate All":
		return m, tea.Sequence(
			tea.Printf("🔄 Generating all code... (Use: vandor generate all)"),
			tea.Quit,
		)
	case "Generate Core":
		return m, tea.Sequence(
			tea.Printf("🔄 Generating core code... (Use: vandor generate core)"),
			tea.Quit,
		)
	case "Generate DB Model":
		return m, tea.Sequence(
			tea.Printf("🗄️ Generating DB model... (Use: vandor generate db-model)"),
			tea.Quit,
		)

	// Vpkg menu items
	case "Add Package":
		return m, tea.Sequence(
			tea.Printf("📦 Adding package... (Use: vandor vpkg add <package-name>)"),
			tea.Quit,
		)
	case "List Packages":
		return m, tea.Sequence(
			tea.Printf("📋 Listing packages... (Use: vandor vpkg list)"),
			tea.Quit,
		)
	case "Search Packages":
		return m, tea.Sequence(
			tea.Printf("🔍 Searching packages... (Use: vandor vpkg search <query>)"),
			tea.Quit,
		)

	default:
		return m, nil
	}
}

func (m model) View() string {
	styles := getStyles()
	if m.choice != "" && m.quitting {
		return styles.Quit.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice))
	}
	if m.quitting {
		return styles.Quit.Render("Not hungry? That's cool.")
	}

	var title string
	switch m.screen {
	case "add":
		title = "🔧 Add Components"
	case "generate":
		title = "⚙️ Generate Code"
	case "vpkg":
		title = "📦 Manage Packages"
	default:
		title = "🚀 Vandor CLI"
	}

	return "\n" + styles.Title.Render(title) + "\n\n" + m.list.View()
}

func (m model) createMainMenu() list.Model {
	items := []list.Item{
		item("Initialize Project"),
		item("Add Component"),
		item("Generate Code"),
		item("Manage Packages"),
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

func (m model) createAddMenu() list.Model {
	items := []list.Item{
		item("Add Schema"),
		item("Add Domain"),
		item("Add Usecase"),
		item("Add Service"),
		item("Add HTTP Handler"),
		item("Add Job"),
		item("Add Scheduler"),
	}

	styles := getStyles()
	const defaultWidth = 60
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "What component would you like to add?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styles.Title
	l.Styles.PaginationStyle = styles.Pagination
	l.Styles.HelpStyle = styles.Help

	return l
}

func (m model) createGenerateMenu() list.Model {
	items := []list.Item{
		item("Generate All"),
		item("Generate Core"),
		item("Generate DB Model"),
		item("Generate Domain"),
		item("Generate Usecase"),
		item("Generate Service"),
		item("Generate Handler"),
	}

	styles := getStyles()
	const defaultWidth = 60
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "What would you like to generate?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styles.Title
	l.Styles.PaginationStyle = styles.Pagination
	l.Styles.HelpStyle = styles.Help

	return l
}

func (m model) createVpkgMenu() list.Model {
	items := []list.Item{
		item("Add Package"),
		item("List Packages"),
		item("Search Packages"),
		item("Update Packages"),
	}

	styles := getStyles()
	const defaultWidth = 60
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Package management:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styles.Title
	l.Styles.PaginationStyle = styles.Pagination
	l.Styles.HelpStyle = styles.Help

	return l
}

type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) Run() error {
	items := []list.Item{
		item("Initialize Project"),
		item("Add Component"),
		item("Generate Code"),
		item("Manage Packages"),
		item("Quit"),
	}

	const defaultWidth = 60

	styles := getStyles()
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "What would you like to do?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styles.Title
	l.Styles.PaginationStyle = styles.Pagination
	l.Styles.HelpStyle = styles.Help

	m := model{list: l, screen: "main"}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return fmt.Errorf("error running TUI: %v", err)
	}

	return nil
}
