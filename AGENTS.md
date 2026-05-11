# luna — Project Rules

## What Is This?

`luna` is an OpenCode-based AI agent for remote Linux systems administration.
It operates exclusively through the `luna-interceptor` MCP server over SSH.

## Security Model

- **Bash tool is disabled.** All remote execution goes through `execute_remote`.
- **Read-only by default.** Mutating commands require `allow_mutations=true`.
- **`allow_mutations` is NEVER set without explicit human approval.**
- **No credentials in chat.** Host configs live in `interceptor/hosts.yaml`.
- **Known hosts verification is enforced.** No `InsecureIgnoreHostKey`.

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
```

## Adding Hosts

```bash
cp interceptor/hosts.yaml.example interceptor/hosts.yaml
# Edit hosts.yaml with real host details
ssh-keyscan <host-address> >> ~/.ssh/known_hosts
```

## Project Structure

```
luna/
├── AGENTS.md                    ← this file
├── opencode.json                ← OpenCode config
├── bin/luna-interceptor       ← built binary (gitignored)
├── interceptor/                 ← Go MCP server source
│   ├── hosts.yaml               ← your hosts (gitignored)
│   └── internal/
│       ├── config/              ← hosts.yaml loader
│       ├── ssh/                 ← SSH + SFTP pool
│       ├── security/            ← command allowlist
│       └── tools/               ← MCP tool handlers
└── instructions/
    ├── instructions.md          ← primary agent prompt
    └── agents/                  ← subagent prompts
        ├── debugger.md
        ├── deployer.md
        └── network.md
```
