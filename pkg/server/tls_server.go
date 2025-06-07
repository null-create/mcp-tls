package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

// TLSOptions defines server options for TLS.
type TLSOptions struct {
	CertFile          string       // PEM-encoded server certificate
	KeyFile           string       // PEM-encoded private key
	ClientCAFile      string       // Optional: PEM-encoded CA cert for client verification
	RequireClientCert bool         // Enforce mTLS
	Addr              string       // e.g., ":8443"
	Router            http.Handler // Chi router or any http.Handler
}

// StartSecureServer configures and starts a TLS-enabled HTTP server.
func StartSecureServer(opts TLSOptions, handlers http.Handler) error {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
	}

	// Setup mTLS if CA is provided
	if opts.ClientCAFile != "" {
		caCert, err := os.ReadFile(opts.ClientCAFile)
		if err != nil {
			return fmt.Errorf("failed to read client CA file: %w", err)
		}
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return fmt.Errorf("invalid client CA cert")
		}
		tlsConfig.ClientCAs = caPool

		if opts.RequireClientCert {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
		}
	}

	server := &http.Server{
		Addr:      opts.Addr,
		Handler:   handlers,
		TLSConfig: tlsConfig,
	}

	fmt.Printf("🔐 MCP-TLS Utility Server running at https://%s\n", opts.Addr)
	return server.ListenAndServeTLS(opts.CertFile, opts.KeyFile)
}
