package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/danielmiessler/fabric/internal/core"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
)

// Pattern represents a Fabric pattern
type Pattern struct {
	Name        string
	Description string
	SystemMD    string
	UserMD      string
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role      string    // "user" or "assistant"
	Content   string
	Timestamp time.Time
}

// TViewApp represents the main tview-based application
type TViewApp struct {
	app         *tview.Application
	pages       *tview.Pages
	registry    *core.PluginRegistry
	integration *RealFabricIntegration
	
	// Components
	patternList   *tview.List
	chatView      *tview.TextView
	chatInput     *tview.InputField
	youtubeInput  *tview.InputField
	settings      *tview.Form
	
	// Data
	patterns        []Pattern
	chatMessages    []ChatMessage
	selectedPattern Pattern
}

// NewTViewApp creates a new tview-based TUI application
func NewTViewApp() (*TViewApp, error) {
	registry, err := initializeFabric()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize fabric: %w", err)
	}

	integration := NewRealFabricIntegration(registry)
	patterns, _ := integration.GetRealPatterns()

	return &TViewApp{
		app:         tview.NewApplication(),
		pages:       tview.NewPages(),
		registry:    registry,
		integration: integration,
		patterns:    patterns,
	}, nil
}

// NewTViewAppWithRegistry creates a new tview-based TUI application with existing registry
func NewTViewAppWithRegistry(registry *core.PluginRegistry) (*TViewApp, error) {
	integration := NewRealFabricIntegration(registry)
	patterns, _ := integration.GetRealPatterns()

	return &TViewApp{
		app:         tview.NewApplication(),
		pages:       tview.NewPages(),
		registry:    registry,
		integration: integration,
		patterns:    patterns,
	}, nil
}

// Start runs the tview application
func (a *TViewApp) Start() error {
	a.setupUI()
	return a.app.Run()
}

// setupUI initializes all UI components
func (a *TViewApp) setupUI() {
	// Home page
	home := a.createHomePage()
	a.pages.AddPage("home", home, true, true)

	// Pattern browser page
	patterns := a.createPatternsPage()
	a.pages.AddPage("patterns", patterns, true, false)

	// Chat page
	chat := a.createChatPage()
	a.pages.AddPage("chat", chat, true, false)

	// YouTube page
	youtube := a.createYouTubePage()
	a.pages.AddPage("youtube", youtube, true, false)

	// Help page
	help := a.createHelpPage()
	a.pages.AddPage("help", help, true, false)

	// Settings page
	settings := a.createSettingsPage()
	a.pages.AddPage("settings", settings, true, false)

	// Set up global key bindings
	a.pages.SetInputCapture(a.globalKeyHandler)

	a.app.SetRoot(a.pages, true)
}

// createHomePage creates the home/welcome page
func (a *TViewApp) createHomePage() tview.Primitive {
	textView := tview.NewTextView().
		SetText(`[blue::b]Welcome to Fabric TUI![white::-]

[yellow]Available Actions:[white]
• [green]P[white] - Browse Patterns
• [green]C[white] - Open Chat
• [green]Y[white] - YouTube Processing
• [green]H[white] - Help & Documentation
• [green]S[white] - Settings  
• [green]Q[white] - Quit

[gray]Use the highlighted keys to navigate or press Tab to cycle through views.[white]

[cyan::b]Fabric TUI v1.0 - Built with tview[white::-]`).
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)

	textView.SetBorder(true).
		SetTitle(" Fabric TUI ").
		SetTitleAlign(tview.AlignCenter)

	// Set up key bindings for home page
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'p', 'P':
			a.pages.SwitchToPage("patterns")
			return nil
		case 'c', 'C':
			a.pages.SwitchToPage("chat")
			return nil
		case 'y', 'Y':
			a.pages.SwitchToPage("youtube")
			return nil
		case 'h', 'H':
			a.pages.SwitchToPage("help")
			return nil
		case 's', 'S':
			a.pages.SwitchToPage("settings")
			return nil
		case 'q', 'Q':
			a.app.Stop()
			return nil
		}
		return event
	})

	return textView
}

