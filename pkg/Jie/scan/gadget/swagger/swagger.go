package swagger

import (
	"fmt"
	"github.com/buger/jsonparser"
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/sql"
	"github.com/yhy0/ChYing/pkg/ui"
	"github.com/yhy0/logging"
	"net/url"
	"strings"
	"time"
)

/**
  @author: yhy
  @since: 2022/9/14
  @desc: 对 swagger api 进行未授权、ssrf、注入等测试
    参考了 https://github.com/lijiejie/swagger-exp ，不过更加强大，会自动对参数进行识别赋值
**/

var Keys = []string{"location", "url", "path"}

type Parameter struct {
	Name  string
	In    string
	Value string
}

func Scan(target string, client *httpx.Client, fuzzUiResp chan ui.Result) {
	u, err := url.Parse(target)
	if err != nil {
		logging.Logger.Errorf("Scan url.Parse(%s) err: %v", target, err)
	}

	host := u.Scheme + "://" + u.Host
	path := u.Path

	var schema = u.Scheme + "://"

	var basePath = ""

	index := strings.LastIndex(path, "/")
	if index != 0 {
		basePath = path[:index]
	}
	resp, err := client.Request(target, "GET", "", nil, "Swagger")
	if err != nil {
		logging.Logger.Errorln("Scan err: ", err)
		if fuzzUiResp != nil {
			fuzzUiResp <- ui.Result{
				Url:    target,
				Method: "GET",
			}
		}
		return
	}

	if fuzzUiResp != nil {
		fuzzUiResp <- ui.Result{
			Url:           target,
			Method:        "GET",
			StatusCode:    resp.StatusCode,
			ContentLength: resp.ContentLength,
			Request:       resp.RequestDump,
			Response:      resp.ResponseDump,
		}
	}
	var (
		swaggerResources bool
		swaggerJson      bool
	)

	if util.Contains(resp.Body, "Swagger UI") { // 传入的是 /swagger-ui.html
		// 访问 /swagger-resources 获取 对应的 api 集合
		t := strings.ReplaceAll(target, "swagger-ui.html", "swagger-resources")
		resp, err = client.Request(t, "GET", "", nil, "Swagger")
		if err != nil {
			if fuzzUiResp != nil {
				fuzzUiResp <- ui.Result{
					Url:    t,
					Method: "GET",
				}
			}
			logging.Logger.Errorln("Scan err: ", err)
			return
		}
		if fuzzUiResp != nil {
			fuzzUiResp <- ui.Result{
				Url:           t,
				Method:        "GET",
				StatusCode:    resp.StatusCode,
				ContentLength: resp.ContentLength,
				Request:       resp.RequestDump,
				Response:      resp.ResponseDump,
			}
		}
		swaggerResources = true
	}

	if util.Contains(resp.Body, "swaggerVersion") {
		swaggerResources = true
	} else if util.Contains(resp.Body, "swagger") {
		swaggerJson = true
	}

	if !swaggerResources && !swaggerJson {
		logging.Logger.Warnf("Swagger parsing failed. Not find swaggerVersion/swagger")
		return
	}

	if swaggerResources { // 传入的是 /swagger_resources ,首先解析出所有版本的 Swagger api 路径 eg: /v2/api-docs, /v1/api-docs, /test/api
		var apis []string
		for _, key := range Keys {
			_, err = jsonparser.ArrayEach([]byte(resp.Body), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
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
			resp, err = client.Request(host+basePath+api, "GET", "", nil, "Swagger")
			if err != nil {
				logging.Logger.Errorf("Scan(%s) err: %v", api, err)
				if fuzzUiResp != nil {
					fuzzUiResp <- ui.Result{
						Url:    host + basePath + api,
						Method: "GET",
					}
				}
				continue
			}
			if fuzzUiResp != nil {
				fuzzUiResp <- ui.Result{
					Url:           host + basePath + api,
					Method:        "GET",
					StatusCode:    resp.StatusCode,
					ContentLength: resp.ContentLength,
					Request:       resp.RequestDump,
					Response:      resp.ResponseDump,
				}
			}
			if util.Contains(resp.Body, "\"swagger\":") { // 解析每个文档
				logging.Logger.Infof("Start test %s...", api)
				parseApi([]byte(resp.Body), u.Host, basePath, schema, client, fuzzUiResp)
			}
		}
	}

	if swaggerJson { // 表示传入的就是swagger api路径，直接解析 eg: /v2/api-docs
		resp, err = client.Request(target, "GET", "", nil, "Swagger")
		if err != nil {
			logging.Logger.Errorf("Scan(%s) err: %v", target, err)
			if fuzzUiResp != nil {
				fuzzUiResp <- ui.Result{
					Url:    target,
					Method: "GET",
				}
			}
			return
		}
		if fuzzUiResp != nil {
			fuzzUiResp <- ui.Result{
				Url:           target,
				Method:        "GET",
				StatusCode:    resp.StatusCode,
				ContentLength: resp.ContentLength,
				Request:       resp.RequestDump,
				Response:      resp.ResponseDump,
			}
		}
		if util.Contains(resp.Body, "\"swagger\":") { // 解析每个文档
			logging.Logger.Infof("Start test %s, please wait...", target)
			parseApi([]byte(resp.Body), u.Host, basePath, schema, client, fuzzUiResp)
		}
	}
}

// ParseApi 解析 swagger api 文档
func parseApi(body []byte, host, basePath, schema string, client *httpx.Client, fuzzUiResp chan ui.Result) {

	hostTmp, _, _, _ := jsonparser.Get(body, "host")
	if len(hostTmp) > 0 {
		host = string(hostTmp)
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
	parametersRef, _, _, _ := jsonparser.Get(body, "parameters") // 某些
	hostByte, _, _, _ := jsonparser.Get(body, "host")

	if len(hostByte) > 0 {
		h := strings.Split(string(hostByte), ":")
		if !util.IsInnerIP(h[0]) {
			host = string(hostByte)
		}
	}
	basePathTmp, _, _, _ = jsonparser.Get(body, "basePath")

	if len(basePathTmp) > 0 {
		basePath = string(basePathTmp)
	}

	paths, _, _, _ = jsonparser.Get(body, "paths") // 要拼接的路径信息

	// $ref 参数中的这种参数的定义
	definitionsRef, _, _, _ = jsonparser.Get(body, "definitions") // 某些参数的定义
	definitionsMap = getDefinitions(definitionsRef)

	// $ref 参数中的这种参数的定义 , 这种可能会作为路径或参数的一部分
	parametersRef, _, _, _ = jsonparser.Get(body, "parameters") // 某些参数的定义
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

				scanApi(method, host, path, queryParams, bodyParams, header, basePath, schema, client, fuzzUiResp)
			}
		}
	}

}

