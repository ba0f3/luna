---
name: wazuh-cli
description: >
  Interact with Wazuh SIEM/XDR via CLI. Manage agents, query vulnerabilities,
  administer rules/decoders, run FIM/SCA, execute active response, manage the
  cluster, and control RBAC security — all with structured JSON or Markdown output
  that is easy to parse and act on programmatically.
---

# wazuh-cli Agent Skill

## Overview

`wazuh-cli` is a command-line tool designed for AI agents to interact with the
Wazuh Server API without any human-in-the-loop. All inputs come from flags;
no interactive prompts are ever issued during normal operation.

**Binary**: `wazuh-cli`  
**Config**: `~/.config/wazuh/config.json`  
**Default output**: JSON (use `--output markdown` for tables)

---

## Installation

### One-liner (Linux & macOS) — fastest
```bash
curl -fsSL https://raw.githubusercontent.com/ba0f3/wazuh-cli/main/install.sh | sh
```

### Homebrew (macOS & Linux)
```bash
brew install ba0f3/tap/wazuh-cli
```

### Go Install
```bash
go install github.com/ba0f3/wazuh-cli@latest
```

### Alternative: Pre-built Binaries
Download the latest binary from the [GitHub Releases](https://github.com/ba0f3/wazuh-cli/releases) page.

### From Source
```bash
git clone https://github.com/ba0f3/wazuh-cli.git
cd wazuh-cli
make build
mv bin/wazuh-cli $HOME/bin/wazuh-cli
```

## Setup

### Option A — Interactive login (recommended for security)

```bash
wazuh-cli auth login --url https://wazuh-server:55000 --user admin --insecure
```

### Option B — Config file (for persistent use)

```bash
wazuh-cli init \
  --url https://wazuh-server:55000 \
  --user admin \
  --password "your-password" \
  --insecure
```

### Option C — Environment variables (for ephemeral/CI use)

```bash
export WAZUH_URL=https://wazuh-server:55000
export WAZUH_USER=admin
export WAZUH_PASSWORD=your-password
```

### Option D — Config command (modify ~/.config/wazuh/config.json)

```bash
wazuh-cli config set url https://wazuh-server:55000
wazuh-cli config set user admin
wazuh-cli config set indexer_url https://wazuh-indexer:9200
wazuh-cli config list

# For password/indexer_password — use -P (stdin) or -p (inline flag)
wazuh-cli config set password -P < /run/secrets/wazuh_pass   # safest
wazuh-cli config set password -p "s3cr3t"                     # inline flag fallback
wazuh-cli config set indexer_password -P < /run/secrets/idx_pass
```

### Option E — Inline flags

```bash
wazuh-cli --url https://wazuh-server:55000 --user admin --password secret agent list
```

---

## Output Format

| Format | Flag | Best for |
|---|---|---|
| JSON | `--output json` (default) | Programmatic parsing, `jq` pipelines |
| Markdown | `--output markdown` | Human-readable tables in reports |
| Raw | `--output raw` | Direct API response pass-through |

**JSON output structure** (Wazuh API data items):
```json
[
  { "id": "001", "name": "web-server-01", "status": "active", ... }
]
```

**Error envelope** (exit code ≠ 0):
```json
{ "error": true, "code": 2, "message": "...", "detail": "..." }
```

**Exit codes**:
- `0` — success
- `1` — client error (config, network)
- `2` — API error
- `3` — authentication failure
- `4` — permission denied (RBAC)

---

## Command Reference

### Alerts (Wazuh Indexer)

> [!NOTE]
> Alerts are stored in the Wazuh Indexer (OpenSearch), not the Manager API.
> You must configure at least the `indexer_url` via `wazuh-cli config set indexer_url`
> to use these commands. If `indexer_user` and `indexer_password` are omitted,
> `wazuh-cli` will automatically use the Manager API credentials.

```bash
# List recent alerts (default limit: 50)
wazuh-cli alert list

# Filter alerts by rule level and time
wazuh-cli alert list --level 10 --from "now-1h"

# Search alerts by agent ID or rule ID
wazuh-cli alert list --agent-id 001 --rule-id 5710

# Get a specific alert by document ID
wazuh-cli alert get eA3_p40BqY8g5Jz_ZfK8

# Get alert statistics grouped by rule level (default)
wazuh-cli alert stats --from "now-24h"

# Get alert statistics grouped by agent
wazuh-cli alert stats --group-by agent
```

### Agents

```bash
# List all active agents
wazuh-cli agent list --status active

# Get specific agent
wazuh-cli agent get 001

# Delete agent
wazuh-cli agent delete 001

# Restart agent
wazuh-cli agent restart 001

# Get status summary (counts by status)
wazuh-cli agent summary

# Upgrade agent
wazuh-cli agent upgrade 001 --version v4.9.0

# Get agent running config
wazuh-cli agent config 001 --component agent --configuration client
```

### Agent Groups

```bash
wazuh-cli group list
wazuh-cli group create production
wazuh-cli group agents production
wazuh-cli group add-agent production 001
wazuh-cli group remove-agent production 001
wazuh-cli group delete production
```

### Rules

```bash
# List all rules
wazuh-cli rule list

# Filter by level 10+
wazuh-cli rule list --level 10-16

# Get specific rule
wazuh-cli rule get 5710

# List rule files
wazuh-cli rule files

# List rule groups
wazuh-cli rule groups
```

### Decoders

```bash
wazuh-cli decoder list
wazuh-cli decoder get sshd
wazuh-cli decoder files
```

### Vulnerabilities

```bash
# List critical vulnerabilities for an agent
wazuh-cli vulnerability list 001 --severity critical

# Vulnerability summary by severity
wazuh-cli vulnerability summary 001
```

### Security Configuration Assessment (SCA)

```bash
# List SCA policies for agent
wazuh-cli sca list 001

# Get SCA checks for a policy
wazuh-cli sca checks 001 cis_debian10
```

### System Inventory (Syscollector)

```bash
wazuh-cli syscollector hardware 001
wazuh-cli syscollector os 001
wazuh-cli syscollector packages 001
wazuh-cli syscollector processes 001
wazuh-cli syscollector ports 001
```

### File Integrity Monitoring (FIM)

```bash
# List FIM entries
wazuh-cli syscheck list 001

# Filter by file
wazuh-cli syscheck list 001 --file /etc/os-release

# Run FIM scan
wazuh-cli syscheck run 001

# Get last scan time
wazuh-cli syscheck last-scan 001
```

### Active Response

```bash
# Block an IP address
wazuh-cli active-response run \
  --agent-id 001 \
  --command firewall-drop \
  --arguments "-", "null", "192.168.1.100"
```

### Cluster

```bash
wazuh-cli cluster status
wazuh-cli cluster nodes
wazuh-cli cluster health
wazuh-cli cluster config
wazuh-cli cluster restart
```

### Manager

```bash
wazuh-cli manager info
wazuh-cli manager status
wazuh-cli manager logs --level error
wazuh-cli manager restart
wazuh-cli manager validation
```

### RBAC Security

```bash
# Users
wazuh-cli security user list
wazuh-cli security user create --username analyst --password P@ssw0rd!

# Roles
wazuh-cli security role list
wazuh-cli security role create --name readonly

# Policies
wazuh-cli security policy list
```

### CDB Lists

```bash
wazuh-cli cdb-list list
wazuh-cli cdb-list get audit-keys
```

### MITRE ATT&CK

```bash
wazuh-cli mitre list
wazuh-cli mitre get T1110
```

### Log Test

```bash
# Test a syslog entry
wazuh-cli logtest run --event "Jan 15 12:34:56 myhost sshd[1234]: Failed password for root"
```

### Tasks

```bash
wazuh-cli task list
wazuh-cli task list --status "In progress"
```

---

## Common Agent Workflows

### Investigate a compromised agent

```bash
# 1. Get recent high-level alerts
wazuh-cli alert list --agent-id 001 --level 12 --from "now-24h" | jq '.[].rule.description'

# 2. Get agent details
wazuh-cli agent get 001

# 2. List critical vulnerabilities
wazuh-cli vulnerability list 001 --severity critical | jq '.[].cve'

# 3. Check running processes
wazuh-cli syscollector processes 001 | jq '.[] | select(.name | test("suspicious"))'

# 4. Check open ports
wazuh-cli syscollector ports 001

# 6. Get recent FIM changes
wazuh-cli syscheck list 001

# 7. Run active response to isolate (if needed)
wazuh-cli active-response run --agent-id 001 --command firewall-drop --arguments "-", "null", "0.0.0.0"
```

### Audit agents by compliance

```bash
# 1. Find non-compliant agents
wazuh-cli agent list --status active | \
  jq -r '.[].id' | \
  xargs -I{} sh -c 'echo {} && wazuh-cli sca list {} | jq ".[] | select(.fail > 0)"'
```

### Find agents with a specific vulnerability (CVE)

```bash
wazuh-cli agent list --status active | jq -r '.[].id' | while read id; do
  wazuh-cli vulnerability list "$id" --output json | \
    jq --arg id "$id" --arg cve "CVE-2024-1234" \
      '.[] | select(.cve == $cve) | {agent: $id, cve: .cve, severity: .severity}'
done
```

---

## Error Handling Patterns

```bash
# Check exit code
wazuh-cli agent list; echo "Exit: $?"

# Parse JSON errors
result=$(wazuh-cli agent get 999)
if echo "$result" | jq -e '.error == true' > /dev/null 2>&1; then
  echo "Error: $(echo "$result" | jq -r '.message')"
fi

# Re-authenticate (token expired)
# The CLI handles this automatically — just retry the command.
```

---

## Security Notes

- Config file requires `0600` permissions (enforced on load)
- JWT tokens are cached at `~/.config/wazuh/token` with `0600` permissions
- Use `config set password -P` to read password from stdin (safest — never in history)
- Use `config set password -p <value>` to pass inline as a flag (not in history, but visible in `ps`)
- Use `WAZUH_PASSWORD` env var instead of `--password` for process visibility
- Use `--password -` on the root command to read one line from stdin before config resolution
- Use `--insecure` only with self-signed certs and on trusted networks
