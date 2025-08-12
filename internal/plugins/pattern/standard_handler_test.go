package pattern

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/danielmiessler/fabric/internal/chat"
	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/plugins/ai"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
)

// MockVendor implements the ai.Vendor interface for testing
type MockVendor struct {
	name     string
	models   []string
	response string
	sendErr  error
}

func (m *MockVendor) GetName() string                     { return m.name }
func (m *MockVendor) GetSetupDescription() string        { return m.name }
func (m *MockVendor) IsConfigured() bool                 { return true }
func (m *MockVendor) Configure() error                   { return nil }
func (m *MockVendor) Setup() error                       { return nil }
func (m *MockVendor) SetupFillEnvFileContent(*bytes.Buffer) {}
func (m *MockVendor) ListModels() ([]string, error)      { return m.models, nil }
func (m *MockVendor) NeedsRawMode(string) bool           { return false }

func (m *MockVendor) Send(ctx context.Context, messages []*chat.ChatCompletionMessage, opts *domain.ChatOptions) (string, error) {
	if m.sendErr != nil {
		return "", m.sendErr
	}
	return m.response, nil
}

func (m *MockVendor) SendStream(messages []*chat.ChatCompletionMessage, opts *domain.ChatOptions, responseChan chan string) error {
	defer close(responseChan)
	
	if m.sendErr != nil {
		return m.sendErr
	}
	
	// Simulate streaming response
	words := []string{"This", "is", "a", "streaming", "response"}
	for _, word := range words {
		responseChan <- word + " "
		time.Sleep(10 * time.Millisecond)
	}
	
	return nil
}

// createTestPattern creates a test pattern directory structure
func createTestPattern(t *testing.T, dir, name, content string) {
	patternDir := filepath.Join(dir, name)
	if err := os.MkdirAll(patternDir, 0755); err != nil {
		t.Fatalf("Failed to create pattern directory: %v", err)
	}
	
	systemFile := filepath.Join(patternDir, "system.md")
	if err := os.WriteFile(systemFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create system.md file: %v", err)
	}
}

