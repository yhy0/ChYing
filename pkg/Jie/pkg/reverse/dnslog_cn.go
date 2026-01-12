package reverse

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
)

/**
  @author: yhy
  @since: 2023/11/18
  @desc: http://dnslog.cn/
**/

type DnslogCn struct {
	Domain  string
	Session string
}

func GetDnslogUrl() *DnslogCn {
	session := util.RandomLetterNumbers(8)

	headers := map[string]string{
		"Cookie": "PHPSESSID=" + session,
	}
	resp, err := httpx.Request("http://www.dnslog.cn/getdomain.php", "GET", "", headers, "Dnslog")

	if err != nil {
		logging.Logger.Errorln(err)
		return nil
	}

	return &DnslogCn{
		resp.Body,
		session,
	}
}

func GetDnslogRecord(session string) string {
	headers := map[string]string{
		"Cookie": "PHPSESSID=" + session,
	}
	resp, err := httpx.Request("http://www.dnslog.cn/getdomain.php", "GET", "", headers, "Dnslog")

	if err != nil {
		logging.Logger.Errorln(err)
		return ""
	}

	if resp.Body == "[]" {
		return ""
	}

	return resp.Body
}
