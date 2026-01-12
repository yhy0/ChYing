package jenkins

import (
	"fmt"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
)

func Unauthorized(u string, client *httpx.Client) bool {
	if req, err := client.Request(u, "GET", "", nil, "Poc Jenkins"); err == nil {
		if req.RespHeader.Get("X-Jenkins-Session") != "" {
			if req2, err := client.Request(u+"/script", "GET", "", nil, "Poc Jenkins"); err == nil {
				if req2.StatusCode == 200 && util.Contains(req2.Body, "Groovy script") {
					logging.Logger.Println(fmt.Sprintf("Found vuln Jenkins Unauthorized script|%s\n", u+"/script"))
					return true
				}
			}
			if req2, err := client.Request(u+"/computer/(master)/scripts", "GET", "", nil, "Poc Jenkins"); err == nil {
				if req2.StatusCode == 200 && util.Contains(req2.Body, "Groovy script") {
					logging.Logger.Println(fmt.Sprintf("Found vuln Jenkins Unauthorized script|%s\n", u+"/computer/(master)/scripts"))
					return true
				}
			}
		}
	}
	return false
}
