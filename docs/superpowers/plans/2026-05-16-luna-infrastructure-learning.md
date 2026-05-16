# Luna Infrastructure Learning Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Luna's first OpenCode-facing infrastructure learning, inventory scan, CVE lookup, CVE impact, and Wazuh enrichment capabilities.

**Architecture:** Implement two read-only MCP tools in the Go interceptor: `scan_host_inventory` for normalized SSH inventory and `lookup_cve` for structured external advisory lookup. Keep phase-1 learning and CVE workflows in the primary Luna prompt, backed by a local `data/infrastructure/` YAML/Markdown knowledge base skeleton.

**Tech Stack:** Go 1.25.10, `github.com/mark3labs/mcp-go`, existing Luna SSH pool, Go standard `encoding/json` and `net/http`, OpenCode instruction Markdown, YAML/Markdown repository data files.

---

## File Structure

- Create `interceptor/internal/tools/inventory.go`: MCP registration, fixed collector command list, scan orchestration, JSON response formatting.
- Create `interceptor/internal/tools/inventory_types.go`: inventory response structs shared by tool and tests.
- Create `interceptor/internal/tools/inventory_parse.go`: parsing helpers for package/service/process/port/container command output and command-line redaction.
- Create `interceptor/internal/tools/inventory_parse_test.go`: parser and redaction unit tests with Debian/RHEL/Alpine style samples.
- Create `interceptor/internal/tools/cve_lookup.go`: MCP registration, CVE ID validation, NVD source orchestration, JSON response formatting.
- Create `interceptor/internal/tools/cve_lookup_test.go`: validation and HTTP-source failure/success unit tests using `httptest`.
- Modify `interceptor/internal/tools/tools.go`: register `scan_host_inventory` and `lookup_cve`.
- Modify `instructions/instructions.md`: add primary Luna workflows for automatic learning, inventory scans, CVE impact mapping, Wazuh enrichment, and evidence/confidence reporting.
- Create `data/infrastructure/index.md`: human entry point for the knowledge base.
- Create `data/infrastructure/README.md`: layout, source-of-truth, redaction, and provenance rules.
- Create `data/infrastructure/hosts/.gitkeep`, `data/infrastructure/software/.gitkeep`, `data/infrastructure/cves/.gitkeep`, `data/infrastructure/scans/.gitkeep`: tracked empty directories.

---

### Task 1: Knowledge Base Skeleton and Luna Prompt Workflows

**Files:**
- Create: `data/infrastructure/README.md`
- Create: `data/infrastructure/index.md`
- Create: `data/infrastructure/hosts/.gitkeep`
- Create: `data/infrastructure/software/.gitkeep`
- Create: `data/infrastructure/cves/.gitkeep`
- Create: `data/infrastructure/scans/.gitkeep`
- Modify: `instructions/instructions.md`

- [ ] **Step 1: Create the infrastructure knowledge base skeleton**

Create `data/infrastructure/README.md`:

```markdown
# Luna Infrastructure Knowledge Base

This folder is Luna's local, human-reviewable infrastructure knowledge base.
YAML files are the source data. Markdown files are for navigation and notes.

## Layout

- `hosts/<host-id>/` records what Luna knows about one SSH target.
- `software/<software-id>.yaml` records where software has been observed.
- `cves/CVE-YYYY-NNNN.yaml` records CVE advisory data and host impact.
- `scans/<timestamp>/` stores point-in-time scan evidence.
- `index.md` is the human entry point.

## Evidence Rules

- Every normalized fact must include source, timestamp, and evidence.
- Fresh scan evidence overrides older conversation-derived facts.
- Conversational claims are recorded as low-confidence unless confirmed.
- Wazuh findings are evidence, not replacements for direct host scans.
- Secrets, credentials, tokens, and private keys must never be stored here.
- Process command arguments that look secret-like must be redacted.
```

Create `data/infrastructure/index.md`:

```markdown
# Infrastructure Index

Luna updates this index when it learns about hosts, software, scans, and CVEs.

## Hosts

No hosts documented yet.

## Software

No software documented yet.

## CVEs

No CVEs documented yet.

## Recent Scans

No scans recorded yet.
```

Create empty `.gitkeep` files in:

```text
data/infrastructure/hosts/.gitkeep
data/infrastructure/software/.gitkeep
data/infrastructure/cves/.gitkeep
data/infrastructure/scans/.gitkeep
```

