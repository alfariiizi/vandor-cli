package command

import (
	"fmt"

	"github.com/alfariiizi/vandor-cli/internal/theme"
)

// ThemeListCommand implements the theme list functionality
type ThemeListCommand struct{}

func NewThemeListCommand() *ThemeListCommand {
	return &ThemeListCommand{}
}

func (c *ThemeListCommand) Execute(ctx *CommandContext) error {
	_, _ = fmt.Fprintf(ctx.Stdout, "Available themes:\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "Catppuccin themes:\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "  mocha    - Catppuccin Mocha (dark, recommended for dark terminals)\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "  latte    - Catppuccin Latte (light, recommended for light terminals)\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "  frappe   - Catppuccin Frappe (medium contrast)\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "Other themes:\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "  default  - Default adaptive theme\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "  dracula  - Dracula inspired theme\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "  auto     - Auto-detect based on system theme (default)\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "\n")
	_, _ = fmt.Fprintf(ctx.Stdout, "Usage: vandor theme set <theme-name>\n")
	return nil
}

func (c *ThemeListCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "list",
		Category:    "theme",
		Description: "List available themes",
		Usage:       "vandor theme list",
		Args:        []string{},
	}
}

func (c *ThemeListCommand) Validate(args []string) error {
	return nil // No arguments required
}

// ThemeSetCommand implements the theme set functionality
type ThemeSetCommand struct{}

func NewThemeSetCommand() *ThemeSetCommand {
	return &ThemeSetCommand{}
}

func (c *ThemeSetCommand) Execute(ctx *CommandContext) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("theme name is required")
	}

	themeName := ctx.Args[0]

	switch themeName {
	case "mocha":
		theme.SetCatppuccinMochaTheme()
		_, _ = fmt.Fprintf(ctx.Stdout, "✅ Theme set to Catppuccin Mocha (dark)\n")
	case "latte":
		theme.SetCatppuccinLatteTheme()
		_, _ = fmt.Fprintf(ctx.Stdout, "✅ Theme set to Catppuccin Latte (light)\n")
	case "frappe":
		theme.SetCatppuccinFrappeTheme()
		_, _ = fmt.Fprintf(ctx.Stdout, "✅ Theme set to Catppuccin Frappe (medium contrast)\n")
	case "dracula":
		theme.SetDraculaTheme()
		_, _ = fmt.Fprintf(ctx.Stdout, "✅ Theme set to Dracula\n")
	case "default":
		theme.SetDefaultTheme()
		_, _ = fmt.Fprintf(ctx.Stdout, "✅ Theme set to Default (adaptive)\n")
	case "auto":
		theme.SetAutoTheme()
		_, _ = fmt.Fprintf(ctx.Stdout, "✅ Theme set to Auto-detect\n")
	default:
		return fmt.Errorf("unknown theme: %s. Use 'vandor theme list' to see available themes", themeName)
	}

	return nil
}

func (c *ThemeSetCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "set",
		Category:    "theme",
		Description: "Set the active theme",
		Usage:       "vandor theme set <theme-name>",
		Args:        []string{"theme-name"},
	}
}

func (c *ThemeSetCommand) Validate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("theme name is required")
	}

	validThemes := []string{"mocha", "latte", "frappe", "dracula", "default", "auto"}
	themeName := args[0]

	for _, valid := range validThemes {
		if themeName == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid theme: %s. Valid themes are: mocha, latte, frappe, dracula, default, auto", themeName)
}

// ThemeInfoCommand implements the theme info functionality
type ThemeInfoCommand struct{}

func NewThemeInfoCommand() *ThemeInfoCommand {
	return &ThemeInfoCommand{}
}

func (c *ThemeInfoCommand) Execute(ctx *CommandContext) error {
	currentTheme := theme.GetCurrentThemeName()
	_, _ = fmt.Fprintf(ctx.Stdout, "Current theme: %s\n", currentTheme)

	// Get current theme description
	switch currentTheme {
	case "mocha":
		_, _ = fmt.Fprintf(ctx.Stdout, "Description: Catppuccin Mocha - Dark theme with warm, cozy colors\n")
	case "latte":
		_, _ = fmt.Fprintf(ctx.Stdout, "Description: Catppuccin Latte - Light theme with soft, pleasant colors\n")
	case "frappe":
		_, _ = fmt.Fprintf(ctx.Stdout, "Description: Catppuccin Frappe - Medium contrast theme\n")
	case "dracula":
		_, _ = fmt.Fprintf(ctx.Stdout, "Description: Dracula inspired theme\n")
	case "default":
		_, _ = fmt.Fprintf(ctx.Stdout, "Description: Default adaptive theme\n")
	case "auto":
		_, _ = fmt.Fprintf(ctx.Stdout, "Description: Auto-detect theme based on system settings\n")
	default:
		_, _ = fmt.Fprintf(ctx.Stdout, "Description: Unknown theme\n")
	}

	return nil
}

func (c *ThemeInfoCommand) GetMetadata() CommandMetadata {
	return CommandMetadata{
		Name:        "info",
		Category:    "theme",
		Description: "Show current theme information",
		Usage:       "vandor theme info",
		Args:        []string{},
	}
}

func (c *ThemeInfoCommand) Validate(args []string) error {
	return nil // No arguments required
}