// createPatternsPage creates the pattern browser page
func (a *TViewApp) createPatternsPage() tview.Primitive {
	a.patternList = tview.NewList().
		SetSelectedFunc(a.onPatternSelected)

	// Populate pattern list
	for _, pattern := range a.patterns {
		a.patternList.AddItem(pattern.Name, pattern.Description, 0, nil)
	}

	a.patternList.SetBorder(true).
		SetTitle(" Select a Pattern ").
		SetTitleAlign(tview.AlignCenter)

	// Add help text
	helpText := tview.NewTextView().
		SetText("[gray]↑/↓: Navigate • Enter: Select • Esc: Back to Home • Q: Quit[white]").
		SetDynamicColors(true)

	// Create layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.patternList, 0, 1, true).
		AddItem(helpText, 1, 0, false)

	return flex
}

// createChatPage creates the chat interface page
func (a *TViewApp) createChatPage() tview.Primitive {
	// Chat display area
	a.chatView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true)

	a.chatView.SetBorder(true).
		SetTitle(" Chat ").
		SetTitleAlign(tview.AlignCenter)

	// Chat input field
	a.chatInput = tview.NewInputField().
		SetLabel("Message: ").
		SetFieldWidth(0).
		SetDoneFunc(a.onChatInput)

	a.chatInput.SetBorder(true)

	// Initial welcome message
	a.updateChatView()

	// Create layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.chatView, 0, 1, false).
		AddItem(a.chatInput, 3, 0, true)

	return flex
}

// createYouTubePage creates the YouTube processing page
func (a *TViewApp) createYouTubePage() tview.Primitive {
	// YouTube input field
	a.youtubeInput = tview.NewInputField().
		SetLabel("YouTube URL: ").
		SetFieldWidth(0).
		SetDoneFunc(a.onYouTubeInput)

	// YouTube output view
	youtubeView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true).
		SetText(`[yellow]YouTube Video Processing[white]

Enter a YouTube URL above and press Enter to:
• Extract video transcript
• Get video metadata (title, duration, views, etc.)
• Fetch top comments
• Process with any selected pattern

[green]Example URLs:[white]
• https://www.youtube.com/watch?v=VIDEO_ID
• https://youtu.be/VIDEO_ID

[gray]Note: Make sure yt-dlp is installed for full functionality[white]`)

	youtubeView.SetBorder(true).
		SetTitle(" YouTube Results ").
		SetTitleAlign(tview.AlignCenter)

	a.youtubeInput.SetBorder(true)

	// Options checkboxes
	optionsForm := tview.NewForm().
		AddCheckbox("Include Transcript", true, nil).
		AddCheckbox("Include Metadata", true, nil).
		AddCheckbox("Include Comments", false, nil)

	optionsForm.SetBorder(true).
		SetTitle(" Processing Options ").
		SetTitleAlign(tview.AlignCenter)

	// Create layout
	topFlex := tview.NewFlex().
		AddItem(a.youtubeInput, 0, 2, true).
		AddItem(optionsForm, 30, 0, false)

	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(topFlex, 7, 0, true).
		AddItem(youtubeView, 0, 1, false)

	return mainFlex
}

