package db

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetSessionHistoryIncludesTargetBoundedUnattributedRows(t *testing.T) {
	previousDB := GlobalDB
	database, err := gorm.Open(sqlite.Open("file:session-history?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.AutoMigrate(&HTTPHistory{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	GlobalDB = database
	t.Cleanup(func() { GlobalDB = previousDB })

	rows := []HTTPHistory{
		{Hid: 1, Host: "api.example.com", ProjectID: "src-auto", SessionID: "session-1"},
		{Hid: 2, Host: "api.example.com", ProjectID: "src-auto", SessionID: ""},
		{Hid: 3, Host: "edge.api.example.com", ProjectID: "default", SessionID: ""},
		{Hid: 4, Host: "unrelated.test", ProjectID: "src-auto", SessionID: ""},
		{Hid: 5, Host: "api.example.com", ProjectID: "another-project", SessionID: ""},
	}
	if err := database.Create(&rows).Error; err != nil {
		t.Fatalf("insert history: %v", err)
	}

	histories, err := GetSessionHistory("src-auto", "", "session-1", []string{"api.example.com"}, true, 100, 0)
	if err != nil {
		t.Fatalf("GetSessionHistory() error = %v", err)
	}
	if len(histories) != 3 {
		t.Fatalf("expected 3 target-bounded rows, got %d", len(histories))
	}
}

func TestGetSessionVulnerabilitiesIncludesTargetBoundedUnattributedRows(t *testing.T) {
	previousDB := GlobalDB
	database, err := gorm.Open(sqlite.Open("file:session-vulnerabilities?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.AutoMigrate(&Vulnerability{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	GlobalDB = database
	t.Cleanup(func() { GlobalDB = previousDB })

	rows := []Vulnerability{
		{VulnID: "v1", Host: "api.example.com", Target: "https://api.example.com/a", ProjectID: "src-auto", SessionID: "session-1"},
		{VulnID: "v2", Host: "edge.api.example.com", Target: "https://edge.api.example.com/b", ProjectID: "src-auto", SessionID: ""},
		{VulnID: "v3", Host: "unrelated.test", Target: "https://unrelated.test/c", ProjectID: "src-auto", SessionID: ""},
		{VulnID: "v4", Host: "api.example.com", Target: "https://api.example.com/d", ProjectID: "another-project", SessionID: ""},
	}
	if err := database.Create(&rows).Error; err != nil {
		t.Fatalf("insert vulnerabilities: %v", err)
	}

	vulnerabilities, err := GetSessionVulnerabilities("src-auto", "", "session-1", []string{"api.example.com"}, true, 100, 0)
	if err != nil {
		t.Fatalf("GetSessionVulnerabilities() error = %v", err)
	}
	if len(vulnerabilities) != 2 {
		t.Fatalf("expected 2 target-bounded rows, got %d", len(vulnerabilities))
	}
}
