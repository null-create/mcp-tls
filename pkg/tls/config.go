package tls

type TLSConfig struct {
	TLSEnabled      bool   // whether the validation server has TLS enabled
	TLSKeyFile      string // (OPTIONAL) path to server.key file if TLS is enabled
	TLSCertFile     string // (OPTIONAL) path to server.crt file if TLS is enabled
	TLSClientCAFile string // (OPTIONAL) path to client ca.crt file if TLS is enabled
}
