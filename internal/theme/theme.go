package theme

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme contains all the color styles for the Vandor CLI
type Theme struct {
	Title        lipgloss.AdaptiveColor
	Item         lipgloss.AdaptiveColor
	SelectedItem lipgloss.AdaptiveColor
	Pagination   lipgloss.AdaptiveColor
	Help         lipgloss.AdaptiveColor
	Quit         lipgloss.AdaptiveColor
	Success      lipgloss.AdaptiveColor
	Warning      lipgloss.AdaptiveColor
	Error        lipgloss.AdaptiveColor
	Info         lipgloss.AdaptiveColor
}

// DefaultTheme provides the default adaptive color theme
func DefaultTheme() *Theme {
	return &Theme{
		Title: lipgloss.AdaptiveColor{
			Light: "#1a1a1a", // Dark text for light backgrounds
			Dark:  "#ffffff", // Light text for dark backgrounds
		},
		Item: lipgloss.AdaptiveColor{
			Light: "#4a4a4a", // Medium gray for light backgrounds
			Dark:  "#d1d5db", // Light gray for dark backgrounds
		},
		SelectedItem: lipgloss.AdaptiveColor{
			Light: "#0ea5e9", // Blue for light backgrounds
			Dark:  "#38bdf8", // Lighter blue for dark backgrounds
		},
		Pagination: lipgloss.AdaptiveColor{
			Light: "#6b7280", // Gray for light backgrounds
			Dark:  "#9ca3af", // Lighter gray for dark backgrounds
		},
		Help: lipgloss.AdaptiveColor{
			Light: "#6b7280", // Gray for light backgrounds
			Dark:  "#9ca3af", // Lighter gray for dark backgrounds
		},
		Quit: lipgloss.AdaptiveColor{
			Light: "#374151", // Dark gray for light backgrounds
			Dark:  "#e5e7eb", // Light gray for dark backgrounds
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#059669", // Green for light backgrounds
			Dark:  "#10b981", // Lighter green for dark backgrounds
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#d97706", // Orange for light backgrounds
			Dark:  "#f59e0b", // Lighter orange for dark backgrounds
		},
		Error: lipgloss.AdaptiveColor{
			Light: "#dc2626", // Red for light backgrounds
			Dark:  "#ef4444", // Lighter red for dark backgrounds
		},
		Info: lipgloss.AdaptiveColor{
			Light: "#2563eb", // Blue for light backgrounds
			Dark:  "#3b82f6", // Lighter blue for dark backgrounds
		},
	}
}

// GetStyles returns lipgloss styles configured with the theme
func (t *Theme) GetStyles() *Styles {
	return &Styles{
		Title: lipgloss.NewStyle().
			MarginLeft(2).
			Foreground(t.Title).
			Bold(true),
		Item: lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(t.Item),
		SelectedItem: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(t.SelectedItem).
			Bold(true),
		Pagination: lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(t.Pagination),
		Help: lipgloss.NewStyle().
			PaddingLeft(4).
			PaddingBottom(1).
			Foreground(t.Help),
		Quit: lipgloss.NewStyle().
			Margin(1, 0, 2, 4).
			Foreground(t.Quit),
		Success: lipgloss.NewStyle().
			Foreground(t.Success).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(t.Warning).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(t.Error).
			Bold(true),
		Info: lipgloss.NewStyle().
			Foreground(t.Info).
			Bold(true),
	}
}

// Styles contains the actual lipgloss styles
type Styles struct {
	Title        lipgloss.Style
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	Pagination   lipgloss.Style
	Help         lipgloss.Style
	Quit         lipgloss.Style
	Success      lipgloss.Style
	Warning      lipgloss.Style
	Error        lipgloss.Style
	Info         lipgloss.Style
}

// CatppuccinMochaTheme provides the Catppuccin Mocha (dark) theme
func CatppuccinMochaTheme() *Theme {
	return &Theme{
		Title: lipgloss.AdaptiveColor{
			Light: "#4c4f69", // Fallback for light mode
			Dark:  "#cdd6f4", // Mocha Text
		},
		Item: lipgloss.AdaptiveColor{
			Light: "#6c6f85", // Fallback for light mode
			Dark:  "#bac2de", // Mocha Subtext1
		},
		SelectedItem: lipgloss.AdaptiveColor{
			Light: "#1e66f5", // Fallback for light mode
			Dark:  "#89b4fa", // Mocha Blue
		},
		Pagination: lipgloss.AdaptiveColor{
			Light: "#6c6f85", // Fallback for light mode
			Dark:  "#a6adc8", // Mocha Subtext0
		},
		Help: lipgloss.AdaptiveColor{
			Light: "#6c6f85", // Fallback for light mode
			Dark:  "#a6adc8", // Mocha Subtext0
		},
		Quit: lipgloss.AdaptiveColor{
			Light: "#5c5f77", // Fallback for light mode
			Dark:  "#cdd6f4", // Mocha Text
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#40a02b", // Fallback for light mode
			Dark:  "#a6e3a1", // Mocha Green
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#df8e1d", // Fallback for light mode
			Dark:  "#f9e2af", // Mocha Yellow
		},
		Error: lipgloss.AdaptiveColor{
			Light: "#d20f39", // Fallback for light mode
			Dark:  "#f38ba8", // Mocha Red
		},
		Info: lipgloss.AdaptiveColor{
			Light: "#1e66f5", // Fallback for light mode
			Dark:  "#89b4fa", // Mocha Blue
		},
	}
}

