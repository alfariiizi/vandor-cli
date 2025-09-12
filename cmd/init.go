package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	Vpkg []VpkgItem `yaml:"vpkg,omitempty"`
}

type VpkgItem struct {
	Name    string   `yaml:"name"`
	Version string   `yaml:"version"`
	Tags    []string `yaml:"tags"`
}

type TemplateConfig struct {
	Repositories map[string]string `yaml:"repositories"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Vandor project",
	Long:  `Initialize a new Vandor project with configuration and optional project setup.`,
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
		"full-backend": "https://github.com/vandordev/vandor-template-full-backend.git",
		"eda":          "https://github.com/vandordev/vandor-template-eda.git",
		"minimal":      "https://github.com/vandordev/vandor-template-minimal.git",
	}
}

// cloneTemplate clones a GitHub repository template
func cloneTemplate(repoURL, targetDir, projectName string) error {
	fmt.Printf("📥 Cloning template from %s...\n", repoURL)
	
	// Clone the repository
	cmd := exec.Command("git", "clone", repoURL, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone template repository: %v", err)
	}
	
	// Remove .git directory to detach from template repo
	gitDir := filepath.Join(targetDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		fmt.Printf("Warning: Could not remove .git directory: %v\n", err)
	}
	
	// Initialize new git repository
	if err := initializeGitRepo(targetDir); err != nil {
		fmt.Printf("Warning: Could not initialize git repository: %v\n", err)
	}
	
	// Replace template placeholders with actual project name
	if err := replaceTemplatePlaceholders(targetDir, projectName); err != nil {
		fmt.Printf("Warning: Could not replace template placeholders: %v\n", err)
	}
	
	return nil
}

// initializeGitRepo initializes a new git repository
func initializeGitRepo(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	return cmd.Run()
}

// replaceTemplatePlaceholders replaces common template placeholders
func replaceTemplatePlaceholders(dir, projectName string) error {
	// This is a simplified implementation - you might want to make this more sophisticated
	// by walking through files and replacing specific placeholders
	fmt.Printf("🔄 Customizing template for project '%s'...\n", projectName)
	return nil
}

// checkGitInstalled checks if git is available
func checkGitInstalled() error {
	cmd := exec.Command("git", "--version")
	return cmd.Run()
}

func initProject() error {
	reader := bufio.NewReader(os.Stdin)

	// Check if vandor-config.yaml already exists
	if _, err := os.Stat("vandor-config.yaml"); err == nil {
		fmt.Print("vandor-config.yaml already exists. Overwrite? [y/N]: ")
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Initialization cancelled.")
			return nil
		}
	}

	config := VandorConfig{}

	// Get project information
	fmt.Print("Project name (e.g., my-app): ")
	projectName, _ := reader.ReadString('\n')
	config.Project.Name = strings.TrimSpace(projectName)

	fmt.Print("Go module path (e.g., github.com/your-org/my-app): ")
	modulePath, _ := reader.ReadString('\n')
	config.Project.Module = strings.TrimSpace(modulePath)

	fmt.Print("Project version [0.1.0]: ")
	version, _ := reader.ReadString('\n')
	version = strings.TrimSpace(version)
	if version == "" {
		version = "0.1.0"
	}
	config.Project.Version = version

	// Set Vandor CLI version
	config.Vandor.CLI = "0.5.0"
	config.Vandor.Language = "go"

	// Ask for architecture type
	fmt.Println("\nSelect architecture type:")
	fmt.Println("1. full-backend (Complete backend with all features)")
	fmt.Println("2. eda (Event-driven architecture)")
	fmt.Println("3. minimal (Minimal setup)")
	fmt.Print("Choose [1-3]: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		config.Vandor.Architecture = "full-backend"
		config.Vpkg = []VpkgItem{
			{Name: "audit-logger", Version: "1.0.0", Tags: []string{"full-backend", "eda"}},
			{Name: "redis-cache", Version: "1.2.0", Tags: []string{"full-backend", "eda", "minimal"}},
		}
	case "2":
		config.Vandor.Architecture = "eda"
		config.Vpkg = []VpkgItem{
			{Name: "audit-logger", Version: "1.0.0", Tags: []string{"full-backend", "eda"}},
			{Name: "redis-cache", Version: "1.2.0", Tags: []string{"full-backend", "eda", "minimal"}},
			{Name: "kafka-bus", Version: "2.0.0", Tags: []string{"eda"}},
		}
	case "3":
		config.Vandor.Architecture = "minimal"
		config.Vpkg = []VpkgItem{
			{Name: "redis-cache", Version: "1.2.0", Tags: []string{"full-backend", "eda", "minimal"}},
		}
	default:
		config.Vandor.Architecture = "minimal"
		config.Vpkg = []VpkgItem{
			{Name: "redis-cache", Version: "1.2.0", Tags: []string{"full-backend", "eda", "minimal"}},
		}
	}

	// Ask if user wants full project setup
	fmt.Print("\nDo you want to create a full project setup? [y/N]: ")
	setupChoice, _ := reader.ReadString('\n')
	createFullSetup := strings.ToLower(strings.TrimSpace(setupChoice)) == "y"

	var useGitHubTemplate bool
	if createFullSetup {
		fmt.Print("Do you want to use GitHub templates (recommended)? [Y/n]: ")
		githubChoice, _ := reader.ReadString('\n')
		useGitHubTemplate = strings.ToLower(strings.TrimSpace(githubChoice)) != "n"
	}

	// Write vandor-config.yaml
	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile("vandor-config.yaml", yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	fmt.Println("\n✅ vandor-config.yaml created successfully!")

	if createFullSetup {
		if useGitHubTemplate {
			if err := createProjectFromGitHubTemplate(config); err != nil {
				fmt.Printf("❌ Failed to create project from GitHub template: %v\n", err)
				fmt.Println("🔄 Falling back to local template creation...")
				if err := createFullProjectSetup(config); err != nil {
					return fmt.Errorf("failed to create project setup: %v", err)
				}
			} else {
				fmt.Println("✅ Project created from GitHub template successfully!")
			}
		} else {
			if err := createFullProjectSetup(config); err != nil {
				return fmt.Errorf("failed to create full project setup: %v", err)
			}
			fmt.Println("✅ Full project setup created successfully!")
		}
	}

	return nil
}

// createProjectFromGitHubTemplate creates a project by cloning from GitHub template
func createProjectFromGitHubTemplate(config VandorConfig) error {
	// Check if git is installed
	if err := checkGitInstalled(); err != nil {
		return fmt.Errorf("git is not installed or not available in PATH")
	}

	// Get template repositories
	templates := getTemplateRepositories()
	
	// Get the repository URL for the selected architecture
	repoURL, exists := templates[config.Vandor.Architecture]
	if !exists {
		return fmt.Errorf("no template repository configured for architecture: %s", config.Vandor.Architecture)
	}

	// Check if current directory is empty (except for vandor-config.yaml)
	files, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read current directory: %v", err)
	}

	// Count non-hidden files (excluding vandor-config.yaml)
	nonHiddenFiles := 0
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") && file.Name() != "vandor-config.yaml" {
			nonHiddenFiles++
		}
	}

	// If directory is not empty, create project in subdirectory
	var targetDir string
	if nonHiddenFiles > 0 {
		if config.Project.Name == "" {
			targetDir = "vandor-project"
		} else {
			targetDir = config.Project.Name
		}
		fmt.Printf("📁 Current directory is not empty. Creating project in '%s/' subdirectory...\n", targetDir)
	} else {
		// Clone directly to current directory
		targetDir = "."
		fmt.Printf("📁 Creating project in current directory...\n")
	}

	// Clone the template
	if targetDir == "." {
		// For current directory, clone to temp and move files
		tempDir := ".vandor-temp"
		if err := cloneTemplate(repoURL, tempDir, config.Project.Name); err != nil {
			return err
		}
		
		// Move files from temp to current directory
		if err := moveTemplateFiles(tempDir, "."); err != nil {
			os.RemoveAll(tempDir) // Clean up on error
			return fmt.Errorf("failed to move template files: %v", err)
		}
		
		// Clean up temp directory
		os.RemoveAll(tempDir)
	} else {
		// Clone directly to subdirectory
		if err := cloneTemplate(repoURL, targetDir, config.Project.Name); err != nil {
			return err
		}
		
		// Move vandor-config.yaml to the project directory
		if err := os.Rename("vandor-config.yaml", filepath.Join(targetDir, "vandor-config.yaml")); err != nil {
			fmt.Printf("Warning: Could not move vandor-config.yaml to project directory: %v\n", err)
		}
	}

	// Update project files with actual configuration
	if err := updateProjectConfiguration(targetDir, config); err != nil {
		fmt.Printf("Warning: Could not update project configuration: %v\n", err)
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
	fmt.Printf("🔧 Updating project configuration...\n")
	
	// Update go.mod if it exists
	goModPath := filepath.Join(projectDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		if err := updateGoMod(goModPath, config.Project.Module); err != nil {
			return fmt.Errorf("failed to update go.mod: %v", err)
		}
	}

	// Update README.md if it exists
	readmePath := filepath.Join(projectDir, "README.md")
	if _, err := os.Stat(readmePath); err == nil {
		if err := updateReadme(readmePath, config.Project.Name); err != nil {
			fmt.Printf("Warning: Could not update README.md: %v\n", err)
		}
	}

	return nil
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

func createFullProjectSetup(config VandorConfig) error {
	// Create basic directory structure
	dirs := []string{
		"cmd/app",
		"internal/core/domain",
		"internal/core/usecase",
		"internal/core/service",
		"internal/core/model",
		"internal/infrastructure/db",
		"internal/delivery/http",
		"internal/vpkg",
		"config",
		"database/schema",
		"scripts",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create go.mod if it doesn't exist
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n", config.Project.Module)
		if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
			return fmt.Errorf("failed to create go.mod: %v", err)
		}
	}

	// Create basic main.go
	mainGoPath := filepath.Join("cmd", "app", "main.go")
	mainGoContent := fmt.Sprintf(`package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Starting %s...")
	log.Println("Application initialized successfully!")
}
`, config.Project.Name)

	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %v", err)
	}

	// Create basic config file
	configPath := filepath.Join("config", "config.yaml.example")
	configContent := `app:
  name: ` + config.Project.Name + `
  port: 8080
  debug: true

database:
  host: localhost
  port: 5432
  name: ` + strings.ReplaceAll(config.Project.Name, "-", "_") + `
  user: postgres
  password: postgres
  ssl_mode: disable
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config.yaml.example: %v", err)
	}

	return nil
}

