package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/yhy0/ChYing/api"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/mitmproxy"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/7/12
   @desc 数据库和历史记录相关方法
**/

// GetDatabaseSize 获取数据库文件大小信息
func (a *App) GetDatabaseSize(dbPath string) Result {
	fileInfo, err := utils.GetDBFileInfo(dbPath)
	if err != nil {
		return Result{
			Error: fmt.Sprintf("获取数据库文件信息失败: %v", err),
		}
	}

	return Result{
		Data: fileInfo,
	}
}

// GetAllDatabaseSizes 获取所有数据库文件的大小信息
func (a *App) GetAllDatabaseSizes() Result {
	dbFiles, err := utils.GetDBFiles(file.ChyingDir)
	if err != nil {
		return Result{
			Error: fmt.Sprintf("获取数据库文件列表失败: %v", err),
		}
	}

	result := make(map[string]interface{})
	for dbName, dbPath := range dbFiles {
		fileInfo, err := utils.GetDBFileInfo(dbPath)
		if err != nil {
			logging.Logger.Warnf("获取数据库 %s 大小信息失败: %v", dbName, err)
			continue
		}
		result[dbName] = fileInfo
	}

	return Result{
		Data: result,
	}
}

// GetHosts 获取项目
func (a *App) GetHosts() []string {
	return db.GetHosts()
}

// GetHistoryByHost 根据主机获取历史记录
func (a *App) GetHistoryByHost(host []string) []*db.HTTPHistory {
	return db.GetHistory(host)
}

// GetHistory 根据ID获取历史记录
func (a *App) GetHistory(id int) *mitmproxy.HTTPBody {
	req, res := db.GetTraffic(id)
	return &mitmproxy.HTTPBody{
		Id:          int64(id),
		RequestRaw:  req.RequestRaw,
		ResponseRaw: res.ResponseRaw,
	}
}

// GetHistoryDumpIndex 已在 mitmproxy.go 中定义

// ClearAllHistoryData 清空所有历史数据（数据库+内存）
func (a *App) ClearAllHistoryData() Result {
	// 清空数据库中的历史记录
	if err := db.ClearAllHistory(); err != nil {
		logging.Logger.Errorf("清空数据库历史记录失败: %v", err)
		return Result{Error: "清空数据库历史记录失败: " + err.Error()}
	}

	// 清空内存中的HTTP Body Map
	CleanHTTPBodyMap(0) // 清空所有内存数据

	// 清空HTTP URL Map
	mitmproxy.ClearHTTPUrlMap()

	// 清空历史记录映射
	HTTPHistoryMap = sync.Map{}

	logging.Logger.Infoln("所有历史数据已清空（数据库+内存）")
	return Result{Data: "所有历史数据已清空"}
}

// CleanMemoryData 清理内存数据，保留指定数量的最新条目
func (a *App) CleanMemoryData(keepCount int) Result {
	if keepCount < 0 {
		keepCount = 0
	}

	// 清理HTTP Body Map
	CleanHTTPBodyMap(keepCount)

	// 如果要完全清空，也清理URL映射
	if keepCount == 0 {
		mitmproxy.ClearHTTPUrlMap()
		HTTPHistoryMap = sync.Map{}
	}

	logging.Logger.Infof("内存数据已清理，保留 %d 个最新条目", keepCount)
	return Result{Data: fmt.Sprintf("内存数据已清理，保留 %d 个最新条目", keepCount)}
}

// GetLocalProjects 获取本地项目列表
func (a *App) GetLocalProjects() Result {
	// 确保API管理器已初始化
	if a.apiManager == nil {
		a.apiManager = api.NewAPIManager()
	}
	r := a.apiManager.GetLocalProjects()
	return Result{Data: r.Data, Error: r.Error}
}