- [ ] **Step 2: Update Luna's primary prompt with learning and vulnerability workflows**

Add these sections to `instructions/instructions.md` after `Workflow Protocol` and before `Sub-Agent Delegation`:

```markdown
## Infrastructure Learning Protocol

Luna automatically maintains `data/infrastructure/` when conversations or tool results reveal infrastructure facts.

### Knowledge Base Rules

- Store structured facts as YAML and human notes as Markdown.
- Record provenance for every fact: source type, timestamp, evidence, and confidence.
- Prefer fresh `scan_host_inventory` evidence over older conversation-derived facts.
- Treat explicit user statements as useful but low-confidence until confirmed by scan or Wazuh evidence.
- Never store credentials, private keys, passwords, tokens, session cookies, or secret values.
- Redact secret-like process arguments before writing command lines.
- Keep Wazuh evidence separate from direct host scan evidence.

### Inventory Scan Workflow

When asked to learn, scan, inventory, or document servers:

1. Call `list_hosts`.
2. Select requested hosts, or ask for scope if the user request is ambiguous.
3. Call `scan_host_inventory` once per selected host.
4. Write scan evidence under `data/infrastructure/scans/<timestamp>/`.
5. Update `data/infrastructure/hosts/<host-id>/` with host, package, service, process, port, container, and vulnerability files.
6. Update `data/infrastructure/software/<software-id>.yaml` cross-references.
7. Update `data/infrastructure/index.md`.
8. Report what was learned, which collectors failed, confidence level, and recommended next checks.

### CVE Impact Workflow

When asked about a CVE:

1. Validate the CVE ID format.
2. Call `lookup_cve`.
3. Check existing `data/infrastructure/` records for affected software and versions.
4. If data is stale or incomplete, ask before running fresh read-only scans.
5. Call `scan_host_inventory` for relevant approved hosts.
6. Use `wazuh-cli agent list --status active` and `wazuh-cli vulnerability list <agent-id>` when Wazuh is configured.
7. Update `data/infrastructure/cves/<CVE>.yaml` with advisory, Wazuh, scan, and impact evidence.
8. Report a host impact matrix with confidence: confirmed, likely, possible, not affected, or unknown.

### Wazuh Enrichment Workflow

Use Wazuh as enrichment when available, without making it mandatory:

- `wazuh-cli agent list --status active`
- `wazuh-cli vulnerability summary <agent-id>`
- `wazuh-cli vulnerability list <agent-id>`

Record Wazuh agent IDs, vulnerability IDs, package names, versions, severity, and timestamps as evidence.
Do not treat Wazuh results as remediation approval.
```

- [ ] **Step 3: Verify prompt/data additions are present**

Run:

```bash
rg -n "Infrastructure Learning Protocol|scan_host_inventory|lookup_cve|Wazuh Enrichment" instructions/instructions.md data/infrastructure
```

Expected: matches in `instructions/instructions.md` and `data/infrastructure/README.md`.

- [ ] **Step 4: Commit Task 1**

Run:

```bash
git add instructions/instructions.md data/infrastructure
git commit -m "Add Luna infrastructure knowledge workflows"
```

Expected: commit succeeds.

---

### Task 2: Inventory Types, Parsers, and Redaction

**Files:**
- Create: `interceptor/internal/tools/inventory_types.go`
- Create: `interceptor/internal/tools/inventory_parse.go`
- Create: `interceptor/internal/tools/inventory_parse_test.go`

- [ ] **Step 1: Add inventory response types**

Create `interceptor/internal/tools/inventory_types.go` with `InventoryScanResult`, `InventoryCollector`, `HostIdentity`, `InventoryPackage`, `InventoryService`, `InventoryProcess`, `InventoryPort`, `InventoryContainer`, and `WazuhHint` structs. Use JSON tags matching the field names from the design spec.

- [ ] **Step 2: Write parser tests before implementation**

Create `interceptor/internal/tools/inventory_parse_test.go` with tests for:

```go
func TestRedactSecretLikeArgs(t *testing.T)
func TestParseDPKGPackages(t *testing.T)
func TestParseRPMPackages(t *testing.T)
func TestParseSystemdServices(t *testing.T)
func TestParseProcessesRedactsCommand(t *testing.T)
func TestParseSSPorts(t *testing.T)
func TestParseWazuhClientKeys(t *testing.T)
```

