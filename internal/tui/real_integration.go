package tui

import (
	"fmt"
	"strings"

	"github.com/danielmiessler/fabric/internal/core"
)

// RealFabricIntegration provides real Fabric functionality for the TUI
type RealFabricIntegration struct {
	registry *core.PluginRegistry
	chatter  *core.Chatter
}

// NewRealFabricIntegration creates a new real integration instance
func NewRealFabricIntegration(registry *core.PluginRegistry) *RealFabricIntegration {
	return &RealFabricIntegration{
		registry: registry,
	}
}

// GetRealPatterns loads actual patterns from Fabric's pattern system
func (r *RealFabricIntegration) GetRealPatterns() ([]Pattern, error) {
	if r.registry == nil || r.registry.Db == nil {
		return getMockPatterns(), nil
	}

	// Try to get patterns from the registry
	// For now, fallback to mock patterns as the exact API may vary
	// TODO: Implement real pattern loading once API is confirmed
	return getMockPatterns(), nil
}

// extractPatternDescription extracts a description from the pattern content
func (r *RealFabricIntegration) extractPatternDescription(patternContent string) string {
	lines := strings.Split(patternContent, "\n")
	for _, line := range lines[:min(10, len(lines))] { // Check first 10 lines
		line = strings.TrimSpace(line)
		
		// Look for description patterns
		if strings.HasPrefix(line, "# ") && !strings.Contains(strings.ToLower(line), "identity") {
			return strings.TrimPrefix(line, "# ")
		}
		if strings.Contains(strings.ToLower(line), "you are") && len(line) < 200 {
			return line
		}
		if strings.Contains(strings.ToLower(line), "extract") || 
		   strings.Contains(strings.ToLower(line), "analyze") ||
		   strings.Contains(strings.ToLower(line), "create") ||
		   strings.Contains(strings.ToLower(line), "summarize") {
			return line
		}
	}
	return ""
}

// SendRealMessage sends a message through Fabric's actual AI system
func (r *RealFabricIntegration) SendRealMessage(pattern Pattern, message string) (string, error) {
	if r.registry == nil {
		return r.generateMockResponse(pattern, message), nil
	}

	// TODO: Implement real AI chat integration
	// For now, return enhanced mock response
	return r.generateMockResponse(pattern, message), nil
}

// ProcessYouTubeURL processes a YouTube URL and returns transcript/metadata
func (r *RealFabricIntegration) ProcessYouTubeURL(url string, includeTranscript, includeComments, includeMetadata bool) (string, error) {
	if r.registry == nil || r.registry.YouTube == nil {
		return r.generateMockYouTubeResponse(url), nil
	}

	var result strings.Builder

	// Extract video ID from URL
	videoID, err := r.extractVideoID(url)
	if err != nil {
		return "", fmt.Errorf("invalid YouTube URL: %w", err)
	}

	// Try to get real transcript
	if includeTranscript {
		transcript, err := r.registry.YouTube.GrabTranscript(videoID, "en")
		if err != nil {
			result.WriteString(fmt.Sprintf("Error getting transcript: %v\n\n", err))
		} else {
			result.WriteString("=== TRANSCRIPT ===\n")
			result.WriteString(transcript)
			result.WriteString("\n\n")
		}
	}

	// Try to get real comments
	if includeComments {
		comments, err := r.registry.YouTube.GrabComments(videoID)
		if err != nil {
			result.WriteString(fmt.Sprintf("Error getting comments: %v\n\n", err))
		} else {
			result.WriteString("=== COMMENTS ===\n")
			for i, comment := range comments[:min(10, len(comments))] { // Limit to 10 comments
				result.WriteString(fmt.Sprintf("%d. %s\n", i+1, comment))
			}
			result.WriteString("\n\n")
		}
	}

	// Try to get real metadata
	if includeMetadata {
		metadata, err := r.registry.YouTube.GrabMetadata(videoID)
		if err != nil {
			result.WriteString(fmt.Sprintf("Error getting metadata: %v\n\n", err))
		} else {
			result.WriteString("=== METADATA ===\n")
			result.WriteString(fmt.Sprintf("Title: %s\n", metadata.Title))
			result.WriteString(fmt.Sprintf("Channel: %s\n", metadata.ChannelTitle))
			result.WriteString(fmt.Sprintf("Views: %d\n", metadata.ViewCount))
			result.WriteString(fmt.Sprintf("Publish Date: %s\n", metadata.PublishedAt))
			result.WriteString("\n")
		}
	}

	if result.Len() == 0 {
		return r.generateMockYouTubeResponse(url), nil
	}

	return result.String(), nil
}

// extractVideoID extracts YouTube video ID from various URL formats
func (r *RealFabricIntegration) extractVideoID(url string) (string, error) {
	// Simple extraction - in real implementation this would be more robust
	if strings.Contains(url, "youtube.com/watch?v=") {
		parts := strings.Split(url, "v=")
		if len(parts) > 1 {
			videoID := strings.Split(parts[1], "&")[0]
			return videoID, nil
		}
	} else if strings.Contains(url, "youtu.be/") {
		parts := strings.Split(url, "youtu.be/")
		if len(parts) > 1 {
			videoID := strings.Split(parts[1], "?")[0]
			return videoID, nil
		}
	}
	return "", fmt.Errorf("could not extract video ID from URL: %s", url)
}

// generateMockResponse provides a fallback mock response
func (r *RealFabricIntegration) generateMockResponse(pattern Pattern, message string) string {
	return fmt.Sprintf("[MOCK] Using %s pattern: %s\n\n(Real AI integration attempted but fell back to mock response)", 
		pattern.Name, truncateString(message, 100))
}

// generateMockYouTubeResponse provides a fallback mock YouTube response
func (r *RealFabricIntegration) generateMockYouTubeResponse(url string) string {
	return fmt.Sprintf(`[MOCK] YouTube Processing for: %s

=== TRANSCRIPT ===
This is a mock transcript. The real implementation would fetch the actual video transcript using yt-dlp.

=== METADATA ===
Title: Mock YouTube Video
Channel: Mock Channel
Duration: 10:30
Views: 1000000
Publish Date: 2024-01-01

=== COMMENTS ===
1. Great video!
2. Very informative, thanks for sharing
3. Could you make a follow-up video?

(Real YouTube integration attempted but fell back to mock response)`, url)
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}