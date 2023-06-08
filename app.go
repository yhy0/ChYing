package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	uuid "github.com/satori/go.uuid"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/ChYing/tools"
	"github.com/yhy0/ChYing/tools/burpSuite"
	"github.com/yhy0/ChYing/tools/decoder"
	"github.com/yhy0/ChYing/tools/fuzz"
	"github.com/yhy0/ChYing/tools/nucleiY"
	"github.com/yhy0/ChYing/tools/swagger"
	"github.com/yhy0/ChYing/tools/twj"
	"github.com/yhy0/logging"
	"os"
	"path/filepath"
	"strings"
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

// Menu 应用菜单
func (a *App) Menu() *menu.Menu {
	return menu.NewMenuFromItems(
		menu.SubMenu("承影", menu.NewMenuFromItems(
			menu.Text("关于", nil, func(_ *menu.CallbackData) {
				a.diag(conf.Description, false)
			}),
			menu.Text("检查更新", nil, func(_ *menu.CallbackData) {
				resp, err := httpx.Get("https://api.github.com/repos/yhy0/ChYing/tags")
				if err != nil {
					a.diag("检查更新出错\n"+err.Error(), true)
					return
				}

				lastVersion, err := jsonparser.GetString([]byte(resp.Body), "[0]", "name")
				if err != nil {
					a.diag("检查更新出错\n"+err.Error(), true)
					return
				}

				needUpdate := conf.Version < lastVersion
				msg := conf.VersionNewMsg
				btns := []string{conf.BtnConfirmText}
				if needUpdate {
					msg = fmt.Sprintf(conf.VersionOldMsg, lastVersion)
					btns = []string{"确定", "取消"}
				}
				selection, err := a.diag(msg, false, btns...)
				if err != nil {
					return
				}
				if needUpdate && selection == conf.BtnConfirmText {
					url := fmt.Sprintf("https://github.com/yhy0/ChYing/releases/tag/%s", lastVersion)
					runtime.BrowserOpenURL(a.ctx, url)
				}
			}),
			menu.Text(
				"主页",
				keys.Combo("H", keys.CmdOrCtrlKey, keys.ShiftKey),
				func(_ *menu.CallbackData) {
					runtime.BrowserOpenURL(a.ctx, "https://github.com/yhy0/ChYing/")
				},
			),
			menu.Separator(),
			menu.Text("退出", keys.CmdOrCtrl("Q"), func(_ *menu.CallbackData) {
				runtime.Quit(a.ctx)
			}),
		)),

		menu.EditMenu(),
		menu.SubMenu("Help", menu.NewMenuFromItems(
			menu.Text(
				"打开配置文件夹",
				keys.Combo("C", keys.CmdOrCtrlKey, keys.ShiftKey),
				func(_ *menu.CallbackData) {
					err := utils.OpenFolder(file.ChyingDir)
					if err != nil {
						a.diag("Failed to open folder: \n"+err.Error(), true)
						return
					}
				},
			),
		)),
	)
}

// diag ...
func (a *App) diag(message string, error bool, buttons ...string) (string, error) {
	if len(buttons) == 0 {
		buttons = []string{
			conf.BtnConfirmText,
		}
	}

	var t runtime.DialogType

	if error {
		t = runtime.ErrorDialog
	} else {
		t = runtime.InfoDialog
	}

	return runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          t,
		Title:         conf.Title,
		Message:       message,
		CancelButton:  conf.BtnCancelText,
		DefaultButton: conf.BtnConfirmText,
		Buttons:       buttons,
	})
}

// startup is called when the app starts. The context is saved ,so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	burpSuite.Ctx = ctx
	// 启动中间人代理
	burpSuite.Init()
	burpSuite.HotConf()

	if utils.IsPortOccupied(burpSuite.Settings.ProxyPort) {
		port, err := utils.GetRandomUnusedPort()
		if err != nil {
			logging.Logger.Errorln(err)
			burpSuite.Settings.ProxyPort = 65530
		}
		burpSuite.Settings.ProxyPort = port
	}

	go burpSuite.Run(burpSuite.Settings.ProxyPort)

	runtime.EventsEmit(ctx, "ProxyPort", burpSuite.Settings.ProxyPort)
	runtime.EventsEmit(ctx, "Exclude", burpSuite.Settings.Exclude)
	runtime.EventsEmit(ctx, "Include", burpSuite.Settings.Include)
	runtime.EventsEmit(ctx, "FilterSuffix", burpSuite.Settings.FilterSuffix)

	// 通知前端各种数据更改
	go func() {
		for {
			select {
			case percentage := <-twj.Percentage:
				runtime.EventsEmit(ctx, "Percentage", percentage)
			case percentage := <-fuzz.FuzzPercentage: // fuzz 的进度条
				runtime.EventsEmit(ctx, "FuzzPercentage", percentage)
			case _fuzz := <-fuzz.FuzzChan: // fuzz 表格数据
				runtime.EventsEmit(ctx, "Fuzz", _fuzz)
			case _swagger := <-swagger.SwaggerChan:
				if _swagger.StatusCode == 403 {
					fuzz.Bypass403(_swagger.Url, _swagger.Method)
				}
				runtime.EventsEmit(ctx, "swagger", _swagger)
			// burp 相关
			case history := <-burpSuite.HttpHistory:
				if len(burpSuite.Settings.Exclude) > 0 {
					if !utils.RegexpStr(burpSuite.Settings.Exclude, history.Host) {
						if len(burpSuite.Settings.Include) > 0 && utils.RegexpStr(burpSuite.Settings.Include, history.Host) {
							runtime.EventsEmit(ctx, "HttpHistory", history)
						} else {
							runtime.EventsEmit(ctx, "HttpHistory", history)
						}
					}
				} else if len(burpSuite.Settings.Include) > 0 && utils.RegexpStr(burpSuite.Settings.Include, history.Host) {
					runtime.EventsEmit(ctx, "HttpHistory", history)
				} else {
					runtime.EventsEmit(ctx, "HttpHistory", history)
				}

			case event := <-nucleiY.ResultEvent:
				res := nucleiY.Result{
					Url:      event.Matched,
					Name:     event.Info.Name,
					Request:  event.Request,
					Response: event.Response,
				}
				runtime.EventsEmit(ctx, "nucleiYRes", res)
			}
		}
	}()

	httpx.NewSession()
}

