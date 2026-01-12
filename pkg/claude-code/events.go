package claudecode

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2026/01/10
   @desc Wails 事件处理
**/

// EventEmitter 事件发射器接口
type EventEmitter interface {
	Emit(eventName string, data interface{})
}

// WailsEventEmitter Wails 事件发射器
type WailsEventEmitter struct {
	app *application.App
}

// NewWailsEventEmitter 创建 Wails 事件发射器
func NewWailsEventEmitter(app *application.App) *WailsEventEmitter {
	return &WailsEventEmitter{app: app}
}

// Emit 发射事件
func (e *WailsEventEmitter) Emit(eventName string, data interface{}) {
	if e.app != nil {
		e.app.Event.Emit(eventName, data)
	}
}

// EventNames 事件名称常量
const (
	// Claude Code 事件
	EventClaudeText       = "claude:text"        // 文本内容
	EventClaudeToolUse    = "claude:tool_use"    // 工具调用开始
	EventClaudeToolResult = "claude:tool_result" // 工具调用结果
	EventClaudeError      = "claude:error"       // 错误
	EventClaudeDone       = "claude:done"        // 完成
	EventClaudeCost       = "claude:cost"        // 费用统计
	EventClaudeSession    = "claude:session"     // 会话信息
)

// StreamEventHandler 流式事件处理器
type StreamEventHandler struct {
	emitter   EventEmitter
	sessionID string
}

// NewStreamEventHandler 创建流式事件处理器
func NewStreamEventHandler(emitter EventEmitter, sessionID string) *StreamEventHandler {
	return &StreamEventHandler{
		emitter:   emitter,
		sessionID: sessionID,
	}
}

// HandleEvent 处理流式事件
func (h *StreamEventHandler) HandleEvent(event StreamEvent) {
	// 添加会话 ID
	event.SessionID = h.sessionID

	switch event.Type {
	case "text":
		h.emitter.Emit(EventClaudeText, event)
	case "tool_use":
		h.emitter.Emit(EventClaudeToolUse, event)
	case "tool_result":
		h.emitter.Emit(EventClaudeToolResult, event)
	case "error":
		h.emitter.Emit(EventClaudeError, event)
	case "done":
		h.emitter.Emit(EventClaudeDone, event)
	case "cost":
		h.emitter.Emit(EventClaudeCost, event)
	default:
		// 未知事件类型，使用通用事件
		h.emitter.Emit("claude:event", event)
	}
}

// CreateEventChannel 创建事件通道并启动处理协程
// 返回一个发送通道，调用者负责关闭
func CreateEventChannel(ctx context.Context, emitter EventEmitter, sessionID string) chan<- StreamEvent {
	eventChan := make(chan StreamEvent, 100)
	handler := NewStreamEventHandler(emitter, sessionID)

	go func() {
		for {
			select {
			case <-ctx.Done():
				// 上下文取消，排空通道后退出
				for {
					select {
					case _, ok := <-eventChan:
						if !ok {
							return
						}
					default:
						return
					}
				}
			case event, ok := <-eventChan:
				if !ok {
					return
				}
				handler.HandleEvent(event)
			}
		}
	}()

	return eventChan
}

// SendEventNonBlocking 非阻塞发送事件，带超时
func SendEventNonBlocking(eventChan chan<- StreamEvent, event StreamEvent, timeout time.Duration) bool {
	select {
	case eventChan <- event:
		return true
	case <-time.After(timeout):
		logging.Logger.Warnf("Event channel full, dropping event: %s", event.Type)
		return false
	}
}

// FrontendEvent 前端事件结构（用于类型安全）
type FrontendEvent struct {
	Type      string      `json:"type"`
	SessionID string      `json:"session_id"`
	Data      interface{} `json:"data"`
}

// TextEvent 文本事件
type TextEvent struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
}

// ToolUseEvent 工具调用事件
type ToolUseEvent struct {
	SessionID string  `json:"session_id"`
	ToolUse   ToolUse `json:"tool_use"`
}

// ToolResultEvent 工具结果事件
type ToolResultEvent struct {
	SessionID string `json:"session_id"`
	ToolID    string `json:"tool_id"`
	Result    string `json:"result"`
	Error     string `json:"error,omitempty"`
}

// ErrorEvent 错误事件
type ErrorEvent struct {
	SessionID string `json:"session_id"`
	Error     string `json:"error"`
}

// DoneEvent 完成事件
type DoneEvent struct {
	SessionID string `json:"session_id"`
}

// CostEvent 费用事件
type CostEvent struct {
	SessionID    string  `json:"session_id"`
	CostUSD      float64 `json:"cost_usd"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
}
