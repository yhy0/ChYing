package reverse

import (
	"encoding/json"
	"fmt"
	"github.com/thoas/go-funk"
	"github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"strings"
)

/**
  @author: yhy
  @since: 2023/5/17
  @desc: https://github.com/yumusb/DNSLog-Platform-Golang

    https://dig.pm 这个好可恶，不遵守 github 的开源文档的 api 规定
**/

type Dig struct {
	Domain     string `json:"domain"`
	Key        string `json:"key"`
	Token      string `json:"token"`
	FullDomain string `json:"fullDomain"`
	MainDomain string `json:"mainDomain"`
	SubDomain  string `json:"subDomain"`
	Msg        string
}

// GetSubDomain 获取子域名
func GetSubDomain() *Dig {
	conf.GlobalConfig.Reverse.Host = strings.TrimRight(conf.GlobalConfig.Reverse.Host, "/")
	resp, err := httpx.Request(conf.GlobalConfig.Reverse.Host+"/get_sub_domain", "POST", fmt.Sprintf("domain=%s", conf.GlobalConfig.Reverse.Domain), nil, "Dnslog")
	if err != nil {
		logging.Logger.Errorln(err)
		return nil
	}

	dig := &Dig{}
	err = json.Unmarshal([]byte(resp.Body), &dig)
	if err != nil {
		logging.Logger.Errorln(err)
		return nil
	}

	// 强制转一下
	if dig.SubDomain != "" {
		dig.Domain = dig.FullDomain
		dig.Key = dig.SubDomain
	}

	return dig
}

// PullLogs 获取dnslog日志
func PullLogs(dig *Dig) bool {
	if dig == nil {
		return false
	}
	resp, err := httpx.Request(conf.GlobalConfig.Reverse.Host+"/get_results", "POST", fmt.Sprintf("domain=%s&token=%s", dig.Domain, dig.Token), nil, "Dnslog")
	if err != nil {
		logging.Logger.Errorln(err)
		return false
	}

	if !funk.Contains(resp.Body, "null") && funk.Contains(resp.Body, dig.Key) {
		dig.Msg = resp.Body
		return true
	}

	return false
}