// CreateLocalProject 创建本地项目
func (a *App) CreateLocalProject(projectID string, projectName string) Result {
	logging.Logger.Infof("🔧 CreateLocalProject 被调用: projectID=%s, projectName=%s", projectID, projectName)
	// 确保API管理器已初始化
	if a.apiManager == nil {
		a.apiManager = api.NewAPIManager()
	}
	r := a.apiManager.CreateLocalProject(projectID, projectName)
	logging.Logger.Infof("🔧 CreateLocalProject 结果: %+v", r)
	return Result{Data: r.Data, Error: r.Error}
}

// DeleteLocalProject 删除本地项目
func (a *App) DeleteLocalProject(projectName string) Result {
	logging.Logger.Infof("🗑️ DeleteLocalProject 被调用: projectName=%s", projectName)
	// 确保API管理器已初始化
	if a.apiManager == nil {
		a.apiManager = api.NewAPIManager()
	}
	r := a.apiManager.DeleteLocalProject(projectName)
	logging.Logger.Infof("🗑️ DeleteLocalProject 结果: %+v", r)
	return Result{Data: r.Data, Error: r.Error}
}

// loadExistingProjectData 加载现有项目数据
func (a *App) loadExistingProjectData() {
	// 确保API管理器已初始化
	if a.apiManager == nil {
		a.apiManager = api.NewAPIManager()
	}
	output.Lock.Lock()
	defer output.Lock.Unlock()

	res := db.GetSCopilotList()

	for _, item := range res {
		output.SCopilotLists = append(output.SCopilotLists, &output.SCopilotList{
			Host:      item.Host,
			InfoCount: item.InfoCount,
			ApiCount:  item.ApiCount,
			VulnCount: item.VulnCount,
		})
		_data := &output.SCopilotData{}
		if err := json.Unmarshal([]byte(item.JsonData), _data); err != nil {
			logging.Logger.Errorln("json unmarshal failed!", item.ID)
			continue
		}
		output.SCopilotMessage[item.Host] = _data
	}

	ipRes := db.GetAllIPInfo()

	for _, item := range ipRes {
		portService := make(map[int]string)
		_portStr := strings.Split(item.PortService, "<<sep>>")
		for _, port := range _portStr {
			if port == "" {
				continue
			}
			_str := strings.Split(port, ":")
			if len(_str) < 2 {
				continue
			}
			_port, err := strconv.Atoi(_str[0])
			if err != nil {
				logging.Logger.Errorf("Failed to parse port: %v", err)
				continue
			}
			portService[_port] = _str[1]
		}

		output.IPInfoList[item.Host] = &output.IPInfo{
			Ip:          item.Ip,
			AllRecords:  strings.Split(item.AllRecords, "<<sep>>"),
			PortService: portService,
			Type:        item.Type,
			Value:       item.Value,
			Cdn:         item.Cdn,
		}
	}

	history, err := db.GetAllHistory("", "", 0, 0)
	if err != nil {
		logging.Logger.Errorln("Failed to get history:", err)
		return
	}
	lock.Lock()
	defer lock.Unlock()

	for _, item := range history {
		httpHistory := mitmproxy.HTTPHistory{
			Id:          item.Hid,
			Host:        item.Host,
			Method:      item.Method,
			FullUrl:     item.FullUrl,
			Path:        item.Path,
			Status:      item.Status,
			Length:      item.Length,
			ContentType: item.ContentType,
			MIMEType:    item.MIMEType,
			Extension:   item.Extension,
			Title:       item.Title,
			IP:          item.IP,
			Color:       item.Color,
			Note:        item.Note,
		}

		wailsApp.Event.Emit("HttpHistory", httpHistory)
		request, response := db.GetTraffic(int(item.Hid))
		mitmproxy.HistoryItemIDGenerator.Add(1)
		mitmproxy.HTTPBodyMap.Store(item.Hid, &mitmproxy.HTTPBody{
			Id:          item.Hid,
			FlowID:      "", // 填充一个空的FlowID，因为历史记录可能没有这个字段
			RequestRaw:  request.RequestRaw,
			ResponseRaw: response.ResponseRaw,
		})
		mitmproxy.SetHTTPUrlMapValue(item.FullUrl, item.Hid)
	}
	output.DataUpdated <- struct{}{}
}