type Message struct {
	Msg   string `json:"msg"`
	Error string `json:"error"`
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
		logging.Logger.Errorln(err)
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
	// 爆破前判断是否有进程在爆破，如果有停止
	if twj.Flag {
		twj.Stop = true
		time.Sleep(1 * time.Second)
		twj.Stop = false
	}
	return twj.GenerateSignature()
}
func (a *App) TwjStop() {
	twj.Stop = true
	time.Sleep(2 * time.Second)
	twj.Stop = false
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
func (a *App) Settings(setting burpSuite.SettingUI) string {
	if burpSuite.Settings.ProxyPort != setting.ProxyPort && utils.IsPortOccupied(setting.ProxyPort) {
		return "端口被占用"
	} else {
		if burpSuite.Settings.ProxyPort != setting.ProxyPort {
			err := burpSuite.Restart(setting.ProxyPort)
			if err != "" {
				logging.Logger.Errorln(err)
				return err
			}
		}

		burpSuite.Settings.ProxyPort = setting.ProxyPort
		burpSuite.Settings.Exclude = utils.SplitStringByLines(setting.Exclude)
		burpSuite.Settings.Include = utils.SplitStringByLines(setting.Include)
		burpSuite.Settings.FilterSuffix = strings.Split(setting.FilterSuffix, ",")

		runtime.EventsEmit(a.ctx, "ProxyPort", burpSuite.Settings.ProxyPort)
		runtime.EventsEmit(a.ctx, "Exclude", strings.Join(burpSuite.Settings.Exclude, "\r\n"))
		runtime.EventsEmit(a.ctx, "Include", strings.Join(burpSuite.Settings.Include, "\r\n"))
		runtime.EventsEmit(a.ctx, "FilterSuffix", strings.Join(burpSuite.Settings.FilterSuffix, ","))

		// 更改配置文件
		exclude := ""
		if len(burpSuite.Settings.Exclude) == 0 {
			exclude = "  - \r\n"
		} else {
			for _, e := range burpSuite.Settings.Exclude {
				exclude += fmt.Sprintf("  - %s\r\n", e)
			}
		}

		include := ""
		if len(burpSuite.Settings.Include) == 0 {
			include = "  - \r\n"
		} else {
			for _, i := range burpSuite.Settings.Include {
				include += fmt.Sprintf("  - %s\r\n", i)
			}
		}
		filterSuffix := ""
		if len(burpSuite.Settings.FilterSuffix) == 0 {
			filterSuffix = "  - \r\n"
		} else {
			for _, i := range burpSuite.Settings.FilterSuffix {
				filterSuffix += fmt.Sprintf("  - %s\r\n", i)
			}
		}

		var defaultYamlByte = []byte(fmt.Sprintf("port: %d\r\nexclude:\r\n%sinclude:\r\n%s\r\nfilterSuffix:\r\n%s", burpSuite.Settings.ProxyPort, exclude, include, filterSuffix))

		err := burpSuite.WriteYamlConfig(defaultYamlByte)
		if err != nil {
			a.diag(err.Error(), true)
			return err.Error()
		}

		return ""
	}
}

// GetBurpSettings 配置
func (a *App) GetBurpSettings() *burpSuite.Setting {
	return burpSuite.Settings
}

// Intercept 拦截包
func (a *App) Intercept(intercept, wait bool, request string) int {
	if intercept {
		burpSuite.Intercept = true
	} else {
		burpSuite.Intercept = false
	}

	if wait && burpSuite.Sum != 0 {
		burpSuite.InterceptBody = request
		burpSuite.Sum -= 1
		<-burpSuite.Done
	}

	return burpSuite.Sum
}

// GetHistoryDump 代理记录
func (a *App) GetHistoryDump(id int) *burpSuite.HTTPBody {
	return burpSuite.HTTPBodyMap.ReadMap(id)
}

// InterceptSend 从 Intercept 发给 Repeater\Intruder 界面处理
func (a *App) InterceptSend(name string) {
	runtime.EventsEmit(a.ctx, name, burpSuite.HttpBodyInter)
	return
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

func (a *App) TaskList(out string) map[string]string {
	return tools.Tasklist(out)
}

// NucleiLoad 加载模板
func (a *App) NucleiLoad() []nucleiY.Options {
	nucleiY.New("")

	var options []nucleiY.Options
	for k, v := range nucleiY.Pocs {
		var child []string
		for _, t := range v {
			child = append(child, t.Info.Name)
		}
		options = append(options, nucleiY.Options{Label: k, Children: child})
	}
	return options
}

// NucleiY 漏洞扫描
func (a *App) NucleiY(target string, tag string, proxy string) string {
	templatesTempDir := filepath.Join(file.ChyingDir, "nucleiY")

	if _, err := os.Stat(templatesTempDir); err != nil {
		// 不存在，创建
		logging.Logger.Errorln("")
		return "nucleiY not find, https://github.com/yhy0/nucleiY"
	}

	return nucleiY.Scan(target, tag, proxy)
}
