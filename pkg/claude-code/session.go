package claudecode

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2026/01/10
   @desc Claude Code 会话管理（支持数据库持久化）
**/

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session // 内存缓存
	mu       sync.RWMutex
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession 创建新会话
func (m *SessionManager) CreateSession(projectID string, context *AgentContext) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := uuid.New().String()
	now := time.Now()

	if context == nil {
		context = &AgentContext{
			ProjectID:   projectID,
			AutoCollect: true,
		}
	}

	session := &Session{
		ID:        sessionID,
		ProjectID: projectID,
		Context:   context,
		CreatedAt: now,
		UpdatedAt: now,
		History:   []ChatMessage{},
	}

	// 保存到内存缓存
	m.sessions[sessionID] = session

	// 持久化到数据库
	m.saveSessionToDB(session)

	return session
}

// GetSession 获取会话
func (m *SessionManager) GetSession(sessionID string) *Session {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if exists {
		return session
	}

	// 从数据库加载
	session = m.loadSessionFromDB(sessionID)
	if session != nil {
		m.mu.Lock()
		m.sessions[sessionID] = session
		m.mu.Unlock()
	}

	return session
}

// GetOrCreateSession 获取或创建会话
func (m *SessionManager) GetOrCreateSession(sessionID, projectID string) *Session {
	session := m.GetSession(sessionID)
	if session != nil {
		return session
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	session = &Session{
		ID:        sessionID,
		ProjectID: projectID,
		Context: &AgentContext{
			ProjectID:   projectID,
			AutoCollect: true,
		},
		CreatedAt: now,
		UpdatedAt: now,
		History:   []ChatMessage{},
	}

	m.sessions[sessionID] = session
	m.saveSessionToDB(session)

	return session
}

// UpdateSession 更新会话
func (m *SessionManager) UpdateSession(session *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session.UpdatedAt = time.Now()
	m.sessions[session.ID] = session

	// 持久化到数据库
	m.saveSessionToDB(session)
}

// DeleteSession 删除会话
func (m *SessionManager) DeleteSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)

	// 从数据库删除
	if err := db.DeleteClaudeSession(sessionID); err != nil {
		logging.Logger.Errorf("删除会话失败: %v", err)
	}
}

// ClearSession 清除会话消息但保留会话
func (m *SessionManager) ClearSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.History = []ChatMessage{}
		session.ConversationID = ""
		session.UpdatedAt = time.Now()
	}

	// 更新数据库
	if err := db.ClearClaudeSession(sessionID); err != nil {
		logging.Logger.Errorf("清除会话失败: %v", err)
	}
}

// ListSessions 列出所有会话
func (m *SessionManager) ListSessions() []*Session {
	// 优先从数据库获取完整列表
	dbSessions, err := db.GetAllClaudeSessions()
	if err != nil {
		logging.Logger.Errorf("从数据库获取会话列表失败: %v", err)
		// 回退到内存缓存
		m.mu.RLock()
		defer m.mu.RUnlock()
		sessions := make([]*Session, 0, len(m.sessions))
		for _, session := range m.sessions {
			sessions = append(sessions, session)
		}
		return sessions
	}

	sessions := make([]*Session, 0, len(dbSessions))
	// 批量更新内存缓存
	m.mu.Lock()
	for _, dbSession := range dbSessions {
		session := m.dbSessionToSession(dbSession)
		if session != nil {
			sessions = append(sessions, session)
			m.sessions[session.ID] = session
		}
	}
	m.mu.Unlock()

	return sessions
}

// ListSessionsByProject 列出项目的所有会话
func (m *SessionManager) ListSessionsByProject(projectID string) []*Session {
	// 优先从数据库获取
	dbSessions, err := db.GetClaudeSessionsByProject(projectID)
	if err != nil {
		logging.Logger.Errorf("从数据库获取项目会话列表失败: %v", err)
		// 回退到内存缓存
		m.mu.RLock()
		defer m.mu.RUnlock()
		sessions := make([]*Session, 0)
		for _, session := range m.sessions {
			if session.ProjectID == projectID {
				sessions = append(sessions, session)
			}
		}
		return sessions
	}

	sessions := make([]*Session, 0, len(dbSessions))
	// 批量更新内存缓存
	m.mu.Lock()
	for _, dbSession := range dbSessions {
		session := m.dbSessionToSession(dbSession)
		if session != nil {
			sessions = append(sessions, session)
			m.sessions[session.ID] = session
		}
	}
	m.mu.Unlock()

	return sessions
}

