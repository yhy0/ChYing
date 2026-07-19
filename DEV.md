## 安装 wails3

注意 wails 应该和 ChYing 目录在一级，也就是 ls
...
ChYing
wails
...

go.mod 中是这样写的
replace github.com/wailsapp/wails/v3 => ../wails/v3


```azure
git clone https://github.com/wailsapp/wails.git
cd wails
git checkout v3-alpha
cd v3/cmd/wails3
go install
```

如果执行 wails3 失败，则需要看 go 的 bin 目录是否已经加入到环境变量 GOPATH
还需要 npm 环境

## Windows 本地开发环境要求

本项目依赖 CGO（go-sqlite3、tree-sitter-javascript 等），Windows 本地开发需要安装 MinGW：

### 安装 MinGW（推荐使用 MSYS2）

1. 下载并安装 MSYS2：https://www.msys2.org/
2. 打开 MSYS2 UCRT64 终端，执行：
   ```bash
   pacman -S mingw-w64-ucrt-x86_64-gcc
   ```
3. 将 MinGW bin 目录添加到系统 PATH 环境变量：
   ```
   C:\msys64\ucrt64\bin
   ```
4. 验证安装：
   ```bash
   gcc --version
   ```

### 或使用 Chocolatey 安装

```powershell
choco install mingw
```

### 确保 CGO 启用

编译时需要确保 `CGO_ENABLED=1`：
```bash
set CGO_ENABLED=1
wails3 task windows:build PRODUCTION=true
```

**注意**：如果编译时 `CGO_ENABLED=0`，会导致 sqlite 数据库无法工作，程序运行时会崩溃。

### PRODUCTION 参数说明

`PRODUCTION=true` 参数用于区分开发构建和生产构建：

| 参数 | 构建标志 | 用途 |
|------|---------|------|
| 不设置或 `PRODUCTION=false` | `-buildvcs=false -gcflags=all="-l"` | 开发调试，保留调试信息 |
| `PRODUCTION=true` | `-tags production -trimpath -buildvcs=false -ldflags="-w -s"` | 生产发布，优化体积，移除调试信息 |

**Windows 本地构建必须使用 `PRODUCTION=true`**，否则可能无法正常运行。

https://github.com/yhy0/ChYing-Inside
https://github.com/wailsapp/wails.git



运行会在 /Users/你的用户名/.config/ChYing 下生成一个文件夹

~/.config/ChYing/proxify_data/cacert.pem  双击安装证书，用于捕获 https 流量

debug
```bash
ulimit -c unlimited
export GOTRACEBACK=crash
go install github.com/go-delve/delve/cmd/dlv@latest
```

# 一次性多版本
wails3 task darwin:package:universal


package
```shell
 wails3 task package
 
 or 

 wails3 task darwin:package
 wails3 task windows:package
 wails3 task linux:package
```
mac 下编译 Windows 平台参考

```bash
brew install mingw-w64
CGO_ENABLED=1
wails3 task windows:package
```

https://icon-sets.iconify.design/icon-park-outline/?category=General

# 前端依赖更新
pnpm install -g npm-check-updates
下载后使用 

cd frontend

ncu -u

# MCP Server

ChYing 内置了一个 MCP (Model Context Protocol) 服务，允许 AI 助手直接与 ChYing 的安全测试能力交互。

## 配置

确保 ChYing 应用已启动，

```json
{
  "mcpServers": {
    "chying": {
      "type": "http",
      "url": "http://127.0.0.1:9090/mcp"
    }
  }
}
```

> **协议说明**：ChYing 实现的是 MCP 2025-03-26 spec 的 **Streamable HTTP** 传输（单 endpoint `/mcp`，由 `mcp-go` 的 `NewStreamableHTTPServer` 提供），不是更早的 SSE 双 endpoint 传输。所以 `"type"` 必须填 `"http"`（部分客户端写作 `"streamable-http"`），不能填 `"sse"` —— 后者会让客户端按废弃协议握手。

默认端口为 `9090`（与 chying-cli 默认值一致），可在 ChYing 配置文件 `~/.config/ChYing/config.yaml` 中通过 `mcp_port` 字段修改。如启动时该端口被占用，ChYing 会自动尝试 `mcp_port + 1 .. mcp_port + 9`，实际监听地址在启动日志和前端 `MCPStarted` 事件中给出。

## Tools

### 查询类

#### `get_http_history`

获取代理捕获的 HTTP 流量历史记录（分页）。

| 参数 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| source | string | 否 | 过滤来源：`local`（本地代理）、`remote`（远程节点）、`all`（默认） |
| limit | number | 否 | 返回最大条数，默认 50，最大 500 |
| offset | number | 否 | 跳过的记录数，默认 0 |

