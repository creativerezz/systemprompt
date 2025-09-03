package pattern

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/plugins/ai"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
)

func TestPatternValidationError_Error(t *testing.T) {
	err := &PatternValidationError{
		PatternName: "test_pattern",
		Field:       "content",
		Message:     "content is empty",
		Severity:    ValidationError,
	}

	expected := "pattern validation error for 'test_pattern' field 'content': content is empty"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestPatternExecutionError_Error(t *testing.T) {
	err := &PatternExecutionError{
		PatternName: "test_pattern",
		HandlerType: "standard",
		VendorName:  "openai",
		Cause:       &PatternValidationError{Message: "validation failed", Severity: ValidationError},
		Recoverable: true,
	}

	expected := "pattern execution error for 'test_pattern' using handler 'standard' and vendor 'openai': pattern validation error for '' field '': validation failed"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestValidationSeverity_String(t *testing.T) {
	tests := []struct {
		severity ValidationSeverity
		expected string
	}{
		{ValidationWarning, "warning"},
		{ValidationError, "error"},
		{ValidationCritical, "critical"},
		{ValidationSeverity(999), "unknown"},
	}

	for _, test := range tests {
		if result := test.severity.String(); result != test.expected {
			t.Errorf("Expected severity string '%s', got '%s'", test.expected, result)
		}
	}
}

func TestPatternResult_Structure(t *testing.T) {
	result := &PatternResult{
		Content:        "test content",
		Metadata:       map[string]interface{}{"key": "value"},
		FileChanges:    []domain.FileChange{{Path: "/test", Operation: "create"}},
		StreamChan:     make(chan string),
		ProcessingTime: time.Second,
	}

	if result.Content != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", result.Content)
	}

	if result.Metadata["key"] != "value" {
		t.Errorf("Expected metadata key 'value', got '%v'", result.Metadata["key"])
	}

	if len(result.FileChanges) != 1 || result.FileChanges[0].Path != "/test" {
		t.Errorf("Expected file change with path '/test', got %v", result.FileChanges)
	}

	if result.ProcessingTime != time.Second {
		t.Errorf("Expected processing time 1s, got %v", result.ProcessingTime)
	}
}

func TestFileChange_Structure(t *testing.T) {
	change := domain.FileChange{
		Path:      "/test/file.go",
		Operation: "update",
		Content:   "package main",
	}

	if change.Path != "/test/file.go" {
		t.Errorf("Expected path '/test/file.go', got '%s'", change.Path)
	}

	if change.Operation != "update" {
		t.Errorf("Expected operation 'update', got '%s'", change.Operation)
	}

	if change.Content != "package main" {
		t.Errorf("Expected content 'package main', got '%s'", change.Content)
	}
}

// BasePatternHandler Tests

func TestNewBasePatternHandler(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}
	
	if handler.GetName() != "TestHandler" {
		t.Errorf("Expected name 'TestHandler', got '%s'", handler.GetName())
	}
	
	if handler.GetSetupDescription() != "TestHandler Pattern Handler" {
		t.Errorf("Expected setup description 'TestHandler Pattern Handler', got '%s'", handler.GetSetupDescription())
	}
	
	// Check that configuration questions are set up
	if handler.EnableStreaming == nil {
		t.Error("Expected EnableStreaming to be configured")
	}
	
	if handler.EnableValidation == nil {
		t.Error("Expected EnableValidation to be configured")
	}
	
	if handler.MaxContextLength == nil {
		t.Error("Expected MaxContextLength to be configured")
	}
	
	// Check default values
	if handler.EnableStreaming.Value != "true" {
		t.Errorf("Expected EnableStreaming default to be 'true', got '%s'", handler.EnableStreaming.Value)
	}
	
	if handler.EnableValidation.Value != "true" {
		t.Errorf("Expected EnableValidation default to be 'true', got '%s'", handler.EnableValidation.Value)
	}
	
	if handler.MaxContextLength.Value != "0" {
		t.Errorf("Expected MaxContextLength default to be '0', got '%s'", handler.MaxContextLength.Value)
	}
}

func TestBasePatternHandler_IsConfigured(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	// Should be configured with default values
	if !handler.IsConfigured() {
		t.Error("Expected handler to be configured with default values")
	}
}

func TestBasePatternHandler_Configure(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	// Set environment variables
	os.Setenv("TESTHANDLER_PATTERN_HANDLER_ENABLE_STREAMING", "false")
	os.Setenv("TESTHANDLER_PATTERN_HANDLER_ENABLE_VALIDATION", "false")
	os.Setenv("TESTHANDLER_PATTERN_HANDLER_MAX_CONTEXT_LENGTH", "1000")
	
	defer func() {
		os.Unsetenv("TESTHANDLER_PATTERN_HANDLER_ENABLE_STREAMING")
		os.Unsetenv("TESTHANDLER_PATTERN_HANDLER_ENABLE_VALIDATION")
		os.Unsetenv("TESTHANDLER_PATTERN_HANDLER_MAX_CONTEXT_LENGTH")
	}()
	
	err := handler.Configure()
	if err != nil {
		t.Errorf("Expected no error during configuration, got %v", err)
	}
	
	if handler.EnableStreaming.Value != "false" {
		t.Errorf("Expected EnableStreaming to be 'false', got '%s'", handler.EnableStreaming.Value)
	}
	
	if handler.EnableValidation.Value != "false" {
		t.Errorf("Expected EnableValidation to be 'false', got '%s'", handler.EnableValidation.Value)
	}
	
	if handler.MaxContextLength.Value != "1000" {
		t.Errorf("Expected MaxContextLength to be '1000', got '%s'", handler.MaxContextLength.Value)
	}
}

