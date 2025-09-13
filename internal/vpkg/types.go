package vpkg

import "time"

// Registry represents the main registry index
type Registry struct {
	Version     string     `yaml:"version"`
	RegistryURL string     `yaml:"registry_url"`
	Packages    []Package  `yaml:"packages"`
	Tags        []Tag      `yaml:"tags"`
}

// Package represents a package in the registry
type Package struct {
	Name         string   `yaml:"name"`
	Title        string   `yaml:"title"`
	Description  string   `yaml:"description"`
	Type         string   `yaml:"type"` // fx-module, cli-command
	Version      string   `yaml:"version"`
	License      string   `yaml:"license"`
	Author       string   `yaml:"author"`
	Repository   string   `yaml:"repository"`
	Tags         []string `yaml:"tags,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty"`
}

// Tag represents a category tag
type Tag struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// PackageMeta represents package metadata from meta.yaml
type PackageMeta struct {
	Name         string   `yaml:"name"`
	Title        string   `yaml:"title"`
	Description  string   `yaml:"description"`
	Type         string   `yaml:"type"`
	Entry        string   `yaml:"entry"`
	Templates    []string `yaml:"templates"`
	Destination  string   `yaml:"destination"`
	Version      string   `yaml:"version"`
	License      string   `yaml:"license"`
	Author       string   `yaml:"author"`
	Tags         []string `yaml:"tags,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty"`
}

// InstalledPackage represents a locally installed package
type InstalledPackage struct {
	Name        string    `yaml:"name"`
	Version     string    `yaml:"version"`
	InstalledAt time.Time `yaml:"installed_at"`
	Path        string    `yaml:"path"`
	Type        string    `yaml:"type"`
	Meta        PackageMeta `yaml:"meta"`
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
	Registry   string
	Dest       string
	Force      bool
	DryRun     bool
	Version    string
}

// ListOptions holds options for listing packages
type ListOptions struct {
	Registry string
	Tags     []string
	Type     string
}