// ListProjects 列出所有有会话的项目
func (m *SessionManager) ListProjects() []string {
	// 优先从数据库获取
	projects, err := db.GetClaudeSessionProjects()
	if err != nil {
		logging.Logger.Errorf("从数据库获取项目列表失败: %v", err)
		// 回退到内存缓存
		m.mu.RLock()
		defer m.mu.RUnlock()
		projectSet := make(map[string]bool)
		for _, session := range m.sessions {
			projectSet[session.ProjectID] = true
		}
		result := make([]string, 0, len(projectSet))
		for projectID := range projectSet {
			result = append(result, projectID)
		}
		return result
	}
	return projects
}

// AddUserMessage 添加用户消息到会话
func (m *SessionManager) AddUserMessage(sessionID, content string) *ChatMessage {
	// 先尝试获取会话（会自动从数据库加载）
	session := m.GetSession(sessionID)
	if session == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 重新获取，因为 GetSession 可能已经更新了缓存
	session = m.sessions[sessionID]
	if session == nil {
		return nil
	}

	msgID := uuid.New().String()
	now := time.Now()

	chatMsg := ChatMessage{
		ID:        msgID,
		Role:      "user",
		Content:   content,
		Timestamp: now,
	}
	session.History = append(session.History, chatMsg)
	session.UpdatedAt = now

	// 持久化到数据库
	m.saveHistoryToDB(sessionID, session.History)

	return &chatMsg
}

// AddAssistantMessage 添加助手消息到会话
func (m *SessionManager) AddAssistantMessage(sessionID, content string, toolUses []ToolUse) *ChatMessage {
	// 先尝试获取会话（会自动从数据库加载）
	session := m.GetSession(sessionID)
	if session == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 重新获取，因为 GetSession 可能已经更新了缓存
	session = m.sessions[sessionID]
	if session == nil {
		return nil
	}

	msgID := uuid.New().String()
	now := time.Now()

	chatMsg := ChatMessage{
		ID:        msgID,
		Role:      "assistant",
		Content:   content,
		Timestamp: now,
		ToolUses:  toolUses,
	}
	session.History = append(session.History, chatMsg)
	session.UpdatedAt = now

	// 持久化到数据库
	m.saveHistoryToDB(sessionID, session.History)

	return &chatMsg
}

// UpdateContext 更新会话上下文
func (m *SessionManager) UpdateContext(sessionID string, context *AgentContext) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.Context = context
		session.UpdatedAt = time.Now()

		// 持久化到数据库
		dbContext := &db.ClaudeSessionContext{
			ProjectID:            context.ProjectID,
			ProjectName:          context.ProjectName,
			SelectedTrafficIDs:   context.SelectedTrafficIDs,
			SelectedVulnIDs:      context.SelectedVulnIDs,
			SelectedFingerprints: context.SelectedFingerprints,
			CustomData:           context.CustomData,
			AutoCollect:          context.AutoCollect,
		}
		if err := db.UpdateClaudeSessionContext(sessionID, dbContext); err != nil {
			logging.Logger.Errorf("更新会话上下文失败: %v", err)
		}
	}
}

// SetConversationID 设置 Claude Code 会话 ID
func (m *SessionManager) SetConversationID(sessionID, conversationID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.ConversationID = conversationID
		session.UpdatedAt = time.Now()

		// 持久化到数据库
		if err := db.UpdateClaudeSessionConversationID(sessionID, conversationID); err != nil {
			logging.Logger.Errorf("更新会话 ConversationID 失败: %v", err)
		}
	}
}

// CleanupOldSessions 清理过期会话
func (m *SessionManager) CleanupOldSessions(maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	count := 0

	// 清理内存缓存
	for id, session := range m.sessions {
		if session.UpdatedAt.Before(cutoff) {
			delete(m.sessions, id)
			count++
		}
	}

	// 清理数据库
	dbCount, err := db.CleanupOldClaudeSessions(maxAge)
	if err != nil {
		logging.Logger.Errorf("清理过期会话失败: %v", err)
	} else if dbCount > 0 {
		logging.Logger.Infof("清理了 %d 个过期会话", dbCount)
	}

	return count
}

// GetSessionCount 获取会话数量
func (m *SessionManager) GetSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// ==================== 数据库辅助方法 ====================