// CatppuccinLatteTheme provides the Catppuccin Latte (light) theme
func CatppuccinLatteTheme() *Theme {
	return &Theme{
		Title: lipgloss.AdaptiveColor{
			Light: "#4c4f69", // Latte Text
			Dark:  "#cdd6f4", // Fallback for dark mode
		},
		Item: lipgloss.AdaptiveColor{
			Light: "#6c6f85", // Latte Subtext1
			Dark:  "#bac2de", // Fallback for dark mode
		},
		SelectedItem: lipgloss.AdaptiveColor{
			Light: "#1e66f5", // Latte Blue
			Dark:  "#89b4fa", // Fallback for dark mode
		},
		Pagination: lipgloss.AdaptiveColor{
			Light: "#8c8fa1", // Latte Subtext0
			Dark:  "#a6adc8", // Fallback for dark mode
		},
		Help: lipgloss.AdaptiveColor{
			Light: "#8c8fa1", // Latte Subtext0
			Dark:  "#a6adc8", // Fallback for dark mode
		},
		Quit: lipgloss.AdaptiveColor{
			Light: "#4c4f69", // Latte Text
			Dark:  "#cdd6f4", // Fallback for dark mode
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#40a02b", // Latte Green
			Dark:  "#a6e3a1", // Fallback for dark mode
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#df8e1d", // Latte Yellow
			Dark:  "#f9e2af", // Fallback for dark mode
		},
		Error: lipgloss.AdaptiveColor{
			Light: "#d20f39", // Latte Red
			Dark:  "#f38ba8", // Fallback for dark mode
		},
		Info: lipgloss.AdaptiveColor{
			Light: "#1e66f5", // Latte Blue
			Dark:  "#89b4fa", // Fallback for dark mode
		},
	}
}

// CatppuccinFrappeTheme provides the Catppuccin Frappe (medium) theme
func CatppuccinFrappeTheme() *Theme {
	return &Theme{
		Title: lipgloss.AdaptiveColor{
			Light: "#4c4f69", // Fallback for light mode
			Dark:  "#c6d0f5", // Frappe Text
		},
		Item: lipgloss.AdaptiveColor{
			Light: "#6c6f85", // Fallback for light mode
			Dark:  "#b5bfe2", // Frappe Subtext1
		},
		SelectedItem: lipgloss.AdaptiveColor{
			Light: "#1e66f5", // Fallback for light mode
			Dark:  "#8caaee", // Frappe Blue
		},
		Pagination: lipgloss.AdaptiveColor{
			Light: "#8c8fa1", // Fallback for light mode
			Dark:  "#a5adce", // Frappe Subtext0
		},
		Help: lipgloss.AdaptiveColor{
			Light: "#8c8fa1", // Fallback for light mode
			Dark:  "#a5adce", // Frappe Subtext0
		},
		Quit: lipgloss.AdaptiveColor{
			Light: "#4c4f69", // Fallback for light mode
			Dark:  "#c6d0f5", // Frappe Text
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#40a02b", // Fallback for light mode
			Dark:  "#a6d189", // Frappe Green
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#df8e1d", // Fallback for light mode
			Dark:  "#e5c890", // Frappe Yellow
		},
		Error: lipgloss.AdaptiveColor{
			Light: "#d20f39", // Fallback for light mode
			Dark:  "#e78284", // Frappe Red
		},
		Info: lipgloss.AdaptiveColor{
			Light: "#1e66f5", // Fallback for light mode
			Dark:  "#8caaee", // Frappe Blue
		},
	}
}

