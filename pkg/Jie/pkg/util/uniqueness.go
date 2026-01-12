package util

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/yhy0/logging"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

/**
   @author yhy
   @since 2023/11/3
   @desc 用来判断是否扫描过
**/

// GenerateUniqueID 生成唯一id(md5), 用来判断是否扫描过 空代表目前逻辑还不支持判断，可以看成没有扫描过
func GenerateUniqueID(req *http.Request) string {
	// 目前暂时不处理文件上传请求， 还没想好怎么测试 自动修改文件名后缀，然后判断返回包里是否包含修改后的后缀？
	// TODO
	contentType := req.Header.Get("Content-Type")
	isFileUpload := strings.Contains(contentType, "multipart/form-data")
	if isFileUpload {
		return ""
	}
	// 为请求生成唯一标识符
	key, err := getRequestKey(req)
	if err != nil {
		// 这种可能是加密的，导致提取不到参数
		if !strings.Contains(err.Error(), "invalid semicolon separator in query") {
			logging.Logger.Errorln(err)
		}
	}

	return key
}

// TODO 对于 ThinkPHP 那种控制器的，这种去重就不行了 index.php?c=search&args=xxxx
// getRequestKey 生成请求的唯一标识符 用于判断是否扫描过 计算方式为 请求方法 + URL（不包括查询参数） + 查询参数名称
func getRequestKey(req *http.Request) (string, error) {
	var host string
	// 不能对原值修改，原值是一个指针
	if req.URL.Scheme == "http" && strings.HasSuffix(req.URL.Host, ":80") {
		host = strings.TrimRight(req.URL.Host, ":80")
	} else if req.URL.Scheme == "https" && strings.HasSuffix(req.URL.Host, ":443") {
		host = strings.TrimRight(req.URL.Host, ":443")
	} else {
		host = req.URL.Host
	}

	// 将请求方法和 URL（不包括查询参数）连接在一起
	data := req.Method + req.URL.Scheme + "://" + host + req.URL.Path

	// 读取响应体
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logging.Logger.Errorln(err)
	}
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	// 提取查询参数名
	paramNames, err := GetReqParameters(req.Method, req.Header.Get("Content-Type"), req.URL, body, nil)

	if err != nil {
		hash := md5.Sum([]byte(data))
		return hex.EncodeToString(hash[:]), err
	}

	// 对查询参数名称进行排序，以确保相同的参数集合具有相同的哈希值
	sort.Strings(paramNames)

	// 将排序后的参数名称连接在一起并添加到数据字符串中
	data += strings.Join(paramNames, "")

	// 计算 MD5 哈希值
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

// SimpleUniqueId 只对 http://testphp.vulnweb.com/redir.php?r=https://beautybeer.blogspot.com/ 这种简单的进行获取唯一 id
func SimpleUniqueId(u string) string {
	parseUrl, err := url.Parse(u)
	if err != nil {
		return ""
	}

	if parseUrl.Scheme == "http" && strings.HasSuffix(parseUrl.Host, ":80") {
		parseUrl.Host = strings.TrimRight(parseUrl.Host, ":80")
	} else if parseUrl.Scheme == "https" && strings.HasSuffix(parseUrl.Host, ":443") {
		parseUrl.Host = strings.TrimRight(parseUrl.Host, ":443")
	}

	// 将请求方法和 URL（不包括查询参数）连接在一起
	data := parseUrl.Scheme + "://" + parseUrl.Host + parseUrl.Path

	// 提取查询参数的名称  有的即使是 POST 请求，url请求路径中也会存在参数，所以这里全部都要提取
	var paramNames []string
	queryParams := parseUrl.Query()
	for paramName := range queryParams {
		paramNames = append(paramNames, paramName)
	}

	// 对查询参数名称进行排序，以确保相同的参数集合具有相同的哈希值
	sort.Strings(paramNames)

	// 将排序后的参数名称连接在一起并添加到数据字符串中
	data += strings.Join(paramNames, "")

	// 计算 MD5 哈希值
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
