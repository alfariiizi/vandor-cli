package vpkg

import "time"

// Registry represents the main registry index (new repository-based format)
type Registry struct {
	Version      string           `yaml:"version"`
	RegistryURL  string           `yaml:"registry_url"`
	Repositories []RepositoryInfo `yaml:"repositories"`
	Tags         []Tag            `yaml:"tags"`
}

// RepositoryInfo represents a repository entry in the registry
type RepositoryInfo struct {
	Name       string `yaml:"name"`       // e.g. "vpkg-vandor-official"
	Repository string `yaml:"repository"` // GitHub repository URL
	MetaURL    string `yaml:"meta_url"`   // URL to meta.yaml
	Author     string `yaml:"author"`     // Repository author
	Verified   bool   `yaml:"verified"`   // Whether repository is verified
}

// RepositoryMeta represents the meta.yaml from a repository (new array format)
type RepositoryMeta struct {
	Version    string    `yaml:"version"`
	Repository string    `yaml:"repository"`
	Author     string    `yaml:"author"`
	License    string    `yaml:"license"`
	Packages   []Package `yaml:"packages"` // Array of packages in this repository
}

// Package represents a single package (now part of RepositoryMeta.Packages)
type Package struct {
	Name         string   `yaml:"name"`
	Title        string   `yaml:"title"`
	Description  string   `yaml:"description"`
	Type         string   `yaml:"type"`      // fx-module, cli-command
	Templates    string   `yaml:"templates"` // Directory path - all templates auto-discovered
	Destination  string   `yaml:"destination"`
	Version      string   `yaml:"version"`
	Tags         []string `yaml:"tags,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty"`
}

// Tag represents a category tag
type Tag struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// PackageWithRepo combines package info with its repository context
type PackageWithRepo struct {
	Package        Package        `json:"package"`
	RepositoryInfo RepositoryInfo `json:"repository"`
	RepositoryMeta RepositoryMeta `json:"repo_meta"`
}

// InstalledPackage represents a locally installed package
type InstalledPackage struct {
	Name        string    `yaml:"name"`
	Version     string    `yaml:"version"`
	InstalledAt time.Time `yaml:"installed_at"`
	Path        string    `yaml:"path"`
	Type        string    `yaml:"type"`
	Meta        Package   `yaml:"meta"` // Use Package type instead of PackageMeta
}

// TemplateContext provides data for template rendering
type TemplateContext struct {
	Module      string // Go module path from go.mod
	VpkgName    string // e.g. "vandor/redis-cache"
	Namespace   string // e.g. "vandor"
	Pkg         string // e.g. "redis-cache"
	Package     string // sanitized package name e.g. "rediscache"
	PackagePath string // e.g. "internal/vpkg/vandor/redis-cache"
	Version     string
	Author      string
	Time        string
	Title       string
	Description string
}

// InstallOptions holds options for package installation
type InstallOptions struct {
	Registry string
	Dest     string
	Force    bool
	DryRun   bool
	Version  string
}

// ListOptions holds options for listing packages
type ListOptions struct {
	Registry string
	Tags     []string
	Type     string
}
