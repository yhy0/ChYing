package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/yhy0/logging"
	"gorm.io/gorm"
)

/**
   @author yhy
   @since 2024/9/24
   @desc //TODO
**/

type HTTPHistory struct {
	ID          int64  `gorm:"primary_key;auto_increment" json:"id"`
	Hid         int64  `json:"hid"`
	Host        string `gorm:"index" json:"host"`
	Method      string `json:"method"`
	FullUrl     string `json:"full_url"`
	Path        string `gorm:"index" json:"path"`
	Params      string `json:"params"`
	Edited      string `json:"edited"`
	Status      string `json:"status"`
	Length      string `json:"length"`
	ContentType string `json:"content_type"`
	MIMEType    string `json:"mime_type"`
	Extension   string `json:"extension"`
	Title       string `json:"title"`
	Comment     string `json:"comment"`
	TLS         string `json:"tls"`
	IP          string `gorm:"index" json:"ip"`
	Color       string `json:"color"`
	Note        string `json:"note"`
	Cookies     string `json:"cookies"`
	Time        string `json:"time"`

	// 新增字段用于区分流量来源
	Source   string `gorm:"index;default:'local'" json:"source"` // 流量来源: 'local' | 'remote' | 'crawler'
	SourceID string `gorm:"index" json:"source_id"`              // 来源标识: 本地IP或服务器ID
	NodeName string `json:"node_name"`                           // 节点名称: 便于识别

	// 项目标识字段
	ProjectID string `gorm:"index;default:'default'" json:"project_id"` // 项目ID，用于区分不同项目的数据

	// Session 标识字段
	SessionID string `gorm:"index;default:''" json:"session_id"` // 扫描会话ID，用于多Agent隔离

	CreatedAt time.Time
	UpdatedAt time.Time
}

// GetSessionHistory returns traffic attributed to a session and can also read
// pre-attribution rows from the same project database, bounded by session targets.
// Legacy rows remain unchanged in ChYing; callers receive read-only references.
func GetSessionHistory(projectID, source, sessionID string, targets []string, includeUnattributed bool, limit, offset int) ([]*HTTPHistory, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	query := GlobalDB.Model(&HTTPHistory{})
	if includeUnattributed {
		query = query.Where(
			"((project_id = ? AND (session_id = ? OR session_id = '' OR session_id IS NULL)) OR ((project_id = '' OR project_id = 'default') AND (session_id = '' OR session_id IS NULL)))",
			projectID, sessionID,
		)
	} else {
		query = query.Where("project_id = ? AND session_id = ?", projectID, sessionID)
	}
	if source != "" && source != "all" {
		query = query.Where("source = ?", source)
	}
	if len(targets) > 0 {
		clauses := make([]string, 0, len(targets))
		args := make([]interface{}, 0, len(targets)*3)
		for _, raw := range targets {
			target := strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(raw), "*."), "."))
			if target == "" {
				continue
			}
			clauses = append(clauses, "(LOWER(host) = ? OR LOWER(host) LIKE ? OR LOWER(host) LIKE ?)")
			args = append(args, target, "%."+target, target+":%")
		}
		if len(clauses) > 0 {
			query = query.Where("("+strings.Join(clauses, " OR ")+")", args...)
		}
	}
	var data []*HTTPHistory
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func AddHistory(data *HTTPHistory) {
	// 设置默认值
	if data.Source == "" {
		data.Source = "local"
	}
	if data.SourceID == "" {
		data.SourceID = "localhost"
	}
	if data.NodeName == "" {
		data.NodeName = "本地节点"
	}
	if data.ProjectID == "" {
		data.ProjectID = "default"
	}

	// 使用 SQLite 数据库
	if GlobalDB == nil {
		logging.Logger.Warnln("数据库未初始化，无法添加历史记录")
		return
	}

	if !GlobalDB.Migrator().HasTable(&HTTPHistory{}) {
		err := GlobalDB.AutoMigrate(&HTTPHistory{})
		if err != nil {
			logging.Logger.Errorln("Table Create err:", err)
			return
		}
	}

	if !ExistHistory(data.Hid) {
		err := RetryOnLocked("AddHistory", func() error {
			return GlobalDB.Create(&data).Error
		}, 3)
		if err != nil {
			logging.Logger.Errorln("AddHistory err:", err)
		}
	}
}

// AddRemoteHistory 添加远程流量历史记录
func AddRemoteHistory(data *HTTPHistory, sourceID, nodeName string) {
	data.Source = "remote"
	data.SourceID = sourceID
	data.NodeName = nodeName
	AddHistory(data)
}

// AddCrawlerHistory 添加爬虫流量历史记录
func AddCrawlerHistory(data *HTTPHistory, sourceID, nodeName string) {
	data.Source = "crawler"
	data.SourceID = sourceID
	data.NodeName = nodeName
	AddHistory(data)
}

func GetLastHid() int64 {
	var hid int64
	if GlobalDB == nil {
		return 0
	}
	GlobalDB.Model(&HTTPHistory{}).Select("hid").Order("hid desc").Limit(1).Find(&hid)
	return hid
}

