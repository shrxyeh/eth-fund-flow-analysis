package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shrxyeh/ethereum-fund-flow/internal/api"
	"github.com/shrxyeh/ethereum-fund-flow/internal/config"
	"github.com/shrxyeh/ethereum-fund-flow/pkg/logger"
)

func main() {
	var (
		defaultAddr = "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH Contract as default
		showHelp    = flag.Bool("help", false, "Show usage information")
		address     = flag.String("address", defaultAddr, "Ethereum address to analyze")
		mode        = flag.String("mode", "both", "Analysis mode: beneficiary, payer, or both")
		port        = flag.String("port", "", "Port to run the server on (overrides .env PORT)")
	)
	
	// Parse flags
	flag.Parse()
	
	// Show help if requested
	if *showHelp {
		printUsage()
		os.Exit(0)
	}
	
	// Validate mode
	if *mode != "beneficiary" && *mode != "payer" && *mode != "both" {
		fmt.Println("Error: mode must be 'beneficiary', 'payer', or 'both'")
		printUsage()
		os.Exit(1)
	}
	
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Override port if specified
	if *port != "" {
		cfg.Port = *port
	}
	
	// Initialize logger
	l := logger.NewLogger()
	
	// Log startup information
	l.Infof("Starting Ethereum Fund Flow Analysis API")
	l.Infof("Default Ethereum address: %s", *address)
	l.Infof("Analysis mode: %s", *mode)
	
	// Start server
	server := api.NewServer(cfg, l)
	server.SetDefaultAddress(*address)
	server.SetAnalysisMode(*mode)
	
	l.Infof("Server starting on port %s", cfg.Port)
	if err := server.Start(); err != nil {
		l.Fatalf("Failed to start server: %v", err)
	}
}

func printUsage() {
	fmt.Println("Ethereum Fund Flow Analysis API")
	fmt.Println("\nUsage:")
	fmt.Println("  ./bin/api [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("  ./bin/api -address=0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2 -mode=beneficiary")
	fmt.Println("  ./bin/api -address=0x7a250d5630b4cf539739df2c5dacb4c659f2488d -mode=payer")
	fmt.Println("  ./bin/api -port=9090")
}