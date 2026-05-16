# luna — Project Rules

## What Is This?

`luna` is an OpenCode-based AI agent for remote Linux systems administration.
It operates exclusively through the `luna-interceptor` MCP server over SSH.

**SSH / OpenSSH parity:** The interceptor is aimed at **full practical compatibility** with the user’s current SSH client—`~/.ssh/known_hosts`, `~/.ssh/config` (supported directives), `SSH_AUTH_SOCK`, and default keys—so hosts reachable with `ssh` should work the same way through Luna unless security policy or an unsupported option forces a difference. See `README.md` (SSH client compatibility).

## Security Model

- **Bash tool is disabled.** All remote execution goes through `execute_remote`.
- **Read-only by default.** Mutating commands require `allow_mutations=true`.
- **`allow_mutations` is NEVER set without explicit human approval.**
- **No credentials in chat.** SSH agent and default keys (`~/.ssh/id_*`) are used.
- **Known hosts verification is enforced.** Hosts must be in `~/.ssh/known_hosts`.

## MCP Interceptor Responses

| Response prefix | Meaning | Action |
|----------------|---------|--------|
| `BLOCKED:` | Command permanently forbidden | Explain to user, do not retry |
| `PERMISSION_REQUIRED:` | Mutating command, not approved | Stop, ask user, then retry with flag |
| Normal output | Command executed | Continue |

## Agents

| Agent | Mode | Role |
|-------|------|------|
| `luna` | primary | Main orchestrator |
| `@debugger` | subagent | Log analysis, root cause |
| `@deployer` | subagent | Approved change execution |
| `@network` | subagent | Connectivity and firewall |

## Building the Interceptor

```bash
cd interceptor
go mod tidy
make build   # → ../bin/luna-interceptor
make test    # → only security/allowlist tests exist currently
make lint    # → requires golangci-lint installed
```

## Project Structure

```
luna/
├── AGENTS.md                    ← this file
├── opencode.json                ← OpenCode config (agents, MCP server, permissions)
├── bin/luna-interceptor         ← built binary (gitignored)
├── interceptor/                 ← Go MCP server source
│   ├── go.mod / go.sum          ← Go module (go 1.25.5, mcp-go, sftp, crypto)
│   ├── Makefile                 ← build, test, lint, tidy, clean, run
│   ├── main.go                  ← MCP stdio server entrypoint
│   └── internal/
│       ├── ssh/                 ← SSH connection pool + SFTP + known_hosts parser
│       ├── security/            ← Command allowlist (ReadOnly / Mutating / Forbidden)
│       └── tools/               ← MCP tool handlers (list_hosts, execute_remote, read_file, transfer_file)
└── instructions/
    ├── instructions.md          ← primary agent prompt (luna)
    └── agents/                  ← subagent prompts
        ├── debugger.md
        ├── deployer.md
        └── network.md
```

## Adding a New MCP Tool

1. Create `interceptor/internal/tools/<name>.go`
2. Add `register<Name>()` call in `tools/tools.go`
3. If the tool executes commands, add classification logic to `security/allowlist.go`
4. Run `make build && make test`

## Infrastructure Knowledge Base

`data/infrastructure/` stores Luna's local infrastructure knowledge base. YAML
files are source data; Markdown files are navigation and notes. Luna records
provenance and confidence for learned facts and redacts secret-like process
arguments before persistence.

## Release

Tags trigger GoReleaser via GitHub Actions (`.github/workflows/release.yml`).
Builds for Linux/Darwin/Windows, amd64 + arm64.

## Important Constraints

- **stdout is reserved for MCP JSON-RPC** — all diagnostics go to stderr.
- `transfer_file` is **always** mutating; it requires `allow_mutations=true` regardless of content.
- `read_file` caps at 100 KB by default (configurable up to 1024 KB).
- `execute_remote` timeout defaults to 30s, max 300s.
- Unknown commands default to **Mutating** (fail-safe).
- Redirection (`>`, `>>`) and command substitution (`` ` ``, `$()`) always classify as **Mutating**.
- The SSH pool caches connections; stale clients are evicted lazily.
