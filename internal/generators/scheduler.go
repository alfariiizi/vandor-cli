package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/vandordev/vandor/internal/utils"
)

type SchedulerData struct {
	ModuleName  string // e.g., "github.com/vandordev/vandor"
	PascalName  string // e.g., "BackupData"
	FunctionName string // e.g., "RegisterBackupDataJob"
}

func GenerateScheduler(schedulerName string) error {
	if schedulerName == "" {
		return fmt.Errorf("scheduler name cannot be empty")
	}

	// Capitalize first letter for PascalCase
	schedulerName = strings.ToUpper(schedulerName[:1]) + schedulerName[1:]

	data := SchedulerData{
		ModuleName:   utils.GetModuleName(),
		PascalName:   schedulerName,
		FunctionName: "Register" + schedulerName + "Job",
	}

	// Check if scheduler already exists
	schedulerPath := filepath.Join("internal", "cron", "scheduler", strings.ToLower(schedulerName)+".go")
	if _, err := os.Stat(schedulerPath); err == nil {
		return fmt.Errorf("scheduler %s already exists at %s", schedulerName, schedulerPath)
	}

	// Generate scheduler file
	f := generateSchedulerFile(data)

	// Ensure directory exists
	dir := filepath.Dir(schedulerPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := f.Save(schedulerPath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("Scheduler %s created successfully at %s\n", schedulerName, schedulerPath)
	return nil
}

func generateSchedulerFile(data SchedulerData) *jen.File {
	f := jen.NewFile("scheduler")

	// Add imports
	f.ImportName("context", "context")
	f.ImportName("log", "log")
	f.ImportName(data.ModuleName+"/internal/core/job", "job")
	f.ImportName(data.ModuleName+"/internal/cron/init", "cron")

	// Generate scheduler registration function
	f.Func().Id(data.FunctionName).Params(
		jen.Id("s").Op("*").Qual("cron", "Scheduler"),
	).Block(
		jen.Id("s").Dot("Scheduler").Dot("Every").Call(jen.Lit(1)).Dot("Minute").Call().Dot("Do").Call(
			jen.Func().Params().Block(
				jen.Qual("log", "Println").Call(jen.Lit("[cron] Running " + data.PascalName + " job...")),
				jen.Line(),
				jen.Id("payload").Op(":=").Qual("job", "LogSystemPayload").Values(jen.Dict{
					jen.Id("Message"): jen.Lit(data.PascalName + " job executed"),
				}),
				jen.Line(),
				jen.If(
					jen.List(jen.Id("_"), jen.Id("err")).Op(":=").Id("s").Dot("Jobs").Dot("LogSystem").Dot("Enqueue").Call(
						jen.Qual("context", "Background").Call(),
						jen.Id("payload"),
					),
					jen.Id("err").Op("!=").Nil(),
				).Block(
					jen.Qual("log", "Printf").Call(
						jen.Lit("[cron] Error enqueueing " + data.PascalName + " job: %v"),
						jen.Id("err"),
					),
				),
			),
		),
	)

	return f
}