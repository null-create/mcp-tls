package main

import (
	"log"

	"github.com/null-create/mcp-tls/pkg/config"
	"github.com/null-create/mcp-tls/pkg/server"
)

func main() {
	router := server.NewRouter()

	cfgs := config.LoadConfig()

	if cfgs.TLSEnabled {
		err := server.StartSecureServer(server.TLSOptions{
			CertFile:          cfgs.TLSCertFile,
			KeyFile:           cfgs.TLSKeyFile,
			ClientCAFile:      cfgs.TLSClientCAFile, // Optional
			RequireClientCert: false,                // Set to true if mTLS is needed
			Addr:              ":8443",
		}, router)

		if err != nil {
			log.Fatalf("❌ TLS server failed: %v", err)
		}
	} else {
		server := server.NewServer(router)
		server.Run()
	}
}
