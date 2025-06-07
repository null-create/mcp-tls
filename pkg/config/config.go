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
	// check tls configs
	tlsEnabled, err := strconv.ParseBool(os.Getenv("MCPTLS_ENABLED"))
	if err != nil {
		log.Fatal(err)
	}
	tlsKeyFile := os.Getenv("MCPTLS_KEY_FILE")
	if tlsKeyFile == "" {
		log.Print("WARNING MCPTLS_KEY_FILE env var not set")
	}
	tlsCertFile := os.Getenv("MCPTLS_CERT_FILE")
	if tlsCertFile == "" {
		log.Print("WARNING MCPTLS_CERT_FILE env var not set")
	}

	// check for custom server port
	serverPort := os.Getenv("MCPTLS_SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	return Config{
		TLSEnabled:  tlsEnabled,
		TLSKeyFile:  tlsKeyFile,
		TLSCertFile: tlsCertFile,
		ServerPort:  serverPort,
	}
}
