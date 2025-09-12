package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/vandordev/vandor/internal/utils"
)

type JobData struct {
	ModuleName string // e.g., "github.com/vandordev/vandor"
	StructName string // e.g., "SendEmail"
	VarName    string // e.g., "sendEmail"
	JobKey     string // e.g., "job:send_email"
}

func GenerateJob(jobName string) error {
	if jobName == "" {
		return fmt.Errorf("job name cannot be empty")
	}

	// Capitalize first letter for struct name
	jobName = strings.ToUpper(jobName[:1]) + jobName[1:]

	data := JobData{
		ModuleName: utils.GetModuleName(),
		StructName: jobName,
		VarName:    strings.ToLower(jobName[:1]) + jobName[1:], // camelCase
		JobKey:     "job:" + strings.ToLower(jobName),          // snake_case with prefix
	}

	// Check if job already exists
	jobPath := filepath.Join("internal", "delivery", "worker", "job", strings.ToLower(jobName)+".go")
	if _, err := os.Stat(jobPath); err == nil {
		return fmt.Errorf("job %s already exists at %s", jobName, jobPath)
	}

	// Generate job file
	f := generateJobFile(data)

	// Ensure directory exists
	dir := filepath.Dir(jobPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := f.Save(jobPath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("Job %s created successfully at %s\n", jobName, jobPath)
	return nil
}

func generateJobFile(data JobData) *jen.File {
	f := jen.NewFile("job")

	// Add imports
	f.ImportName("context", "context")
	f.ImportName("encoding/json", "json")
	f.ImportName("log", "log")
	f.ImportName("strings", "strings")
	f.ImportName(data.ModuleName+"/internal/core/domain", "domain_entries")
	f.ImportName(data.ModuleName+"/internal/core/model", "model")
	f.ImportName(data.ModuleName+"/internal/delivery/http/api", "api")
	f.ImportName(data.ModuleName+"/internal/delivery/http/method", "method")
	f.ImportName(data.ModuleName+"/internal/delivery/worker", "worker")
	f.ImportName(data.ModuleName+"/internal/infrastructure/db", "db")
	f.ImportName(data.ModuleName+"/internal/types", "types")
	f.ImportName(data.ModuleName+"/pkg/validator", "validator")
	f.ImportName("github.com/hibiken/asynq", "asynq")
	f.ImportName("github.com/danielgtaylor/huma/v2", "huma")

	// Generate payload struct
	f.Comment("Payload definition")
	f.Type().Id(data.StructName + "Payload").Struct(
		jen.Comment("TODO: add fields for payload"),
	)

	// Generate job type alias
	f.Comment("Job type alias")
	f.Type().Id(data.StructName).Qual("model", "Job").Types(jen.Id(data.StructName + "Payload"))

	// Generate HTTP input struct
	f.Comment("HTTP input for job enqueue endpoint")
	f.Type().Id(data.StructName + "HTTPInput").Struct(
		jen.Id("JobSecret").String().Tag(map[string]string{
			"header":   "X-Job-Secret",
			"required": "true",
		}),
		jen.Id("Body").Id(data.StructName + "Payload").Tag(map[string]string{
			"json":        "body",
			"contentType": "application/json",
		}),
	)

	// Generate implementation struct
	f.Comment("Job implementation")
	f.Type().Id(data.VarName).Struct(
		jen.Id("api").Qual("huma", "API"),
		jen.Id("client").Op("*").Qual("db", "Client"),
		jen.Id("domain").Op("*").Qual("domain_entries", "Domain"),
		jen.Id("validator").Qual("validator", "Validator"),
		jen.Id("worker").Op("*").Qual("worker", "Client"),
	)

	// Generate constructor
	f.Func().Id("New" + data.StructName).Params(
		jen.Id("client").Op("*").Qual("db", "Client"),
		jen.Id("domain").Op("*").Qual("domain_entries", "Domain"),
		jen.Id("validator").Qual("validator", "Validator"),
		jen.Id("worker").Op("*").Qual("worker", "Client"),
		jen.Id("api").Op("*").Qual("api", "HttpApi"),
	).Id(data.StructName).Block(
		jen.Return(jen.Op("&").Id(data.VarName).Values(jen.Dict{
			jen.Id("client"):    jen.Id("client"),
			jen.Id("domain"):    jen.Id("domain"),
			jen.Id("validator"): jen.Id("validator"),
			jen.Id("worker"):    jen.Id("worker"),
			jen.Id("api"):       jen.Id("api").Dot("JobAPI"),
		})),
	)

	// Generate Key method
	f.Func().Params(jen.Id("j").Op("*").Id(data.VarName)).Id("Key").Params().String().Block(
		jen.Return(jen.Lit(data.JobKey)),
	)

	// Generate Enqueue method
	f.Comment("Enqueue method will be called by clients to enqueue a new job")
	f.Func().Params(jen.Id("j").Op("*").Id(data.VarName)).Id("Enqueue").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("payload").Id(data.StructName + "Payload"),
	).Params(jen.Op("*").Qual("asynq", "TaskInfo"), jen.Error()).Block(
		jen.If(jen.Id("err").Op(":=").Id("j").Dot("validator").Dot("Validate").Call(jen.Id("payload")), jen.Id("err").Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Id("err")),
		),
		jen.Line(),
		jen.List(jen.Id("data"), jen.Id("_")).Op(":=").Qual("json", "Marshal").Call(jen.Id("payload")),
		jen.Id("task").Op(":=").Qual("asynq", "NewTask").Call(jen.Id("j").Dot("Key").Call(), jen.Id("data")),
		jen.Return(jen.Id("j").Dot("worker").Dot("EnqueueContext").Call(jen.Id("ctx"), jen.Id("task"), jen.Qual("asynq", "Queue").Call(jen.Lit("default")))),
	)

	// Generate HTTPRegisterRoute method
	f.Comment("HTTPRegisterRoute registers the HTTP route for job enqueueing")
	f.Func().Params(jen.Id("j").Op("*").Id(data.VarName)).Id("HTTPRegisterRoute").Params().Block(
		jen.Id("path").Op(":=").Lit("/").Op("+").Qual("strings", "Split").Call(jen.Id("j").Dot("Key").Call(), jen.Lit(":")).Index(jen.Lit(1)),
		jen.Line(),
		jen.Qual("method", "POST").Call(
			jen.Id("j").Dot("api"),
			jen.Id("path"),
			jen.Qual("method", "Operation").Values(jen.Dict{
				jen.Id("Summary"):     jen.Lit(data.StructName + " Job"),
				jen.Id("Description"): jen.Lit("Enqueue a new " + data.StructName + " job"),
				jen.Id("Tags"):        jen.Index().String().Values(jen.Lit("Job")),
				jen.Id("Job"):         jen.Lit(true),
			}),
			jen.Func().Params(
				jen.Id("ctx").Qual("context", "Context"),
				jen.Id("input").Op("*").Id(data.StructName + "HTTPInput"),
			).Params(jen.Op("*").Qual("model", "JobHTTPHandlerResponse"), jen.Error()).Block(
				jen.List(jen.Id("taskInfo"), jen.Id("err")).Op(":=").Id("j").Dot("Enqueue").Call(jen.Id("ctx"), jen.Id("input").Dot("Body")),
				jen.If(jen.Id("err").Op("!=").Nil()).Block(
					jen.Return(jen.Nil(), jen.Id("err")),
				),
				jen.Return(
					jen.Params(jen.Op("*").Qual("model", "JobHTTPHandlerResponse")).Call(
						jen.Qual("types", "GenerateOutputResponseData").Call(
							jen.Qual("model", "JobHTTPHandlerData").Values(jen.Dict{
								jen.Id("TaskID"): jen.Id("taskInfo").Dot("ID"),
							}),
						),
					),
					jen.Nil(),
				),
			),
		),
	)

	// Generate Handle method
	f.Comment("Handle method processes the job when it is executed by the worker (server worker)")
	f.Func().Params(jen.Id("j").Op("*").Id(data.VarName)).Id("Handle").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("payload").Id(data.StructName + "Payload"),
	).Error().Block(
		jen.Comment("TODO: implement job logic"),
		jen.Qual("log", "Println").Call(jen.Lit("Handling "+data.StructName+" job"), jen.Lit("payload"), jen.Id("payload")),
		jen.Return(jen.Nil()),
	)

	return f
}