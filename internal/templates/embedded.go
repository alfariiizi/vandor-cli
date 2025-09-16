package templates

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alfariiizi/vandor-cli/internal/utils"
)

//go:embed usecase/*.tmpl domain/*.tmpl service/*.tmpl job/*.tmpl handler/*.tmpl scheduler/*.tmpl seed/*.tmpl enum/*.tmpl
var templateFS embed.FS

// TemplateData contains common template variables
type TemplateData struct {
	ModuleName string // e.g., "github.com/user/project"
	Name       string // e.g., "CreateUser"
	Receiver   string // e.g., "createUser"
	NameSnake  string // e.g., "create_user"
	PathName   string // e.g., "create-user" (for URLs)
	Group      string // e.g., "user" (for services/handlers)
	Method     string // e.g., "POST" (for handlers)
}

// TemplateConfig defines configuration for template generation
type TemplateConfig struct {
	ComponentType string            // "usecase", "domain", "service", etc.
	Name          string            // Component name
	Group         string            // Group name (for services/handlers)
	Method        string            // HTTP method (for handlers)
	OutputDir     string            // Output directory
	ExtraData     map[string]string // Additional template variables
}

// TemplateManager handles template operations
type TemplateManager struct {
	fs embed.FS
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		fs: templateFS,
	}
}

// Generate generates code from templates
func (tm *TemplateManager) Generate(config TemplateConfig) error {
	if config.Name == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	// Prepare template data
	data, err := tm.prepareTemplateData(config)
	if err != nil {
		return fmt.Errorf("failed to prepare template data: %w", err)
	}

	// Get template path
	templatePath := fmt.Sprintf("%s/%s.go.tmpl", config.ComponentType, config.ComponentType)

	// Check if template exists
	if !tm.templateExists(templatePath) {
		return fmt.Errorf("template not found: %s", templatePath)
	}

	// Generate the file
	return tm.generateFile(templatePath, config, data)
}

// prepareTemplateData prepares the data for template rendering
func (tm *TemplateManager) prepareTemplateData(config TemplateConfig) (TemplateData, error) {
	moduleName, err := utils.DetectGoModule()
	if err != nil {
		return TemplateData{}, fmt.Errorf("failed to detect Go module: %w", err)
	}

	// Ensure proper capitalization for type names
	name := utils.ToPascalCase(config.Name)
	receiver := utils.ToCamelCase(config.Name)
	nameSnake := utils.ToSnakeCase(config.Name)
	pathName := utils.ToKebabCase(config.Name)

	data := TemplateData{
		ModuleName: moduleName,
		Name:       name,
		Receiver:   receiver,
		NameSnake:  nameSnake,
		PathName:   pathName,
		Group:      config.Group,
		Method:     strings.ToUpper(config.Method),
	}

	return data, nil
}

// templateExists checks if a template file exists
func (tm *TemplateManager) templateExists(templatePath string) bool {
	_, err := tm.fs.ReadFile(templatePath)
	return err == nil
}

// generateFile generates a file from template
func (tm *TemplateManager) generateFile(templatePath string, config TemplateConfig, data TemplateData) error {
	// Read template content
	content, err := tm.fs.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Parse template
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Debug: Print the receiver value
	fmt.Printf("🔍 Debug - Receiver: '%s', Name: '%s'\n", data.Receiver, data.Name)

	// Determine output file path
	outputPath, err := tm.getOutputPath(config, data)
	if err != nil {
		return fmt.Errorf("failed to determine output path: %w", err)
	}

	fmt.Printf("🔍 Debug - Output path: '%s'\n", outputPath)

	// Ensure output directory exists
	if errDir := os.MkdirAll(filepath.Dir(outputPath), 0755); errDir != nil {
		return fmt.Errorf("failed to create output directory: %w", errDir)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("✅ Generated %s: %s\n", config.ComponentType, outputPath)
	return nil
}

// getOutputPath determines the output file path
func (tm *TemplateManager) getOutputPath(config TemplateConfig, data TemplateData) (string, error) {
	// Use PascalCase for filename (more readable and conventional)
	filename := data.Name + ".go"

	if config.OutputDir != "" {
		return filepath.Join(config.OutputDir, filename), nil
	}

	// Default paths based on component type
	switch config.ComponentType {
	case "usecase":
		return filepath.Join("internal", "core", "usecase", filename), nil
	case "domain":
		return filepath.Join("internal", "core", "domain", filename), nil
	case "service":
		if config.Group == "" {
			return "", fmt.Errorf("group is required for service generation")
		}
		return filepath.Join("internal", "core", "service", config.Group, filename), nil
	case "job":
		return filepath.Join("internal", "core", "job", filename), nil
	case "handler":
		if config.Group == "" {
			return "", fmt.Errorf("group is required for handler generation")
		}
		return filepath.Join("internal", "delivery", "http", "handler", config.Group, filename), nil
	case "scheduler":
		return filepath.Join("internal", "core", "scheduler", filename), nil
	case "seed":
		return filepath.Join("internal", "infrastructure", "seed", filename), nil
	case "enum":
		return filepath.Join("internal", "core", "enum", filename), nil
	default:
		return "", fmt.Errorf("unknown component type: %s", config.ComponentType)
	}
}

// ListAvailableTemplates returns a list of available templates
func (tm *TemplateManager) ListAvailableTemplates() ([]string, error) {
	var templates []string

	entries, err := tm.fs.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Check if the directory contains a template file
			templateFile := fmt.Sprintf("%s/%s.go.tmpl", entry.Name(), entry.Name())
			if tm.templateExists(templateFile) {
				templates = append(templates, entry.Name())
			}
		}
	}

	return templates, nil
}

// ValidateComponentType checks if a component type is supported
func (tm *TemplateManager) ValidateComponentType(componentType string) error {
	templates, err := tm.ListAvailableTemplates()
	if err != nil {
		return err
	}

	for _, template := range templates {
		if template == componentType {
			return nil
		}
	}

	return fmt.Errorf("unsupported component type: %s. Available: %s", componentType, strings.Join(templates, ", "))
}
