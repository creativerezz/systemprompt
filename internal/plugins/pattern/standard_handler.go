package pattern

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielmiessler/fabric/internal/chat"
	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/plugins"
	"github.com/danielmiessler/fabric/internal/plugins/ai"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
)

// StandardPatternHandler handles existing patterns with system.md/user.md structure
type StandardPatternHandler struct {
	*BasePatternHandler
}

// NewStandardPatternHandler creates a new standard pattern handler
func NewStandardPatternHandler(vendorManager *ai.VendorsManager, db *fsdb.Db) *StandardPatternHandler {
	base := NewBasePatternHandler("Standard Pattern Handler", vendorManager, db)
	
	return &StandardPatternHandler{
		BasePatternHandler: base,
	}
}

// ValidatePattern validates the pattern directory structure and content
func (h *StandardPatternHandler) ValidatePattern(pattern *fsdb.Pattern) error {
	// First run base validation
	if err := h.BasePatternHandler.ValidatePattern(pattern); err != nil {
		return err
	}

	if !plugins.ParseBoolElseFalse(h.EnableValidation.Value) {
		return nil // Validation disabled
	}

	// Validate pattern structure for file-based patterns
	if !strings.HasPrefix(pattern.Name, "/") && !strings.HasPrefix(pattern.Name, "\\") && 
	   !strings.HasPrefix(pattern.Name, "~") && !strings.HasPrefix(pattern.Name, ".") {
		
		// This is a pattern name, validate directory structure
		if err := h.validatePatternDirectory(pattern.Name); err != nil {
			return err
		}
	}

	// Validate pattern content
	if err := h.validatePatternContent(pattern); err != nil {
		return err
	}

	return nil
}

// validatePatternDirectory validates the pattern directory structure
func (h *StandardPatternHandler) validatePatternDirectory(patternName string) error {
	// Check main patterns directory
	mainPatternDir := filepath.Join(h.db.Patterns.Dir, patternName)
	systemFile := filepath.Join(mainPatternDir, h.db.Patterns.SystemPatternFile)
	
	// Check custom patterns directory if configured
	var customSystemFile string
	if h.db.Patterns.CustomPatternsDir != "" {
		customPatternDir := filepath.Join(h.db.Patterns.CustomPatternsDir, patternName)
		customSystemFile = filepath.Join(customPatternDir, h.db.Patterns.SystemPatternFile)
	}

	// Pattern must exist in either main or custom directory
	mainExists := false
	customExists := false

	if _, err := os.Stat(systemFile); err == nil {
		mainExists = true
	}

	if customSystemFile != "" {
		if _, err := os.Stat(customSystemFile); err == nil {
			customExists = true
		}
	}

	if !mainExists && !customExists {
		return &PatternValidationError{
			PatternName: patternName,
			Field:       "directory",
			Message:     fmt.Sprintf("pattern directory not found: %s", patternName),
			Severity:    ValidationError,
		}
	}

	// Validate system.md file exists and is readable
	systemFileToCheck := systemFile
	if customExists {
		systemFileToCheck = customSystemFile
	}

	if _, err := os.Stat(systemFileToCheck); err != nil {
		return &PatternValidationError{
			PatternName: patternName,
			Field:       "system_file",
			Message:     fmt.Sprintf("system.md file not found or not readable: %s", err.Error()),
			Severity:    ValidationError,
		}
	}

	return nil
}

// validatePatternContent validates the pattern content
func (h *StandardPatternHandler) validatePatternContent(pattern *fsdb.Pattern) error {
	content := strings.TrimSpace(pattern.Pattern)

	// Check for empty content
	if content == "" {
		return &PatternValidationError{
			PatternName: pattern.Name,
			Field:       "content",
			Message:     "pattern content is empty",
			Severity:    ValidationError,
		}
	}

	// Check for common pattern issues
	if strings.Contains(content, "{{") && !strings.Contains(content, "}}") {
		return &PatternValidationError{
			PatternName: pattern.Name,
			Field:       "content",
			Message:     "unclosed template variable found",
			Severity:    ValidationWarning,
		}
	}

	// Check for malformed markdown (basic check)
	if strings.Count(content, "#") > 0 {
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "# ") && 
			   !strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") &&
			   !strings.HasPrefix(line, "#### ") && !strings.HasPrefix(line, "##### ") &&
			   !strings.HasPrefix(line, "###### ") {
				return &PatternValidationError{
					PatternName: pattern.Name,
					Field:       "content",
					Message:     fmt.Sprintf("malformed markdown header at line %d: %s", i+1, line),
					Severity:    ValidationWarning,
				}
			}
		}
	}

	return nil
}

