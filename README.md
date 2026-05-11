# luna

An OpenCode-based AI agent for remote Linux systems administration via SSH.
All remote operations are routed through a Go MCP interceptor with a security gate.

## SSH client compatibility

Luna aims to be **as compatible as practical with your existing OpenSSH client**—the same host you reach with `ssh` should work through the interceptor without re‑inventing credentials or host trust. Concretely, the Go SSH stack follows your **`~/.ssh/known_hosts`**, **`~/.ssh/config`** (where the parser supports it: e.g. `User`, `IdentityFile`, `CertificateFile`, `IdentitiesOnly`, `StrictHostKeyChecking`, `HostKeyAlgorithms`), **`SSH_AUTH_SOCK`**, and common **`~/.ssh/id_*`** keys. Gaps can still exist (e.g. `ProxyJump`, `Match`, or vendor‑specific agent behavior); treat parity as a **design goal** and report mismatches.

## Architecture

```
OpenCode (LLM) ──MCP stdio──► luna-interceptor (Go binary)
                                         │
                               ┌─────────┴─────────┐
                               │                   │
                             SSH                 SFTP
                               │                   │
                          Remote Linux         Remote Linux
                          (execute_remote)     (read_file / transfer_file)
```

## Agents

| Agent | Mode | Role |
|-------|------|------|
| `luna` | Primary | Main orchestrator |
| `@debugger` | Subagent | Log analysis, root cause investigation |
| `@deployer` | Subagent | Approved change execution |
| `@network` | Subagent | Connectivity, firewall, DNS diagnostics |

## MCP Tools

| Tool | Description | Mutations |
|------|-------------|-----------|
| `list_hosts` | List configured hosts | No |
| `execute_remote` | Run shell command via SSH | No (by default) |
| `read_file` | Fetch file via SFTP | No |
| `transfer_file` | Upload file via SFTP | Yes — requires approval |

## Quick Start

### 1. Install the interceptor

Download the latest release from GitHub into the `bin/` directory:

```bash
mkdir -p bin
curl -sL https://github.com/ba0f3/luna/releases/latest/download/luna-interceptor_Linux_x86_64.tar.gz | tar xz -C bin/ luna-interceptor
```

*(Alternatively, build from source: `cd interceptor && go mod tidy && make build`)*

### 2. Open in OpenCode

```bash
cd /path/to/luna
opencode
```

The `luna-interceptor` MCP server starts automatically.

## Security Model

- **Bash is disabled** at the OpenCode level — no local shell execution.
- **Read-only by default** — mutating commands need `allow_mutations=true`.
- **`allow_mutations` is never set without explicit human approval.**
- **Known hosts verification** — requires hosts to be in `~/.ssh/known_hosts`.
- **No credentials in prompts** — SSH agent and default keys are used for authentication.
- **OpenSSH parity** — see [SSH client compatibility](#ssh-client-compatibility); behavior should match your normal `ssh` setup where supported.

## Adding a New MCP Tool

1. Create `interceptor/internal/tools/<name>.go`
2. Add `register<Name>()` call in `tools/tools.go`
3. Add commands to `security/allowlist.go` (read-only or mutating)
4. Run `make build && make test`

## License

MIT
