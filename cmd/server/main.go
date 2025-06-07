package main

import (
	"log"

	"github.com/null-create/mcp-tls/pkg/config"
	"github.com/null-create/mcp-tls/pkg/server"
)

func main() {
	router := server.NewRouter()

	cfgs := config.LoadConfigs()

	if cfgs.TLSConfig.TLSEnabled {
		err := server.StartSecureServer(server.TLSOptions{
			CertFile:          cfgs.TLSConfig.TLSCertFile,
			KeyFile:           cfgs.TLSConfig.TLSKeyFile,
			ClientCAFile:      cfgs.TLSConfig.TLSClientCAFile, // Optional
			RequireClientCert: false,                          // Set to true if mTLS is needed
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
