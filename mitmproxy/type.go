package mitmproxy

import (
	"sync"
)

/**
  @author: yhy
  @since: 2023/5/7
  @desc: //TODO
**/

var EventDataChan = make(chan *EventData, 9999)

type EventData struct {
	Name string `json:"name"`
	Data any    `json:"data"`
}

type HTTPHistory struct {
	Id                int64  `json:"id"`
	FlowID            string `json:"flow_id,omitempty"`
	Host              string `json:"host"`
	Method            string `json:"method"`
	FullUrl           string `json:"full_url"`
	Path              string `json:"path"`
	Status            string `json:"status"`
	Length            string `json:"length"`
	ContentType       string `json:"content_type"`
	MIMEType          string `json:"mime_type"`
	Extension         string `json:"extension"`
	Title             string `json:"title"`
	IP                string `json:"ip"`
	Note              string `json:"note"`
	Color             string `json:"color"`
	Timestamp         string `json:"timestamp,omitempty"`
	ResponseTimestamp string `json:"response_timestamp,omitempty"`
}

type HTTPBody struct {
	Id                int64   `json:"id"`
	FlowID            string  `json:"flow_id,omitempty"`
	Title             string  `json:"title"`
	TargetUrl         string  `json:"targetUrl"`
	RequestRaw        string  `json:"request_raw"`
	ResponseRaw       string  `json:"response_raw"`
	ResponseTimestamp float64 `json:"response_timestamp,omitempty"`
	ServerDurationMs  float64 `json:"server_duration_ms,omitempty"` // 添加前端期望的字段
}

type IntruderRes struct {
	Id        int64    `json:"id"`
	Payload   []string `json:"payload"`
	Status    int      `json:"status"`    // 改为数字类型
	Length    int      `json:"length"`    // 改为数字类型
	TimeMs    int      `json:"timeMs"`    // 添加响应时间字段
	Timestamp int64    `json:"timestamp"` // 添加时间戳字段
}

type SMap struct {
	sync.RWMutex
	Map map[int64]HTTPBody
}

func (l *SMap) WriteMap(key int64, value HTTPBody) {
	l.Lock()
	l.Map[key] = value
	l.Unlock()
}

func (l *SMap) ReadMap(key int64) HTTPBody {
	l.RLock()
	value := l.Map[key]
	l.RUnlock()
	return value
}

// GetIntruderMap 安全地获取 IntruderMap 中的 SMap
func GetIntruderMap(uuid string) (*SMap, bool) {
	value, ok := IntruderMap.Load(uuid)
	if !ok {
		return nil, false
	}
	smap, ok := value.(*SMap)
	return smap, ok
}

// SetIntruderMap 安全地设置 IntruderMap 中的 SMap
func SetIntruderMap(uuid string, smap *SMap) {
	IntruderMap.Store(uuid, smap)
}

// GetOrCreateIntruderMap 获取或创建 IntruderMap 中的 SMap
func GetOrCreateIntruderMap(uuid string) *SMap {
	value, _ := IntruderMap.LoadOrStore(uuid, &SMap{
		Map: make(map[int64]HTTPBody),
	})
	return value.(*SMap)
}

// GetHTTPUrlMapValue 安全地获取 HTTPUrlMap 中的值
func GetHTTPUrlMapValue(url string) (int64, bool) {
	value, ok := HTTPUrlMap.Load(url)
	if !ok {
		return 0, false
	}
	id, ok := value.(int64)
	return id, ok
}

// SetHTTPUrlMapValue 安全地设置 HTTPUrlMap 中的值
func SetHTTPUrlMapValue(url string, id int64) {
	HTTPUrlMap.Store(url, id)
}

// ClearHTTPUrlMap 清空 HTTPUrlMap
func ClearHTTPUrlMap() {
	HTTPUrlMap.Range(func(key, value interface{}) bool {
		HTTPUrlMap.Delete(key)
		return true
	})
}
