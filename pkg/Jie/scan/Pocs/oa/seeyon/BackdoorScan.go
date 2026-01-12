package seeyon

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
)

// test233.jsp pass:rebeyond
// qwerasdf.jsp?pwd=zhengbianlu&cmd=cmd+/c+whoami
// SeeyonUpdate.jspx pass:rebeyond
// test123456.jsp?pwd=asasd3344&cmd=cmd+/c+whoami
// qwer960452.jsp?pwd=el38A9485&cmd=cmd+/c+whoami
// a234.jspx pass:rebeyond
// test10086.jsp 蚁剑密码: test
// peiqi10086.jsp 蚁剑密码: peiqi

func BackdoorScan(u string, client *httpx.Client) bool {
	backurls := []string{"/seeyon/test233.jsp", "/seeyon/SeeyonUpdate.jspx", "/seeyon/SeeyonUpdate1.jspx", "/seeyon/test123456.jsp", "/seeyon/test1234567.jsp", "/seeyon/qwerasdf.jsp", "/seeyon/qwer960452.jsp", "/seeyon/ping123456.jsp", "/seeyon/common/designer/pageLayout/test233.jsp", "/seeyon/common/designer/pageLayout/test10086.jsp", "/seeyon/common/designer/pageLayout/a234.jspx", "/seeyon/common/designer/pageLayout/peiqi10086.jsp"}
	for _, backurl := range backurls {
		if req, err := client.Request(u+backurl, "GET", "", nil, "Poc SeeYon"); err == nil {
			if req.StatusCode == 200 && (!strings.Contains(req.Body, "error") || strings.Contains(req.Body, "java.lang.NullPointerException")) && !strings.Contains(req.Body, "Burp") {
				return true
			}
		}
	}
	return false
}
