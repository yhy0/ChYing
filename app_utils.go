package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/browser"
	"github.com/shirou/gopsutil/process"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/coder/twj"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/ChYing/pkg/gadgets/fuzz"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"

	reqv3 "github.com/imroc/req/v3"
)

/**
   @author yhy
   @since 2024/7/12
   @desc 工具方法
**/

// GetMemoryUsage 获取当前进程的系统内存使用情况
func (a *App) GetMemoryUsage() MemoryInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mainPid := int32(os.Getpid())
	mainProc, err := process.NewProcess(mainPid)
	if err != nil {
		logging.Logger.Errorln("获取主进程信息失败:", err)
		return fallbackMemoryInfo(m)
	}

	// 递归收集主进程及其所有子进程
	allPids := map[int32]struct{}{mainPid: {}}
	var collectChildren func(p *process.Process)
	collectChildren = func(p *process.Process) {
		children, err := p.Children()
		if err == nil {
			for _, child := range children {
				allPids[child.Pid] = struct{}{}
				collectChildren(child)
			}
		}
	}
	collectChildren(mainProc)
	// 统计内存
	var totalRSS, totalVMS uint64
	for pid := range allPids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}
		memInfo, err := p.MemoryInfo()
		if err == nil {
			totalRSS += memInfo.RSS
			totalVMS += memInfo.VMS
		}
	}

	return MemoryInfo{
		Alloc:          totalRSS,
		AllocFormatted: formatBytes(totalRSS),
		Sys:            totalVMS,
		SysFormatted:   formatBytes(totalVMS),
		NumGC:          m.NumGC,
		NumGoroutine:   runtime.NumGoroutine(),
	}
}

// fallbackMemoryInfo 备用内存信息
func fallbackMemoryInfo(m runtime.MemStats) MemoryInfo {
	return MemoryInfo{
		Alloc:          m.Alloc,
		AllocFormatted: formatBytes(m.Alloc),
		Sys:            m.Sys,
		SysFormatted:   formatBytes(m.Sys),
		NumGC:          m.NumGC,
		NumGoroutine:   runtime.NumGoroutine(),
	}
}

// formatBytes 格式化字节数为可读的字符串
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetAllRequestIDs 获取所有已记录的请求ID列表
func GetAllRequestIDs() []int64 {
	var ids []int64
	httpx.HTTPBodyMap.Range(func(key, value interface{}) bool {
		if id, ok := key.(int64); ok {
			ids = append(ids, id)
		}
		return true
	})
	return ids
}

// CleanHTTPBodyMap 清理HTTPBodyMap中的旧数据，保持内存使用在合理范围内
func CleanHTTPBodyMap(maxEntries int) {
	count := 0
	httpx.HTTPBodyMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	if count > maxEntries {
		// 删除最老的条目（这里简单地删除一些条目）
		deleteCount := count - maxEntries
		deletedCount := 0
		httpx.HTTPBodyMap.Range(func(key, value interface{}) bool {
			if deletedCount < deleteCount {
				httpx.HTTPBodyMap.Delete(key)
				deletedCount++
			}
			return deletedCount < deleteCount
		})
		logging.Logger.Infof("Cleaned %d old entries from HTTPBodyMap, current count: %d", deletedCount, maxEntries)
	}
}

// getMsg 获取消息
func getMsg(host string) *Msg {
	output.Lock.Lock()
	defer output.Lock.Unlock()
	ipInfo := output.IPInfoList[output.SCopilotMessage[host].HostNoPort]
	var paramsCnt int
	if output.SCopilotMessage[host].CollectionMsg.Parameters != nil {
		paramsCnt = len(output.SCopilotMessage[host].CollectionMsg.Parameters.Keys())
	}

	msg := &Msg{
		Target:       host,
		UUID:         host,
		IpAddress:    output.SCopilotMessage[host].IpAddress,
		SiteMap:      output.SCopilotMessage[host].SiteMap,
		Fingerprint:  output.SCopilotMessage[host].Fingerprints,
		APICnt:       len(output.SCopilotMessage[host].CollectionMsg.Api),
		SubdomainCnt: len(output.SCopilotMessage[host].CollectionMsg.Subdomain),
		ParamsCnt:    paramsCnt,
		InnerIpCnt:   len(output.SCopilotMessage[host].CollectionMsg.InnerIp),
		OtherCnt:     len(output.SCopilotMessage[host].CollectionMsg.Phone) + len(output.SCopilotMessage[host].CollectionMsg.Email) + len(output.SCopilotMessage[host].CollectionMsg.IdCard) + len(output.SCopilotMessage[host].CollectionMsg.Others),
	}
	if ipInfo != nil {
		msg.CDN = ipInfo.Cdn
		msg.IPMsg = strings.Trim(fmt.Sprintf("%s %s", ipInfo.Value, ipInfo.Type), " ")
		msg.Records = ipInfo.AllRecords
		msg.PortInfo = ipInfo.PortService
	}
	return msg
}

