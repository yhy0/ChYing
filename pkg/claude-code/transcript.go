package claudecode

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2026/01/25
   @desc Claude CLI Transcript 文件解析
**/

// TranscriptEntry 表示 jsonl 文件中的一行
type TranscriptEntry struct {
	Type      string                 `json:"type"`      // "user", "assistant", "result", "system"
	Message   *TranscriptMessage     `json:"message"`   // 消息内容
	SessionID string                 `json:"sessionId"` // 会话 ID
	UUID      string                 `json:"uuid"`      // 消息 UUID
	Timestamp string                 `json:"timestamp"` // 时间戳
	Extra     map[string]interface{} `json:"-"`         // 其他字段
}

// TranscriptMessage 表示消息内容
type TranscriptMessage struct {
	Role    string                   `json:"role"`    // "user", "assistant", "system"
	Content interface{}              `json:"content"` // 可以是 string 或 []ContentBlock
	Model   string                   `json:"model,omitempty"`
	Usage   map[string]interface{}   `json:"usage,omitempty"`
}

// ContentBlock 表示内容块
type ContentBlock struct {
	Type  string                 `json:"type"`  // "text", "tool_use", "tool_result"
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

// ComputeTranscriptPath 计算 transcript 文件路径
// cwd: 工作目录
// sessionID: Claude CLI 会话 ID
// 注意: Claude CLI 将路径中的 "/" 和 "." 都替换为 "-"
func ComputeTranscriptPath(cwd, sessionID string) string {
	if cwd == "" || sessionID == "" {
		return ""
	}
	// Claude CLI 的编码方式：将 "/" 和 "." 都替换为 "-"
	encodedCwd := strings.ReplaceAll(cwd, "/", "-")
	encodedCwd = strings.ReplaceAll(encodedCwd, ".", "-")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".claude", "projects", encodedCwd, sessionID+".jsonl")
}

// ReadTranscript 读取并解析 transcript 文件
func ReadTranscript(transcriptPath string) ([]ChatMessage, error) {
	if transcriptPath == "" {
		return nil, fmt.Errorf("transcript path is empty")
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ChatMessage{}, nil
		}
		return nil, fmt.Errorf("failed to open transcript file: %w", err)
	}
	defer file.Close()

	var messages []ChatMessage
	scanner := bufio.NewScanner(file)
	// 增加 buffer 大小以处理大消息
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024) // 最大 10MB 每行

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			logging.Logger.Warnf("Failed to parse transcript line %d: %v", lineNum, err)
			continue
		}

		// 跳过 result 类型的条目
		if entry.Type == "result" {
			continue
		}

		// 转换为 ChatMessage
		chatMsg := convertTranscriptEntryToMessage(entry, lineNum)
		if chatMsg != nil {
			messages = append(messages, *chatMsg)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading transcript file: %w", err)
	}

	return messages, nil
}

// convertTranscriptEntryToMessage 将 TranscriptEntry 转换为 ChatMessage
func convertTranscriptEntryToMessage(entry TranscriptEntry, lineNum int) *ChatMessage {
	if entry.Message == nil {
		return nil
	}

	msg := &ChatMessage{
		ID:        entry.UUID,
		Role:      entry.Message.Role,
		Timestamp: parseTimestamp(entry.Timestamp),
	}

	// 处理 content
	switch content := entry.Message.Content.(type) {
	case string:
		msg.Content = content
	case []interface{}:
		// 内容是数组，需要解析每个块
		var textParts []string
		var toolUses []ToolUse

		for _, block := range content {
			blockMap, ok := block.(map[string]interface{})
			if !ok {
				continue
			}

			blockType, _ := blockMap["type"].(string)
			switch blockType {
			case "text":
				if text, ok := blockMap["text"].(string); ok {
					textParts = append(textParts, text)
				}
			case "tool_use":
				toolUse := ToolUse{
					ID:     getString(blockMap, "id"),
					Name:   getString(blockMap, "name"),
					Status: "completed",
				}
				if input, ok := blockMap["input"].(map[string]interface{}); ok {
					inputJSON, _ := json.Marshal(input)
					toolUse.Input = inputJSON
				}
				toolUses = append(toolUses, toolUse)
			case "tool_result":
				// 工具结果通常作为单独的消息处理
				toolUseID := getString(blockMap, "tool_use_id")
				if toolUseID != "" {
					// 查找对应的 tool_use 并更新结果
					for i := range toolUses {
						if toolUses[i].ID == toolUseID {
							if result, ok := blockMap["content"].(string); ok {
								toolUses[i].Result = result
							}
							break
						}
					}
				}
			}
		}

		msg.Content = strings.Join(textParts, "\n")
		if len(toolUses) > 0 {
			msg.ToolUses = toolUses
		}
	}

	// 如果消息内容为空且没有工具调用，跳过
	if msg.Content == "" && len(msg.ToolUses) == 0 {
		return nil
	}

	return msg
}

// parseTimestamp 解析时间戳
func parseTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Now()
	}

	// 尝试多种格式
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, ts); err == nil {
			return t
		}
	}

	return time.Now()
}

// getString 从 map 中安全获取字符串
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// GetTranscriptHistory 获取 transcript 历史（供外部调用）
// claudeSessionID: Claude CLI 的会话 ID
// workingDir: 工作目录
func GetTranscriptHistory(claudeSessionID, workingDir string) ([]ChatMessage, error) {
	if claudeSessionID == "" {
		return nil, fmt.Errorf("claude session ID is empty")
	}

	transcriptPath := ComputeTranscriptPath(workingDir, claudeSessionID)
	if transcriptPath == "" {
		return nil, fmt.Errorf("failed to compute transcript path")
	}

	return ReadTranscript(transcriptPath)
}
