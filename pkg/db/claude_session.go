package db

import (
	"encoding/json"
	"time"

	"github.com/yhy0/logging"
	"gorm.io/gorm"
)

/**
   @author yhy
   @since 2026/01/11
   @desc Claude AI 会话数据库模型
**/

// ClaudeSession Claude 会话模型
type ClaudeSession struct {
	ID             int64     `gorm:"primary_key;auto_increment" json:"id"`
	SessionID      string    `gorm:"index;unique;not null" json:"session_id"` // 会话唯一标识
	ProjectID      string    `gorm:"index;not null" json:"project_id"`        // 项目ID
	Title          string    `json:"title"`                                   // 会话标题（基于第一条消息）
	Context        string    `gorm:"type:text" json:"context"`                // 上下文 JSON
	History        string    `gorm:"type:text" json:"history"`                // 消息历史 JSON
	ConversationID string    `json:"conversation_id"`                         // Claude Code 会话 ID
	MessageCount   int       `gorm:"default:0" json:"message_count"`          // 消息数量
	TotalCostUSD   float64   `gorm:"default:0" json:"total_cost_usd"`         // 总费用
	TotalTokens    int       `gorm:"default:0" json:"total_tokens"`           // 总 token 数
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ClaudeSession) TableName() string {
	return "claude_sessions"
}

// ClaudeSessionMessage 会话消息（用于 JSON 序列化）
type ClaudeSessionMessage struct {
	ID           string                 `json:"id"`
	Role         string                 `json:"role"`
	Content      string                 `json:"content"`
	Timestamp    time.Time              `json:"timestamp"`
	ToolUses     []ClaudeSessionToolUse `json:"tool_uses,omitempty"`
	CostUSD      float64                `json:"cost_usd,omitempty"`
	InputTokens  int                    `json:"input_tokens,omitempty"`
	OutputTokens int                    `json:"output_tokens,omitempty"`
}