// saveSessionToDB 保存会话到数据库
func (m *SessionManager) saveSessionToDB(session *Session) {
	// 序列化上下文
	contextJSON := ""
	if session.Context != nil {
		if data, err := json.Marshal(session.Context); err == nil {
			contextJSON = string(data)
		}
	}

	// 序列化历史
	historyJSON := "[]"
	if len(session.History) > 0 {
		if data, err := json.Marshal(session.History); err == nil {
			historyJSON = string(data)
		}
	}

	// 生成标题
	title := ""
	for _, msg := range session.History {
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

	dbSession := &db.ClaudeSession{
		SessionID:      session.ID,
		ProjectID:      session.ProjectID,
		Title:          title,
		Context:        contextJSON,
		History:        historyJSON,
		ConversationID: session.ConversationID,
		MessageCount:   len(session.History),
		CreatedAt:      session.CreatedAt,
		UpdatedAt:      session.UpdatedAt,
	}

	// 检查是否存在
	if db.ExistClaudeSession(session.ID) {
		if err := db.UpdateClaudeSession(dbSession); err != nil {
			logging.Logger.Errorf("更新会话到数据库失败: %v", err)
		}
	} else {
		if err := db.AddClaudeSession(dbSession); err != nil {
			logging.Logger.Errorf("保存会话到数据库失败: %v", err)
		}
	}
}

// saveHistoryToDB 保存历史到数据库
func (m *SessionManager) saveHistoryToDB(sessionID string, history []ChatMessage) {
	// 转换为数据库格式
	dbHistory := make([]db.ClaudeSessionMessage, len(history))
	for i, msg := range history {
		toolUses := make([]db.ClaudeSessionToolUse, len(msg.ToolUses))
		for j, tu := range msg.ToolUses {
			toolUses[j] = db.ClaudeSessionToolUse{
				ID:     tu.ID,
				Name:   tu.Name,
				Input:  tu.Input,
				Status: tu.Status,
				Result: tu.Result,
				Error:  tu.Error,
			}
		}
		dbHistory[i] = db.ClaudeSessionMessage{
			ID:           msg.ID,
			Role:         msg.Role,
			Content:      msg.Content,
			Timestamp:    msg.Timestamp,
			ToolUses:     toolUses,
			CostUSD:      msg.CostUSD,
			InputTokens:  msg.InputTokens,
			OutputTokens: msg.OutputTokens,
		}
	}

	if err := db.UpdateClaudeSessionHistory(sessionID, dbHistory); err != nil {
		logging.Logger.Errorf("保存会话历史到数据库失败: %v", err)
	}
}

// loadSessionFromDB 从数据库加载会话
func (m *SessionManager) loadSessionFromDB(sessionID string) *Session {
	dbSession, err := db.GetClaudeSession(sessionID)
	if err != nil || dbSession == nil {
		return nil
	}

	return m.dbSessionToSession(dbSession)
}

// dbSessionToSession 将数据库会话转换为内存会话
func (m *SessionManager) dbSessionToSession(dbSession *db.ClaudeSession) *Session {
	// 解析上下文
	var context *AgentContext
	if dbSession.Context != "" {
		var dbContext db.ClaudeSessionContext
		if err := json.Unmarshal([]byte(dbSession.Context), &dbContext); err == nil {
			context = &AgentContext{
				ProjectID:            dbContext.ProjectID,
				ProjectName:          dbContext.ProjectName,
				SelectedTrafficIDs:   dbContext.SelectedTrafficIDs,
				SelectedVulnIDs:      dbContext.SelectedVulnIDs,
				SelectedFingerprints: dbContext.SelectedFingerprints,
				CustomData:           dbContext.CustomData,
				AutoCollect:          dbContext.AutoCollect,
			}
		}
	}

	// 解析历史
	var history []ChatMessage
	if dbSession.History != "" {
		var dbHistory []db.ClaudeSessionMessage
		if err := json.Unmarshal([]byte(dbSession.History), &dbHistory); err == nil {
			history = make([]ChatMessage, len(dbHistory))
			for i, msg := range dbHistory {
				toolUses := make([]ToolUse, len(msg.ToolUses))
				for j, tu := range msg.ToolUses {
					toolUses[j] = ToolUse{
						ID:     tu.ID,
						Name:   tu.Name,
						Input:  tu.Input,
						Status: tu.Status,
						Result: tu.Result,
						Error:  tu.Error,
					}
				}
				history[i] = ChatMessage{
					ID:           msg.ID,
					Role:         msg.Role,
					Content:      msg.Content,
					Timestamp:    msg.Timestamp,
					ToolUses:     toolUses,
					CostUSD:      msg.CostUSD,
					InputTokens:  msg.InputTokens,
					OutputTokens: msg.OutputTokens,
				}
			}
		}
	}

	return &Session{
		ID:             dbSession.SessionID,
		ProjectID:      dbSession.ProjectID,
		Context:        context,
		CreatedAt:      dbSession.CreatedAt,
		UpdatedAt:      dbSession.UpdatedAt,
		History:        history,
		ConversationID: dbSession.ConversationID,
	}
}

// LoadAllSessionsFromDB 从数据库加载所有会话到内存（启动时调用）
func (m *SessionManager) LoadAllSessionsFromDB() {
	dbSessions, err := db.GetAllClaudeSessions()
	if err != nil {
		logging.Logger.Errorf("从数据库加载会话失败: %v", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, dbSession := range dbSessions {
		session := m.dbSessionToSession(dbSession)
		if session != nil {
			m.sessions[session.ID] = session
		}
	}

	logging.Logger.Infof("从数据库加载了 %d 个会话", len(dbSessions))
}
