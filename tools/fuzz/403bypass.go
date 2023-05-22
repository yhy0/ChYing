package fuzz

import (
	"fmt"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/ChYing/tools"
	"github.com/yhy0/logging"
	"net/url"
	"strings"
	"unicode"
)

/**
	@author: yhy
	@since: 2023/5/6
	@desc: bypass 403 	https://github.com/devploit/dontgo403
	todo 应该考虑一下带参数的，现在不能处理带参数的，直接拼接了，不好
**/

func Bypass403(uri, m string) {
	if !strings.HasSuffix(uri, "/") {
		uri += "/"
	}

	if m == "" {
		m = "GET"
	} else {
		m = strings.ToUpper(m)
	}

	result := method(uri, m)
	if result != nil {
		FuzzChan <- tools.Result{
			Url:           result.Url,
			Method:        m,
			StatusCode:    result.StatusCode,
			ContentLength: result.ContentLength,
			Request:       result.Request,
			Response:      result.Response,
		}
		return
	}

	result = headers(uri, m)
	if result != nil {
		FuzzChan <- tools.Result{
			Url:           result.Url,
			Method:        m,
			StatusCode:    result.StatusCode,
			ContentLength: result.ContentLength,
			Request:       result.Request,
			Response:      result.Response,
		}
		return
	}

	result = endPaths(uri, m)
	if result != nil {
		FuzzChan <- tools.Result{
			Url:           result.Url,
			Method:        m,
			StatusCode:    result.StatusCode,
			ContentLength: result.ContentLength,
			Request:       result.Request,
			Response:      result.Response,
		}
		return
	}

	result = midPaths(uri, m)
	if result != nil {
		FuzzChan <- tools.Result{
			Url:           result.Url,
			Method:        m,
			StatusCode:    result.StatusCode,
			ContentLength: result.ContentLength,
			Request:       result.Request,
			Response:      result.Response,
		}
		return
	}

	result = capital(uri, m)
	if result != nil {
		FuzzChan <- tools.Result{
			Url:           result.Url,
			Method:        m,
			StatusCode:    result.StatusCode,
			ContentLength: result.ContentLength,
			Request:       result.Request,
			Response:      result.Response,
		}
		return
	}

	result = http10(uri, m)
	if result != nil {
		FuzzChan <- tools.Result{
			Url:           result.Url,
			Method:        m,
			StatusCode:    result.StatusCode,
			ContentLength: result.ContentLength,
			Request:       result.Request,
			Response:      result.Response,
		}
		return
	}

	return
}

// method 通过更改请求方法，尝试绕过 403
func method(uri, m string) *tools.Result {
	logging.Logger.Infoln(file.Bypass403["httpmethods.txt"])
	ch := make(chan struct{}, 5)
	result := &tools.Result{}
	var flag = false
	for _, line := range file.Bypass403["httpmethods.txt"] {
		if m == line {
			continue
		}
		if flag {
			break
		}
		ch <- struct{}{}
		go func(line string) {
			resp, err := httpx.Request(uri, line, "", false, nil)
			if err != nil {
				<-ch
				return
			}
			<-ch
			if resp != nil && resp.StatusCode == 200 {
				flag = true
				result = &tools.Result{
					Url:           uri,
					Method:        line,
					StatusCode:    resp.StatusCode,
					ContentLength: resp.ContentLength,
					Request:       resp.RequestDump,
					Response:      resp.ResponseDump,
				}
				return
			}
		}(line)
	}
	close(ch)
	if flag {
		return result
	}
	return nil
}

// headers 通过添加header，尝试绕过 403
func headers(uri, m string) *tools.Result {
	ch := make(chan struct{}, 10)
	result := &tools.Result{}
	var flag = false

	for _, ip := range file.Bypass403["ips.txt"] {
		for _, line := range file.Bypass403["headers.txt"] {
			if flag {
				break
			}
			ch <- struct{}{}
			go func(ip, line string) {
				header := make(map[string]string)
				header[line] = ip

				resp, err := httpx.Request(uri, m, "", false, header)
				if err != nil {
					<-ch
					return
				}
				<-ch
				if resp != nil && resp.StatusCode == 200 {
					flag = true
					result = &tools.Result{
						Url:           uri,
						Method:        m,
						StatusCode:    resp.StatusCode,
						ContentLength: resp.ContentLength,
						Request:       resp.RequestDump,
						Response:      resp.ResponseDump,
					}
					return
				}
			}(ip, line)
		}

	}

	if flag {
		return result
	}

	for _, line := range file.Bypass403["simpleheaders.txt"] {
		if flag {
			break
		}
		ch <- struct{}{}
		go func(line string) {
			x := strings.Split(line, " ")
			header := make(map[string]string)
			header[x[0]] = x[1]
			resp, err := httpx.Request(uri, m, "", false, header)
			if err != nil {
				<-ch
				return
			}
			<-ch
			if resp != nil && resp.StatusCode == 200 {
				flag = true
				result = &tools.Result{
					Url:           uri,
					Method:        m,
					StatusCode:    resp.StatusCode,
					ContentLength: resp.ContentLength,
					Request:       resp.RequestDump,
					Response:      resp.ResponseDump,
				}
				return
			}
		}(line)
	}

	if flag {
		return result
	}
	close(ch)
	return nil
}

