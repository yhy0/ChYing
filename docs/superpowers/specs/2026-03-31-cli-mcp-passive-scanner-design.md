# ChYing CLI + MCP 被动扫描服务设计

## 概述

将 ChYing 的被动扫描能力从 GUI 桌面应用中解耦，提供一个独立的 CLI 入口。CLI 模式在 Docker 中以服务形式运行，通过 MCP (SSE) 接口暴露扫描结果，供 Agent 或人工查询。

**核心定位：** 人机两用 — Agent 通过 MCP 调用获取结果，人也可以直接用 CLI 做被动扫描。

## 架构

```
┌────────────────────────────────────────────────────────────┐
│ Docker Container                                           │
│                                                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ chying-cli serve                                    │   │
│  │                                                     │   │
│  │  ┌──────────┐    ┌────────────────┐   ┌──────────┐ │   │
│  │  │ Proxify  │───>│ PassiveScan    │──>│ Jie Scan │ │   │
│  │  │ :9080    │    │ (async)        │   │ Plugins  │ │   │
│  │  └──────────┘    └────────────────┘   └────┬─────┘ │   │
│  │       │                                     │       │   │
│  │       │ session_id 标记                      │       │   │
│  │       ▼                                     ▼       │   │
│  │  ┌──────────┐                         ┌──────────┐  │   │
│  │  │ SQLite   │◄────────────────────────│ output   │  │   │
│  │  │ (GORM)   │                         │ Channel  │  │   │
│  │  └──────────┘                         └────┬─────┘  │   │
│  │       ▲                                    │        │   │
│  │       │                                    ▼        │   │
│  │  ┌──────────┐                         ┌──────────┐  │   │
│  │  │ MCP SSE  │                         │ CLI      │  │   │
│  │  │ :9090    │                         │ Output   │  │   │
│  │  └──────────┘                         └──────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────┘
      ▲ proxy          ▲ MCP SSE
      │                │
┌─────┴────┐    ┌──────┴───┐
│ Agent    │    │ Agent    │
│ Browser  │    │ MCP      │
│ / HTTP   │    │ Client   │
└──────────┘    └──────────┘
```

**设计原则：**
- GUI 入口 (`main.go`) 和前端完全不动
- 所有修改对 GUI 模式向后兼容 — `session_id` 为空时等同于现有行为
- CLI 直接复用现有包：`mitmproxy/`、`pkg/Jie/`、`pkg/db/`、`pkg/mcpserver/`

## CLI 入口

### 文件结构

```
cmd/chying-cli/
├── main.go       # cobra 根命令
├── serve.go      # serve 子命令：初始化 + 启动全部服务
└── output.go     # CLI 终端输出格式化（漏洞、流量摘要）
```

### 启动参数

```bash
chying-cli serve \
  --proxy-port 9080        # 代理监听端口（默认 9080）
  --mcp-port 9090          # MCP SSE 端口（默认 9090）
  --bind 0.0.0.0           # 监听地址（默认 127.0.0.1，Docker 场景用 0.0.0.0）
  --project default        # 项目名，对应 SQLite 数据库名
  --config /path/to.yaml   # 指定配置文件路径（可选）
  --quiet                  # 静默模式，不在终端输出流量和漏洞
```

### 初始化顺序

从现有 `App.StepXxx` 方法中抽取的核心逻辑，跳过 Wails 依赖部分：

```
1.  logging.New(...)              → 日志初始化
2.  file.New()                    → 文件系统路径（~/.config/ChYing/）
3.  ants.NewPool(conf.Parallelism)→ 协程池
4.  wappalyzer.New()              → 指纹识别引擎
5.  conf.InitConfig()             → 加载/创建配置
6.  conf.HotConf()                → 配置热重载
7.  conf.SyncJieConfig()          → 同步 Jie 扫描器配置
8.  db.Init(project, "sqlite")    → 数据库初始化 + schema 迁移
9.  mode.Passive()                → 设为被动扫描模式
10. go mitmproxy.Proxify()        → 启动代理 + 被动扫描插件
11. mcpserver.StartHTTPServer()   → 启动 MCP SSE 服务
12. go cliEventLoop()             → 监听 output.OutChannel，输出漏洞到终端
13. signal.Notify(SIGINT, SIGTERM)→ 优雅退出
```

