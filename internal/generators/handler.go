package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/alfariiizi/vandor-cli/internal/utils"
)

type HandlerData struct {
	ModuleName string // e.g., "github.com/alfariiizi/vandor-cli"
	Name       string // e.g., "GetUser"
	Group      string // e.g., "user"
	GroupTitle string // e.g., "User"
	Path       string // e.g., "users/{id}"
	Method     string // e.g., "GET"
	Receiver   string // e.g., "getUser"
}

func GenerateHandler(name, group, method string) error {
	if name == "" || group == "" || method == "" {
		return fmt.Errorf("all parameters (name, group, method) are required")
	}

	// Generate path from name (convert to lowercase)
	path := strings.ToLower(name)

	// Normalize inputs
	name = strings.ToUpper(name[:1]) + name[1:]
	method = strings.ToUpper(method)
	group = strings.ToLower(group)

	data := HandlerData{
		ModuleName: utils.GetModuleName(),
		Name:       name,
		Group:      group,
		GroupTitle: strings.ToUpper(group[:1]) + group[1:],
		Path:       path,
		Method:     method,
		Receiver:   strings.ToLower(name[:1]) + name[1:],
	}

	// Check if handler already exists
	handlerPath := filepath.Join("internal", "delivery", "http", "handler", group, strings.ToLower(name)+".go")
	if _, err := os.Stat(handlerPath); err == nil {
		return fmt.Errorf("handler %s already exists at %s", name, handlerPath)
	}

	// Generate handler file
	f := generateHandlerFile(data)

	// Ensure directory exists
	dir := filepath.Dir(handlerPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := f.Save(handlerPath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("Handler %s created successfully at %s\n", name, handlerPath)
	return nil
}

func generateHandlerFile(data HandlerData) *jen.File {
	f := jen.NewFile(data.Group + "_handler")

	// Add imports
	f.ImportName("context", "context")
	f.ImportName(data.ModuleName+"/internal/core/model", "model")
	f.ImportName(data.ModuleName+"/internal/core/service", "service")
	f.ImportName(data.ModuleName+"/internal/infrastructure/db", "db")
	f.ImportName(data.ModuleName+"/internal/delivery/http/api", "api")
	f.ImportName(data.ModuleName+"/internal/delivery/http/method", "method")
	f.ImportName(data.ModuleName+"/internal/types", "types")
	f.ImportName("github.com/danielgtaylor/huma/v2", "huma")

	// Generate payload struct based on method
	generatePayloadStruct(f, data)

	// Generate input struct
	generateInputStruct(f, data)

	// Generate output types
	f.Type().Id(data.Name+"Output").Qual("types", "OutputResponseData").Types(jen.Id(data.Name + "Data"))

	f.Type().Id(data.Name+"Data").Struct(
		jen.Comment("Example response data"),
		jen.Id("ID").String().Tag(map[string]string{
			"json":    "id",
			"example": "123",
		}),
		jen.Id("Name").String().Tag(map[string]string{
			"json":    "name",
			"example": "Book",
		}),
		jen.Id("Description").String().Tag(map[string]string{
			"json":    "description",
			"example": "A great book",
		}),
	)

	// Generate handler type
	f.Type().Id(data.Name+"Handler").Qual("model", "HTTPHandler").Types(
		jen.Id(data.Name+"Input"),
		jen.Id(data.Name+"Output"),
	)

	// Generate implementation struct
	f.Type().Id(data.Receiver).Struct(
		jen.Id("api").Qual("huma", "API"),
		jen.Id("service").Op("*").Qual("service", "Services"),
		jen.Id("client").Op("*").Qual("db", "Client"),
	)

	// Generate constructor
	f.Func().Id("New"+data.Name).Params(
		jen.Id("api").Op("*").Qual("api", "HttpApi"),
		jen.Id("service").Op("*").Qual("service", "Services"),
		jen.Id("client").Op("*").Qual("db", "Client"),
	).Id(data.Name+"Handler").Block(
		jen.Id("h").Op(":=").Op("&").Id(data.Receiver).Values(jen.Dict{
			jen.Id("api"):     jen.Id("api").Dot("BaseAPI"),
			jen.Id("service"): jen.Id("service"),
			jen.Id("client"):  jen.Id("client"),
		}),
		jen.Id("h").Dot("RegisterRoutes").Call(),
		jen.Return(jen.Id("h")),
	)

	// Generate RegisterRoutes method
	generateRegisterRoutesMethod(f, data)

	// Generate GenerateResponse method
	f.Func().Params(jen.Id("h").Op("*").Id(data.Receiver)).Id("GenerateResponse").Params(
		jen.Id("data").Id(data.Name + "Data"),
	).Op("*").Id(data.Name + "Output").Block(
		jen.Return(jen.Params(jen.Op("*").Id(data.Name + "Output")).Call(
			jen.Qual("types", "GenerateOutputResponseData").Call(jen.Id("data")),
		)),
	)

	// Generate Handler method
	f.Func().Params(jen.Id("h").Op("*").Id(data.Receiver)).Id("Handler").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("input").Op("*").Id(data.Name+"Input"),
	).Params(jen.Op("*").Id(data.Name+"Output"), jen.Error()).Block(
		jen.Comment("TODO: Implement handler logic here"),
		jen.Line(),
		jen.Return(jen.Id("h").Dot("GenerateResponse").Call(jen.Id(data.Name+"Data").Values()), jen.Nil()),
	)

	return f
}

func generatePayloadStruct(f *jen.File, data HandlerData) {
	switch data.Method {
	case "POST":
		f.Type().Id(data.Name+"Payload").Struct(
			jen.Comment("Example POST input"),
			jen.Id("Name").String().Tag(map[string]string{
				"json":     "name",
				"doc":      "Name of the item",
				"example":  "Book",
				"required": "true",
			}),
			jen.Id("Description").String().Tag(map[string]string{
				"json": "description",
			}),
		)
	case "PUT", "PATCH":
		f.Type().Id(data.Name+"Payload").Struct(
			jen.Comment("Example "+data.Method+" input"),
			jen.Id("Name").Op("*").String().Tag(map[string]string{
				"json":    "name",
				"doc":     "Name of the item",
				"example": "Book",
			}),
			jen.Id("Description").Op("*").String().Tag(map[string]string{
				"json": "description",
			}),
		)
	}
}

func generateInputStruct(f *jen.File, data HandlerData) {
	structFields := []jen.Code{}

	switch data.Method {
	case "GET":
		structFields = append(structFields,
			jen.Comment("Example GET input"),
			jen.Id("ID").String().Tag(map[string]string{
				"path":    "id",
				"doc":     "ID of the item",
				"example": "123",
			}),
			jen.Id("Query").String().Tag(map[string]string{
				"query":   "q",
				"doc":     "Query parameter for filtering",
				"example": "search term",
			}),
		)
	case "DELETE":
		structFields = append(structFields,
			jen.Id("ID").String().Tag(map[string]string{
				"path":    "id",
				"doc":     "ID of the item to delete",
				"example": "123",
			}),
		)
	case "PUT", "PATCH":
		structFields = append(structFields,
			jen.Id("ID").String().Tag(map[string]string{
				"path":    "id",
				"doc":     "ID of the item to update",
				"example": "123",
			}),
			jen.Id("Body").Id(data.Name+"Payload").Tag(map[string]string{
				"json":        "body",
				"contentType": "application/json",
			}),
		)
	default: // POST and others
		structFields = append(structFields,
			jen.Comment("JSON body for "+data.Method),
			jen.Id("Body").Id(data.Name+"Payload").Tag(map[string]string{
				"json":        "body",
				"contentType": "application/json",
			}),
		)
	}

	f.Comment("NOTE:")
	f.Comment("Hint Tags for input parameters")
	f.Comment("@ref: https://huma.rocks/features/request-inputs")
	f.Comment("")
	f.Comment("Tag       | Description                           | Example")
	f.Comment("-------------------------------------------------------------------")
	f.Comment("path      | Name of the path parameter            | path:\"thing-id\"")
	f.Comment("query     | Name of the query string parameter    | query:\"q\"")
	f.Comment("header    | Name of the header parameter          | header:\"Authorization\"")
	f.Comment("cookie    | Name of the cookie parameter          | cookie:\"session\"")
	f.Comment("required  | Mark a query/header param as required | required:\"true\"")
	f.Line()

	f.Type().Id(data.Name + "Input").Struct(structFields...)
}

func generateRegisterRoutesMethod(f *jen.File, data HandlerData) {
	methodCall := jen.Id("method").Dot(data.Method).Call(
		jen.Id("api"),
		jen.Lit("/"+data.Path),
		jen.Qual("method", "Operation").Values(jen.Dict{
			jen.Id("Summary"):     jen.Lit(data.Name),
			jen.Id("Description"): jen.Lit(data.Name + " handler"),
			jen.Id("Tags"):        jen.Index().String().Values(jen.Lit(data.GroupTitle)),
			jen.Id("BearerAuth"):  jen.Lit(false),
		}),
		jen.Id("h").Dot("Handler"),
	)

	f.Func().Params(jen.Id("h").Op("*").Id(data.Receiver)).Id("RegisterRoutes").Params().Block(
		jen.Id("api").Op(":=").Id("h").Dot("api"),
		methodCall,
	)
}
