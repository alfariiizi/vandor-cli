package vpkg

import (
	"encoding/json"
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

// DiscoverTemplateFiles discovers all template files in a package's templates directory
// Supports multiple template extensions: .tmpl, .templ, .gotmpl
// Uses optimized discovery with maximum 3 levels of nesting for performance
func (r *RegistryClient) DiscoverTemplateFiles(packageWithRepo *PackageWithRepo, templatesDir string) ([]string, error) {
	// Try GitHub API first for accurate directory structure
	files, err := r.discoverViaGitHubAPI(packageWithRepo, templatesDir)
	if err == nil && len(files) > 0 {
		return files, nil
	}

	// Fallback to optimized pattern matching with depth limit
	files, err = r.discoverViaOptimizedPatterns(packageWithRepo, templatesDir)
	if err == nil && len(files) > 0 {
		return files, nil
	}

	return nil, fmt.Errorf("no template files found in %s", templatesDir)
}

// discoverViaGitHubAPI uses GitHub API to efficiently discover all template files
func (r *RegistryClient) discoverViaGitHubAPI(packageWithRepo *PackageWithRepo, templatesDir string) ([]string, error) {
	// Extract GitHub repo info from MetaURL
	// https://raw.githubusercontent.com/user/repo/main/meta.yaml -> user/repo
	repoURL := packageWithRepo.RepositoryInfo.MetaURL
	if !strings.Contains(repoURL, "github.com") {
		return nil, fmt.Errorf("not a GitHub repository")
	}

	// Parse GitHub repo from URL
	parts := strings.Split(repoURL, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid GitHub URL format")
	}

	owner := parts[3]
	repo := parts[4]
	branch := "main" // Default to main branch

	// GitHub API URL for directory contents
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		owner, repo, templatesDir, branch)

	return r.fetchGitHubDirectoryContents(apiURL, "")
}

// fetchGitHubDirectoryContents recursively fetches directory contents from GitHub API
// Limited to maximum depth of 3 levels for performance
func (r *RegistryClient) fetchGitHubDirectoryContents(apiURL, pathPrefix string) ([]string, error) {
	return r.fetchGitHubDirectoryContentsWithDepth(apiURL, pathPrefix, 0, 3)
}

// fetchGitHubDirectoryContentsWithDepth recursively fetches with depth limit
func (r *RegistryClient) fetchGitHubDirectoryContentsWithDepth(apiURL, pathPrefix string, currentDepth, maxDepth int) ([]string, error) {
	if currentDepth > maxDepth {
		return nil, nil // Stop recursion at max depth
	}

	resp, err := r.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed with status %d", resp.StatusCode)
	}

	var items []struct {
		Name string `json:"name"`
		Type string `json:"type"`
		URL  string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	var templateFiles []string
	templateExtensions := []string{".tmpl", ".templ", ".gotmpl"}

	for _, item := range items {
		itemPath := item.Name
		if pathPrefix != "" {
			itemPath = filepath.Join(pathPrefix, item.Name)
		}

		if item.Type == "file" {
			// Check if it's a template file
			for _, ext := range templateExtensions {
				if strings.HasSuffix(item.Name, ext) {
					templateFiles = append(templateFiles, itemPath)
					break
				}
			}
		} else if item.Type == "dir" && currentDepth < maxDepth {
			// Recursively fetch directory contents with depth limit
			subFiles, err := r.fetchGitHubDirectoryContentsWithDepth(item.URL, itemPath, currentDepth+1, maxDepth)
			if err == nil {
				templateFiles = append(templateFiles, subFiles...)
			}
		}
	}

	return templateFiles, nil
}

// discoverViaOptimizedPatterns uses optimized pattern matching with maximum 3 levels depth
func (r *RegistryClient) discoverViaOptimizedPatterns(packageWithRepo *PackageWithRepo, templatesDir string) ([]string, error) {
	var templateFiles []string
	baseURL := strings.Replace(packageWithRepo.RepositoryInfo.MetaURL, "/meta.yaml", "", 1)
	templateExtensions := []string{".tmpl", ".templ", ".gotmpl"}

	// Get package name for targeted patterns
	parts := strings.Split(packageWithRepo.Package.Name, "/")
	pkgName := parts[len(parts)-1] // e.g., "redis-cache"

	// Define limited, high-probability patterns with max 3 levels
	patterns := []string{
		// Level 0: Root files
		"README.md",
		"main.go",
		pkgName + ".go",
		strings.Split(pkgName, "-")[0] + ".go", // redis.go from redis-cache

		// Level 1: Common subdirectories
		"cmd/main.go",
		"internal/handler.go",
		"pkg/service.go",
		"somDir/test.go", // Your specific case
		"src/main.go",

		// Level 2: Two levels deep
		"cmd/server/main.go",
		"internal/pkg/handler.go",
		"pkg/internal/service.go",
		"src/internal/main.go",
		"cmd/cli/main.go",

		// Level 3: Three levels deep (maximum)
		"cmd/server/internal/main.go",
		"internal/pkg/handler/service.go",
		"pkg/internal/service/handler.go",
	}

	// Try each pattern with all template extensions
	for _, pattern := range patterns {
		for _, ext := range templateExtensions {
			templateFile := pattern + ext
			if r.tryTemplateFile(baseURL, templatesDir, templateFile) {
				templateFiles = append(templateFiles, templateFile)
			}
		}
	}

	if len(templateFiles) == 0 {
		return nil, fmt.Errorf("no template files found with optimized patterns")
	}

	return templateFiles, nil
}

// tryTemplateFile checks if a template file exists at the given path
func (r *RegistryClient) tryTemplateFile(baseURL, templatesDir, pattern string) bool {
	testURL := fmt.Sprintf("%s/%s/%s", baseURL, templatesDir, pattern)

	// Create a client with shorter timeout for existence checks
	client := &http.Client{
		Timeout: 5 * time.Second, // Shorter timeout for existence checks
	}

	resp, err := client.Get(testURL)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
