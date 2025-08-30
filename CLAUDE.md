# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Fabric is an open-source framework for augmenting humans using AI. It organizes AI prompts by real-world tasks called "Patterns" and provides both CLI and web interfaces for executing these patterns against various AI providers (OpenAI, Anthropic, Gemini, Ollama, etc.).

## Common Commands

### Go Development (Main Application)
```bash
# Build the main fabric binary
go build -o fabric cmd/fabric/main.go

# Build the terminal UI
go build -o fabric-tui cmd/fabric-tui/main.go

# Run tests 
go test ./...

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
```

### Web Interface Development
```bash
# Navigate to web directory
cd web/

# Install dependencies 
npm install
# or 
pnpm install

# Run development server
npm run dev
# or
pnpm run dev

# Build for production
npm run build

# Run linting
npm run lint

# Run tests
npm test
```

### Testing
```bash
# Run all Go tests
go test ./...

# Run specific package tests
go test ./internal/core/
go test ./internal/plugins/ai/

# Run web frontend tests
cd web/ && npm test
```

## Architecture Overview

### Core Architecture
- **cmd/**: Main executables (fabric, fabric-api, fabric-tui, code_helper, to_pdf)
- **internal/**: Core business logic organized by domain
  - **cli/**: Command-line interface and flag handling
  - **core/**: Core domain logic (chatter, plugin registry)
  - **plugins/**: AI providers, database, patterns, templates
  - **server/**: REST API server components
  - **tools/**: Utility functions and converters
  - **tui/**: Terminal user interface components (Bubble Tea)
- **data/**: Patterns (AI prompts) and strategies
- **web/**: Svelte-based web interface

### Key Concepts
- **Patterns**: Structured AI prompts stored as Markdown files in directories under `data/patterns/`
- **Providers**: AI service integrations (OpenAI, Anthropic, Gemini, etc.)
- **Contexts**: Reusable prompt contexts stored in user config
- **Sessions**: Conversation state management
- **Extensions**: Template-based system for custom functionality

### Pattern Structure
Each pattern is a directory containing:
- `system.md`: The main prompt content
- `user.md`: Optional user prompt template
- `README.md`: Optional pattern documentation

### Configuration
- User config directory: `~/.config/fabric/`
- Patterns directory: `~/.config/fabric/patterns/`
- Custom patterns: User-configurable directory (separate from built-in patterns)
- Environment variables for API keys: `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, etc.

## Development Workflow

### Adding New Patterns
1. Create directory: `data/patterns/new_pattern_name/`
2. Add `system.md` with the prompt content
3. Optionally add `user.md` for user input templates
4. Test with: `fabric --pattern new_pattern_name`

### Adding AI Providers
1. Create new provider in `internal/plugins/ai/`
2. Implement the provider interface
3. Register in `internal/plugins/ai/vendors.go`
4. Add configuration options in CLI flags

### Web Interface Development
- Framework: SvelteKit with TypeScript
- Styling: TailwindCSS with Skeleton UI
- Build: Vite
- Testing: Vitest
- API communication with Go backend via REST endpoints

### Terminal UI Development
- Framework: Bubble Tea (Charmbracelet)
- Components: Bubbles for interactive elements
- Styling: Lipgloss for terminal styling
- Architecture: Model-View-Update (MVU) pattern
- Integration: Shares core Fabric registry and patterns

### Testing Approach
- Go: Standard `go test` framework with test files ending in `_test.go`
- Web: Vitest for unit testing, manual testing for integration
- Pattern validation through CLI execution

## Key Files to Understand
- `cmd/fabric/main.go`: Main CLI entry point
- `cmd/fabric-tui/main.go`: Terminal UI entry point
- `internal/cli/cli.go`: Core CLI logic and command handling
- `internal/tui/app.go`: Main TUI application and state management
- `internal/plugins/ai/vendors.go`: AI provider registration
- `internal/core/chatter.go`: Core conversation logic
- `web/src/routes/`: Web interface pages and API routes
- `data/patterns/`: All built-in AI patterns

## Common Patterns for Changes
- AI providers: Follow existing provider structure in `internal/plugins/ai/`
- CLI flags: Add in `internal/cli/flags.go` and corresponding handlers
- Web features: Create Svelte components in `web/src/lib/components/`
- TUI components: Create new views in `internal/tui/` following Bubble Tea MVU pattern
- Database operations: Use the plugin pattern in `internal/plugins/db/`
- New executables: Add to `cmd/` directory with proper module structure

## TUI Development Notes
- The terminal UI uses Bubble Tea's Model-View-Update architecture
- All TUI components should implement the `tea.Model` interface
- State management is centralized in the main App model
- Use Lipgloss for consistent styling across components
- Integration with Fabric core happens through the shared plugin registry
- See `TUI_README.md` for detailed TUI development guidance