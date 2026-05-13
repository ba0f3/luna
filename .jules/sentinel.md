## 2024-05-18 - Insecure prefix matching in allowlist bypass

**Vulnerability:** The security interceptor used a string prefix check (`strings.HasPrefix`) to block mutating commands like `sed -i` while allowing `sed` as read-only.
**Learning:** This approach is inherently flawed for command-line tools because flags can be reordered. An attacker can place another flag (like `-e`) *before* the blocked flag (e.g., `sed -e 's/a/b/' -i file.txt`), causing the prefix check to fail and allowing the mutating command to bypass the block.
**Prevention:** Rely on regular expressions across the full command string for checking the presence of security-sensitive flags, rather than simple prefix matching. Make sure that regex matches full flags carefully by using `(?:\s|$)` boundaries and `[^\s]*` to catch attached options like `-ibak` without capturing parts of filenames like `my-file-i.txt`.
