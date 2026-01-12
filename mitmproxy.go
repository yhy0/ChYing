package main

import (
	"encoding/json"
	"fmt"

	"github.com/iancoleman/orderedmap"
	uuid "github.com/satori/go.uuid"
	"github.com/yhy0/ChYing/mitmproxy"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/collection"
	"github.com/yhy0/ChYing/pkg/db"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/8/12
   @desc //TODO
**/

// RepeaterData  CurrentRepeater 中回退、前进使用
var RepeaterData *orderedmap.OrderedMap
var CurrentRepeater map[string]int

func init() {
	RepeaterData = orderedmap.New()
	CurrentRepeater = make(map[string]int)
}

// SetProxyInterceptMode 设置代理拦截模式 (适配新的 proxify 库)
func (a *App) SetProxyInterceptMode(enable bool) {
	// 启用/禁用请求和响应拦截
	mitmproxy.SetInterceptMode(true, enable)  // 拦截请求
	mitmproxy.SetInterceptMode(false, enable) // 拦截响应
}

// ForwardProxyInterceptData 转发拦截的数据 (适配新的 proxify 库)
func (a *App) ForwardProxyInterceptData(id string, modifiedBody string, action string) {
	mitmproxy.ForwardInterceptedData(id, modifiedBody, action)
}

func (a *App) GetHistoryAll() []*mitmproxy.HTTPHistory {
	var res []*mitmproxy.HTTPHistory

	// 检查数据库是否已初始化
	if db.GlobalDB == nil {
		logging.Logger.Warnln("数据库未初始化，返回空历史记录")
		return res
	}

	// 优先从当前项目获取历史记录，如果没有则获取所有项目的记录
	projectID := "default" // 可以根据当前项目设置

	history, err := db.GetAllHistory(projectID, "", 1000, 0) // 获取指定项目的前1000条记录
	if err != nil {
		logging.Logger.Errorln("Failed to get history:", err)
		// 如果获取指定项目失败，尝试获取所有项目的记录
		history, err = db.GetAllHistory("", "", 1000, 0)
		if err != nil {
			logging.Logger.Errorln("Failed to get all history:", err)
			return res
		}
	}

	if history == nil {
		return res
	}

	for _, item := range history {
		httpHistory := &mitmproxy.HTTPHistory{
			Id:          item.Hid,
			Host:        item.Host,
			Method:      item.Method,
			FullUrl:     item.FullUrl,
			Path:        item.Path,
			Status:      item.Status,
			Length:      item.Length,
			ContentType: item.ContentType,
			MIMEType:    item.MIMEType,  // ✅ 修复：添加MIME类型映射
			Extension:   item.Extension, // ✅ 修复：添加扩展名映射
			Title:       item.Title,     // ✅ 修复：添加标题映射
			IP:          item.IP,        // ✅ 修复：添加IP映射
			Color:       item.Color,
			Note:        item.Note, // ✅ 修复：添加备注映射
		}
		res = append(res, httpHistory)
	}

	logging.Logger.Infoln("GetHistoryAll: 成功获取", len(res), "条历史记录")
	return res
}

// GetHistoryDumpIndex 代理记录
func (a *App) GetHistoryDumpIndex(_url string) *mitmproxy.HTTPBody {
	id, ok := mitmproxy.GetHTTPUrlMapValue(_url)
	if !ok {
		return nil
	}
	_data, _ok := mitmproxy.HTTPBodyMap.Load(id)
	if _ok {
		return _data.(*mitmproxy.HTTPBody)
	}
	return nil
}

// GetHistoryDump 代理记录
func (a *App) GetHistoryDump(id int64) *mitmproxy.HTTPBody {
	lock.Lock()
	defer lock.Unlock()
	_data, _ok := mitmproxy.HTTPBodyMap.Load(id)
	if _ok {
		return _data.(*mitmproxy.HTTPBody)
	} else {
		return nil
	}

	// RepeaterData 每个 tab 的历史记录 使用另一个函数实现（目前前端保存了，）
	// val, ok := RepeaterData.Get(strconv.FormatInt(id, 10))
	// if !ok {
	// 	_data.UUID = strconv.FormatInt(id, 10)
	// 	RepeaterData.Set(_data.UUID, []data.HTTPBody{
	// 		_data,
	// 	})
	// 	CurrentRepeater[_data.UUID] = 0
	// } else {
	// 	_data = val.([]data.HTTPBody)[CurrentRepeater[_data.UUID]]
	// }
	//
	// return _data
}

// RawRequest Repeater 请求
func (a *App) RawRequest(request string, target string, _uuid string) *Result {
	// 说明第一次
	if _uuid == "" {
		logging.Logger.Debugln("_uuid")
		_uuid = uuid.NewV4().String()
	}

	resp, err := httpx.Raw(request, target)
	if err != nil {
		logging.Logger.Errorln("Repeater Raw err:", err)
		return &Result{
			Error: err.Error(),
		}
	}

	httpBody := mitmproxy.HTTPBody{
		TargetUrl:         target,
		RequestRaw:        request,
		ResponseRaw:       resp.ResponseDump,
		ResponseTimestamp: resp.ServerDurationMs,
		ServerDurationMs:  resp.ServerDurationMs, // 添加前端期望的字段
	}

	logging.Logger.Debugln("Repeater Raw:", target, resp.StatusCode, _uuid)

	return &Result{
		Data: httpBody,
	}
}

func (a *App) RawRequestForward(_uuid string) *Result {
	lock.Lock()
	defer lock.Unlock()
	id, ok := CurrentRepeater[_uuid]
	if !ok {
		return &Result{ // 还没有手动请求
			Error: "null",
		}
	}

	val, _ := RepeaterData.Get(_uuid)
	if id+1 >= len(val.([]mitmproxy.HTTPBody)) { // 已经是最后一个了
		return &Result{
			Error: "null",
		}
	}
	// 当前加一
	CurrentRepeater[_uuid] = id + 1
	return &Result{
		Data: val.([]mitmproxy.HTTPBody)[id+1],
	}
}

func (a *App) RawRequestBack(_uuid string) *Result {
	lock.Lock()
	defer lock.Unlock()
	id, ok := CurrentRepeater[_uuid]
	if !ok {
		return &Result{ // 还没有手动请求
			Error: "null",
		}
	}

	val, _ := RepeaterData.Get(_uuid)
	if id-1 < 0 { // 已经是最后一个了
		return &Result{
			Error: "null",
		}
	}
	// 当前减一
	CurrentRepeater[_uuid] = id - 1
	return &Result{
		Data: val.([]mitmproxy.HTTPBody)[id-1],
	}
}

func (a *App) GetStep(_uuid string) []int {
	lock.Lock()
	defer lock.Unlock()
	val, ok := RepeaterData.Get(_uuid)
	if !ok {
		return []int{0, 0}
	}
	return []int{CurrentRepeater[_uuid], len(val.([]mitmproxy.HTTPBody)) - 1}
}

func (a *App) GetStepValue(_uuid string, step int) *Result {
	lock.Lock()
	defer lock.Unlock()

	val, ok := RepeaterData.Get(_uuid)
	if !ok {
		return &Result{
			Error: "null",
		}
	}
	return &Result{
		Data: val.([]mitmproxy.HTTPBody)[step],
	}
}

// GetRepeater 每次切换 tab 时，重新取值
func (a *App) GetRepeater() []mitmproxy.HTTPBody {
	lock.Lock()
	defer lock.Unlock()
	var res []mitmproxy.HTTPBody
	for _, k := range RepeaterData.Keys() {
		val, ok := RepeaterData.Get(k)
		if !ok {
			continue
		}
		id := CurrentRepeater[k]
		res = append(res, val.([]mitmproxy.HTTPBody)[id])
	}

	return res
}

func (a *App) UpdateRepeaterTabId(_uuid string, id int) {
	lock.Lock()
	defer lock.Unlock()
	CurrentRepeater[_uuid] = id
}

type AttackData struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
	Len  int    `json:"len"`
}

