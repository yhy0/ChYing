package swagger

import (
	"bytes"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/thoas/go-funk"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/ChYing/tools"
	"github.com/yhy0/Jie/conf"
	"github.com/yhy0/Jie/pkg/input"
	"github.com/yhy0/Jie/pkg/output"
	"github.com/yhy0/Jie/pkg/util"
	"github.com/yhy0/Jie/scan/sqlmap"
	"github.com/yhy0/logging"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

/**
  @author: yhy
  @since: 2022/9/14
  @desc: 对 swagger api 进行未授权、ssrf、注入等测试
	参考 https://github.com/lijiejie/swagger-exp
**/

var Keys = []string{"location", "url", "path"}

type Parameter struct {
	Name  string
	In    string
	Value string
}

var SwaggerChan chan tools.Result

func init() {
	SwaggerChan = make(chan tools.Result, 1)
}

func Scan(target string) {
	// 每次扫描前都重新读取一遍配置字典
	file.ReadFiles()

	u, err := url.Parse(target)
	if err != nil {
		logging.Logger.Errorf("Scan url.Parse(%s) err: %v", target, err)
		return
	}

	host := u.Scheme + "://" + u.Host
	path := u.Path

	var schema = u.Scheme + "://"

	var basePath = ""

	index := strings.LastIndex(path, "/")
	if index == -1 {
		logging.Logger.Errorln(target, " 格式不正确")
		return
	}
	if index != 0 {
		basePath = path[:index]
	}
	req, err := httpx.Request(target, "GET", "", false, nil)
	if err != nil {
		logging.Logger.Errorln("Scan err: ", err)
		return
	}

	SwaggerChan <- tools.Result{
		Url:           target,
		Method:        "GET",
		StatusCode:    req.StatusCode,
		ContentLength: req.ContentLength,
		Request:       req.RequestDump,
		Response:      req.ResponseDump,
	}

	var (
		swaggerResources bool
		swaggerJson      bool
	)

	if funk.Contains(req.Body, "Swagger UI") { // 传入的是 /swagger-ui.html
		// 访问 /swagger-resources 获取 对应的 api 集合
		t := strings.ReplaceAll(target, "swagger-ui.html", "swagger-resources")
		req, err = httpx.Request(t, "GET", "", false, nil)
		if err != nil {
			logging.Logger.Errorln("Scan err: ", err)
			return
		}
		swaggerResources = true

		SwaggerChan <- tools.Result{
			Url:           t,
			Method:        "GET",
			StatusCode:    req.StatusCode,
			ContentLength: req.ContentLength,
			Request:       req.RequestDump,
			Response:      req.ResponseDump,
		}
	}

	if funk.Contains(req.Body, "swaggerVersion") {
		swaggerResources = true
	} else if funk.Contains(req.Body, "swagger") {
		swaggerJson = true
	}

	if !swaggerResources && !swaggerJson {
		logging.Logger.Warnf("Swagger parsing failed. Not find swaggerVersion/swagger")
		return
	}

	if swaggerResources { // 传入的是 /swagger_resources ,首先解析出所有版本的 Swagger api 路径 eg: /v2/api-docs, /v1/api-docs, /test/api
		var apis []string
		for _, key := range Keys {
			_, err = jsonparser.ArrayEach([]byte(req.Body), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				if err != nil {
					logging.Logger.Errorln(target, " jsonparser.ArrayEach error:", err.Error())
				}

				api, _, _, _ := jsonparser.Get(value, key)
				if len(api) > 0 {
					apis = append(apis, string(api))

				}
			})
			if len(apis) > 0 { // 说明找到路径了
				break
			}
		}
		if len(apis) > 0 {
			logging.Logger.Infof("Swagger-resources parsing success. %s find %d api doc: %v", target, len(apis), apis)
		} else {
			logging.Logger.Warnln("Swagger-resources parsing success. Not find api doc. Program Exit.")
			return
		}

		for _, api := range apis { // 获取全部 api 文档路径
			req, err = httpx.Request(host+basePath+api, "GET", "", false, nil)
			if err != nil {
				logging.Logger.Errorf("Scan(%s) err: %v", api, err)
				continue
			}

			SwaggerChan <- tools.Result{
				Url:           host + basePath + api,
				Method:        "GET",
				StatusCode:    req.StatusCode,
				ContentLength: req.ContentLength,
				Request:       req.RequestDump,
				Response:      req.ResponseDump,
			}

			if funk.Contains(req.Body, "\"swagger\":") { // 解析每个文档
				logging.Logger.Infof("Start test %s...", api)
				parseApi([]byte(req.Body), u.Host, basePath, schema)
			}
		}
	}

	if swaggerJson { // 表示传入的就是swagger api路径，直接解析 eg: /v2/api-docs
		req, err = httpx.Request(target, "GET", "", false, nil)
		if err != nil {
			logging.Logger.Errorf("Scan(%s) err: %v", target, err)
			return
		}

		SwaggerChan <- tools.Result{
			Url:           target,
			Method:        "GET",
			StatusCode:    req.StatusCode,
			ContentLength: req.ContentLength,
			Request:       req.RequestDump,
			Response:      req.ResponseDump,
		}

		if funk.Contains(req.Body, "\"swagger\":") { // 解析每个文档
			logging.Logger.Infof("Start test %s, please wait...", target)
			parseApi([]byte(req.Body), u.Host, basePath, schema)
		}
	}
}

