package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	TLSEnabled  bool
	TLSKeyFile  string
	TLSCertFile string
	ServerPort  string
}

// LoadConfig() loads the program configuration from environment variables.
func LoadConfig() Config {
	tlsEnabled, err := strconv.ParseBool(os.Getenv("MCPTLS_ENABLED"))
	if err != nil {
		log.Fatal(err)
	}
	serverPort := os.Getenv("MCPTLS_SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	return Config{
		TLSEnabled:  tlsEnabled,
		TLSKeyFile:  os.Getenv("MCPTLS_KEY_FILE"),
		TLSCertFile: os.Getenv("MCPTLS_CERT_FILE"),
		ServerPort:  serverPort,
	}
}
