package entgo

import (
	"fmt"
	"os"
	"os/exec"
)

func RegenerateEntgo() error {
	// Generate ent code
	cmd := exec.Command("go", "run", "./cmd/entgo/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate DB model: %w", err)
	}

	// Run goimports on the generated code
	cmd = exec.Command("goimports", "-w", "./internal/infrastructure/db/rest/.")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run goimports: %w", err)
	}

	fmt.Printf("✅ Successfully regenerated DB Model\n")
	return nil
}
