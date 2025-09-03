# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

Fabric is an open-source framework for augmenting humans using AI. It organizes AI prompts by real-world tasks called "Patterns" and provides both CLI and web interfaces for executing these patterns against various AI providers (OpenAI, Anthropic, Gemini, Ollama, etc.).

## Common Commands

### Building and Running

```bash
# Build main fabric binary
go build -o fabric cmd/fabric/main.go

# Build terminal UI
go build -o fabric-tui cmd/fabric-tui/main.go

# Install from source
go install github.com/danielmiessler/fabric/cmd/fabric@latest

# Run setup (creates config directory and patterns)
fabric --setup

# Update patterns from repository
fabric --updatepatterns

# Launch terminal UI
fabric --tui
# or
fabric -i

# Start API server (default port 8080)
fabric --serve --address "0.0.0.0:8080"
```

### Testing

```bash
# Run all Go tests
go test ./...

# Run specific package tests
go test ./internal/core/
go test ./internal/plugins/ai/
go test ./internal/cli/

# Run single test function
go test -run TestFunctionName ./internal/core/
go test -run TestChatter ./internal/core/

# Run tests with verbose output
go test -v ./...

# Check formatting (requires Nix)
nix flake check
```

### Web Interface Development

```bash
# Navigate to web directory
cd web/

# Install dependencies 
pnpm install

# Run development server
pnpm run dev

# Build for production
pnpm run build

# Run linting
pnpm run lint

# Format code
pnpm run format

# Run web frontend tests
pnpm test
```

## Architecture

### Core Structure

**Main Components:**
- `cmd/`: Entry points for executables (fabric, fabric-api, fabric-tui, code_helper, to_pdf)
- `internal/`: Core business logic
  - `cli/`: Command-line interface and flag handling
  - `core/`: Core domain logic (chatter, plugin registry)
  - `plugins/`: Modular provider system
    - `ai/`: AI provider integrations (OpenAI, Anthropic, Gemini, Ollama, etc.)
    - `db/`: Database abstraction (fsdb, sqlite)
    - `pattern/`: Pattern management and execution
    - `template/`: Template engine for extensibility
  - `server/`: REST API components
  - `tui/`: Terminal UI using Bubble Tea framework
- `data/patterns/`: Built-in AI prompts organized by task
- `web/`: SvelteKit-based web interface

### Key Concepts

**Patterns**: AI prompts stored as directories under `data/patterns/`, each containing:
- `system.md`: Main prompt content
- `user.md`: Optional user prompt template
- `README.md`: Optional documentation

**Plugin Architecture**: All AI providers, databases, and patterns use a unified plugin registry system allowing dynamic loading and configuration.

**Conversation Management**: The `Chatter` interface handles multi-turn conversations with context preservation across all providers.

**Extension System**: Template-based extensibility through `internal/plugins/template/` allowing custom functionality.

## Development Patterns

### Adding New AI Providers
1. Create provider in `internal/plugins/ai/your_provider/`
2. Implement the provider interface matching existing patterns
3. Register in `internal/plugins/ai/vendors.go`
4. Add CLI flags in `internal/cli/flags.go`

### Adding New Patterns
1. Create directory: `data/patterns/pattern_name/`
2. Add `system.md` with prompt content
3. Test with: `fabric --pattern pattern_name`

### Terminal UI Development
- Uses Bubble Tea's Model-View-Update architecture
- Components in `internal/tui/` implement `tea.Model` interface
- State centralized in main App model
- Styling with Lipgloss for consistency

### Web Interface Development
- Framework: SvelteKit with TypeScript
- Styling: TailwindCSS with Skeleton UI
- Build: Vite
- API communication via REST endpoints in Go backend

## Important Files

- `cmd/fabric/main.go`: Main CLI entry point
- `internal/cli/cli.go`: Core CLI logic and command handling
- `internal/core/chatter.go`: Conversation management
- `internal/plugins/ai/vendors.go`: AI provider registration
- `internal/tui/app.go`: Terminal UI application state
- `web/src/routes/`: Web interface pages and API routes

## Configuration

- User config: `~/.config/fabric/`
- Patterns: `~/.config/fabric/patterns/`
- Environment variables: `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, etc.
- Custom patterns: User-configurable separate directory
- Per-pattern model mapping: `FABRIC_MODEL_PATTERN_NAME=vendor|model`

## Docker Development

```bash
# Build Docker image
docker build -t fabric .

# Run with environment variables
docker run -e OPENAI_API_KEY=your_key -p 8080:8080 fabric
```

## CI/CD

GitHub Actions workflow runs on push/PR to main:
- Runs Go tests with `go test -v ./...`
- Checks formatting with Nix flake
- Excludes pattern and markdown changes from triggering builds
