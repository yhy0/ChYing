package output

import "github.com/iancoleman/orderedmap"

/**
   @author yhy
   @since 2023/10/16
   @desc //TODO
**/

// 漏洞等级
var (
    Low      = "Low"
    Medium   = "Medium"
    High     = "High"
    Critical = "Critical"
)

type VulMessage struct {
    DataType string   `json:"data_type"`
    VulnData VulnData `json:"vul_data"`
    Plugin   string   `json:"plugin"`
    Level    string   `json:"level"`
}

type VulnData struct {
    CreateTime  string `json:"create_time"`
    VulnType    string `json:"vuln_type"`
    Target      string `json:"target"`
    Ip          string `json:"ip"`
    Method      string `json:"method"`
    Param       string `json:"param"`
    Payload     string `json:"payload"`
    CURLCommand string `json:"curl_command"`
    Description string `json:"description"`
    Request     string `json:"request"`
    Header      string `json:"header"`
    Response    string `json:"response"`
}

type SCopilotData struct {
    Target        string         `json:"target"`
    Ip            string         `json:"ip"`
    IpAddress     string         `json:"ip_address"`
    HostNoPort    string         `json:"host_no_port"`
    SiteMap       []string       `json:"site_map"`
    Fingerprints  []string       `json:"fingerprints"`
    VulMessage    []VulMessage   `json:"vul_message"`
    VulPlugin     map[string]int `json:"vul_plugin"`
    InfoMsg       []PluginMsg    `json:"info_msg"`
    InfoPlugin    map[string]int `json:"info_plugin"`
    PluginMsg     []PluginMsg    `json:"plugin_msg"`
    CollectionMsg Collection     `json:"collection_msg"`
}

type Collection struct {
    Subdomain   []string               `json:"subdomains"`
    OtherDomain []string               `json:"other_domains"`
    PublicIp    []string               `json:"public_ip"`
    InnerIp     []string               `json:"inner_ip"`
    Phone       []string               `json:"phone"`
    Email       []string               `json:"email"`
    IdCard      []string               `json:"id_card"`
    Others      []string               `json:"others"`
    Urls        []string               `json:"urls"`
    Api         []string               `json:"api"`
    Parameters  *orderedmap.OrderedMap `json:"parameters"`
}

// IsEmpty 判断 Collection 是否为空
func (c Collection) IsEmpty() bool {
    // 根据业务逻辑定义 "空" 的条件
    return len(c.Subdomain) == 0 && len(c.OtherDomain) == 0 && len(c.PublicIp) == 0 && len(c.InnerIp) == 0 && len(c.Phone) == 0 && len(c.Email) == 0 && len(c.IdCard) == 0 && len(c.Others) == 0 && len(c.Urls) == 0 && len(c.Api) == 0 && c.Parameters == nil
}

type PluginMsg struct {
    Url      string   `json:"url"`
    Plugin   string   `json:"plugin"`
    Result   []string `json:"result"` // 插件结果
    Request  string   `json:"request"`
    Response string   `json:"response"`
}

type SCopilotList struct {
    Host      string `json:"host"`
    InfoCount int    `json:"info_count"`
    ApiCount  int    `json:"api_count"`
    VulnCount int    `json:"vuln_count"`
}

type IPInfo struct {
    Ip          string         `json:"ip"`
    AllRecords  []string       `json:"all_records"`
    PortService map[int]string `json:"port_service"`
    Type        string         `json:"type"` // cdn 、waf、cloud
    Value       string         `json:"value"`
    Cdn         bool           `json:"cdn"`
}
