---
name: use-chying
description: Open or focus a ChYing desktop project, wait for its local MCP server, and return the exact ready MCP URL. Use when an agent needs ChYing HTTP history, Repeater, Intruder, passive-scan, or project-scoped MCP tools without asking the user to click through the desktop project selector.
---

# Use ChYing

Open an explicitly named existing ChYing project through the bundled `chyingctl` command. Treat its JSON response as the source of truth for the desktop PID, selected project, and MCP URL.

## Open a project

1. Require the caller to provide the exact ChYing project name. Never infer it from a target, repository, cwd, or bounty Case.
2. Locate `chyingctl` in this order:
   - `command -v chyingctl`
   - `/Applications/ChYing.app/Contents/MacOS/chyingctl`
   - `$HOME/Applications/ChYing.app/Contents/MacOS/chyingctl`
   - `<ChYing repository>/bin/ChYing.app/Contents/MacOS/chyingctl`
   - `<ChYing repository>/bin/ChYing.dev.app/Contents/MacOS/chyingctl`
3. Run:

   ```bash
   chyingctl open --project <project> --wait-mcp --json --timeout 60s
   ```

4. Accept the result only when all of these are true:
   - `status` is `ready`.
   - `project` exactly equals the requested project.
   - `mcp_url` is present and uses loopback HTTP.
   - `pid` is a live ChYing process.
5. Use the returned `mcp_url`; never assume ChYing selected port 9090 or scan 9090–9099 after a successful response.

The same-project call is idempotent and focuses the existing window. If ChYing is already running with another project, stop and report the conflict. Do not terminate ChYing, switch its database, or retry with a guessed project.

## Check status

Run `chyingctl status --json` when the caller only needs the current desktop state. Treat `stopped`, `failed`, a stale PID, or a missing MCP URL as not ready.

## Failure handling

- If the project does not exist, report the exact name and ask the user to create or select the correct ChYing project.
- If `chyingctl` is unavailable, ask for the ChYing application path; pass it with `--app` or set `CHYING_APP`.
- If startup times out, include the last reported status and error. Do not fall back to an unrelated listener.
- Never place credentials, raw HTTP traffic, or tokens in `runtime.json`; it is only a readiness contract.
- Leave the desktop app running unless the user explicitly asks to quit it.
