package main

import "github.com/null-create/mcp-tls/pkg/server"

func main() {
	server := server.NewServer()
	server.Run()
}
