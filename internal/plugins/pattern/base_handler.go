package pattern

import (
	"context"
	"fmt"
	"strconv"

	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/plugins"
	"github.com/danielmiessler/fabric/internal/plugins/ai"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
)

// BasePatternHandler provides common functionality for all pattern handlers
type BasePatternHandler struct {
	*plugins.PluginBase

	// Configuration
	EnableStreaming  *plugins.SetupQuestion
	EnableValidation *plugins.SetupQuestion
	MaxContextLength *plugins.SetupQuestion

	// Runtime state
	vendorManager *ai.VendorsManager
	db            *fsdb.Db
}

// NewBasePatternHandler creates a new base pattern handler
func NewBasePatternHandler(name string, vendorManager *ai.VendorsManager, db *fsdb.Db) *BasePatternHandler {
	handler := &BasePatternHandler{
		vendorManager: vendorManager,
		db:            db,
	}

	handler.PluginBase = &plugins.PluginBase{
		Name:             name,
		SetupDescription: name + " Pattern Handler",
		EnvNamePrefix:    plugins.BuildEnvVariablePrefix(name + " Pattern Handler"),
	}

	// Setup configuration questions
	handler.EnableStreaming = handler.AddSetupQuestionBool("Enable Streaming", false)
	handler.EnableStreaming.Value = "true"

	handler.EnableValidation = handler.AddSetupQuestionBool("Enable Validation", false)
	handler.EnableValidation.Value = "true"

	handler.MaxContextLength = handler.AddSetupQuestionCustom("Max Context Length", false,
		"Enter the maximum context length for patterns (0 for unlimited)")
	handler.MaxContextLength.Value = "0"

	return handler
}

// SupportsStreaming returns whether this handler supports streaming
func (h *BasePatternHandler) SupportsStreaming() bool {
	return plugins.ParseBoolElseFalse(h.EnableStreaming.Value)
}

// SupportsFileOperations returns whether this handler supports file operations
func (h *BasePatternHandler) SupportsFileOperations() bool {
	return false // Base handler doesn't support file operations
}

// GetSupportedPatternTypes returns the pattern types this handler supports
func (h *BasePatternHandler) GetSupportedPatternTypes() []string {
	return []string{"standard"} // Base handler supports standard patterns
}

// ValidatePattern provides basic pattern validation
func (h *BasePatternHandler) ValidatePattern(pattern *fsdb.Pattern) error {
	if !plugins.ParseBoolElseFalse(h.EnableValidation.Value) {
		return nil // Validation disabled
	}

	if pattern == nil {
		return &PatternValidationError{
			PatternName: "unknown",
			Field:       "pattern",
			Message:     "pattern cannot be nil",
			Severity:    ValidationCritical,
		}
	}

	if pattern.Name == "" {
		return &PatternValidationError{
			PatternName: pattern.Name,
			Field:       "name",
			Message:     "pattern name cannot be empty",
			Severity:    ValidationError,
		}
	}

	if pattern.Pattern == "" {
		return &PatternValidationError{
			PatternName: pattern.Name,
			Field:       "pattern",
			Message:     "pattern content cannot be empty",
			Severity:    ValidationError,
		}
	}

	// Check context length if configured
	if h.MaxContextLength.Value != "0" {
		if maxLength, err := strconv.Atoi(h.MaxContextLength.Value); err == nil && maxLength > 0 {
			if len(pattern.Pattern) > maxLength {
				return &PatternValidationError{
					PatternName: pattern.Name,
					Field:       "pattern",
					Message:     fmt.Sprintf("pattern content exceeds maximum length of %d characters", maxLength),
					Severity:    ValidationError,
				}
			}
		}
	}

	return nil
}

// ExecutePattern provides base pattern execution (to be overridden by specific handlers)
func (h *BasePatternHandler) ExecutePattern(ctx context.Context, pattern *fsdb.Pattern, request *domain.ChatRequest, opts *domain.ChatOptions) (*PatternResult, error) {
	return nil, &PatternExecutionError{
		PatternName: pattern.Name,
		HandlerType: h.GetName(),
		Cause:       fmt.Errorf("ExecutePattern not implemented in base handler"),
		Recoverable: false,
	}
}

// ProcessResult processes the pattern execution result
func (h *BasePatternHandler) ProcessResult(result *PatternResult, opts *domain.ChatOptions) (string, error) {
	if result == nil {
		return "", fmt.Errorf("result cannot be nil")
	}

	if result.Error != nil {
		return "", result.Error
	}

	return result.Content, nil
}

// GetVendorManager returns the vendor manager
func (h *BasePatternHandler) GetVendorManager() *ai.VendorsManager {
	return h.vendorManager
}

// GetDatabase returns the database
func (h *BasePatternHandler) GetDatabase() *fsdb.Db {
	return h.db
}