package sqlmap

import (
	"encoding/json"
	"fmt"
	"github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	JieOutput "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"
)

/**
  @author: yhy
  @since: 2023/10/26
  @desc: 转发给 sql, 由专业的工具 sql 进行扫描注入
**/

type Plugin struct {
	SeenRequests sync.Map
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
	if conf.GlobalConfig.SqlmapApi.Url == "" {
		logging.Logger.Warnln("sqlmap api 未配置")
		return
	}

	if p.IsScanned(in.UniqueId) {
		return
	}
	// 转发之前先判断是否有参数，存在参数才转发
	params := util.ExtractParameters(in.Url, in.Method, in.RequestBody, in.Headers)

	if len(params) == 0 {
		return
	}

	// 创建新任务
	taskID := createTask()
	if taskID == "" {
		return
	}
	// 开始扫描
	startScan(taskID, in)
	logging.Logger.Debugln("Sqlmap Scan started for target:", in.Url, taskID)

	// 监控任务状态
	go getTaskStatus(taskID, in.Url)
}

func (p *Plugin) IsScanned(key string) bool {
	if key == "" {
		return false
	}
	if _, ok := p.SeenRequests.Load(key); ok {
		return true
	}
	p.SeenRequests.Store(key, true)
	return false
}

func (p *Plugin) Name() string {
	return "sqlmapApi"
}

var client = &http.Client{}

func createTask() string {
	body := request("GET", "/task/new", "")
	result := &struct {
		Taskid  string `json:"taskid"`
		Success bool   `json:"success"`
	}{}

	err := json.Unmarshal(body, result)

	if err != nil {
		logging.Logger.Errorln(err, body)
		return ""
	}

	return result.Taskid
}

func startScan(taskID string, in *input.CrawlResult) bool {
	data := option{
		Url:         in.Url,
		Method:      in.Method,
		Data:        in.RequestBody,
		RandomAgent: true,
		Level:       1,
		Risk:        1,
		Verbose:     5,
		Proxy:       conf.GlobalConfig.Http.Proxy,
	}

	for k, v := range in.Headers {
		data.Headers += k + ": " + v + "\r\n"
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		logging.Logger.Println("Error converting struct to JSON:", err)
		return false
	}

	body := request("POST", fmt.Sprintf("/scan/%s/start", taskID), string(jsonData))

	result := &struct {
		Engineid int  `json:"engineid"`
		Success  bool `json:"success"`
	}{}

	err = json.Unmarshal(body, result)
	if err != nil {
		logging.Logger.Errorln(err)
		return false
	}

	return result.Success
}

func getTaskStatus(taskID string, target string) {
	statusURL := fmt.Sprintf("/scan/%s/status", taskID)
	resultURL := fmt.Sprintf("/scan/%s/data", taskID)
	for {
		statusBody := request("GET", statusURL, "")
		statusResult := &struct {
			Status     string `json:"status"`
			Returncode int    `json:"returncode"`
			Success    bool   `json:"success"`
		}{}
		err := json.Unmarshal(statusBody, statusResult)
		if err != nil {
			logging.Logger.Errorln(err)
			return
		}
		// 任务还没有完成
		if statusResult.Status != "terminated" && statusResult.Status != "not running" {
			time.Sleep(5 * time.Second)
			continue
		}

		resultBody := request("GET", resultURL, "")
		result := &struct {
			Success bool                     `json:"success"`
			Error   []string                 `json:"error"`
			Data    []map[string]interface{} `json:"data"`
		}{}

		err = json.Unmarshal(resultBody, result)
		if err != nil {
			logging.Logger.Errorln(err)
			return
		}

		if len(result.Data) > 0 {
			for index := range result.Data {
				tmpType := make([]interface{}, 0)
				if reflect.TypeOf(result.Data[index]["value"]) == reflect.TypeOf(tmpType) {
					sqlmapValues := result.Data[index]["value"].([]interface{})
					for vIndex := range sqlmapValues {
						param := sqlmapValues[vIndex].(map[string]interface{})["parameter"].(string)
						injectTypes := sqlmapValues[vIndex].(map[string]interface{})["data"].(map[string]interface{})
						for _, iValue := range injectTypes {
							injectTitle := iValue.(map[string]interface{})["title"].(string)
							injectPayload := iValue.(map[string]interface{})["payload"].(string)
							JieOutput.OutChannel <- JieOutput.VulMessage{
								DataType: "web_vul",
								Plugin:   "SQL Injection",
								VulnData: JieOutput.VulnData{
									CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
									Target:      target,
									Param:       param,
									Payload:     injectPayload,
									Description: injectTitle,
								},
								Level: JieOutput.Critical,
							}
							logging.Logger.Infof("Sqlmap 检测到%s %s 参数存在 sql 注入[%s:%s]", target, param, injectTitle, injectPayload)
							return
						}
					}
				}
			}
			return
		}
		return
	}
}

func request(method, endpoint string, body string) []byte {
	Transport := &http.Transport{
		MaxIdleConnsPerHost: -1,
		DisableKeepAlives:   true,
	}

	if conf.GlobalConfig.Http.Proxy != "" {
		proxyURL, _ := url.Parse(conf.GlobalConfig.Http.Proxy)
		Transport.Proxy = http.ProxyURL(proxyURL)
	}

	if strings.HasSuffix(conf.GlobalConfig.SqlmapApi.Url, "/") {
		conf.GlobalConfig.SqlmapApi.Url = strings.TrimRight(conf.GlobalConfig.SqlmapApi.Url, "/")
	}

	if !strings.HasPrefix(conf.GlobalConfig.SqlmapApi.Url, "https://") && !strings.HasPrefix(conf.GlobalConfig.SqlmapApi.Url, "http://") {
		conf.GlobalConfig.SqlmapApi.Url = "http://" + conf.GlobalConfig.SqlmapApi.Url
	}

	req, err := http.NewRequest(method, conf.GlobalConfig.SqlmapApi.Url+endpoint, strings.NewReader(body))
	if err != nil {
		logging.Logger.Println("Error creating request:", err)
		return nil
	}
	client.Transport = Transport
	req.Header.Set("Content-Type", "application/json")
	if conf.GlobalConfig.SqlmapApi.Username != "" && conf.GlobalConfig.SqlmapApi.Password != "" {
		req.SetBasicAuth(conf.GlobalConfig.SqlmapApi.Username, conf.GlobalConfig.SqlmapApi.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		logging.Logger.Println("Error executing request:", err)
		return nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.Logger.Println("Error reading response body:", err)
		return nil
	}

	return respBody
}
