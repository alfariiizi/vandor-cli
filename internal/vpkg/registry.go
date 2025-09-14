package vpkg

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultRegistryURL = "https://raw.githubusercontent.com/alfariiizi/vpkg-registry/main/registry.yaml"
)

// RegistryClient handles fetching data from the package registry
type RegistryClient struct {
	registryURL string
	httpClient  *http.Client
}

// NewRegistryClient creates a new registry client
func NewRegistryClient(registryURL string) *RegistryClient {
	if registryURL == "" {
		registryURL = DefaultRegistryURL
	}

	return &RegistryClient{
		registryURL: registryURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchRegistry fetches and parses the registry index (new repository-based format)
func (r *RegistryClient) FetchRegistry() (*Registry, error) {
	resp, err := r.httpClient.Get(r.registryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry request failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry response: %w", err)
	}

	var registry Registry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry YAML: %w", err)
	}

	return &registry, nil
}

// FindPackage finds a package by name across all repositories
func (r *RegistryClient) FindPackage(name string) (*PackageWithRepo, error) {
	registry, err := r.FetchRegistry()
	if err != nil {
		return nil, err
	}

	// Search through all repositories
	for _, repoInfo := range registry.Repositories {
		repoMeta, err := r.FetchRepositoryMeta(repoInfo.MetaURL)
		if err != nil {
			// Continue searching other repositories if one fails
			continue
		}

		// Look for package in this repository
		for _, pkg := range repoMeta.Packages {
			if pkg.Name == name {
				return &PackageWithRepo{
					Package:        pkg,
					RepositoryInfo: repoInfo,
					RepositoryMeta: *repoMeta,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("package %s not found in any repository", name)
}

// FetchRepositoryMeta fetches the meta.yaml file for a repository
func (r *RegistryClient) FetchRepositoryMeta(metaURL string) (*RepositoryMeta, error) {
	resp, err := r.httpClient.Get(metaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository meta: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("repository meta request failed with status %d: %s", resp.StatusCode, metaURL)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read repository meta: %w", err)
	}

	var meta RepositoryMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse repository meta YAML: %w", err)
	}

	return &meta, nil
}

// FetchPackageFile fetches a specific file from a package's repository
func (r *RegistryClient) FetchPackageFile(packageWithRepo *PackageWithRepo, filePath string) ([]byte, error) {
	// Build URL from repository base URL + file path
	baseURL := strings.Replace(packageWithRepo.RepositoryInfo.MetaURL, "/meta.yaml", "", 1)
	fileURL := fmt.Sprintf("%s/%s", baseURL, filePath)

	resp, err := r.httpClient.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file %s: %w", filePath, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file request failed with status %d: %s", resp.StatusCode, fileURL)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return data, nil
}

// ListPackages returns all packages from all repositories, optionally filtered
func (r *RegistryClient) ListPackages(opts ListOptions) ([]Package, error) {
	registry, err := r.FetchRegistry()
	if err != nil {
		return nil, err
	}

	var allPackages []Package

	// Collect packages from all repositories
	for _, repoInfo := range registry.Repositories {
		repoMeta, err := r.FetchRepositoryMeta(repoInfo.MetaURL)
		if err != nil {
			// Log warning but continue with other repositories
			fmt.Printf("Warning: Failed to fetch repository %s: %v\n", repoInfo.Name, err)
			continue
		}

		// Add all packages from this repository
		allPackages = append(allPackages, repoMeta.Packages...)
	}

	// Apply filters
	var filtered []Package
	for _, pkg := range allPackages {
		// Filter by type if specified
		if opts.Type != "" && pkg.Type != opts.Type {
			continue
		}

		// Filter by tags if specified
		if len(opts.Tags) > 0 {
			hasTag := false
			for _, tag := range opts.Tags {
				for _, pkgTag := range pkg.Tags {
					if pkgTag == tag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		filtered = append(filtered, pkg)
	}

	return filtered, nil
}

// DiscoverTemplateFiles discovers all .tmpl files in a package's templates directory
func (r *RegistryClient) DiscoverTemplateFiles(packageWithRepo *PackageWithRepo, templatesDir string) ([]string, error) {
	// This function will recursively discover all template files from a remote repository
	// For now, we'll simulate this by fetching a directory listing or using known patterns
	// In a real implementation, you'd need to either:
	// 1. Use GitHub API to list directory contents
	// 2. Have a manifest file listing all templates
	// 3. Try common template patterns

	var templateFiles []string

	// Common template patterns to try
	commonPatterns := []string{
		"README.md.tmpl",
		"main.go.tmpl",
		"config.go.tmpl",
		"service.go.tmpl",
		"cmd/main.go.tmpl",
		"templates/main.go.tmpl",
		"templates/README.md.tmpl",
	}

	baseURL := strings.Replace(packageWithRepo.RepositoryInfo.MetaURL, "/meta.yaml", "", 1)

	// Try to fetch each common pattern
	for _, pattern := range commonPatterns {
		testURL := fmt.Sprintf("%s/%s/%s", baseURL, templatesDir, pattern)
		resp, err := r.httpClient.Get(testURL)
		if err != nil {
			continue // Skip files that don't exist
		}
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			templateFiles = append(templateFiles, filepath.Join(templatesDir, pattern))
		}
	}

	// If no templates found with common patterns, try the entry file
	if len(templateFiles) == 0 && packageWithRepo.Package.Entry != "" {
		entryTemplate := strings.Replace(packageWithRepo.Package.Entry, filepath.Ext(packageWithRepo.Package.Entry), ".tmpl", 1)
		testURL := fmt.Sprintf("%s/%s/%s", baseURL, templatesDir, entryTemplate)
		resp, err := r.httpClient.Get(testURL)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				templateFiles = append(templateFiles, filepath.Join(templatesDir, entryTemplate))
			}
		}
	}

	return templateFiles, nil
}