func TestBasePatternHandler_SupportsStreaming(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	// Test default value (true)
	if !handler.SupportsStreaming() {
		t.Error("Expected SupportsStreaming to return true by default")
	}
	
	// Test false value
	handler.EnableStreaming.Value = "false"
	if handler.SupportsStreaming() {
		t.Error("Expected SupportsStreaming to return false when disabled")
	}
	
	// Test true value
	handler.EnableStreaming.Value = "true"
	if !handler.SupportsStreaming() {
		t.Error("Expected SupportsStreaming to return true when enabled")
	}
}

func TestBasePatternHandler_SupportsFileOperations(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	// Base handler should not support file operations
	if handler.SupportsFileOperations() {
		t.Error("Expected base handler to not support file operations")
	}
}

func TestBasePatternHandler_GetSupportedPatternTypes(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	types := handler.GetSupportedPatternTypes()
	if len(types) != 1 || types[0] != "standard" {
		t.Errorf("Expected supported types to be ['standard'], got %v", types)
	}
}

func TestBasePatternHandler_ValidatePattern(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	// Test with nil pattern
	err := handler.ValidatePattern(nil)
	if err == nil {
		t.Error("Expected error for nil pattern")
	}
	
	validationErr, ok := err.(*PatternValidationError)
	if !ok {
		t.Errorf("Expected PatternValidationError, got %T", err)
	} else {
		if validationErr.Severity != ValidationCritical {
			t.Errorf("Expected ValidationCritical, got %v", validationErr.Severity)
		}
	}
	
	// Test with empty name
	pattern := &fsdb.Pattern{
		Name:    "",
		Pattern: "test content",
	}
	
	err = handler.ValidatePattern(pattern)
	if err == nil {
		t.Error("Expected error for empty pattern name")
	}
	
	// Test with empty content
	pattern = &fsdb.Pattern{
		Name:    "test",
		Pattern: "",
	}
	
	err = handler.ValidatePattern(pattern)
	if err == nil {
		t.Error("Expected error for empty pattern content")
	}
	
	// Test with valid pattern
	pattern = &fsdb.Pattern{
		Name:    "test",
		Pattern: "test content",
	}
	
	err = handler.ValidatePattern(pattern)
	if err != nil {
		t.Errorf("Expected no error for valid pattern, got %v", err)
	}
	
	// Test with validation disabled
	handler.EnableValidation.Value = "false"
	err = handler.ValidatePattern(nil)
	if err != nil {
		t.Errorf("Expected no error when validation is disabled, got %v", err)
	}
	
	// Reset validation for context length test
	handler.EnableValidation.Value = "true"
	
	// Test context length validation
	longContent := make([]byte, 100001) // Exceed the 100000 character limit
	for i := range longContent {
		longContent[i] = 'a'
	}
	
	longPattern := &fsdb.Pattern{
		Name:    "long_test",
		Pattern: string(longContent),
	}
	
	handler.MaxContextLength.Value = "50000" // Set a limit
	err = handler.ValidatePattern(longPattern)
	if err == nil {
		t.Error("Expected error for pattern exceeding context length")
	}
	
	// Test with unlimited context length
	handler.MaxContextLength.Value = "0"
	err = handler.ValidatePattern(longPattern)
	if err != nil {
		t.Errorf("Expected no error with unlimited context length, got %v", err)
	}
}

func TestBasePatternHandler_ExecutePattern(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	pattern := &fsdb.Pattern{
		Name:    "test",
		Pattern: "test content",
	}
	
	ctx := context.Background()
	result, err := handler.ExecutePattern(ctx, pattern, nil, nil)
	
	// Base handler should return an error indicating it's not implemented
	if err == nil {
		t.Error("Expected error from base handler ExecutePattern")
	}
	
	if result != nil {
		t.Error("Expected nil result from base handler ExecutePattern")
	}
	
	execErr, ok := err.(*PatternExecutionError)
	if !ok {
		t.Errorf("Expected PatternExecutionError, got %T", err)
	} else {
		if execErr.PatternName != "test" {
			t.Errorf("Expected pattern name 'test', got '%s'", execErr.PatternName)
		}
		if execErr.HandlerType != "TestHandler" {
			t.Errorf("Expected handler type 'TestHandler', got '%s'", execErr.HandlerType)
		}
		if execErr.Recoverable {
			t.Error("Expected error to be non-recoverable")
		}
	}
}

func TestBasePatternHandler_ProcessResult(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	// Test with nil result
	output, err := handler.ProcessResult(nil, nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}
	if output != "" {
		t.Errorf("Expected empty output for nil result, got '%s'", output)
	}
	
	// Test with result containing error
	result := &PatternResult{
		Error: &PatternValidationError{Message: "test error", Severity: ValidationError},
	}
	
	output, err = handler.ProcessResult(result, nil)
	if err == nil {
		t.Error("Expected error when result contains error")
	}
	if output != "" {
		t.Errorf("Expected empty output when result contains error, got '%s'", output)
	}
	
	// Test with valid result
	result = &PatternResult{
		Content: "test output",
	}
	
	output, err = handler.ProcessResult(result, nil)
	if err != nil {
		t.Errorf("Expected no error for valid result, got %v", err)
	}
	if output != "test output" {
		t.Errorf("Expected output 'test output', got '%s'", output)
	}
}

func TestBasePatternHandler_GetVendorManager(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	if handler.GetVendorManager() != vendorManager {
		t.Error("Expected GetVendorManager to return the same instance")
	}
}

func TestBasePatternHandler_GetDatabase(t *testing.T) {
	vendorManager := &ai.VendorsManager{}
	db := &fsdb.Db{}
	
	handler := NewBasePatternHandler("TestHandler", vendorManager, db)
	
	if handler.GetDatabase() != db {
		t.Error("Expected GetDatabase to return the same instance")
	}
}