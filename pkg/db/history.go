package db

import (
	"fmt"
	"time"

	"github.com/yhy0/logging"
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

	CreatedAt time.Time
	UpdatedAt time.Time
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
func GetAllHistory(projectID string, source string, limit, offset int) ([]*HTTPHistory, error) {
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

func UpdateMarker(hid int64, color string, note string) {
	if GlobalDB == nil {
		logging.Logger.Warnln("数据库未初始化，无法更新标记")
		return
	}
	GlobalDB.Model(&HTTPHistory{}).Where("hid = ?", hid).Update("color", color).Update("note", note)
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
