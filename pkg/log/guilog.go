package log

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"strings"
)

/**
  @author: yhy
  @since: 2023/4/20
  @desc: //TODO
**/

var GuiLog *GuiLogger

// GuiLogger logrus 自定义 hook
type GuiLogger struct {
	Ctx context.Context
}

// Levels 只定义 error 和 panic 等级的日志,其他日志等级不会触发 hook
func (g *GuiLogger) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
}

// Fire 将异常日志写入到指定日志文件中
func (g *GuiLogger) Fire(entry *logrus.Entry) error {

	message := entry.Message

	fileVal := ""

	packageName := strings.Split(SplitLast(entry.Caller.Function, "/"), ".")
	fileName := SplitLast(entry.Caller.File, "/")
	fileVal = fmt.Sprintf("[%s:%s(%s):%d]", packageName[0], fileName, packageName[1], entry.Caller.Line)
	messageFormat := "%s"

	timestamp := fmt.Sprintf("[%s]", entry.Time.Format("15:04:05"))

	runtime.EventsEmit(g.Ctx, "log", fmt.Sprintf("%s %s %s "+messageFormat+"\n", timestamp, entry.Level.String(), fileVal, message))
	return nil
}

// SplitLast 字符串分割获取最后一位
func SplitLast(str, sep string) string {
	arr := strings.Split(str, sep)
	return arr[len(arr)-1]
}
