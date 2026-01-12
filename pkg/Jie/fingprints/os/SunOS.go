package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type SunOSPlugin struct{}

func (p SunOSPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "SunOS") {
            return true
        }
    }

    if v, ok := headers["X-Powered-By"]; ok {
        if strings.Contains(strings.Join(v, ""), "SunOS") {
            return true
        }
    }
    return false
}

func (p SunOSPlugin) Name() string {
    return "SunOS"
}
