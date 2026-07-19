# Agent 打开 ChYing 桌面项目

macOS 应用包内置了轻量控制命令 `chyingctl`。Agent 不需要点击项目选择页，可以直接请求桌面应用打开一个已有项目，并等待 MCP 服务真正开始监听：

支持 Agent Skills 的客户端可以直接使用仓库内的 [`skills/use-chying/SKILL.md`](../skills/use-chying/SKILL.md)，其中包含项目打开、就绪校验和冲突处理的完整约束。

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

也可以查询当前状态：

```bash
/Applications/ChYing.app/Contents/MacOS/chyingctl status --json
```

运行约束：

- 仅接受 `~/.config/ChYing/db/<项目>/<项目>.db` 中已经存在的项目。
- 同一项目重复调用是幂等的，只会唤起已有窗口并返回 MCP 地址。
- 桌面应用已经打开其他项目时会明确失败，不会在后台静默切换数据库。
- 启动和就绪状态原子写入 `~/.config/ChYing/runtime.json`，权限为 `0600`，其中不保存密钥或流量内容。
- 本地开发包可用 `bin/ChYing.dev.app/Contents/MacOS/chyingctl`；也可通过 `--app` 或 `CHYING_APP` 指定应用路径。

不需要等待结果时可传 `--wait-mcp=false`。默认超时为 60 秒，可通过 `--timeout 90s` 调整。
