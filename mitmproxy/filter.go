package mitmproxy

import (
	"net"
	"regexp"
	"strings"

	"github.com/yhy0/ChYing/conf"
)

/**
   @author yhy
   @since 2024/10/19
   @desc 过滤数据，go-mitmproxy 写的过滤有点没有看懂，不起作用，流量还是有，自己实现过滤，每个插件的每个方法好像都要调用一遍这个函数
**/

func Filter(host string) bool {
	filter := false
	if len(conf.Config.Exclude) > 0 {
		if MatchHost(host, conf.Config.Exclude) { // 在过滤条件内，不走后续流程
			filter = true
		}
	}
	if !filter { // 过滤条件中没有这个，看看 Include 中是否设置，只要某些流量
		if len(conf.Config.Include) > 0 {
			if !MatchHost(host, conf.Config.Include) {
				filter = true // Include 中没有设置，不走后续流程
			} else {
				filter = false
			}
		}
	}

	return filter
}

// MatchHost detect hosts is match address
func MatchHost(address string, hosts []*conf.Scope) bool {
	hostname, _ := splitHostPort(address)
	for _, host := range hosts {
		if host.Enabled {
			if host.Regexp {
				if match, _ := regexp.MatchString(host.Prefix, hostname); match {
					return true
				}
			} else {
				if hostname == host.Prefix {
					return true
				}
			}

		}
	}
	return false
}

func splitHostPort(address string) (string, string) {
	// 使用 net.SplitHostPort 正确处理 IPv6 地址 (如 [::1]:8080)
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		// 如果解析失败，可能是没有端口的地址
		// 检查是否是 IPv6 地址（包含多个冒号但没有方括号）
		if strings.Count(address, ":") > 1 && !strings.HasPrefix(address, "[") {
			// 纯 IPv6 地址，没有端口
			return address, ""
		}
		// 尝试简单的最后一个冒号分割（用于 IPv4）
		index := strings.LastIndex(address, ":")
		if index == -1 {
			return address, ""
		}
		return address[:index], address[index+1:]
	}
	return host, port
}
