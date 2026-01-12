package util

import (
    "github.com/projectdiscovery/retryabledns"
    "github.com/yhy0/ChYing/pkg/Jie/lib/cdncheck"
    "github.com/yhy0/logging"
    "net"
)

/**
  @author: yhy
  @since: 2023/10/30
  @desc: 检测是否为 cdn
**/

var client = cdncheck.New()

func CheckCdn(target string) (cdn bool, value string, itemType string, dnsData *retryabledns.DNSData) {
    var err error
    
    ip := net.ParseIP(target)
    // 这种是域名
    if ip == nil {
        cdn, value, itemType, err, dnsData = client.CheckDomainWithFallback(target)
        if err != nil {
            logging.Logger.Errorln(err)
            return
        }
    } else {
        // 内网 ip 不检查 cdn
        if IsInnerIP(ip.String()) {
            cdn = false
        } else {
            // checks if an IP is contained in the cdn denylist
            cdn, value, itemType, err = client.Check(ip)
            if err != nil {
                logging.Logger.Errorln(err)
                return
            }
        }
    }
    
    if itemType == "cdn" || Contains(value, "cdn") {
        cdn = true
    }
    return
}
