package main

import (
	"context"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/ChYing/pkg/log"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/ChYing/tools/burpSuite"
	"github.com/yhy0/ChYing/tools/decoder"
	"github.com/yhy0/ChYing/tools/fuzz"
	"github.com/yhy0/ChYing/tools/swagger"
	"github.com/yhy0/ChYing/tools/twj"
	"github.com/yhy0/logging"
	"time"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 启动中间人代理
	go burpSuite.Run(conf.ProxyPort)

	runtime.EventsEmit(ctx, "ProxyPort", conf.ProxyPort)
	// 通知前端各种数据更改
	go func() {
		for {
			select {
			case percentage := <-twj.Percentage:
				runtime.EventsEmit(ctx, "Percentage", percentage)
			case percentage := <-fuzz.FuzzPercentage: // fuzz 的进度条
				runtime.EventsEmit(ctx, "FuzzPercentage", percentage)
			case _fuzz := <-fuzz.FuzzChan: // fuzz 表格数据
				// todo 先在这里对 403 页面进行 bypass 测试，后续再看看有没有必要在前端显示做一个按钮开关控制
				if _fuzz.StatusCode == 403 {
					bypass := fuzz.Bypass403(_fuzz.Url, "")
					if bypass != nil {
						fuzz.FuzzChan <- *bypass
					}
				}
				runtime.EventsEmit(ctx, "Fuzz", _fuzz)
			case _swagger := <-swagger.SwaggerChan:
				if _swagger.StatusCode == 403 {
					bypass := fuzz.Bypass403(_swagger.Url, _swagger.Method)
					if bypass != nil {
						fuzz.FuzzChan <- *bypass
					}
				}
				runtime.EventsEmit(ctx, "swagger", _swagger)
			// burp 相关
			case history := <-burpSuite.HttpHistory:
				runtime.EventsEmit(ctx, "HttpHistory", history)
			}
		}
	}()

	log.GuiLog = &log.GuiLogger{
		Ctx: ctx,
	}
	logging.Logger.AddHook(log.GuiLog)

	httpx.NewSession()
}

type Message struct {
	Msg   string
	Error string
}

func (a *App) Parser(jwt string) *twj.Jwt {
	parseJWT, err := twj.ParseJWT(jwt)

	if err != nil {
		parseJWT = &twj.Jwt{
			Header:       "",
			Payload:      "",
			Message:      err.Error(),
			SignatureStr: "",
		}
		return parseJWT
	}

	return parseJWT
}

func (a *App) Verify(jwt string, secret string) (msg Message) {
	parseJWT, err := twj.Verify(jwt, secret)

	if err != nil {
		fmt.Println(err)
		msg.Msg = ""
		msg.Error = err.Error()
		return
	}
	h, err := json.Marshal(parseJWT)

	msg.Msg = string(h)
	msg.Error = ""
	return
}

func (a *App) Brute() string {
	return twj.GenerateSignature()
}

func (a *App) Proxy(proxy string) (msg Message) {
	if proxy == "" {
		conf.Proxy = ""
		httpx.NewSession()
	} else {
		_, err := httpx.ValidateProxyURL(proxy)
		if err != nil {
			msg.Msg = "代理设置失败"
			msg.Error = err.Error()
			return
		}
		msg.Msg = "代理设置成功: " + proxy
		msg.Error = ""
		conf.Proxy = proxy
		httpx.NewSession()
		return
	}
	return
}

// Swagger 扫描
func (a *App) Swagger(target string) {
	if target != "" {
		swagger.Scan(target)
	}
}

func (a *App) Fuzz(target string, actions []string, filePath string) string {
	if target != "" && len(actions) > 0 {
		err := fuzz.Fuzz(target, actions, filePath)
		if err != nil {
			return err.Error()
		}
	} else {
		return "目标和模式不能为空"
	}
	return ""
}

func (a *App) FuzzStop() {
	fuzz.Stop = true
	time.Sleep(2 * time.Second)
	fuzz.Stop = false
}

// burp 相关

// Settings 配置
func (a *App) Settings(port string) string {
	logging.Logger.Infoln(conf.ProxyPort, port)
	if conf.ProxyPort == port {
		return ""
	}

	if utils.IsPortOccupied(port) {
		return "端口被占用"
	} else {
		err := burpSuite.Restart(port)
		logging.Logger.Errorln(err)
		if err != "" {
			return err
		}
		conf.ProxyPort = port
		logging.Logger.Infoln(conf.ProxyPort, port)
		runtime.EventsEmit(a.ctx, "ProxyPort", conf.ProxyPort)
		return ""
	}
}

// GetProxyPort 配置
func (a *App) GetProxyPort() string {
	return conf.ProxyPort
}

// GetHistoryDump 代理记录
func (a *App) GetHistoryDump(id int) *burpSuite.HTTPBody {
	return burpSuite.HTTPBodyMap.ReadMap(id)
}

// SendToRepeater 发给 Repeater 界面处理
func (a *App) SendToRepeater(id int) {
	runtime.EventsEmit(a.ctx, "RepeaterBody", burpSuite.HTTPBodyMap.ReadMap(id))
	return
}

// Raw Repeater 请求
func (a *App) Raw(request string, target string, id string) (httpBody *burpSuite.HTTPBody) {
	// 说明第一次
	if id == "" {
		id = uuid.NewV4().String()
	}

	resp, err := httpx.Raw(request, target)
	if err != nil {
		return
	}

	httpBody = &burpSuite.HTTPBody{
		TargetUrl: target,
		Request:   resp.RequestDump,
		Response:  resp.ResponseDump,
		UUID:      id,
	}
	value, ok := burpSuite.RepeaterBodyMap[id]
	if ok {
		_id := len(value)
		value[_id] = httpBody
		return
	}

	// 初始化
	burpSuite.RepeaterBodyMap[id] = make(map[int]*burpSuite.HTTPBody)

	burpSuite.RepeaterBodyMap[id][0] = httpBody

	return
}

// SendToIntruder 发给 Intruder 界面处理
func (a *App) SendToIntruder(id int) {
	runtime.EventsEmit(a.ctx, "IntruderBody", burpSuite.HTTPBodyMap.ReadMap(id))
	return
}

// Intruder 处理 Intruder 传来的参数
func (a *App) Intruder(target string, req string, payloads []string, rules []string, attackType string, uuid string) {
	for i, rule := range rules {
		if rule == "" {
			rules[i] = "None"
		}
	}

	burpSuite.Intruder(target, req, payloads, rules, attackType, uuid, a.ctx)
}

// GetAttackDump Intruder attack 记录
func (a *App) GetAttackDump(uuid string, id int) *burpSuite.HTTPBody {
	fmt.Println(uuid, id)
	return burpSuite.IntruderMap[uuid].ReadMap(id)
}

func (a *App) Decoder(str string, mode string) string {
	switch mode {
	case "DecodeUnicode":
		return decoder.DecodeUnicode(str)
	case "EncodeUnicode":
		return decoder.EncodeUnicode(str)
	case "DecodeURL":
		return decoder.DecodeURL(str)
	case "EncodeURL":
		return decoder.EncodeURL(str)
	case "DecodeBase64":
		return decoder.DecodeBase64(str)
	case "EncodeBase64":
		return decoder.EncodeBase64(str)
	case "DecodeHex":
		return decoder.DecodeHex(str)
	case "EncodeHex":
		return decoder.EncodeHex(str)
	case "MD5":
		return decoder.Md5(str)
	default:
		return str
	}
}
