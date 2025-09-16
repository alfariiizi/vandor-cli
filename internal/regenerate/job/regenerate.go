package job

import (
	"fmt"
	"os"
	"os/exec"
)

func RegenerateJob() error {
	// For now, execute the original command
	// TODO: Implement native Go regeneration logic
	cmd := exec.Command("go", "run", "./cmd/job/cmd-regenerate-job/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to regenerate jobs: %w", err)
	}

	fmt.Printf("✅ Successfully regenerated jobs\n")
	return nil
}