// DraculaTheme provides a Dracula-inspired color theme
func DraculaTheme() *Theme {
	return &Theme{
		Title: lipgloss.AdaptiveColor{
			Light: "#44475a", // Dracula Comment (adapted for light)
			Dark:  "#f8f8f2", // Dracula Foreground
		},
		Item: lipgloss.AdaptiveColor{
			Light: "#6272a4", // Dracula Comment
			Dark:  "#f8f8f2", // Dracula Foreground
		},
		SelectedItem: lipgloss.AdaptiveColor{
			Light: "#8be9fd", // Dracula Cyan
			Dark:  "#8be9fd", // Dracula Cyan
		},
		Pagination: lipgloss.AdaptiveColor{
			Light: "#6272a4", // Dracula Comment
			Dark:  "#6272a4", // Dracula Comment
		},
		Help: lipgloss.AdaptiveColor{
			Light: "#6272a4", // Dracula Comment
			Dark:  "#6272a4", // Dracula Comment
		},
		Quit: lipgloss.AdaptiveColor{
			Light: "#44475a", // Dracula Comment (adapted)
			Dark:  "#f8f8f2", // Dracula Foreground
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#50fa7b", // Dracula Green
			Dark:  "#50fa7b", // Dracula Green
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#f1fa8c", // Dracula Yellow
			Dark:  "#f1fa8c", // Dracula Yellow
		},
		Error: lipgloss.AdaptiveColor{
			Light: "#ff5555", // Dracula Red
			Dark:  "#ff5555", // Dracula Red
		},
		Info: lipgloss.AdaptiveColor{
			Light: "#bd93f9", // Dracula Purple
			Dark:  "#bd93f9", // Dracula Purple
		},
	}
}

// detectSystemTheme attempts to detect if the system is using a dark or light theme
func detectSystemTheme() bool {
	// Check environment variables that might indicate theme preference
	if colorTerm := os.Getenv("COLORFGBG"); colorTerm != "" {
		// COLORFGBG format is usually "foreground;background"
		// If background is light (high number), it's likely a light theme
		parts := strings.Split(colorTerm, ";")
		if len(parts) >= 2 {
			// This is a heuristic - not perfect but works for many terminals
			bg := parts[1]
			return bg == "0" || bg == "8" // Dark backgrounds
		}
	}

	// Check other environment variables
	if term := os.Getenv("TERM"); term != "" {
		// Some terminals set specific TERM values for dark themes
		if strings.Contains(strings.ToLower(term), "dark") {
			return true
		}
	}

	// Check for macOS dark mode (if running on macOS)
	if os.Getenv("XPC_SERVICE_NAME") != "" { // Likely macOS
		// This is a simplified check - in practice you might want to use osascript
		// For now, we'll default to dark theme on macOS
		return true
	}

	// Default to dark theme if we can't detect (most modern terminals default to dark)
	return true
}

// GetAutoTheme returns the appropriate Catppuccin theme based on system detection
func GetAutoTheme() *Theme {
	isDark := detectSystemTheme()
	if isDark {
		return CatppuccinMochaTheme() // Dark theme
	}
	return CatppuccinLatteTheme() // Light theme
}

// Global theme instance - now defaults to auto-detected Catppuccin theme
var currentTheme = GetAutoTheme()
var currentStyles = currentTheme.GetStyles()
var currentThemeName = "auto"

// GetCurrentStyles returns the current theme styles
func GetCurrentStyles() *Styles {
	return currentStyles
}

// SetTheme sets a new theme and updates the styles
func SetTheme(theme *Theme) {
	currentTheme = theme
	currentStyles = currentTheme.GetStyles()
}

// GetCurrentThemeName returns the name of the current theme
func GetCurrentThemeName() string {
	return currentThemeName
}

// SetDefaultTheme sets the default theme
func SetDefaultTheme() {
	currentThemeName = "default"
	SetTheme(DefaultTheme())
}

// SetCatppuccinMochaTheme sets the Catppuccin Mocha theme (dark)
func SetCatppuccinMochaTheme() {
	currentThemeName = "mocha"
	SetTheme(CatppuccinMochaTheme())
}

// SetCatppuccinLatteTheme sets the Catppuccin Latte theme (light)
func SetCatppuccinLatteTheme() {
	currentThemeName = "latte"
	SetTheme(CatppuccinLatteTheme())
}

// SetCatppuccinFrappeTheme sets the Catppuccin Frappe theme (medium)
func SetCatppuccinFrappeTheme() {
	currentThemeName = "frappe"
	SetTheme(CatppuccinFrappeTheme())
}

// SetAutoTheme automatically selects the appropriate theme based on system detection
func SetAutoTheme() {
	currentThemeName = "auto"
	SetTheme(GetAutoTheme())
}

// SetDraculaTheme sets the Dracula theme
func SetDraculaTheme() {
	currentThemeName = "dracula"
	SetTheme(DraculaTheme())
}
