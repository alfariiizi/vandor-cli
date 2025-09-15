package vpkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// ProgressState represents different stages of installation
type ProgressState int

const (
	StateDiscovering ProgressState = iota
	StateDownloading
	StateRendering
	StateInstalling
	StateCompleted
	StateError
)

// ProgressStep represents a single step in the installation process
type ProgressStep struct {
	Name        string
	Description string
	Progress    float64 // 0.0 to 1.0
	State       ProgressState
	Error       error
}

// ProgressModel is the Bubble Tea model for installation progress
type ProgressModel struct {
	progress    progress.Model
	steps       []ProgressStep
	currentStep int
	packageName string
	totalFiles  int
	processed   int
	width       int
	done        bool
	err         error
	ctx         context.Context
	cancel      context.CancelFunc
}

// ProgressMsg represents progress updates
type ProgressMsg struct {
	Step        int
	Progress    float64
	Description string
	TotalFiles  int
	Processed   int
	Error       error
}

// CompletedMsg indicates installation completion
type CompletedMsg struct {
	Success bool
	Error   error
}

// NewProgressModel creates a new progress model
func NewProgressModel(packageName string) ProgressModel {
	ctx, cancel := context.WithCancel(context.Background())

	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(50),
		progress.WithoutPercentage(),
	)

	steps := []ProgressStep{
		{
			Name:        "Discovery",
			Description: "Discovering template files...",
			State:       StateDiscovering,
		},
		{
			Name:        "Download",
			Description: "Downloading templates...",
			State:       StateDownloading,
		},
		{
			Name:        "Render",
			Description: "Rendering templates...",
			State:       StateRendering,
		},
		{
			Name:        "Install",
			Description: "Installing files...",
			State:       StateInstalling,
		},
	}

	return ProgressModel{
		progress:    prog,
		steps:       steps,
		currentStep: 0,
		packageName: packageName,
		width:       80,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Init implements tea.Model
func (m ProgressModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancel()
			return m, tea.Quit
		case "q":
			if m.done {
				return m, tea.Quit
			}
		}

	case ProgressMsg:
		if msg.Step >= 0 && msg.Step < len(m.steps) {
			m.currentStep = msg.Step
			m.steps[msg.Step].Progress = msg.Progress
			m.steps[msg.Step].Description = msg.Description

			if msg.TotalFiles > 0 {
				m.totalFiles = msg.TotalFiles
			}
			if msg.Processed > 0 {
				m.processed = msg.Processed
			}

			if msg.Error != nil {
				m.steps[msg.Step].State = StateError
				m.steps[msg.Step].Error = msg.Error
				m.err = msg.Error
				m.done = true
			}
		}

	case CompletedMsg:
		m.done = true
		if msg.Success {
			// Mark all steps as completed
			for i := range m.steps {
				m.steps[i].State = StateCompleted
				m.steps[i].Progress = 1.0
			}
		} else {
			m.err = msg.Error
			if m.currentStep < len(m.steps) {
				m.steps[m.currentStep].State = StateError
				m.steps[m.currentStep].Error = msg.Error
			}
		}

		// Auto-quit after showing completion for 2 seconds
		return m, tea.Sequence(
			tea.Tick(2*time.Second, func(time.Time) tea.Msg {
				return tea.Quit()
			}),
		)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model
func (m ProgressModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Header
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(1, 0)

	b.WriteString(titleStyle.Render(fmt.Sprintf("📦 Installing %s", m.packageName)))
	b.WriteString("\n\n")

	// Progress steps
	for i, step := range m.steps {
		b.WriteString(m.renderStep(i, step))
		b.WriteString("\n")
	}

	// Overall progress
	if m.totalFiles > 0 {
		b.WriteString("\n")
		overallProgress := float64(m.processed) / float64(m.totalFiles)
		progressBar := m.progress.ViewAs(overallProgress)

		fileCountStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

		b.WriteString(fmt.Sprintf("📁 Files: %s\n", fileCountStyle.Render(fmt.Sprintf("%d/%d", m.processed, m.totalFiles))))
		b.WriteString(progressBar)
		b.WriteString("\n")
	}

	// Status
	b.WriteString("\n")
	if m.done {
		if m.err != nil {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)
			b.WriteString(errorStyle.Render("❌ Installation failed!"))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(m.err.Error()))
		} else {
			successStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true)
			b.WriteString(successStyle.Render("✅ Installation completed successfully!"))
		}
		b.WriteString("\n\n")

		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
		b.WriteString(helpStyle.Render("Press 'q' to quit"))
	} else {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
		b.WriteString(helpStyle.Render("Press 'ctrl+c' to cancel"))
	}

	return b.String()
}

