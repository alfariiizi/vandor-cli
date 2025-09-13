package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/alfariiizi/vandor-cli/internal/theme"
)

type VandorConfig struct {
	Project struct {
		Name    string `yaml:"name"`
		Module  string `yaml:"module"`
		Version string `yaml:"version"`
	} `yaml:"project"`
	Vandor struct {
		CLI          string `yaml:"cli"`
		Architecture string `yaml:"architecture"`
		Language     string `yaml:"language"`
	} `yaml:"vandor"`
}

type VpkgItem struct {
	Name    string   `yaml:"name"`
	Version string   `yaml:"version"`
	Tags    []string `yaml:"tags"`
}

type TemplateConfig struct {
	Repositories map[string]string `yaml:"repositories"`
}

// Dependency represents a required tool with installation instructions
type Dependency struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	CheckCommand []string `json:"check_command"`
	InstallCmd   []string `json:"install_command"`
	ManualURL    string   `json:"manual_url"`
	Required     bool     `json:"required"`
}

// ArchitectureDependencies defines dependencies for each architecture type
type ArchitectureDependencies struct {
	Dependencies []Dependency `json:"dependencies"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Vandor project by cloning GitHub templates",
	Long:  `Initialize a new Vandor project by cloning from GitHub templates and customizing it with your project details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initProject(); err != nil {
			er(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// getTemplateRepositories returns the default GitHub template repositories
func getTemplateRepositories() map[string]string {
	return map[string]string{
		"full-backend": "https://github.com/alfariiizi/vandor-backend-template.git",
		"eda":          "https://github.com/alfariiizi/vandor-eda-template.git",
		"minimal":      "https://github.com/alfariiizi/vandor-minimal-template.git",
	}
}

// cloneTemplate clones a GitHub repository template
func cloneTemplate(repoURL, targetDir, projectName string) error {
	styles := theme.GetCurrentStyles()

	// Clone the repository
	cmd := exec.Command("git", "clone", repoURL, targetDir)
	cmd.Stdout = nil // Suppress git clone output
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone template repository: %v", err)
	}

	fmt.Println(styles.Success.Render("✅ Template cloned successfully!"))

	// Remove .git directory to detach from template repo
	gitDir := filepath.Join(targetDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  Warning: Could not remove .git directory: %v", err)))
	}

	// Initialize new git repository
	if err := initializeGitRepo(targetDir); err != nil {
		fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  Warning: Could not initialize git repository: %v", err)))
	} else {
		fmt.Println(styles.Success.Render("✅ New git repository initialized!"))
	}

	return nil
}

// initializeGitRepo initializes a new git repository
func initializeGitRepo(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	return cmd.Run()
}

// createVandorConfig creates the vandor-config.yaml file in the target directory
func createVandorConfig(targetDir string, config VandorConfig) error {
	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	configPath := filepath.Join(targetDir, "vandor-config.yaml")
	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// runGoModTidy runs 'go mod tidy' in the specified directory
func runGoModTidy(dir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil

	return cmd.Run()
}

// getArchitectureDependencies returns the dependencies for a given architecture
func getArchitectureDependencies(architecture string) *ArchitectureDependencies {
	switch architecture {
	case "full-backend":
		return &ArchitectureDependencies{
			Dependencies: []Dependency{
				{
					Name:         "air",
					Description:  "Hot reload tool for Go development",
					CheckCommand: []string{"air", "--version"},
					InstallCmd:   []string{"go", "install", "github.com/cosmtrek/air@latest"},
					ManualURL:    "https://github.com/cosmtrek/air",
					Required:     true,
				},
				{
					Name:         "atlas",
					Description:  "Database migration and schema management tool",
					CheckCommand: []string{"atlas", "version"},
					InstallCmd:   []string{"go", "install", "ariga.io/atlas/cmd/atlas@latest"},
					ManualURL:    "https://atlasgo.io/getting-started",
					Required:     true,
				},
			},
		}
	case "eda":
		return &ArchitectureDependencies{
			Dependencies: []Dependency{
				{
					Name:         "air",
					Description:  "Hot reload tool for Go development",
					CheckCommand: []string{"air", "--version"},
					InstallCmd:   []string{"go", "install", "github.com/cosmtrek/air@latest"},
					ManualURL:    "https://github.com/cosmtrek/air",
					Required:     false, // Optional for EDA
				},
			},
		}
	case "minimal":
		// No specific dependencies for minimal setup
		return &ArchitectureDependencies{
			Dependencies: []Dependency{},
		}
	default:
		return &ArchitectureDependencies{
			Dependencies: []Dependency{},
		}
	}
}

// Example of how to add dependencies for new architectures:
//
// case "microservices":
//     return &ArchitectureDependencies{
//         Dependencies: []Dependency{
//             {
//                 Name:         "docker",
//                 Description:  "Container runtime for microservices",
//                 CheckCommand: []string{"docker", "--version"},
//                 InstallCmd:   []string{"curl", "-fsSL", "https://get.docker.com", "-o", "get-docker.sh"},
//                 ManualURL:    "https://docs.docker.com/get-docker/",
//                 Required:     true,
//             },
//             {
//                 Name:         "kubectl",
//                 Description:  "Kubernetes command-line tool",
//                 CheckCommand: []string{"kubectl", "version", "--client"},
//                 InstallCmd:   []string{"go", "install", "k8s.io/kubectl@latest"},
//                 ManualURL:    "https://kubernetes.io/docs/tasks/tools/install-kubectl/",
//                 Required:     false,
//             },
//         },
//     }

// checkAndInstallDependencies checks and installs dependencies for the given architecture
func checkAndInstallDependencies(architecture string) error {
	styles := theme.GetCurrentStyles()
	reader := bufio.NewReader(os.Stdin)

	deps := getArchitectureDependencies(architecture)

	// Skip if no dependencies
	if len(deps.Dependencies) == 0 {
		fmt.Println(styles.Success.Render("✅ No additional dependencies required for this architecture"))
		return nil
	}

	fmt.Println()
	// Capitalize first letter of architecture
	archTitle := strings.ToUpper(string(architecture[0])) + architecture[1:]
	fmt.Println(styles.Title.Render(fmt.Sprintf("🔧 Checking Dependencies for %s Architecture", archTitle)))

	// Build description of what's needed
	var depNames []string
	for _, dep := range deps.Dependencies {
		depNames = append(depNames, fmt.Sprintf("'%s' (%s)", dep.Name, dep.Description))
	}
	fmt.Println(styles.Info.Render(fmt.Sprintf("This template requires: %s", strings.Join(depNames, ", "))))

	for _, dep := range deps.Dependencies {
		if err := checkAndInstallSingleDependency(dep, reader, styles); err != nil {
			if dep.Required {
				return err
			}
			fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  %s installation skipped (optional)", dep.Name)))
		}
	}

	fmt.Println(styles.Success.Render("🎉 All dependencies are ready!"))
	return nil
}

// checkAndInstallSingleDependency checks and installs a single dependency
func checkAndInstallSingleDependency(dep Dependency, reader *bufio.Reader, styles *theme.Styles) error {
	// Check if dependency is installed
	installed := isDependencyInstalled(dep.CheckCommand)

	if installed {
		fmt.Println(styles.Success.Render(fmt.Sprintf("✅ %s is already installed", dep.Name)))
		return nil
	}

	requiredText := ""
	if dep.Required {
		requiredText = " (required)"
	} else {
		requiredText = " (optional)"
	}

	fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  %s is not installed%s", dep.Name, requiredText)))
	fmt.Print(styles.Item.Render(fmt.Sprintf("Install %s automatically? [Y/n]: ", dep.Name)))

	choice, _ := reader.ReadString('\n')
	choice = strings.ToLower(strings.TrimSpace(choice))

	if choice == "" || choice == "y" || choice == "yes" {
		if err := installDependency(dep); err != nil {
			fmt.Println(styles.Error.Render(fmt.Sprintf("❌ Failed to install %s: %v", dep.Name, err)))
			showManualInstallation(dep, styles)

			if dep.Required {
				return fmt.Errorf("%s installation failed", dep.Name)
			}
			return err
		}
		fmt.Println(styles.Success.Render(fmt.Sprintf("✅ %s installed successfully!", dep.Name)))
		return nil
	}

	// User chose not to auto-install
	showManualInstallation(dep, styles)

	if dep.Required {
		fmt.Print(styles.Item.Render(fmt.Sprintf("Continue without %s? [y/N]: ", dep.Name)))
		continueChoice, _ := reader.ReadString('\n')
		continueChoice = strings.ToLower(strings.TrimSpace(continueChoice))
		if continueChoice != "y" && continueChoice != "yes" {
			return fmt.Errorf("%s is required for %s development", dep.Name, dep.Description)
		}
	}

	return nil
}

// isDependencyInstalled checks if a command-line tool is available using the check command
func isDependencyInstalled(checkCommand []string) bool {
	if len(checkCommand) == 0 {
		return false
	}

	// Try the specific check command first
	cmd := exec.Command(checkCommand[0], checkCommand[1:]...)
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil
	if err := cmd.Run(); err == nil {
		return true
	}

	// Fallback: try with 'which' on Unix-like systems
	cmd = exec.Command("which", checkCommand[0])
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err == nil {
		return true
	}

	// Fallback: try with 'where' on Windows
	cmd = exec.Command("where", checkCommand[0])
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

// installDependency installs a dependency using its install command
func installDependency(dep Dependency) error {
	styles := theme.GetCurrentStyles()
	fmt.Println(styles.Info.Render(fmt.Sprintf("📦 Installing %s (%s)...", dep.Name, dep.Description)))

	if len(dep.InstallCmd) == 0 {
		return fmt.Errorf("no install command specified for %s", dep.Name)
	}

	// Execute the install command
	cmd := exec.Command(dep.InstallCmd[0], dep.InstallCmd[1:]...)
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil

	return cmd.Run()
}

// showManualInstallation shows manual installation instructions for a dependency
func showManualInstallation(dep Dependency, styles *theme.Styles) {
	fmt.Println()
	fmt.Println(styles.Info.Render(fmt.Sprintf("📖 Manual %s Installation:", dep.Name)))
	if len(dep.InstallCmd) > 0 {
		fmt.Println(styles.Item.Render(fmt.Sprintf("  %s", strings.Join(dep.InstallCmd, " "))))
	}
	if dep.ManualURL != "" {
		fmt.Println(styles.Item.Render(fmt.Sprintf("  or visit: %s", dep.ManualURL)))
	}
	fmt.Println()
}

// checkGitInstalled checks if git is available
func checkGitInstalled() error {
	cmd := exec.Command("git", "--version")
	return cmd.Run()
}

func initProject() error {
	styles := theme.GetCurrentStyles()
	reader := bufio.NewReader(os.Stdin)

	// Welcome message
	fmt.Println(styles.Title.Render("🚀 Vandor Project Initializer"))
	fmt.Println(styles.Info.Render("Initialize a new Vandor project from GitHub templates"))
	fmt.Println()

	// Check if git is installed
	if err := checkGitInstalled(); err != nil {
		fmt.Println(styles.Error.Render("❌ Git is not installed or not available in PATH"))
		return fmt.Errorf("git is required for cloning templates")
	}

	config := VandorConfig{}

	// Get project information
	fmt.Print(styles.Item.Render("Project name (e.g., my-clinic-app): "))
	projectName, _ := reader.ReadString('\n')
	config.Project.Name = strings.TrimSpace(projectName)

	if config.Project.Name == "" {
		fmt.Println(styles.Error.Render("❌ Project name is required"))
		return fmt.Errorf("project name cannot be empty")
	}

	fmt.Print(styles.Item.Render("Go module path (e.g., github.com/your-org/my-clinic-app): "))
	modulePath, _ := reader.ReadString('\n')
	config.Project.Module = strings.TrimSpace(modulePath)

	if config.Project.Module == "" {
		fmt.Println(styles.Error.Render("❌ Go module path is required"))
		return fmt.Errorf("go module path cannot be empty")
	}

	fmt.Print(styles.Item.Render("Project version [0.1.0]: "))
	version, _ := reader.ReadString('\n')
	version = strings.TrimSpace(version)
	if version == "" {
		version = "0.1.0"
	}
	config.Project.Version = version

	vandorVersion, _, _ := getVersionInfo()

	// Set Vandor CLI version
	config.Vandor.CLI = vandorVersion
	config.Vandor.Language = "go"

	// Ask for architecture type
	fmt.Println()
	fmt.Println(styles.Title.Render("🏗️  Select Architecture Type:"))
	fmt.Println(styles.Item.Render("1. full-backend (Complete backend with all features)"))
	fmt.Println(styles.Item.Render("2. eda (Event-driven architecture)"))
	fmt.Println(styles.Item.Render("3. minimal (Minimal setup)"))
	fmt.Print(styles.SelectedItem.Render("Choose [1-3]: "))

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		config.Vandor.Architecture = "full-backend"
	case "2":
		config.Vandor.Architecture = "eda"
	case "3":
		config.Vandor.Architecture = "minimal"
	default:
		config.Vandor.Architecture = "minimal"
		fmt.Println(styles.Warning.Render("⚠️  No valid choice selected, defaulting to minimal"))
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✅ Selected architecture: %s", config.Vandor.Architecture)))

	// Check dependencies for the selected architecture
	if err := checkAndInstallDependencies(config.Vandor.Architecture); err != nil {
		fmt.Println(styles.Error.Render(fmt.Sprintf("❌ Dependency check failed: %v", err)))
		return err
	}

	// Ask where to create the project
	fmt.Println()
	fmt.Println(styles.Title.Render("📁 Project Location:"))
	fmt.Print(styles.Item.Render(fmt.Sprintf("Create new directory '%s' or use current directory? [new/current]: ", config.Project.Name)))
	locationChoice, _ := reader.ReadString('\n')
	locationChoice = strings.ToLower(strings.TrimSpace(locationChoice))

	var targetDir string
	if locationChoice == "current" || locationChoice == "c" {
		targetDir = "."
		fmt.Println(styles.Info.Render("📂 Using current directory"))
	} else {
		targetDir = config.Project.Name
		fmt.Println(styles.Info.Render(fmt.Sprintf("📂 Creating new directory: %s", targetDir)))
	}

	// Clone the project from GitHub template
	if err := createProjectFromGitHubTemplate(config, targetDir); err != nil {
		fmt.Println(styles.Error.Render(fmt.Sprintf("❌ Failed to create project from GitHub template: %v", err)))
		return err
	}

	fmt.Println()
	fmt.Println(styles.Success.Render("🎉 Project created successfully!"))

	return nil
}

// createProjectFromGitHubTemplate creates a project by cloning from GitHub template
func createProjectFromGitHubTemplate(config VandorConfig, targetDir string) error {
	styles := theme.GetCurrentStyles()

	// Get template repositories
	templates := getTemplateRepositories()

	// Get the repository URL for the selected architecture
	repoURL, exists := templates[config.Vandor.Architecture]
	if !exists {
		return fmt.Errorf("no template repository configured for architecture: %s", config.Vandor.Architecture)
	}

	fmt.Println()
	fmt.Println(styles.Info.Render(fmt.Sprintf("📡 Cloning %s template from GitHub...", config.Vandor.Architecture)))
	fmt.Println(styles.Item.Render(fmt.Sprintf("Repository: %s", repoURL)))

	// Clone the template
	if targetDir == "." {
		// For current directory, clone to temp and move files
		tempDir := ".vandor-temp"
		if err := cloneTemplate(repoURL, tempDir, config.Project.Name); err != nil {
			return err
		}

		// Move files from temp to current directory
		if err := moveTemplateFiles(tempDir, "."); err != nil {
			if rmErr := os.RemoveAll(tempDir); rmErr != nil {
				fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  Warning: failed to clean up temp directory: %v", rmErr)))
			}
			return fmt.Errorf("failed to move template files: %v", err)
		}

		// Clean up temp directory
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  Warning: failed to clean up temp directory: %v", err)))
		}
	} else {
		// Clone directly to subdirectory
		if err := cloneTemplate(repoURL, targetDir, config.Project.Name); err != nil {
			return err
		}
	}

	// Update project files with actual configuration
	fmt.Println(styles.Info.Render("🔄 Customizing template with your project details..."))
	if err := updateProjectConfiguration(targetDir, config); err != nil {
		fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  Warning: Could not update project configuration: %v", err)))
	}

	// Create/update vandor-config.yaml in the target directory
	if err := createVandorConfig(targetDir, config); err != nil {
		fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  Warning: Could not create vandor-config.yaml: %v", err)))
	}

	// Ask if user wants to run go mod tidy
	fmt.Println()
	fmt.Print(styles.Item.Render("Run 'go mod tidy' to clean up dependencies? [Y/n]: "))
	reader := bufio.NewReader(os.Stdin)
	tidyChoice, _ := reader.ReadString('\n')
	tidyChoice = strings.ToLower(strings.TrimSpace(tidyChoice))

	if tidyChoice == "" || tidyChoice == "y" || tidyChoice == "yes" {
		fmt.Println(styles.Info.Render("📦 Running go mod tidy..."))
		if err := runGoModTidy(targetDir); err != nil {
			fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  Warning: go mod tidy failed: %v", err)))
		} else {
			fmt.Println(styles.Success.Render("✅ Dependencies cleaned up successfully!"))
		}
	} else {
		fmt.Println(styles.Info.Render("ℹ️  Skipping go mod tidy - remember to run it later"))
	}

	return nil
}

// moveTemplateFiles moves files from source to destination
func moveTemplateFiles(src, dst string) error {
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		dstPath := filepath.Join(dst, file.Name())

		if err := os.Rename(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to move %s: %v", file.Name(), err)
		}
	}

	return nil
}

// updateProjectConfiguration updates project files with actual configuration
func updateProjectConfiguration(projectDir string, config VandorConfig) error {
	styles := theme.GetCurrentStyles()

	// Replace all module names in the entire project
	if err := replaceModuleNamesInProject(projectDir, config.Project.Module); err != nil {
		return fmt.Errorf("failed to replace module names: %v", err)
	}

	// Update README.md if it exists
	readmePath := filepath.Join(projectDir, "README.md")
	if _, err := os.Stat(readmePath); err == nil {
		if err := updateReadme(readmePath, config.Project.Name); err != nil {
			fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  Warning: Could not update README.md: %v", err)))
		} else {
			fmt.Println(styles.Success.Render("✅ README.md updated!"))
		}
	}

	return nil
}

// replaceModuleNamesInProject walks through all Go files and replaces module imports
func replaceModuleNamesInProject(projectDir, newModuleName string) error {
	styles := theme.GetCurrentStyles()

	// First, find the current module name from go.mod
	goModPath := filepath.Join(projectDir, "go.mod")
	oldModuleName, err := extractModuleNameFromGoMod(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read current module name: %v", err)
	}

	if oldModuleName == "" {
		fmt.Println(styles.Warning.Render("⚠️  Warning: Could not detect current module name"))
		return nil
	}

	fmt.Println(styles.Info.Render(fmt.Sprintf("🔄 Replacing module name: %s → %s", oldModuleName, newModuleName)))

	// Update go.mod file
	if updateErr := updateGoMod(goModPath, newModuleName); updateErr != nil {
		return fmt.Errorf("failed to update go.mod: %v", updateErr)
	}
	fmt.Println(styles.Success.Render("✅ go.mod updated!"))

	// Walk through all .go files and replace import statements
	fileCount := 0
	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Replace module imports in Go files
		if err := replaceModuleInGoFile(path, oldModuleName, newModuleName); err != nil {
			fmt.Println(styles.Warning.Render(fmt.Sprintf("⚠️  Warning: Could not update %s: %v", path, err)))
		} else {
			fileCount++
		}

		return nil
	})

	if err != nil {
		return err
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✅ Updated %d Go files with new module name!", fileCount)))
	return nil
}

// extractModuleNameFromGoMod extracts the current module name from go.mod
func extractModuleNameFromGoMod(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}

	return "", fmt.Errorf("module name not found in go.mod")
}

