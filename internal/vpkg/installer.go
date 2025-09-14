package vpkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alfariiizi/vandor-cli/internal/utils"
)

// Installer handles package installation and removal
type Installer struct {
	registryClient *RegistryClient
}

// NewInstaller creates a new package installer
func NewInstaller(registryURL string) *Installer {
	return &Installer{
		registryClient: NewRegistryClient(registryURL),
	}
}

// Install installs a package with the given options
func (i *Installer) Install(packageName string, opts InstallOptions) error {
	// Parse package name and version
	name, version := parsePackageSpec(packageName)
	if version == "" {
		version = opts.Version
	}

	// Find package in repositories
	packageWithRepo, err := i.registryClient.FindPackage(name)
	if err != nil {
		return fmt.Errorf("failed to find package: %w", err)
	}

	pkg := packageWithRepo.Package

	// Determine destination path - always relative to project root
	projectRoot, err := i.findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	destPath := opts.Dest
	if destPath == "" {
		destPath = pkg.Destination
	}
	if destPath == "" {
		destPath = fmt.Sprintf("internal/vpkg/%s", name)
	}

	// Make destination path absolute from project root
	if !filepath.IsAbs(destPath) {
		destPath = filepath.Join(projectRoot, destPath)
	}

	// Check if package already exists
	if !opts.Force && i.packageExists(destPath) {
		return fmt.Errorf("package already exists at %s (use --force to overwrite)", destPath)
	}

	// Prepare template context
	ctx, err := i.prepareTemplateContext(name, &pkg, destPath)
	if err != nil {
		return fmt.Errorf("failed to prepare template context: %w", err)
	}

	// Create destination directory
	if !opts.DryRun {
		if errDir := os.MkdirAll(destPath, 0755); errDir != nil {
			return fmt.Errorf("failed to create destination directory: %w", errDir)
		}
	}

	// Discover and install templates from the templates directory
	templateFiles, err := i.registryClient.DiscoverTemplateFiles(packageWithRepo, pkg.Templates)
	if err != nil {
		return fmt.Errorf("failed to discover template files: %w", err)
	}

	if len(templateFiles) == 0 {
		return fmt.Errorf("no template files found in %s", pkg.Templates)
	}

	// Install each discovered template
	for _, templatePath := range templateFiles {
		if err := i.installTemplate(packageWithRepo, templatePath, destPath, ctx, opts); err != nil {
			return fmt.Errorf("failed to install template %s: %w", templatePath, err)
		}
	}

	// Write metadata file
	if !opts.DryRun {
		if err := i.writeInstalledMeta(destPath, name, version, &pkg); err != nil {
			return fmt.Errorf("failed to write package metadata: %w", err)
		}
	}

	// Generate usage receipt
	if !opts.DryRun {
		i.printUsageReceipt(name, &pkg, ctx)
	}

	return nil
}

// Remove removes an installed package
func (i *Installer) Remove(packageName string, backup bool) error {
	// Find installed package
	installedPath := i.findInstalledPackage(packageName)
	if installedPath == "" {
		return fmt.Errorf("package %s is not installed", packageName)
	}

	// Create backup if requested
	if backup {
		backupPath := installedPath + ".backup." + time.Now().Format("20060102-150405")
		if err := os.Rename(installedPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("Package backed up to: %s\n", backupPath)
		return nil
	}

	// Remove the package directory
	if err := os.RemoveAll(installedPath); err != nil {
		return fmt.Errorf("failed to remove package: %w", err)
	}

	fmt.Printf("Package %s removed successfully\n", packageName)
	return nil
}

// ListInstalled lists all installed packages
func (i *Installer) ListInstalled() ([]InstalledPackage, error) {
	var installed []InstalledPackage

	vpkgDir := "internal/vpkg"
	if _, err := os.Stat(vpkgDir); os.IsNotExist(err) {
		return installed, nil
	}

	// Walk through vpkg directory
	err := filepath.Walk(vpkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}

		// Look for meta.yaml files
		if info.Name() == "meta.yaml" && strings.Contains(path, "internal/vpkg") {
			pkg, err := i.loadInstalledPackage(path)
			if err == nil {
				installed = append(installed, *pkg)
			}
		}

		return nil
	})

	return installed, err
}

