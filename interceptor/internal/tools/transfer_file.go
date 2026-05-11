package tools

import (
	"context"
	"fmt"
	"log"

	"github.com/ba0f3/luna/interceptor/internal/ssh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTransferFile(s *server.MCPServer, pool *ssh.Pool) {
	tool := mcp.NewTool("transfer_file",
		mcp.WithDescription(`Upload text content to a file on a remote host via SFTP.

MUTATING OPERATION — always requires allow_mutations=true.
Never set allow_mutations=true without explicit user approval first.

Use this to deploy config files, scripts, or other text content.
The content replaces the entire remote file. Binary files are not supported.`),
		mcp.WithString("host",
			mcp.Required(),
			mcp.Description("Target host in format [user@]hostname[:port] (e.g. ubuntu@192.168.1.50)"),
		),
		mcp.WithString("remote_path",
			mcp.Required(),
			mcp.Description("Absolute destination path on the remote host (e.g. /etc/nginx/nginx.conf)"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Text content to write to the remote file (UTF-8)"),
		),
		mcp.WithString("permissions",
			mcp.Description("Octal file permissions string (default: 0644, e.g. 0755 for executables)"),
		),
		mcp.WithBoolean("allow_mutations",
			mcp.Required(),
			mcp.Description("Must be true — confirms user approved this file write operation"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		host, err := req.RequireString("host")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		remotePath, err := req.RequireString("remote_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := req.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		allowMutations := req.GetBool("allow_mutations", false)
		if !allowMutations {
			return mcp.NewToolResultText(
				"PERMISSION_REQUIRED: transfer_file modifies system state.\n\n" +
					"Ask the human user for explicit approval, then retry with allow_mutations=true.",
			), nil
		}

		permissions := req.GetString("permissions", "0644")

		log.Printf("transfer_file APPROVED host=%s path=%q size=%d bytes perm=%s",
			host, remotePath, len(content), permissions)

		if err := pool.WriteFile(host, remotePath, []byte(content), permissions); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("transfer_file error: %v", err)), nil
		}

		log.Printf("transfer_file OK host=%s path=%q", host, remotePath)

		return mcp.NewToolResultText(fmt.Sprintf(
			"OK: Wrote %d bytes to %s:%s (permissions: %s)",
			len(content), host, remotePath, permissions,
		)), nil
	})
}