// Intruder 处理 Intruder 传来的参数
func (a *App) Intruder(target string, req string, payloads string, attackType string, len int, tabId string) {
	name := fmt.Sprintf("Attack-%s", tabId)
	wailsApp.Event.Emit("Attack-Data", AttackData{
		Name: name,
		UUID: tabId,
		Len:  len,
	})
	logging.Logger.Debugln(target)

	logging.Logger.Debugln("========")
	logging.Logger.Debugln(payloads)
	logging.Logger.Debugln("========")
	logging.Logger.Debugln(attackType)
	// 解析 JSON 到结构体切片
	var items []mitmproxy.PayloadItem
	err := json.Unmarshal([]byte(payloads), &items)
	if err != nil {
		logging.Logger.Errorln("解析 JSON 出错:", err)
		return
	}

	mitmproxy.Intruder(target, req, items, attackType, tabId)
}

// GetAttackDump Intruder attack 记录
func (a *App) GetAttackDump(_uuid string, id int64) mitmproxy.HTTPBody {
	logging.Logger.Debugln("GetAttackDump", _uuid, id)
	if smap, ok := mitmproxy.GetIntruderMap(_uuid); ok {
		return smap.ReadMap(id)
	}
	return mitmproxy.HTTPBody{}
}

func (a *App) Jsluice(body string) map[string]interface{} {
	result := make(map[string]interface{})
	apis := collection.Api("", body, "application/javascript")
	urls, secrets := utils.Jsluice(body, apis)
	result["urls"] = urls
	result["secrets"] = secrets

	return result
}

// ClearMemoryHistoryData 清除所有存储的HTTP历史记录相关数据，但不重置ID生成器。
// 这个函数应该由前端在用户请求清除所有历史时调用。
func (a *App) ClearMemoryHistoryData() {
	// 清空 HTTPBodyMap
	mitmproxy.HTTPBodyMap.Range(func(key, value interface{}) bool {
		mitmproxy.HTTPBodyMap.Delete(key)
		return true
	})

	// 清空 HTTPUrlMap
	mitmproxy.ClearHTTPUrlMap()

	// 清空临时缓存（以防万一）
	mitmproxy.TempHistoryCache.Range(func(key, value interface{}) bool {
		mitmproxy.TempHistoryCache.Delete(key)
		return true
	})
	mitmproxy.TempRequestRawCache.Range(func(key, value interface{}) bool {
		mitmproxy.TempRequestRawCache.Delete(key)
		return true
	})

	logging.Logger.Infof("All history data (HTTPBodyMap, HTTPUrlMap, TempCaches) cleared. HistoryItemIDGenerator state remains: %d", mitmproxy.HistoryItemIDGenerator.Load())
}

// GetAuthorizationCheckStatus 获取越权检测状态
func (a *App) GetAuthorizationCheckStatus() bool {
	return mitmproxy.AuthorizationCheckEnabled.Load()
}
