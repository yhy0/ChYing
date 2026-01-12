package programing

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type PythonPlugin struct{}

func (p PythonPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "Python") {
            return true
        }
    }
    return false
}

func (p PythonPlugin) Name() string {
    return "Python"
}
