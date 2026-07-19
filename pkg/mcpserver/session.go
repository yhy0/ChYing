package mcpserver

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yhy0/ChYing/pkg/db"
	"gorm.io/gorm"
)

// ScanSession 扫描会话
type ScanSession struct {
	SessionID   string    `json:"session_id"`
	ProjectID   string    `json:"project_id"`
	Targets     []string  `json:"targets"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Active      bool      `json:"active"`
}

// sessionStore 内存中的 session 存储
var sessionStore sync.Map

func persistSession(session *ScanSession) {
	if session == nil {
		return
	}
	targets, err := json.Marshal(session.Targets)
	if err != nil {
		return
	}
	_ = db.SaveMCPScanSession(&db.MCPScanSession{
		SessionID: session.SessionID, ProjectID: session.ProjectID, TargetsJSON: string(targets),
		Description: session.Description, CreatedAt: session.CreatedAt, Active: session.Active,
	})
}

func sessionFromRecord(record *db.MCPScanSession) *ScanSession {
	if record == nil {
		return nil
	}
	var targets []string
	_ = json.Unmarshal([]byte(record.TargetsJSON), &targets)
	return &ScanSession{
		SessionID: record.SessionID, ProjectID: record.ProjectID, Targets: targets,
		Description: record.Description, CreatedAt: record.CreatedAt, Active: record.Active,
	}
}

// RegisterSession 注册新的扫描会话
func RegisterSession(targets []string, description string, projectID ...string) *ScanSession {
	project := "default"
	if len(projectID) > 0 && projectID[0] != "" {
		project = projectID[0]
	}
	session := &ScanSession{
		SessionID:   uuid.New().String(),
		ProjectID:   project,
		Targets:     targets,
		Description: description,
		CreatedAt:   time.Now(),
		Active:      true,
	}
	sessionStore.Store(session.SessionID, session)
	persistSession(session)
	return session
}

// GetSession 获取指定会话
func GetSession(sessionID string) (*ScanSession, bool) {
	val, ok := sessionStore.Load(sessionID)
	if ok {
		return val.(*ScanSession), true
	}
	record, err := db.GetMCPScanSession(sessionID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, false
	}
	session := sessionFromRecord(record)
	if session == nil {
		return nil, false
	}
	sessionStore.Store(sessionID, session)
	return session, true
}

// ListSessions 列出所有活跃会话
func ListSessions() []*ScanSession {
	sessions := make([]*ScanSession, 0)
	seen := make(map[string]bool)
	sessionStore.Range(func(key, value interface{}) bool {
		s := value.(*ScanSession)
		if s.Active {
			sessions = append(sessions, s)
			seen[s.SessionID] = true
		}
		return true
	})
	if records, err := db.ListActiveMCPScanSessions(db.CurrentProjectName); err == nil {
		for i := range records {
			if seen[records[i].SessionID] {
				continue
			}
			session := sessionFromRecord(&records[i])
			if session != nil {
				sessionStore.Store(session.SessionID, session)
				sessions = append(sessions, session)
			}
		}
	}
	return sessions
}

// ConfigureSession 修改会话的目标列表
func ConfigureSession(sessionID string, addTargets, removeTargets []string) (*ScanSession, bool) {
	original, ok := GetSession(sessionID)
	if !ok {
		return nil, false
	}
	// Copy-on-write: create a new session to avoid race conditions
	session := &ScanSession{
		SessionID:   original.SessionID,
		ProjectID:   original.ProjectID,
		Targets:     make([]string, len(original.Targets)),
		Description: original.Description,
		CreatedAt:   original.CreatedAt,
		Active:      original.Active,
	}
	copy(session.Targets, original.Targets)

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
	persistSession(session)
	return session, true
}

// CloseSession 关闭会话
func CloseSession(sessionID string) bool {
	original, ok := GetSession(sessionID)
	if !ok {
		return false
	}
	// Copy-on-write: create a new session to avoid race conditions
	session := &ScanSession{
		SessionID:   original.SessionID,
		ProjectID:   original.ProjectID,
		Targets:     original.Targets,
		Description: original.Description,
		CreatedAt:   original.CreatedAt,
		Active:      false,
	}
	sessionStore.Store(sessionID, session)
	persistSession(session)
	return true
}
