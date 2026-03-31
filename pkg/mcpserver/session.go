package mcpserver

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ScanSession 扫描会话
type ScanSession struct {
	SessionID   string    `json:"session_id"`
	Targets     []string  `json:"targets"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Active      bool      `json:"active"`
}

// sessionStore 内存中的 session 存储
var sessionStore sync.Map

// RegisterSession 注册新的扫描会话
func RegisterSession(targets []string, description string) *ScanSession {
	session := &ScanSession{
		SessionID:   uuid.New().String(),
		Targets:     targets,
		Description: description,
		CreatedAt:   time.Now(),
		Active:      true,
	}
	sessionStore.Store(session.SessionID, session)
	return session
}

// GetSession 获取指定会话
func GetSession(sessionID string) (*ScanSession, bool) {
	val, ok := sessionStore.Load(sessionID)
	if !ok {
		return nil, false
	}
	return val.(*ScanSession), true
}

// ListSessions 列出所有活跃会话
func ListSessions() []*ScanSession {
	var sessions []*ScanSession
	sessionStore.Range(func(key, value interface{}) bool {
		s := value.(*ScanSession)
		if s.Active {
			sessions = append(sessions, s)
		}
		return true
	})
	return sessions
}

// ConfigureSession 修改会话的目标列表
func ConfigureSession(sessionID string, addTargets, removeTargets []string) (*ScanSession, bool) {
	val, ok := sessionStore.Load(sessionID)
	if !ok {
		return nil, false
	}
	session := val.(*ScanSession)

	if len(removeTargets) > 0 {
		removeSet := make(map[string]bool)
		for _, t := range removeTargets {
			removeSet[t] = true
		}
		var filtered []string
		for _, t := range session.Targets {
			if !removeSet[t] {
				filtered = append(filtered, t)
			}
		}
		session.Targets = filtered
	}

	if len(addTargets) > 0 {
		existing := make(map[string]bool)
		for _, t := range session.Targets {
			existing[t] = true
		}
		for _, t := range addTargets {
			if !existing[t] {
				session.Targets = append(session.Targets, t)
			}
		}
	}

	sessionStore.Store(sessionID, session)
	return session, true
}

// CloseSession 关闭会话
func CloseSession(sessionID string) bool {
	val, ok := sessionStore.Load(sessionID)
	if !ok {
		return false
	}
	session := val.(*ScanSession)
	session.Active = false
	sessionStore.Store(sessionID, session)
	return true
}