// ParseApi 解析 swagger api 文档
func parseApi(body []byte, host, basePath, schema string) {
	hostByte, _, _, _ := jsonparser.Get(body, "host")

	if len(hostByte) > 0 {
		h := strings.Split(string(hostByte), ":")
		if !util.IsInnerIP(h[0]) {
			host = string(hostByte)
		}
	}
	basePathTmp, _, _, _ := jsonparser.Get(body, "basePath")

	if len(basePathTmp) > 0 {
		basePath = string(basePathTmp)
	}

	paths, _, _, _ := jsonparser.Get(body, "paths") // 要拼接的路径信息

	// $ref 参数中的这种参数的定义
	definitionsRef, _, _, _ := jsonparser.Get(body, "definitions") // 某些参数的定义
	definitionsMap := getDefinitions(definitionsRef)

	// $ref 参数中的这种参数的定义 , 这种可能会作为路径或参数的一部分
	parametersRef, _, _, _ := jsonparser.Get(body, "parameters") // 某些参数的定义
	parametersMap := getParameters(parametersRef)

	// 构造参数对 api 进行发包
	if len(paths) > 0 {
		apiMap := getMap(paths)

		for path, v := range apiMap {
			var (
				method string
				des    string
			)
			// 构造参数
			for method, des = range getMap([]byte(v)) {
				var (
					queryParams string
					bodyParams  string
					contentType string
				)
				header := make(map[string]string)

				consumes, _, _, _ := jsonparser.Get([]byte(des), "consumes")

				contentType = strings.ReplaceAll(strings.ReplaceAll(string(consumes), "[\"", ""), "\"]", "")
				parameters, _, _, _ := jsonparser.Get([]byte(des), "parameters")

				if len(parameters) > 0 {
					_, err := jsonparser.ArrayEach(parameters, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
						var (
							paraFormat string
							isRequired string
							paramsStr  string
						)

						arg, _, _, _ := jsonparser.Get(value, "name") // 参数名
						inTmp, _, _, _ := jsonparser.Get(value, "in") // 参数出现位置
						required, _, _, _ := jsonparser.Get(value, "required")
						defaultByte, _, _, _ := jsonparser.Get(value, "default")

						parameterRef, _, _, _ := jsonparser.Get(value, "$ref")
						if len(parameterRef) > 0 {
							parameter := strings.ReplaceAll(string(parameterRef), "#/parameters/", "")

							if parametersMap[parameter].In == "header" {
								header[parametersMap[parameter].Name] = parametersMap[parameter].Value
							} else if parametersMap[parameter].In == "path" {
								//  /replication/executions/{id} >  /replication/executions/1
								path = strings.ReplaceAll(path, "{"+parametersMap[parameter].Name+"}", parametersMap[parameter].Value)
							} else if parametersMap[parameter].In == "query" {
								queryParams += fmt.Sprintf("&%s", parametersMap[parameter].Value)
							} else {
								bodyParams += fmt.Sprintf("&%s", parametersMap[parameter].Value)
							}
						}

						if len(required) > 0 {
							isRequired = "*"
						}

						if len(defaultByte) > 0 {
							paramsStr = fmt.Sprintf("&%s=%s", string(arg), string(defaultByte))
						} else {
							if strings.Contains(string(value), "\"format\"") {
								paraFormatTmp, _, _, _ := jsonparser.Get(value, "format")
								paraFormat = string(paraFormatTmp)
								if len(paraFormatTmp) > 0 {
									paramsStr = fmt.Sprintf("&%s=%s%s%s", string(arg), isRequired, paraFormat, isRequired)
								}

								if strings.Contains(string(value), "\"items\"") {
									paraFormatTmp, _, _, _ := jsonparser.Get(value, "items", "format")
									paraFormat = string(paraFormatTmp)
									paramsStr = fmt.Sprintf("&%s=%s%s%s", string(arg), isRequired, paraFormat, isRequired)
								}

							} else if strings.Contains(string(value), "\"schema\"") {
								paraFormatTmp, _, _, _ := jsonparser.Get(value, "schema")
								if strings.Contains(string(paraFormatTmp), "\"$ref\"") {
									paraFormatTmp, _, _, _ = jsonparser.Get(paraFormatTmp, "$ref")
									paraFormat = strings.ReplaceAll(string(paraFormatTmp), "#/definitions/", "")

									paraFormat = fmt.Sprintf("%s", definitionsMap[paraFormat])

									if len(paraFormatTmp) > 0 {
										paramsStr = fmt.Sprintf("%s", paraFormat)
									}
								} else if strings.Contains(string(paraFormatTmp), "\"format\"") {
									paraFormatTmp, _, _, _ := jsonparser.Get(value, "format")
									paraFormat = string(paraFormatTmp)
									if len(paraFormatTmp) > 0 {
										paramsStr = fmt.Sprintf("&%s=%s%s%s", string(arg), isRequired, paraFormat, isRequired)
									}
								} else if strings.Contains(string(paraFormatTmp), "\"type\"") {
									paraFormatTmp, _, _, _ := jsonparser.Get(value, "type")
									paraFormat = string(paraFormatTmp)
									if len(paraFormatTmp) > 0 {
										paramsStr = fmt.Sprintf("&%s=%s%s%s", string(arg), isRequired, paraFormat, isRequired)
									}
								}

							} else {
								paraFormatTmp, _, _, _ := jsonparser.Get(value, "type")
								paraFormat = string(paraFormatTmp)

								if len(paraFormatTmp) > 0 {
									paramsStr = fmt.Sprintf("&%s=%s%s%s", string(arg), isRequired, paraFormat, isRequired)
								}
							}
						}

						if strings.EqualFold(string(inTmp), "header") {
							if len(defaultByte) > 0 {
								header[string(arg)] = string(defaultByte)
							} else {
								if paraFormat == "string" {
									header[string(arg)] = "test"
								} else {
									header[string(arg)] = "1"
								}
							}

						} else if strings.Contains(string(inTmp), "path") {
							if paraFormat == "string" {
								path = strings.ReplaceAll(path, "{"+string(arg)+"}", "test")
							} else {
								path = strings.ReplaceAll(path, "{"+string(arg)+"}", "1")
							}
						} else if strings.Contains(string(inTmp), "query") {
							queryParams += paramsStr
						} else {
							bodyParams += paramsStr
						}
					})

					if err != nil {
						logging.Logger.Errorln("ParseApi parameters error", err.Error())
						continue
					}

				}

				queryParams = strings.Replace(queryParams, "&", "", 1)
				bodyParams = strings.Replace(bodyParams, "&", "", 1)

				if contentType != "" {
					header["Content-Type"] = contentType
				}

				if strings.Contains(des, "#/definitions/") {
					header["Content-Type"] = "application/json"
				}

				scanApi(method, host, path, queryParams, bodyParams, header, basePath, schema)
			}
		}
	}

}

