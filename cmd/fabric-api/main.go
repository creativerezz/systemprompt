package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/danielmiessler/fabric/internal/api"
	"github.com/danielmiessler/fabric/internal/core"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
)

func main() {
	// Load configuration from environment variables
	config := api.LoadConfig()

	// Initialize Fabric database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	configDir := filepath.Join(homeDir, ".config", "fabric")
	fabricDb := fsdb.NewDb(configDir)

	// Initialize Fabric core
	registry, err := core.NewPluginRegistry(fabricDb)
	if err != nil {
		log.Fatalf("Failed to create Fabric registry: %v", err)
	}
	if err := registry.Configure(); err != nil {
		log.Fatalf("Failed to configure Fabric core: %v", err)
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