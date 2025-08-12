# Technology Stack

## Primary Technologies

- **Language**: Go 1.24+
- **CLI Framework**: jessevdk/go-flags for command-line parsing
- **Web Framework**: Gin (Go web framework)
- **Database**: SQLite3 for local storage
- **Frontend**: Svelte/SvelteKit with TypeScript
- **Package Manager**: pnpm (for web interface)

## Key Dependencies

### Backend (Go)

- `github.com/anthropics/anthropic-sdk-go` - Anthropic AI integration
- `github.com/openai/openai-go` - OpenAI integration
- `github.com/ollama/ollama` - Local LLM support
- `github.com/gin-gonic/gin` - Web server
- `github.com/mattn/go-sqlite3` - Database
- `github.com/atotto/clipboard` - Clipboard operations
- `github.com/go-git/go-git/v5` - Git operations
- `gopkg.in/yaml.v3` - YAML configuration

### Frontend (Web)

- Svelte/SvelteKit framework
- TypeScript for type safety
- Tailwind CSS for styling
- Skeleton UI components
- Vite for build tooling

## Build System

### Go Application

```bash
# Install from source
go install github.com/danielmiessler/fabric/cmd/fabric@latest

# Build locally
go build -o fabric ./cmd/fabric

# Run tests
go test ./...
```

### Web Interface

```bash
# Install dependencies
pnpm install

# Development server
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview
```

## Development Environment

- **Nix Flake**: Available for reproducible development environment
- **Docker**: Containerized deployment with multi-stage builds
- **Dev Container**: VS Code dev container support
- **Go Version**: Requires Go 1.24+ (uses toolchain go1.24.2)

## Common Commands

```bash
# Setup fabric configuration
fabric --setup

# List available patterns
fabric --listpatterns

# Use a pattern
fabric --pattern summarize < input.txt

# Stream output
fabric --stream --pattern analyze_claims

# Update patterns
fabric --updatepatterns

# Start web server
fabric --serve

# Start with Ollama endpoints
fabric --serveOllama
```

## Architecture Notes

- Modular plugin system for AI vendors
- Pattern-based prompt organization
- Session and context management
- Support for custom patterns directory
- RESTful API for web interface integration
