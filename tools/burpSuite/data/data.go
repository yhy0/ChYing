package data

import "sync"

/**
  @author: yhy
  @since: 2023/4/24
  @desc: //TODO
**/

type HTTPHistory struct {
	Id        int    `json:"id"`
	Host      string `json:"host"`
	Method    string `json:"method"`
	URL       string `json:"url"`
	Params    string `json:"params"`
	Edited    string `json:"edited"`
	Status    string `json:"status"`
	Length    string `json:"length"`
	MIMEType  string `json:"mime_type"`
	Extension string `json:"extension"`
	Title     string `json:"title"`
	Comment   string `json:"comment"`
	TLS       string `json:"tls"`
	IP        string `json:"ip"`
	Cookies   string `json:"cookies"`
	Time      string `json:"time"`
}

type SMap struct {
	sync.RWMutex
	Map map[int]*HTTPBody
}

func (l *SMap) WriteMap(key int, value *HTTPBody) {
	l.Lock()
	l.Map[key] = value
	l.Unlock()
}

func (l *SMap) ReadMap(key int) *HTTPBody {
	l.RLock()
	value, _ := l.Map[key]
	l.RUnlock()
	return value
}

var HttpHistory chan HTTPHistory

var HTTPId chan int
var HTTPBodyMap *SMap

// RepeaterBodyMap Repeater 中回退、前进使用
var RepeaterBodyMap map[string]map[int]*HTTPBody

type HTTPBody struct {
	TargetUrl string `json:"targetUrl"`
	Request   string `json:"request"`
	Response  string `json:"response"`
	UUID      string `json:"uuid"`
}

func init() {
	HttpHistory = make(chan HTTPHistory, 1)

	HTTPId = make(chan int, 1)

	HTTPBodyMap = &SMap{
		Map: make(map[int]*HTTPBody),
	}

	RepeaterBodyMap = make(map[string]map[int]*HTTPBody)
}