**用途**：浏览代理捕获的所有 HTTP 请求，了解目标应用的接口和流量全貌。

---

#### `get_traffic_detail`

获取单条流量的完整请求和响应原始数据。

| 参数 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| hid | number | 否 | 流量的 History ID（来自 `get_http_history` / `get_traffic_by_host` 的结果） |
| id | number | 否 | 流量的数据库 ID（来自 `query_by_dsl` 的结果） |

> `hid` 和 `id` 至少提供一个。

**用途**：查看某条请求的完整报文（请求头、请求体、响应头、响应体），用于分析具体接口的行为。

---

#### `query_by_dsl`

使用 DSL 表达式查询流量历史。

| 参数 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| dsl_query | string | 是 | DSL 查询表达式 |

**可用字段**：`id`, `url`, `path`, `method`, `host`, `status`, `length`, `content_type`, `timestamp`, `request`, `request_body`, `response`, `response_body`, `request_headers`, `response_headers`, `status_reason`

**可用函数/运算符**：`contains()`, `regex()`, `==`, `!=`, `&&`, `||`

**示例**：
```
status == 200 && contains(response_body, "admin")
contains(host, "api.example.com") && method == "POST"
regex(path, "/api/v[0-9]+/users")
status != 200 && contains(request_headers, "Authorization")
```

**用途**：精确筛选流量，例如查找包含敏感信息的响应、特定 API 路径的请求等。

---

#### `get_hosts`

获取流量历史中所有唯一的主机名列表。无参数。

**用途**：快速了解代理捕获到了哪些目标主机，用于确定测试范围。

---

#### `get_traffic_by_host`

按主机名筛选流量，默认排除静态资源。

| 参数 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| host | string | 是 | 主机名，如 `example.com` |
| exclude_extensions | string | 否 | 排除的文件扩展名（逗号分隔）。设为 `none` 不排除。默认排除 js/css/图片/字体/媒体 |

**用途**：聚焦分析某个目标主机的 API 流量，自动过滤掉静态资源干扰。

---

#### `get_vulnerabilities`

获取已发现的漏洞列表（分页）。

| 参数 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| source | string | 否 | 过滤来源：`local`、`remote`、`all`（默认） |
| limit | number | 否 | 返回最大条数，默认 100，最大 500 |
| offset | number | 否 | 跳过的记录数，默认 0 |

**返回字段**：漏洞 ID、类型、目标、主机、方法、路径、插件、严重等级、参数、Payload、描述、cURL 命令、请求/响应原文、发现时间。

**用途**：查看被动/主动扫描发现的漏洞，获取漏洞详情和复现命令。

---

#### `get_statistics`

获取当前项目的综合统计信息。无参数。

**返回内容**：项目名称、流量总数、主机数量、主机列表、漏洞统计（按等级和类型分布）。

**用途**：快速获取项目安全测试的全局视图。

---

### 主动测试类

#### `send_request`

发送原始 HTTP 请求（Repeater 功能）。

| 参数 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| target | string | 是 | 目标 URL，含协议（如 `https://example.com`） |
| raw_request | string | 是 | 原始 HTTP 请求文本（请求头 + 可选的请求体） |

**示例**：
```
GET /api/users HTTP/1.1
Host: example.com
Cookie: session=abc123
```

```
POST /api/login HTTP/1.1
Host: example.com
Content-Type: application/json

{"username":"admin","password":"test"}
```

**用途**：手动发送/修改请求来验证漏洞、测试接口行为。

---

#### `run_intruder`

运行 Intruder 攻击，使用不同 Payload 批量发送请求。

| 参数 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| target | string | 是 | 目标 URL，含协议 |
| raw_request | string | 是 | 原始请求，用 `§` 标记 Payload 位置（如 `GET /api/users/§1§`） |
| payloads | string | 是 | JSON 数组格式的 Payload 集。如 `[["admin","test","user"]]` 或 `[["admin","test"],["123","password"]]` |
| attack_type | string | 是 | 攻击类型：`sniper`、`battering-ram`、`pitchfork`、`cluster-bomb` |

**攻击类型说明**：
- **sniper**：单 Payload 集，逐个位置依次替换
- **battering-ram**：单 Payload 集，所有位置同时替换相同值
- **pitchfork**：多 Payload 集，各位置并行迭代
- **cluster-bomb**：多 Payload 集，测试所有组合

> 最大请求组合数为 1000，超出需拆分批次。

**用途**：批量测试参数值，用于爆破、模糊测试、权限绕过验证等场景。

---

### 工具类

#### `get_current_project`

获取当前项目信息。无参数。

**返回内容**：项目名称、代理地址/端口/启用状态、MCP 端口、版本号。

**用途**：确认当前工作环境和配置状态。