// ClaudeSessionToolUse 工具调用（用于 JSON 序列化）
type ClaudeSessionToolUse struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Input  json.RawMessage `json:"input"`
	Status string          `json:"status"`
	Result string          `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// ClaudeSessionContext 会话上下文（用于 JSON 序列化）
type ClaudeSessionContext struct {
	ProjectID            string   `json:"project_id"`
	ProjectName          string   `json:"project_name"`
	SelectedTrafficIDs   []string `json:"selected_traffic_ids,omitempty"`
	SelectedVulnIDs      []string `json:"selected_vuln_ids,omitempty"`
	SelectedFingerprints []string `json:"selected_fingerprints,omitempty"`
	CustomData           string   `json:"custom_data,omitempty"`
	AutoCollect          bool     `json:"auto_collect"`
}

// AddClaudeSession 添加会话
func AddClaudeSession(session *ClaudeSession) error {
	if GlobalDB == nil {
		return nil
	}
	return RetryOnLocked("AddClaudeSession", func() error {
		return GlobalDB.Create(session).Error
	}, 3)
}

// UpdateClaudeSession 更新会话
func UpdateClaudeSession(session *ClaudeSession) error {
	if GlobalDB == nil {
		return nil
	}
	return RetryOnLocked("UpdateClaudeSession", func() error {
		// 使用 session_id 作为条件更新，避免主键为 0 时 Save 执行 INSERT 导致唯一约束冲突
		return GlobalDB.Model(&ClaudeSession{}).
			Where("session_id = ?", session.SessionID).
			Updates(map[string]interface{}{
				"project_id":      session.ProjectID,
				"title":           session.Title,
				"context":         session.Context,
				"history":         session.History,
				"conversation_id": session.ConversationID,
				"message_count":   session.MessageCount,
				"total_cost_usd":  session.TotalCostUSD,
				"total_tokens":    session.TotalTokens,
				"updated_at":      time.Now(),
			}).Error
	}, 3)
}

// GetClaudeSession 根据 SessionID 获取会话
func GetClaudeSession(sessionID string) (*ClaudeSession, error) {
	if GlobalDB == nil {
		return nil, nil
	}
	var session ClaudeSession
	err := GlobalDB.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetClaudeSessionsByProject 获取项目的所有会话
func GetClaudeSessionsByProject(projectID string) ([]*ClaudeSession, error) {
	if GlobalDB == nil {
		return nil, nil
	}
	var sessions []*ClaudeSession
	err := GlobalDB.Where("project_id = ?", projectID).
		Order("updated_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetAllClaudeSessions 获取所有会话
func GetAllClaudeSessions() ([]*ClaudeSession, error) {
	if GlobalDB == nil {
		return nil, nil
	}
	var sessions []*ClaudeSession
	err := GlobalDB.Order("updated_at DESC").Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// DeleteClaudeSession 删除会话
func DeleteClaudeSession(sessionID string) error {
	if GlobalDB == nil {
		return nil
	}
	return RetryOnLocked("DeleteClaudeSession", func() error {
		return GlobalDB.Where("session_id = ?", sessionID).Delete(&ClaudeSession{}).Error
	}, 3)
}

// ClearClaudeSession 清除会话消息但保留会话
func ClearClaudeSession(sessionID string) error {
	if GlobalDB == nil {
		return nil
	}
	return RetryOnLocked("ClearClaudeSession", func() error {
		return GlobalDB.Model(&ClaudeSession{}).
			Where("session_id = ?", sessionID).
			Updates(map[string]interface{}{
				"history":         "[]",
				"conversation_id": "",
				"message_count":   0,
				"updated_at":      time.Now(),
			}).Error
	}, 3)
}

// UpdateClaudeSessionHistory 更新会话历史
func UpdateClaudeSessionHistory(sessionID string, history []ClaudeSessionMessage) error {
	if GlobalDB == nil {
		return nil
	}
	historyJSON, err := json.Marshal(history)
	if err != nil {
		logging.Logger.Errorf("序列化会话历史失败: %v", err)
		return err
	}

	// 生成标题（基于第一条用户消息）
	title := ""
	for _, msg := range history {
		if msg.Role == "user" {
			content := msg.Content
			// 使用 rune 切片来安全截取 UTF-8 字符串
			runes := []rune(content)
			if len(runes) > 50 {
				title = string(runes[:50]) + "..."
			} else {
				title = content
			}
			break
		}
	}

	return RetryOnLocked("UpdateClaudeSessionHistory", func() error {
		updates := map[string]interface{}{
			"history":       string(historyJSON),
			"message_count": len(history),
			"updated_at":    time.Now(),
		}
		if title != "" {
			updates["title"] = title
		}
		return GlobalDB.Model(&ClaudeSession{}).
			Where("session_id = ?", sessionID).
			Updates(updates).Error
	}, 3)
}

// UpdateClaudeSessionContext 更新会话上下文
func UpdateClaudeSessionContext(sessionID string, context *ClaudeSessionContext) error {
	if GlobalDB == nil {
		return nil
	}
	contextJSON, err := json.Marshal(context)
	if err != nil {
		logging.Logger.Errorf("序列化会话上下文失败: %v", err)
		return err
	}

	return RetryOnLocked("UpdateClaudeSessionContext", func() error {
		return GlobalDB.Model(&ClaudeSession{}).
			Where("session_id = ?", sessionID).
			Updates(map[string]interface{}{
				"context":    string(contextJSON),
				"updated_at": time.Now(),
			}).Error
	}, 3)
}

// UpdateClaudeSessionConversationID 更新 Claude Code 会话 ID
func UpdateClaudeSessionConversationID(sessionID, conversationID string) error {
	if GlobalDB == nil {
		return nil
	}
	return RetryOnLocked("UpdateClaudeSessionConversationID", func() error {
		return GlobalDB.Model(&ClaudeSession{}).
			Where("session_id = ?", sessionID).
			Updates(map[string]interface{}{
				"conversation_id": conversationID,
				"updated_at":      time.Now(),
			}).Error
	}, 3)
}

// UpdateClaudeSessionCost 更新会话费用统计
func UpdateClaudeSessionCost(sessionID string, costUSD float64, tokens int) error {
	if GlobalDB == nil {
		return nil
	}
	return RetryOnLocked("UpdateClaudeSessionCost", func() error {
		return GlobalDB.Model(&ClaudeSession{}).
			Where("session_id = ?", sessionID).
			Updates(map[string]interface{}{
				"total_cost_usd": gorm.Expr("total_cost_usd + ?", costUSD),
				"total_tokens":   gorm.Expr("total_tokens + ?", tokens),
				"updated_at":     time.Now(),
			}).Error
	}, 3)
}

// ExistClaudeSession 检查会话是否存在
func ExistClaudeSession(sessionID string) bool {
	if GlobalDB == nil {
		return false
	}
	var count int64
	GlobalDB.Model(&ClaudeSession{}).Where("session_id = ?", sessionID).Count(&count)
	return count > 0
}

// GetClaudeSessionProjects 获取所有唯一的项目ID列表
func GetClaudeSessionProjects() ([]string, error) {
	if GlobalDB == nil {
		return nil, nil
	}
	var projectIDs []string
	err := GlobalDB.Model(&ClaudeSession{}).
		Distinct("project_id").
		Pluck("project_id", &projectIDs).Error
	if err != nil {
		return nil, err
	}
	return projectIDs, nil
}

// CleanupOldClaudeSessions 清理过期会话
func CleanupOldClaudeSessions(maxAge time.Duration) (int64, error) {
	if GlobalDB == nil {
		return 0, nil
	}
	cutoff := time.Now().Add(-maxAge)
	var count int64

	err := RetryOnLocked("CleanupOldClaudeSessions", func() error {
		result := GlobalDB.Where("updated_at < ?", cutoff).Delete(&ClaudeSession{})
		count = result.RowsAffected
		return result.Error
	}, 3)

	return count, err
}