// startEventLoop 启动事件循环
func (a *App) startEventLoop() {
	for {
		select {
		// 数据更改
		case notify := <-Notify:
			wailsApp.Event.Emit("Notify", notify)
		case percentage := <-Percentage:
			wailsApp.Event.Emit("ReScanPercentage", percentage)
		case percentage := <-RePercentage:
			wailsApp.Event.Emit("RePercentage", percentage)
		case percentage := <-twj.Percentage:
			wailsApp.Event.Emit("Percentage", percentage)
		case percentage := <-fuzz.Percentage: // fuzz 的进度条
			wailsApp.Event.Emit("FuzzPercentage", percentage)
		case _fuzz := <-fuzz.Chan: // fuzz 表格数据
			wailsApp.Event.Emit("Fuzz", _fuzz)
		case <-output.DataUpdated:
			var msg []*Msg
			for _, list := range output.SCopilotLists {
				msg = append(msg, getMsg(list.Host))
			}
			wailsApp.Event.Emit("Dashboard", msg)
		case vuln := <-output.OutChannel:
			logging.Logger.Infoln(aurora.Red(vuln.PrintScreen()).String())

			// 🆕 将漏洞数据持久化到数据库
			go func(v output.VulMessage) {
				// 转换为数据库格式
				vulnData := &db.Vulnerability{
					VulnID:      fmt.Sprintf("%s-%s-%d", v.VulnData.VulnType, v.VulnData.Target, time.Now().UnixNano()),
					VulnType:    v.VulnData.VulnType,
					Target:      v.VulnData.Target,
					Host:        v.VulnData.Target, // 使用Target作为Host
					Method:      v.VulnData.Method,
					Path:        "", // VulnData中没有Path字段
					Plugin:      v.Plugin,
					Level:       v.Level,
					IP:          v.VulnData.Ip,
					Param:       v.VulnData.Param,
					Payload:     v.VulnData.Payload,
					Description: v.VulnData.Description,
					CurlCommand: v.VulnData.CURLCommand,
					Request:     v.VulnData.Request,
					Response:    v.VulnData.Response,
					Source:      "local",
					SourceID:    "localhost",
					NodeName:    "本地节点",
					ProjectID:   "default", // 暂时使用默认项目ID
				}

				// 添加到数据库
				if err := db.AddVulnerability(vulnData); err != nil {
					logging.Logger.Errorf("漏洞数据入库失败: %v", err)
				} else {
					logging.Logger.Infof("漏洞数据已入库: %s - %s", vulnData.VulnType, vulnData.Target)
				}
			}(vuln)

			wailsApp.Event.Emit("VulMessage", vuln)
		case scanMsg := <-httpx.RequestScanMsgChannel:
			wailsApp.Event.Emit("RequestScanMsg", scanMsg)
		case httpMarker := <-conf.HttpMarkerChan:
			wailsApp.Event.Emit("HttpMarker", httpMarker)
			db.UpdateMarker(httpMarker.Id, httpMarker.Color, httpMarker.Note)
		}
	}
}

// GetVulnerabilities 获取漏洞列表
// projectID: 项目ID，传空字符串或"all"获取所有项目的漏洞
// source: 来源过滤，传空字符串或"all"获取所有来源
// limit: 限制数量，0表示不限制
// offset: 偏移量
func (a *App) GetVulnerabilities(projectID string, source string, limit int, offset int) Result {
	vulnerabilities, err := db.GetAllVulnerabilities(projectID, source, limit, offset)
	if err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: vulnerabilities}
}

