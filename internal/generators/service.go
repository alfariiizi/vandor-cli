package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/vandordev/vandor/internal/utils"
)

type ServiceData struct {
	ModuleName  string // e.g., "github.com/vandordev/vandor"
	Name        string // e.g., "CreateUser"
	ServiceName string // e.g., "create_user"
	Receiver    string // e.g., "createUser"
}

func GenerateService(serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	// Capitalize first letter for type name
	serviceName = strings.ToUpper(serviceName[:1]) + serviceName[1:]

	data := ServiceData{
		ModuleName:  utils.GetModuleName(),
		Name:        serviceName,
		ServiceName: strings.ToLower(serviceName), // snake_case for package name
		Receiver:    strings.ToLower(serviceName[:1]) + serviceName[1:], // camelCase for receiver
	}

	// Check if service already exists
	servicePath := filepath.Join("internal", "core", "service", strings.ToLower(serviceName)+".go")
	if _, err := os.Stat(servicePath); err == nil {
		return fmt.Errorf("service %s already exists at %s", serviceName, servicePath)
	}

	// Generate service file
	f := generateServiceFile(data)

	// Ensure directory exists
	dir := filepath.Dir(servicePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := f.Save(servicePath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("Service %s created successfully at %s\n", serviceName, servicePath)
	return nil
}

func generateServiceFile(data ServiceData) *jen.File {
	f := jen.NewFile(data.ServiceName + "_service")

	// Add imports
	f.ImportName("context", "context")
	f.ImportName(data.ModuleName+"/internal/core/domain", "domain_entries")
	f.ImportName(data.ModuleName+"/internal/core/model", "model")
	f.ImportName(data.ModuleName+"/internal/core/usecase", "usecase")
	f.ImportName(data.ModuleName+"/internal/infrastructure/db", "db")
	f.ImportName(data.ModuleName+"/internal/pkg/logger", "logger")
	f.ImportName(data.ModuleName+"/internal/pkg/validator", "validator")

	// Generate Input struct
	f.Type().Id(data.Name + "Input").Struct(
		jen.Comment("TODO: Define fields"),
	)

	// Generate Output struct
	f.Type().Id(data.Name + "Output").Struct(
		jen.Comment("TODO: Define fields"),
	)

	// Generate service type alias
	f.Type().Id(data.Name).Qual("model", "Service").Types(
		jen.Id(data.Name + "Input"),
		jen.Id(data.Name + "Output"),
	)

	// Generate implementation struct
	f.Type().Id(data.Receiver).Struct(
		jen.Id("domain").Op("*").Qual("domain_entries", "Domain"),
		jen.Id("client").Op("*").Qual("db", "Client"),
		jen.Id("usecase").Op("*").Qual("usecase", "Usecases"),
		jen.Id("validator").Qual("validator", "Validator"),
	)

	// Generate constructor
	f.Func().Id("New" + data.Name).Params(
		jen.Id("domain").Op("*").Qual("domain_entries", "Domain"),
		jen.Id("client").Op("*").Qual("db", "Client"),
		jen.Id("usecase").Op("*").Qual("usecase", "Usecases"),
		jen.Id("validator").Qual("validator", "Validator"),
	).Id(data.Name).Block(
		jen.Return(jen.Op("&").Id(data.Receiver).Values(jen.Dict{
			jen.Id("domain"):    jen.Id("domain"),
			jen.Id("client"):    jen.Id("client"),
			jen.Id("usecase"):   jen.Id("usecase"),
			jen.Id("validator"): jen.Id("validator"),
		})),
	)

	// Generate Validate method
	f.Func().Params(jen.Id("s").Op("*").Id(data.Receiver)).Id("Validate").Params(
		jen.Id("input").Id(data.Name + "Input"),
	).Error().Block(
		jen.Return(jen.Id("s").Dot("validator").Dot("Validate").Call(jen.Id("input"))),
	)

	// Generate Execute method
	f.Func().Params(jen.Id("s").Op("*").Id(data.Receiver)).Id("Execute").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("input").Id(data.Name + "Input"),
	).Params(jen.Op("*").Id(data.Name + "Output"), jen.Error()).Block(
		jen.If(jen.Id("err").Op(":=").Id("s").Dot("Validate").Call(jen.Id("input")), jen.Id("err").Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Id("err")),
		),
		jen.Line(),
		jen.List(jen.Id("res"), jen.Id("err")).Op(":=").Id("s").Dot("Process").Call(jen.Id("ctx"), jen.Id("input")),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Qual("logger", "Get").Call().Dot("Error").Call().
				Dot("Str").Call(jen.Lit("service"), jen.Lit(data.Name)).
				Dot("Str").Call(jen.Lit("method"), jen.Lit("Process")).
				Dot("Err").Call(jen.Id("err")).
				Dot("Msg").Call(jen.Lit("failed to execute service")),
			jen.Return(jen.Nil(), jen.Id("err")),
		),
		jen.Line(),
		jen.If(jen.Id("err").Op(":=").Id("s").Dot("Observer").Call(jen.Id("ctx"), jen.Id("input"), jen.Id("res")), jen.Id("err").Op("!=").Nil()).Block(
			jen.Qual("logger", "Get").Call().Dot("Error").Call().
				Dot("Str").Call(jen.Lit("service"), jen.Lit(data.Name)).
				Dot("Str").Call(jen.Lit("method"), jen.Lit("Observer")).
				Dot("Err").Call(jen.Id("err")).
				Dot("Msg").Call(jen.Lit("failed to execute observer")),
		),
		jen.Line(),
		jen.Return(jen.Id("res"), jen.Nil()),
	)

	// Generate Observer method
	f.Func().Params(jen.Id("s").Op("*").Id(data.Receiver)).Id("Observer").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("input").Id(data.Name + "Input"),
		jen.Id("output").Op("*").Id(data.Name + "Output"),
	).Error().Block(
		jen.Comment("TODO: Implement observer logic"),
		jen.Comment("This is optional. You can leave this blank if not needed."),
		jen.Line(),
		jen.Return(jen.Nil()),
	)

	// Generate Process method
	f.Func().Params(jen.Id("s").Op("*").Id(data.Receiver)).Id("Process").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("input").Id(data.Name + "Input"),
	).Params(jen.Op("*").Id(data.Name + "Output"), jen.Error()).Block(
		jen.Comment("TODO: Implement logic"),
		jen.Line(),
		jen.Return(jen.Nil(), jen.Nil()),
	)

	return f
}