**与 GUI 模式的差异：**
- 不启动 Wails 应用
- 不调用 `wailsApp.Event.Emit()` — CLI 模式用 `cliEventLoop` 替代
- MCP/Proxy 监听地址可配置为 `0.0.0.0`（GUI 模式保持 `127.0.0.1`）

### CLI 终端输出

非 quiet 模式下实时打印：

```
[14:23:01] 🔍 Proxy listening on 0.0.0.0:9080
[14:23:01] 🔌 MCP server on 0.0.0.0:9090/mcp
[14:23:15] → GET https://target.com/api/users [200] 1.2KB
[14:23:18] 🔴 [HIGH] SQL Injection | https://target.com/api/users?id=1 | Param: id
[14:24:01] 🟡 [MEDIUM] Sensitive info | https://target.com/api/config | AWS_ACCESS_KEY_ID
```

`--quiet` 模式下只输出启动信息和错误。

## Session 隔离机制

支持多个 Agent 并发共用同一个 ChYing 实例，通过 session 隔离各自的流量和漏洞结果。

### 数据模型

```go
// pkg/mcpserver/session.go
type ScanSession struct {
    SessionID   string    // UUID，自动生成
    Targets     []string  // 目标域名列表，如 ["target-a.com", "*.target-a.com"]
    Description string    // 可选描述
    CreatedAt   time.Time
    Active      bool      // 是否活跃
}
```

Session 信息存储在内存中（`sync.Map`），不需要持久化 — 重启后 session 清空，Agent 需要重新注册。

### 流程

```
1. Agent 调用 register_session(targets=["target-a.com"]) → 返回 session_id

2. Agent 配置代理时附带 session 标识：
   HTTP 请求头: X-ChYing-Session: <session_id>

3. ChYing 代理 onRequestCallback 中：
   - 读取 X-ChYing-Session header
   - 从 header 中剥离（不转发给目标服务器）
   - 关联到 HTTPHistory 记录的 session_id 字段

4. 被动扫描发现漏洞时：
   - 漏洞记录继承流量的 session_id

5. Agent 查询时：
   get_vulnerabilities(session_id="abc-123")         → 只返回该 session 的漏洞
   get_new_findings_since(session_id="abc-123", ...) → 增量查询
```

### 数据库改动

为以下表增加 `session_id` 字段（可为空，GUI 模式不填）：

- `http_histories` 表：`session_id VARCHAR(36) DEFAULT ''`
- `vulnerabilities` 表：`session_id VARCHAR(36) DEFAULT ''`

所有查询方法增加可选的 `session_id` 参数。不传时返回全量数据（向后兼容 GUI 模式）。

## MCP Tools

### 复用现有 Tools（11 个，全部保留）

| Tool | 说明 |
|------|------|
| `get_http_history` | 查流量列表（分页） |
| `get_traffic_detail` | 查完整请求/响应 |
| `query_by_dsl` | DSL 高级查询 |
| `get_hosts` | 获取所有域名 |
| `get_traffic_by_host` | 按域名查流量 |
| `get_vulnerabilities` | 获取漏洞列表 |
| `get_statistics` | 获取统计信息 |
| `send_request` | 发送请求（Repeater） |
| `run_intruder` | 运行 Intruder |
| `get_current_project` | 获取当前项目 |

以上 tools 均增加可选参数 `session_id`，传入时按 session 过滤。

### 新增 Tools

#### `register_session`

注册扫描会话，返回 session_id。

```
参数:
  targets: string[]  (必填) 目标域名列表
  description: string (可选) 会话描述

返回:
  { session_id: "uuid", targets: [...], created_at: "..." }
```

#### `get_scan_status`

查询扫描状态。

```
参数:
  session_id: string (可选) 指定 session

返回:
  {
    proxy_running: true,
    total_requests: 1234,
    total_vulnerabilities: 5,
    scan_queue_depth: 3,
    active_sessions: 2,
    uptime: "2h30m"
  }
```

#### `get_new_findings_since`

增量查询，返回指定时间之后的新发现。Agent 不需要轮询全量数据。

```
参数:
  session_id: string  (可选) 指定 session
  since: string       (必填) ISO 8601 时间戳
  type: string        (可选) "vulnerabilities" | "traffic" | "all"

返回:
  {
    vulnerabilities: [...],
    new_hosts: [...],
    traffic_count: 42,
    query_time: "2026-03-31T14:30:00Z"  // 用于下次增量查询
  }
```

