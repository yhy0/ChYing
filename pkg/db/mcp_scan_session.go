package db

import (
	"time"
)

// MCPScanSession persists MCP traffic-isolation sessions so external clients can
// resume the same project/session attribution after ChYing restarts.
type MCPScanSession struct {
	SessionID   string    `gorm:"primaryKey" json:"session_id"`
	ProjectID   string    `gorm:"index;not null" json:"project_id"`
	TargetsJSON string    `gorm:"type:text;not null" json:"targets_json"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Active      bool      `gorm:"index;not null;default:true" json:"active"`
}

func SaveMCPScanSession(session *MCPScanSession) error {
	if GlobalDB == nil || session == nil {
		return nil
	}
	return RetryOnLocked("SaveMCPScanSession", func() error {
		return GlobalDB.Save(session).Error
	}, 5)
}

func GetMCPScanSession(sessionID string) (*MCPScanSession, error) {
	if GlobalDB == nil || sessionID == "" {
		return nil, nil
	}
	var session MCPScanSession
	result := GlobalDB.Where("session_id = ?", sessionID).First(&session)
	if result.Error != nil {
		return nil, result.Error
	}
	return &session, nil
}

func ListActiveMCPScanSessions(projectID string) ([]MCPScanSession, error) {
	if GlobalDB == nil {
		return nil, nil
	}
	var sessions []MCPScanSession
	query := GlobalDB.Where("active = ?", true)
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if err := query.Order("created_at ASC").Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}
