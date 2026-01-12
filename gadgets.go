package main

import (
	"encoding/json"
	"github.com/yhy0/ChYing/pkg/coder/twj"
	"github.com/yhy0/ChYing/pkg/decoder"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
	"strings"
	"time"
)

/**
   @author yhy
   @since 2024/8/12
   @desc //TODO
**/

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

func (a *App) Verify(jwt string, secret string) (res Result) {
	parseJWT, err := twj.Verify(jwt, secret)

	if err != nil {
		logging.Logger.Errorln(err)

		res.Error = err.Error()
		return
	}
	h, err := json.Marshal(parseJWT)

	res.Data = string(h)
	return
}

func (a *App) Brute(jwtMsg string, jwtPath string) string {
	// 爆破前判断是否有进程在爆破，如果有停止
	if twj.Flag {
		twj.Stop = true
		time.Sleep(1 * time.Second)
		twj.Stop = false
	}
	return twj.GenerateSignature(jwtMsg, jwtPath)
}

func (a *App) TwjStop() {
	twj.Stop = true
	time.Sleep(2 * time.Second)
	twj.Stop = false
}

func (a *App) PredictionApi(api string) []string {
	logging.Logger.Debug(strings.Split(api, "\n"))
	return utils.PredictionApi(strings.Split(api, "\n"), 1)
}
