package vpkg

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	configPath   string
	vpkgDir      string
	registryURL  string
	communityURL string
}

type Package struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Tags        []string `yaml:"tags" json:"tags"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Author      string   `yaml:"author,omitempty" json:"author,omitempty"`
	Repository  string   `yaml:"repository,omitempty" json:"repository,omitempty"`
}

type VandorConfig struct {
	Project struct {
		Name    string `yaml:"name"`
		Module  string `yaml:"module"`
		Version string `yaml:"version"`
	} `yaml:"project"`
	Vandor struct {
		CLI          string `yaml:"cli"`
		Architecture string `yaml:"architecture"`
		Language     string `yaml:"language"`
	} `yaml:"vandor"`
	Vpkg []Package `yaml:"vpkg,omitempty"`
}

type PackageRegistry struct {
	Packages []Package `yaml:"packages" json:"packages"`
}

func NewManager() (*Manager, error) {
	return &Manager{
		configPath:   "vandor-config.yaml",
		vpkgDir:      "internal/vpkg",
		registryURL:  "https://raw.githubusercontent.com/alfarizi/vandor-packages/main/registry.yaml",
		communityURL: "https://raw.githubusercontent.com/vandor-community/packages/main/registry.yaml",
	}, nil
}

func (m *Manager) loadConfig() (*VandorConfig, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config VandorConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

func (m *Manager) saveConfig(config *VandorConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

func (m *Manager) AddPackage(packageName string) error {
	// Load current config
	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	// Check if package is already installed
	for _, pkg := range config.Vpkg {
		if pkg.Name == packageName {
			return fmt.Errorf("package %s is already installed", packageName)
		}
	}

	// Find package in registry
	pkg, err := m.findPackageInRegistry(packageName)
	if err != nil {
		return err
	}

	// Check if package is compatible with current architecture
	if !m.isPackageCompatible(pkg, config.Vandor.Architecture) {
		return fmt.Errorf("package %s is not compatible with architecture %s", packageName, config.Vandor.Architecture)
	}

	// Download and install package
	if err := m.installPackage(pkg); err != nil {
		return fmt.Errorf("failed to install package: %v", err)
	}

	// Add to config
	config.Vpkg = append(config.Vpkg, *pkg)

	// Save config
	return m.saveConfig(config)
}

func (m *Manager) RemovePackage(packageName string) error {
	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	// Find and remove package from config
	found := false
	newVpkg := []Package{}
	for _, pkg := range config.Vpkg {
		if pkg.Name != packageName {
			newVpkg = append(newVpkg, pkg)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("package %s is not installed", packageName)
	}

	config.Vpkg = newVpkg

	// Remove package files
	packageDir := filepath.Join(m.vpkgDir, packageName)
	if err := os.RemoveAll(packageDir); err != nil {
		return fmt.Errorf("failed to remove package files: %v", err)
	}

	// Save config
	return m.saveConfig(config)
}

func (m *Manager) ListPackages() ([]Package, error) {
	config, err := m.loadConfig()
	if err != nil {
		return nil, err
	}

	return config.Vpkg, nil
}

func (m *Manager) UpdatePackage(packageName string) error {
	// Find latest version in registry
	pkg, err := m.findPackageInRegistry(packageName)
	if err != nil {
		return err
	}

	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	// Check if package is installed
	found := false
	for i, installedPkg := range config.Vpkg {
		if installedPkg.Name == packageName {
			if installedPkg.Version == pkg.Version {
				return fmt.Errorf("package %s is already up to date (v%s)", packageName, pkg.Version)
			}

			// Update package
			if err := m.installPackage(pkg); err != nil {
				return fmt.Errorf("failed to update package: %v", err)
			}

			config.Vpkg[i] = *pkg
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("package %s is not installed", packageName)
	}

	return m.saveConfig(config)
}

func (m *Manager) UpdateAllPackages() error {
	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	for i, installedPkg := range config.Vpkg {
		// Find latest version
		pkg, err := m.findPackageInRegistry(installedPkg.Name)
		if err != nil {
			fmt.Printf("Warning: Could not find package %s in registry, skipping\n", installedPkg.Name)
			continue
		}

		if installedPkg.Version != pkg.Version {
			fmt.Printf("Updating %s from v%s to v%s\n", pkg.Name, installedPkg.Version, pkg.Version)

			if err := m.installPackage(pkg); err != nil {
				return fmt.Errorf("failed to update package %s: %v", pkg.Name, err)
			}

			config.Vpkg[i] = *pkg
		}
	}

	return m.saveConfig(config)
}

func (m *Manager) SearchPackages(query string) ([]Package, error) {
	// Search in official registry
	officialPackages, _ := m.fetchPackagesFromRegistry(m.registryURL)

	// Search in community registry
	communityPackages, _ := m.fetchPackagesFromRegistry(m.communityURL)

	// Combine results
	allPackages := append(officialPackages, communityPackages...)

	if query == "" {
		return allPackages, nil
	}

	// Filter by query
	var results []Package
	queryLower := strings.ToLower(query)
	for _, pkg := range allPackages {
		if strings.Contains(strings.ToLower(pkg.Name), queryLower) ||
			strings.Contains(strings.ToLower(pkg.Description), queryLower) {
			results = append(results, pkg)
		}
	}

	return results, nil
}

func (m *Manager) findPackageInRegistry(packageName string) (*Package, error) {
	// Try official registry first
	packages, err := m.fetchPackagesFromRegistry(m.registryURL)
	if err == nil {
		for _, pkg := range packages {
			if pkg.Name == packageName {
				return &pkg, nil
			}
		}
	}

	// Try community registry
	packages, err = m.fetchPackagesFromRegistry(m.communityURL)
	if err == nil {
		for _, pkg := range packages {
			if pkg.Name == packageName {
				return &pkg, nil
			}
		}
	}

	return nil, fmt.Errorf("package %s not found in any registry", packageName)
}

func (m *Manager) fetchPackagesFromRegistry(url string) ([]Package, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var registry PackageRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, err
	}

	return registry.Packages, nil
}

func (m *Manager) isPackageCompatible(pkg *Package, architecture string) bool {
	for _, tag := range pkg.Tags {
		if tag == architecture {
			return true
		}
	}
	return false
}

func (m *Manager) installPackage(pkg *Package) error {
	// Create package directory
	packageDir := filepath.Join(m.vpkgDir, pkg.Name)
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return err
	}

	// Create package info file
	infoFile := filepath.Join(packageDir, "package.yaml")
	data, err := yaml.Marshal(pkg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(infoFile, data, 0644); err != nil {
		return err
	}

	// Download package files (this is a simplified implementation)
	// In a real implementation, you would download the actual package files
	// from the repository and extract them to the package directory

	// Create a basic setup file as example
	setupContent := fmt.Sprintf(`# %s Package Setup

This package provides %s functionality.

## Installation

This package has been automatically installed in your project.

## Usage

Please refer to the documentation for usage instructions.

## Version

Current version: %s
`, pkg.Name, pkg.Description, pkg.Version)

	setupFile := filepath.Join(packageDir, "README.md")
	return os.WriteFile(setupFile, []byte(setupContent), 0644)
}