// ExistHistory 检查历史记录是否存在
func ExistHistory(hid int64) bool {
	if GlobalDB == nil {
		return false
	}
	var count int64
	GlobalDB.Model(&HTTPHistory{}).Where("hid = ?", hid).Count(&count)
	return count > 0
}

// GetAllHistory 获取所有历史记录，支持项目ID过滤
func GetAllHistory(projectID string, source string, limit, offset int, sessionID ...string) ([]*HTTPHistory, error) {
	var sid string
	if len(sessionID) > 0 {
		sid = sessionID[0]
	}

	// 使用 SQLite 数据库
	var data []*HTTPHistory
	query := GlobalDB.Model(&HTTPHistory{})

	// 添加项目ID过滤
	if projectID != "" && projectID != "all" {
		query = query.Where("project_id = ?", projectID)
	}

	// 添加来源过滤
	if source != "" && source != "all" {
		query = query.Where("source = ?", source)
	}

	if sid != "" {
		query = query.Where("session_id = ?", sid)
	}

	if limit > 0 {
		query = query.Order("created_at DESC").Limit(limit).Offset(offset)
	} else {
		query = query.Order("created_at DESC")
	}

	query.Find(&data)

	return data, nil
}

func GetHistory(host []string) []*HTTPHistory {
	var history []*HTTPHistory
	if GlobalDB == nil {
		return history
	}

	globalDBTmp := GlobalDB.Model(&HTTPHistory{})
	for i, h := range host {
		if i > 0 {
			globalDBTmp = globalDBTmp.Or("host = ?", h)
		} else {
			globalDBTmp = globalDBTmp.Where("host = ?", h)
		}
	}
	globalDBTmp.Find(&history)

	return history
}

// GetHistoryByHost returns traffic for the current project and optional agent session.
// GetHistory is kept for existing desktop callers that intentionally query the whole project.
func GetHistoryByHost(projectID string, hosts []string, sessionID string) ([]*HTTPHistory, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	query := GlobalDB.Model(&HTTPHistory{})
	if projectID != "" && projectID != "all" {
		query = query.Where("project_id = ?", projectID)
	}
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	if len(hosts) > 0 {
		query = query.Where("host IN ?", hosts)
	}
	var history []*HTTPHistory
	if err := query.Order("created_at DESC").Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

func GetHosts() []string {
	var hosts []string
	if GlobalDB == nil {
		return hosts
	}

	globalDBTmp := GlobalDB.Model(&HTTPHistory{})
	var history []*HTTPHistory
	globalDBTmp.Select("distinct host").Find(&history)

	for _, h := range history {
		hosts = append(hosts, h.Host)
	}
	return hosts
}

// GetHistoryByID 根据 ID 获取单条历史记录
func GetHistoryByID(id int64) (*HTTPHistory, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var history HTTPHistory
	err := GlobalDB.Where("id = ?", id).First(&history).Error
	if err != nil {
		return nil, err
	}

	return &history, nil
}

// GetHistoryByHid 根据 Hid 获取单条历史记录
func GetHistoryByHid(hid int64) (*HTTPHistory, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var history HTTPHistory
	err := GlobalDB.Where("hid = ?", hid).First(&history).Error
	if err != nil {
		return nil, err
	}

	return &history, nil
}

// GetHistoriesByHids 根据 Hid 列表批量获取历史记录
func GetHistoriesByHids(hids []int64) ([]*HTTPHistory, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	if len(hids) == 0 {
		return []*HTTPHistory{}, nil
	}

	var histories []*HTTPHistory
	err := GlobalDB.Where("hid IN ?", hids).Find(&histories).Error
	if err != nil {
		return nil, err
	}

	return histories, nil
}

func UpdateMarker(hid int64, color string, note string) {
	if GlobalDB == nil {
		logging.Logger.Warnln("数据库未初始化，无法更新标记")
		return
	}
	GlobalDB.Model(&HTTPHistory{}).Where("hid = ?", hid).Update("color", color).Update("note", note)
}

// GetNewHistorySince 获取指定时间之后的新历史记录
func GetNewHistorySince(since time.Time, sessionID string) ([]*HTTPHistory, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var data []*HTTPHistory
	query := GlobalDB.Model(&HTTPHistory{}).Where("created_at > ?", since)
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	query.Order("created_at ASC").Find(&data)
	return data, nil
}

// GetHostsBySession 获取指定 session 的所有域名
func GetHostsBySession(sessionID string) []string {
	var hosts []string
	if GlobalDB == nil {
		return hosts
	}
	query := GlobalDB.Model(&HTTPHistory{}).Select("DISTINCT host")
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	var histories []*HTTPHistory
	query.Find(&histories)
	for _, h := range histories {
		hosts = append(hosts, h.Host)
	}
	return hosts
}

// SearchHistory 搜索包含关键词的流量记录
// searchIn: "url" | "request_body" | "response_body" | "all"
func SearchHistory(projectID, keyword, searchIn string, limit int) ([]*HTTPHistory, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}
	if limit <= 0 {
		limit = 50
	}

	likeKeyword := "%" + keyword + "%"

	switch searchIn {
	case "url":
		var data []*HTTPHistory
		GlobalDB.Model(&HTTPHistory{}).
			Where("project_id = ? AND (full_url LIKE ? OR path LIKE ?)", projectID, likeKeyword, likeKeyword).
			Order("created_at DESC").Limit(limit).Find(&data)
		return data, nil

	case "request_body":
		// Join with requests table to search in request body
		var data []*HTTPHistory
		GlobalDB.Model(&HTTPHistory{}).
			Joins("JOIN requests ON requests.request_id = http_histories.hid").
			Where("http_histories.project_id = ? AND requests.request_raw LIKE ?", projectID, likeKeyword).
			Order("http_histories.created_at DESC").Limit(limit).Find(&data)
		return data, nil

	case "response_body":
		// Join with responses table to search in response body
		var data []*HTTPHistory
		GlobalDB.Model(&HTTPHistory{}).
			Joins("JOIN responses ON responses.request_id = http_histories.hid").
			Where("http_histories.project_id = ? AND responses.response_raw LIKE ?", projectID, likeKeyword).
			Order("http_histories.created_at DESC").Limit(limit).Find(&data)
		return data, nil

	default: // "all"
		var data []*HTTPHistory
		GlobalDB.Model(&HTTPHistory{}).
			Joins("LEFT JOIN requests ON requests.request_id = http_histories.hid").
			Joins("LEFT JOIN responses ON responses.request_id = http_histories.hid").
			Where("http_histories.project_id = ? AND (http_histories.full_url LIKE ? OR http_histories.path LIKE ? OR requests.request_raw LIKE ? OR responses.response_raw LIKE ?)",
				projectID, likeKeyword, likeKeyword, likeKeyword, likeKeyword).
			Order("http_histories.created_at DESC").Limit(limit).
			Select("DISTINCT http_histories.*").Find(&data)
		return data, nil
	}
}

