package main

import (
	"fmt"
	"log"
	"os"

	"github.com/danielmiessler/fabric/internal/api"
	"github.com/danielmiessler/fabric/internal/core"
)

func main() {
	// Load configuration from environment variables
	config := api.LoadConfig()

	// Initialize Fabric core
	registry := core.NewPluginRegistry()
	if err := registry.Setup(); err != nil {
		log.Fatalf("Failed to setup Fabric core: %v", err)
	}

	// Configure AI vendors
	registry.ConfigureVendors()

	// Create and start API server
	server := api.NewAPIServer(config, registry)
	
	fmt.Printf("Starting Fabric API server on %s\n", config.Address)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}