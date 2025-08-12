package pattern

import (
	"bytes"
	"context"
	"testing"

	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
)

// MockPatternHandler is a simple mock implementation for testing
type MockPatternHandler struct {
	name           string
	supportedTypes []string
}

func NewMockPatternHandler(name string, supportedTypes []string) *MockPatternHandler {
	return &MockPatternHandler{
		name:           name,
		supportedTypes: supportedTypes,
	}
}

func (m *MockPatternHandler) GetName() string {
	return m.name
}

func (m *MockPatternHandler) GetSetupDescription() string {
	return m.name + " handler"
}

func (m *MockPatternHandler) IsConfigured() bool {
	return true
}

func (m *MockPatternHandler) Configure() error {
	return nil
}

func (m *MockPatternHandler) Setup() error {
	return nil
}

func (m *MockPatternHandler) SetupFillEnvFileContent(buffer *bytes.Buffer) {
	// Mock implementation
}

func (m *MockPatternHandler) GetSupportedPatternTypes() []string {
	return m.supportedTypes
}

func (m *MockPatternHandler) SupportsStreaming() bool {
	for _, t := range m.supportedTypes {
		if t == "streaming" {
			return true
		}
	}
	return false
}

func (m *MockPatternHandler) SupportsFileOperations() bool {
	for _, t := range m.supportedTypes {
		if t == "file_operation" {
			return true
		}
	}
	return false
}

func (m *MockPatternHandler) ValidatePattern(pattern *fsdb.Pattern) error {
	return nil
}

func (m *MockPatternHandler) ExecutePattern(ctx context.Context, pattern *fsdb.Pattern, request *domain.ChatRequest, opts *domain.ChatOptions) (*PatternResult, error) {
	return &PatternResult{Content: "mock result"}, nil
}

func (m *MockPatternHandler) ProcessResult(result *PatternResult, opts *domain.ChatOptions) (string, error) {
	return result.Content, nil
}

func TestPatternRegistry_RegisterHandler(t *testing.T) {
	registry := NewPatternRegistry()
	handler := NewMockPatternHandler("test", []string{"standard"})

	// Test successful registration
	err := registry.RegisterHandler("test", handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test duplicate registration
	err = registry.RegisterHandler("test", handler)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test empty name
	err = registry.RegisterHandler("", handler)
	if err == nil {
		t.Error("Expected error for empty name")
	}

	// Test nil handler
	err = registry.RegisterHandler("nil_test", nil)
	if err == nil {
		t.Error("Expected error for nil handler")
	}
}

func TestPatternRegistry_GetHandler(t *testing.T) {
	registry := NewPatternRegistry()
	handler := NewMockPatternHandler("test", []string{"standard"})

	// Test getting non-existent handler
	_, err := registry.GetHandler("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent handler")
	}

	// Register and test getting existing handler
	registry.RegisterHandler("test", handler)
	retrieved, err := registry.GetHandler("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if retrieved.GetName() != handler.GetName() {
		t.Error("Retrieved handler does not match registered handler")
	}
}

func TestPatternRegistry_ListHandlers(t *testing.T) {
	registry := NewPatternRegistry()

	// Test empty registry
	handlers := registry.ListHandlers()
	if len(handlers) != 0 {
		t.Errorf("Expected 0 handlers, got %d", len(handlers))
	}

	// Add handlers and test
	handler1 := NewMockPatternHandler("handler1", []string{"standard"})
	handler2 := NewMockPatternHandler("handler2", []string{"streaming"})

	registry.RegisterHandler("handler1", handler1)
	registry.RegisterHandler("handler2", handler2)

	handlers = registry.ListHandlers()
	if len(handlers) != 2 {
		t.Errorf("Expected 2 handlers, got %d", len(handlers))
	}

	// Check that both handlers are in the list
	found1, found2 := false, false
	for _, name := range handlers {
		if name == "handler1" {
			found1 = true
		}
		if name == "handler2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Not all registered handlers found in list")
	}
}

func TestPatternRegistry_DefaultHandler(t *testing.T) {
	registry := NewPatternRegistry()
	handler := NewMockPatternHandler("default", []string{"standard"})

	// Test getting default handler when none set
	defaultHandler := registry.GetDefaultHandler()
	if defaultHandler != nil {
		t.Error("Expected nil default handler")
	}

	// Set and test default handler
	registry.SetDefaultHandler(handler)
	defaultHandler = registry.GetDefaultHandler()
	if defaultHandler.GetName() != handler.GetName() {
		t.Error("Default handler does not match set handler")
	}
}

