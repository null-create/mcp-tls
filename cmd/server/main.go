package main

import (
	"log"
	"os"

	"github.com/null-create/mcp-tls/pkg/server"
)

func main() {
	router := server.NewRouter()

	tlsEnabled := os.Getenv("TLS_ENABLED")
	if tlsEnabled != "" && tlsEnabled == "true" {
		err := server.StartSecureServer(server.TLSOptions{
			CertFile:          "certs/server.crt",
			KeyFile:           "certs/server.key",
			ClientCAFile:      "certs/ca.crt", // Optional
			RequireClientCert: false,          // Set to true if mTLS is needed
			Addr:              ":8443",
		}, router)

		if err != nil {
			log.Fatalf("TLS server failed: %v", err)
		}
	} else {
		server := server.NewServer(router)
		server.Run()
	}
}