// renderStep renders a single step with its progress
func (m ProgressModel) renderStep(index int, step ProgressStep) string {
	var icon string
	var style lipgloss.Style

	switch {
	case step.State == StateError:
		icon = "❌"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	case step.State == StateCompleted:
		icon = "✅"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	case index == m.currentStep:
		icon = "🔄"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	case index < m.currentStep:
		icon = "✅"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	default:
		icon = "⏳"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	}

	stepText := fmt.Sprintf("%s %s", icon, step.Name)

	// Add progress bar for current step
	if index == m.currentStep && step.Progress > 0 && step.State != StateCompleted && step.State != StateError {
		miniProgress := progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(20),
			progress.WithoutPercentage(),
		)
		progressBar := miniProgress.ViewAs(step.Progress)
		stepText += fmt.Sprintf(" %s", progressBar)
	}

	result := style.Render(stepText)

	// Add description
	if step.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginLeft(3)
		result += "\n" + descStyle.Render(step.Description)
	}

	// Add error details
	if step.Error != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			MarginLeft(3)
		result += "\n" + errorStyle.Render(fmt.Sprintf("Error: %s", step.Error.Error()))
	}

	return result
}

// SendProgress sends a progress update
func SendProgress(step int, progress float64, description string, totalFiles, processed int, err error) tea.Cmd {
	return func() tea.Msg {
		return ProgressMsg{
			Step:        step,
			Progress:    progress,
			Description: description,
			TotalFiles:  totalFiles,
			Processed:   processed,
			Error:       err,
		}
	}
}

// SendCompleted sends a completion message
func SendCompleted(success bool, err error) tea.Cmd {
	return func() tea.Msg {
		return CompletedMsg{
			Success: success,
			Error:   err,
		}
	}
}

// ProgressInstaller wraps the regular installer with progress tracking
type ProgressInstaller struct {
	*Installer
	model   ProgressModel
	program *tea.Program
}

// NewProgressInstaller creates a new installer with progress tracking
func NewProgressInstaller(registryURL, packageName string) *ProgressInstaller {
	installer := NewInstaller(registryURL)
	model := NewProgressModel(packageName)

	// Create program with proper TTY input handling
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	return &ProgressInstaller{
		Installer: installer,
		model:     model,
		program:   program,
	}
}

// InstallWithProgress installs a package with visual progress
func (pi *ProgressInstaller) InstallWithProgress(packageName string, opts InstallOptions) error {
	// Check if we have a proper TTY for the progress UI
	if !pi.canUseTUI() {
		// Fallback to simple progress without TUI
		return pi.InstallWithSimpleProgress(packageName, opts)
	}

	// Channel to communicate errors from installation goroutine
	errChan := make(chan error, 1)
	doneChan := make(chan struct{}, 1)

	// Start installation in a goroutine
	go func() {
		defer close(doneChan)
		err := pi.installWithProgressTracking(packageName, opts)

		// Send completion message
		if err != nil {
			pi.program.Send(SendCompleted(false, err))
		} else {
			pi.program.Send(SendCompleted(true, nil))
		}

		errChan <- err
	}()

	// Run the TUI and wait for completion
	_, tuiErr := pi.program.Run()

	// Wait for installation to complete
	<-doneChan
	installErr := <-errChan

	// If TUI failed but installation succeeded, that's still OK
	if tuiErr != nil && installErr == nil {
		// TUI failed but installation worked, just print a simple success message
		if installErr == nil {
			fmt.Printf("✅ Package '%s' installed successfully!\n", packageName)
		}
	}

	return installErr
}

