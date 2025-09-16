package enum

import (
	"fmt"
	"os"
	"os/exec"
)

func RegenerateEnum() error {
	// For now, execute the original command
	// TODO: Implement native Go regeneration logic
	cmd := exec.Command("go", "run", "./cmd/enum/cmd/main.go", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to regenerate enum: %w", err)
	}

	fmt.Printf("✅ Successfully regenerated enum\n")
	return nil
}