// endPaths 通过添加 path 后缀，尝试绕过 403
func endPaths(uri, m string) *tools.Result {
	ch := make(chan struct{}, 5)
	result := &tools.Result{}
	var flag = false
	for _, line := range file.Bypass403["endpaths.txt"] {
		if flag {
			break
		}
		ch <- struct{}{}
		go func(line string) {
			resp, err := httpx.Request(uri+line, m, "", false, nil)
			if err != nil {
				<-ch
				return
			}
			<-ch
			if resp != nil && resp.StatusCode == 200 {
				flag = true
				result = &tools.Result{
					Url:           uri + line,
					Method:        m,
					StatusCode:    resp.StatusCode,
					ContentLength: resp.ContentLength,
					Request:       resp.RequestDump,
					Response:      resp.ResponseDump,
				}
				return
			}
		}(line)
	}
	close(ch)
	if flag {
		return result
	}
	return nil
}

// midPaths 在 path 路径中间添加字符，尝试绕过 403
func midPaths(uri, m string) *tools.Result {
	ch := make(chan struct{}, 5)
	result := &tools.Result{}
	var flag = false

	x := strings.Split(uri, "/")
	var uripath string

	if uri[len(uri)-1:] == "/" {
		uripath = x[len(x)-2]
	} else {
		uripath = x[len(x)-1]
	}

	baseuri := strings.ReplaceAll(uri, uripath, "")
	baseuri = baseuri[:len(baseuri)-1]

	for _, line := range file.Bypass403["midpaths.txt"] {
		if flag {
			break
		}
		ch <- struct{}{}
		go func(line string) {
			var fullpath string
			if uri[len(uri)-1:] == "/" {
				fullpath = baseuri + line + uripath + "/"
			} else {
				fullpath = baseuri + "/" + line + uripath
			}

			resp, err := httpx.Request(fullpath, m, "", false, nil)
			if err != nil {
				<-ch
				return
			}
			<-ch
			if resp != nil && resp.StatusCode == 200 {
				flag = true
				result = &tools.Result{
					Url:           fullpath,
					Method:        m,
					StatusCode:    resp.StatusCode,
					ContentLength: resp.ContentLength,
					Request:       resp.RequestDump,
					Response:      resp.ResponseDump,
				}
				return
			}
		}(line)
	}
	close(ch)
	if flag {
		return result
	}
	return nil
}

// capital 通过将URI最后部分中的每个字母大写, 尝试绕过 403
func capital(uri, m string) *tools.Result {
	ch := make(chan struct{}, 5)
	result := &tools.Result{}
	var flag = false
	x := strings.Split(uri, "/")
	var uripath string

	if uri[len(uri)-1:] == "/" {
		uripath = x[len(x)-2]
	} else {
		uripath = x[len(x)-1]
	}
	baseuri := strings.ReplaceAll(uri, uripath, "")
	baseuri = baseuri[:len(baseuri)-1]

	for _, z := range uripath {
		if flag {
			break
		}
		ch <- struct{}{}
		go func(z rune) {
			newpath := strings.Map(func(r rune) rune {
				if r == z {
					return unicode.ToUpper(r)
				} else {
					return r
				}
			}, uripath)

			var fullpath string
			if uri[len(uri)-1:] == "/" {
				fullpath = baseuri + newpath + "/"
			} else {
				fullpath = baseuri + "/" + newpath
			}

			resp, err := httpx.Request(fullpath, m, "", false, nil)
			if err != nil {
				<-ch
				return
			}
			<-ch
			if resp != nil && resp.StatusCode == 200 {
				flag = true
				result = &tools.Result{
					Url:           fullpath,
					Method:        m,
					StatusCode:    resp.StatusCode,
					ContentLength: resp.ContentLength,
					Request:       resp.RequestDump,
					Response:      resp.ResponseDump,
				}
				return
			}
		}(z)
	}
	close(ch)
	if flag {
		return result
	}
	return nil
}

func http10(uri, m string) *tools.Result {
	u, err := url.Parse(uri)
	if err != nil {
		logging.Logger.Errorln("Error url.Parse:", err)
		return nil
	}
	// 设置请求行和请求头
	raw := fmt.Sprintf("GET %s HTTP/1.0\r\n"+
		"\r\n"+
		"\r\n", u.Path+"?"+u.RawQuery)

	resp, err := httpx.Request10(u.Host, raw)
	if err != nil {
		return nil
	}
	if resp != nil && resp.StatusCode == 200 {
		return &tools.Result{
			Url:           uri,
			Method:        "GET",
			StatusCode:    resp.StatusCode,
			ContentLength: resp.ContentLength,
			Request:       resp.RequestDump,
			Response:      resp.ResponseDump,
		}
	}

	return nil
}