// GetVulnerabilityStats 获取漏洞统计信息
// projectID: 项目ID，传空字符串或"all"获取所有项目的统计
func (a *App) GetVulnerabilityStats(projectID string) Result {
	stats, err := db.GetVulnerabilityStatistics(projectID)
	if err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: stats}
}

// ClearVulnerabilities 清空漏洞数据
func (a *App) ClearVulnerabilities() Result {
	err := db.ClearAllVulnerabilities()
	if err != nil {
		return Result{Error: err.Error()}
	}
	return Result{Data: "漏洞数据已清空"}
}

// OpenConfigDir 打开配置目录
func (a *App) OpenConfigDir() error {
	return utils.OpenFolder(file.ChyingDir)
}

// githubRelease GitHub Release API 响应结构
type githubRelease struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
}

// CheckForUpdates 检查版本更新
func (a *App) CheckForUpdates() Result {
	updateInfo, err := checkGitHubRelease()
	if err != nil {
		logging.Logger.Warnf("版本检查失败: %v", err)
		return Result{Error: fmt.Sprintf("版本检查失败: %v", err)}
	}
	return Result{Data: updateInfo}
}

// GetCurrentVersion 获取当前版本号
func (a *App) GetCurrentVersion() string {
	return conf.Version
}

// autoCheckForUpdates 自动检查版本更新（后台静默执行）
func (a *App) autoCheckForUpdates() {
	// 延迟 3 秒，等待前端完全加载
	time.Sleep(3 * time.Second)

	updateInfo, err := checkGitHubRelease()
	if err != nil {
		logging.Logger.Warnf("自动版本检查失败: %v", err)
		return
	}

	if updateInfo.HasUpdate {
		logging.Logger.Infof("发现新版本: %s (当前: %s)", updateInfo.LatestVersion, updateInfo.CurrentVersion)
		if wailsApp != nil {
			wailsApp.Event.Emit("UpdateAvailable", updateInfo)
		}
	} else {
		logging.Logger.Infof("当前已是最新版本: %s", updateInfo.CurrentVersion)
	}
}

// checkGitHubRelease 请求 GitHub API 获取最新 Release 并对比版本
func checkGitHubRelease() (*UpdateInfo, error) {
	client := reqv3.C().SetTimeout(10 * time.Second).SetUserAgent("ChYing-UpdateChecker")

	resp, err := client.R().Get("https://api.github.com/repos/yhy0/CHYing/releases/latest")
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API 返回状态码: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.Unmarshal(resp.Bytes(), &release); err != nil {
		return nil, fmt.Errorf("解析 GitHub Release 响应失败: %w", err)
	}

	if release.TagName == "" {
		return nil, fmt.Errorf("GitHub Release 中未找到版本标签")
	}

	// 清理版本号前缀 v
	latestTag := strings.TrimPrefix(release.TagName, "v")
	currentTag := strings.TrimPrefix(conf.Version, "v")

	latestVer, err := version.NewVersion(latestTag)
	if err != nil {
		return nil, fmt.Errorf("解析最新版本号 '%s' 失败: %w", release.TagName, err)
	}

	currentVer, err := version.NewVersion(currentTag)
	if err != nil {
		return nil, fmt.Errorf("解析当前版本号 '%s' 失败: %w", conf.Version, err)
	}

	updateInfo := &UpdateInfo{
		HasUpdate:      latestVer.GreaterThan(currentVer),
		CurrentVersion: conf.Version,
		LatestVersion:  release.TagName,
		ReleaseURL:     release.HTMLURL,
		ReleaseNotes:   release.Body,
		PublishedAt:    release.PublishedAt,
	}

	return updateInfo, nil
}

// OpenURL 在系统默认浏览器中打开 URL
func (a *App) OpenURL(url string) Result {
	if err := browser.OpenURL(url); err != nil {
		logging.Logger.Errorf("打开 URL 失败: %v", err)
		return Result{Error: fmt.Sprintf("打开 URL 失败: %v", err)}
	}
	return Result{Data: "OK"}
}