// ExecutePattern executes the pattern using the existing AI vendor system
func (h *StandardPatternHandler) ExecutePattern(ctx context.Context, pattern *fsdb.Pattern, request *domain.ChatRequest, opts *domain.ChatOptions) (*PatternResult, error) {
	startTime := time.Now()

	result := &PatternResult{
		Metadata: make(map[string]interface{}),
	}

	// Validate pattern first
	if err := h.ValidatePattern(pattern); err != nil {
		result.Error = &PatternExecutionError{
			PatternName: pattern.Name,
			HandlerType: h.GetName(),
			Cause:       err,
			Recoverable: false,
		}
		return result, result.Error
	}

	// Get a vendor for execution
	vendor := h.getVendorForExecution(opts.Model)
	if vendor == nil {
		result.Error = &PatternExecutionError{
			PatternName: pattern.Name,
			HandlerType: h.GetName(),
			Cause:       fmt.Errorf("no suitable AI vendor found for model: %s", opts.Model),
			Recoverable: true,
		}
		return result, result.Error
	}

	// Create a session for pattern execution
	session := &fsdb.Session{}
	
	// Add system message with pattern content
	if pattern.Pattern != "" {
		session.Append(&chat.ChatCompletionMessage{
			Role:    chat.ChatMessageRoleSystem,
			Content: pattern.Pattern,
		})
	}

	// Add user message if provided
	if request.Message != nil && request.Message.Content != "" {
		session.Append(request.Message)
	}

	messages := session.GetVendorMessages()
	if len(messages) == 0 {
		result.Error = &PatternExecutionError{
			PatternName: pattern.Name,
			HandlerType: h.GetName(),
			Cause:       fmt.Errorf("no messages to send to AI vendor"),
			Recoverable: false,
		}
		return result, result.Error
	}

	// Execute with streaming or non-streaming based on configuration
	var content string
	var err error

	if h.SupportsStreaming() && opts != nil {
		content, err = h.executeWithStreaming(ctx, vendor, messages, opts, result)
	} else {
		content, err = vendor.Send(ctx, messages, opts)
	}

	if err != nil {
		result.Error = &PatternExecutionError{
			PatternName: pattern.Name,
			HandlerType: h.GetName(),
			VendorName:  vendor.GetName(),
			Cause:       err,
			Recoverable: true,
		}
		return result, result.Error
	}

	result.Content = content
	result.ProcessingTime = time.Since(startTime)
	result.Metadata["vendor"] = vendor.GetName()
	result.Metadata["model"] = opts.Model
	result.Metadata["pattern_name"] = pattern.Name

	return result, nil
}

// executeWithStreaming executes the pattern with streaming support
func (h *StandardPatternHandler) executeWithStreaming(ctx context.Context, vendor ai.Vendor, messages []*chat.ChatCompletionMessage, opts *domain.ChatOptions, result *PatternResult) (string, error) {
	responseChan := make(chan string, 100)
	result.StreamChan = responseChan

	errChan := make(chan error, 1)
	done := make(chan struct{})
	var content strings.Builder

	go func() {
		defer close(done)
		
		if streamErr := vendor.SendStream(messages, opts, responseChan); streamErr != nil {
			errChan <- streamErr
		}
	}()

	// Collect streaming responses
	for {
		select {
		case response, ok := <-responseChan:
			if !ok {
				// Channel closed, streaming finished
				goto streamingDone
			}
			content.WriteString(response)
		case err := <-errChan:
			return "", err
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

streamingDone:
	// Wait for goroutine to finish
	<-done

	// Check for any final errors
	select {
	case err := <-errChan:
		return "", err
	default:
		// No errors
	}

	return content.String(), nil
}

// getVendorForExecution gets the appropriate vendor for pattern execution
func (h *StandardPatternHandler) getVendorForExecution(model string) ai.Vendor {
	if h.vendorManager == nil {
		return nil
	}

	if model == "" {
		// Return first available vendor
		if len(h.vendorManager.Vendors) > 0 {
			return h.vendorManager.Vendors[0]
		}
		return nil
	}

	// Find vendor by model
	models, err := h.vendorManager.GetModels()
	if err != nil {
		return nil
	}

	vendorName := models.FindGroupsByItemFirst(model)
	return h.vendorManager.FindByName(vendorName)
}

// GetSupportedPatternTypes returns the pattern types this handler supports
func (h *StandardPatternHandler) GetSupportedPatternTypes() []string {
	return []string{"standard", "default"}
}