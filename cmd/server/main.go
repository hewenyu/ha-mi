package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/boringsoft/ha-mi/internal/api"
	"github.com/boringsoft/ha-mi/internal/config"
	"github.com/boringsoft/ha-mi/internal/db"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file (supports .yaml, .yml, .json)")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err)
		os.Exit(1)
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		fmt.Printf("Error creating database directory: %s\n", err)
		os.Exit(1)
	}

	// Initialize database
	database, err := db.New(cfg.Database.Path)
	if err != nil {
		fmt.Printf("Error connecting to database: %s\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize database schema
	if err := database.Initialize(); err != nil {
		fmt.Printf("Error initializing database: %s\n", err)
		os.Exit(1)
	}

	// Create and start server
	server := api.NewServer(cfg, database)
	if err := server.Start(); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("HA-MI server is running...")
	fmt.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal
	server.WaitForShutdown()
}
