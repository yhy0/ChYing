package burpSuite

/**
  @author: yhy
  @since: 2023/5/7
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

type HTTPBody struct {
	TargetUrl string `json:"targetUrl"`
	Request   string `json:"request"`
	Response  string `json:"response"`
	UUID      string `json:"uuid"`
}

type IntruderRes struct {
	Id      int      `json:"id"`
	Payload []string `json:"payload"`
	Status  string   `json:"status"`
	Length  string   `json:"length"`
}

type Setting struct {
	ProxyPort int      `json:"port"`
	Exclude   []string `json:"exclude"` // Exclude 排除显示的域名
	Include   []string `json:"include"`
}

// SettingUI 前端映射使用
type SettingUI struct {
	ProxyPort int    `json:"port"`
	Exclude   string `json:"exclude"` // Exclude 排除显示的域名
	Include   string `json:"include"`
}