func TestPatternRegistry_DetectPatternType(t *testing.T) {
	registry := NewPatternRegistry()

	tests := []struct {
		name        string
		pattern     *fsdb.Pattern
		expectedType string
	}{
		{
			name: "streaming pattern",
			pattern: &fsdb.Pattern{
				Name:    "stream_test",
				Pattern: "This pattern uses streaming output",
			},
			expectedType: "streaming",
		},
		{
			name: "file operation pattern",
			pattern: &fsdb.Pattern{
				Name:    "create_coding_feature",
				Pattern: "This pattern creates files",
			},
			expectedType: "file_operation",
		},
		{
			name: "standard pattern",
			pattern: &fsdb.Pattern{
				Name:    "summarize",
				Pattern: "Summarize the following text",
			},
			expectedType: "standard",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := registry.DetectPatternType(test.pattern)
			if result != test.expectedType {
				t.Errorf("Expected pattern type '%s', got '%s'", test.expectedType, result)
			}
		})
	}
}

func TestPatternRegistry_GetHandlerForPattern(t *testing.T) {
	registry := NewPatternRegistry()
	
	standardHandler := NewMockPatternHandler("standard", []string{"standard"})
	streamingHandler := NewMockPatternHandler("streaming", []string{"streaming"})
	defaultHandler := NewMockPatternHandler("default", []string{"standard", "streaming"})

	registry.RegisterHandler("standard", standardHandler)
	registry.RegisterHandler("streaming", streamingHandler)
	registry.SetDefaultHandler(defaultHandler)

	// Test getting handler for streaming pattern
	streamingPattern := &fsdb.Pattern{
		Name:    "stream_test",
		Pattern: "This pattern uses streaming output",
	}

	handler := registry.GetHandlerForPattern(streamingPattern)
	if handler.GetName() != streamingHandler.GetName() {
		t.Error("Expected streaming handler for streaming pattern")
	}

	// Test getting handler for standard pattern
	standardPattern := &fsdb.Pattern{
		Name:    "summarize",
		Pattern: "Summarize the following text",
	}

	handler = registry.GetHandlerForPattern(standardPattern)
	if handler.GetName() != standardHandler.GetName() {
		t.Error("Expected standard handler for standard pattern")
	}
}

func TestPatternRegistry_GetHandlerCount(t *testing.T) {
	registry := NewPatternRegistry()

	if registry.GetHandlerCount() != 0 {
		t.Errorf("Expected 0 handlers, got %d", registry.GetHandlerCount())
	}

	handler1 := NewMockPatternHandler("handler1", []string{"standard"})
	handler2 := NewMockPatternHandler("handler2", []string{"streaming"})

	registry.RegisterHandler("handler1", handler1)
	if registry.GetHandlerCount() != 1 {
		t.Errorf("Expected 1 handler, got %d", registry.GetHandlerCount())
	}

	registry.RegisterHandler("handler2", handler2)
	if registry.GetHandlerCount() != 2 {
		t.Errorf("Expected 2 handlers, got %d", registry.GetHandlerCount())
	}
}

func TestPatternRegistry_UnregisterHandler(t *testing.T) {
	registry := NewPatternRegistry()
	handler := NewMockPatternHandler("test", []string{"standard"})

	// Test unregistering non-existent handler
	err := registry.UnregisterHandler("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent handler")
	}

	// Register, then unregister handler
	registry.RegisterHandler("test", handler)
	if registry.GetHandlerCount() != 1 {
		t.Error("Handler not registered properly")
	}

	err = registry.UnregisterHandler("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if registry.GetHandlerCount() != 0 {
		t.Error("Handler not unregistered properly")
	}
}

func TestPatternRegistry_Clear(t *testing.T) {
	registry := NewPatternRegistry()
	handler1 := NewMockPatternHandler("handler1", []string{"standard"})
	handler2 := NewMockPatternHandler("handler2", []string{"streaming"})
	defaultHandler := NewMockPatternHandler("default", []string{"standard"})

	registry.RegisterHandler("handler1", handler1)
	registry.RegisterHandler("handler2", handler2)
	registry.SetDefaultHandler(defaultHandler)

	if registry.GetHandlerCount() != 2 {
		t.Error("Handlers not registered properly")
	}

	if registry.GetDefaultHandler() == nil {
		t.Error("Default handler not set properly")
	}

	registry.Clear()

	if registry.GetHandlerCount() != 0 {
		t.Error("Handlers not cleared properly")
	}

	if registry.GetDefaultHandler() != nil {
		t.Error("Default handler not cleared properly")
	}
}