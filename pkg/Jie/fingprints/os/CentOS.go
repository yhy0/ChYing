package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type CentOSPlugin struct{}

func (p CentOSPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "CentOS") {
            return true
        }
    }

    if v, ok := headers["X-Powered-By"]; ok {
        if strings.Contains(strings.Join(v, ""), "CentOS") {
            return true
        }
    }
    return false
}

func (p CentOSPlugin) Name() string {
    return "CentOS"
}
