package weblogic

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
)

func CVE_2018_2894(url string, client *httpx.Client) bool {
	if req, err := client.Request(url+"/ws_utc/begin.do", "GET", "", nil, "Poc WebLogic"); err == nil {
		if req2, err2 := client.Request(url+"/ws_utc/config.do", "GET", "", nil, "Poc WebLogic"); err2 == nil {
			if req.StatusCode == 200 || req2.StatusCode == 200 {
				logging.Logger.Infoln("[Vulnerable] CVE_2018_2894 ", url)
				return true
			}
		}
	}
	logging.Logger.Debugln("[Safety] CVE_2018_2894 ", url)
	return false
}
