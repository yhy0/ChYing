# Agent 打开 ChYing 桌面项目

macOS 应用包内置了轻量控制命令 `chyingctl`。Agent 可以请求桌面应用打开一个已有项目并等待 MCP 就绪；若用户已经在 UI 里选好项目并点了「开始」，也可以直接依附当前就绪状态，不必再提供项目名。

支持 Agent Skills 的客户端可以直接使用仓库内的 [`skills/use-chying/SKILL.md`](../skills/use-chying/SKILL.md)，其中包含依附、打开、就绪校验和冲突处理的完整约束。

## 依附当前已就绪项目

用户手动打开项目后：

```bash
/Applications/ChYing.app/Contents/MacOS/chyingctl status --json
# 或
/Applications/ChYing.app/Contents/MacOS/chyingctl open --wait-mcp --json
```

`open` 省略 `--project` 时不会启动或切换项目，只会在 `status=ready` 时返回当前 `project` 与 `mcp_url`。

## 打开指定项目

```bash
/Applications/ChYing.app/Contents/MacOS/chyingctl open \
  --project src-auto \
  --wait-mcp \
  --json
```

成功时会返回当前项目、实际 MCP 地址和桌面进程 PID：

```json
{
  "version": 1,
  "status": "ready",
  "project": "src-auto",
  "mcp_url": "http://127.0.0.1:9090/mcp",
  "pid": 12345
}
```

运行约束：

- 带 `--project` 时仅接受 `~/.config/ChYing/db/<项目>/<项目>.db` 中已经存在的项目。
- 同一项目重复调用是幂等的，只会唤起已有窗口并返回 MCP 地址。
- 桌面应用已经打开其他项目时会明确失败，不会在后台静默切换数据库。
- 省略 `--project` 时只依附当前就绪项目；未就绪则失败并提示用户在 UI 打开或补上 `--project`。
- 启动和就绪状态原子写入 `~/.config/ChYing/runtime.json`，权限为 `0600`，其中不保存密钥或流量内容。
- 本地开发包可用 `bin/ChYing.dev.app/Contents/MacOS/chyingctl`；也可通过 `--app` 或 `CHYING_APP` 指定应用路径。

不需要等待结果时可传 `--wait-mcp=false`。默认超时为 60 秒，可通过 `--timeout 90s` 调整。
