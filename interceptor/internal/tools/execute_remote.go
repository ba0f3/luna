package tools

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ba0f3/luna/interceptor/internal/security"
	"github.com/ba0f3/luna/interceptor/internal/ssh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerExecuteRemote(s *server.MCPServer, pool *ssh.Pool) {
	tool := mcp.NewTool("execute_remote",
		mcp.WithDescription(`Execute a shell command on a remote Linux host via SSH.

READ-ONLY BY DEFAULT: Commands that modify system state are blocked unless
allow_mutations is explicitly set to true. Never set allow_mutations=true
without explicit user approval first.

Returns: stdout, stderr, exit_code, duration, and security classification.

BLOCKED response means the command is permanently forbidden (catastrophic op).
PERMISSION_REQUIRED response means the command is mutating — stop and ask the
human user for approval, then retry with allow_mutations=true.`),
		mcp.WithString("host",
			mcp.Required(),
			mcp.Description("Target host in format [user@]hostname[:port] (e.g. ubuntu@192.168.1.50). Uses current user and port 22 if omitted."),
		),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("Shell command to execute on the remote host"),
		),
		mcp.WithNumber("timeout_sec",
			mcp.Description("Execution timeout in seconds (default: 30, max: 300)"),
		),
		mcp.WithBoolean("allow_mutations",
			mcp.Description("Set to true ONLY after explicit user approval to permit state-changing commands"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		host, err := req.RequireString("host")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		command, err := req.RequireString("command")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Parse timeout (capped at 300s).
		timeoutSec := req.GetFloat("timeout_sec", 30)
		if timeoutSec <= 0 || timeoutSec > 300 {
			timeoutSec = 30
		}
		timeout := time.Duration(timeoutSec) * time.Second

		allowMutations := req.GetBool("allow_mutations", false)

		// ── Security gate ──────────────────────────────────────────────
		check := security.Classify(command)
		log.Printf("execute_remote host=%s class=%s allow_mutations=%v cmd=%q",
			host, check.Class, allowMutations, command)

		switch check.Class {
		case security.Forbidden:
			log.Printf("BLOCKED host=%s cmd=%q reason=%s", host, command, check.Reason)
			return mcp.NewToolResultText(fmt.Sprintf(
				"BLOCKED: %s\n\nCommand: %q\n\nThis command is permanently forbidden and cannot be executed.",
				check.Reason, command,
			)), nil

		case security.Mutating:
			if !allowMutations {
				log.Printf("PERMISSION_REQUIRED host=%s cmd=%q", host, command)
				return mcp.NewToolResultText(fmt.Sprintf(
					"PERMISSION_REQUIRED: %s\n\nCommand: %q\n\nAsk the human user for explicit approval, then retry with allow_mutations=true.",
					check.Reason, command,
				)), nil
			}
			log.Printf("MUTATING APPROVED host=%s cmd=%q", host, command)
		}
		// ── End security gate ──────────────────────────────────────────

		result, err := pool.Execute(host, command, timeout)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("SSH execution error: %v", err)), nil
		}

		log.Printf("execute_remote host=%s exit=%d duration=%s",
			host, result.ExitCode, result.Duration)

		return mcp.NewToolResultText(formatExecResult(host, command, check, allowMutations, result)), nil
	})
}

func formatExecResult(host, command string, check security.CheckResult, mutationsAllowed bool, r ssh.ExecResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Host:     %s\n", host))
	b.WriteString(fmt.Sprintf("Command:  %s\n", command))
	b.WriteString(fmt.Sprintf("Class:    %s\n", check.Class))
	b.WriteString(fmt.Sprintf("Exit:     %d\n", r.ExitCode))
	b.WriteString(fmt.Sprintf("Duration: %s\n", r.Duration.Round(time.Millisecond)))
	if mutationsAllowed {
		b.WriteString("Mutations: APPROVED\n")
	}

	b.WriteString("\n--- STDOUT ---\n")
	if strings.TrimSpace(r.Stdout) == "" {
		b.WriteString("(empty)\n")
	} else {
		b.WriteString(r.Stdout)
		if !strings.HasSuffix(r.Stdout, "\n") {
			b.WriteString("\n")
		}
	}

	if r.Stderr != "" {
		b.WriteString("\n--- STDERR ---\n")
		b.WriteString(r.Stderr)
		if !strings.HasSuffix(r.Stderr, "\n") {
			b.WriteString("\n")
		}
	}

	return b.String()
}
