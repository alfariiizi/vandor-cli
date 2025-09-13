package vpkg

import (
	"fmt"
	"io"
	"net/http"
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

// FetchRegistry fetches and parses the registry index
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

// FindPackage finds a package by name in the registry
func (r *RegistryClient) FindPackage(name string) (*Package, error) {
	registry, err := r.FetchRegistry()
	if err != nil {
		return nil, err
	}

	for _, pkg := range registry.Packages {
		if pkg.Name == name {
			return &pkg, nil
		}
	}

	return nil, fmt.Errorf("package %s not found in registry", name)
}

// FetchPackageMeta fetches the meta.yaml file for a specific package
func (r *RegistryClient) FetchPackageMeta(packageName string) (*PackageMeta, error) {
	// Convert package name to URL path
	// e.g. "vandor/redis-cache" -> "packages/vandor/redis-cache/meta.yaml"
	parts := strings.Split(packageName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid package name format: %s (expected namespace/name)", packageName)
	}

	baseURL := strings.Replace(r.registryURL, "/registry.yaml", "", 1)
	metaURL := fmt.Sprintf("%s/packages/%s/%s/meta.yaml", baseURL, parts[0], parts[1])

	resp, err := r.httpClient.Get(metaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package meta: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("package meta request failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read package meta: %w", err)
	}

	var meta PackageMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse package meta YAML: %w", err)
	}

	return &meta, nil
}

// FetchPackageFile fetches a specific file from a package
func (r *RegistryClient) FetchPackageFile(packageName, filePath string) ([]byte, error) {
	parts := strings.Split(packageName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid package name format: %s", packageName)
	}

	baseURL := strings.Replace(r.registryURL, "/registry.yaml", "", 1)
	fileURL := fmt.Sprintf("%s/packages/%s/%s/%s", baseURL, parts[0], parts[1], filePath)

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

// ListPackages returns all packages from the registry, optionally filtered
func (r *RegistryClient) ListPackages(opts ListOptions) ([]Package, error) {
	registry, err := r.FetchRegistry()
	if err != nil {
		return nil, err
	}

	var filtered []Package
	for _, pkg := range registry.Packages {
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