// scanApi 对拼接的 api 进行未授权、ssrf、注入等测试
func scanApi(method, baseUrl, path, queryParams, bodyParams string, header map[string]string, basePath, schema string) {
	//util.HttpProxy = "http://127.0.0.1:8080"
	//delete 方法有点危险，不进行测试
	if strings.EqualFold(method, "delete") {
		return
	}

	// 该 api 可能是一个删除、更新方法，安全着想不进行测试
	if funk.Contains(path, "delete") || funk.Contains(path, "del") || funk.Contains(path, "update") {
		return
	}

	queryParams = strings.ReplaceAll(queryParams, "*string*", "test")
	queryParams = strings.ReplaceAll(queryParams, "*int64*", "1")
	queryParams = strings.ReplaceAll(queryParams, "*int32*", "1")
	queryParams = strings.ReplaceAll(queryParams, "*number*", "1")
	queryParams = strings.ReplaceAll(queryParams, "*integer*", "1")
	queryParams = strings.ReplaceAll(queryParams, "*boolean*", "true")
	queryParams = strings.ReplaceAll(queryParams, "*int8*", "1")
	queryParams = strings.ReplaceAll(queryParams, "*object*", "{}")
	queryParams = strings.ReplaceAll(queryParams, "=string", "=test")

	bodyParams = strings.ReplaceAll(bodyParams, "*string*", "test")
	bodyParams = strings.ReplaceAll(bodyParams, "*int64*", "1")
	bodyParams = strings.ReplaceAll(bodyParams, "*int32*", "1")
	bodyParams = strings.ReplaceAll(bodyParams, "*integer*", "1")
	bodyParams = strings.ReplaceAll(bodyParams, "*int8*", "1")
	bodyParams = strings.ReplaceAll(bodyParams, "*number*", "1")
	bodyParams = strings.ReplaceAll(bodyParams, "*boolean*", "true")
	bodyParams = strings.ReplaceAll(bodyParams, "=string", "=test")

	host := schema + baseUrl

	if basePath != "" {
		host = host + basePath
	}

	// https://127.0.0.1/api/v2.0/   /test/ -> https://127.0.0.1/api/v2.0/test/
	if strings.HasSuffix(host, "/") && strings.HasPrefix(path, "/") {
		index := strings.LastIndex(host, "/")
		host = host[:index]
	}

	var target string
	if queryParams != "" {
		target = host + path + "?" + queryParams
	} else {
		target = host + path
	}

	if strings.Contains(bodyParams, "*file*") { //文件上传
		fileBody := strings.Split(bodyParams, "&")
		var (
			name     string
			fileName string
		)
		args := make(map[string]string)
		for _, i := range fileBody {
			if strings.Contains(i, "*file*") {
				tmp := strings.Split(i, "=")
				name = tmp[0]
				fileName = "test.txt"
			} else {
				tmp := strings.Split(i, "=")
				args[tmp[0]] = tmp[1]
			}
		}
		// 测试文件上传
		body := &bytes.Buffer{}                                           // 初始化body参数
		writer := multipart.NewWriter(body)                               // 实例化multipart
		part, err := writer.CreateFormFile(name, filepath.Base(fileName)) // 创建multipart 文件字段
		if err != nil {
			logging.Logger.Errorln("CreateFormFile error", err.Error())
			return
		}

		_, err = part.Write([]byte("test")) // 写入文件数据到multipart
		if err != nil {
			logging.Logger.Errorln("Write file data to multipart error", err.Error())
			return
		}

		for key, val := range args {
			_ = writer.WriteField(key, val) // 写入body中额外参数
		}
		err = writer.Close()
		if err != nil {
			logging.Logger.Errorln("writer.Close error", err.Error())
			return
		}

		resp, err := httpx.UploadRequest(target, args, name, fileName)
		if err != nil {
			logging.Logger.Errorln("UploadRequest err", err)
			return
		}

		if resp.StatusCode == 200 {
			output.OutChannel <- output.VulMessage{
				DataType: "web_vul",
				Plugin:   "Swagger 文件上传",
				VulnData: output.VulnData{
					CreateTime: time.Now().Format("2006-01-02 15:04:05"),
					Target:     target,
					Method:     method,
					Param:      bodyParams,
					Payload:    "",
					Request:    resp.RequestDump,
					Response:   resp.Body,
				},
				Level: output.Medium,
			}
		}

	} else { //ssrf、 文件读取测试
		sensitiveWords := []string{"url=", "path=", "uri=", "api=", "target=", "host=", "domain=", "ip=", "file="}

		var queryParams2 string
		var bodyParams2 string

		// ssrf、 文件读取测试
		for _, i := range sensitiveWords {
			if strings.Contains(queryParams, i) {
				tmp := queryParams
				queryParams = strings.ReplaceAll(tmp, i+"test", i+"https://www.baidu.com/")
				queryParams2 = strings.ReplaceAll(tmp, i+"test", i+"/etc/passwd")
			}
			if strings.Contains(bodyParams, i) {
				tmp := bodyParams
				bodyParams = strings.ReplaceAll(tmp, i+"test", i+"https://www.baidu.com/")
				bodyParams2 = strings.ReplaceAll(tmp, i+"test", i+"/etc/passwd")
			}
		}

		scan(method, target, bodyParams, header)

		if queryParams2 != "" {
			target = host + path + "?" + queryParams2
			scan(method, target, bodyParams, header)
		}

		if bodyParams2 != "" {
			scan(method, target, bodyParams2, header)
		}
	}

}

