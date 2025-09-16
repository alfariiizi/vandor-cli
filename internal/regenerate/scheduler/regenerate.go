package scheduler

import (
	"fmt"
	"os"
	"os/exec"
)

func RegenerateScheduler() error {
	// For now, execute the original command
	// TODO: Implement native Go regeneration logic
	cmd := exec.Command("go", "run", "./cmd/scheduler/cmd-regenerate-scheduler/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to regenerate scheduler: %w", err)
	}

	fmt.Printf("✅ Successfully regenerated scheduler\n")
	return nil
}
