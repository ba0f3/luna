package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ba0f3/luna/interceptor/internal/ssh"
	"github.com/ba0f3/luna/interceptor/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "luna"
	serverVersion = "1.0.0"
)

func main() {
	// All diagnostic output must go to stderr — stdout is reserved for MCP JSON-RPC.
	log.SetOutput(os.Stderr)
	log.SetPrefix("[luna] ")
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)

	// Build SSH connection pool.
	pool := ssh.NewPool()

	// Build and register MCP server.
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	tools.Register(s, pool)

	log.Printf("starting MCP stdio server")

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