func scan(method, target, bodyParams string, header map[string]string) {
	//util.HttpProxy = "http://127.0.0.1:8080"
	if strings.EqualFold(method, "get") {
		req, err := httpx.Request(target, "GET", "", false, header)
		if err != nil {
			logging.Logger.Errorf("scanApi(GET %s) err %v", target, err)
			return
		}

		SwaggerChan <- tools.Result{
			Url:           target,
			Method:        "GET",
			StatusCode:    req.StatusCode,
			ContentLength: req.ContentLength,
			Request:       req.RequestDump,
			Response:      req.ResponseDump,
		}

		// 可能是未授权, 然后对 200 的进行ssrf、注入测试
		if req.StatusCode == 200 && !util.In(req.Body, conf.Page403Content) && !util.In(req.Body, conf.Page403Content) {
			logging.Logger.Infof("Possibly unauthorized access: GET %s", target)

			in := &input.CrawlResult{
				Url:     target,
				Headers: header,
				Kv:      bodyParams,
				Method:  method,
			}

			sqlmap.Scan(in)

			output.OutChannel <- output.VulMessage{
				DataType: "web_vul",
				Plugin:   "Swagger unauthorized",
				VulnData: output.VulnData{
					CreateTime: time.Now().Format("2006-01-02 15:04:05"),
					Target:     target,
					Method:     method,
					Param:      bodyParams,
					Payload:    "",
					Request:    req.RequestDump,
					Response:   req.Body,
				},
				Level: output.Low,
			}

			if funk.Contains(req.Body, "www.baidu.com/img/sug_bd.png") {
				logging.Logger.Infof("存在 SSRF漏洞: GET %s", target)

				output.OutChannel <- output.VulMessage{
					DataType: "web_vul",
					Plugin:   "Swagger SSRF",
					VulnData: output.VulnData{
						CreateTime: time.Now().Format("2006-01-02 15:04:05"),
						Target:     target,
						Method:     method,
						Param:      bodyParams,
						Payload:    "",
						Request:    req.RequestDump,
						Response:   req.Body,
					},
					Level: output.Critical,
				}
			}

			if funk.Contains(req.Body, "root:x:0:0:root:/root:") {
				logging.Logger.Infof("存在任意文件读取漏洞: GET %s", target)

				output.OutChannel <- output.VulMessage{
					DataType: "web_vul",
					Plugin:   "Swagger File Reading",
					VulnData: output.VulnData{
						CreateTime: time.Now().Format("2006-01-02 15:04:05"),
						Target:     target,
						Method:     method,
						Param:      bodyParams,
						Payload:    "",
						Request:    req.RequestDump,
						Response:   req.Body,
					},
					Level: output.Critical,
				}
			}
		}

	} else {
		req, err := httpx.Request(target, method, bodyParams, false, header)
		if err != nil {
			logging.Logger.Errorf("scanApi(%s %s) err %v", method, target, err)
			return
		}

		SwaggerChan <- tools.Result{
			Url:           target,
			Method:        method,
			StatusCode:    req.StatusCode,
			ContentLength: req.ContentLength,
			Request:       req.RequestDump,
			Response:      req.ResponseDump,
		}

		if req.StatusCode == 200 && !util.In(req.Body, conf.Page403Content) && !util.In(req.Body, conf.Page403Content) { // 可能是未授权
			payload := fmt.Sprintf("%s %s %s", method, target, bodyParams)
			logging.Logger.Infof("Possibly unauthorized access: %s", payload)

			in := &input.CrawlResult{
				Url:     target,
				Headers: header,
				Kv:      bodyParams,
				Method:  method,
			}

			sqlmap.Scan(in)

			output.OutChannel <- output.VulMessage{
				DataType: "web_vul",
				Plugin:   "Swagger unauthorized",
				VulnData: output.VulnData{
					CreateTime: time.Now().Format("2006-01-02 15:04:05"),
					Target:     target,
					Method:     method,
					Param:      bodyParams,
					Payload:    "",
					Request:    req.RequestDump,
					Response:   req.Body,
				},
				Level: output.Low,
			}

			if funk.Contains(req.Body, "www.baidu.com/img/sug_bd.png") {

				logging.Logger.Infof("存在SSRF漏洞漏洞: %s ", payload)
				output.OutChannel <- output.VulMessage{
					DataType: "web_vul",
					Plugin:   "Swagger SSRF",
					VulnData: output.VulnData{
						CreateTime: time.Now().Format("2006-01-02 15:04:05"),
						Target:     target,
						Method:     method,
						Param:      bodyParams,
						Payload:    "",
						Request:    req.RequestDump,
						Response:   req.Body,
					},
					Level: output.Critical,
				}
			}

			if funk.Contains(req.Body, "root:x:0:0:root:/root:") {
				logging.Logger.Infof("存在任意文件读取漏洞: %s ", payload)

				output.OutChannel <- output.VulMessage{
					DataType: "web_vul",
					Plugin:   "Swagger File Reading",
					VulnData: output.VulnData{
						CreateTime: time.Now().Format("2006-01-02 15:04:05"),
						Target:     target,
						Method:     method,
						Param:      bodyParams,
						Payload:    "",
						Request:    req.RequestDump,
						Response:   req.Body,
					},
					Level: output.Critical,
				}
			}
		}
	}
}

