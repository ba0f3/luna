## 2025-05-12 - [Critical] Fix command obfuscation bypass via shell escaping
**Vulnerability:** The command string parser (`mvdan.cc/sh/v3/syntax`) retains backslash escapes in `*syntax.Lit` values. This allowed attackers to bypass command validation by obfuscating forbidden or mutating commands (e.g., `s\udo`, `\rm`).
**Learning:** Shell parsers provide exact literal strings; static validation must perform "quote removal" and unescaping to normalize the command to what Bash will actually execute.
**Prevention:** An `unescape` helper was added to strip literal backslashes during AST walking before comparing against allow/deny lists.