// installWithProgressTracking performs the actual installation with progress updates
func (pi *ProgressInstaller) installWithProgressTracking(packageName string, opts InstallOptions) error {
	// Step 1: Discovery
	pi.program.Send(SendProgress(0, 0.1, "Parsing package name...", 0, 0, nil))

	name, version := parsePackageSpec(packageName)
	if version == "" {
		version = opts.Version
	}

	pi.program.Send(SendProgress(0, 0.3, "Finding package in registry...", 0, 0, nil))

	packageWithRepo, err := pi.registryClient.FindPackage(name)
	if err != nil {
		pi.program.Send(SendProgress(0, 0, "Failed to find package", 0, 0, err))
		return fmt.Errorf("failed to find package: %w", err)
	}

	pkg := packageWithRepo.Package

	pi.program.Send(SendProgress(0, 0.6, "Determining destination path...", 0, 0, nil))

	projectRoot, err := pi.findProjectRoot()
	if err != nil {
		pi.program.Send(SendProgress(0, 0, "Failed to find project root", 0, 0, err))
		return fmt.Errorf("failed to find project root: %w", err)
	}

	destPath := opts.Dest
	if destPath == "" {
		destPath = pkg.Destination
	}
	if destPath == "" {
		destPath = fmt.Sprintf("internal/vpkg/%s", name)
	}

	if !filepath.IsAbs(destPath) {
		destPath = filepath.Join(projectRoot, destPath)
	}

	if !opts.Force && pi.packageExists(destPath) {
		existsErr := fmt.Errorf("package already exists at %s (use --force to overwrite)", destPath)
		pi.program.Send(SendProgress(0, 0, "Package already exists", 0, 0, existsErr))
		return existsErr
	}

	pi.program.Send(SendProgress(0, 0.8, "Discovering template files...", 0, 0, nil))

	templateFiles, err := pi.registryClient.DiscoverTemplateFiles(packageWithRepo, pkg.Templates)
	if err != nil {
		pi.program.Send(SendProgress(0, 0, "Failed to discover templates", 0, 0, err))
		return fmt.Errorf("failed to discover template files: %w", err)
	}

	if len(templateFiles) == 0 {
		noFilesErr := fmt.Errorf("no template files found in %s", pkg.Templates)
		pi.program.Send(SendProgress(0, 0, "No template files found", 0, 0, noFilesErr))
		return noFilesErr
	}

	pi.program.Send(SendProgress(0, 1.0, fmt.Sprintf("Found %d template files", len(templateFiles)), len(templateFiles), 0, nil))

	// Step 2: Download
	pi.program.Send(SendProgress(1, 0.1, "Preparing template context...", len(templateFiles), 0, nil))

	ctx, err := pi.prepareTemplateContext(name, &pkg, destPath)
	if err != nil {
		pi.program.Send(SendProgress(1, 0, "Failed to prepare context", len(templateFiles), 0, err))
		return fmt.Errorf("failed to prepare template context: %w", err)
	}

	if !opts.DryRun {
		pi.program.Send(SendProgress(1, 0.2, "Creating destination directory...", len(templateFiles), 0, nil))
		if err := pi.createDestinationDir(destPath); err != nil {
			pi.program.Send(SendProgress(1, 0, "Failed to create directory", len(templateFiles), 0, err))
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	// Step 3: Download and render templates
	for i, templatePath := range templateFiles {
		progress := float64(i) / float64(len(templateFiles))

		pi.program.Send(SendProgress(1, progress, fmt.Sprintf("Downloading %s...", templatePath), len(templateFiles), i, nil))

		if err := pi.installTemplate(packageWithRepo, templatePath, destPath, ctx, opts); err != nil {
			pi.program.Send(SendProgress(1, progress, fmt.Sprintf("Failed to install %s", templatePath), len(templateFiles), i, err))
			return fmt.Errorf("failed to install template %s: %w", templatePath, err)
		}

		pi.program.Send(SendProgress(2, progress, fmt.Sprintf("Rendered %s", templatePath), len(templateFiles), i+1, nil))
	}

	pi.program.Send(SendProgress(1, 1.0, "All templates downloaded", len(templateFiles), len(templateFiles), nil))
	pi.program.Send(SendProgress(2, 1.0, "All templates rendered", len(templateFiles), len(templateFiles), nil))

	// Step 4: Install
	if !opts.DryRun {
		pi.program.Send(SendProgress(3, 0.5, "Writing package metadata...", len(templateFiles), len(templateFiles), nil))

		if err := pi.writeInstalledMeta(destPath, name, version, &pkg); err != nil {
			pi.program.Send(SendProgress(3, 0, "Failed to write metadata", len(templateFiles), len(templateFiles), err))
			return fmt.Errorf("failed to write package metadata: %w", err)
		}

		pi.program.Send(SendProgress(3, 1.0, "Installation completed!", len(templateFiles), len(templateFiles), nil))
	} else {
		pi.program.Send(SendProgress(3, 1.0, "Dry run completed!", len(templateFiles), len(templateFiles), nil))
	}

	return nil
}

// createDestinationDir creates the destination directory
func (pi *ProgressInstaller) createDestinationDir(destPath string) error {
	return os.MkdirAll(destPath, 0755)
}

// canUseTUI checks if we can use the TUI (TTY is available)
func (pi *ProgressInstaller) canUseTUI() bool {
	// Always allow TUI if FORCE_TUI environment variable is set (for testing)
	if os.Getenv("FORCE_TUI") != "" {
		return true
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	// For better compatibility, only require stdout to be a terminal
	// Some environments may not have stdin as terminal but can still display TUI
	return true
}

// InstallWithSimpleProgress performs installation with enhanced text progress
func (pi *ProgressInstaller) InstallWithSimpleProgress(packageName string, opts InstallOptions) error {
	// Header with package info
	fmt.Printf("\n")
	fmt.Printf("╭─────────────────────────────────────────────────────────────╮\n")
	fmt.Printf("│                    📦 VPKG INSTALLER                       │\n")
	fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n")
	fmt.Printf("📦 Package: %s\n", packageName)
	if opts.DryRun {
		fmt.Printf("🧪 Mode: Dry Run (no changes will be made)\n")
	}
	fmt.Printf("\n")

	// Step 1: Discovery Phase
	fmt.Printf("╭─ Step 1/4: Discovery Phase ────────────────────────────────╮\n")

	name, version := parsePackageSpec(packageName)
	if version == "" {
		version = opts.Version
	}
	fmt.Printf("│ 🔍 Parsing package specification... %s", name)
	if version != "" {
		fmt.Printf("@%s", version)
	}
	fmt.Printf("\n")

	fmt.Printf("│ 🌐 Finding package in registry...")
	packageWithRepo, err := pi.registryClient.FindPackage(name)
	if err != nil {
		fmt.Printf(" ❌\n")
		fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n")
		return fmt.Errorf("failed to find package: %w", err)
	}
	fmt.Printf(" ✅\n")

	pkg := packageWithRepo.Package
	fmt.Printf("│ 📋 Package Info:\n")
	fmt.Printf("│   • Name: %s\n", pkg.Name)
	fmt.Printf("│   • Type: %s\n", pkg.Type)
	fmt.Printf("│   • Version: %s\n", pkg.Version)
	fmt.Printf("│   • Description: %s\n", pkg.Description)

	// Determine destination path
	projectRoot, err := pi.findProjectRoot()
	if err != nil {
		fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n")
		return fmt.Errorf("failed to find project root: %w", err)
	}

	destPath := opts.Dest
	if destPath == "" {
		destPath = pkg.Destination
	}
	if destPath == "" {
		destPath = fmt.Sprintf("internal/vpkg/%s", name)
	}

	if !filepath.IsAbs(destPath) {
		destPath = filepath.Join(projectRoot, destPath)
	}

	fmt.Printf("│ 📁 Destination: %s\n", destPath)

	// Check if package already exists
	if !opts.Force && pi.packageExists(destPath) {
		fmt.Printf("│ ⚠️  Package already exists (use --force to overwrite)\n")
		fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n")
		return fmt.Errorf("package already exists at %s (use --force to overwrite)", destPath)
	}

	fmt.Printf("│ 🔍 Discovering template files...")
	templateFiles, err := pi.registryClient.DiscoverTemplateFiles(packageWithRepo, pkg.Templates)
	if err != nil {
		fmt.Printf(" ❌\n")
		fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n")
		return fmt.Errorf("failed to discover template files: %w", err)
	}

	if len(templateFiles) == 0 {
		fmt.Printf(" ❌\n")
		fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n")
		return fmt.Errorf("no template files found in %s", pkg.Templates)
	}
	fmt.Printf(" ✅ Found %d files\n", len(templateFiles))
	fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n\n")

	// Step 2: Preparation Phase
	fmt.Printf("╭─ Step 2/4: Preparation Phase ──────────────────────────────╮\n")
	fmt.Printf("│ ⚙️  Preparing template context...")
	ctx, err := pi.prepareTemplateContext(name, &pkg, destPath)
	if err != nil {
		fmt.Printf(" ❌\n")
		fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n")
		return fmt.Errorf("failed to prepare template context: %w", err)
	}
	fmt.Printf(" ✅\n")

	if !opts.DryRun {
		fmt.Printf("│ 📁 Creating destination directory...")
		if err := pi.createDestinationDir(destPath); err != nil {
			fmt.Printf(" ❌\n")
			fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n")
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
		fmt.Printf(" ✅\n")
	}
	fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n\n")

	// Step 3: Installation Phase
	fmt.Printf("╭─ Step 3/4: Installation Phase ─────────────────────────────╮\n")
	fmt.Printf("│ 📁 Installing %d template files:\n", len(templateFiles))

	// Create progress bar representation
	barWidth := 50
	for i, templatePath := range templateFiles {
		progress := float64(i) / float64(len(templateFiles))
		filledWidth := int(progress * float64(barWidth))

		// Progress bar
		bar := "│ ["
		for j := 0; j < barWidth; j++ {
			if j < filledWidth {
				bar += "█"
			} else {
				bar += "░"
			}
		}
		bar += fmt.Sprintf("] %d/%d", i, len(templateFiles))
		fmt.Printf("\r%s", bar)

		if opts.DryRun {
			time.Sleep(50 * time.Millisecond) // Simulate work for demo
		} else {
			if err := pi.installTemplate(packageWithRepo, templatePath, destPath, ctx, opts); err != nil {
				fmt.Printf("\n│ ❌ Failed to install %s\n", templatePath)
				fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n")
				return fmt.Errorf("failed to install template %s: %w", templatePath, err)
			}
		}

		// Update progress bar to completion for this file
		progress = float64(i+1) / float64(len(templateFiles))
		filledWidth = int(progress * float64(barWidth))
		bar = "│ ["
		for j := 0; j < barWidth; j++ {
			if j < filledWidth {
				bar += "█"
			} else {
				bar += "░"
			}
		}
		bar += fmt.Sprintf("] %d/%d", i+1, len(templateFiles))
		fmt.Printf("\r%s", bar)

		// Show completed file
		outputPath := filepath.Join(destPath, pi.removeTemplateExtension(templatePath))
		fmt.Printf("\n│ ✅ %s\n", outputPath)
	}

	// Final progress bar
	fmt.Printf("│ [")
	for j := 0; j < barWidth; j++ {
		fmt.Printf("█")
	}
	fmt.Printf("] %d/%d ✅ All files processed!\n", len(templateFiles), len(templateFiles))
	fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n\n")

	// Step 4: Finalization Phase
	fmt.Printf("╭─ Step 4/4: Finalization Phase ─────────────────────────────╮\n")
	if !opts.DryRun {
		fmt.Printf("│ 📝 Writing package metadata...")
		if err := pi.writeInstalledMeta(destPath, name, version, &pkg); err != nil {
			fmt.Printf(" ❌\n")
			fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n")
			return fmt.Errorf("failed to write package metadata: %w", err)
		}
		fmt.Printf(" ✅\n")
	}
	fmt.Printf("│ 🎉 Installation completed successfully!\n")
	fmt.Printf("╰─────────────────────────────────────────────────────────────╯\n\n")

	// Usage information
	if !opts.DryRun {
		pi.printUsageReceipt(name, &pkg, ctx)
	}

	return nil
}

// removeTemplateExtension removes template extensions from file path
func (pi *ProgressInstaller) removeTemplateExtension(templatePath string) string {
	extensions := []string{".tmpl", ".templ", ".gotmpl"}

	for _, ext := range extensions {
		if strings.HasSuffix(templatePath, ext) {
			return strings.TrimSuffix(templatePath, ext)
		}
	}

	return templatePath
}
