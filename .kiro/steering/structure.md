# Project Structure

## Root Directory Layout

```
fabric/
├── cmd/                    # Main applications
│   ├── fabric/            # Primary CLI application
│   ├── code_helper/       # Code analysis helper
│   ├── generate_changelog/ # Changelog generation tool
│   └── to_pdf/            # LaTeX to PDF converter
├── internal/              # Private application code
├── data/                  # Static data files
├── web/                   # Web interface (Svelte)
├── scripts/               # Build and utility scripts
├── docs/                  # Documentation
├── completions/           # Shell completion scripts
└── nix/                   # Nix configuration
```

## Internal Package Organization

```
internal/
├── cli/                   # Command-line interface logic
├── core/                  # Core business logic and plugin registry
├── domain/                # Domain models and types
├── plugins/               # Plugin system
│   ├── ai/               # AI vendor integrations
│   ├── db/               # Database operations
│   ├── strategy/         # Prompt strategies
│   └── template/         # Template processing
├── server/               # Web server implementation
├── tools/                # Utility tools
└── util/                 # Common utilities
```

## Data Directory Structure

```
data/
├── patterns/             # AI prompt patterns (200+ subdirectories)
│   ├── summarize/       # Each pattern has its own directory
│   │   ├── system.md    # System prompt
│   │   └── user.md      # User prompt (optional)
│   └── analyze_claims/
└── strategies/          # Prompt strategies (JSON files)
    ├── cot.json        # Chain of Thought
    ├── cod.json        # Chain of Draft
    └── standard.json   # Standard strategy
```

## Web Interface Structure

```
web/
├── src/
│   ├── lib/             # Reusable components
│   ├── routes/          # SvelteKit routes
│   └── app.html         # Main HTML template
├── static/              # Static assets
├── package.json         # Node.js dependencies
└── svelte.config.js     # Svelte configuration
```

## Configuration Files

- `go.mod` / `go.sum` - Go module dependencies
- `flake.nix` - Nix development environment
- `.devcontainer/` - VS Code dev container setup
- `scripts/docker/` - Docker configuration
- `completions/` - Shell completion scripts for bash, zsh, fish

## Key Conventions

### Pattern Structure

- Each pattern lives in `data/patterns/{pattern_name}/`
- `system.md` contains the main prompt
- `user.md` contains user-specific instructions (optional)
- Patterns use Markdown for maximum readability

### Go Package Naming

- `cmd/` for executable commands
- `internal/` for private packages (not importable by external projects)
- Domain-driven organization within `internal/`
- Plugin-based architecture for extensibility

### File Naming

- Go files use snake_case
- Markdown files use lowercase with underscores
- Configuration files follow standard conventions (.json, .yaml, .md)

### Import Organization

- Standard library imports first
- Third-party imports second
- Local imports last
- Grouped with blank lines between sections

## Build Artifacts

- Binary output: `fabric` (main CLI)
- Web build output: `web/build/`
- Docker images: Multi-stage builds with Alpine base
- Nix packages: Reproducible builds via flake.nix