// GetDistinctPaths 获取去重的路径列表，按主机分组
func GetDistinctPaths(projectID, host string) ([]string, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var paths []string
	query := GlobalDB.Model(&HTTPHistory{}).Select("DISTINCT path")
	if projectID != "" && projectID != "all" {
		query = query.Where("project_id = ?", projectID)
	}
	if host != "" {
		query = query.Where("host = ?", host)
	}
	query.Order("path ASC").Pluck("path", &paths)
	return paths, nil
}

// GetTrafficStatistics 获取流量统计信息
func GetTrafficStatistics(projectID string) (map[string]interface{}, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	stats := make(map[string]interface{})

	baseQuery := func() *gorm.DB {
		q := GlobalDB.Model(&HTTPHistory{})
		if projectID != "" && projectID != "all" {
			q = q.Where("project_id = ?", projectID)
		}
		return q
	}

	// 总流量数
	var totalTraffic int64
	baseQuery().Count(&totalTraffic)
	stats["traffic_count"] = totalTraffic

	// 主机数
	var hostCount int64
	baseQuery().Select("COUNT(DISTINCT host)").Scan(&hostCount)
	stats["host_count"] = hostCount

	// 按方法统计
	var methodStats []struct {
		Method string `json:"method"`
		Count  int64  `json:"count"`
	}
	baseQuery().Select("method, count(*) as count").Group("method").Find(&methodStats)
	methodMap := make(map[string]int64)
	for _, s := range methodStats {
		methodMap[s.Method] = s.Count
	}
	stats["by_method"] = methodMap

	// 按状态码统计
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	baseQuery().Select("status, count(*) as count").Group("status").Find(&statusStats)
	statusMap := make(map[string]int64)
	for _, s := range statusStats {
		statusMap[s.Status] = s.Count
	}
	stats["by_status"] = statusMap

	// 按 content_type 统计
	var ctStats []struct {
		ContentType string `json:"content_type"`
		Count       int64  `json:"count"`
	}
	baseQuery().Select("content_type, count(*) as count").Where("content_type != ''").Group("content_type").Order("count DESC").Limit(20).Find(&ctStats)
	ctMap := make(map[string]int64)
	for _, s := range ctStats {
		ctMap[s.ContentType] = s.Count
	}
	stats["by_content_type"] = ctMap

	return stats, nil
}

// ClearAllHistory 清空所有历史记录数据
func ClearAllHistory() error {
	// 清空 SQLite 数据库表
	if GlobalDB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	err := GlobalDB.Exec("DELETE FROM http_histories").Error
	if err != nil {
		logging.Logger.Errorln("ClearAllHistory - http_histories err:", err)
		return err
	}

	err = GlobalDB.Exec("DELETE FROM requests").Error
	if err != nil {
		logging.Logger.Errorln("ClearAllHistory - requests err:", err)
		return err
	}

	err = GlobalDB.Exec("DELETE FROM responses").Error
	if err != nil {
		logging.Logger.Errorln("ClearAllHistory - responses err:", err)
		return err
	}

	logging.Logger.Infoln("所有历史记录已从数据库中清空")
	return nil
}
