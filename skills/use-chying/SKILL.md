---
name: use-chying
description: Open or attach to a ChYing desktop project, wait for its local MCP server, and return the exact ready MCP URL. Use when an agent needs ChYing HTTP history, Repeater, Intruder, passive-scan, or project-scoped MCP tools. Prefer attaching when the user already opened a project in the desktop UI.
---

# Use ChYing

Use the bundled `chyingctl` command. Treat its JSON response as the source of truth for the desktop PID, selected project, and MCP URL.

## Prefer attach when the user already opened ChYing

If the human selected a project in the desktop UI and clicked Start, do **not** invent or require a project name first.

1. Locate `chyingctl` in this order:
   - `command -v chyingctl`
   - `/Applications/ChYing.app/Contents/MacOS/chyingctl`
   - `$HOME/Applications/ChYing.app/Contents/MacOS/chyingctl`
   - `<ChYing repository>/bin/ChYing.app/Contents/MacOS/chyingctl`
   - `<ChYing repository>/bin/ChYing.dev.app/Contents/MacOS/chyingctl`
2. Run either:

   ```bash
   chyingctl status --json
   ```

   or attach without naming a project:

   ```bash
   chyingctl open --wait-mcp --json --timeout 60s
   ```

3. Accept the result only when all of these are true:
   - `status` is `ready`.
   - `project` is non-empty.
   - `mcp_url` is present and uses loopback HTTP.
   - `pid` is a live ChYing process.
4. Use the returned `project` and `mcp_url` as the active ChYing identity. Never assume port 9090 or scan 9090–9099.

## Open a named project when nothing is ready

Only when status/attach is not ready, and the caller already knows the exact project name:

```bash
chyingctl open --project <project> --wait-mcp --json --timeout 60s
```

Rules:

- Require the exact ChYing project name for this path. Never infer it from a target, repository, or cwd.
- Same-project call is idempotent and focuses the existing window.
- If ChYing is already running with another project, stop and report the conflict. Do not terminate ChYing, switch its database, or retry with a guessed project.

## Failure handling

- If nothing is ready and no project name is available, ask the user to open the correct project in the ChYing desktop UI (or provide the exact project name).
- If the named project does not exist, report the exact name and ask the user to create or select the correct ChYing project.
- If `chyingctl` is unavailable, ask for the ChYing application path; pass it with `--app` or set `CHYING_APP`.
- If startup times out, include the last reported status and error. Do not fall back to an unrelated listener.
- Never place credentials, raw HTTP traffic, or tokens in `runtime.json`; it is only a readiness contract.
- Leave the desktop app running unless the user explicitly asks to quit it.
