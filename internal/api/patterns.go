package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// listPatterns handles GET /api/patterns
func (s *APIServer) listPatterns(c *gin.Context) {
	patternsDB := s.registry.PatternsLoader.Patterns
	if patternsDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Pattern database not available",
			"code":    "INTERNAL_ERROR",
		})
		return
	}

	names, err := patternsDB.GetNames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get pattern names",
			"code":    "INTERNAL_ERROR",
		})
		return
	}
	
	// Convert to API response format
	patternList := make([]map[string]interface{}, 0, len(names))
	for _, name := range names {
		patternList = append(patternList, map[string]interface{}{
			"name":        name,
			"description": "", // Pattern description would need to be extracted from pattern content
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    patternList,
		"count":   len(patternList),
	})
}

// getPattern handles GET /api/patterns/:name
func (s *APIServer) getPattern(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Pattern name is required",
			"code":    "BAD_REQUEST",
		})
		return
	}

	// Clean the pattern name
	name = strings.TrimSpace(name)
	name = filepath.Clean(name)

	patternsDB := s.registry.PatternsLoader.Patterns
	if patternsDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Pattern database not available",
			"code":    "INTERNAL_ERROR",
		})
		return
	}

	pattern, err := patternsDB.Get(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Pattern not found",
			"code":    "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"name":        name,
			"description": pattern.Description,
			"pattern":     pattern.Pattern,
		},
	})
}

// processYouTube handles POST /api/youtube/process
func (s *APIServer) processYouTube(c *gin.Context) {
	var request struct {
		URL     string `json:"url" binding:"required"`
		Pattern string `json:"pattern"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"code":    "BAD_REQUEST",
			"details": err.Error(),
		})
		return
	}

	// For now, return a placeholder response
	// This would need to be implemented based on your YouTube processing logic
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "YouTube processing not yet implemented",
		"data": map[string]interface{}{
			"url":     request.URL,
			"pattern": request.Pattern,
		},
	})
}