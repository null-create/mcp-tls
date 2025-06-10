package config

import (
	"log"
	"os"
)

type Config struct {
	ServerPort string // (OPTIONAL) server port. defaults to 8080
	Proxy      bool   // Whether the server functions as a proxy server (defaults to false)
}

// LoadConfigs() loads the global configurations from environment variables.
func LoadConfigs() Config {
	serverPort := os.Getenv("MCPTLS_SERVER_PORT")
	if serverPort == "" {
		log.Print("⚠️ WARNING MCPTLS_SERVER_PORT env var not set. Using defaults port 8080")
		serverPort = "8080"
	}

	return Config{
		ServerPort: serverPort,
	}
}
