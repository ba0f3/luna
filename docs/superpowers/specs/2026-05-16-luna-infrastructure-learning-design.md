# Luna Infrastructure Learning and CVE Impact — Design Plan

**Date:** 2026-05-16
**Status:** Draft for review
**Scope:** OpenCode Luna instructions, MCP interceptor read-only tools, local infrastructure knowledge base

## Summary

Add OpenCode-facing Luna capabilities for automatic infrastructure learning, structured server inventory, CVE lookup, CVE impact mapping, and Wazuh vulnerability enrichment. The first implementation should use a hybrid design: add a small number of read-only MCP tools for normalized data collection, then teach Luna how to persist and cross-reference that data in a local `data/infrastructure/` knowledge base.

## Goals

1. Automatically learn infrastructure facts from conversations and tool outputs.
2. Scan servers through one read-only MCP call instead of many fragile `execute_remote` loops.
3. Store host, software, service, process, port, container, scan, and vulnerability data in a cross-reference-friendly folder structure.
4. Look up CVEs from external advisory sources when users ask about a CVE.
5. Map CVEs to known and freshly scanned servers based on installed software, services, versions, containers, and Wazuh evidence.
6. Use `wazuh-cli` vulnerability data when available, without making Wazuh mandatory.
7. Preserve Luna's approval model: scans and lookups are read-only; remediation remains human-approved.

## Non-Goals

- Automatic patching or package upgrades.
- Running unauthenticated network-wide scanners.
- Replacing Wazuh as a vulnerability source of record.
- Storing credentials, SSH keys, passwords, or tokens in the knowledge base.
- Treating conversational claims as equally reliable as scan evidence.

## Recommended Approach

Use a hybrid MCP-tool and OpenCode-workflow design.

### Why This Approach

- `scan_host_inventory` gives Luna one stable read-only call with normalized JSON output.
- `lookup_cve` gives Luna structured external advisory data instead of relying on free-form browsing.
- OpenCode instructions remain responsible for orchestration, explanation, approval boundaries, and documentation.
- The local data folder stays human-readable and can be reviewed or edited by operators.

## Architecture

```text
OpenCode Luna
  |
  |-- MCP: list_hosts
  |-- MCP: scan_host_inventory(host)
  |-- MCP: lookup_cve(cve_id)
  |-- MCP: execute_remote / read_file for targeted follow-up
  |
  |-- local data/infrastructure/
        |-- hosts/<host-id>/
        |-- software/<software-id>.yaml
        |-- cves/CVE-YYYY-NNNN.yaml
        |-- scans/<timestamp>/
        |-- index.md
```

## New MCP Tools

### `scan_host_inventory`

Read-only SSH inventory collector. It should use the existing SSH pool and security posture.

Inputs:

- `host` required.
- `timeout_sec` optional, capped like `execute_remote`.
- Optional scan sections can be added later, but phase 1 should use a conservative default set.

Collector output:

- Host identity: hostname, OS release, kernel, architecture, uptime.
- Packages: `dpkg`, `rpm`, `apk`, `snap`, or other available package managers.
- Services: systemd units and active service state when systemd exists.
- Processes: running process name, PID, user, command, CPU, memory.
- Ports: listening TCP/UDP sockets, process names/PIDs when visible.
- Containers: Docker and Kubernetes read-only summaries when available.
- Wazuh hints: local Wazuh agent ID/name/version when discoverable through read-only files or commands.
- Process command lines with secret-like arguments redacted by default.

Constraints:

- No remote writes.
- No package index refresh.
- No service restarts.
- No privileged escalation.
- Unknown or unavailable collectors return partial results with explicit errors.

### `lookup_cve`

External advisory lookup tool.

Inputs:

- `cve_id` required.

Output:

- CVE ID, title/summary, publication and modification dates.
- Severity and CVSS vectors where available.
- Affected products and version ranges.
- Known exploited status when available.
- References to NVD, GitHub advisories, vendor advisories, and CISA KEV if applicable.
- Remediation guidance from advisory sources.

The tool should keep advisory evidence separate from local host exposure evidence.

`lookup_cve` should live in the Luna interceptor so CVE lookup is a structured MCP capability rather than an ad hoc prompt workflow.

## Knowledge Base Layout

```text
data/infrastructure/
  index.md
  hosts/
    <host-id>/
      host.yaml
      processes.yaml
      services.yaml
      packages.yaml
      ports.yaml
      containers.yaml
      vulnerabilities.yaml
      notes.md
  software/
    <software-id>.yaml
  cves/
    CVE-YYYY-NNNN.yaml
  scans/
    YYYY-MM-DDTHH-MM-SSZ/
      manifest.yaml
      findings.yaml
```

YAML is the source data format. Markdown is used only for navigation and human notes.

## Cross-Reference Model

### Host Records

Host files answer: "What is on this server?"

Each host record should include:

- Stable host ID.
- SSH target used by Luna.
- Last scan timestamp.
- OS and kernel facts.
- Discovered services, processes, packages, ports, containers, and Wazuh agent mapping.
- Evidence references back to scan files or conversation notes.

### Software Records

Software files answer: "Where is this software running?"

Each software record should include:

- Normalized software name.
- Observed versions.
- Hosts and source evidence.
- Related services, ports, containers, and packages.
- Related CVEs.

### CVE Records

CVE files answer: "Which hosts may be affected by this CVE?"

Each CVE record should include:

- Advisory summary.
- Affected product/version rules.
- Known exploited status.
- Host impact matrix.
- Wazuh findings.
- Scan evidence.
- Confidence level: confirmed, likely, possible, not affected, unknown.

### Scan Records

Scan files preserve point-in-time evidence.

Each scan should include:

- Timestamp.
- Host.
- Tool version or schema version.
- Successful and failed collectors.
- Raw normalized findings used to update host/software/CVE records.

## Automatic Learning Workflow

Luna should automatically update `data/infrastructure/` after relevant conversations and tool results.

Rules:

- Persist infrastructure facts from scans, diagnostics, CVE analysis, Wazuh output, and explicit user statements.
- Store provenance for every fact: source type, timestamp, and short evidence text or scan reference.
- Separate observed facts from inferred notes.
- Prefer fresh scan evidence over older conversation-derived data.
- Never store secrets or credentials.
- If a fact is uncertain, mark it as `confidence: low` or write it to `notes.md` instead of a normalized record.

## Inventory Scan Workflow

When the user asks Luna to learn, scan, document, or inventory servers:

1. Call `list_hosts`.
2. Select requested hosts or ask for scope if ambiguous.
3. Call `scan_host_inventory` once per host.
4. Write scan evidence under `data/infrastructure/scans/<timestamp>/`.
5. Update host records under `data/infrastructure/hosts/<host-id>/`.
6. Update software cross-reference files.
7. Update `data/infrastructure/index.md`.
8. Report what was learned, what failed, and what needs approval or follow-up.

## CVE Impact Workflow

When the user asks about a CVE:

1. Validate the CVE ID format.
2. Call `lookup_cve`.
3. Check existing `data/infrastructure/` records for potentially affected software.
4. Ask whether to run fresh scans if existing data is stale or incomplete.
5. If approved for read-only scanning scope, call `scan_host_inventory` for relevant hosts.
6. Query Wazuh via `wazuh-cli vulnerability list` or summaries when Wazuh is configured.
7. Create or update `data/infrastructure/cves/<CVE>.yaml`.
8. Report an impact matrix with confidence and evidence.

## Wazuh Enrichment

Use `wazuh-cli` as an enrichment source when configured.

Expected commands:

```bash
wazuh-cli agent list --status active
wazuh-cli vulnerability summary <agent-id>
wazuh-cli vulnerability list <agent-id>
```

Wazuh evidence should be recorded with source metadata. It should not override direct scan findings unless the evidence clearly identifies the same host/software/version.

## Security and Approval Boundaries

- `scan_host_inventory` is read-only and must not require `allow_mutations=true`.
- `lookup_cve` is read-only and external-network-facing.
- `transfer_file` and all remediation commands remain approval-gated.
- CVE remediation plans must be proposed first and executed only after explicit human approval.
- Unknown commands in ad hoc follow-up still default to mutating under the existing interceptor policy.

## OpenCode Instruction Changes

Update `instructions/instructions.md` with:

- Automatic learning rules.
- Knowledge base write policy.
- Inventory scan workflow.
- CVE lookup and impact workflow.
- Wazuh enrichment workflow.
- Output sections for "Evidence", "Confidence", and "Next Scan/Remediation".

Consider adding subagents later:

- `@inventory` for host inventory and data maintenance.
- `@vulnerability` for CVE impact analysis and Wazuh correlation.

Phase 1 keeps these workflows in the primary `luna` prompt to avoid adding orchestration complexity too early. Dedicated `@inventory` and `@vulnerability` subagents are deferred until the primary workflows prove stable.

## Testing Plan

Interceptor tests:

- `scan_host_inventory` rejects missing host.
- Collector parser handles Debian, RHEL, Alpine, and missing-command outputs.
- Partial collector failures still return valid JSON.
- No mutating shell fragments appear in the fixed collector command set.
- `lookup_cve` validates CVE IDs.
- `lookup_cve` handles source failures and returns structured errors.

Instruction/workflow checks:

- Luna writes host records after a scan.
- Luna writes CVE records after lookup.
- Luna labels stale data and asks about fresh scans.
- Luna keeps Wazuh evidence separate from direct scan evidence.
- Luna does not propose or execute remediation without approval.

Manual verification:

- Build interceptor: `cd interceptor && make build`.
- Run tests: `cd interceptor && make test`.
- In OpenCode, run an inventory scan against a known host.
- Ask about a CVE and verify host/software/CVE cross-references are updated.

## Implementation Phases

### Phase 1: Knowledge Base and Instructions

- Add `data/infrastructure/` skeleton and schema examples.
- Update Luna instructions with automatic learning and CVE workflows.
- Add documentation for the knowledge base layout.

### Phase 2: `scan_host_inventory`

- Add MCP tool implementation.
- Add read-only collectors and normalized output structs.
- Add tests for parsing and failure handling.

### Phase 3: `lookup_cve`

- Add MCP tool implementation inside the Luna interceptor.
- Query external CVE/advisory sources.
- Normalize advisory output.
- Add tests for validation and source failure handling.

### Phase 4: Wazuh Integration

- Add Luna workflow guidance for `wazuh-cli`.
- Add host-to-Wazuh-agent mapping in the knowledge base.
- Record Wazuh vulnerability evidence in host and CVE files.

## Resolved Design Decisions

1. Phase 1 keeps inventory and vulnerability workflows in the primary Luna prompt. New subagents can be added later if the workflows become large or repetitive.
2. `lookup_cve` lives in the Luna interceptor as a structured read-only MCP tool.
3. `scan_host_inventory` redacts process arguments that look secret-like by default.