#### `configure_session`

动态修改 session 配置。

```
参数:
  session_id: string  (必填)
  add_targets: string[] (可选) 追加目标域名
  remove_targets: string[] (可选) 移除目标域名

返回:
  { session_id: "...", targets: [...updated...] }
```

#### `close_session`

关闭并清理 session。

```
参数:
  session_id: string (必填)

返回:
  { closed: true, summary: { total_requests: N, total_vulns: M } }
```

## Agent 集成方式

### 浏览器场景

```python
# Playwright
browser = playwright.chromium.launch(
    proxy={"server": "http://chying:9080"}
)
# 页面请求自动经过 ChYing 被动扫描
```

Session header 通过 Playwright 的 `extra_http_headers` 或 proxy 配置注入。

### 非浏览器 HTTP 场景

```bash
# 环境变量设代理
export HTTP_PROXY=http://chying:9080
export HTTPS_PROXY=http://chying:9080

# 之后所有 HTTP 工具自动走代理
curl http://target/api/users
python -c "import requests; requests.get('http://target')"
sqlmap -u "http://target?id=1"
```

### 非 HTTP 场景（PWN/Crypto）

不走代理，直连目标。被动扫描器不参与。

## Docker 部署

### Dockerfile.cli

```dockerfile
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o chying-cli ./cmd/chying-cli/

FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite-libs
COPY --from=builder /app/chying-cli /usr/local/bin/
VOLUME /data
EXPOSE 9080 9090
ENTRYPOINT ["chying-cli", "serve", "--bind", "0.0.0.0"]
```

### docker-compose.cli.yml

```yaml
services:
  chying:
    build:
      context: .
      dockerfile: Dockerfile.cli
    ports:
      - "9080:9080"
      - "9090:9090"
    volumes:
      - chying-data:/data
    command: ["serve", "--proxy-port", "9080", "--mcp-port", "9090", "--bind", "0.0.0.0"]

  # Agent 使用示例
  agent:
    image: your-agent:latest
    environment:
      - HTTP_PROXY=http://chying:9080
      - HTTPS_PROXY=http://chying:9080
      - CHYING_MCP_URL=http://chying:9090/mcp
    depends_on:
      - chying

volumes:
  chying-data:
```

## 改动文件清单

| 操作 | 文件 | 说明 |
|------|------|------|
| 新增 | `cmd/chying-cli/main.go` | CLI 入口，cobra 根命令 |
| 新增 | `cmd/chying-cli/serve.go` | serve 子命令，初始化 + 启动 |
| 新增 | `cmd/chying-cli/output.go` | 终端输出格式化 |
| 新增 | `pkg/mcpserver/session.go` | Session 管理 |
| 新增 | `pkg/mcpserver/tools_session.go` | Session 相关 MCP tools |
| 新增 | `pkg/mcpserver/tools_realtime.go` | `get_scan_status`、`get_new_findings_since` |
| 新增 | `Dockerfile.cli` | CLI Docker 构建 |
| 新增 | `docker-compose.cli.yml` | CLI 部署配置 |
| 修改 | `pkg/mcpserver/server.go` | 注册新 tools；`StartHTTPServer` 支持 bind addr |
| 修改 | `pkg/db/db.go` | 表增加 `session_id` 字段 |
| 修改 | `pkg/db/` 查询方法 | 增加 `session_id` 可选过滤 |
| 修改 | `mitmproxy/proxify.go` | `onRequestCallback` 读取 `X-ChYing-Session` header |
| 修改 | `mitmproxy/passiveScanPlugin.go` | 传递 session_id 到扫描任务 |
| 不动 | `main.go` | GUI 入口 |
| 不动 | `app_*.go` | Wails 绑定方法 |
| 不动 | `frontend/` | 前端 |

## 测试策略

1. **单元测试：** Session 管理的 CRUD、DB 查询的 session 过滤
2. **集成测试：** CLI 启动 → 代理转发 → 被动扫描触发 → MCP 查询结果
3. **手动验证：** Docker compose 起服务 → 浏览器设代理访问目标 → MCP 客户端查结果
4. **兼容性验证：** GUI 模式启动，确认所有功能不受影响（session_id 为空的向后兼容）
