// Package utils provides utility functions for the command application.
package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

func GetModuleName() string {
	cmd := exec.Command("go", "list", "-m")
	modNameBytes, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	moduleName := string(bytes.TrimSpace(modNameBytes))
	return moduleName
}

// DetectGoModule detects the Go module name from go.mod
func DetectGoModule() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to detect Go module: %w", err)
	}
	return string(bytes.TrimSpace(output)), nil
}
