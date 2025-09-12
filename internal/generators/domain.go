package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/alfariiizi/vandor-cli/internal/utils"
)

type DomainData struct {
	ModuleName  string // e.g., "github.com/alfariiizi/vandor-cli"
	Name        string // e.g., "User"
	LowerName   string // e.g., "user"
	PackageName string // e.g., "domain"
}

func GenerateDomain(domainName string) error {
	if domainName == "" {
		return fmt.Errorf("domain name cannot be empty")
	}

	// Capitalize first letter
	domainName = strings.ToUpper(domainName[:1]) + domainName[1:]

	data := DomainData{
		ModuleName:  utils.GetModuleName(),
		Name:        domainName,
		LowerName:   strings.ToLower(domainName),
		PackageName: "domain",
	}

	// Check if domain already exists
	domainPath := filepath.Join("internal", "core", "domain", "model", strings.ToLower(domainName)+".go")
	if _, err := os.Stat(domainPath); err == nil {
		return fmt.Errorf("domain %s already exists at %s", domainName, domainPath)
	}

	// Generate domain file
	f := generateDomainFile(data)

	// Ensure directory exists
	dir := filepath.Dir(domainPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := f.Save(domainPath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("Domain %s created successfully at %s\n", domainName, domainPath)
	return nil
}

func generateDomainFile(data DomainData) *jen.File {
	f := jen.NewFile(data.PackageName)

	// Add imports
	f.ImportName("fmt", "fmt")
	f.ImportName(data.ModuleName+"/internal/core/domain/builder", "domain_builder")
	f.ImportName(data.ModuleName+"/internal/infrastructure/db", "db")

	// Generate struct
	f.Type().Id(data.Name).Struct(
		jen.Op("*").Qual("db", data.Name),
		jen.Id("client").Op("*").Qual("db", "Client"),
	)

	// Generate constructor function
	f.Func().Id("New"+data.Name+"Domain").
		Params(jen.Id("client").Op("*").Qual("db", "Client")).
		Qual("domain_builder", "Domain").Types(
			jen.Op("*").Qual("db", data.Name),
			jen.Op("*").Id(data.Name),
		).Block(
		jen.Return(
			jen.Qual("domain_builder", "NewDomain").Call(
				jen.Func().Params(
					jen.Id("e").Op("*").Qual("db", data.Name),
					jen.Id("c").Op("*").Qual("db", "Client"),
				).Op("*").Id(data.Name).Block(
					jen.Return(jen.Op("&").Id(data.Name).Values(
						jen.Dict{
							jen.Id(data.Name): jen.Id("e"),
							jen.Id("client"):  jen.Id("c"),
						},
					)),
				),
				jen.Id("client"),
			),
		),
	)

	// Add comment for TODO methods
	f.Comment("TODO: Add your domain methods here")
	f.Comment("Example methods:")
	f.Line()

	// Generate String method
	f.Func().Params(jen.Id(data.LowerName).Op("*").Id(data.Name)).Id("String").Params().String().Block(
		jen.Return(jen.Qual("fmt", "Sprintf").Call(
			jen.Lit(data.Name+"{ID: %d}"),
			jen.Id(data.LowerName).Dot("ID"),
		)),
	)

	// Add comment for additional methods
	f.Line()
	f.Comment("Add more business logic methods as needed")

	return f
}