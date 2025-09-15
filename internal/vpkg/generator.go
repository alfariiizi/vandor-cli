package vpkg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alfariiizi/vandor-cli/internal/utils"
)

// TemplateGenerator converts Go files to template files
type TemplateGenerator struct {
	patterns []PatternRule
}

// PatternRule defines a pattern replacement rule
type PatternRule struct {
	Name        string         // Rule name for debugging
	Pattern     *regexp.Regexp // Regex pattern to match
	Replacement string         // Template replacement
	Description string         // Human-readable description
}

// NewTemplateGenerator creates a new template generator with default rules
func NewTemplateGenerator() *TemplateGenerator {
	generator := &TemplateGenerator{}
	generator.setupDefaultPatterns()
	return generator
}

// Generate converts Go files to template files
func (g *TemplateGenerator) Generate(opts GenerateOptions) error {
	inputInfo, err := os.Stat(opts.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read input path: %w", err)
	}

	if inputInfo.IsDir() {
		return g.generateFromDirectory(opts)
	} else {
		return g.generateFromFile(opts)
	}
}

// generateFromDirectory processes all .go files in a directory
func (g *TemplateGenerator) generateFromDirectory(opts GenerateOptions) error {
	if opts.Verbose {
		fmt.Printf("🔍 Scanning directory: %s\n", opts.InputPath)
	}

	// Ensure output directory exists
	if !opts.DryRun {
		if err := os.MkdirAll(opts.OutputPath, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	processedFiles := 0
	err := filepath.WalkDir(opts.InputPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(opts.InputPath, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Generate output path
		outputFile := filepath.Join(opts.OutputPath, relPath+".tmpl")

		// Process file
		fileOpts := GenerateOptions{
			InputPath:   path,
			OutputPath:  outputFile,
			PackageName: opts.PackageName,
			DryRun:      opts.DryRun,
			Verbose:     opts.Verbose,
		}

		if err := g.generateFromFile(fileOpts); err != nil {
			return fmt.Errorf("failed to process %s: %w", path, err)
		}

		processedFiles++
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("✅ Processed %d Go files\n", processedFiles)
	if opts.DryRun {
		fmt.Printf("💡 Use --dry-run=false to actually generate files\n")
	}

	return nil
}

// generateFromFile processes a single .go file
func (g *TemplateGenerator) generateFromFile(opts GenerateOptions) error {
	if opts.Verbose {
		fmt.Printf("📝 Processing: %s -> %s\n", opts.InputPath, opts.OutputPath)
	}

	// Read input file
	content, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Analyze Go file structure
	analysis, err := g.analyzeGoFile(opts.InputPath, content)
	if err != nil {
		return fmt.Errorf("failed to analyze Go file: %w", err)
	}

	// Convert to template
	templateContent, err := g.convertToTemplate(string(content), analysis, opts)
	if err != nil {
		return fmt.Errorf("failed to convert to template: %w", err)
	}

	if opts.DryRun {
		fmt.Printf("📄 Would generate: %s\n", opts.OutputPath)
		if opts.Verbose {
			fmt.Printf("━━━ Template Preview ━━━\n")
			lines := strings.Split(templateContent, "\n")
			for i, line := range lines {
				if i >= 20 { // Show first 20 lines
					fmt.Printf("... (%d more lines)\n", len(lines)-20)
					break
				}
				fmt.Printf("%3d │ %s\n", i+1, line)
			}
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━\n")
		}
		return nil
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(opts.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write template file
	if err := os.WriteFile(opts.OutputPath, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("✅ Generated: %s\n", opts.OutputPath)
	return nil
}

// GoFileAnalysis contains information about a Go file
type GoFileAnalysis struct {
	PackageName   string
	Imports       []string
	StructNames   []string
	FunctionNames []string
	ConstNames    []string
	VarNames      []string
	TypeNames     []string
}

// analyzeGoFile parses a Go file and extracts structural information
func (g *TemplateGenerator) analyzeGoFile(filePath string, content []byte) (*GoFileAnalysis, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	analysis := &GoFileAnalysis{
		PackageName: node.Name.Name,
	}

	// Walk the AST to extract information
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ImportSpec:
			if x.Path != nil {
				analysis.Imports = append(analysis.Imports, strings.Trim(x.Path.Value, `"`))
			}

		case *ast.TypeSpec:
			analysis.TypeNames = append(analysis.TypeNames, x.Name.Name)
			if _, ok := x.Type.(*ast.StructType); ok {
				analysis.StructNames = append(analysis.StructNames, x.Name.Name)
			}

		case *ast.FuncDecl:
			if x.Name != nil {
				analysis.FunctionNames = append(analysis.FunctionNames, x.Name.Name)
			}

		case *ast.GenDecl:
			for _, spec := range x.Specs {
				if s := spec.(*ast.ValueSpec); s != nil {
					for _, name := range s.Names {
						if x.Tok == token.CONST {
							analysis.ConstNames = append(analysis.ConstNames, name.Name)
						} else if x.Tok == token.VAR {
							analysis.VarNames = append(analysis.VarNames, name.Name)
						}
					}
				}
			}
		}
		return true
	})

	return analysis, nil
}

// convertToTemplate applies pattern rules to convert Go code to template
func (g *TemplateGenerator) convertToTemplate(content string, analysis *GoFileAnalysis, opts GenerateOptions) (string, error) {
	result := content

	// Determine package name for templates
	packageName := opts.PackageName
	if packageName == "" {
		packageName = analysis.PackageName
	}

	// Sanitize package name for Go identifier
	packageIdent := utils.ToGoIdentifier(strings.ReplaceAll(packageName, "-", ""))

	if opts.Verbose {
		fmt.Printf("🔧 Using package context: %s -> %s\n", packageName, packageIdent)
		fmt.Printf("📊 Analysis: %d structs, %d functions, %d imports\n",
			len(analysis.StructNames), len(analysis.FunctionNames), len(analysis.Imports))
	}

	// Apply pattern rules
	for _, rule := range g.patterns {
		if opts.Verbose && rule.Pattern.MatchString(result) {
			fmt.Printf("🔄 Applying rule: %s\n", rule.Name)
		}
		result = rule.Pattern.ReplaceAllString(result, rule.Replacement)
	}

	// Apply context-specific replacements
	result = g.applyContextualReplacements(result, analysis, packageIdent, opts)

	return result, nil
}

// setupDefaultPatterns initializes the default pattern replacement rules
func (g *TemplateGenerator) setupDefaultPatterns() {
	g.patterns = []PatternRule{
		// Package declaration
		{
			Name:        "package-declaration",
			Pattern:     regexp.MustCompile(`^package\s+\w+`),
			Replacement: `package {{.Package}}`,
			Description: "Replace package declaration",
		},

		// Import paths that look like project modules
		{
			Name:        "local-imports",
			Pattern:     regexp.MustCompile(`"([^"]*)(internal|pkg|cmd|api)/([^"]*)"`),
			Replacement: `"{{.Module}}/$1$2/$3"`,
			Description: "Replace local import paths",
		},

		// Third-party imports should not be replaced
		{
			Name:        "preserve-third-party",
			Pattern:     regexp.MustCompile(`"github\.com/redis/go-redis/v9"`),
			Replacement: `"github.com/redis/go-redis/v9"`,
			Description: "Preserve third-party imports",
		},

		// Go module replacement in import paths (be more specific)
		{
			Name:        "module-imports",
			Pattern:     regexp.MustCompile(`"github\.com/[^/]+/[^/]+/(internal|pkg|cmd|api)/([^"]*)"`),
			Replacement: `"{{.Module}}/$1/$2"`,
			Description: "Replace module-based imports",
		},

		// Common naming patterns for title case
		{
			Name:        "title-case-patterns",
			Pattern:     regexp.MustCompile(`\b([A-Z][a-zA-Z]*)(Client|Service|Manager|Handler|Config|Module)\b`),
			Replacement: `{{Pascal .Pkg}}$2`,
			Description: "Replace title case type names",
		},

		// Function constructor patterns
		{
			Name:        "constructor-functions",
			Pattern:     regexp.MustCompile(`\bfunc\s+New([A-Z][a-zA-Z]*)\(`),
			Replacement: `func New{{Pascal .Pkg}}(`,
			Description: "Replace constructor function names",
		},

		// Variable naming patterns
		{
			Name:        "camel-case-variables",
			Pattern:     regexp.MustCompile(`\b([a-z][a-zA-Z]*)(Client|Service|Manager|Config)\b`),
			Replacement: `{{Camel .Pkg}}$2`,
			Description: "Replace camelCase variable names",
		},

		// Constants with package prefix
		{
			Name:        "constant-names",
			Pattern:     regexp.MustCompile(`\bconst\s+([A-Z][A-Z_]*[A-Z])\b`),
			Replacement: `const {{Upper .Pkg}}_$1`,
			Description: "Replace constant names with package prefix",
		},

		// Fx module string should use template variable
		{
			Name:        "fx-module-string",
			Pattern:     regexp.MustCompile(`fx\.Module\("([^"]+)",`),
			Replacement: `fx.Module("{{.Package}}",`,
			Description: "Replace fx.Module string with template variable",
		},

		// Function names that contain package references
		{
			Name:        "function-references",
			Pattern:     regexp.MustCompile(`\bfx\.Provide\(New([A-Z][a-zA-Z]*)\)`),
			Replacement: `fx.Provide(New{{Pascal .Pkg}})`,
			Description: "Replace function references in fx.Provide",
		},
	}
}

// applyContextualReplacements applies specific replacements based on file analysis
func (g *TemplateGenerator) applyContextualReplacements(content string, analysis *GoFileAnalysis, packageIdent string, opts GenerateOptions) string {
	result := content

	// Replace specific struct names found in the file
	for _, structName := range analysis.StructNames {
		// Only replace if it looks like it could be a main type
		if strings.Contains(strings.ToLower(structName), strings.ToLower(opts.PackageName)) {

			// Don't replace if it's inside certain contexts (like redis.NewClient)
			pattern := regexp.MustCompile(fmt.Sprintf(`\b%s\b(?!\.)`, regexp.QuoteMeta(structName)))
			replacement := "{{Pascal .Pkg}}"

			if opts.Verbose {
				fmt.Printf("🔄 Replacing struct: %s -> %s\n", structName, replacement)
			}

			result = pattern.ReplaceAllString(result, replacement)
		}
	}

	// Add template header comment
	header := fmt.Sprintf(`// Package {{.Package}} provides {{.Description}}
// Generated from %s
//
// Template Variables:
// - {{.Package}} = %s (Go identifier)
// - {{.Pkg}} = %s (original name)
// - {{.Title}} = {{.Title}}
// - {{.ImportPath}} = {{.ImportPath}}
// - {{.Module}} = {{.Module}}
//
`, filepath.Base(opts.InputPath), packageIdent, opts.PackageName)

	result = header + result

	return result
}
