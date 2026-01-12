package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yhy0/ChYing/mitmproxy"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/collection"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/7/12
   @desc 扫描目标管理相关方法
**/

// Dashboard 获取仪表盘数据
func (a *App) Dashboard() []*Msg {
	var msg []*Msg
	for _, list := range output.SCopilotLists {
		msg = append(msg, getMsg(list.Host))
	}
	logging.Logger.Infoln("[+] Dashboard 获取数据成功！", len(msg))
	return msg
}

// GetTargetInfo 获取目标信息
func (a *App) GetTargetInfo(host string) *Msg {
	return getMsg(host)
}

// GetCollectionMsg 获取收集信息
func (a *App) GetCollectionMsg(host string) output.Collection {
	output.Lock.Lock()
	defer output.Lock.Unlock()
	logging.Logger.Infoln("[+] GetCollectionMsg 获取数据成功！", host, output.SCopilotMessage[host].CollectionMsg)
	return output.SCopilotMessage[host].CollectionMsg
}

// RegressionScan 回归扫描，当修改配置文件后，进行回归扫描，比如新增了提取api 的正则，回溯重扫
func (a *App) RegressionScan() {
	RegressionScan()
}

// RegressionScan 回归扫描实现
func RegressionScan() {
	go func() {
		page := 1
		pageSize := 100
		var count = 1
		for {
			offset := (page - 1) * pageSize

			total, result := db.GetResponse(offset, pageSize)

			if len(result) > 0 {
				for _, res := range result {
					float, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(count)/float64(total)*100), 64)
					Percentage <- float
					count += 1
					if mitmproxy.Filter(res.Host) {
						continue
					}

					var hostNoPort string
					_host := strings.Split(res.Host, ":")
					if len(_host) > 1 {
						hostNoPort = _host[0]
					} else {
						hostNoPort = res.Host
					}
					collectionMsg := collection.Info(res.Host+res.Path, hostNoPort, res.ResponseRaw, res.ContentType)

					msg := output.SCopilotData{
						Target:        res.Host,
						CollectionMsg: collectionMsg,
					}
					// 更新数据
					output.SCopilot(res.Host, msg)

					logging.Logger.Infoln("[+] RegressionScan 扫描中...", total, count, fmt.Sprintf("%.2f", float64(count)/float64(total)*100), float)
				}
				page += 1
			} else {
				break
			}

			if total <= int64(count) {
				break
			}
		}
		Notify <- []string{"回归扫描完成！", "success"}
	}()
}

// Search 搜索
func (a *App) Search(pattern string, host []string) {
	go func() {
		var count = 1
		result := db.GetResponseByHost(host)
		total := len(result)
		if len(result) > 0 {
			for _, res := range result {
				logging.Logger.Infoln("[+] Search ...", total, count, fmt.Sprintf("%.2f", float64(count)/float64(total)*100), res.RequestId, res.Url)

				err, _result := utils.Search(pattern, res.ResponseRaw)

				float, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(count)/float64(total)*100), 64)
				count += 1

				RePercentage <- float
				if err != nil {
					Notify <- []string{"正则解析失败！", " error"}
					continue
				}

				if len(_result) == 0 {
					continue
				}
				_res := make(map[string][]string)
				if res.Url == "" {
					res.Url = "test"
				}
				_res[res.Url] = _result
				wailsApp.Event.Emit("RegexpResult", _res)
			}
		}
		Notify <- []string{"扫描完成！", "success"}
	}()
}

// GetRequestScanPBody 根据ID获取HTTPBody信息，供前端查询使用
func (a *App) GetRequestScanPBody(id int64) Result {
	// 首先尝试从内存中获取
	if body, ok := httpx.HTTPBodyMap.Load(id); ok {
		if httpBody, typeOk := body.(*httpx.HttpBody); typeOk {
			return Result{
				Data:  httpBody,
				Error: "nil",
			}
		}
		return Result{
			Error: fmt.Sprintf("stored data is not HttpBody type for id: %d", id),
		}
	}

	// 如果内存中没有，尝试从数据库中获取
	req, res := db.GetTraffic(int(id))
	if req != nil || res != nil {
		httpBody := &httpx.HttpBody{
			Id:          id,
			RequestRaw:  "",
			ResponseRaw: "",
		}
		if req != nil {
			httpBody.RequestRaw = req.RequestRaw
		}
		if res != nil {
			httpBody.ResponseRaw = res.ResponseRaw
		}
		return Result{
			Data:  httpBody,
			Error: "nil",
		}
	}

	return Result{
		Error: fmt.Sprintf("no http body found for id: %d", id),
	}
}

// GetScanTargets 获取扫描目标列表
func (a *App) GetScanTargets(status string, limit int, offset int) Result {
	var targetStatus db.ScanTargetStatus
	if status != "" {
		targetStatus = db.ScanTargetStatus(status)
	}

	targets, total, err := db.GetScanTargets(targetStatus, limit, offset)
	if err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: map[string]interface{}{
		"targets": targets,
		"total":   total,
	}}
}

// AddScanTarget 添加扫描目标
func (a *App) AddScanTarget(target db.ScanTarget) Result {
	if target.CreatedBy == "" {
		target.CreatedBy = "本地用户"
	}

	if err := db.AddScanTarget(&target); err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: "添加成功"}
}

// BatchAddScanTargets 批量添加扫描目标
func (a *App) BatchAddScanTargets(targets []string, targetType string, createdBy string) Result {
	if createdBy == "" {
		createdBy = "本地用户"
	}

	err := db.BatchAddScanTargets(targets, db.ScanTargetType(targetType), nil, createdBy)
	if err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: fmt.Sprintf("批量添加成功，共 %d 个目标", len(targets))}
}

// UpdateScanTarget 更新扫描目标
func (a *App) UpdateScanTarget(target db.ScanTarget) Result {
	if err := db.UpdateScanTarget(&target); err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: "更新成功"}
}

// DeleteScanTarget 删除扫描目标
func (a *App) DeleteScanTarget(id int64) Result {
	if err := db.DeleteScanTarget(id); err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: "删除成功"}
}

// GetScanTargetByID 根据ID获取扫描目标
func (a *App) GetScanTargetByID(id int64) Result {
	target, err := db.GetScanTargetByID(id)
	if err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: target}
}

// UpdateScanTargetStatus 更新扫描目标状态
func (a *App) UpdateScanTargetStatus(id int64, status string, message string) Result {
	err := db.UpdateScanTargetStatus(id, db.ScanTargetStatus(status), message)
	if err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: "状态更新成功"}
}

// GetScanStatistics 获取扫描统计信息
func (a *App) GetScanStatistics() Result {
	stats, err := db.GetScanStatistics()
	if err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: stats}
}

// CleanupCompletedTargets 清理已完成的目标
func (a *App) CleanupCompletedTargets(days int) Result {
	err := db.CleanupCompletedTargets(days)
	if err != nil {
		return Result{Error: err.Error()}
	}

	return Result{Data: "清理完成"}
}

// GetDefaultScanConfig 获取默认扫描配置
func (a *App) GetDefaultScanConfig() Result {
	config := &db.ScanConfig{
		EnablePortScan:    true,
		EnableDirScan:     true,
		EnableVulnScan:    true,
		EnableFingerprint: true,
		EnableCrawler:     false,
		EnableXSS:         false,
		EnableSQL:         false,
		EnableBypass403:   false,
		Threads:           10,
		Timeout:           30,
		CustomHeaders:     []string{},
		Exclude:           []string{},
		PortRange:         "80,443,8080,8443,3000,5000,8000,9000",
		MaxDepth:          2,
	}

	return Result{Data: config}
}
