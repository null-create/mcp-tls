package main

import (
	"github.com/null-create/mcp-tls/pkg/server"
)

func main() {
	router := server.NewRouter()
	server := server.NewServer(router)
	server.Run()
}
