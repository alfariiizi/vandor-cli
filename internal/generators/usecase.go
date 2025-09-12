package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/alfariiizi/vandor-cli/internal/utils"
)

type UsecaseData struct {
	ModuleName string // e.g., "github.com/alfariiizi/vandor-cli"
	Name       string // e.g., "CreateUser"
	Receiver   string // e.g., "createUser"
}

func GenerateUsecase(usecaseName string) error {
	if usecaseName == "" {
		return fmt.Errorf("usecase name cannot be empty")
	}

	// Capitalize first letter for type name
	usecaseName = strings.ToUpper(usecaseName[:1]) + usecaseName[1:]

	data := UsecaseData{
		ModuleName: utils.GetModuleName(),
		Name:       usecaseName,
		Receiver:   strings.ToLower(usecaseName[:1]) + usecaseName[1:], // camelCase for receiver
	}

	// Check if usecase already exists
	usecasePath := filepath.Join("internal", "core", "usecase", strings.ToLower(usecaseName)+".go")
	if _, err := os.Stat(usecasePath); err == nil {
		return fmt.Errorf("usecase %s already exists at %s", usecaseName, usecasePath)
	}

	// Generate usecase file
	f := generateUsecaseFile(data)

	// Ensure directory exists
	dir := filepath.Dir(usecasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := f.Save(usecasePath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("Usecase %s created successfully at %s\n", usecaseName, usecasePath)
	return nil
}

func generateUsecaseFile(data UsecaseData) *jen.File {
	f := jen.NewFile("usecase")

	// Add imports
	f.ImportName("context", "context")
	f.ImportName("log", "log")
	f.ImportName(data.ModuleName+"/internal/core/domain", "domain_entries")
	f.ImportName(data.ModuleName+"/internal/core/model", "model")
	f.ImportName(data.ModuleName+"/internal/infrastructure/db", "db")
	f.ImportName(data.ModuleName+"/internal/pkg/validator", "validator")
	f.ImportName(data.ModuleName+"/internal/infrastructure/sse", "sse")

	// Generate Input struct
	f.Type().Id(data.Name + "Input").Struct(
		jen.Comment("TODO: Define fields"),
	)

	// Generate Output struct
	f.Type().Id(data.Name + "Output").Struct(
		jen.Comment("TODO: Define fields"),
	)

	// Generate usecase type alias
	f.Type().Id(data.Name).Qual("model", "Usecase").Types(
		jen.Id(data.Name+"Input"),
		jen.Id(data.Name+"Output"),
	)

	// Generate implementation struct
	f.Type().Id(data.Receiver).Struct(
		jen.Id("client").Op("*").Qual("db", "Client"),
		jen.Id("domain").Op("*").Qual("domain_entries", "Domain"),
		jen.Id("validator").Qual("validator", "Validator"),
		jen.Id("sse").Op("*").Qual("sse", "Manager"),
	)

	// Generate constructor
	f.Func().Id("New"+data.Name).Params(
		jen.Id("client").Op("*").Qual("db", "Client"),
		jen.Id("domain").Op("*").Qual("domain_entries", "Domain"),
		jen.Id("validator").Qual("validator", "Validator"),
		jen.Id("sse").Op("*").Qual("sse", "Manager"),
	).Id(data.Name).Block(
		jen.Return(jen.Op("&").Id(data.Receiver).Values(jen.Dict{
			jen.Id("client"):    jen.Id("client"),
			jen.Id("domain"):    jen.Id("domain"),
			jen.Id("validator"): jen.Id("validator"),
			jen.Id("sse"):       jen.Id("sse"),
		})),
	)

	// Generate Validate method
	f.Func().Params(jen.Id("uc").Op("*").Id(data.Receiver)).Id("Validate").Params(
		jen.Id("input").Id(data.Name + "Input"),
	).Error().Block(
		jen.Return(jen.Id("uc").Dot("validator").Dot("Validate").Call(jen.Id("input"))),
	)

	// Generate Execute method
	f.Func().Params(jen.Id("uc").Op("*").Id(data.Receiver)).Id("Execute").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("input").Id(data.Name+"Input"),
	).Params(jen.Op("*").Id(data.Name+"Output"), jen.Error()).Block(
		jen.Id("log").Op(":=").Qual("logger", "Get").Call(),
		jen.Line(),
		jen.If(jen.Id("err").Op(":=").Id("uc").Dot("Validate").Call(jen.Id("input")), jen.Id("err").Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Id("err")),
		),
		jen.Line(),
		jen.List(jen.Id("res"), jen.Id("err")).Op(":=").Id("uc").Dot("Process").Call(jen.Id("ctx"), jen.Id("input")),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Id("log").Dot("Error").Call().
				Dot("Str").Call(jen.Lit("usecase"), jen.Lit(data.Name)).
				Dot("Str").Call(jen.Lit("error"), jen.Id("err").Dot("Error").Call()).
				Dot("Msg").Call(jen.Lit("Failed to process "+data.Name)),
			jen.Return(jen.Nil(), jen.Id("err")),
		),
		jen.Line(),
		jen.If(jen.Id("err").Op(":=").Id("uc").Dot("Observer").Call(jen.Id("ctx"), jen.Id("input")), jen.Id("err").Op("!=").Nil()).Block(
			jen.Id("log").Dot("Printf").Call(jen.Lit("Observer usecase '"+data.Name+"' error: %s"), jen.Id("err").Dot("Error").Call()),
		),
		jen.Line(),
		jen.Return(jen.Id("res"), jen.Nil()),
	)

	// Generate Observer method
	f.Func().Params(jen.Id("uc").Op("*").Id(data.Receiver)).Id("Observer").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("input").Id(data.Name+"Input"),
	).Error().Block(
		jen.Comment("TODO: Implement observer logic"),
		jen.Comment("This is optional. You can leave this blank if not needed."),
		jen.Line(),
		jen.Return(jen.Nil()),
	)

	// Generate SendEvent method
	f.Func().Params(jen.Id("uc").Op("*").Id(data.Receiver)).Id("SendEvent").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("input").Id(data.Name+"Input"),
		jen.Id("output").Id(data.Name+"Output"),
	).Error().Block(
		jen.Comment("TODO: Implement event sending logic"),
		jen.Comment("This is optional. You can leave this blank if not needed."),
		jen.Line(),
		jen.Return(jen.Nil()),
	)

	// Generate Process method
	f.Func().Params(jen.Id("uc").Op("*").Id(data.Receiver)).Id("Process").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("input").Id(data.Name+"Input"),
	).Params(jen.Op("*").Id(data.Name+"Output"), jen.Error()).Block(
		jen.Comment("TODO: Implement logic"),
		jen.Line(),
		jen.Return(jen.Op("&").Id(data.Name+"Output").Values(), jen.Nil()),
	)

	return f
}
