package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/null-create/mcp-tls/pkg/tls"
)

type Config struct {
	ServerPort string // (OPTIONAL) server port. defaults to 8080
	TargetURL  string // target for proxy server to pass/receive from
	TLSConfig  tls.TLSConfig
}

// LoadConfigs() loads the program configuration from environment variables.
func LoadConfigs() Config {
	// proxy target url
	targetURL := os.Getenv("MCPTLS_TARGET_URL")
	if targetURL == "" {
		log.Fatal("❌ MCPTLS_TARGET_URL must be set")
	}

	// check tls configs
	tlsEnabled, err := strconv.ParseBool(os.Getenv("MCPTLS_ENABLED"))
	if err != nil {
		log.Fatal(err)
	}
	tlsKeyFile := os.Getenv("MCPTLS_KEY_FILE")
	if tlsKeyFile == "" {
		log.Print("⚠️ WARNING MCPTLS_KEY_FILE env var not set. Using defaults.")
		tlsKeyFile = filepath.Join("certs", "server.key")
	}
	tlsCertFile := os.Getenv("MCPTLS_CERT_FILE")
	if tlsCertFile == "" {
		log.Print("⚠️ WARNING MCPTLS_CERT_FILE env var not set. Using defaults.")
		tlsCertFile = filepath.Join("certs", "server.crt")
	}
	tlsClientCAFile := os.Getenv("MCPTLS_CLIENT_CA_FILE")
	if tlsClientCAFile == "" {
		log.Print("⚠️ WARNING MCPTLS_CLIENT_CA_FILE env var not set. Using defaults.")
		tlsClientCAFile = filepath.Join("certs", "ca.crt")
	}

	// check for custom server port
	serverPort := os.Getenv("MCPTLS_SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	return Config{
		TLSConfig: tls.TLSConfig{
			TLSEnabled:      tlsEnabled,
			TLSKeyFile:      tlsKeyFile,
			TLSCertFile:     tlsCertFile,
			TLSClientCAFile: tlsClientCAFile,
		},
		ServerPort: serverPort,
		TargetURL:  targetURL,
	}
}