func getMap(body []byte) map[string]string {
	apiMap := make(map[string]string)

	// You can use `ObjectEach` helper to iterate objects { "key1":object1, "key2":object2, .... "keyN":objectN }
	jsonparser.ObjectEach(body, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		//fmt.Printf("Key: %s\n Value: %s\n Type: %s\n", string(key), string(value), dataType)
		apiMap[string(key)] = string(value)
		return nil
	})

	return apiMap
}

func getDefinitions(definitions []byte) map[string]string {
	definitionsMap := getMap(definitions)

	if definitionsMap != nil {
		for path, v := range definitionsMap {
			properties, _, _, _ := jsonparser.Get([]byte(v), "properties")

			var paramsStr string

			for k, value := range getMap(properties) {
				if strings.Contains(value, "\"example\"") {
					paraExample, _, _, _ := jsonparser.Get([]byte(value), "example")
					if len(paraExample) > 0 {
						paramsStr += fmt.Sprintf("\"%s\":\"%s\",", k, string(paraExample))
					}
				} else if strings.Contains(value, "\"format\"") {
					paraFormat, _, _, _ := jsonparser.Get([]byte(value), "format")
					if len(paraFormat) > 0 {
						paramsStr += fmt.Sprintf("\"%s\":*%s*,", k, string(paraFormat))
					}
				} else if strings.Contains(value, "\"$ref\"") {
					if strings.Contains(value, "\"items\"") {
						var paraFormatTmp []byte
						flag := false
						if strings.Contains(value, "\"additionalProperties\"") {
							flag = true
							paraFormatTmp, _, _, _ = jsonparser.Get([]byte(value), "additionalProperties", "items")
						} else {
							paraFormatTmp, _, _, _ = jsonparser.Get([]byte(value), "items")
						}

						if strings.Contains(string(paraFormatTmp), "\"items\"") {
							var (
								paraFormatTmp1 []byte
								paraFormatType []byte
							)
							if flag {
								paraFormatTmp1, _, _, _ = jsonparser.Get([]byte(value), "additionalProperties", "items", "items", "$ref")
								paraFormatType, _, _, _ = jsonparser.Get([]byte(value), "additionalProperties", "items", "type")
								paraFormat := strings.ReplaceAll(string(paraFormatTmp1), "#/definitions/", "")
								if paraFormat != path { // todo 有的结构本身又作为参数，这是什么操作
									if string(paraFormatType) == "array" {
										paramsStr += fmt.Sprintf("\"%s\":{\"additionalProp\":[**%s**]},", k, paraFormat)
									} else {
										paramsStr += fmt.Sprintf("\"%s\":{\"additionalProp\":**%s**},", k, paraFormat)
									}
								} else {
									paramsStr += fmt.Sprintf("\"%s\":\"*%s*\",", k, paraFormat)
								}

							} else {
								paraFormatTmp1, _, _, _ = jsonparser.Get([]byte(value), "items", "items", "$ref")
								paraFormatType, _, _, _ = jsonparser.Get([]byte(value), "items", "type")
								paraFormat := strings.ReplaceAll(string(paraFormatTmp1), "#/definitions/", "")
								if paraFormat != path { // todo 有的结构本身又作为参数，这是什么操作
									if string(paraFormatType) == "array" {
										paramsStr += fmt.Sprintf("\"%s\":[**%s**],", k, paraFormat)
									} else {
										paramsStr += fmt.Sprintf("\"%s\":**%s**,", k, paraFormat)
									}
								} else {
									paramsStr += fmt.Sprintf("\"%s\":\"*%s*\",", k, paraFormat)
								}

							}

						} else {
							var (
								paraFormatTmp1 []byte
								paraFormatType []byte
							)

							if flag {
								paraFormatTmp1, _, _, _ = jsonparser.Get([]byte(value), "additionalProperties", "items", "$ref")
								paraFormatType, _, _, _ = jsonparser.Get([]byte(value), "additionalProperties", "type")
								paraFormat := strings.ReplaceAll(string(paraFormatTmp1), "#/definitions/", "")
								if paraFormat != path { // todo 有的结构本身又作为参数，这是什么操作
									if string(paraFormatType) == "array" {
										paramsStr += fmt.Sprintf("\"%s\":{\"additionalProp\":[**%s**]},", k, paraFormat)
									} else {
										paramsStr += fmt.Sprintf("\"%s\":{\"additionalProp\":**%s**},", k, paraFormat)
									}
								} else {
									paramsStr += fmt.Sprintf("\"%s\":\"*%s*\",", k, paraFormat)
								}
							} else {
								paraFormatTmp1, _, _, _ = jsonparser.Get([]byte(value), "items", "$ref")
								paraFormatType, _, _, _ = jsonparser.Get([]byte(value), "type")
								paraFormat := strings.ReplaceAll(string(paraFormatTmp1), "#/definitions/", "")
								if paraFormat != path { // todo 有的结构本身又作为参数，这是什么操作
									if string(paraFormatType) == "array" {
										paramsStr += fmt.Sprintf("\"%s\":[**%s**],", k, paraFormat)
									} else {
										paramsStr += fmt.Sprintf("\"%s\":**%s**,", k, paraFormat)
									}
								} else {
									paramsStr += fmt.Sprintf("\"%s\":\"*%s*\",", k, paraFormat)
								}
							}

						}

					} else {
						paraFormatTmp, _, _, _ := jsonparser.Get([]byte(value), "$ref")
						paraFormat := strings.ReplaceAll(string(paraFormatTmp), "#/definitions/", "")
						if paraFormat != path { // todo 有的结构本身又作为参数，这是什么操作
							paramsStr += fmt.Sprintf("\"%s\":**%s**,", k, paraFormat)
						} else {
							paramsStr += fmt.Sprintf("\"%s\":\"*%s*\",", k, paraFormat)
						}
					}

				} else if strings.Contains(value, "\"items\"") {

					paraExample, _, _, _ := jsonparser.Get([]byte(value), "items", "example")
					if len(paraExample) > 0 {
						paramsStr += fmt.Sprintf("\"%s\":[%s],", k, string(paraExample))
					} else {
						paraFormat, _, _, _ := jsonparser.Get([]byte(value), "items", "type")
						if len(paraFormat) > 0 {
							paramsStr += fmt.Sprintf("\"%s\":[*%s*],", k, string(paraFormat))
						}
					}

				} else {
					paraFormat, _, _, _ := jsonparser.Get([]byte(value), "type")
					if len(paraFormat) > 0 {
						paramsStr += fmt.Sprintf("\"%s\":*%s*,", k, string(paraFormat))
					}
				}
			}

			paramsStr = strings.ReplaceAll(paramsStr, "*int32*", "1")
			paramsStr = strings.ReplaceAll(paramsStr, "*integer*", "1")
			paramsStr = strings.ReplaceAll(paramsStr, "*number*", "1")
			paramsStr = strings.ReplaceAll(paramsStr, "*object*", "{}")
			paramsStr = strings.ReplaceAll(paramsStr, "*double*", "1")
			paramsStr = strings.ReplaceAll(paramsStr, "*int64*", "1")
			paramsStr = strings.ReplaceAll(paramsStr, "*string*", "\"test\"")
			paramsStr = strings.ReplaceAll(paramsStr, "*boolean*", "true")
			paramsStr = strings.ReplaceAll(paramsStr, "*date-time*", "\"2022-09-15T14:13:19.272Z\"")
			paramsStr = strings.TrimRight(paramsStr, ",")
			paramsStr = "{" + paramsStr + "}"

			definitionsMap[path] = paramsStr

		}

		var limit = 0
		for {
			limit += 1
			var num = 0
			definitionsMapTmp := definitionsMap
			for k, v := range definitionsMapTmp {
				if strings.Contains(v, "**") {
					num += 1
					regex := regexp.MustCompile(`\*\*[A-Za-z0-9-_=.]+\*\*`)
					value := regex.FindAllString(v, -1)

					for _, kk := range value {
						kkk := strings.ReplaceAll(kk, "**", "")
						reValue := definitionsMap[kkk]
						if strings.Count(reValue, "**") == 0 {
							definitionsMap[k] = strings.ReplaceAll(v, kk, reValue)
						}
					}
				}

			}

			if num == 0 || limit > 50 { // 防止死循环
				break
			}
		}

		return definitionsMap
	}

	return nil
}

