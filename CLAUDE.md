# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

承影（ChYing）是一个基于 Wails v3 的桌面安全测试工具，类似轻量级 Burp Suite。后端 Go，前端 Vue 3 + TypeScript，数据库 SQLite。集成了 Jie 扫描器用于被动/主动漏洞检测。

## 开发环境要求

- Go 1.25+，CGO_ENABLED=1（go-sqlite3 依赖 CGO）
- Node.js + npm
- wails3 CLI（从 wails 仓库 v3-alpha 分支安装：`cd wails/v3/cmd/wails3 && go install`）
- wails 源码需与 ChYing 目录同级（go.mod 中 `replace github.com/wailsapp/wails/v3 => ../wails/v3`）
- Windows 需要 MinGW（CGO 编译依赖）

## 常用命令

```bash
# 开发模式（热重载）
wails3 dev -config ./build/config.yml -port 9245

# 构建（当前平台）
wails3 task build                    # 开发构建
wails3 task build DEV=true           # 带调试信息

# 打包
wails3 task package                  # 当前平台
wails3 task darwin:package           # macOS .app
wails3 task darwin:package:universal # macOS 通用二进制
wails3 task windows:package          # Windows（需 PRODUCTION=true）
wails3 task linux:package            # Linux

# 前端
cd frontend && npm install           # 安装依赖
cd frontend && npm run dev           # 前端独立开发
cd frontend && npm run build         # 生产构建

# 生成绑定（Go struct -> TypeScript）
wails3 generate bindings -clean=true -ts

# Go
go mod tidy
go test ./test/...                   # 运行测试
go test ./test/ -run TestXxx -v      # 运行单个测试
```

## 架构

### 根目录 Go 文件 —— App 层（Wails 绑定）

`App` struct 是 Wails 暴露给前端的核心服务，其方法按职责拆分：

- `main.go` — 入口，Wails 应用创建、窗口配置、事件转发循环 `EventNotification()`
- `app.go` — App struct 定义、类型定义（Result/InitStep/InitProgress 等）、全局变量、init()
- `app_initialization.go` — 分步初始化流程（7 步：基础组件 → 配置加载 → 数据库 → 表结构 → 代理启动 → 项目加载 → 完成）
- `app_proxy.go` — 代理作用域配置、DSL 查询、匹配替换规则、越权检测、监听器管理
- `app_config.go` — 配置管理（读取/更新/保存配置文件）
- `app_database.go` — 数据库和历史记录操作
- `app_scan.go` — 扫描目标管理
- `app_window.go` — 窗口管理（文件选择对话框等）
- `app_utils.go` — 工具方法（内存信息、Nmap 扫描等）
- `app_claude.go` — Claude Code CLI 集成和 A2A 协议支持
- `mitmproxy.go` — Repeater（重放器）、Intruder（入侵者）、流量历史查询
- `extension.go` — 编解码器（URL/Base64/Hex/Unicode/MD5）
- `gadgets.go` — JWT 解析/签名/爆破、Fuzz、API 预测

### 核心包

- `mitmproxy/` — HTTP/HTTPS 代理核心，基于 projectdiscovery/proxify 的本地修改版（`lib/proxify/`）。包含流量拦截（intercept.go）、Intruder 攻击（intruder.go）、匹配替换（matchreplace.go）、越权检测（authcheck.go）、被动扫描插件（passiveScanPlugin.go）、DSL 查询（dsl.go）
- `conf/` — 统一配置管理，使用 viper 热加载 YAML 配置。`conf.go` 热加载逻辑，`config_manager.go` 配置读写，`default.go` 默认配置，`type.go` 配置结构体定义
- `pkg/db/` — SQLite 数据库层（gorm），管理 HTTP 历史、请求/响应、扫描目标、漏洞等
- `pkg/Jie/` — 集成的 Jie 扫描引擎（漏洞检测：XSS/SQL注入/SSRF/命令执行等）
- `pkg/claude-code/` — Claude Code SDK 客户端封装（CLI 调用 + A2A 协议）
- `api/` — API 管理层，封装 Config/Proxy/Vulnerability 三类 API

### 前端

Vue 3 + TypeScript + UnoCSS，Glassmorphism（液态玻璃）UI 风格。

- `frontend/src/views/` — 页面视图（ProjectSelection、ClaudeAgent、ScanLog、Vulnerability）
- `frontend/src/components/` — 按功能模块组织：proxy、repeater、intruder、decoder、plugins、scan、claude、settings 等
- `frontend/src/store/` — Pinia 状态管理
- `frontend/src/composables/` — Vue 组合式函数
- `frontend/bindings/` — wails3 自动生成的 Go→TS 绑定（.gitignore 中）

### 本地修改的依赖

- `lib/proxify/` — 修改版 proxify（修复 Stop() 端口释放）
- `lib/jsluice/` — JS 分析库
- `lib/gowsdl/` — WSDL 解析
- `lib/webUnPack/` — Web 解包工具

## 关键约定

- 前后端通信通过 Wails 绑定（App struct 方法）和 Wails 事件（`wailsApp.Event.Emit`）
- 统一返回结构 `Result{Data, Error}` — 不使用类型别名（Wails v3 binding 生成器的限制）
- 配置文件位于 `~/.config/ChYing/`，HTTPS 证书在 `~/.config/ChYing/proxify_data/cacert.pem`
- 日志文件在 `~/.config/ChYing/` 目录下
- Windows 构建必须使用 `PRODUCTION=true`，否则 sqlite 崩溃
- go.mod 中有多个 replace 指令处理依赖冲突，修改依赖时需注意
