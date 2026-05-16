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
