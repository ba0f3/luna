## 2024-05-18 - Insecure prefix matching in allowlist bypass

**Vulnerability:** The security interceptor used a string prefix check (`strings.HasPrefix`) to block mutating commands like `sed -i` while allowing `sed` as read-only.
**Learning:** This approach is inherently flawed for command-line tools because flags can be reordered. An attacker can place another flag (like `-e`) *before* the blocked flag (e.g., `sed -e 's/a/b/' -i file.txt`), causing the prefix check to fail and allowing the mutating command to bypass the block.
**Prevention:** Rely on regular expressions across the full command string for checking the presence of security-sensitive flags, rather than simple prefix matching. Make sure that regex matches full flags carefully by using `(?:\s|$)` boundaries and `[^\s]*` to catch attached options like `-ibak` without capturing parts of filenames like `my-file-i.txt`.

## 2025-02-18 - Command Wrapper Bypass via Start-of-String Regex and ReadOnly Allowlist

**Vulnerability:** The security interceptor used start-of-string or operator boundaries `(?:^|[|&;])\s*` for forbidden commands like `sudo` and `useradd`. This allowed bypasses when forbidden commands were passed as arguments to command wrappers or safe commands (e.g., `env sudo rm -rf /`, `echo sudo rm -rf /`). Additionally, the `env` command was listed in `readOnlyPrefixes`, incorrectly classifying arbitrary command execution through it as safe (e.g., `env touch /tmp/hacked`).
**Learning:** Command line argument parsing and classification must account for command runners (`env`, `time`, `xargs`, `awk system()`) which execute their arguments. Furthermore, when analyzing an unquoted, normalized command string, relying on start-of-string anchors `^` for blocking dangerous binaries is insufficient because the malicious binary can be placed anywhere in the argument list.
**Prevention:** Use word boundaries `\b` rather than start-of-string boundaries `^` to catch forbidden binaries anywhere in the command string (e.g., `(?i)\bsudo\b`). Remove command runners like `env` from read-only allowlists, ensuring they fall back to `Mutating` status where human approval is required.
