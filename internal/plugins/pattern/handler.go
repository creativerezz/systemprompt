package pattern

import (
	"context"
	"fmt"
	"time"

	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/plugins"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
)

// PatternHandler defines the interface for all pattern handlers
type PatternHandler interface {
	plugins.Plugin

	// Pattern processing
	ValidatePattern(pattern *fsdb.Pattern) error
	ExecutePattern(ctx context.Context, pattern *fsdb.Pattern, request *domain.ChatRequest, opts *domain.ChatOptions) (*PatternResult, error)

	// Handler capabilities
	SupportsStreaming() bool
	SupportsFileOperations() bool
	GetSupportedPatternTypes() []string

	// Result processing
	ProcessResult(result *PatternResult, opts *domain.ChatOptions) (string, error)
}

// PatternResult represents the result of pattern execution
type PatternResult struct {
	Content        string
	Metadata       map[string]interface{}
	FileChanges    []domain.FileChange
	StreamChan     chan string
	Error          error
	ProcessingTime time.Duration
}

// PatternValidationError represents validation errors
type PatternValidationError struct {
	PatternName string
	Field       string
	Message     string
	Severity    ValidationSeverity
}

func (e *PatternValidationError) Error() string {
	return fmt.Sprintf("pattern validation error for '%s' field '%s': %s", e.PatternName, e.Field, e.Message)
}

// PatternExecutionError represents execution errors
type PatternExecutionError struct {
	PatternName string
	HandlerType string
	VendorName  string
	Cause       error
	Recoverable bool
}

func (e *PatternExecutionError) Error() string {
	return fmt.Sprintf("pattern execution error for '%s' using handler '%s' and vendor '%s': %s", 
		e.PatternName, e.HandlerType, e.VendorName, e.Cause.Error())
}

// ValidationSeverity represents the severity of validation issues
type ValidationSeverity int

const (
	ValidationWarning ValidationSeverity = iota
	ValidationError
	ValidationCritical
)

// String returns the string representation of ValidationSeverity
func (v ValidationSeverity) String() string {
	switch v {
	case ValidationWarning:
		return "warning"
	case ValidationError:
		return "error"
	case ValidationCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// PatternMetadata represents handler-specific metadata for patterns
type PatternMetadata struct {
	HandlerType     string            `yaml:"handler_type,omitempty"`
	RequiredVendors []string          `yaml:"required_vendors,omitempty"`
	Capabilities    []string          `yaml:"capabilities,omitempty"`
	ValidationRules map[string]string `yaml:"validation_rules,omitempty"`
	StreamingMode   string            `yaml:"streaming_mode,omitempty"`
}