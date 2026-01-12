package brute

import (
	"fmt"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
)

func JbossBrute(url string, client *httpx.Client) (username string, password string) {
	if req, err := client.Basic(url+"/jmx-console/", "GET", "", nil, "asdasdascsacacs", "asdasdascsacacs", "Brute"); err == nil {
		if req.StatusCode == 401 {
			for uspa := range jbossuserpass {
				if req2, err2 := client.Basic(url+"/jmx-console/", "GET", "", nil, jbossuserpass[uspa].Username, jbossuserpass[uspa].Password, "Brute"); err2 == nil {
					if req2.StatusCode == 200 || req2.StatusCode == 403 {
						logging.Logger.Infof(fmt.Sprintf("Found vuln Jboss password|%s:%s|%s\n", jbossuserpass[uspa].Username, jbossuserpass[uspa].Password, url))
						return jbossuserpass[uspa].Username, jbossuserpass[uspa].Password
					}
				}
			}
		}
	}
	return "", ""
}
