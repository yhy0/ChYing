package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type UbuntuPlugin struct{}

func (p UbuntuPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "Ubuntu") {
            return true
        }
    }

    if v, ok := headers["X-Powered-By"]; ok {
        if strings.Contains(strings.Join(v, ""), "Ubuntu") {
            return true
        }
    }
    return false
}

func (p UbuntuPlugin) Name() string {
    return "Ubuntu"
}