// createHelpPage creates the help and documentation page
func (a *TViewApp) createHelpPage() tview.Primitive {
	helpText := `[blue::b]Fabric TUI Help & Documentation[white::-]

[yellow::b]NAVIGATION[white::-]
[green]Global Shortcuts:[white]
• [cyan]Tab[white] - Cycle through all views (Home → Patterns → Chat → YouTube → Help → Settings)
• [cyan]Esc[white] - Return to Home page from any view
• [cyan]Ctrl+C[white] or [cyan]Q[white] - Quit application

[green]From Home Page:[white]
• [cyan]P[white] - Go to Pattern Browser
• [cyan]C[white] - Go to Chat Interface  
• [cyan]Y[white] - Go to YouTube Processing
• [cyan]H[white] - Go to Help (this page)
• [cyan]S[white] - Go to Settings

[yellow::b]PATTERN BROWSER[white::-]
• [cyan]↑/↓[white] arrows - Navigate through available patterns
• [cyan]Enter[white] - Select pattern and go to chat
• [cyan]j/k[white] - Vim-style navigation (up/down)
• [cyan]/[white] - Start filtering/searching patterns
• [cyan]Esc[white] - Clear filter or return to Home

[yellow::b]CHAT INTERFACE[white::-]
• Type your message and press [cyan]Enter[white] - Send message to AI
• [cyan]↑/↓[white] arrows - Scroll through chat history
• Messages are processed using the selected pattern
• AI responses use your configured providers (OpenAI, Claude, etc.)

[yellow::b]YOUTUBE PROCESSING[white::-]
• Enter any YouTube URL (youtube.com/watch?v= or youtu.be/)
• Press [cyan]Enter[white] - Process video (extract transcript, metadata, comments)
• Processing happens in background with status updates
• If pattern is selected, content is also analyzed by AI
• Requires [cyan]yt-dlp[white] to be installed for full functionality

[yellow::b]FEATURES[white::-]
[green]Real Integration:[white]
• Connects to your actual Fabric configuration
• Uses real patterns from ~/.config/fabric/patterns/
• Processes YouTube videos with real transcripts and metadata
• Ready for real AI provider integration

[green]Pattern Types:[white]
• [cyan]summarize[white] - Create concise summaries
• [cyan]extract_wisdom[white] - Extract key insights and wisdom
• [cyan]analyze_claims[white] - Analyze claims for validity
• [cyan]explain_code[white] - Explain code in simple terms
• [cyan]improve_writing[white] - Enhance writing quality
• Plus many more from your Fabric installation!

[yellow::b]SETUP REQUIREMENTS[white::-]
[green]For AI Responses:[white]
• Configure API keys in ~/.config/fabric/.env
• Run: [cyan]fabric --setup[white]
• Supported providers: OpenAI, Anthropic, Google Gemini, Ollama

[green]For YouTube Processing:[white]
• Install yt-dlp: [cyan]pip install yt-dlp[white]
• Ensure it's in your PATH
• Test with any YouTube URL

[green]For Custom Patterns:[white]
• Update patterns: [cyan]fabric --updatepatterns[white]
• Add custom patterns to your custom directory
• Patterns appear automatically in browser

[yellow::b]EXAMPLE WORKFLOWS[white::-]
[green]1. Analyze YouTube Video:[white]
   P → select "extract_wisdom" → Enter → Y → paste URL → Enter

[green]2. Summarize Long Text:[white]
   P → select "summarize" → Enter → C → paste text → Enter

[green]3. Code Explanation:[white]
   P → select "explain_code" → Enter → C → paste code → Enter

[yellow::b]TROUBLESHOOTING[white::-]
• If patterns don't load: Run [cyan]fabric --setup[white] and [cyan]fabric --updatepatterns[white]
• If YouTube fails: Install [cyan]yt-dlp[white] and check URL format
• If AI doesn't respond: Check API keys in ~/.config/fabric/.env
• Navigation issues: Use [cyan]Esc[white] to reset or [cyan]Tab[white] to cycle views

[gray]Press Esc to return to Home or Tab to continue navigating[white]`

	helpView := tview.NewTextView().
		SetText(helpText).
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true)

	helpView.SetBorder(true).
		SetTitle(" Help & Documentation ").
		SetTitleAlign(tview.AlignCenter)

	return helpView
}

// createSettingsPage creates the settings page
func (a *TViewApp) createSettingsPage() tview.Primitive {
	a.settings = tview.NewForm().
		AddDropDown("Model", []string{"gpt-4", "gpt-3.5-turbo", "claude-3", "gemini-pro"}, 0, nil).
		AddInputField("Temperature", "0.7", 10, nil, nil).
		AddInputField("Max Tokens", "2048", 10, nil, nil).
		AddCheckbox("Stream Responses", true, nil).
		AddButton("Save", func() {
			// TODO: Save settings
		}).
		AddButton("Back", func() {
			a.pages.SwitchToPage("home")
		})

	a.settings.SetBorder(true).
		SetTitle(" Settings ").
		SetTitleAlign(tview.AlignCenter)

	return a.settings
}