// updateGoMod updates the go.mod file with the correct module name
func updateGoMod(goModPath, moduleName string) error {
	if moduleName == "" {
		return nil // Skip if no module name provided
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}

	// Replace the module line
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "module ") {
			lines[i] = fmt.Sprintf("module %s", moduleName)
			break
		}
	}

	updatedContent := strings.Join(lines, "\n")
	return os.WriteFile(goModPath, []byte(updatedContent), 0644)
}

// replaceModuleInGoFile replaces module import statements in a Go file
func replaceModuleInGoFile(filePath, oldModule, newModule string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Use regex to replace import statements
	// This handles both single imports and grouped imports
	importRegex := regexp.MustCompile(`"` + regexp.QuoteMeta(oldModule) + `([/\w-]*)"`)
	updatedContent := importRegex.ReplaceAllString(string(content), `"`+newModule+`$1"`)

	// Only write if content changed
	if updatedContent != string(content) {
		return os.WriteFile(filePath, []byte(updatedContent), 0644)
	}

	return nil
}

// updateReadme updates the README.md file with the correct project name
func updateReadme(readmePath, projectName string) error {
	if projectName == "" {
		return nil // Skip if no project name provided
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	// Simple replacement of template placeholders
	updatedContent := strings.ReplaceAll(string(content), "{{PROJECT_NAME}}", projectName)
	updatedContent = strings.ReplaceAll(updatedContent, "PROJECT_NAME", projectName)

	return os.WriteFile(readmePath, []byte(updatedContent), 0644)
}