func TestStandardPatternHandler_ValidatePattern(t *testing.T) {
	// Create temporary directory for test patterns
	tempDir, err := os.MkdirTemp("", "fabric-test-patterns-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test database
	db := &fsdb.Db{
		Patterns: &fsdb.PatternsEntity{
			StorageEntity: &fsdb.StorageEntity{
				Dir: tempDir,
			},
			SystemPatternFile: "system.md",
		},
	}

	// Create mock vendor manager
	mockVendor := &MockVendor{
		name:     "test-vendor",
		models:   []string{"test-model"},
		response: "Test response",
	}
	vendorManager := ai.NewVendorsManager()
	vendorManager.AddVendors(mockVendor)

	// Create handler
	handler := NewStandardPatternHandler(vendorManager, db)
	handler.EnableValidation.Value = "true"

	tests := []struct {
		name        string
		pattern     *fsdb.Pattern
		setupFunc   func()
		expectError bool
		errorType   ValidationSeverity
	}{
		{
			name:        "nil pattern",
			pattern:     nil,
			expectError: true,
			errorType:   ValidationCritical,
		},
		{
			name: "empty pattern name",
			pattern: &fsdb.Pattern{
				Name:    "",
				Pattern: "test content",
			},
			expectError: true,
			errorType:   ValidationError,
		},
		{
			name: "empty pattern content",
			pattern: &fsdb.Pattern{
				Name:    "test-pattern",
				Pattern: "",
			},
			expectError: true,
			errorType:   ValidationError,
		},
		{
			name: "valid pattern with directory",
			pattern: &fsdb.Pattern{
				Name:    "test-pattern",
				Pattern: "# Test Pattern\n\nThis is a test pattern.",
			},
			setupFunc: func() {
				createTestPattern(t, tempDir, "test-pattern", "# Test Pattern\n\nThis is a test pattern.")
			},
			expectError: false,
		},
		{
			name: "pattern with unclosed template variable",
			pattern: &fsdb.Pattern{
				Name:    "test-pattern-template",
				Pattern: "Hello {{name, this is incomplete",
			},
			setupFunc: func() {
				createTestPattern(t, tempDir, "test-pattern-template", "Hello {{name, this is incomplete")
			},
			expectError: true,
			errorType:   ValidationWarning,
		},
		{
			name: "pattern with malformed markdown header",
			pattern: &fsdb.Pattern{
				Name:    "test-pattern-markdown",
				Pattern: "#Invalid header without space",
			},
			setupFunc: func() {
				createTestPattern(t, tempDir, "test-pattern-markdown", "#Invalid header without space")
			},
			expectError: true,
			errorType:   ValidationWarning,
		},
		{
			name: "file path pattern (should skip directory validation)",
			pattern: &fsdb.Pattern{
				Name:    "/tmp/test-pattern.md",
				Pattern: "# File Pattern\n\nThis is a file-based pattern.",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			err := handler.ValidatePattern(tt.pattern)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if validationErr, ok := err.(*PatternValidationError); ok {
					if validationErr.Severity != tt.errorType {
						t.Errorf("Expected error severity %v, got %v", tt.errorType, validationErr.Severity)
					}
				} else {
					t.Errorf("Expected PatternValidationError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestStandardPatternHandler_ExecutePattern(t *testing.T) {
	// Create mock vendor manager
	mockVendor := &MockVendor{
		name:     "test-vendor",
		models:   []string{"test-model"},
		response: "This is a test response from the AI vendor",
	}
	vendorManager := ai.NewVendorsManager()
	vendorManager.AddVendors(mockVendor)

	// Create test database
	db := &fsdb.Db{
		Patterns: &fsdb.PatternsEntity{
			SystemPatternFile: "system.md",
		},
	}

	// Create handler
	handler := NewStandardPatternHandler(vendorManager, db)
	handler.EnableValidation.Value = "false" // Disable validation for execution tests
	handler.EnableStreaming.Value = "false"  // Disable streaming for non-streaming tests

	tests := []struct {
		name        string
		pattern     *fsdb.Pattern
		request     *domain.ChatRequest
		opts        *domain.ChatOptions
		expectError bool
		checkResult func(*testing.T, *PatternResult)
	}{
		{
			name: "successful execution",
			pattern: &fsdb.Pattern{
				Name:    "test-pattern",
				Pattern: "You are a helpful assistant. Please respond to the user's message.",
			},
			request: &domain.ChatRequest{
				Message: &chat.ChatCompletionMessage{
					Role:    chat.ChatMessageRoleUser,
					Content: "Hello, how are you?",
				},
			},
			opts: &domain.ChatOptions{
				Model: "test-model",
			},
			expectError: false,
			checkResult: func(t *testing.T, result *PatternResult) {
				if result.Content != "This is a test response from the AI vendor" {
					t.Errorf("Expected specific response, got: %s", result.Content)
				}
				if result.ProcessingTime == 0 {
					t.Error("Expected processing time to be set")
				}
				if result.Metadata["vendor"] != "test-vendor" {
					t.Errorf("Expected vendor metadata to be 'test-vendor', got: %v", result.Metadata["vendor"])
				}
			},
		},
		{
			name: "execution with vendor error",
			pattern: &fsdb.Pattern{
				Name:    "test-pattern",
				Pattern: "Test pattern content",
			},
			request: &domain.ChatRequest{
				Message: &chat.ChatCompletionMessage{
					Role:    chat.ChatMessageRoleUser,
					Content: "Test message",
				},
			},
			opts: &domain.ChatOptions{
				Model: "test-model",
			},
			expectError: true,
		},
		{
			name: "execution with no suitable vendor",
			pattern: &fsdb.Pattern{
				Name:    "test-pattern",
				Pattern: "Test pattern content",
			},
			request: &domain.ChatRequest{
				Message: &chat.ChatCompletionMessage{
					Role:    chat.ChatMessageRoleUser,
					Content: "Test message",
				},
			},
			opts: &domain.ChatOptions{
				Model: "non-existent-model",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up vendor error for specific test
			if tt.name == "execution with vendor error" {
				mockVendor.sendErr = fmt.Errorf("vendor error")
			} else {
				mockVendor.sendErr = nil
			}

			ctx := context.Background()
			result, err := handler.ExecutePattern(ctx, tt.pattern, tt.request, tt.opts)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if executionErr, ok := err.(*PatternExecutionError); ok {
					if executionErr.PatternName != tt.pattern.Name {
						t.Errorf("Expected pattern name %s in error, got %s", tt.pattern.Name, executionErr.PatternName)
					}
				} else {
					t.Errorf("Expected PatternExecutionError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}

				if result == nil {
					t.Error("Expected result but got nil")
					return
				}

				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

func TestStandardPatternHandler_ExecutePatternWithStreaming(t *testing.T) {
	// Create mock vendor manager
	mockVendor := &MockVendor{
		name:     "test-vendor",
		models:   []string{"test-model"},
		response: "This is a streaming response",
	}
	vendorManager := ai.NewVendorsManager()
	vendorManager.AddVendors(mockVendor)

	// Create test database
	db := &fsdb.Db{
		Patterns: &fsdb.PatternsEntity{
			SystemPatternFile: "system.md",
		},
	}

	// Create handler with streaming enabled
	handler := NewStandardPatternHandler(vendorManager, db)
	handler.EnableValidation.Value = "false"
	handler.EnableStreaming.Value = "true"

	pattern := &fsdb.Pattern{
		Name:    "test-pattern",
		Pattern: "You are a helpful assistant.",
	}

	request := &domain.ChatRequest{
		Message: &chat.ChatCompletionMessage{
			Role:    chat.ChatMessageRoleUser,
			Content: "Hello",
		},
	}

	opts := &domain.ChatOptions{
		Model: "test-model",
	}

	ctx := context.Background()
	result, err := handler.ExecutePattern(ctx, pattern, request, opts)

	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	// Check that streaming channel was created
	if result.StreamChan == nil {
		t.Error("Expected streaming channel to be set")
	}

	// Check that content was collected from streaming
	if !strings.Contains(result.Content, "streaming") {
		t.Errorf("Expected streaming content, got: %s", result.Content)
	}
}

func TestStandardPatternHandler_GetSupportedPatternTypes(t *testing.T) {
	handler := NewStandardPatternHandler(nil, nil)
	
	types := handler.GetSupportedPatternTypes()
	
	expectedTypes := []string{"standard", "default"}
	if len(types) != len(expectedTypes) {
		t.Errorf("Expected %d pattern types, got %d", len(expectedTypes), len(types))
	}
	
	for i, expectedType := range expectedTypes {
		if i >= len(types) || types[i] != expectedType {
			t.Errorf("Expected pattern type %s at index %d, got %s", expectedType, i, types[i])
		}
	}
}