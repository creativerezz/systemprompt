package pattern

import (
	"fmt"
	"strings"
	"sync"

	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
)

// PatternRegistry manages pattern handlers
type PatternRegistry struct {
	handlers       map[string]PatternHandler
	defaultHandler PatternHandler
	typeDetectors  []PatternTypeDetector
	mu             sync.RWMutex
}

// PatternTypeDetector detects pattern types for automatic handler selection
type PatternTypeDetector func(pattern *fsdb.Pattern) string

// NewPatternRegistry creates a new pattern registry
func NewPatternRegistry() *PatternRegistry {
	return &PatternRegistry{
		handlers:      make(map[string]PatternHandler),
		typeDetectors: make([]PatternTypeDetector, 0),
	}
}

// RegisterHandler registers a pattern handler
func (r *PatternRegistry) RegisterHandler(name string, handler PatternHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("handler name cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	if _, exists := r.handlers[name]; exists {
		return fmt.Errorf("handler '%s' already registered", name)
	}

	r.handlers[name] = handler
	return nil
}

// GetHandler retrieves a handler by name
func (r *PatternRegistry) GetHandler(name string) (PatternHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[name]
	if !exists {
		return nil, fmt.Errorf("handler '%s' not found", name)
	}

	return handler, nil
}

// GetHandlerForPattern automatically selects the best handler for a pattern
func (r *PatternRegistry) GetHandlerForPattern(pattern *fsdb.Pattern) PatternHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Detect pattern type
	patternType := r.detectPatternTypeInternal(pattern)
	
	// Try to find a handler that supports this pattern type
	for _, handler := range r.handlers {
		supportedTypes := handler.GetSupportedPatternTypes()
		for _, supportedType := range supportedTypes {
			if supportedType == patternType {
				return handler
			}
		}
	}

	// Fall back to default handler
	if r.defaultHandler != nil {
		return r.defaultHandler
	}

	// Return the first available handler as last resort
	for _, handler := range r.handlers {
		return handler
	}

	return nil
}

// detectPatternTypeInternal is the internal pattern type detection (without locking)
func (r *PatternRegistry) detectPatternTypeInternal(pattern *fsdb.Pattern) string {
	// Try pattern type detection first
	for _, detector := range r.typeDetectors {
		patternType := detector(pattern)
		if patternType != "" {
			return patternType
		}
	}

	// Use default detector
	return DefaultPatternTypeDetector(pattern)
}

// ListHandlers returns all registered handler names
func (r *PatternRegistry) ListHandlers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		names = append(names, name)
	}

	return names
}

// SetDefaultHandler sets the default handler
func (r *PatternRegistry) SetDefaultHandler(handler PatternHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.defaultHandler = handler
}

// AddTypeDetector adds a pattern type detector
func (r *PatternRegistry) AddTypeDetector(detector PatternTypeDetector) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.typeDetectors = append(r.typeDetectors, detector)
}

// GetDefaultHandler returns the default handler
func (r *PatternRegistry) GetDefaultHandler() PatternHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultHandler
}

// GetHandlerCount returns the number of registered handlers
func (r *PatternRegistry) GetHandlerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.handlers)
}

// UnregisterHandler removes a handler from the registry
func (r *PatternRegistry) UnregisterHandler(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[name]; !exists {
		return fmt.Errorf("handler '%s' not found", name)
	}

	delete(r.handlers, name)
	return nil
}

// Clear removes all handlers and resets the default handler
func (r *PatternRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers = make(map[string]PatternHandler)
	r.defaultHandler = nil
	r.typeDetectors = make([]PatternTypeDetector, 0)
}

// DetectPatternType detects the pattern type using registered detectors
func (r *PatternRegistry) DetectPatternType(pattern *fsdb.Pattern) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.detectPatternTypeInternal(pattern)
}

// DefaultPatternTypeDetector is a basic pattern type detector
func DefaultPatternTypeDetector(pattern *fsdb.Pattern) string {
	if pattern == nil {
		return ""
	}

	content := strings.ToLower(pattern.Pattern)

	// Detect streaming patterns
	if strings.Contains(content, "stream") || strings.Contains(content, "real-time") {
		return "streaming"
	}

	// Detect file operation patterns
	if strings.Contains(content, "create_coding_feature") || 
	   strings.Contains(content, "file") && (strings.Contains(content, "create") || strings.Contains(content, "modify")) {
		return "file_operation"
	}

	// Default to standard pattern
	return "standard"
}