package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	EtherscanAPIKey string
	Port            string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	etherscanAPIKey := os.Getenv("ETHERSCAN_API_KEY")
	if etherscanAPIKey == "" {
		return nil, fmt.Errorf("ETHERSCAN_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	return &Config{
		EtherscanAPIKey: etherscanAPIKey,
		Port:            port,
	}, nil
}