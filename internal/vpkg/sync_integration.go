package vpkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SyncCapability represents a vpkg package's sync capability
type SyncCapability struct {
	PackageName string
	Provider    string // Go file that provides sync functions
	Commands    []string
}

// VpkgSyncManager manages sync integration for installed vpkg packages
type VpkgSyncManager struct {
	projectRoot string
}

// NewVpkgSyncManager creates a new vpkg sync manager
func NewVpkgSyncManager() *VpkgSyncManager {
	return &VpkgSyncManager{
		projectRoot: ".",
	}
}

// DiscoverSyncCapabilities discovers all installed vpkg packages with sync capabilities
func (m *VpkgSyncManager) DiscoverSyncCapabilities() ([]SyncCapability, error) {
	var capabilities []SyncCapability

	// Check if vpkg directory exists
	vpkgDir := filepath.Join("internal", "vpkg")
	if _, err := os.Stat(vpkgDir); os.IsNotExist(err) {
		return capabilities, nil // No vpkg packages installed
	}

	// Walk through installed packages
	err := filepath.Walk(vpkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for meta.yaml files
		if info.Name() == "meta.yaml" {
			capability, err := m.extractSyncCapability(path)
			if err != nil {
				// Skip packages with invalid meta.yaml
				return nil
			}
			if capability != nil {
				capabilities = append(capabilities, *capability)
			}
		}

		return nil
	})

	return capabilities, err
}

// extractSyncCapability extracts sync capability from a package's meta.yaml
func (m *VpkgSyncManager) extractSyncCapability(metaPath string) (*SyncCapability, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta PackageMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	// Check if package has sync capabilities
	hasSyncCapability := false
	for _, capability := range meta.Capabilities {
		if capability == "sync-integration" {
			hasSyncCapability = true
			break
		}
	}

	if !hasSyncCapability || meta.Sync.Provider == "" {
		return nil, nil // No sync capability
	}

	return &SyncCapability{
		PackageName: meta.Name,
		Provider:    meta.Sync.Provider,
		Commands:    meta.Sync.Commands,
	}, nil
}

// ExecuteSyncCapabilities executes sync functions for all capable packages
func (m *VpkgSyncManager) ExecuteSyncCapabilities() error {
	capabilities, err := m.DiscoverSyncCapabilities()
	if err != nil {
		return fmt.Errorf("failed to discover sync capabilities: %w", err)
	}

	if len(capabilities) == 0 {
		return nil // No packages with sync capabilities
	}

	fmt.Printf("🔄 Syncing vpkg packages with sync capabilities...\n")

	for _, capability := range capabilities {
		if err := m.executeSyncCapability(capability); err != nil {
			return fmt.Errorf("failed to sync %s: %w", capability.PackageName, err)
		}
	}

	return nil
}

// executeSyncCapability executes sync for a specific package capability
func (m *VpkgSyncManager) executeSyncCapability(capability SyncCapability) error {
	// Find the package directory
	packageDir, err := m.findPackageDirectory(capability.PackageName)
	if err != nil {
		return err
	}

	// Find the sync provider file
	providerPath := filepath.Join(packageDir, capability.Provider)
	if _, err := os.Stat(providerPath); os.IsNotExist(err) {
		return fmt.Errorf("sync provider not found: %s", providerPath)
	}

	fmt.Printf("   Syncing %s...\n", capability.PackageName)

	// Execute the sync function
	// For now, we'll use go run to execute the sync function
	// In the future, we could compile and load as a plugin
	if err := m.executeGoSyncFunction(providerPath); err != nil {
		return err
	}

	return nil
}

// findPackageDirectory finds the directory of an installed vpkg package
func (m *VpkgSyncManager) findPackageDirectory(packageName string) (string, error) {
	// Parse package name (e.g., "vandor/http-huma" -> "vandor", "http-huma")
	parts := strings.Split(packageName, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid package name format: %s", packageName)
	}

	namespace, pkg := parts[0], parts[1]
	packageDir := filepath.Join("internal", "vpkg", namespace, pkg)

	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		return "", fmt.Errorf("package directory not found: %s", packageDir)
	}

	return packageDir, nil
}

// executeGoSyncFunction executes a Go sync function
func (m *VpkgSyncManager) executeGoSyncFunction(providerPath string) error {
	// Create a temporary main.go that calls the sync function
	tempDir, err := os.MkdirTemp("", "vpkg-sync")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Create main.go that imports and calls the sync function
	mainGoContent := fmt.Sprintf(`package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Change to project root directory
	if err := os.Chdir("%s"); err != nil {
		fmt.Printf("Error changing directory: %%v\n", err)
		os.Exit(1)
	}

	// Import and call the sync function dynamically
	// For now, we'll shell out to go run the sync provider directly
	fmt.Println("Executing vpkg sync function...")
}
`, m.projectRoot)

	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		return err
	}

	// For now, we'll use a simple approach - check if the provider has a main function
	// and execute it directly. In the future, we could use Go plugins or other approaches.

	// Try to execute the sync provider as a standalone program
	cmd := exec.Command("go", "run", providerPath)
	cmd.Dir = m.projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Extended PackageMeta to include sync capabilities
type PackageMeta struct {
	Name         string   `yaml:"name"`
	Capabilities []string `yaml:"capabilities,omitempty"`
	Sync         struct {
		Provider string   `yaml:"provider,omitempty"`
		Commands []string `yaml:"commands,omitempty"`
	} `yaml:"sync,omitempty"`
}
