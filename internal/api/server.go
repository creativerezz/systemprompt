package api

import (
	"log/slog"
	"net/http"

	"github.com/danielmiessler/fabric/internal/core"
	"github.com/gin-gonic/gin"
)

// APIServer represents the main API server
type APIServer struct {
	config   *Config
	registry *core.PluginRegistry
	router   *gin.Engine
}

// NewAPIServer creates a new API server instance
func NewAPIServer(config *Config, registry *core.PluginRegistry) *APIServer {
	// Set Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	server := &APIServer{
		config:   config,
		registry: registry,
		router:   gin.New(),
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// Start starts the API server
func (s *APIServer) Start() error {
	return s.router.Run(s.config.Address)
}

// setupMiddleware configures middleware for the server
func (s *APIServer) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logger middleware
	s.router.Use(gin.Logger())

	// CORS middleware
	s.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// API Key authentication middleware (if configured)
	if s.config.APIKey != "" {
		s.router.Use(s.apiKeyMiddleware())
	} else {
		slog.Warn("Starting API server without API key authentication. This may pose security risks.")
	}
}

// setupRoutes configures all API routes
func (s *APIServer) setupRoutes() {
	// Health check endpoint (no auth required)
	s.router.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.router.Group("/api")
	{
		// Pattern routes
		patterns := v1.Group("/patterns")
		{
			patterns.GET("", s.listPatterns)
			patterns.GET("/:name", s.getPattern)
		}

		// YouTube routes
		youtube := v1.Group("/youtube")
		{
			youtube.POST("/process", s.processYouTube)
		}
	}
}

// apiKeyMiddleware validates API key authentication
func (s *APIServer) apiKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.GetHeader("Authorization")
			if apiKey != "" && len(apiKey) > 7 && apiKey[:7] == "Bearer " {
				apiKey = apiKey[7:]
			}
		}

		if apiKey != s.config.APIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or missing API key",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// healthCheck handles the health check endpoint
func (s *APIServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "fabric-api",
		"version": "1.0.0",
	})
}