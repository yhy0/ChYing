package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type WindowsCEPlugin struct{}

func (p WindowsCEPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "WinCE") {
            return true
        }
    }

    return false
}

func (p WindowsCEPlugin) Name() string {
    return "Windows CE"
}