// scanApi 对拼接的 api 进行未授权、ssrf、注入等测试
func scanApi(method, baseUrl, path, queryParams, bodyParams string, header map[string]string, basePath, schema string, client *httpx.Client, fuzzUiResp chan ui.Result) {
	// util.HttpProxy = "http://127.0.0.1:8080"
	// delete 方法有点危险，不进行测试
	if strings.EqualFold(method, "delete") {
		return
	}

	// 该 api 可能是一个删除、更新方法，安全着想不进行测试
	if util.Contains(path, "delete") || util.Contains(path, "del") || util.Contains(path, "update") {
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

	if strings.Contains(bodyParams, "*file*") { // 文件上传
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
		resp, err := client.Upload(target, args, name, fileName)
		if err != nil {
			logging.Logger.Errorln("UploadRequest err", err)
			if fuzzUiResp != nil {
				fuzzUiResp <- ui.Result{
					Url:    target,
					Method: "Post",
				}
			}
			return
		}

		if fuzzUiResp != nil {
			fuzzUiResp <- ui.Result{
				Url:           target,
				Method:        "Post",
				StatusCode:    resp.StatusCode,
				ContentLength: resp.ContentLength,
				Request:       resp.RequestDump,
				Response:      resp.ResponseDump,
			}
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

	} else { // ssrf、 文件读取测试
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

		scanSwagger(method, target, bodyParams, header, client, fuzzUiResp)

		if queryParams2 != "" {
			target = host + path + "?" + queryParams2
			scanSwagger(method, target, bodyParams, header, client, fuzzUiResp)
		}

		if bodyParams2 != "" {
			scanSwagger(method, target, bodyParams2, header, client, fuzzUiResp)
		}
	}

}

func scanSwagger(method, target, bodyParams string, header map[string]string, client *httpx.Client, fuzzUiResp chan ui.Result) {
	parse, err := url.Parse(target)
	if err != nil {
		return
	}
	// util.HttpProxy = "http://127.0.0.1:8080"
	if strings.EqualFold(method, "get") {
		res, err := client.Request(target, "GET", "", header, "Swagger")
		if err != nil {
			logging.Logger.Errorf("scanApi(GET %s) err %v", target, err)
			if fuzzUiResp != nil {
				fuzzUiResp <- ui.Result{
					Url:    target,
					Method: "GET",
				}
			}
			return
		}
		if fuzzUiResp != nil {
			fuzzUiResp <- ui.Result{
				Url:           target,
				Method:        "GET",
				StatusCode:    res.StatusCode,
				ContentLength: res.ContentLength,
				Request:       res.RequestDump,
				Response:      res.ResponseDump,
			}
		}
		// 可能是未授权, 然后对 200 的进行ssrf、注入测试
		if res.StatusCode == 200 && !util.IsBlackHtml(res.Body, res.RespHeader["Content-Type"], parse.Path) {
			logging.Logger.Infof("Possibly unauthorized access: GET %s", target)

			in := &input.CrawlResult{
				Url:     target,
				Headers: header,
				Kv:      bodyParams,
				Method:  method,
			}
			sqlPlugin := &sql.Plugin{}
			sqlPlugin.Scan(target, "", in, client)

			output.OutChannel <- output.VulMessage{
				DataType: "web_vul",
				Plugin:   "Swagger unauthorized",
				VulnData: output.VulnData{
					CreateTime: time.Now().Format("2006-01-02 15:04:05"),
					Target:     target,
					Method:     method,
					Param:      bodyParams,
					Payload:    "",
					Request:    res.RequestDump,
					Response:   res.Body,
				},
				Level: output.Low,
			}

			if util.Contains(res.Body, "<title>百度一下，你就知道</title>") {
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
						Request:    res.RequestDump,
						Response:   res.Body,
					},
					Level: output.Critical,
				}
			}

			if util.Contains(res.Body, "root:x:0:0:root:/root:") {
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
						Request:    res.RequestDump,
						Response:   res.Body,
					},
					Level: output.Critical,
				}
			}
		}

	} else {
		res, err := client.Request(target, method, bodyParams, header, "Swagger")
		if err != nil {
			logging.Logger.Errorf("scanApi(%s %s) err %v", method, target, err)
			if fuzzUiResp != nil {
				fuzzUiResp <- ui.Result{
					Url:    target,
					Method: method,
				}
			}
			return
		}
		if fuzzUiResp != nil {
			fuzzUiResp <- ui.Result{
				Url:           target,
				Method:        method,
				StatusCode:    res.StatusCode,
				ContentLength: res.ContentLength,
				Request:       res.RequestDump,
				Response:      res.ResponseDump,
			}
		}
		if res.StatusCode == 200 && !util.IsBlackHtml(res.Body, res.RespHeader["Content-Type"], parse.Path) { // 可能是未授权
			payload := fmt.Sprintf("%s %s %s", method, target, bodyParams)
			logging.Logger.Infof("Possibly unauthorized access: %s", payload)

			in := &input.CrawlResult{
				Url:     target,
				Headers: header,
				Kv:      bodyParams,
				Method:  method,
			}
			sqlPlugin := &sql.Plugin{}
			sqlPlugin.Scan(target, "", in, client)

			output.OutChannel <- output.VulMessage{
				DataType: "web_vul",
				Plugin:   "Swagger unauthorized",
				VulnData: output.VulnData{
					CreateTime: time.Now().Format("2006-01-02 15:04:05"),
					Target:     target,
					Method:     method,
					Param:      bodyParams,
					Payload:    "",
					Request:    res.RequestDump,
					Response:   res.Body,
				},
				Level: output.Low,
			}

			if util.Contains(res.Body, "<title>百度一下，你就知道</title>") {

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
						Request:    res.RequestDump,
						Response:   res.Body,
					},
					Level: output.Critical,
				}
			}

			if util.Contains(res.Body, "root:x:0:0:root:/root:") {
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
						Request:    res.RequestDump,
						Response:   res.Body,
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
		// fmt.Printf("Key: %s\n Value: %s\n Type: %s\n", string(key), string(value), dataType)
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
