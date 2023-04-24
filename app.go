package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/ChYing/pkg/log"
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
			case swagger := <-swagger.SwaggerChan:
				runtime.EventsEmit(ctx, "swagger", swagger)
			}
		}
	}()

	log.GuiLog = &log.GuiLogger{
		Ctx: ctx,
	}
	logging.Logger.AddHook(log.GuiLog)

	httpx.NewSession()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
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