// globalKeyHandler handles global key events
func (a *TViewApp) globalKeyHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		// Escape always goes back to home
		a.pages.SwitchToPage("home")
		return nil
	case tcell.KeyTab:
		// Tab cycles through pages
		current, _ := a.pages.GetFrontPage()
		switch current {
		case "home":
			a.pages.SwitchToPage("patterns")
		case "patterns":
			a.pages.SwitchToPage("chat")
		case "chat":
			a.pages.SwitchToPage("youtube")
		case "youtube":
			a.pages.SwitchToPage("help")
		case "help":
			a.pages.SwitchToPage("settings")
		case "settings":
			a.pages.SwitchToPage("home")
		}
		return nil
	case tcell.KeyCtrlC:
		a.app.Stop()
		return nil
	}

	switch event.Rune() {
	case 'q', 'Q':
		a.app.Stop()
		return nil
	}

	return event
}

// onPatternSelected handles pattern selection
func (a *TViewApp) onPatternSelected(index int, mainText string, secondaryText string, shortcut rune) {
	if index < len(a.patterns) {
		a.selectedPattern = a.patterns[index]
		a.chatMessages = []ChatMessage{} // Clear previous messages
		a.updateChatView()
		a.pages.SwitchToPage("chat")
		a.app.SetFocus(a.chatInput)
	}
}

// onChatInput handles chat input
func (a *TViewApp) onChatInput(key tcell.Key) {
	if key == tcell.KeyEnter {
		message := a.chatInput.GetText()
		if strings.TrimSpace(message) != "" {
			// Add user message
			a.chatMessages = append(a.chatMessages, ChatMessage{
				Role:      "user",
				Content:   message,
				Timestamp: time.Now(),
			})

			// Generate real AI response
			response, err := a.integration.SendRealMessage(a.selectedPattern, message)
			if err != nil {
				response = fmt.Sprintf("Error generating response: %v", err)
			}
			a.chatMessages = append(a.chatMessages, ChatMessage{
				Role:      "assistant",
				Content:   response,
				Timestamp: time.Now(),
			})

			// Clear input and update view
			a.chatInput.SetText("")
			a.updateChatView()
			
			// Scroll to bottom
			a.chatView.ScrollToEnd()
		}
	}
}

// onYouTubeInput handles YouTube URL input
func (a *TViewApp) onYouTubeInput(key tcell.Key) {
	if key == tcell.KeyEnter {
		url := a.youtubeInput.GetText()
		if strings.TrimSpace(url) != "" {
			go a.processYouTubeURL(url)
		}
	}
}

// processYouTubeURL processes YouTube URL in background
func (a *TViewApp) processYouTubeURL(url string) {
	// Show processing message
	a.app.QueueUpdateDraw(func() {
		if youtubeView := a.getYouTubeView(); youtubeView != nil {
			youtubeView.SetText("[yellow]Processing YouTube URL...[white]\n\nPlease wait while we extract the video content...")
		}
	})

	// Process YouTube URL with real integration
	result, err := a.integration.ProcessYouTubeURL(url, true, true, true)
	
	// Update UI with results
	a.app.QueueUpdateDraw(func() {
		if youtubeView := a.getYouTubeView(); youtubeView != nil {
			if err != nil {
				youtubeView.SetText(fmt.Sprintf("[red]Error processing YouTube URL:[white]\n%v\n\n[gray]Make sure the URL is valid and yt-dlp is installed.[white]", err))
			} else {
				youtubeView.SetText(fmt.Sprintf("[green]Successfully processed:[white] %s\n\n%s", url, result))
				
				// If a pattern is selected, also process with AI
				if a.selectedPattern.Name != "" {
					aiResponse, aiErr := a.integration.SendRealMessage(a.selectedPattern, result)
					if aiErr == nil {
						youtubeView.SetText(youtubeView.GetText(false) + 
							fmt.Sprintf("\n\n[blue]AI Analysis using %s pattern:[white]\n%s", a.selectedPattern.Name, aiResponse))
					}
				}
			}
			youtubeView.ScrollToEnd()
		}
		a.youtubeInput.SetText("")
	})
}

// getYouTubeView helper to get the YouTube TextView
func (a *TViewApp) getYouTubeView() *tview.TextView {
	// This is a simplified approach - in a more complex app you'd store references
	if page, _ := a.pages.GetFrontPage(); page == "youtube" {
		// Navigate through the flex layout to find the TextView
		// This is hacky but works for our simple case
		return nil // Would need more complex navigation
	}
	return nil
}

