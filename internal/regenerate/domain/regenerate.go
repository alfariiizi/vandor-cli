package domain

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/alfariiizi/vandor-cli/internal/utils"
)

//go:embed domain_registry.tmpl
var domainRegistryTemplate string

type DomainInfo struct {
	Name        string // e.g., "User"
	LowerName   string // e.g., "user"
	PackageName string // e.g., "domain"
}

type DomainRegistryData struct {
	Domains []DomainInfo
}

func RegenerateDomain() error {
	domainsPath := filepath.Join("internal", "core", "domain", "model")

	// Discover all domain files
	domains, err := discoverDomains(domainsPath)
	if err != nil {
		return fmt.Errorf("error discovering domains: %w", err)
	}

	if len(domains) == 0 {
		fmt.Println("No domains found in", domainsPath)
		return nil
	}

	// Sort domains by name for consistent output
	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Name < domains[j].Name
	})

	data := DomainRegistryData{
		Domains: domains,
	}

	// Generate domain registry
	registryPath := filepath.Join("internal", "core", "domain", "domain.go")
	if err := generateDomainRegistry(data, registryPath); err != nil {
		return fmt.Errorf("error generating domain registry: %w", err)
	}

	fmt.Printf("Domain registry generated successfully at %s\n", registryPath)
	fmt.Printf("Registered %d domains:\n", len(domains))
	for _, domain := range domains {
		fmt.Printf("  - %s\n", domain.Name)
	}
	return nil
}

func discoverDomains(domainsPath string) ([]DomainInfo, error) {
	var domains []DomainInfo

	// Walk through domain model directory
	err := filepath.Walk(domainsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Look for domain struct and NewXXXDomain function
		domainInfo := extractDomainInfo(node, path)
		if domainInfo != nil {
			domains = append(domains, *domainInfo)
		}

		return nil
	})

	return domains, err
}

func extractDomainInfo(node *ast.File, filePath string) *DomainInfo {
	var domainStruct *ast.TypeSpec
	var newDomainFunc *ast.FuncDecl

	// Look for domain struct and New function
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			// Look for struct type
			if structType, ok := x.Type.(*ast.StructType); ok {
				// Check if it embeds db.SomeType
				if hasDBEmbedding(structType) {
					domainStruct = x
				}
			}
		case *ast.FuncDecl:
			// Look for New*Domain function
			if x.Name != nil && strings.HasPrefix(x.Name.Name, "New") && strings.HasSuffix(x.Name.Name, "Domain") {
				newDomainFunc = x
			}
		}
		return true
	})

	// If we found both struct and New function, extract domain info
	if domainStruct != nil && newDomainFunc != nil {
		domainName := domainStruct.Name.Name
		return &DomainInfo{
			Name:        domainName,
			LowerName:   strings.ToLower(domainName),
			PackageName: "domain",
		}
	}

	return nil
}

func hasDBEmbedding(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		// Check for embedded field like *db.User
		if len(field.Names) == 0 { // Embedded field has no names
			if starExpr, ok := field.Type.(*ast.StarExpr); ok {
				if selExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "db" {
						return true
					}
				}
			}
		}
	}
	return false
}

type TemplateData struct {
	ModuleName string
	Domains    []DomainInfo
}

func generateDomainRegistry(data DomainRegistryData, outputPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Parse embedded template
	tmpl, err := template.New("domain_registry").Parse(domainRegistryTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	templateData := TemplateData{
		ModuleName: utils.GetModuleName(),
		Domains:    data.Domains,
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	// Execute template
	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
