package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/pkg/gadgets/fuzz"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2024/8/12
   @desc //TODO
**/

func (a *App) Fuzz(target string, actions []string, filePath string) string {
	if target == "" || len(actions) == 0 {
		return "目标或模式不能为空"
	}
	targets := strings.Split(target, "\n")
	Pool, _ = ants.NewPool(conf.Parallelism)

	for _, t := range targets {
		Pool.Submit(func() {
			logging.Logger.Infoln("[fuzz] start", t, actions)
			err := fuzz.Fuzz(t, actions)
			if err != nil {
				logging.Logger.Errorln("[fuzz] error:", t, err)
				Notify <- []string{fmt.Sprintf("fuzz error: %s\n %s", t, err.Error()), "error"}
				return
			}
			Notify <- []string{fmt.Sprintf("fuzz finish: %s", t), "success"}
		})
	}

	return ""
}

func (a *App) FuzzStop() {
	fuzz.Stop = true
	time.Sleep(2 * time.Second)
	fuzz.Stop = false
}
