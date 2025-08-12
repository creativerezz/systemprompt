// Package pattern provides a standardized interface for processing AI patterns
// within the Fabric framework. It includes pattern handlers, validation,
// execution, and result processing capabilities.
package pattern

// This package implements the pattern handler system as specified in the
// pattern-handlers feature specification. It provides:
//
// 1. PatternHandler interface - defines the contract for all pattern handlers
// 2. BasePatternHandler - provides common functionality and plugin integration
// 3. PatternRegistry - manages handler registration and discovery
// 4. Error types - structured error handling for validation and execution
// 5. Result types - standardized execution results
//
// The pattern handlers integrate seamlessly with the existing Fabric plugin
// architecture and maintain backward compatibility with all existing patterns.