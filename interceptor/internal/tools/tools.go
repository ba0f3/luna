package tools

import (
	"github.com/ba0f3/luna/interceptor/internal/ssh"
	"github.com/mark3labs/mcp-go/server"
)

// Register wires all luna MCP tools onto the server.
func Register(s *server.MCPServer, pool *ssh.Pool) {
	registerListHosts(s)
	registerExecuteRemote(s, pool)
	registerReadFile(s, pool)
	registerTransferFile(s, pool)
	registerScanHostInventory(s, pool)
	registerLookupCVE(s)
}