// installTemplate installs a single template file
func (i *Installer) installTemplate(packageWithRepo *PackageWithRepo, templatePath, destPath string, ctx TemplateContext, opts InstallOptions) error {
	// Build full template file path for fetching
	// templatePath is relative like "redis.go.tmpl", we need "packages/redis-cache/templates/redis.go.tmpl"
	fullTemplatePath := fmt.Sprintf("%s/%s", packageWithRepo.Package.Templates, templatePath)

	// Fetch template content from repository
	content, err := i.registryClient.FetchPackageFile(packageWithRepo, fullTemplatePath)
	if err != nil {
		return err
	}

	// Preserve directory structure from templates but remove template extensions
	// redis.go.tmpl -> redis.go
	// someDir/others.go.tmpl -> someDir/others.go
	// cmd/main.go.templ -> cmd/main.go
	relPath := i.removeTemplateExtension(templatePath)

	outputPath := filepath.Join(destPath, relPath)
	outputDir := filepath.Dir(outputPath)

	if opts.DryRun {
		fmt.Printf("Would create: %s\n", outputPath)
		return nil
	}

	// Create directory structure
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", outputDir, err)
	}

	// Render template if it's a template file (supports multiple extensions)
	if i.isTemplateFile(templatePath) {
		outputName := filepath.Base(relPath)
		tmpl, err := template.New(outputName).Funcs(templateFuncs()).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() { _ = file.Close() }()

		if err := tmpl.Execute(file, ctx); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
	} else {
		// Copy file as-is (for non-template files like static assets)
		if err := os.WriteFile(outputPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	fmt.Printf("✓ Created: %s\n", outputPath)
	return nil
}

// prepareTemplateContext creates the template context
func (i *Installer) prepareTemplateContext(packageName string, pkg *Package, destPath string) (TemplateContext, error) {
	module, err := utils.DetectGoModule()
	if err != nil {
		return TemplateContext{}, fmt.Errorf("failed to detect Go module: %w", err)
	}

	parts := strings.Split(packageName, "/")
	namespace := parts[0]
	pkgName := parts[1]

	// Sanitize package name for Go identifier
	packageIdent := utils.ToGoIdentifier(strings.ReplaceAll(pkgName, "-", ""))

	return TemplateContext{
		Module:      module,
		VpkgName:    packageName,
		Namespace:   namespace,
		Pkg:         pkgName,
		Package:     packageIdent,
		PackagePath: destPath,
		Version:     pkg.Version,
		Author:      "", // Author is now at repository level
		Time:        time.Now().Format(time.RFC3339),
		Title:       pkg.Title,
		Description: pkg.Description,
	}, nil
}

// templateFuncs returns template helper functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"Title":   utils.ToTitle,
		"Camel":   utils.ToCamelCase,
		"Snake":   utils.ToSnakeCase,
		"Kebab":   utils.ToKebabCase,
		"Upper":   strings.ToUpper,
		"Lower":   strings.ToLower,
		"Pascal":  utils.ToPascalCase,
		"GoIdent": utils.ToGoIdentifier,
	}
}

// packageExists checks if a package already exists at the given path
func (i *Installer) packageExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// findInstalledPackage finds the installation path of a package
func (i *Installer) findInstalledPackage(packageName string) string {
	// Check default location
	defaultPath := fmt.Sprintf("internal/vpkg/%s", packageName)
	if i.packageExists(defaultPath) {
		return defaultPath
	}

	// TODO: Search all vpkg directories for the package
	return ""
}

