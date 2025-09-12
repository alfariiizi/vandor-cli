# Vandor CLI

A powerful CLI tool for managing Go projects with hexagonal architecture, code
generation, and package management.

## Installation

### 📦 Quick Install (Recommended)

**One-line installation:**

```bash
curl -fsSL https://raw.githubusercontent.com/alfariiizi/vandor-cli/main/install-vandor.sh | bash
```

**Or with wget:**

```bash
wget -qO- https://raw.githubusercontent.com/alfariiizi/vandor-cli/main/install-vandor.sh | bash
```

> **Note**: If no releases are available yet, the script will provide instructions for building from source.

### 💾 Manual Download

1. Visit the
   [releases page](https://github.com/alfariiizi/vandor-cli/releases/latest)
2. Download the binary for your platform
3. Extract and move to PATH:

**Linux/macOS:**

```bash
tar -xzf vandor-linux-amd64.tar.gz
sudo mv vandor /usr/local/bin/
chmod +x /usr/local/bin/vandor
```

**Windows:**

```powershell
Expand-Archive vandor-windows-amd64.zip
Move-Item vandor.exe C:\Windows\System32\
```

### 🔧 Build from Source

**Requirements:** Go 1.21+

```bash
git clone https://github.com/alfariiizi/vandor-cli.git
cd vandor-cli
go build -o vandor main.go
sudo mv vandor /usr/local/bin/
```

### ✅ Verify Installation

```bash
vandor version        # Check installed version
vandor --help         # Show all commands
```

## Quick Start

### Initialize a new project

```bash
vandor init
```

This will create a `vandor-config.yaml` file and optionally set up a full
project structure.

### Launch TUI (Interactive Mode)

```bash
vandor tui
```

This launches an interactive terminal interface for easier project management.

### Set a Beautiful Theme

```bash
vandor theme set mocha    # Dark theme for dark terminals
vandor theme set latte    # Light theme for light terminals
vandor theme set auto     # Auto-detect (default)
```

### Keep Vandor Up to Date

```bash
vandor upgrade check      # Check for updates
vandor upgrade           # Upgrade to latest version (with source fallback)
vandor upgrade source    # Build from source (when no releases available)
```

## Commands

### Project Initialization

- `vandor init` - Initialize a new Vandor project with configuration

### Component Management

- `vandor add schema <name>` - Create a new database schema
- `vandor add domain <name>` - Create a new domain
- `vandor add usecase <name>` - Create a new usecase
- `vandor add service <group> <name>` - Create a new service
- `vandor add job <name>` - Create a new background job
- `vandor add scheduler <name>` - Create a new scheduler
- `vandor add enum <name>` - Create a new enum
- `vandor add seed <name>` - Create a new database seed
- `vandor add handler <group> <name> <method>` - Create a new HTTP handler
- `vandor add handler-crud <model>` - Generate CRUD handlers for a model
- `vandor add service-handler <group> <name> <method>` - Create service and
  handler together

### Code Generation

- `vandor sync all` - Generate all code
- `vandor sync core` - Generate core components (domains, usecases, services)
- `vandor sync domain` - Generate domain code
- `vandor sync usecase` - Generate usecase code
- `vandor sync service` - Generate service code
- `vandor sync job` - Generate job code
- `vandor sync scheduler` - Generate scheduler code
- `vandor sync enum` - Generate enum code
- `vandor sync seed` - Generate seed code
- `vandor sync handler` - Generate HTTP handler code
- `vandor sync db-model` - Generate database models using Ent

### Package Management (vpkg)

- `vandor vpkg add <package-name>` - Add a Vandor package
- `vandor vpkg remove <package-name>` - Remove a Vandor package
- `vandor vpkg list` - List installed packages
- `vandor vpkg search [query]` - Search available packages
- `vandor vpkg update [package-name]` - Update packages

### Utility Commands

- `vandor version` - Show version information
  - `vandor version --detailed` - Show detailed system information
- `vandor tui` - Launch interactive TUI
- `vandor help` - Show help information

### Theme Management

- `vandor theme list` - List available themes
- `vandor theme set <theme-name>` - Set active theme
- `vandor theme info` - Show current theme information

**Available themes:**

- `mocha` - Catppuccin Mocha (dark theme)
- `latte` - Catppuccin Latte (light theme)
- `frappe` - Catppuccin Frappe (medium contrast)
- `dracula` - Dracula inspired theme
- `auto` - Auto-detect based on system theme (default)

### Update Management

- `vandor upgrade` - Upgrade to the latest version (with fallback to source)
- `vandor upgrade check` - Check if newer version is available  
- `vandor upgrade source` - Build and install from source code
- `vandor install script` - Generate installation script (for developers)
- `vandor install instructions` - Show installation instructions
- `vandor uninstall` - Uninstall Vandor CLI from system

## Configuration

The `vandor-config.yaml` file contains project configuration:

```yaml
project:
  name: my-clinic-app
  module: github.com/your-org/my-clinic-app
  version: "0.1.0"

vandor:
  cli: "0.5.0"
  architecture: "full-backend" # full-backend, eda, or minimal
  language: go

vpkg:
  - name: audit-logger
    version: "1.0.0"
    tags: [full-backend, eda]
  - name: redis-cache
    version: "1.2.0"
    tags: [full-backend, eda, minimal]
```

## Architecture Types

- **full-backend**: Complete backend with all features
- **eda**: Event-driven architecture
- **minimal**: Minimal setup for lightweight projects

## Vandor Packages (vpkg)

Vandor packages are reusable components that can be easily installed into your
project. They include:

- **Official packages**: Maintained by the Vandor team
- **Community packages**: Created by the community
- Setup instructions and documentation
- Compatible with different architecture types

### Popular Packages

- `redis-cache` - Redis caching integration
- `audit-logger` - Audit logging functionality
- `kafka-bus` - Kafka event bus integration

## Development

### Prerequisites

- Go 1.21 or later
- Git

### Building

```bash
go build -o vandor
```

### Running Tests

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for
details.

## Support

- GitHub Issues:
  [Report bugs or request features](https://github.com/alfariiizi/vandor-cli/issues)
- Documentation: [Full documentation](https://docs.vandor.dev)
- Community: [Discord Server](https://discord.gg/vandor)
