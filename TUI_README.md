# Fabric Terminal User Interface (TUI)

A beautiful, interactive terminal interface for Fabric built with Bubble Tea.

## Features

- 🎯 **Pattern Browser**: Browse and search through all available Fabric patterns
- 💬 **Interactive Chat**: Chat with AI using selected patterns in real-time  
- ⚙️ **Settings Management**: Configure models, API keys, and preferences
- 🎨 **Beautiful UI**: Modern terminal interface with syntax highlighting
- ⌨️ **Keyboard Navigation**: Full keyboard support with vim-style bindings

## Installation

### Prerequisites

Make sure Go 1.24+ is installed:
```bash
go version
```

### Install Dependencies

```bash
# Install Bubble Tea dependencies
go mod tidy
```

### Build the TUI

```bash
# Build the standalone TUI
go build -o fabric-tui cmd/fabric-tui/main.go

# Or build with the main fabric binary (includes -i/--tui flag)
go build -o fabric cmd/fabric/main.go
```

## Usage

### Standalone TUI
```bash
./fabric-tui
```

### Integrated with main Fabric CLI
```bash
# Launch TUI mode
./fabric --tui
# or
./fabric -i
```

## Interface Overview

The TUI has four main views accessible via Tab navigation:

### 1. Home View
- Welcome screen with navigation options
- Quick access to main features

### 2. Pattern Browser
- Browse all available Fabric patterns
- Search/filter patterns by name
- View pattern descriptions
- Select patterns for chat

### 3. Chat Interface  
- Real-time chat with selected pattern
- Message history with timestamps
- Streaming responses (when supported)
- User and Assistant message styling

### 4. Settings
- Configure default AI model
- Set temperature and other parameters
- Manage API keys
- Set custom patterns directory

## Keyboard Shortcuts

### Global
- `Tab` - Next view
- `Shift+Tab` - Previous view  
- `q` or `Ctrl+C` - Quit
- `?` - Help

### Pattern Browser
- `↑/↓` or `j/k` - Navigate patterns
- `/` - Filter/search patterns
- `Enter` - Select pattern and go to chat
- `Esc` - Clear filter

### Chat Interface
- Type and `Enter` - Send message
- `Esc` - Focus/unfocus input
- `↑/↓` - Scroll message history

### Settings
- `↑/↓` or `j/k` - Navigate settings
- `Enter` - Edit setting (future implementation)

## Architecture

The TUI is built using the Bubble Tea framework with the following components:

- `internal/tui/app.go` - Main application and model
- `internal/tui/patterns.go` - Pattern browser component
- `internal/tui/chat.go` - Chat interface component  
- `internal/tui/settings.go` - Settings management
- `internal/tui/keys.go` - Keyboard shortcuts
- `cmd/fabric-tui/main.go` - Standalone TUI entry point

## Integration with Fabric

The TUI integrates seamlessly with Fabric's existing architecture:

- Uses the same pattern system from `data/patterns/`
- Leverages existing AI provider integrations
- Shares configuration with main Fabric CLI
- Accesses the same plugin registry

## Development

### Adding New Views

1. Create a new component file in `internal/tui/`
2. Implement the Bubble Tea Model interface:
   ```go
   type MyModel struct { /* ... */ }
   func (m MyModel) Init() tea.Cmd { /* ... */ }
   func (m MyModel) Update(tea.Msg) (MyModel, tea.Cmd) { /* ... */ }
   func (m MyModel) View() string { /* ... */ }
   ```
3. Add the view to the main app state machine
4. Update keyboard shortcuts and help text

### Styling

The TUI uses Lipgloss for styling. Key style definitions are in:
- `internal/tui/app.go` - Global styles
- Component files - Component-specific styles

### Testing

```bash
# Run tests
go test ./internal/tui/...

# Run with race detection
go test -race ./internal/tui/...
```

## Future Enhancements

- [ ] **Real Fabric Integration** - Replace mock data with actual Fabric patterns and AI calls
- [ ] **Pattern Editor** - Create and edit patterns directly in the TUI
- [ ] **Session Management** - Save and restore chat sessions
- [ ] **File Attachments** - Support for image and document inputs
- [ ] **Syntax Highlighting** - Code syntax highlighting in responses
- [ ] **Export Options** - Export chats to various formats
- [ ] **Themes** - Multiple color themes and customization
- [ ] **Plugin System** - Extensible plugin architecture
- [ ] **Context Management** - Visual context/memory management
- [ ] **Model Comparison** - Side-by-side model comparisons

## Troubleshooting

### TUI doesn't start
- Ensure terminal supports ANSI colors and mouse
- Try with `TERM=xterm-256color`

### Patterns not loading
- Run `fabric --updatepatterns` first
- Check `~/.config/fabric/patterns/` directory exists

### Performance issues
- Reduce terminal size if rendering is slow
- Check for high CPU usage from other processes

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

## License

Same as main Fabric project - MIT License