// writeInstalledMeta writes metadata for an installed package
func (i *Installer) writeInstalledMeta(destPath, name, version string, pkg *Package) error {
	installed := InstalledPackage{
		Name:        name,
		Version:     version,
		InstalledAt: time.Now(),
		Path:        destPath,
		Type:        pkg.Type,
		Meta:        *pkg,
	}

	data, err := yaml.Marshal(installed)
	if err != nil {
		return err
	}

	metaPath := filepath.Join(destPath, "meta.yaml")
	return os.WriteFile(metaPath, data, 0644)
}

// loadInstalledPackage loads an installed package from its meta.yaml
func (i *Installer) loadInstalledPackage(metaPath string) (*InstalledPackage, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var installed InstalledPackage
	if err := yaml.Unmarshal(data, &installed); err != nil {
		return nil, err
	}

	return &installed, nil
}

// printUsageReceipt prints installation success message and usage instructions
func (i *Installer) printUsageReceipt(packageName string, pkg *Package, ctx TemplateContext) {
	fmt.Printf("\n✅ Package %s installed successfully!\n\n", packageName)

	if pkg.Type == "fx-module" {
		fmt.Printf("📦 Import the package:\n")
		fmt.Printf("   import %s \"%s/%s\"\n\n", ctx.Package, ctx.Module, ctx.PackagePath)

		fmt.Printf("🔧 Wire into Fx application:\n")
		fmt.Printf("   app := fx.New(\n")
		fmt.Printf("       %s.Module,\n", ctx.Package)
		fmt.Printf("       // ... other modules\n")
		fmt.Printf("   )\n\n")

		if len(pkg.Dependencies) > 0 {
			fmt.Printf("📋 Dependencies to add:\n")
			for _, dep := range pkg.Dependencies {
				fmt.Printf("   go get %s\n", dep)
			}
			fmt.Printf("\n")
		}
	} else if pkg.Type == "cli-command" {
		fmt.Printf("🚀 Run as CLI command:\n")
		fmt.Printf("   vandor vpkg exec %s [args]\n\n", packageName)

		fmt.Printf("🔧 Or embed in your application:\n")
		fmt.Printf("   import %s \"%s/%s\"\n", ctx.Package, ctx.Module, ctx.PackagePath)
		fmt.Printf("   // Use %s.Command() to get Cobra command\n\n", ctx.Package)
	}

	fmt.Printf("📖 See README in %s for detailed usage instructions.\n", ctx.PackagePath)
}

// parsePackageSpec parses package@version format
func parsePackageSpec(spec string) (name, version string) {
	if strings.Contains(spec, "@") {
		parts := strings.Split(spec, "@")
		return parts[0], parts[1]
	}
	return spec, ""
}

// removeTemplateExtension removes template extensions from file path
// Supports .tmpl, .templ, .gotmpl extensions
func (i *Installer) removeTemplateExtension(templatePath string) string {
	extensions := []string{".tmpl", ".templ", ".gotmpl"}

	for _, ext := range extensions {
		if strings.HasSuffix(templatePath, ext) {
			return strings.TrimSuffix(templatePath, ext)
		}
	}

	return templatePath // Return as-is if no template extension found
}

// isTemplateFile checks if a file is a template file based on its extension
func (i *Installer) isTemplateFile(filePath string) bool {
	extensions := []string{".tmpl", ".templ", ".gotmpl"}

	for _, ext := range extensions {
		if strings.HasSuffix(filePath, ext) {
			return true
		}
	}

	return false
}

// findProjectRoot finds the project root by looking for vandor-config.yaml or go.mod
func (i *Installer) findProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up the directory tree looking for project markers
	dir := currentDir
	for {
		// Check for vandor-config.yaml (primary marker)
		vandorConfig := filepath.Join(dir, "vandor-config.yaml")
		if _, err := os.Stat(vandorConfig); err == nil {
			return dir, nil
		}

		// Check for go.mod as fallback
		goMod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			return dir, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no project root found (looking for vandor-config.yaml or go.mod)")
}