func getParameters(parameters []byte) map[string]Parameter {
	parametersMap := getMap(parameters)
	argsMap := make(map[string]Parameter)
	if parametersMap != nil {
		for k, v := range parametersMap {
			parameterName, _, _, _ := jsonparser.Get([]byte(v), "name")
			parameterType, _, _, _ := jsonparser.Get([]byte(v), "type")
			parameterIn, _, _, _ := jsonparser.Get([]byte(v), "in")

			var value string
			if string(parameterType) == "boolean" {
				if string(parameterIn) == "path" || string(parameterIn) == "header" {
					value = "true"
				} else {
					value = string(parameterName) + "=*boolean*"
				}

			} else if string(parameterType) == "int" || string(parameterType) == "int8" || string(parameterType) == "integer" || string(parameterType) == "int32" || string(parameterType) == "int64" || string(parameterType) == "number" || string(parameterType) == "double" {
				if string(parameterIn) == "path" || string(parameterIn) == "header" {
					value = "1"
				} else {
					value = string(parameterName) + "=*int32*"
				}

			} else {
				if string(parameterIn) == "path" || string(parameterIn) == "header" {
					value = "test"
				} else {
					value = string(parameterName) + "=*string*"
				}
			}

			argsMap[k] = Parameter{
				Name:  string(parameterName),
				Value: value,
				In:    string(parameterIn),
			}

		}
	}
	return argsMap
}
