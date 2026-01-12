package conf

import (
    "github.com/thoas/go-funk"
    "github.com/yhy0/ChYing/conf/file"
    "github.com/yhy0/logging"
    "strings"
    "sync"
)

/**
   @author yhy
   @since 2024/12/27
   @desc //TODO
**/

var Address sync.Map // history 的 id 和 f *proxy.Flow 中的 Id uuid.UUID 对应，主要用于在 Jie 中更新数据对应
var HttpMarkerChan = make(chan HttpMarker, 1)

type HttpMarker struct {
    Id    int64  `json:"id"`
    Color string `json:"color"`
    Note  string `json:"note"`
}

func MarkerMitmRule(request string, response string) (string, string) {
    var note []string
    var color []string
    
    for _, rule := range file.MitmRules {
        if rule.EnableForRequest || rule.EnableForBody || rule.EnableForHeader {
            matches := rule.RegexCompiled.FindString(request)
            if matches != "" {
                color = append(color, rule.Color)
                note = append(note, rule.VerboseName)
                logging.Logger.Debugln(matches, color, rule)
            }
        }
        
        if rule.EnableForResponse || rule.EnableForHeader || rule.EnableForBody {
            matches := rule.RegexCompiled.FindString(response)
            if matches != "" {
                color = append(color, rule.Color)
                note = append(note, rule.VerboseName)
                logging.Logger.Debugln(matches, color, rule)
            }
        }
    }
    if len(color) > 0 {
        if funk.Contains(color, "red") {
            return "red", strings.Join(note, ", ")
        }
        if funk.Contains(color, "yellow") {
            return "yellow", strings.Join(note, ", ")
        }
    }
    return "", strings.Join(note, ", ")
}
