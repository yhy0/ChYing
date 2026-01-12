package jsonp

import (
	"errors"
	"fmt"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"io"
	"net/url"
	"sync"
	"time"
)

/**
  @author: yhy
  @since: 2023/1/5
  @desc: https://github.com/wrenchonline/glint/blob/main/pkg/pocs/jsonp/jsonp.go
**/

type Plugin struct {
	SeenRequests sync.Map
}

func (p *Plugin) Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) {
	if p.IsScanned(in.UniqueId) {
		return
	}

	isvul, info, err := CheckSenseJsonp(in, client)
	if err != nil {
		logging.Logger.Errorf("check jsonp error: %v", err)
		return
	}

	if isvul {
		output.OutChannel <- output.VulMessage{
			DataType: "web_vul",
			Plugin:   "JSONP",
			VulnData: output.VulnData{
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
				Target:     in.Url,
				Method:     in.Method,
				Ip:         in.Ip,
				Param:      "",
				Request:    info.Request,
				Response:   info.Response,
				Payload:    in.Url,
			},
			Level: output.Medium,
		}
		return
	}
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
	return "jsonp"
}

type JsonpInfo struct {
	Request  string
	Response string
}

func CheckSenseJsonp(in *input.CrawlResult, client *httpx.Client) (bool, *JsonpInfo, error) {
	queryMap, _, err := UrlParser(in.Url)
	if err != nil {
		return false, nil, err
	}

	isCallback, callbackFuncName, err := CheckJSIsCallback(queryMap)
	if err != nil {
		return false, nil, err
	}
	if isCallback {
		//    referer： host 请求
		normalRespContent, _, err := GetJsResponse(in, client)
		if err != nil {
			logging.Logger.Errorf("GetJsResponse error: %v", err)
			return false, nil, err
		}
		isJsonpNormal, err := CheckJsRespAst(normalRespContent, callbackFuncName)
		if err != nil {
			logging.Logger.Errorf("GetJsResponse error: %v", err)
			return false, nil, err
		}
		// 如果包含敏感字段 将 referer 置空 再请求一次
		if isJsonpNormal {
			in.Headers["Referer"] = ""
			noRefererContent, info, err := GetJsResponse(in, client)
			if err != nil {
				logging.Logger.Errorf("GetJsResponse error: %v", err)
				return false, nil, err
			}
			isJsonp, err := CheckJsRespAst(noRefererContent, callbackFuncName)
			if err != nil {
				logging.Logger.Errorf("GetJsResponse error: %v", err)
				return false, nil, err
			}
			return isJsonp, info, nil
		}

	}
	return false, nil, nil
}

func UrlParser(jsUrl string) (url.Values, string, error) {
	urlParser, err := url.Parse(jsUrl)
	if err != nil {
		return nil, "", err
	}
	// 拼接原始referer
	domainString := urlParser.Scheme + "://" + urlParser.Host
	return urlParser.Query(), domainString, nil
}

func CheckJSIsCallback(queryMap url.Values) (bool, string, error) {
	var re = regexp.MustCompile(`(?m)(?i)(callback)|(jsonp)|(^cb$)|(function)`)
	for k, v := range queryMap {
		regResult := re.FindAllString(k, -1)
		if len(regResult) > 0 && len(v) > 0 {
			return true, v[0], nil
		}
	}
	return false, "", nil
}

func CheckIsSensitiveKey(key string) (bool, error) {
	var re = regexp.MustCompile(`(?m)(?i)(uid)|(userid)|(user_id)|(nin)|(name)|(username)|(nick)`)
	regResult := re.FindAllString(key, -1)
	if len(regResult) > 0 {
		return true, nil
	}
	return false, nil
}

func GetJsResponse(in *input.CrawlResult, client *httpx.Client) (string, *JsonpInfo, error) {
	res, err := client.Request(in.Url, in.Method, "", in.Headers, "Jsonp")
	if err != nil {
		return "", nil, err
	}

	if res.StatusCode != 200 {
		return "", nil, errors.New(fmt.Sprintf("Fake Origin Referer Fail. Status code: %d", res.StatusCode))
	}
	info := &JsonpInfo{
		res.RequestDump,
		res.ResponseDump,
	}

	return res.Body, info, nil
}

func CheckJsRespAst(content string, funcName string) (bool, error) {
	// var params = []string{}
	// var vardiscover bool
	var Valid_Callback bool = false
	var Valid_Key bool = false

	obj := js.Options{}
	ast, err := js.Parse(parse.NewInputString(content), obj)
	if err != nil {
		return false, err
	}

	logging.Logger.Debugf("Scope: %s", ast.Scope.String())
	logging.Logger.Debugf("JS: %s", ast.String())
	// ast.BlockStmt.String()
	l := js.NewLexer(parse.NewInputString(content))
	for {
		tt, text := l.Next()
		// fmt.Println("text", string(text))
		switch tt {
		case js.ErrorToken:
			if l.Err() != io.EOF {
				logging.Logger.Errorln("Error on line:", l.Err())
			}
			return Valid_Key, nil
		case js.VarToken:
			// vardiscover = true
		case js.StringToken:
			if Valid_Callback {
				bexist, err := CheckIsSensitiveKey(string(text))
				if err != nil {
					return false, err
				}
				if bexist {
					Valid_Key = true
				}
			}
		case js.IdentifierToken:
			Identifier := string(text)
			// fmt.Println("IdentifierToken", Identifier)
			if Identifier == funcName {
				Valid_Callback = true
			}
		}
	}
}