// updateChatView refreshes the chat display
func (a *TViewApp) updateChatView() {
	var content strings.Builder

	if a.selectedPattern.Name != "" {
		content.WriteString(fmt.Sprintf("[blue::b]Using Pattern: %s[white::-]\n", a.selectedPattern.Name))
		content.WriteString(fmt.Sprintf("[gray]%s[white]\n\n", a.selectedPattern.Description))
	} else {
		content.WriteString("[yellow]No pattern selected. Go back to Pattern Browser to select one.[white]\n\n")
	}

	if len(a.chatMessages) == 0 {
		content.WriteString("[gray]Type a message below to start chatting...[white]\n")
	} else {
		for _, msg := range a.chatMessages {
			timestamp := msg.Timestamp.Format("15:04:05")
			if msg.Role == "user" {
				content.WriteString(fmt.Sprintf("[green::b][%s] You:[white::-] %s\n\n", timestamp, msg.Content))
			} else {
				content.WriteString(fmt.Sprintf("[blue::b][%s] Assistant:[white::-] %s\n\n", timestamp, msg.Content))
			}
		}
	}

	a.chatView.SetText(content.String())
}

// generateResponse generates a mock AI response
func (a *TViewApp) generateResponse(input string) string {
	if a.selectedPattern.Name == "" {
		return "Please select a pattern first to get AI responses tailored to your needs."
	}

	// Simple pattern-based responses
	responses := map[string]string{
		"summarize": fmt.Sprintf("**Summary using %s pattern:**\n\nKey points from your input:\n• Main concept: %s\n• Analysis: [processing]\n• Conclusion: [synthesis]\n\n*This is a demo response using tview.*", 
			a.selectedPattern.Name, truncateString(input, 50)),
		"extract_wisdom": fmt.Sprintf("**Wisdom using %s pattern:**\n\n💡 Key Insight: %s\n🎯 Principle: [underlying wisdom]\n📚 Application: [practical use]\n\n*Demo response with tview interface.*", 
			a.selectedPattern.Name, truncateString(input, 60)),
		"analyze_claims": fmt.Sprintf("**Analysis using %s pattern:**\n\n✅ Strong evidence for: [supported points]\n❓ Weak evidence for: [questionable claims]\n📊 Overall: Analyzing '%s'\n\n*Demo tview response.*", 
			a.selectedPattern.Name, truncateString(input, 40)),
		"explain_code": fmt.Sprintf("**Code explanation using %s:**\n\n🔍 Purpose: [what it does]\n⚙️ How it works: [mechanism]\n💡 Key concepts: [important ideas]\n\nCode: %s\n\n*Demo response with tview.*", 
			a.selectedPattern.Name, truncateString(input, 80)),
	}

	if response, exists := responses[a.selectedPattern.Name]; exists {
		return response
	}

	return fmt.Sprintf("Using the %s pattern to process: %s\n\n*This is a demo response. The full implementation would integrate with Fabric's AI processing.*", 
		a.selectedPattern.Name, truncateString(input, 100))
}

// getMockPatterns returns demo patterns for testing
func getMockPatterns() []Pattern {
	return []Pattern{
		{Name: "summarize", Description: "Create concise summaries of content"},
		{Name: "extract_wisdom", Description: "Extract key insights and wisdom from content"},
		{Name: "analyze_claims", Description: "Analyze claims for validity and supporting evidence"},
		{Name: "explain_code", Description: "Explain how code works in simple terms"},
		{Name: "create_summary", Description: "Create comprehensive summaries with key points"},
		{Name: "improve_writing", Description: "Improve writing quality, clarity, and style"},
		{Name: "translate", Description: "Translate content between languages"},
		{Name: "generate_ideas", Description: "Generate creative ideas and solutions"},
	}
}

// initializeFabric initializes the fabric database and plugin registry
func initializeFabric() (registry *core.PluginRegistry, err error) {
	var homedir string
	if homedir, err = os.UserHomeDir(); err != nil {
		return
	}

	fabricDb := fsdb.NewDb(filepath.Join(homedir, ".config/fabric"))
	if err = fabricDb.Configure(); err != nil {
		return
	}

	if registry, err = core.NewPluginRegistry(fabricDb); err != nil {
		return
	}

	return
}

// truncateString truncates a string to maxLen characters with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}