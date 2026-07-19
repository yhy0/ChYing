package mcpserver

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/logging"
)

// NewChYingMCPServer 创建并注册所有 MCP tools
func NewChYingMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		"ChYing Security Platform",
		conf.Version,
		server.WithToolCapabilities(true),
	)

	// 查询类工具
	s.AddTool(getHttpHistoryTool(), handleGetHttpHistory)
	s.AddTool(getTrafficDetailTool(), handleGetTrafficDetail)
	s.AddTool(queryByDSLTool(), handleQueryByDSL)
	s.AddTool(getHostsTool(), handleGetHosts)
	s.AddTool(getTrafficByHostTool(), handleGetTrafficByHost)
	s.AddTool(getVulnerabilitiesTool(), handleGetVulnerabilities)
	s.AddTool(getStatisticsTool(), handleGetStatistics)

	// 主动测试类工具
	s.AddTool(sendRequestTool(), handleSendRequest)
	s.AddTool(runIntruderTool(), handleRunIntruder)

	// 工具类
	s.AddTool(getCurrentProjectTool(), handleGetCurrentProject)

	// Session 管理工具
	s.AddTool(registerSessionTool(), handleRegisterSession)
	s.AddTool(configureSessionTool(), handleConfigureSession)
	s.AddTool(closeSessionTool(), handleCloseSession)

	// 实时状态工具
	s.AddTool(getScanStatusTool(), handleGetScanStatus)
	s.AddTool(getNewFindingsSinceTool(), handleGetNewFindingsSince)

	return s
}

// StartHTTPServer 启动 MCP HTTP SSE Server，仅绑定到 localhost
// 返回实际监听地址（可能因端口冲突而 fallback 至 port+1..port+9），如果启动失败返回错误。
//
// 行为：从 `port` 开始最多尝试 10 个连续端口；遇到 "address already in use" 自动递增重试。
// 这样 GUI 与 CLI 默认端口（9090）即便被其他进程占用，也能在 9091..9099 范围内找到空位。
// 真正监听的端口通过返回值和日志（"MCP server listening on ..."）暴露给上层，
// 上层会通过 wails 事件 MCPStarted 把地址告诉前端，外部 MCP 客户端可据此更新连接 URL。
func StartHTTPServer(port int, bindAddr ...string) (string, error) {
	s := NewChYingMCPServer()
	httpServer := server.NewStreamableHTTPServer(s)

	host := "127.0.0.1"
	if len(bindAddr) > 0 && bindAddr[0] != "" {
		host = bindAddr[0]
	}

	mux := http.NewServeMux()
	mux.Handle("/mcp", httpServer)

	const maxAttempts = 10

	var (
		listener net.Listener
		addr     string
		lastErr  error
	)
	for i := 0; i < maxAttempts; i++ {
		tryPort := port + i
		tryAddr := fmt.Sprintf("%s:%d", host, tryPort)
		ln, err := net.Listen("tcp", tryAddr)
		if err == nil {
			listener = ln
			addr = tryAddr
			if i > 0 {
				logging.Logger.Warnf("MCP server: 端口 %d 被占用，自动 fallback 到 %d", port, tryPort)
			}
			break
		}
		lastErr = err
		// 仅在 "address already in use" 类错误时继续重试；其他错误（如权限、地址非法）应立即报错。
		if !isAddrInUseErr(err) {
			break
		}
	}
	if listener == nil {
		logging.Logger.Errorf("MCP server listen error: %v (尝试了端口 %d..%d)", lastErr, port, port+maxAttempts-1)
		return "", fmt.Errorf("MCP server listen error: %w (尝试了端口 %d..%d)", lastErr, port, port+maxAttempts-1)
	}
	logging.Logger.Infof("MCP server listening on %s", addr)

	// 启动成功，在新 goroutine 中 serve
	go func() {
		if err := http.Serve(listener, mux); err != nil {
			logging.Logger.Errorf("MCP server error: %v", err)
		}
	}()

	return addr, nil
}

// isAddrInUseErr 判断错误是否为 "address already in use"。
// 比 strings.Contains 更稳：跨 macOS/Linux/Windows 都能用 syscall.EADDRINUSE 命中。
func isAddrInUseErr(err error) bool {
	if err == nil {
		return false
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var sysErr *os.SyscallError
		if errors.As(opErr.Err, &sysErr) {
			return errors.Is(sysErr.Err, syscall.EADDRINUSE)
		}
	}
	return false
}
