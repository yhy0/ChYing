package input

import (
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "net/url"
)

/**
  @author: yhy
  @since: 2023/1/5
  @desc: //TODO
**/

// CrawlResult 如果是被动代理模式的结果，需要考虑到重复数据的问题，防止重复发payload
type CrawlResult struct {
    Id           string            `json:"id"`        // 和被动代理的 id 对应
    Target       string            `json:"target"`    // 这个是总目标，代表是哪个网站，并不是要扫描的链接 eg: http://testphp.vulnweb.com/
    Host         string            `json:"host"`      // 目标的 host 值，作为 key 用来区分，带端口的
    Url          string            `json:"url"`       // 这个才是代表要扫描的 url,  eg: http://testphp.vulnweb.com/comment.php?aid=1
    ParseUrl     *url.URL          `json:"parse_url"` // 解析后的 url
    Ip           string            `json:"ip"`        // 这里表示是 ip 或者 host ，都不带端口
    Cdn          bool              `json:"cdn"`
    Port         int               `json:"port"`
    UniqueId     string            `json:"unique_id"` // 唯一 ID，用来判断是否扫描
    Method       string            `json:"method"`    // 请求方法
    Headers      map[string]string `json:"headers"`   // 请求头
    RequestBody  string            `json:"request_body"`
    ContentType  string            `json:"content_type"`
    Resp         *httpx.Response   `json:"resp"`
    RawRequest   string            `json:"raw_request"`
    RawResponse  string            `json:"raw_response"`
    Fingerprints []string          `json:"fingerprints"` // 指纹，有的扫描插件需要匹配到指纹才会进行扫描
    Source       string            `json:"source"`       // 来源
    File         string            `json:"file"`
    Kv           string            `json:"kv"`          // 参数名和参数值  user=admin&password=admin
    ParamNames   []string          `json:"param_names"` // 请求中的参数名  user,password，
    Waf          []string          `json:"waf"`         // 是否存在 waf
    Archive      map[string]string `json:"archive"`     // 从 web.archive.org 获取到的历史 url
}
