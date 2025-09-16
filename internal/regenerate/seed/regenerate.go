package seed

import (
	"fmt"
	"os"
	"os/exec"
)

func RegenerateSeed() error {
	// For now, execute the original command
	// TODO: Implement native Go regeneration logic
	cmd := exec.Command("go", "run", "./cmd/seed/cmd-generate/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to regenerate seed: %w", err)
	}

	fmt.Printf("✅ Successfully regenerated seed\n")
	return nil
}