Test expectations:

- `--password supersecret` becomes `--password [REDACTED]`.
- `--token=abc123` becomes `--token=[REDACTED]`.
- `AWS_SECRET_ACCESS_KEY=secret` becomes `AWS_SECRET_ACCESS_KEY=[REDACTED]`.
- DPKG lines parse as `manager=dpkg`, `name`, `version`, `arch`.
- RPM lines parse as `manager=rpm`, `name`, `version`, `arch`.
- systemd tab-separated lines parse service state and description.
- process command parsing redacts secret-like arguments.
- `ss` output captures protocol, state, local address, and process text.
- Wazuh `client.keys` first line captures agent ID and agent name.

- [ ] **Step 3: Run parser tests and verify they fail**

Run:

```bash
cd interceptor && go test ./internal/tools -run 'TestRedact|TestParse'
```

Expected: FAIL because parser functions are not implemented.

- [ ] **Step 4: Implement parsers and redaction**

Create `interceptor/internal/tools/inventory_parse.go` with:

- `redactSecretLikeArgs(command string) string`
- `isSecretArgName(name string) bool`
- `parseOSRelease(out string) map[string]string`
- `parseDPKGPackages(out string) []InventoryPackage`
- `parseRPMPackages(out string) []InventoryPackage`
- `parseAPKPackages(out string) []InventoryPackage`
- `parseSystemdServices(out string) []InventoryService`
- `parsePSProcesses(out string) []InventoryProcess`
- `parseSSPorts(out string) []InventoryPort`
- `parseDockerContainers(out string) []InventoryContainer`
- `parseWazuhClientKeys(out string) WazuhHint`

Use `strings.Fields`, `strings.Split`, and `strings.Cut`; do not add third-party dependencies.

- [ ] **Step 5: Run parser tests and verify they pass**

Run:

```bash
cd interceptor && go test ./internal/tools -run 'TestRedact|TestParse'
```

Expected: PASS.

- [ ] **Step 6: Commit Task 2**

Run:

```bash
git add interceptor/internal/tools/inventory_types.go interceptor/internal/tools/inventory_parse.go interceptor/internal/tools/inventory_parse_test.go
git commit -m "Add inventory parsing and redaction"
```

Expected: commit succeeds.

---

### Task 3: `scan_host_inventory` MCP Tool

**Files:**
- Create: `interceptor/internal/tools/inventory.go`
- Modify: `interceptor/internal/tools/tools.go`

- [ ] **Step 1: Add the inventory MCP tool implementation**

Create `interceptor/internal/tools/inventory.go` with:

- `const inventorySchemaVersion = "luna.inventory.v1"`
- `type inventoryCommand struct { name string; command string; parse func(*InventoryScanResult, string) }`
- `registerScanHostInventory(s *server.MCPServer, pool *ssh.Pool)`
- `runInventoryScan(pool *ssh.Pool, host string, timeout time.Duration) InventoryScanResult`
- `inventoryCollectors() []inventoryCommand`

Collectors must be fixed read-only commands:

```text
hostname
cat /etc/os-release
uname -srmo
uname -m
uptime -p
dpkg-query -W -f='${Package}\t${Version}\t${Architecture}\n'
rpm -qa --qf '%{NAME}\t%{VERSION}-%{RELEASE}\t%{ARCH}\n'
apk info -vv
systemctl list-units --type=service --all --no-legend --no-pager --plain
ps -eo user=,pid=,pcpu=,pmem=,args= --no-headers
ss -H -tulpen
docker ps -a --format '{{.ID}}\t{{.Image}}\t{{.Names}}\t{{.Status}}'
awk '{print $1, $2}' /var/ossec/etc/client.keys
```

The tool must return indented JSON and must preserve partial collector errors in the `collectors` array.

- [ ] **Step 2: Register the tool**

Modify `interceptor/internal/tools/tools.go` so `Register` includes:

```go
registerScanHostInventory(s, pool)
```

- [ ] **Step 3: Run formatting and tests**

Run:

```bash
cd interceptor && gofmt -w internal/tools/inventory*.go
cd interceptor && go test ./internal/tools
```

Expected: PASS.

- [ ] **Step 4: Commit Task 3**

Run:

```bash
git add interceptor/internal/tools/inventory.go interceptor/internal/tools/inventory_parse.go interceptor/internal/tools/tools.go
git commit -m "Add scan_host_inventory MCP tool"
```

