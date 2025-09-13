# GitHub Actions

This directory contains GitHub Actions workflows for the Vandor CLI project.

## Workflows

### `go.yml` - Main CI/CD Pipeline

This workflow handles:

- **Testing**: Runs tests with coverage reporting
- **Linting**: Code quality checks with golangci-lint
- **Building**: Cross-platform binary builds for Linux, macOS, and Windows
- **Releasing**: Automated GitHub releases with binaries when tags are pushed
- **Docker**: Multi-platform Docker images published to GitHub Container Registry

#### Triggers

- **Push**: `main` and `develop` branches, version tags (`v*`)
- **Pull Request**: `main` branch

#### Artifacts

- Cross-platform binaries (Linux, macOS, Windows)
- Docker images for `linux/amd64` and `linux/arm64`
- Release archives (`.tar.gz` for Unix, `.zip` for Windows)

#### Build Matrix

| OS | Architecture | Output |
|----|--------------|--------|
| Linux | amd64, arm64 | `vandor-linux-amd64`, `vandor-linux-arm64` |
| macOS | amd64, arm64 | `vandor-darwin-amd64`, `vandor-darwin-arm64` |
| Windows | amd64 | `vandor-windows-amd64.exe` |

#### Version Information

Build-time variables are injected via ldflags:
- `version` - Git tag or "dev"
- `commit` - Short commit hash
- `date` - Build timestamp

#### Release Process

1. Create a git tag: `git tag v1.0.0`
2. Push the tag: `git push origin v1.0.0`  
3. GitHub Actions automatically creates a release with:
   - Generated changelog
   - Cross-platform binaries
   - Installation instructions

#### Docker Images

Images are published to `ghcr.io/alfariiizi/vandor-cli` with tags:
- `main` - Latest main branch
- `v1.0.0` - Specific version tags
- `1.0` - Major.minor version
- `sha-abc123` - Commit-specific tags