Expected: commit succeeds.

---

### Task 4: `lookup_cve` MCP Tool

**Files:**
- Create: `interceptor/internal/tools/cve_lookup.go`
- Create: `interceptor/internal/tools/cve_lookup_test.go`
- Modify: `interceptor/internal/tools/tools.go`

- [ ] **Step 1: Write CVE lookup tests first**

Create `interceptor/internal/tools/cve_lookup_test.go` with:

```go
func TestValidateCVEID(t *testing.T)
func TestFetchNVDLookupParsesBasicFields(t *testing.T)
func TestFetchNVDLookupHandlesHTTPFailure(t *testing.T)
```

Use `httptest.NewServer` for NVD success and HTTP failure responses. The success response must include `id`, `published`, `lastModified`, English `descriptions`, one reference URL, and `cvssMetricV31` with `baseScore`, `baseSeverity`, and `vectorString`.

- [ ] **Step 2: Run tests and verify they fail**

Run:

```bash
cd interceptor && go test ./internal/tools -run 'TestValidateCVE|TestFetchNVD'
```

Expected: FAIL because CVE lookup functions are not implemented.

- [ ] **Step 3: Implement `lookup_cve`**

Create `interceptor/internal/tools/cve_lookup.go` with:

- `const nvdCVEAPI = "https://services.nvd.nist.gov/rest/json/cves/2.0"`
- `normalizeCVEID(raw string) (string, bool)`
- `fetchNVDLookup(baseURL, cveID string, client *http.Client) (CVELookupResult, error)`
- `registerLookupCVE(s *server.MCPServer)`
- `CVELookupResult` with schema version, ID, source, summary, published, last modified, severity, CVSS score/vector, references, and errors.
- NVD response structs local to the file.

Invalid CVE IDs return a text result prefixed with `INVALID_CVE_ID:`. External source errors return structured JSON with an `errors` array instead of crashing the tool.

- [ ] **Step 4: Register the CVE tool**

Modify `interceptor/internal/tools/tools.go` so `Register` includes:

```go
registerLookupCVE(s)
```

- [ ] **Step 5: Run formatting and tests**

Run:

```bash
cd interceptor && gofmt -w internal/tools/cve_lookup.go internal/tools/cve_lookup_test.go internal/tools/tools.go
cd interceptor && go test ./internal/tools
```

Expected: PASS.

- [ ] **Step 6: Commit Task 4**

Run:

```bash
git add interceptor/internal/tools/cve_lookup.go interceptor/internal/tools/cve_lookup_test.go interceptor/internal/tools/tools.go
git commit -m "Add lookup_cve MCP tool"
```

Expected: commit succeeds.

---

### Task 5: Integration Verification and Documentation Alignment

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`

- [ ] **Step 1: Document new MCP tools**

Update the MCP tools table in `README.md` to include:

```markdown
| `scan_host_inventory` | Run a fixed read-only inventory scan and return normalized JSON | No |
| `lookup_cve` | Query external CVE advisory sources and return normalized JSON | No |
```

Update `AGENTS.md` with:

```markdown
## Infrastructure Knowledge Base

`data/infrastructure/` stores Luna's local infrastructure knowledge base. YAML files are source data; Markdown files are navigation and notes. Luna records provenance and confidence for learned facts and redacts secret-like process arguments before persistence.
```

- [ ] **Step 2: Run full verification**

Run:

```bash
cd interceptor && make test
cd interceptor && make build
```

Expected: both commands succeed.

- [ ] **Step 3: Review git diff**

Run:

```bash
git diff --stat
git status --short
```

Expected: only intended Luna infrastructure learning files are modified or added.

- [ ] **Step 4: Commit Task 5**

Run:

```bash
git add README.md AGENTS.md
git commit -m "Document Luna infrastructure learning tools"
```

Expected: commit succeeds.

---

## Self-Review Checklist

- Spec coverage: Tasks cover the data folder, primary Luna workflows, `scan_host_inventory`, `lookup_cve`, Wazuh enrichment instructions, redaction, docs, and verification.
- Placeholder scan: No open implementation placeholders remain.
- Type consistency: `InventoryScanResult`, `CVELookupResult`, `registerScanHostInventory`, and `registerLookupCVE` are named consistently across tasks.

